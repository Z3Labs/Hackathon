package diagnosis

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/Z3Labs/Hackathon/backend/internal/config"
	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
	"github.com/Z3Labs/Hackathon/backend/internal/types"
)

type diagnosisClient struct {
	ctx         context.Context
	reportModel model.ReportModel
	aiClient    AIClient
}

// New 创建诊断客户端
func New(ctx context.Context, svcCtx *svc.ServiceContext, aiConfig config.AIConfig) DiagnosisClient {
	return &diagnosisClient{
		ctx:         ctx,
		reportModel: svcCtx.ReportModel,
		aiClient:    NewMCPClient(aiConfig),
	}
}

// GenerateReport 生成诊断报告
func (c *diagnosisClient) GenerateReport(req *types.PostAlertCallbackReq) (string, error) {
	if !req.NeedHandle {
		return "", nil
	}
	deploymentId := req.Annotations["deployment_id"]
	// 1. 构建提示词
	prompt := buildPromptTemplate(req)

	// 2. 调用 AI 接口（通过 MCP 查询指标并生成诊断报告）
	reportContent, tokensUsed, err := c.aiClient.GenerateCompletion(c.ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("AI 调用失败: %w", err)
	}

	// 3. 保存报告到数据库
	report := &model.Report{
		DeploymentId: deploymentId,
		Content:      reportContent, // 直接存储文本报告
		CreatedTime:  time.Now(),
		UpdatedTime:  time.Now(),
	}

	if err := c.reportModel.Insert(c.ctx, report); err != nil {
		return "", fmt.Errorf("保存报告失败: %w", err)
	}

	logx.Infof("部署 %s 诊断报告生成成功，Token 消耗: %d", deploymentId, tokensUsed)

	return reportContent, nil
}
