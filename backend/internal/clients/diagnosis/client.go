package diagnosis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/Z3Labs/Hackathon/backend/internal/config"
	"github.com/Z3Labs/Hackathon/backend/internal/model"
	"github.com/Z3Labs/Hackathon/backend/internal/svc"
)

type diagnosisClient struct {
	metricModel model.MetricModel
	reportModel model.ReportModel
	aiClient    AIClient
}

// New 创建诊断客户端
func New(svcCtx *svc.ServiceContext, aiConfig config.AIConfig) DiagnosisClient {
	return &diagnosisClient{
		metricModel: svcCtx.MetricModel,
		reportModel: svcCtx.ReportModel,
		aiClient:    NewOpenAIClient(aiConfig),
	}
}

// GenerateReport 生成诊断报告
func (c *diagnosisClient) GenerateReport(ctx context.Context, deploymentId string) (string, error) {
	// 1. 查询该部署的指标数据
	metrics, err := c.metricModel.FindByDeploymentId(ctx, deploymentId)
	if err != nil {
		return "", fmt.Errorf("查询指标数据失败: %w", err)
	}

	if len(metrics) == 0 {
		return "", fmt.Errorf("部署 %s 没有指标数据", deploymentId)
	}

	// 2. 检测异常指标
	anomalies := c.detectAnomalies(metrics)
	if len(anomalies) == 0 {
		logx.Infof("部署 %s 未检测到异常，无需生成报告", deploymentId)
		return "", nil
	}

	// 3. 构建提示词
	prompt := buildPromptTemplate(metrics, anomalies)

	// 4. 调用 AI 接口
	reportContent, tokensUsed, err := c.aiClient.GenerateCompletion(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("AI 调用失败: %w", err)
	}

	// 5. 提取 JSON 内容（AI 可能返回带说明的文本）
	reportJSON := extractJSON(reportContent)

	// 6. 保存报告到数据库
	report := &model.Report{
		DeploymentId: deploymentId,
		Content:      reportJSON, // 直接存储 JSON 字符串
		CreatedTime:  time.Now(),
		UpdatedTime:  time.Now(),
	}

	if err := c.reportModel.Insert(ctx, report); err != nil {
		return "", fmt.Errorf("保存报告失败: %w", err)
	}

	logx.Infof("部署 %s 诊断报告生成成功，Token 消耗: %d", deploymentId, tokensUsed)

	return reportJSON, nil
}

// detectAnomalies 检测异常指标（静态阈值检测）
func (c *diagnosisClient) detectAnomalies(metrics []*model.Metric) []*model.Metric {
	var anomalies []*model.Metric

	// 阈值配置
	thresholds := map[string]float64{
		model.MetricCPUUsage:        80.0,
		model.MetricMemoryUsage:     90.0,
		model.MetricDiskUsage:       90.0,
		model.MetricDiskIOWait:      50.0,
		model.MetricGoroutines:      10000,
		model.MetricGCPauseDuration: 100.0,
	}

	for _, m := range metrics {
		// 从 Metric map 中提取指标名称
		metricName := m.Metric["__name__"]

		// 从 Value 数组中解析指标值
		if len(m.Value) != 2 {
			continue
		}
		valueStr, ok := m.Value[1].(string)
		if !ok {
			continue
		}
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			continue
		}

		// 检查是否超过阈值
		if threshold, exists := thresholds[metricName]; exists {
			if value > threshold {
				anomalies = append(anomalies, m)
			}
		}
	}

	return anomalies
}
