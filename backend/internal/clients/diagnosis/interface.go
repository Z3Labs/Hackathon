package diagnosis

import "context"

type DiagnosisClient interface {
	// GenerateReport 为指定部署生成诊断报告
	// 返回: 报告内容(JSON字符串)、错误信息
	GenerateReport(ctx context.Context, deploymentId string) (string, error)
}
