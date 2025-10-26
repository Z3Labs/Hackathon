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
	logx.Logger
}

// New 创建诊断客户端
func New(ctx context.Context, svcCtx *svc.ServiceContext, aiConfig config.AIConfig) DiagnosisClient {
	return &diagnosisClient{
		ctx:         ctx,
		reportModel: svcCtx.ReportModel,
		aiClient:    NewMCPClient(aiConfig),
		Logger:      logx.WithContext(ctx),
	}
}

// GenerateReport 生成诊断报告
func (c *diagnosisClient) GenerateReport(req *types.PostAlertCallbackReq) (string, error) {
	if !req.NeedHandle {
		return "", nil
	}
	deploymentId := req.Labels["deploymentId"]

	// TODO 锁，待优化
	deploy, _ := c.reportModel.FindByDeploymentId(c.ctx, deploymentId)
	if deploy != nil {
		return "", fmt.Errorf("部署 %s 的诊断报告已存在，避免重复生成", deploymentId)
	}

	// 1. 先插入一条状态为"生成中"的记录
	report := &model.Report{
		DeploymentId: deploymentId,
		Content:      "",
		Status:       model.ReportStatusGenerating,
		CreatedTime:  time.Now(),
		UpdatedTime:  time.Now(),
	}

	if err := c.reportModel.Insert(c.ctx, report); err != nil {
		return "", fmt.Errorf("创建报告记录失败: %w", err)
	}

	// 2. 构建提示词
	prompt := buildPromptTemplate(req)

	// 3. 调用 AI 接口（通过 MCP 查询指标并生成诊断报告）
	reportContent, tokensUsed, err := c.aiClient.GenerateCompletion(c.ctx, prompt)
	if err != nil {
		// AI 调用失败，更新状态为失败
		report.Status = model.ReportStatusFailed
		report.Content = err.Error()
		report.UpdatedTime = time.Now()
		if updateErr := c.reportModel.Update(c.ctx, report); updateErr != nil {
			c.Errorf("更新报告状态失败: %v", updateErr)
		}
		return "", fmt.Errorf("AI 调用失败: %w", err)
	}

	// 4. 更新报告内容和状态为完成
	report.Content = reportContent
	report.Status = model.ReportStatusCompleted
	report.UpdatedTime = time.Now()

	if err := c.reportModel.Update(c.ctx, report); err != nil {
		return "", fmt.Errorf("更新报告失败: %w", err)
	}

	c.Infof("部署 %s 诊断报告生成成功，Token 消耗: %d", deploymentId, tokensUsed)

	return reportContent, nil
}
