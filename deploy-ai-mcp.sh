#!/bin/bash

# AI MCP 诊断服务自动打包和部署脚本
# 使用方法: ./deploy-ai-mcp.sh <服务器IP>
# 例如: ./deploy-ai-mcp.sh 192.168.1.100

set -e  # 遇到错误立即退出

if [ $# -lt 1 ]; then
    echo "使用方法: $0 <服务器IP>"
    echo "例如: $0 192.168.1.100"
    exit 1
fi

SERVER_IP=$1
SERVER_USER="root"  # 默认使用root用户
DEPLOY_DIR="/servers/backend/ai-mcp-call"
CONTAINER_NAME="diagnosis-service"
IMAGE_NAME="diagnosis-service:latest"
IMAGE_FILE="diagnosis-service-latest.tar.gz"

echo "🚀 开始 AI MCP 诊断服务自动部署到服务器: $SERVER_IP (用户: $SERVER_USER)"

# 1. 检查Docker环境
if ! command -v docker &> /dev/null; then
    echo "❌ 错误: Docker未安装，请先安装Docker"
    exit 1
fi

# 2. 进入AI诊断服务目录
echo "📁 进入AI诊断服务目录..."
cd backend/internal/clients/diagnosis/py

# 3. 清理旧的镜像包
echo "🧹 清理旧的镜像包..."
rm -f diagnosis-service-*.tar.gz

# 4. 构建Docker镜像（跨平台构建 linux/amd64）
echo "🔨 开始构建Docker镜像（目标平台: linux/amd64）..."
./build-docker.sh latest linux/amd64

if [ ! -f "$IMAGE_FILE" ]; then
    echo "❌ 构建失败，镜像包不存在: $IMAGE_FILE"
    exit 1
fi

echo "✅ Docker镜像构建成功"

# 5. 检查文件大小
FILE_SIZE=$(du -h "$IMAGE_FILE" | cut -f1)
echo "📦 镜像包大小: $FILE_SIZE"

# 6. 上传到服务器（先上传为 .new 文件）
echo "📤 上传镜像包到服务器..."
ssh $SERVER_USER@$SERVER_IP "mkdir -p $DEPLOY_DIR"
scp "$IMAGE_FILE" $SERVER_USER@$SERVER_IP:$DEPLOY_DIR/${IMAGE_FILE}.new

# 7. 在服务器上执行部署
echo "🔄 在服务器上执行部署..."
ssh $SERVER_USER@$SERVER_IP << EOF
    set -e
    
    echo "📁 进入部署目录: $DEPLOY_DIR"
    cd $DEPLOY_DIR
    
    # 备份旧的镜像包
    TIMESTAMP=\$(date +%Y%m%d_%H%M%S)
    if [ -f "$IMAGE_FILE" ]; then
        echo "💾 备份旧的镜像包..."
        mv $IMAGE_FILE ${IMAGE_FILE%.tar.gz}.backup.\${TIMESTAMP}.tar.gz
        echo "✅ 旧镜像包已重命名为: ${IMAGE_FILE%.tar.gz}.backup.\${TIMESTAMP}.tar.gz"
    fi
    
    # 重命名新上传的镜像包
    echo "📦 处理新镜像包..."
    mv ${IMAGE_FILE}.new $IMAGE_FILE
    
    # 加载新镜像
    echo "📥 加载新的Docker镜像..."
    docker load < $IMAGE_FILE
    
    # 停止并重命名旧容器（备份）
    echo "⏹️  停止并备份旧容器..."
    if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}\$"; then
        docker stop $CONTAINER_NAME || true
        docker rename $CONTAINER_NAME ${CONTAINER_NAME}.backup.\${TIMESTAMP} || true
        echo "✅ 旧容器已停止并重命名为: ${CONTAINER_NAME}.backup.\${TIMESTAMP}"
    else
        echo "ℹ️  未发现旧容器"
    fi
    
    # 启动新容器
    echo "▶️  启动新容器..."
    docker run -d --name $CONTAINER_NAME --restart unless-stopped $IMAGE_NAME
    
    # 等待容器启动
    echo "⏳ 等待容器启动..."
    sleep 3
    
    # 检查容器状态
    echo "🔍 检查容器状态..."
    if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}\$"; then
        echo "✅ 容器启动成功"
        docker ps --filter "name=${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Image}}"
    else
        echo "❌ 容器启动失败"
        docker ps -a --filter "name=${CONTAINER_NAME}"
        docker logs $CONTAINER_NAME
        exit 1
    fi
    
    # 删除悬空镜像（被夺取tag的旧镜像）
    echo "🧹 清理悬空镜像..."
    DANGLING_IMAGES=\$(docker images -f "dangling=true" -q)
    if [ -n "\$DANGLING_IMAGES" ]; then
        docker rmi \$DANGLING_IMAGES || true
        echo "✅ 已删除悬空镜像"
    else
        echo "ℹ️  未发现悬空镜像"
    fi
    
    # 测试容器健康状态
    echo "🔍 测试容器健康状态..."
    if docker exec $CONTAINER_NAME python --version &> /dev/null; then
        echo "✅ 容器健康检查通过"
    else
        echo "⚠️  容器健康检查失败，但容器仍在运行"
    fi
EOF

if [ $? -eq 0 ]; then
    echo ""
    echo "🎉 AI MCP 诊断服务部署成功！"
    echo "📦 容器名称: $CONTAINER_NAME"
    echo "🖼️  镜像名称: $IMAGE_NAME"
    echo ""
    echo "📋 常用管理命令:"
    echo "  查看容器状态: ssh $SERVER_USER@$SERVER_IP 'docker ps --filter name=$CONTAINER_NAME'"
    echo "  查看容器日志: ssh $SERVER_USER@$SERVER_IP 'docker logs $CONTAINER_NAME'"
    echo "  进入容器: ssh $SERVER_USER@$SERVER_IP 'docker exec -it $CONTAINER_NAME /bin/bash'"
    echo "  测试诊断服务: ssh $SERVER_USER@$SERVER_IP 'docker exec $CONTAINER_NAME python /app/diagnosis_runner.py --help'"
    echo "  重启容器: ssh $SERVER_USER@$SERVER_IP 'docker restart $CONTAINER_NAME'"
    echo "  停止容器: ssh $SERVER_USER@$SERVER_IP 'docker stop $CONTAINER_NAME'"
else
    echo "❌ AI MCP 诊断服务部署失败"
    exit 1
fi
