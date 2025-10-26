# MockServer 设计文档

MockServer 仓库地址：https://github.com/Z3Labs/MockServer

## 1. 核心设计

MockServer 是一个**故障注入服务**，通过 HTTP API 模拟各种系统故障，用于测试 AI 诊断系统的能力。

### 1.1 设计精髓

```bash
# 一条 curl 命令触发多类型故障，持续 60 秒
curl -X POST http://mock-server:8080/faults \
  -H "Content-Type: application/json" \
  -d '{
    "faults": [
      {"type": "cpu", "severity": 90, "duration": 60},
      {"type": "memory", "severity": 80, "duration": 60},
      {"type": "network", "latency": 500, "duration": 60}
    ]
  }'
```

### 1.2 核心能力

✅ **curl 触发**：无需登录，一条命令即可模拟故障  
✅ **多类型并发**：同时触发 CPU、内存、网络、磁盘等多种故障  
✅ **持续时间控制**：精确设定故障持续时间（秒级）  
✅ **立即生效**：故障实时影响 Prometheus 指标  

---

## 2. API 设计

### 2.1 触发故障

```bash
POST /api/v1/faults
```

**请求示例**：

```bash
curl -X POST http://localhost:8080/api/v1/faults \
  -H "Content-Type: application/json" \
  -d '{
    "faults": [
      {
        "type": "cpu",
        "severity": 90,
        "duration": 30
      },
      {
        "type": "memory",
        "severity": 85,
        "duration": 30
      },
      {
        "type": "network",
        "latency": 1000,
        "loss": 10,
        "duration": 30
      }
    ]
  }'
```

**响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "faultId": "fault-20250124153045",
    "status": "active",
    "expiresAt": "2025-01-24T15:31:15Z"
  }
}
```

### 2.2 查询故障状态

```bash
GET /api/v1/faults/{faultId}
```

### 2.3 取消故障

```bash
DELETE /api/v1/faults/{faultId}
```

---

## 3. 支持的故障类型

### 3.1 CPU 高负载

```json
{
  "type": "cpu",
  "severity": 90,
  "duration": 60
}
```

- `severity`: CPU 使用率（0-100）
- `duration`: 持续时间（秒）

### 3.2 内存泄漏

```json
{
  "type": "memory",
  "severity": 80,
  "duration": 120
}
```

- `severity`: 内存使用率（0-100）
- `duration`: 持续时间（秒）

### 3.3 网络延迟/丢包

```json
{
  "type": "network",
  "latency": 500,
  "loss": 10,
  "duration": 60
}
```

- `latency`: 延迟毫秒数
- `loss`: 丢包率（0-100）
- `duration`: 持续时间（秒）

### 3.4 磁盘 IO 阻塞

```json
{
  "type": "disk",
  "severity": 95,
  "duration": 60
}
```

- `severity`: IO 使用率（0-100）
- `duration`: 持续时间（秒）

### 3.5 同时触发多种故障

```json
{
  "faults": [
    {"type": "cpu", "severity": 90, "duration": 60},
    {"type": "memory", "severity": 80, "duration": 60},
    {"type": "network", "latency": 500, "duration": 60},
    {"type": "disk", "severity": 85, "duration": 60}
  ]
}
```

---

## 4. 技术实现

### 4.1 架构

```
┌─────────────────────────────────────────┐
│         MockServer (独立服务)              │
│  Port: 8080                             │
│                                         │
│  ┌───────────────────────────────────┐ │
│  │      HTTP API 接口层                │ │
│  │  POST /api/v1/faults               │ │
│  │  GET  /api/v1/faults/{id}          │ │
│  │  DELETE /api/v1/faults/{id}        │ │
│  └───────────────────────────────────┘ │
│                 │                       │
│  ┌───────────────────────────────────┐ │
│  │      故障管理器                      │ │
│  │  - 故障调度                          │ │
│  │  - 定时器管理                        │ │
│  │  - 并发控制                          │ │
│  └───────────────────────────────────┘ │
│                 │                       │
│  ┌──────────────┴──────────────┐       │
│  │  CPU  │  Memory │  Network  │ ...  │
│  │  故障引擎    │     故障引擎    │     故障引擎  │      │
│  └──────────────────────────────┘       │
│                 │                       │
└─────────────────┼────────────────────┘
                 │ 影响 Node Exporter 指标
                 ▼
      ┌──────────────────────┐
      │   Prometheus         │
      │   (监控系统)          │
      └──────────────────────┘
```

### 4.2 故障引擎实现

每个故障类型有独立的**故障引擎**，通过 goroutine 运行：

```go
type FaultEngine interface {
    // 启动故障
    Start(ctx context.Context, config FaultConfig) error
    
    // 停止故障
    Stop(ctx context.Context) error
    
    // 获取当前状态
    Status() FaultStatus
}

// CPU 故障引擎示例
type CPUFaultEngine struct {
    // 通过消耗 CPU 资源实现
    workers []chan struct{}
}

func (e *CPUFaultEngine) Start(ctx context.Context, config CPUFaultConfig) error {
    // 启动 CPU 密集型 goroutine
    go e.cpuStress(config.Severity, config.Duration)
}
```

### 4.3 故障调度

```go
type FaultManager struct {
    activeFaults map[string]*FaultContext
    mu           sync.RWMutex
}

func (m *FaultManager) TriggerFaults(faults []FaultConfig) (string, error) {
    faultId := generateId()
    
    // 并发启动所有故障
    for _, fault := range faults {
        engine := m.getEngine(fault.Type)
        go engine.Start(ctx, fault)
    }
    
    // 定时器：duration 后自动停止
    go m.scheduleStop(faultId, maxDuration(faults))
    
    return faultId, nil
}
```

---

## 5. 使用场景

### 5.1 AI 诊断测试

```bash
# 1. 触发复杂故障（CPU + 内存 + 网络）
curl -X POST http://localhost:8080/api/v1/faults \
  -d '{"faults": [
    {"type": "cpu", "severity": 95, "duration": 60},
    {"type": "memory", "severity": 90, "duration": 60},
    {"type": "network", "latency": 1000, "duration": 60}
  ]}'

# 2. 等待 Prometheus 采集异常指标
sleep 30

# 3. 调用 AI 诊断 API
curl -X POST http://backend:8888/diagnosis/analyze \
  -d '{"deploymentId": "deploy-123"}'

# 3. 验证 AI 能否正确识别多类型故障
```

### 5.2 Hackathon 演示

快速触发故障，展示 AI 诊断能力：

```bash
# 演示场景 1: CPU 高负载
curl -X POST http://localhost:8080/api/v1/faults \
  -d '{"faults": [{"type": "cpu", "severity": 95, "duration": 60}]}'

# 60 秒后自动恢复，可再次触发
```

### 5.3 自动化测试

```go
func TestAIDiagnosis_MultipleFaults(t *testing.T) {
    // 触发多个故障
    resp := mockClient.CreateFault(FaultConfig{
        Faults: []Fault{
            {Type: "cpu", Severity: 90, Duration: 60},
            {Type: "memory", Severity: 80, Duration: 60},
        },
    })
    
    // 等待指标采集
    time.Sleep(30 * time.Second)
    
    // 验证 AI 诊断
    report := diagnoseClient.Analyze("deploy-123")
    assert.Contains(report.Summary, "CPU")
    assert.Contains(report.Summary, "Memory")
}
```

---

## 6. 与系统集成

### 6.1 在部署流程中的位置

```
部署服务
    ↓ 创建部署
MockServer (故障注入)
    ↓ 触发故障
Node Exporter (指标采集)
    ↓ 指标上报
Prometheus (监控)
    ↓ 告警触发
AI 诊断系统
    ↓ 分析故障
生成诊断报告
```

### 6.2 配置示例

```yaml
# hackathon-api.yaml
mockserver:
  enabled: true
  host: "mock-server"
  port: 8080
  defaultDuration: 60
  
faults:
  cpu:
    enabled: true
  memory:
    enabled: true
  network:
    enabled: true
  disk:
    enabled: true
```

---

## 7. 快速开始

### 7.1 启动 MockServer

```bash
# 使用 Docker
docker run -d -p 8080:8080 z3labs/mockserver:latest

# 或本地启动
go run cmd/mockserver/main.go -port 8080
```

### 7.2 触发故障示例

```bash
# 触发 CPU 高负载，持续 60 秒
curl -X POST http://localhost:8080/api/v1/faults \
  -H "Content-Type: application/json" \
  -d '{"faults": [{"type": "cpu", "severity": 95, "duration": 60}]}'

# 触发网络延迟 + 丢包
curl -X POST http://localhost:8080/api/v1/faults \
  -d '{"faults": [{"type": "network", "latency": 500, "loss": 20, "duration": 60}]}'

# 同时触发多种故障
curl -X POST http://localhost:8080/api/v1/faults \
  -d '{"faults": [
    {"type": "cpu", "severity": 90, "duration": 60},
    {"type": "memory", "severity": 85, "duration": 60}
  ]}'
```

---

## 总结

MockServer 是一个**极简的故障注入服务**，核心设计精髓：

1. **curl 即可触发**：无需复杂配置，一条命令模拟故障
2. **多类型并发**：同时触发 CPU、内存、网络等多种故障
3. **持续时间控制**：精确控制故障持续时间
4. **用于 AI 测试**：验证 AI 诊断系统能否识别和分析故障

适用于 Hackathon 演示、AI 诊断测试、故障演练等场景。

