package diagnosis

import (
	"context"

	"github.com/Z3Labs/Hackathon/backend/internal/types"
)

type DiagnosisClient interface {
	// GenerateReport 为指定部署生成诊断报告
	// 返回: 报告内容(JSON字符串)、错误信息
	GenerateReport(req *types.PostAlertCallbackReq) (string, error)
}

type AIClient interface {
	GenerateCompletion(ctx context.Context, prompt string) (response string, tokensUsed int, err error)
}
