package diagnosis

import (
	"fmt"
	"strings"

	"github.com/Z3Labs/Hackathon/backend/internal/types"
)

// buildPromptTemplate 基于告警信息构建完整的 AI prompt
func buildPromptTemplate(req *types.PostAlertCallbackReq) string {
	// 构建标签信息
	labelsStr := formatMap(req.Labels)

	// 构建注解信息
	annotationsStr := formatMap(req.Annotations)

	// 提取关键的描述信息（通常在 annotations 的 description 字段）
	description := req.Desc
	if desc, ok := req.Annotations["description"]; ok && desc != "" {
		description = desc
	}

	prompt := fmt.Sprintf(`你是一个专业的 DevOps 运维诊断专家，擅长分析系统告警并定位问题根因。

**收到以下告警信息**：

告警名称: %s
告警状态: %s
严重程度: %s
描述信息: %s
触发值: %.2f
开始时间: %s
接收时间: %s
结束时间: %s
告警源: %s
需要处理: %t
紧急程度: %t

**标签信息**:
%s

**注解信息**:
%s

**你的任务**：

1. **使用 Prometheus MCP 工具查询相关指标**
   - 使用 get_targets() 检查 Prometheus 抓取目标状态
   - 使用 execute_query() 查询关键指标（CPU、内存、网络、应用指标等）
   - 使用 execute_range_query() 获取时间范围内的趋势数据
   - 根据告警信息中的标签 instance 精准查询相关实例的指标

2. **分析发布失败的根本原因**
   - 结合告警信息和查询到的指标数据
   - 分析指标之间的关联关系
   - 识别异常模式和趋势
   - 定位问题的根本原因

3. **生成诊断报告**
   请以清晰的文本格式输出诊断报告，包含以下部分：
   
   【问题概述】
   简要描述告警反映的问题
   
   【根因分析】
   详细说明问题的根本原因，引用具体的指标数据和分析过程
   
   【影响范围】
   说明问题影响的系统范围和严重程度
   
   【解决方案】
   提供具体的解决步骤和建议

**重要提示**：
- 请使用 MCP 工具主动查询所需的指标数据，不要等待提供
- 在诊断报告中，只输出分析结果和建议，不需要列出查询到的原始指标数据
- 报告应该简洁明了，便于运维人员快速理解和处理
- 如果某些指标查询失败，请说明并基于现有信息进行分析

现在请开始诊断分析：`,
		req.Alertname,
		req.Status,
		req.Severity,
		description,
		req.Values,
		req.StartsAt,
		req.ReceiveAt,
		req.EndsAt,
		req.GeneratorURL,
		req.NeedHandle,
		req.IsEmergent,
		labelsStr,
		annotationsStr,
	)

	return prompt
}

// formatMap 格式化 map 为易读的字符串
func formatMap(m map[string]string) string {
	if len(m) == 0 {
		return "（无）"
	}

	var lines []string
	for key, value := range m {
		lines = append(lines, fmt.Sprintf("  - %s: %s", key, value))
	}
	return strings.Join(lines, "\n")
}
