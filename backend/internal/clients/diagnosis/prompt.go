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
   - 使用 get_metric_metadata(metric: "metric_name") - 获取指标元数据（类型、说明等）
   - 若不知道有哪些指标，可以使用 list_metrics() 列出所有可用指标名称，确保指标输入正确
   - 根据告警信息中的标签 hostname 精准查询相关实例的指标

2. **分析发布失败的根本原因**
   - 结合告警信息和查询到的指标数据
   - 分析指标之间的关联关系
   - 识别异常模式和趋势
   - 定位问题的根本原因

**重要提示**：
- 请使用 MCP 工具主动查询所需的指标数据，不要等待提供
- 重点查询 描述信息 提及到的指标，以及 go-runtime 相关的指标（因为是go服务发布）
- 在诊断报告中，只输出分析结果和建议，不需要列出查询到的原始指标数据
- 报告应该简洁明了，便于运维人员快速理解和处理
- 如果某些指标查询失败，请说明并基于现有信息进行分析

3. %s

4. **输出格式**
   重要：请严格按照以下JSON格式输出!!!，你的输出只有一个json，不要用 markdown代码块 标记或任何额外的文本说明
   
   {
     "promQL": ["查询语句1", "查询语句2"],
     "content": "报告内容"
   }
   
   其中：
   - promQL: 字符串数组，包含你在诊断分析过程中识别出的异常指标的Prometheus查询语句（每个查询语句一个元素）。如果没有发现异常指标，则为空数组[]
   - content: 字符串，包含详细的诊断报告，格式如下：
     
     【问题概述】
     简要描述告警反映的问题
     
     【根因分析】
     详细说明问题的根本原因，引用具体的指标数据和分析过程
     
     【影响范围】
     说明问题影响的系统范围和严重程度
     
     【解决方案】
     提供具体的解决步骤和建议

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
		fmt.Sprintf(github_search_prompt, req.RepoAddress),
	)

	return prompt
}

// 若问题不存在则输出分析结果
var github_search_prompt = `根据以上排查信息，
若确定问题的存在，则进一步分析 GitHub 仓库 "%s" 最新 release 中的潜在 bug：

  1. 用 "get_latest_release" 获取最新 release，从 body 中提取该次发布的 PR 编号，若该次发布存在pr，则继续，否则结束分析。

  2. 逐个分析 PR，对每个 PR 编号，依次调用以下工具：
  ### 2.1 获取 PR 基本信息
  工具：pull_request_read
  - method: "get"
  - pullNumber: [PR编号]
  ### 2.2 获取代码变更文件
  工具：pull_request_read
  - method: "get_files"
  - pullNumber: [PR编号]
  ### 2.3 获取代码 Diff
  工具：pull_request_read
  - method: "get_diff"
  - pullNumber: [PR编号]

  3. 分析每个 PR 的 diff，查找常见的致命 bug（忽略不会导致发布失败的小问题）：
     - 空指针问题
     - 资源泄漏（未关闭连接、文件句柄）
     - 并发安全
     - 逻辑错误
	 - ...

  4. 若查找到可能的错误，则输出：PR编号 + 文件路径 + 问题描述 + 建议修复`

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
