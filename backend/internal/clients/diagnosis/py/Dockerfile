# AI 诊断服务 Docker 镜像
# 包含 Python 环境、Prometheus MCP Server (Python)、GitHub MCP Server (Go) 和诊断脚本

# ==================== 阶段 1: 编译 GitHub MCP Server ====================
FROM golang:1.24-alpine AS github-mcp-builder

# 安装编译依赖
RUN apk add --no-cache git ca-certificates

# 克隆并编译 github-mcp-server
WORKDIR /build
RUN git clone https://github.com/github/github-mcp-server.git . && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o github-mcp-server cmd/github-mcp-server/main.go

# ==================== 阶段 2: Python 运行环境 ====================
FROM python:3.11-slim

# 设置工作目录
WORKDIR /app

# 安装系统依赖（git 用于 clone MCP Server）
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    git \
    ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# 安装 Prometheus MCP Server（Python 实现）
RUN git clone https://github.com/spongehah/prometheus-mcp-server.git /tmp/prometheus-mcp && \
    cd /tmp/prometheus-mcp && \
    pip install --no-cache-dir -e . && \
    cd / && \
    rm -rf /tmp/prometheus-mcp/.git

# 复制 Python 依赖文件
COPY requirements.txt .

# 安装 Python 依赖
RUN pip install --no-cache-dir -r requirements.txt

# 复制 GitHub MCP Server 二进制（从第一阶段）
COPY --from=github-mcp-builder /build/github-mcp-server /usr/local/bin/github-mcp-server
RUN chmod +x /usr/local/bin/github-mcp-server

# 复制所有 Python 脚本到容器
COPY simple_anthropic_mcp.py .
COPY diagnosis_runner.py .

# 设置环境变量（可选，运行时会被覆盖）
ENV PYTHONUNBUFFERED=1

# 健康检查（检查 Python 和 MCP Server 是否可用）
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD python --version && python -c "import prometheus_mcp_server" || exit 1

# 保持容器运行，等待 docker exec 调用
CMD ["tail", "-f", "/dev/null"]
