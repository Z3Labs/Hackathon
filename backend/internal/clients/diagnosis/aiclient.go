package diagnosis

import (
	"context"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"

	"github.com/Z3Labs/Hackathon/backend/internal/config"
)

type AIClient interface {
	GenerateCompletion(ctx context.Context, prompt string) (response string, tokensUsed int, err error)
}

type openaiClient struct {
	client  *openai.Client
	model   string
	timeout time.Duration
}

// NewOpenAIClient 创建 OpenAI 兼容的 AI 客户端
// 支持 OpenAI、Claude (via proxy)、通义千问等所有兼容 OpenAI API 的服务
func NewOpenAIClient(cfg config.AIConfig) AIClient {
	config := openai.DefaultConfig(cfg.APIKey)

	// 如果配置了自定义 BaseURL，则使用自定义端点
	if cfg.BaseURL != "" {
		config.BaseURL = cfg.BaseURL
	}

	return &openaiClient{
		client:  openai.NewClientWithConfig(config),
		model:   cfg.Model,
		timeout: time.Duration(cfg.Timeout) * time.Second,
	}
}

func (c *openaiClient) GenerateCompletion(ctx context.Context, prompt string) (string, int, error) {
	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// 调用 OpenAI API
	resp, err := c.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   2000,
	})

	if err != nil {
		return "", 0, err
	}

	if len(resp.Choices) == 0 {
		return "", 0, fmt.Errorf("AI 返回空响应")
	}

	content := resp.Choices[0].Message.Content
	tokensUsed := resp.Usage.TotalTokens

	return content, tokensUsed, nil
}
