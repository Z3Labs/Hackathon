package diagnosis

import (
	"context"
	"testing"

	"github.com/Z3Labs/Hackathon/backend/internal/config"
	"github.com/Z3Labs/Hackathon/backend/internal/types"
)

// TestBuildPromptTemplate 测试 prompt 构建
func TestBuildPromptTemplate(t *testing.T) {
	req := &types.PostAlertCallbackReq{
		Key:       "test-alert-001",
		Status:    "firing",
		Alertname: "HighCPUUsage",
		Severity:  "critical",
		Desc:      "CPU 使用率过高",
		StartsAt:  "2025-01-15T14:23:15Z",
		ReceiveAt: "2025-01-15T14:23:16Z",
		Values:    92.5,
		NeedHandle: true,
		Labels: map[string]string{
			"instance": "localhost:9301",
			"job":      "node_exporter",
		},
		Annotations: map[string]string{
			"description":   "节点 CPU 使用率超过 80% 阈值，当前触发值: 92.5%",
			"deployment_id": "test-deployment-123",
		},
	}

	prompt := buildPromptTemplate(req)

	// 验证 prompt 不为空
	if prompt == "" {
		t.Error("buildPromptTemplate() returned empty prompt")
	}

	// 验证 prompt 包含关键信息
	if !contains(prompt, "HighCPUUsage") {
		t.Error("prompt should contain alert name")
	}
	if !contains(prompt, "localhost:9301") {
		t.Error("prompt should contain instance label")
	}
	if !contains(prompt, "92.5") {
		t.Error("prompt should contain trigger value")
	}

	t.Logf("Generated prompt:\n%s", prompt)
}

// TestMCPClient_GenerateCompletion 集成测试 - 测试完整的 MCP 调用流程
// 注意：这是一个集成测试，需要：
// 1. Docker 环境（运行 MCP Server）
// 2. 网络访问（AI API 和 Prometheus）
// 3. 可能需要较长时间（AI 多轮对话）
//
// 运行方式：go test -v -run TestMCPClient_GenerateCompletion -timeout 5m
func TestMCPClient_GenerateCompletion(t *testing.T) {
	// 如果在 CI 环境，跳过集成测试
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 1. 配置初始化（使用 .env 中的配置）
	cfg := config.AIConfig{
		APIKey:        "c566eaba-81e9-408c-8c4f-a17775560377",
		BaseURL:       "https://api-inference.modelscope.cn",
		Model:         "Qwen/Qwen3-Coder-480B-A35B-Instruct",
		PrometheusURL: "http://150.158.152.112:9300",
		Timeout:       120,
	}

	// 2. 创建 MCP 客户端
	client := NewMCPClient(cfg)

	// 3. Mock 告警数据
	req := &types.PostAlertCallbackReq{
		Key:       "test-alert-001",
		Status:    "firing",
		Alertname: "HighCPUUsage",
		Severity:  "critical",
		Desc:      "CPU 使用率过高",
		StartsAt:  "2025-01-15T14:23:15Z",
		ReceiveAt: "2025-01-15T14:23:16Z",
		EndsAt:    "",
		Values:    92.5,
		GeneratorURL: "http://150.158.152.112:9300/graph?g0.expr=...",
		NeedHandle:   true,
		IsEmergent:   true,
		Labels: map[string]string{
			"instance":  "localhost:9301",
			"job":       "node_exporter",
			"alertname": "HighCPUUsage",
		},
		Annotations: map[string]string{
			"description":   "节点 CPU 使用率超过 80% 阈值，当前触发值: 92.5%",
			"deployment_id": "test-deployment-123",
			"summary":       "CPU 使用率告警",
		},
	}

	// 4. 构建 prompt
	prompt := buildPromptTemplate(req)
	t.Logf("Generated prompt:\n%s\n", prompt)

	// 5. 调用 GenerateCompletion
	ctx := context.Background()
	report, tokensUsed, err := client.GenerateCompletion(ctx, prompt)

	// 6. 验证结果
	if err != nil {
		t.Fatalf("GenerateCompletion() failed: %v", err)
	}

	if report == "" {
		t.Error("GenerateCompletion() returned empty report")
	}

	t.Logf("Tokens used: %d", tokensUsed)
	t.Logf("\n================================================================================")
	t.Logf("诊断报告")
	t.Logf("================================================================================")
	t.Logf("%s", report)
	t.Logf("================================================================================")
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
