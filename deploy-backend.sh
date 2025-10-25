#!/bin/bash

# 后端自动打包和部署脚本
# 使用方法: ./deploy-backend.sh <服务器IP>
# 例如: ./deploy-backend.sh 192.168.1.100

set -e  # 遇到错误立即退出

if [ $# -lt 1 ]; then
    echo "使用方法: $0 <服务器IP>"
    echo "例如: $0 192.168.1.100"
    exit 1
fi

SERVER_IP=$1
SERVER_USER="root"  # 默认使用root用户
BACKEND_DIR="/servers/backend"
SERVICE_NAME="hackathon"

echo "🚀 开始后端自动部署到服务器: $SERVER_IP (用户: $SERVER_USER)"

# 1. 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: Go未安装，请先安装Go"
    exit 1
fi

# 2. 进入后端目录
cd backend

# 3. 清理旧的构建文件
echo "🧹 清理旧的构建文件..."
rm -f hackathon hackathon-*

# 4. 构建Go程序
echo "🔨 开始构建Go程序..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o hackathon hackathon.go

if [ ! -f "hackathon" ]; then
    echo "❌ 构建失败，hackathon可执行文件不存在"
    exit 1
fi

echo "✅ Go程序构建成功"

# 5. 检查文件大小
FILE_SIZE=$(du -h hackathon | cut -f1)
echo "📦 构建文件大小: $FILE_SIZE"

# 6. 上传到服务器
echo "📤 上传文件到服务器..."
scp hackathon $SERVER_USER@$SERVER_IP:$BACKEND_DIR/hackathon.new

# 7. 在服务器上执行部署
echo "🔄 在服务器上执行部署..."
ssh $SERVER_USER@$SERVER_IP << EOF
    set -e
    
    echo "📁 进入后端目录: $BACKEND_DIR"
    cd $BACKEND_DIR
    
    # 备份当前版本
    if [ -f "hackathon" ]; then
        echo "💾 备份当前版本..."
        cp hackathon hackathon.backup.\$(date +%Y%m%d_%H%M%S)
    fi
    
    # 停止服务
    echo "⏹️  停止hackathon服务..."
    systemctl stop $SERVICE_NAME || true
    
    # 替换可执行文件
    echo "🔄 替换可执行文件..."
    mv hackathon.new hackathon
    chmod +x hackathon
    
    # 启动服务
    echo "▶️  启动hackathon服务..."
    systemctl start $SERVICE_NAME
    
    # 等待服务启动
    echo "⏳ 等待服务启动..."
    sleep 3
    
    # 检查服务状态
    echo "🔍 检查服务状态..."
    if systemctl is-active --quiet $SERVICE_NAME; then
        echo "✅ hackathon服务启动成功"
        systemctl status $SERVICE_NAME --no-pager -l
    else
        echo "❌ hackathon服务启动失败"
        systemctl status $SERVICE_NAME --no-pager -l
        exit 1
    fi
    
    # 检查端口监听
    echo "🔍 检查端口监听..."
    if netstat -tlnp | grep -q ":8888"; then
        echo "✅ 端口8888监听正常"
    else
        echo "❌ 端口8888未监听"
        exit 1
    fi
EOF

if [ $? -eq 0 ]; then
    echo ""
    echo "🎉 后端部署成功！"
    echo "📡 API地址: http://$SERVER_IP:8888"
    echo "🔗 测试接口: http://$SERVER_IP:8888/api/ping"
    echo ""
    echo "📋 常用管理命令:"
    echo "  查看服务状态: ssh $SERVER_USER@$SERVER_IP 'systemctl status $SERVICE_NAME'"
    echo "  查看服务日志: ssh $SERVER_USER@$SERVER_IP 'journalctl -u $SERVICE_NAME -f'"
    echo "  重启服务: ssh $SERVER_USER@$SERVER_IP 'systemctl restart $SERVICE_NAME'"
else
    echo "❌ 后端部署失败"
    exit 1
fi
