package diagnosis

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/config"
)

type mcpClient struct {
	pythonPath     string        // Python 解释器路径
	scriptPath     string        // Python 脚本路径
	apiKey         string        // AI API Key
	baseURL        string        // AI Base URL
	model          string        // AI 模型名称
	prometheusURL  string        // Prometheus URL
	timeout        time.Duration // 超时时间
}

// NewMCPClient 创建 MCP AI 客户端
func NewMCPClient(cfg config.AIConfig) AIClient {
	// 获取当前文件所在目录
	_, filename, _, _ := runtime.Caller(0)
	diagnosisDir := filepath.Dir(filename)
	scriptPath := filepath.Join(diagnosisDir, "py", "diagnosis_runner.py")
	
	// 使用虚拟环境中的 Python（包含所有依赖）
	pythonPath := filepath.Join(diagnosisDir, "py", "venv", "bin", "python")

	return &mcpClient{
		pythonPath:    pythonPath,
		scriptPath:    scriptPath,
		apiKey:        cfg.APIKey,
		baseURL:       cfg.BaseURL,
		model:         cfg.Model,
		prometheusURL: cfg.PrometheusURL,
		timeout:       time.Duration(cfg.Timeout) * time.Second,
	}
}

func (c *mcpClient) GenerateCompletion(ctx context.Context, prompt string) (string, int, error) {
	// 设置超时（MCP 调用需要更长时间，因为涉及多轮工具调用）
	ctx, cancel := context.WithTimeout(ctx, c.timeout*2)
	defer cancel()

	// 构建 Python 脚本参数
	args := []string{
		c.scriptPath,
		"--prompt", prompt,
		"--api-key", c.apiKey,
		"--base-url", c.baseURL,
		"--model", c.model,
		"--prometheus-url", c.prometheusURL,
	}

	// 执行 Python 脚本
	cmd := exec.CommandContext(ctx, c.pythonPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 执行命令
	err := cmd.Run()
	if err != nil {
		return "", 0, fmt.Errorf("执行 Python 脚本失败: %w, stderr: %s", err, stderr.String())
	}

	// 直接返回文本结果（不再解析 JSON）
	result := strings.TrimSpace(stdout.String())
	
	if result == "" {
		return "", 0, fmt.Errorf("Python 脚本返回空结果")
	}

	// MCP 模式下无法获取准确的 token 使用量，返回 0
	return result, 0, nil
}
