package diagnosis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPConfig MCP 服务器配置
type MCPConfig struct {
	Name    string            // 服务器名称（如 "prometheus", "github"）
	Command string            // 启动命令（如 "python"）
	Args    []string          // 命令参数（如 ["-m", "prometheus_mcp_server.main"]）
	Env     map[string]string // 环境变量
}

// DiagnosisConfig 诊断配置
type DiagnosisConfig struct {
	Prompt           string      // AI 提示词
	AnthropicAPIKey  string      // Anthropic API Key
	AnthropicBaseURL string      // API Base URL（可选，用于兼容服务）
	Model            string      // 模型名称
	MaxTokens        int64       // 最大 token 数
	MaxIterations    int         // 最大迭代次数
	MCPServers       []MCPConfig // MCP 服务器配置列表
}

// MCPDiagnosisClient MCP 诊断客户端
type MCPDiagnosisClient struct {
	anthropicClient anthropic.Client
	mcpClient       *mcp.Client
	sessions        map[string]*mcp.ClientSession // server_name -> session
	toolRouter      map[string]string             // tool_name -> server_name
}

// NewMCPDiagnosisClient 创建新的 MCP 诊断客户端
func NewMCPDiagnosisClient(apiKey, baseURL string) *MCPDiagnosisClient {
	var opts []option.RequestOption
	opts = append(opts, option.WithAPIKey(apiKey))
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	return &MCPDiagnosisClient{
		anthropicClient: anthropic.NewClient(opts...),
		mcpClient: mcp.NewClient(&mcp.Implementation{
			Name:    "hackathon-diagnosis-client",
			Version: "v1.0.0",
		}, nil),
		sessions:   make(map[string]*mcp.ClientSession),
		toolRouter: make(map[string]string),
	}
}

// ConnectMCPServers 连接到所有 MCP 服务器
func (c *MCPDiagnosisClient) ConnectMCPServers(ctx context.Context, configs []MCPConfig) error {
	for _, config := range configs {
		log.Printf("[%s MCP] 启动中...", config.Name)

		// 创建命令
		cmd := exec.CommandContext(ctx, config.Command, config.Args...)

		// 设置环境变量
		if len(config.Env) > 0 {
			env := make([]string, 0, len(config.Env))
			for k, v := range config.Env {
				env = append(env, fmt.Sprintf("%s=%s", k, v))
			}
			cmd.Env = env
		}

		// 创建传输
		transport := &mcp.CommandTransport{Command: cmd}

		// 连接到 MCP 服务器
		session, err := c.mcpClient.Connect(ctx, transport, nil)
		if err != nil {
			return fmt.Errorf("连接 %s MCP 失败: %w", config.Name, err)
		}

		// 保存会话
		c.sessions[config.Name] = session
		log.Printf("[%s MCP] ✅ 连接成功", config.Name)
	}

	return nil
}

// DiscoverTools 发现所有工具并构建路由表
func (c *MCPDiagnosisClient) DiscoverTools(ctx context.Context) ([]*mcp.Tool, error) {
	var allTools []*mcp.Tool

	for serverName, session := range c.sessions {
		// 获取工具列表
		result, err := session.ListTools(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("获取 %s 工具列表失败: %w", serverName, err)
		}

		log.Printf("[%s MCP] 加载了 %d 个工具:", serverName, len(result.Tools))
		for _, tool := range result.Tools {
			log.Printf("  - %s", tool.Name)

			// 构建路由表
			c.toolRouter[tool.Name] = serverName
			allTools = append(allTools, tool)
		}
	}

	log.Printf("[MCP] 总共加载了 %d 个工具", len(allTools))
	return allTools, nil
}

// ConvertToAnthropicTools 将 MCP 工具转换为 Anthropic 格式
func (c *MCPDiagnosisClient) ConvertToAnthropicTools(mcpTools []*mcp.Tool) []anthropic.ToolUnionParam {
	tools := make([]anthropic.ToolUnionParam, 0, len(mcpTools))

	for _, tool := range mcpTools {
		// 将 InputSchema 转换为 anthropic.ToolInputSchemaParam
		var inputSchemaParam anthropic.ToolInputSchemaParam
		
		if tool.InputSchema != nil {
			// 将 MCP InputSchema 转换为 map
			schemaBytes, _ := json.Marshal(tool.InputSchema)
			var schemaMap map[string]any
			json.Unmarshal(schemaBytes, &schemaMap)
			
			// 提取 properties
			if props, ok := schemaMap["properties"].(map[string]any); ok {
				inputSchemaParam.Properties = props
			}
		}

		toolParam := anthropic.ToolParam{
			Name:        tool.Name,
			Description: anthropic.String(tool.Description),
			InputSchema: inputSchemaParam,
		}

		tools = append(tools, anthropic.ToolUnionParam{
			OfTool: &toolParam,
		})
	}

	return tools
}

// CallToolViaRouter 通过路由器调用工具
func (c *MCPDiagnosisClient) CallToolViaRouter(ctx context.Context, toolName string, arguments map[string]any) (string, error) {
	// 查找对应的服务器
	serverName, ok := c.toolRouter[toolName]
	if !ok {
		return "", fmt.Errorf("未找到工具: %s", toolName)
	}

	session, ok := c.sessions[serverName]
	if !ok {
		return "", fmt.Errorf("未找到服务器会话: %s", serverName)
	}

	log.Printf("[%s MCP] 调用工具: %s", serverName, toolName)
	argJSON, _ := json.MarshalIndent(arguments, "", "  ")
	log.Printf("[%s MCP] 参数: %s", serverName, string(argJSON))

	// 调用 MCP 工具
	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: arguments,
	})
	if err != nil {
		return "", fmt.Errorf("工具调用失败: %w", err)
	}

	// 提取文本结果
	var resultTexts []string
	for _, content := range result.Content {
		// MCP Content 需要通过 JSON 序列化来提取文本
		contentBytes, _ := json.Marshal(content)
		var contentMap map[string]any
		json.Unmarshal(contentBytes, &contentMap)
		if text, ok := contentMap["text"].(string); ok {
			resultTexts = append(resultTexts, text)
		}
	}

	resultText := strings.Join(resultTexts, "\n")
	log.Printf("[%s MCP] 结果: %s", serverName, truncate(resultText, 200))

	return resultText, nil
}

// Diagnose 执行诊断分析
func (c *MCPDiagnosisClient) Diagnose(ctx context.Context, config DiagnosisConfig) (string, error) {
	// 1. 连接 MCP 服务器
	if err := c.ConnectMCPServers(ctx, config.MCPServers); err != nil {
		return "", err
	}
	defer c.Close()

	// 2. 发现工具并构建路由表
	mcpTools, err := c.DiscoverTools(ctx)
	if err != nil {
		return "", err
	}

	// 3. 转换为 Anthropic 工具格式
	tools := c.ConvertToAnthropicTools(mcpTools)

	// 4. 构建初始消息
	log.Println("[AI] 开始分析...")
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(config.Prompt)),
	}

	// 5. 多轮对话循环
	for iteration := 0; iteration < config.MaxIterations; iteration++ {
		log.Printf("[AI] 第 %d 轮对话", iteration+1)

		// 调用 Anthropic API
		response, err := c.anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.Model(config.Model),
			MaxTokens: config.MaxTokens,
			Messages:  messages,
			Tools:     tools,
		})
		if err != nil {
			return "", fmt.Errorf("API 调用失败: %w", err)
		}

		log.Printf("[AI] Stop reason: %s", response.StopReason)

		// 检查是否完成
		if response.StopReason == "end_turn" {
			// 提取最终响应文本
			var finalText strings.Builder
			for _, block := range response.Content {
				if textBlock, ok := block.AsAny().(anthropic.TextBlock); ok {
					finalText.WriteString(textBlock.Text)
				}
			}

			log.Println("[AI] ✅ 分析完成")
			log.Println("#####") // 切割日志和 AI 结果
			return strings.TrimSpace(finalText.String()), nil
		}

		// 添加 AI 响应到消息历史
		messages = append(messages, response.ToParam())

		// 处理工具调用
		var toolResults []anthropic.ContentBlockParamUnion
		for _, block := range response.Content {
			if toolUse, ok := block.AsAny().(anthropic.ToolUseBlock); ok {
				// 调用工具 - 解析 Input
				var arguments map[string]any
				if len(toolUse.Input) > 0 {
					json.Unmarshal(toolUse.Input, &arguments)
				}

				result, err := c.CallToolViaRouter(ctx, toolUse.Name, arguments)

				if err != nil {
					// 工具调用失败，返回错误
					toolResults = append(toolResults, anthropic.NewToolResultBlock(
						toolUse.ID,
						fmt.Sprintf("工具调用失败: %v", err),
						true, // isError
					))
				} else {
					// 工具调用成功
					toolResults = append(toolResults, anthropic.NewToolResultBlock(
						toolUse.ID,
						result,
						false,
					))
				}
			}
		}

		// 如果没有工具调用，退出循环
		if len(toolResults) == 0 {
			break
		}

		// 添加工具结果到消息历史
		messages = append(messages, anthropic.MessageParam{
			Role:    anthropic.MessageParamRoleUser,
			Content: toolResults,
		})
	}

	return "诊断分析达到最大迭代次数，请检查配置或联系技术支持。", nil
}

// Close 关闭所有 MCP 会话
func (c *MCPDiagnosisClient) Close() {
	for name, session := range c.sessions {
		if err := session.Close(); err != nil {
			log.Printf("[%s MCP] 关闭会话失败: %v", name, err)
		} else {
			log.Printf("[%s MCP] 会话已关闭", name)
		}
	}
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
