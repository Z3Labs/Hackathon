# MCP 智能诊断系统使用指南

## 🎯 功能说明

集成了 Model Context Protocol (MCP)，让 AI 能够**主动查询 Prometheus 指标**进行智能诊断。

### 两种模式对比

| 特性 | OpenAI 兼容模式 | MCP 模式（推荐） |
|------|----------------|-----------------|
| 数据获取 | 预先查询好的指标 | AI 主动查询 Prometheus |
| 分析能力 | 基于固定提示词 | 智能选择查询指标 |
| 响应时间 | 快（~5秒） | 慢（~30-60秒） |
| 准确性 | 受限于提示词 | 更全面、更准确 |
| 依赖 | 无 | Python + Docker |

## 📦 安装依赖（MCP 模式）

```bash
# 1. 进入 Python 目录
cd backend/internal/clients/diagnosis/py

# 2. 创建虚拟环境
python3 -m venv venv
source venv/bin/activate

# 3. 安装依赖
pip install -r requirements.txt
```

## ⚙️ 配置方式

### 方式 1：环境变量（推荐）

```bash
# 复制配置模板
cp backend/.env.example backend/.env

# 编辑 .env 文件
vim backend/.env
```

MCP 模式配置示例：

```bash
AI_BASE_URL=https://api-inference.modelscope.cn
AI_API_KEY=c566eaba-81e9-408c-8c4f-a17775560377
AI_MODEL=Qwen/Qwen3-Coder-480B-A35B-Instruct
AI_USE_MCP=true
PROMETHEUS_URL=http://150.158.152.112:9300
```

### 方式 2：修改 YAML 配置

直接修改 `backend/etc/hackathon-api.yaml`（不推荐，建议使用环境变量）

## 🚀 启动服务

```bash
cd backend

# 加载环境变量并启动
export $(cat .env | xargs) && go run hackathon.go
```

## 📊 工作流程

### OpenAI 兼容模式
```
告警触发 → Go 后端
    ↓
提前查询 Prometheus 指标
    ↓
构建提示词（包含指标数据）
    ↓
调用 AI API
    ↓
生成诊断报告
```

### MCP 模式
```
告警触发 → Go 后端
    ↓
调用 Python MCP Bridge
    ↓
Python 启动 MCP Server
    ↓
AI 主动调用工具：
  - get_targets() → 检查监控目标
  - execute_query() → 查询 CPU/内存/Go runtime
  - execute_range_query() → 获取趋势数据
    ↓
AI 基于实时数据生成报告
    ↓
返回 JSON 结构化报告
```

## 🔍 AI 可以查询的指标

MCP 模式下，AI 可以主动查询以下类型的指标：

### 系统资源
- CPU 使用率：`node_cpu_seconds_total`
- 内存使用：`node_memory_*`
- 磁盘 I/O：`node_disk_*`
- 网络流量：`node_network_*`

### Go Runtime
- Goroutines：`go_goroutines`
- 堆内存：`go_memstats_heap_*`
- GC 统计：`go_gc_duration_seconds`

### 应用指标
- HTTP 请求：`http_requests_total`
- 错误率：`http_requests_total{status=~"5.."}`
- 进程信息：`process_*`

## 📝 诊断报告格式

```json
{
  "anomaly_metrics": [
    {
      "metric_name": "process_cpu_seconds_total",
      "current_value": "0.0002333",
      "threshold": "0.80",
      "severity": "warning",
      "description": "CPU 使用率正常"
    }
  ],
  "root_cause": "未发现明显的资源瓶颈",
  "details": "系统资源使用正常...",
  "recommendations": [
    "检查应用程序日志",
    "确认依赖服务状态"
  ]
}
```

## 🐛 故障排查

### 1. Python 脚本执行失败

```bash
# 检查 Python 环境
which python3
python3 --version

# 检查脚本权限
chmod +x backend/internal/clients/diagnosis/py/diagnosis_runner.py

# 手动测试
cd backend/internal/clients/diagnosis/py
python3 diagnosis_runner.py --help
```

### 2. MCP Server 连接失败

```bash
# 检查 Docker
docker --version
docker ps

# 测试 MCP Server
docker run -i --rm \
  -e PROMETHEUS_URL=http://150.158.152.112:9300 \
  ghcr.io/pab1it0/prometheus-mcp-server:latest
```

### 3. Prometheus 无法访问

```bash
# 测试 Prometheus 连接
curl http://150.158.152.112:9300/api/v1/query?query=up
```

### 4. 查看详细日志

Go 服务日志会显示：
- `使用 MCP 模式进行智能诊断` - 启用 MCP
- `使用 OpenAI 兼容模式进行智能诊断` - 使用传统模式

## 💡 性能优化建议

1. **调整超时时间**：MCP 模式需要更长时间，建议设置 `Timeout: 120`
2. **并发控制**：避免同时触发大量 MCP 诊断请求
3. **缓存结果**：相同告警可以复用最近的诊断结果
4. **指标预热**：提前拉取 Docker 镜像

## 🔐 安全建议

1. 不要在代码中硬编码 API Key
2. 使用环境变量或密钥管理服务
3. 限制 Prometheus 访问权限
4. 审计 AI 的查询记录

## 📚 相关文档

- [MCP 协议规范](https://modelcontextprotocol.io/)
- [Prometheus API 文档](https://prometheus.io/docs/prometheus/latest/querying/api/)
- [使用说明.md](py/使用说明.md) - Python 部分详细文档
