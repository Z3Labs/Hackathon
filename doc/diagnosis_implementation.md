# 智能部署诊断系统 - 技术实现方案

## 1. 系统概述

### 1.1 核心目标
为部署系统添加 AI 智能诊断能力，通过分析 Node Exporter 指标数据，自动生成诊断报告。

### 1.2 核心流程
```
接收告警 → 构建提示词 → Docker 容器 → Python MCP → AI 多轮对话 → 生成报告
                                    ↓                ↓
                              启动 Python 环境    调用 Prometheus/GitHub MCP
                                    ↓                ↓
                              执行诊断脚本        实时查询指标/代码
```

### 1.3 技术栈
- **后端框架**: go-zero
- **数据库**: MongoDB
- **AI 服务**: Claude API / OpenAI API / 通义千问（可配置，通过 MCP）
- **MCP 工具**: Prometheus MCP（指标查询）+ GitHub MCP（代码分析，可选）
- **运行环境**: Docker 容器（Python 3.12 + MCP Server）
- **指标来源**: Prometheus（实时查询，无需存储）

---

## 2. 数据模型设计

### 2.1 核心数据结构

**重要变更**：当前实现基于 MCP 架构，**不再存储指标数据**，指标由 Prometheus MCP 实时查询。

#### Deployment 模型（保持原样）
```go
type Deployment struct {
    Id          string    `bson:"_id,omitempty" json:"id,omitempty"`
    // TODO: 添加业务字段
    CreatedTime time.Time `bson:"createdTime" json:"createdTime"`
    UpdatedTime time.Time `bson:"updatedTime" json:"updatedTime"`
}
```

#### ~~Metric 模型~~（已废弃）

**MCP 架构下不需要此模型**，原因：
- 指标数据由 Prometheus MCP 实时查询，无需存储
- AI 通过 MCP 工具直接调用 Prometheus API
- 避免数据冗余和同步问题

#### Report 模型（独立表，存储诊断报告）
```go
type Report struct {
    Id           string       `bson:"_id,omitempty" json:"id,omitempty"`
    DeploymentId string       `bson:"deploymentId"  json:"deploymentId"` // 关联的部署ID
    Content      string       `bson:"content"       json:"content"`      // AI 生成的报告（纯文本）
    Status       ReportStatus `bson:"status"        json:"status"`       // 报告生成状态
    CreatedTime  time.Time    `bson:"createdTime"   json:"createdTime"`
    UpdatedTime  time.Time    `bson:"updatedTime"   json:"updatedTime"`
}

// 报告状态枚举
type ReportStatus string

const (
    ReportStatusGenerating ReportStatus = "generating" // 生成中
    ReportStatusCompleted  ReportStatus = "completed"  // 已完成
    ReportStatusFailed     ReportStatus = "failed"     // 生成失败
)

type ReportModel interface {
    Insert(ctx context.Context, report *Report) error
    FindByDeploymentId(ctx context.Context, deploymentId string) (*Report, error)
    Update(ctx context.Context, report *Report) error
    DeleteByDeploymentId(ctx context.Context, deploymentId string) error
}
```

**状态说明**：
- `generating`: 报告生成中（MCP 多轮对话进行中）
- `completed`: 报告生成成功
- `failed`: 报告生成失败（AI 调用失败或其他错误）

**报告格式（AI 输出格式）**:
当前实现中，报告以 JSON 格式存储，包含以下字段：
```json
{
  "promQL": ["查询语句1", "查询语句2"],
  "content": "【问题概述】\n...\n【根因分析】\n...\n【影响范围】\n...\n【解决方案】\n..."
}
```

其中：
- `promQL`: 字符串数组，包含诊断过程中识别出的异常指标的 Prometheus 查询语句
- `content`: 字符串，包含详细的诊断报告，格式如下：
  ```
  【问题概述】
  简要描述告警反映的问题
  
  【根因分析】
  详细说明问题的根本原因，引用具体的指标数据和分析过程
  
  【影响范围】
  说明问题影响的系统范围和严重程度
  
  【解决方案】
  提供具体的解决步骤和建议
  ```

### 2.2 MongoDB 集合设计

**集合1: Deployment**
```javascript
// 主键索引
{ "_id": 1 }

// 查询优化索引（根据业务需求添加）
{ "createdTime": -1 }
```

**集合2: Reports**
```javascript
// 主键索引
{ "_id": 1 }

// 部署ID唯一索引（一个部署只有一份报告）
{ "deploymentId": 1 } unique

// 时间索引
{ "createdTime": -1 }

// 状态索引（用于查询生成中/失败的报告）
{ "status": 1, "updatedTime": -1 }
```

**注意**：Metrics 集合已废弃，不再需要创建。

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
    BaseURL        string `json:",optional"` // API 基础 URL，从环境变量读取
    APIKey         string                    // API 密钥，从环境变量读取
    Model          string `json:",default=gpt-4"`                               // 模型名称
    Timeout        int    `json:",default=30"`                                  // 超时时间（秒）
    PrometheusURL  string `json:",optional"`                                    // Prometheus URL（MCP 模式需要）
    GitHubToken    string `json:",optional"`                                    // GitHub Personal Access Token（可选）
    GitHubToolsets string `json:",default=repos,issues,pull_requests,releases"` // GitHub MCP 工具集
}
```

**重要变更**：
- **新增 `PrometheusURL`**: MCP 模式下必需，用于 Prometheus MCP 连接
- **新增 `GitHubToken`**: 可选，提供后自动启用 GitHub MCP 进行代码分析
- **新增 `GitHubToolsets`**: GitHub MCP 工具集配置（默认包含 repos, issues, pull_requests, releases）
- **移除 `MaxRetries`**: MCP 模式下由 Python 脚本内部处理
- **移除 `PrometheusConfig`**: 统一到 `AIConfig` 中

**AI 服务商支持**：
- `BaseURL`: 支持不同 AI 服务商的端点
  - OpenAI: `https://api.openai.com/v1`
  - Claude: `https://api.anthropic.com/v1`
  - 通义千问: `https://dashscope.aliyuncs.com/compatible-mode/v1`
  - ModelScope: `https://api-inference.modelscope.cn`
- `Model`: 根据不同服务商使用不同模型名称
  - OpenAI: `gpt-4`, `gpt-3.5-turbo`
  - Claude: `claude-3-5-sonnet-20241022`
  - 通义千问: `qwen-max`, `qwen-turbo`
  - ModelScope: `Qwen/Qwen3-Coder-480B-A35B-Instruct`

### 3.2 配置文件

```yaml
# backend/etc/hackathon-api.yaml
Name: hackathon-api
Host: 0.0.0.0
Port: 8888

Mongo:
  URL: mongodb://localhost:27017
  Database: hackathon

AI:
  BaseURL: ${AI_BASE_URL}          # 从环境变量读取
  APIKey: ${AI_API_KEY}            # 从环境变量读取
  Model: claude-3-5-sonnet-20241022
  Timeout: 60                       # MCP 模式需要更长时间
  PrometheusURL: http://localhost:9090  # Prometheus URL（必需）
  GitHubToken: ${GITHUB_TOKEN}     # GitHub Token（可选，提供后自动启用 GitHub MCP）
  GitHubToolsets: repos,issues,pull_requests,releases

Qiniu:
  AccessKey: ${QINIU_ACCESS_KEY}
  SecretKey: ${QINIU_SECRET_KEY}
  Bucket: your-bucket
  DownloadHost: https://your-cdn.com

VM:
  VMUIURL: http://localhost:8428
```

**环境变量示例**：
```bash
# OpenAI
export AI_BASE_URL="https://api.openai.com/v1"
export AI_API_KEY="sk-xxx"

# Claude
export AI_BASE_URL="https://api.anthropic.com/v1"
export AI_API_KEY="sk-ant-xxx"

# 通义千问
export AI_BASE_URL="https://dashscope.aliyuncs.com/compatible-mode/v1"
export AI_API_KEY="sk-xxx"

# ModelScope（免费）
export AI_BASE_URL="https://api-inference.modelscope.cn"
export AI_API_KEY="your-modelscope-api-key"

# GitHub MCP（可选）
export GITHUB_TOKEN="ghp_xxx"  # 提供后自动启用 GitHub 代码分析功能

# 七牛云
export QINIU_ACCESS_KEY="xxx"
export QINIU_SECRET_KEY="xxx"
```

**Docker 容器准备**：
MCP 模式需要 Docker 容器运行 Python 环境和 MCP Server：
```bash
cd backend/internal/clients/diagnosis/py
./build-docker.sh  # 构建镜像
```

---

## 4. DiagnosisClient 实现

### 4.1 接口定义

```go
// backend/internal/clients/diagnosis/interface.go
package diagnosis

import (
    "context"
    "github.com/Z3Labs/Hackathon/backend/internal/types"
)

type DiagnosisClient interface {
    // GenerateReport 为指定部署生成诊断报告
    // 接收告警回调请求，返回报告内容、错误信息
    GenerateReport(req *types.PostAlertCallbackReq) (string, error)
}

type AIClient interface {
    GenerateCompletion(ctx context.Context, prompt string) (response string, tokensUsed int, err error)
}
```

**重要变更**：
- 输入参数从 `deploymentId` 改为 `*types.PostAlertCallbackReq`（告警回调请求）
- 新增 `AIClient` 接口，用于 AI 调用的抽象层

### 4.2 核心实现

```go
// backend/internal/clients/diagnosis/client.go
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
        aiClient:    NewMCPClient(aiConfig),  // 使用 MCP 客户端
        Logger:      logx.WithContext(ctx),
    }
}

// GenerateReport 生成诊断报告
func (c *diagnosisClient) GenerateReport(req *types.PostAlertCallbackReq) (string, error) {
    deploymentId := req.Labels["deploymentId"]

    // 1. 检查是否已存在报告（避免重复生成）
    existingReport, _ := c.reportModel.FindByDeploymentId(c.ctx, deploymentId)
    if existingReport != nil {
        return "", fmt.Errorf("部署 %s 的诊断报告已存在，避免重复生成", deploymentId)
    }

    // 2. 先插入一条状态为"生成中"的记录
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

    // 3. 构建提示词（基于告警信息）
    prompt := buildPromptTemplate(req)

    // 4. 调用 AI 接口（通过 MCP 查询指标并生成诊断报告）
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

    // 5. 更新报告内容和状态为完成
    report.Content = reportContent
    report.Status = model.ReportStatusCompleted
    report.UpdatedTime = time.Now()

    if err := c.reportModel.Update(c.ctx, report); err != nil {
        return "", fmt.Errorf("更新报告失败: %w", err)
    }

    c.Infof("部署 %s 诊断报告生成成功，Token 消耗: %d", deploymentId, tokensUsed)

    return reportContent, nil
}
```

**关键变更**：
- **移除指标查询逻辑**：不再从数据库查询指标，改为基于告警回调信息
- **移除异常检测逻辑**：MCP 模式下，AI 通过 Prometheus MCP 实时查询指标进行分析
- **新增报告状态管理**：支持 `generating`、`completed`、`failed` 三种状态
- **错误处理增强**：AI 调用失败时，更新报告状态为 `failed` 并保存错误信息

### 4.3 MCP 客户端实现

**核心架构**：通过 Docker 容器运行 Python MCP 环境

```go
// backend/internal/clients/diagnosis/mcpclient.go
package diagnosis

import (
    "bytes"
    "context"
    "fmt"
    "os/exec"
    "regexp"
    "strings"
    "time"

    "github.com/zeromicro/go-zero/core/logx"
    "github.com/Z3Labs/Hackathon/backend/internal/config"
)

const (
    diagnosisContainerName = "diagnosis-service"
    diagnosisImageName     = "diagnosis-service:latest"
    pyReturnSplit          = "#####"  // Python 脚本输出分隔符
)

var jsonRegex = regexp.MustCompile(`\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}`)

type mcpClient struct {
    containerName  string        // Docker 容器名称
    scriptPath     string        // 容器内 Python 脚本路径
    apiKey         string        // AI API Key
    baseURL        string        // AI Base URL
    model          string        // AI 模型名称
    prometheusURL  string        // Prometheus URL
    githubToken    string        // GitHub Token（可选）
    githubToolsets string        // GitHub MCP 工具集
    timeout        time.Duration // 超时时间
    logger         logx.Logger   // 日志记录器
}

// NewMCPClient 创建 MCP AI 客户端
func NewMCPClient(cfg config.AIConfig) AIClient {
    client := &mcpClient{
        containerName:  diagnosisContainerName,
        scriptPath:     "/app/diagnosis_runner.py",
        apiKey:         cfg.APIKey,
        baseURL:        cfg.BaseURL,
        model:          cfg.Model,
        prometheusURL:  cfg.PrometheusURL,
        githubToken:    cfg.GitHubToken,
        githubToolsets: cfg.GitHubToolsets,
        timeout:        time.Duration(cfg.Timeout) * time.Second,
        logger:         logx.WithContext(context.Background()),
    }

    // 启动容器（如果未运行）
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := client.ensureContainer(ctx); err != nil {
        client.logger.Errorf("启动诊断服务容器失败: %v", err)
    }

    return client
}

func (c *mcpClient) GenerateCompletion(ctx context.Context, prompt string) (string, int, error) {
    // 设置超时（MCP 调用需要更长时间）
    ctx, cancel := context.WithTimeout(ctx, c.timeout*2)
    defer cancel()

    // 确保容器运行
    if err := c.ensureContainer(ctx); err != nil {
        return "", 0, fmt.Errorf("确保容器运行失败: %w", err)
    }

    // 构建 docker exec 命令参数
    args := []string{
        "exec", "-i",
        c.containerName,
        "python", c.scriptPath,
        "--prompt", prompt,
        "--api-key", c.apiKey,
        "--base-url", c.baseURL,
        "--model", c.model,
        "--prometheus-url", c.prometheusURL,
    }

    // 添加 GitHub MCP 参数（如果配置）
    if c.githubToken != "" {
        args = append(args, "--github-token", c.githubToken)
        if c.githubToolsets != "" {
            args = append(args, "--github-toolsets", c.githubToolsets)
        }
    }

    cmd := exec.CommandContext(ctx, "docker", args...)

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    // 执行命令
    err := cmd.Run()
    if err != nil {
        return "", 0, fmt.Errorf("生成分析报告失败: %w\nstdout: %s\nstderr: %s", 
            err, stdout.String(), stderr.String())
    }

    // 解析输出
    result := strings.TrimSpace(stdout.String())
    if result == "" {
        return "", 0, fmt.Errorf("Python 脚本返回空结果")
    }

    split := strings.Split(result, pyReturnSplit)
    if len(split) > 1 {
        c.logger.Infof("Python 脚本执行成功，日志:\n%s", split[0])
        returnValue := strings.TrimSpace(strings.Join(split[1:], pyReturnSplit))
        
        // 提取 JSON
        findString := jsonRegex.FindString(returnValue)
        if findString != "" {
            return findString, 0, nil  // MCP 模式下无法获取准确的 token 数
        }
        return "", 0, fmt.Errorf(returnValue)
    }
    return "", 0, fmt.Errorf(result)
}

// ensureContainer 确保容器运行
func (c *mcpClient) ensureContainer(ctx context.Context) error {
    if c.isContainerRunning(ctx) {
        return nil
    }
    return c.startContainer(ctx)
}

// 其他容器管理方法（isContainerRunning, startContainer 等）...
```

**关键特性**：
- **Docker 容器化**：隔离 Python 环境，避免依赖冲突
- **自动容器管理**：自动启动和管理 Docker 容器
- **MCP 多工具支持**：同时支持 Prometheus MCP 和 GitHub MCP
- **错误处理**：完善的日志记录和错误传递

---

## 5. 提示词工程

### 5.1 提示词模板（基于告警回调）

**核心变更**：MCP 模式下，提示词基于告警回调信息构建，AI 通过 MCP 工具主动查询指标数据。

```go
// backend/internal/clients/diagnosis/prompt.go
package diagnosis

import (
    "fmt"
    "strings"

    "github.com/Z3Labs/Hackathon/backend/internal/types"
)

func buildPromptTemplate(req *types.PostAlertCallbackReq) string {
    // 构建标签信息
    labelsStr := formatMap(req.Labels)

    // 构建注解信息
    annotationsStr := formatMap(req.Annotations)

    // 提取描述信息
    description := req.Desc
    if desc, ok := req.Annotations["description"]; ok && desc != "" {
        description = desc
    }

    prompt := fmt.Sprintf(`你是一个专业的 DevOps 运维诊断专家，擅长分析系统告警并定位问题根因。

**收到以下告警信息**：

告警类型：%s
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
   - 使用 get_metric_metadata(metric: "metric_name") 获取指标元数据
   - 使用 list_metrics() 列出所有可用指标名称
   - 根据告警信息中的标签 hostname 精准查询相关实例的指标

2. **分析发布失败的根本原因**
   - 结合告警信息和查询到的指标数据
   - 分析指标之间的关联关系
   - 识别异常模式和趋势
   - 定位问题的根本原因

3. %s

4. **输出格式**
   重要：请严格按照以下JSON格式输出!!!，你的输出只有一个json，不要用 markdown代码块 标记或任何额外的文本说明
   
   {
     "promQL": ["查询语句1", "查询语句2"],
     "content": "报告内容"
   }
   
   其中：
   - promQL: 字符串数组，包含你在诊断分析过程中识别出的异常指标的Prometheus查询语句
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
        req.Key,
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
        fmt.Sprintf(github_search_prompt, req.RepoAddress, req.Tag),
    )

    return prompt
}

// GitHub 代码分析提示词（可选）
var github_search_prompt = `根据以上排查信息，
若确定问题的存在，则进一步分析 GitHub 仓库 "%s" 发布 release 中的潜在 bug：

  1. 用 "get_release_by_tag" 获取指定 tag %s 的release，若没有查到相关信息，则使用 "get_latest_release"获取最新一个release，
 然后从 body 中提取该次发布的 PR 编号，若该次发布存在pr，则继续，否则结束分析。

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

```

**关键特性**：
- **MCP 工具指导**：明确指示 AI 使用哪些 MCP 工具查询数据
- **GitHub 代码分析**：可选的 GitHub MCP 集成，分析发布相关的代码变更
- **结构化输出**：要求 AI 返回包含 promQL 和 content 的 JSON 格式
- **灵活性**：AI 可以根据告警信息自主决定查询哪些指标

---

## 6. Python MCP 环境

### 6.1 目录结构

```
backend/internal/clients/diagnosis/py/
├── Dockerfile                    # Docker 镜像定义
├── build-docker.sh              # 构建脚本
├── requirements.txt             # Python 依赖
├── diagnosis_runner.py          # 主入口脚本
├── simple_anthropic_mcp.py      # MCP 集成逻辑
├── .env.template                # 环境变量模板
└── .env                         # 本地环境变量（不提交到 Git）
```

### 6.2 Dockerfile

```dockerfile
FROM python:3.12-slim

WORKDIR /app

# 安装 Node.js（GitHub MCP 需要）
RUN apt-get update && apt-get install -y nodejs npm curl && rm -rf /var/lib/apt/lists/*

# 安装 Python 依赖
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# 安装 GitHub MCP Server
RUN npm install -g @modelcontextprotocol/server-github

# 复制脚本
COPY . .

# 保持容器运行
CMD ["tail", "-f", "/dev/null"]
```

### 6.3 核心依赖

```txt
# requirements.txt
anthropic>=0.40.0
mcp>=1.3.2
prometheus-mcp-server>=0.3.0
python-dotenv>=1.0.0
```

---

---

## 7. 实施步骤

### 阶段 1: Docker 环境搭建（已完成）
1. ✅ 创建 Python MCP 环境目录结构
2. ✅ 编写 Dockerfile 和 requirements.txt
3. ✅ 实现 diagnosis_runner.py 入口脚本
4. ✅ 实现 simple_anthropic_mcp.py MCP 集成
5. ✅ 构建 Docker 镜像

### 阶段 2: Go 后端集成（已完成）
6. ✅ 扩展 Config 添加 AI 配置（PrometheusURL、GitHubToken）
7. ✅ 实现 Report Model（新增 Status 字段）
8. ✅ 实现 MCP 客户端（mcpclient.go）
9. ✅ 实现 DiagnosisClient 核心逻辑
10. ✅ 实现提示词构建（基于告警回调）

### 阶段 3: 测试与优化（进行中）
11. ✅ 端到端测试（告警 → MCP → 报告生成）
12. 🔲 性能优化（容器启动时间、MCP 调用超时）
13. 🔲 错误处理完善（网络失败、AI 超时）
14. 🔲 提示词优化（提高报告质量）

### 阶段 4: 功能扩展（可选）
15. 🔲 支持更多 AI 服务商（OpenAI、通义千问）
16. 🔲 添加 MCP 工具缓存机制
17. 🔲 实现报告历史版本管理
18. 🔲 添加诊断报告评分功能

---

## 8. 关键技术细节

### 8.1 MCP 架构核心流程

**MCP 调用链路**：
```
Go Backend → Docker Exec → Python Container → MCP Client → AI + Tools
     ↓                            ↓                           ↓
  构建提示词              启动 MCP Sessions         多轮对话查询指标
     ↓                            ↓                           ↓
  传递参数                  Prometheus MCP              生成诊断报告
                              GitHub MCP (可选)
```

**关键特性**：
- **多轮对话**：AI 可以多次调用 MCP 工具，逐步收集数据
- **工具路由**：同时支持 Prometheus 和 GitHub 两个 MCP Server
- **实时查询**：无需预存指标数据，按需实时查询

### 8.2 Docker 容器管理

**容器生命周期管理**：
```go
// 1. 检查容器是否运行
if c.isContainerRunning(ctx) {
    return nil
}

// 2. 尝试启动已存在的容器
if c.tryStartExistingContainer(ctx) {
    return nil
}

// 3. 创建新容器
docker run -d --name diagnosis-service --restart unless-stopped diagnosis-service:latest
```

**重启策略**：
- `unless-stopped`：自动重启，除非手动停止
- 系统重启后自动恢复运行
- 确保服务高可用性

### 8.3 错误处理与状态管理

**报告状态流转**：
```
generating (生成中) → completed (成功) / failed (失败)
       ↓                    ↓              ↓
  先插入记录          更新内容         保存错误信息
```

**错误处理策略**：
- AI 调用失败：更新报告状态为 `failed`，保存错误信息
- 容器启动失败：记录日志，返回错误给调用方
- 超时处理：MCP 调用设置 2 倍超时时间（默认 120 秒）

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
