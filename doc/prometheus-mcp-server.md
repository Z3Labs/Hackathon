📊 Prometheus MCP Server 概述

这是一个 Model Context Protocol (MCP) 服务器，可以让 AI 助手（如
Claude）通过标准化接口访问和分析 Prometheus 指标。

  ---
🛠️ 核心功能（Tools）

MCP 服务器提供了以下工具，AI 可以主动调用：

1. 查询工具

- execute_query - 执行即时 PromQL 查询，获取当前值
- execute_range_query - 执行范围查询，获取时间序列数据

2. 发现工具

- list_metrics() - 列出所有可用指标名称
- get_metric_metadata(metric: "metric_name") - 获取指标元数据（类型、说明等）
- get_targets() - 获取所有 Prometheus 抓取目标信息

  ---
📦 安装方式

方法 1: Claude Code CLI 集成（推荐）

claude mcp add prometheus \
--env PROMETHEUS_URL=http://your-prometheus:9090 \
-- docker run -i --rm -e PROMETHEUS_URL
ghcr.io/pab1it0/prometheus-mcp-server:latest

方法 2: Docker 运行

# 基本运行
docker run -i --rm \
-e PROMETHEUS_URL="http://your-prometheus:9090" \
ghcr.io/pab1it0/prometheus-mcp-server:latest

# 带认证
docker run -i --rm \
-e PROMETHEUS_URL="http://your-prometheus:9090" \
-e PROMETHEUS_USERNAME="admin" \
-e PROMETHEUS_PASSWORD="password" \
ghcr.io/pab1it0/prometheus-mcp-server:latest

方法 3: 本地安装

git clone https://github.com/spongehah/prometheus-mcp-server.git
cd prometheus-mcp-server
uv pip install -e .
python -m prometheus_mcp_server.main

  ---
⚙️ 配置说明

环境变量配置（.env 文件）

# 必需：Prometheus 服务器地址
PROMETHEUS_URL=http://your-prometheus-server:9090

# 可选：认证（二选一）
# 基本认证
PROMETHEUS_USERNAME=your_username
PROMETHEUS_PASSWORD=your_password

# Bearer Token 认证
PROMETHEUS_TOKEN=your_token

# 可选：传输模式
PROMETHEUS_MCP_SERVER_TRANSPORT=stdio  # 可选: http, stdio, sse
PROMETHEUS_MCP_BIND_HOST=localhost     # HTTP 模式使用
PROMETHEUS_MCP_BIND_PORT=8080          # HTTP 模式使用

Claude Desktop 配置

在 Claude Desktop 的配置文件中添加：
```json
{
    "mcpServers": {
        "prometheus": {
            "command": "uv",
            "args": [
                "--directory",
                "/full/path/to/prometheus-mcp-server",
                "run",
                "src/prometheus_mcp_server/main.py"
            ],
            "env": {
                "PROMETHEUS_URL": "http://your-prometheus-server:9090",
                "PROMETHEUS_USERNAME": "your_username",
                "PROMETHEUS_PASSWORD": "your_password"
            }
        }
    }
}
```
