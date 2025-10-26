package diagnosis

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/Z3Labs/Hackathon/backend/internal/config"
)

const (
	// Docker 容器配置
	diagnosisContainerName = "diagnosis-service"
	diagnosisImageName     = "diagnosis-service:latest"

	pyReturnSplit = "#####"
)

var jsonRegex = regexp.MustCompile(`\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}`)

type mcpClient struct {
	containerName  string        // Docker 容器名称
	scriptPath     string        // 容器内 Python 脚本路径
	apiKey         string        // AI API Key
	baseURL        string        // AI Base URL
	model          string        // AI 模型名称
	prometheusURL  string        // Prometheus URL
	githubToken    string        // GitHub Personal Access Token
	githubToolsets string        // GitHub MCP 工具集
	timeout        time.Duration // 超时时间
	logger         logx.Logger   // 日志记录器
}

// NewMCPClient 创建 MCP AI 客户端
func NewMCPClient(cfg config.AIConfig) AIClient {
	client := &mcpClient{
		containerName:  diagnosisContainerName,
		scriptPath:     "/app/diagnosis_runner.py", // 容器内脚本路径
		apiKey:         cfg.APIKey,
		baseURL:        cfg.BaseURL,
		model:          cfg.Model,
		prometheusURL:  cfg.PrometheusURL,
		githubToken:    cfg.GitHubToken,    // GitHub Token
		githubToolsets: cfg.GitHubToolsets, // GitHub 工具集
		timeout:        time.Duration(cfg.Timeout) * time.Second,
		logger:         logx.WithContext(context.Background()), // 初始化日志记录器
	}

	// 启动容器（如果未运行）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.ensureContainer(ctx); err != nil {
		client.logger.Errorf("启动诊断服务容器失败: %v", err)
	}

	return client
}

func (c *mcpClient) GenerateCompletion(ctx context.Context, prompt string) (string, int, error) {
	// 设置超时（MCP 调用需要更长时间，因为涉及多轮工具调用）
	ctx, cancel := context.WithTimeout(ctx, c.timeout*2)
	defer cancel()

	// 优先使用 Go 版本的 MCP 调用
	report, err := SimpleDiagnosis(prompt, c.apiKey, c.prometheusURL, c.githubToken, true, c.githubToolsets, c.baseURL, c.model)
	if err == nil {
		returnValue := strings.Trim(strings.TrimSpace(report), "\n")
		findString := jsonRegex.FindString(returnValue)
		if findString != "" {
			return findString, 0, nil
		}
		return "", 0, fmt.Errorf("%s", returnValue)
	}
	c.logger.Errorf("Go 版本 MCP 调用失败，回退到 Docker 容器调用: %v", err)

	return c.CallWithPyDocker(ctx, prompt)
}

func (c *mcpClient) CallWithPyDocker(ctx context.Context, prompt string) (string, int, error) {
	// 若报错，则回退到使用 Docker 容器调用 Python 脚本
	// 确保容器运行
	if err := c.ensureContainer(ctx); err != nil {
		return "", 0, fmt.Errorf("确保容器运行失败: %w", err)
	}

	// 使用 docker exec 调用容器内的 Python 脚本
	args := []string{
		"exec", "-i",
		c.containerName,
		"python", c.scriptPath,
		"--prompt", prompt,
		"--api-key", c.apiKey,
		"--base-url", c.baseURL,
		"--model", c.model,
		"--prometheus-url", c.prometheusURL,
	}

	// 添加 GitHub MCP 参数（如果提供了 token，自动启用）
	if c.githubToken != "" {
		args = append(args, "--github-token", c.githubToken)
		if c.githubToolsets != "" {
			args = append(args, "--github-toolsets", c.githubToolsets)
		}
	}

	cmd := exec.CommandContext(ctx, "docker", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 执行命令
	err := cmd.Run()
	if err != nil {
		return "", 0, fmt.Errorf("生成分析报告失败: %w\n, stdout: %s\n stderr: %s", err, stdout.String(), stderr.String())
	}

	// 直接返回文本结果
	result := strings.TrimSpace(stdout.String())

	if result == "" {
		return "", 0, fmt.Errorf("python 脚本返回空结果")
	}
	split := strings.Split(result, pyReturnSplit)

	// MCP 模式下无法获取准确的 token 使用量，返回 0
	if len(split) > 1 {
		c.logger.Infof("Python 脚本执行成功，执行日志: \n%s", split[0])
		c.logger.Infof("report: \n%s", split[1])
		returnValue := strings.Trim(strings.TrimSpace(strings.Join(split[1:], pyReturnSplit)), "\n")
		findString := jsonRegex.FindString(returnValue)
		if findString != "" {
			return findString, 0, nil
		}
		return "", 0, fmt.Errorf("%s", returnValue)
	}
	return "", 0, fmt.Errorf("%s", result)
}

// ensureContainer 确保诊断服务容器正在运行
func (c *mcpClient) ensureContainer(ctx context.Context) error {
	// 检查容器是否运行
	if c.isContainerRunning(ctx) {
		return nil
	}

	// 容器未运行，尝试启动
	return c.startContainer(ctx)
}

// isContainerRunning 检查容器是否正在运行
func (c *mcpClient) isContainerRunning(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "docker", "ps", "--filter", fmt.Sprintf("name=%s", c.containerName), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// 检查输出中是否包含容器名称
	return strings.TrimSpace(string(output)) == c.containerName
}

// startContainer 启动诊断服务容器
func (c *mcpClient) startContainer(ctx context.Context) error {
	// 先尝试启动已存在的容器
	if c.tryStartExistingContainer(ctx) {
		return nil
	}

	// 容器不存在，创建并启动新容器
	c.logger.Infof("诊断服务启动容器: %s", c.containerName)

	args := []string{
		"run",
		"-d",                      // 后台运行
		"--name", c.containerName, // 容器名称
		"--restart", "unless-stopped", // 自动重启策略
		diagnosisImageName, // 镜像名称
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("启动容器失败: %w, output: %s", err, string(output))
	}

	c.logger.Info("诊断服务容器启动成功")
	return nil
}

// tryStartExistingContainer 尝试启动已存在但停止的容器
func (c *mcpClient) tryStartExistingContainer(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "docker", "start", c.containerName)
	err := cmd.Run()
	if err == nil {
		c.logger.Info("诊断服务已启动现有容器")
		return true
	}
	return false
}

// StopContainer 停止诊断服务容器（用于优雅关闭）
func (c *mcpClient) StopContainer() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c.logger.Infof("诊断服务停止容器: %s", c.containerName)

	cmd := exec.CommandContext(ctx, "docker", "stop", c.containerName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("停止容器失败: %w", err)
	}

	c.logger.Info("诊断服务容器已停止")
	return nil
}
