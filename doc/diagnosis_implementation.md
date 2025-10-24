# 智能部署诊断系统 - 技术实现方案

## 1. 系统概述

### 1.1 核心目标
为部署系统添加 AI 智能诊断能力，通过分析 Node Exporter 指标数据，自动生成诊断报告。

### 1.2 核心流程
```
部署ID → 查询指标数据 → AI 分析 → 生成报告 → 保存映射关系
```

### 1.3 技术栈
- **后端框架**: go-zero
- **数据库**: MongoDB
- **AI 服务**: Claude API / OpenAI API / 通义千问（可配置）
- **指标来源**: Node Exporter（Prometheus 生态）

---

## 2. 数据模型设计

### 2.1 核心数据结构

#### Deployment 模型（保持原样）
```go
type Deployment struct {
    Id          string    `bson:"_id,omitempty" json:"id,omitempty"`
    // TODO: 添加业务字段
    CreatedTime time.Time `bson:"createdTime" json:"createdTime"`
    UpdatedTime time.Time `bson:"updatedTime" json:"updatedTime"`
}
```

#### Metric 模型（独立表）

**设计说明**：直接存储 Prometheus HTTP API 返回的原始 JSON 格式，无需任何转换。

```go
type Metric struct {
    Id           string            `bson:"_id,omitempty" json:"id,omitempty"`
    DeploymentId string            `bson:"deploymentId" json:"deploymentId"` // 关联的部署ID

    // 直接对应 Prometheus 即时查询返回格式
    Metric       map[string]string `bson:"metric" json:"metric"`   // 包含 __name__ 和所有标签
    Value        []interface{}     `bson:"value" json:"value"`     // [timestamp(float64), "value"(string)]

    CreatedTime  time.Time         `bson:"createdTime" json:"createdTime"`
}

type MetricModel interface {
    Insert(ctx context.Context, metric *Metric) error
    FindByDeploymentId(ctx context.Context, deploymentId string) ([]*Metric, error)
    DeleteByDeploymentId(ctx context.Context, deploymentId string) error
}
```

**Prometheus 原始返回格式示例**：
```json
{
  "status": "success",
  "data": {
    "resultType": "vector",
    "result": [
      {
        "metric": {
          "__name__": "node_cpu_seconds_total",
          "cpu": "0",
          "mode": "idle",
          "instance": "localhost:9100",
          "job": "node-exporter"
        },
        "value": [1435781451.781, "12345.67"]
      }
    ]
  }
}
```

**MongoDB 存储格式示例**：
```json
{
  "_id": "ObjectId(...)",
  "deploymentId": "deploy-123",
  "metric": {
    "__name__": "node_cpu_seconds_total",
    "cpu": "0",
    "mode": "idle",
    "instance": "localhost:9100",
    "job": "node-exporter"
  },
  "value": [1435781451.781, "12345.67"],
  "createdTime": "2025-01-24T10:00:00Z"
}
```

**字段说明**：
- `Metric`: 标签集合，包含：
  - `__name__`: 指标名称（如 `node_cpu_seconds_total`）
  - 其他标签: 如 `cpu`, `mode`, `instance`, `job` 等
- `Value`: 数组格式 `[timestamp, value]`
  - `value[0]`: float64 类型的 Unix 时间戳（秒）
  - `value[1]`: string 类型的指标值（Prometheus 使用字符串避免 JSON 无法表示 NaN/Inf）

**数据解析示例**：
```go
// 从数据库读取 Metric
metric, _ := metricModel.FindById(ctx, id)

// 解析指标名称
metricName := metric.Metric["__name__"]  // "node_cpu_seconds_total"

// 解析其他标签
cpu := metric.Metric["cpu"]              // "0"
mode := metric.Metric["mode"]            // "idle"

// 解析时间戳
timestampFloat, _ := metric.Value[0].(float64)
timestamp := time.Unix(int64(timestampFloat), 0)

// 解析指标值
valueStr, _ := metric.Value[1].(string)
value, _ := strconv.ParseFloat(valueStr, 64)  // 12345.67

fmt.Printf("指标: %s, 值: %.2f, 时间: %s\n",
    metricName, value, timestamp.Format("2006-01-02 15:04:05"))
```

**常见指标名称常量**（用于异常检测）：
```go
const (
    MetricCPUUsage           = "node_cpu_seconds_total"
    MetricLoadAvg            = "node_load1"
    MetricMemoryUsage        = "node_memory_MemAvailable_bytes"
    MetricDiskUsage          = "node_filesystem_avail_bytes"
    MetricDiskIOWait         = "node_disk_io_time_seconds_total"
    MetricGoroutines         = "go_goroutines"
    MetricGCPauseDuration    = "go_gc_duration_seconds"
    MetricNetworkReceive     = "node_network_receive_bytes_total"
    MetricNetworkTransmit    = "node_network_transmit_bytes_total"
)
```

#### Report 模型（独立表，存储为 JSON 字符串）
```go
type Report struct {
    Id           string    `bson:"_id,omitempty" json:"id,omitempty"`
    DeploymentId string    `bson:"deploymentId" json:"deploymentId"`     // 关联的部署ID
    Content      string    `bson:"content" json:"content"`               // AI 生成的报告（JSON 字符串）
    AIModel      string    `bson:"aiModel" json:"aiModel"`               // 使用的 AI 模型
    TokensUsed   int       `bson:"tokensUsed" json:"tokensUsed"`         // Token 消耗
    CreatedTime  time.Time `bson:"createdTime" json:"createdTime"`
    UpdatedTime  time.Time `bson:"updatedTime" json:"updatedTime"`
}

type ReportModel interface {
    Insert(ctx context.Context, report *Report) error
    FindByDeploymentId(ctx context.Context, deploymentId string) (*Report, error)
    Update(ctx context.Context, report *Report) error
    DeleteByDeploymentId(ctx context.Context, deploymentId string) error
}
```

**报告 JSON 格式（AI 输出格式）**:
```json
{
  "anomalyIndicators": [
    {
      "metricName": "cpu_usage_percent",
      "currentValue": 92.5,
      "baselineValue": 45.0,
      "deviation": "当前值超过基线 105%，达到 92.5%"
    }
  ],
  "rootCauseAnalysis": "详细的根因分析（200-300字）",
  "immediateActions": [
    "立即操作建议1",
    "立即操作建议2"
  ],
  "longTermOptimization": [
    "长期优化建议1",
    "长期优化建议2"
  ]
}
```

### 2.2 MongoDB 集合设计

**集合1: Deployment**
```javascript
// 主键索引
{ "_id": 1 }

// 查询优化索引（根据业务需求添加）
{ "createdTime": -1 }
```

**集合2: Metrics**
```javascript
// 主键索引
{ "_id": 1 }

// 部署ID索引（核心查询字段）
{ "deploymentId": 1, "createdTime": -1 }

// 复合索引（按部署ID和指标名称查询）
// 注意：metric.__name__ 是嵌套字段
{ "deploymentId": 1, "metric.__name__": 1, "createdTime": -1 }

// 可选：按时间戳查询（value[0] 是时间戳）
// 如果需要按指标时间戳排序，可以考虑将 value[0] 单独提取为字段
```

**集合3: Reports**
```javascript
// 主键索引
{ "_id": 1 }

// 部署ID唯一索引（一个部署只有一份报告）
{ "deploymentId": 1 } unique

// 时间索引
{ "createdTime": -1 }
```

---

## 3. 配置设计

### 3.1 Config 结构扩展

```go
// backend/internal/config/config.go
package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
    rest.RestConf

    MongoDB struct {
        URI      string
        Database string
    }

    // 新增：AI 服务配置
    AI AIConfig

    // 新增：Prometheus 配置（可选，如果需要主动抓取指标）
    Prometheus PrometheusConfig `json:",optional"`
}

type AIConfig struct {
    BaseURL    string `json:",optional"`           // API 基础 URL，从环境变量读取
    APIKey     string                              // API 密钥，从环境变量读取
    Model      string `json:",default=gpt-4"`      // 模型名称
    Timeout    int    `json:",default=30"`         // 超时时间（秒）
    MaxRetries int    `json:",default=3"`          // 重试次数
}

type PrometheusConfig struct {
    NodeExporterURL string // Node Exporter 地址，如 http://localhost:9100
}
```

**说明**：
- `BaseURL`: 支持不同 AI 服务商的端点
  - OpenAI: `https://api.openai.com/v1`
  - Claude (via OpenAI-compatible): `https://api.anthropic.com/v1` (需要代理)
  - 通义千问: `https://dashscope.aliyuncs.com/compatible-mode/v1`
  - 本地模型: `http://localhost:11434/v1` (如 Ollama)
- `Model`: 根据不同服务商使用不同模型名称
  - OpenAI: `gpt-4`, `gpt-3.5-turbo`
  - Claude: `claude-3-5-sonnet-20241022`
  - 通义千问: `qwen-max`, `qwen-turbo`

### 3.2 配置文件

```yaml
# backend/etc/hackathon-api.yaml
Name: hackathon-api
Host: 0.0.0.0
Port: 8888

MongoDB:
  URI: mongodb://localhost:27017
  Database: hackathon

AI:
  BaseURL: ${AI_BASE_URL}      # 从环境变量读取，如 https://api.openai.com/v1
  APIKey: ${AI_API_KEY}        # 从环境变量读取
  Model: gpt-4                 # 或 claude-3-5-sonnet-20241022, qwen-max
  Timeout: 30
  MaxRetries: 3

# 可选：如果需要主动抓取指标
Prometheus:
  NodeExporterURL: http://localhost:9100
```

**环境变量示例**：
```bash
# OpenAI
export AI_BASE_URL="https://api.openai.com/v1"
export AI_API_KEY="sk-xxx"

# Claude (通过 OpenAI 兼容端点)
export AI_BASE_URL="https://api.anthropic.com/v1"
export AI_API_KEY="sk-ant-xxx"

# 通义千问
export AI_BASE_URL="https://dashscope.aliyuncs.com/compatible-mode/v1"
export AI_API_KEY="sk-xxx"
```

---

## 4. DiagnosisClient 实现

### 4.1 接口定义

```go
// backend/internal/clients/diagnosis/interface.go
package diagnosis

import "context"

type DiagnosisClient interface {
    // GenerateReport 为指定部署生成诊断报告
    // 返回: 报告内容(JSON字符串)、错误信息
    GenerateReport(ctx context.Context, deploymentId string) (string, error)
}
```

### 4.2 核心实现

```go
// backend/internal/clients/diagnosis/client.go
package diagnosis

import (
    "context"
    "fmt"
    "strconv"
    "time"

    "github.com/Z3Labs/Hackathon/backend/internal/config"
    "github.com/Z3Labs/Hackathon/backend/internal/model"
    "github.com/zeromicro/go-zero/core/logx"
)

type diagnosisClient struct {
    metricModel model.MetricModel
    reportModel model.ReportModel
    aiConfig    config.AIConfig
    aiClient    AIClient
}

// New 创建诊断客户端
func New(svcCtx *svc.ServiceContext) DiagnosisClient {
    return &diagnosisClient{
        metricModel: svcCtx.MetricModel,
        reportModel: svcCtx.ReportModel,
        aiConfig:    svcCtx.Config.AI,
        aiClient:    NewOpenAIClient(svcCtx.Config.AI),
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
        Content:      reportJSON,        // 直接存储 JSON 字符串
        AIModel:      c.aiConfig.Model,
        TokensUsed:   tokensUsed,
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
```

### 4.3 AI 客户端实现（统一使用 OpenAI 格式）

```go
// backend/internal/clients/diagnosis/ai_client.go
package diagnosis

import (
    "context"
    "time"

    "github.com/Z3Labs/Hackathon/backend/internal/config"
    openai "github.com/sashabaranov/go-openai"
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
```

**说明**：
- 使用 `github.com/sashabaranov/go-openai` SDK，这是 OpenAI 官方推荐的 Go SDK
- 通过配置不同的 `BaseURL`，可以兼容多个 AI 服务商：
  - **OpenAI**: 默认 `https://api.openai.com/v1`
  - **Claude**: 通过代理端点（需要第三方代理服务）
  - **通义千问**: `https://dashscope.aliyuncs.com/compatible-mode/v1`
  - **本地模型**: `http://localhost:11434/v1` (如 Ollama)
- 所有服务商使用统一的接口，简化实现

---

## 5. 提示词工程

### 5.1 提示词模板

```go
// backend/internal/clients/diagnosis/prompt.go
package diagnosis

import (
    "encoding/json"
    "fmt"
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
1. 根因分析需引用具体时间戳和数值（如"14:23:15 时 CPU 从 45%% 突增至 92.5%%"）
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
```

### 5.2 Few-shot 示例（可选优化）

在提示词中添加示例可以提高 AI 输出质量：

```go
func addFewShotExample() string {
    return `
**示例场景**: CPU 使用率异常

输入数据:
- cpu_usage_percent: 92.5 (时间: 14:23:15)
- go_goroutines: 8500
- go_gc_pause_duration_ms: 85

期望输出:
{
  "anomalyIndicators": [
    {
      "metricName": "cpu_usage_percent",
      "currentValue": 92.5,
      "baselineValue": 45.0,
      "deviation": "当前值超过基线 105%%，达到 92.5%%"
    }
  ],
  "rootCauseAnalysis": "2025-01-24 14:23:15 时刻，CPU 使用率从正常的 45%% 突增至 92.5%%，持续超过 5 分钟。结合 goroutine 数量从 2000 激增至 8500（增长 325%%），以及 GC 暂停时间从 10ms 上升至 85ms，判断为高并发请求导致的计算资源竞争。系统负载达到 15.8（8核心机器），表明 CPU 调度队列严重堆积。",
  "immediateActions": [
    "执行 'curl http://localhost:6060/debug/pprof/goroutine?debug=1' 查看 Goroutine 堆栈，定位泄漏点",
    "临时限流：在负载均衡层设置 max_conns=1000 或 rate_limit=100r/s",
    "查看最近 10 分钟的访问日志：'tail -n 10000 /var/log/access.log | awk '{print $1}' | sort | uniq -c | sort -rn | head -20'"
  ],
  "longTermOptimization": [
    "优化高频接口的算法复杂度，考虑添加缓存层（Redis）减少重复计算",
    "实现应用层限流中间件（令牌桶算法），设置每个 IP 的 QPS 上限",
    "评估水平扩容方案：当前 8 核 CPU 已接近瓶颈，建议增加 2-3 个实例并配置负载均衡"
  ]
}
`
}
```

---

## 6. ServiceContext 集成

```go
// backend/internal/svc/servicecontext.go
package svc

import (
    "github.com/Z3Labs/Hackathon/backend/internal/clients/diagnosis"
    "github.com/Z3Labs/Hackathon/backend/internal/config"
    "github.com/Z3Labs/Hackathon/backend/internal/model"
)

type ServiceContext struct {
    Config          config.Config
    DeploymentModel model.DeploymentModel
    DiagnosisClient diagnosis.DiagnosisClient
}

func NewServiceContext(c config.Config) *ServiceContext {
    // 初始化 DeploymentModel
    deploymentModel := model.NewDeploymentModel(c.MongoDB.URI, c.MongoDB.Database)

    // 初始化 DiagnosisClient
    diagnosisClient := diagnosis.New(deploymentModel, c.AI)

    return &ServiceContext{
        Config:          c,
        DeploymentModel: deploymentModel,
        DiagnosisClient: diagnosisClient,
    }
}
```

---

---

## 7. 实施步骤

### 阶段 1: 基础架构（1-2天）
1. ✅ 扩展 Config 添加 AI 配置
2. ✅ 扩展 Deployment Model 添加 Metrics 和 Report 字段
3. ✅ 创建 diagnosis 包目录结构
4. ✅ 实现 AI 客户端接口定义

### 阶段 2: 核心功能（2-3天）
5. ✅ 实现 DiagnosisClient 核心逻辑
6. ✅ 实现 Claude API 客户端
7. ✅ 实现提示词构建逻辑
8. ✅ 实现 JSON 响应解析
9. ✅ 集成到 ServiceContext

### 阶段 3: 测试验证（1-2天）
10. ✅ 准备 Mock 指标数据
11. ✅ 单元测试（异常检测、提示词生成）
12. ✅ 集成测试（端到端流程）
13. ✅ 优化提示词模板

### 阶段 4: 扩展优化（可选）
14. 🔲 实现 OpenAI 客户端
15. 🔲 实现通义千问客户端
16. 🔲 添加动态基线检测算法
17. 🔲 添加重试机制和错误处理

---

## 8. 关键技术细节

### 8.1 异常检测算法

**静态阈值检测**（第一阶段实现）:
```go
thresholds := map[string]float64{
    "cpu_usage_percent":        80.0,
    "memory_usage_percent":     90.0,
    "disk_usage_percent":       90.0,
    "disk_io_wait_percent":     50.0,
    "go_goroutines":            10000,
    "go_gc_pause_duration_ms":  100.0,
}
```

**动态基线检测**（可选优化）:
基于历史数据计算 3σ 偏离。

### 8.2 Token 成本控制

```go
// 每次调用前检查配额
func (c *diagnosisClient) checkQuota(ctx context.Context) error {
    // 查询今日已消耗 Token 数
    todayUsage := getTodayTokenUsage(ctx)

    // 设置每日上限（如 100,000 tokens）
    if todayUsage > 100000 {
        return fmt.Errorf("今日 Token 配额已用完")
    }

    return nil
}
```

### 8.3 错误处理与重试

```go
func (c *diagnosisClient) callAIWithRetry(ctx context.Context, prompt string) (string, int, error) {
    var lastErr error

    for i := 0; i < c.aiConfig.MaxRetries; i++ {
        response, tokens, err := c.aiClient.GenerateCompletion(ctx, prompt)
        if err == nil {
            return response, tokens, nil
        }

        lastErr = err
        logx.Errorf("AI 调用失败（第 %d 次重试）: %v", i+1, err)

        // 指数退避
        time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
    }

    return "", 0, fmt.Errorf("AI 调用失败（已重试 %d 次）: %w", c.aiConfig.MaxRetries, lastErr)
}
```

---

## 9. 依赖包

```bash
# OpenAI SDK（兼容多个 AI 服务商）
go get github.com/sashabaranov/go-openai
```

**说明**：
- 只需要一个 SDK，通过配置不同的 BaseURL 支持多个服务商
- 无需额外的 HTTP 客户端，SDK 内置

---

## 10. 配置示例与环境变量

### 10.1 OpenAI

```bash
# 设置环境变量
export AI_BASE_URL="https://api.openai.com/v1"
export AI_API_KEY="sk-xxx"

# 启动服务
cd backend
go run hackathon.go -f etc/hackathon-api.yaml
```

### 10.2 通义千问

```bash
# 设置环境变量
export AI_BASE_URL="https://dashscope.aliyuncs.com/compatible-mode/v1"
export AI_API_KEY="sk-xxx"  # 通义千问的 API KEY

# 修改配置文件中的 Model
# Model: qwen-max

# 启动服务
cd backend
go run hackathon.go -f etc/hackathon-api.yaml
```

### 10.3 Claude (通过代理)

```bash
# 需要第三方 OpenAI-compatible 代理服务
export AI_BASE_URL="https://your-proxy-service.com/v1"
export AI_API_KEY="sk-ant-xxx"

# 修改配置文件中的 Model
# Model: claude-3-5-sonnet-20241022

# 启动服务
cd backend
go run hackathon.go -f etc/hackathon-api.yaml
```

---

## 11. 总结

本实现方案提供了一个简洁高效的 AI 智能诊断系统架构，核心特点：

### 11.1 架构优势
- ✅ **数据模型解耦**：Metrics、Reports 独立存储，易于扩展和查询
- ✅ **简化数据处理**：Report 直接存储 JSON 字符串，前端解析，无需后端结构体映射
- ✅ **统一 AI 接口**：使用 OpenAI SDK，一套代码支持多个 AI 服务商
- ✅ **配置灵活**：BaseURL 和 APIKey 从环境变量读取，方便切换服务商
- ✅ **接口简洁**：DiagnosisClient 只有一个核心方法 `GenerateReport`

### 11.2 核心实现流程

```
1. 查询指标数据 (MetricModel.FindByDeploymentId)
   ↓
2. 异常检测 (静态阈值)
   ↓
3. 构建提示词 (buildPromptTemplate)
   ↓
4. 调用 AI (OpenAI SDK)
   ↓
5. 提取 JSON (extractJSON)
   ↓
6. 保存报告 (ReportModel.Insert) - 直接存储 JSON 字符串
```

### 11.3 关键设计决策

| 设计点 | 方案 | 优势 |
|--------|------|------|
| **数据模型** | 独立的 Metrics 和 Reports 表 | 解耦、易扩展、查询灵活 |
| **Report 存储** | JSON 字符串 | 简化后端逻辑、AI 格式变更无需改代码 |
| **AI 集成** | 统一使用 OpenAI SDK | 一套代码支持多服务商 |
| **配置管理** | 环境变量 + YAML | 安全、灵活、易部署 |
| **Client 初始化** | 传入 ServiceContext | 充分利用 go-zero 依赖注入 |

### 11.4 下一步工作

按照文档的实施步骤，可以快速完成核心功能开发：
1. **阶段 1**：创建 Model (Metric、Report)
2. **阶段 2**：实现 DiagnosisClient
3. **阶段 3**：集成到 ServiceContext
4. **阶段 4**：测试与优化
