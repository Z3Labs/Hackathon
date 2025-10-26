package diagnosis

import (
	"context"
	"fmt"
)

// SimpleDiagnosis 简化版诊断函数，类似 Python 的 simple_diagnosis
// 这个函数封装了完整的诊断流程，便于直接调用
func SimpleDiagnosis(
	prompt string,
	anthropicAPIKey string,
	prometheusURL string,
	githubToken string,
	enableGitHubMCP bool,
	githubToolsets string,
	anthropicBaseURL string,
	model string,
) (string, error) {
	// 创建客户端
	client := NewMCPDiagnosisClient(anthropicAPIKey, anthropicBaseURL)
	
	// 构建 MCP 服务器配置
	var mcpServers []MCPConfig
	
	// 1. Prometheus MCP 配置（使用 Docker）
	mcpServers = append(mcpServers, MCPConfig{
		Name:    "prometheus",
		Command: "docker",
		Args: []string{
			"run",
			"-i",
			"--rm",
			"-e", "PROMETHEUS_URL=" + prometheusURL,
			"ghcr.io/pab1it0/prometheus-mcp-server:latest",
		},
		Env: map[string]string{}, // 环境变量通过 -e 参数传递
	})
	
	// 2. GitHub MCP 配置（使用 Docker，可选）
	if enableGitHubMCP && githubToken != "" {
		mcpServers = append(mcpServers, MCPConfig{
			Name:    "github",
			Command: "docker",
			Args: []string{
				"run",
				"-i",
				"--rm",
				"-e", "GITHUB_PERSONAL_ACCESS_TOKEN=" + githubToken,
				"-e", "GITHUB_TOOLSETS=" + githubToolsets,
				"ghcr.io/github/github-mcp-server:latest",
			},
			Env: map[string]string{}, // 环境变量通过 -e 参数传递
		})
	}
	
	// 构建诊断配置
	config := DiagnosisConfig{
		Prompt:           prompt,
		AnthropicAPIKey:  anthropicAPIKey,
		AnthropicBaseURL: anthropicBaseURL,
		Model:            model,
		MaxTokens:        4096,
		MaxIterations:    20,
		MCPServers:       mcpServers,
	}
	
	// 执行诊断
	ctx := context.Background()
	return client.Diagnose(ctx, config)
}

// BuildPrometheusPrompt 构建 Prometheus 诊断提示词
func BuildPrometheusPrompt(alertName, instance string) string {
	return fmt.Sprintf(`你是一个专业的 DevOps 运维诊断专家。

收到以下告警信息：
- 告警名称: %s
- 实例: %s

请使用 Prometheus MCP 工具查询相关指标，分析问题原因并生成诊断报告。`, alertName, instance)
}

// BuildGitHubAnalysisPrompt 构建 GitHub 代码分析提示词
func BuildGitHubAnalysisPrompt(repoOwner, repoName string) string {
	return fmt.Sprintf(`你是一个代码审查专家。

请分析 %s/%s 仓库的最新 release：
1. 获取最新 release
2. 提取其中的 PR 列表
3. 分析每个 PR 的代码变更，查找潜在的 bug
4. 生成详细的 bug 报告

同时检查 Prometheus 监控指标是否正常。`, repoOwner, repoName)
}
