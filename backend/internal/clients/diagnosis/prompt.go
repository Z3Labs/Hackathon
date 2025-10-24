package diagnosis

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/model"
)

func buildPromptTemplate(metrics []*model.Metric, anomalies []*model.Metric) string {
	// 构建异常指标描述
	anomalyDesc := buildAnomalyDescription(anomalies)

	// 获取所有指标（用于上下文）
	metricsContext := buildMetricsContext(metrics)

	// 识别异常场景类型
	scenarioType := identifyScenarioType(anomalies)

	prompt := fmt.Sprintf(`你是一个专业的 DevOps 运维诊断专家，擅长分析监控指标并定位系统问题。

**任务**: 分析以下服务的监控指标异常，生成诊断报告。

**异常场景类型**: %s

**异常指标**:
%s

**完整指标上下文**:
%s

**输出格式要求**:
请严格按照以下 JSON 格式输出（不要包含任何其他文字）:
{
  "anomalyIndicators": [
    {
      "metricName": "指标名称",
      "currentValue": 当前数值,
      "baselineValue": 正常基线值（根据经验估算）,
      "deviation": "偏离程度的具体描述"
    }
  ],
  "rootCauseAnalysis": "详细的根因分析（200-300字）。必须：1) 引用具体的指标数值和时间戳；2) 分析多个指标之间的关联关系；3) 给出技术层面的根本原因；4) 使用专业术语。",
  "immediateActions": [
    "立即操作建议1（具体可执行的命令或操作步骤）",
    "立即操作建议2"
  ],
  "longTermOptimization": [
    "长期优化建议1（架构或配置层面的改进）",
    "长期优化建议2"
  ]
}

**分析要点**:
1. 根因分析需引用具体时间戳和数值（如"14:23:15 时 CPU 从 45%%%% 突增至 92.5%%%%"）
2. 分析多个指标的关联性（如 CPU 高 + Goroutine 激增 → 高并发问题）
3. 立即操作建议要具体可执行（包含命令、参数、阈值）
4. 长期优化建议要有技术深度（架构、算法、配置优化）

现在请分析上述数据并生成报告（只输出 JSON，不要其他内容）:`,
		scenarioType,
		anomalyDesc,
		metricsContext,
	)

	return prompt
}

func buildAnomalyDescription(anomalies []*model.Metric) string {
	var lines []string
	for _, m := range anomalies {
		metricName := m.Metric["__name__"]

		// 解析值和时间戳
		if len(m.Value) == 2 {
			timestampFloat, _ := m.Value[0].(float64)
			valueStr, _ := m.Value[1].(string)
			value, _ := strconv.ParseFloat(valueStr, 64)
			timestamp := time.Unix(int64(timestampFloat), 0)

			lines = append(lines, fmt.Sprintf("- %s: %.2f (时间: %s)",
				metricName, value, timestamp.Format("15:04:05")))
		}
	}
	return strings.Join(lines, "\n")
}

func buildMetricsContext(metrics []*model.Metric) string {
	var lines []string
	for _, m := range metrics {
		metricName := m.Metric["__name__"]

		// 解析值
		if len(m.Value) == 2 {
			valueStr, _ := m.Value[1].(string)
			value, _ := strconv.ParseFloat(valueStr, 64)

			lines = append(lines, fmt.Sprintf("- %s: %.2f", metricName, value))
		}
	}
	return strings.Join(lines, "\n")
}

func identifyScenarioType(anomalies []*model.Metric) string {
	// 简单的场景识别逻辑
	hasHighCPU := false
	hasHighMemory := false
	hasHighGoroutines := false

	for _, m := range anomalies {
		metricName := m.Metric["__name__"]

		switch metricName {
		case model.MetricCPUUsage:
			hasHighCPU = true
		case model.MetricMemoryUsage:
			hasHighMemory = true
		case model.MetricGoroutines:
			hasHighGoroutines = true
		}
	}

	if hasHighCPU && hasHighGoroutines {
		return "高并发负载异常"
	} else if hasHighMemory {
		return "内存资源异常"
	} else if hasHighCPU {
		return "CPU 资源异常"
	}

	return "综合资源异常"
}

func extractJSON(response string) string {
	// 提取 JSON 部分（AI 可能返回带说明的文本）
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start >= 0 && end > start {
		return response[start : end+1]
	}

	return response
}
