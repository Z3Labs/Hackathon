#!/bin/bash

# 前端自动打包和部署脚本
# 使用方法: ./deploy-frontend-auto.sh <服务器IP>
# 例如: ./deploy-frontend-auto.sh 192.168.1.100

set -e  # 遇到错误立即退出

if [ $# -lt 1 ]; then
    echo "使用方法: $0 <服务器IP>"
    echo "例如: $0 192.168.1.100"
    exit 1
fi

SERVER_IP=$1
SERVER_USER="root"  # 默认使用root用户
FRONTEND_DIR="/servers/frontend"

echo "🚀 开始前端自动部署到服务器: $SERVER_IP (用户: $SERVER_USER)"

# 1. 检查Node.js环境
if ! command -v node &> /dev/null; then
    echo "❌ 错误: Node.js未安装，请先安装Node.js"
    exit 1
fi

if ! command -v npm &> /dev/null; then
    echo "❌ 错误: npm未安装，请先安装npm"
    exit 1
fi

# 2. 进入前端目录
cd frontend

# 3. 检查package.json
if [ ! -f "package.json" ]; then
    echo "❌ 错误: package.json不存在"
    exit 1
fi

# 4. 安装依赖（如果需要）
if [ ! -d "node_modules" ]; then
    echo "📦 安装npm依赖..."
    npm install
fi

# 5. 清理旧的构建文件
echo "🧹 清理旧的构建文件..."
rm -rf dist/

# 6. 构建前端项目
echo "🔨 开始构建前端项目..."
npm run build

if [ ! -d "dist" ]; then
    echo "❌ 构建失败，dist目录不存在"
    exit 1
fi

echo "✅ 前端项目构建成功"

# 7. 检查构建文件
DIST_SIZE=$(du -sh dist/ | cut -f1)
echo "📦 构建文件大小: $DIST_SIZE"

# 8. 创建临时压缩包
echo "📦 创建部署包..."
TEMP_DIR="/tmp/hackathon-frontend-$(date +%s)"
mkdir -p $TEMP_DIR
cp -r dist/* $TEMP_DIR/
cd $TEMP_DIR
tar -czf ../frontend-dist.tar.gz .
cd - > /dev/null

# 9. 上传到服务器
echo "📤 上传文件到服务器..."
scp /tmp/frontend-dist.tar.gz $SERVER_USER@$SERVER_IP:/tmp/

# 10. 在服务器上执行部署
echo "🔄 在服务器上执行部署..."
ssh $SERVER_USER@$SERVER_IP << EOF
    set -e
    
    echo "📁 进入前端目录: $FRONTEND_DIR"
    mkdir -p $FRONTEND_DIR
    cd $FRONTEND_DIR
    
    # 备份当前版本
    if [ -d "index.html" ] || [ -f "index.html" ]; then
        echo "💾 备份当前版本..."
        BACKUP_DIR="backup_\$(date +%Y%m%d_%H%M%S)"
        mkdir -p ../\$BACKUP_DIR
        cp -r . ../\$BACKUP_DIR/ 2>/dev/null || true
    fi
    
    # 清理当前文件
    echo "🧹 清理当前文件..."
    rm -rf *
    
    # 解压新文件
    echo "📦 解压新文件..."
    cd $FRONTEND_DIR
    tar -xzf /tmp/frontend-dist.tar.gz
    rm -f /tmp/frontend-dist.tar.gz
    
    # 设置文件权限
    echo "🔐 设置文件权限..."
    chown -R www-data:www-data $FRONTEND_DIR
    chmod -R 755 $FRONTEND_DIR
    
    # 检查nginx配置
    echo "🔍 检查nginx配置..."
    if [ ! -f "/etc/nginx/sites-available/hackathon" ]; then
        echo "⚠️  nginx配置文件不存在，创建默认配置..."
        cat > /etc/nginx/sites-available/hackathon << 'NGINX_EOF'
server {
    listen 13354;
    server_name _;
    
    root $FRONTEND_DIR;
    index index.html;
    
    # 前端路由支持
    location / {
        try_files \$uri \$uri/ /index.html;
    }
    
    # API代理到后端
    location /api/ {
        proxy_pass http://localhost:8888/api/;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
    
    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
NGINX_EOF
        
        # 启用站点
        ln -sf /etc/nginx/sites-available/hackathon /etc/nginx/sites-enabled/
        rm -f /etc/nginx/sites-enabled/default
    fi
    
    # 测试nginx配置
    echo "🔍 测试nginx配置..."
    nginx -t
    
    # 重启nginx
    echo "🔄 重启nginx..."
    systemctl reload nginx
    
    # 检查nginx状态
    echo "🔍 检查nginx状态..."
    if systemctl is-active --quiet nginx; then
        echo "✅ nginx服务运行正常"
    else
        echo "❌ nginx服务异常"
        systemctl status nginx --no-pager -l
        exit 1
    fi
    
    # 检查端口监听
    echo "🔍 检查端口监听..."
    if netstat -tlnp | grep -q ":13354"; then
        echo "✅ 端口13354监听正常"
    else
        echo "❌ 端口13354未监听"
        exit 1
    fi
EOF

# 11. 清理临时文件
echo "🧹 清理临时文件..."
rm -rf $TEMP_DIR /tmp/frontend-dist.tar.gz

if [ $? -eq 0 ]; then
    echo ""
    echo "🎉 前端部署成功！"
    echo "🌐 前端地址: http://$SERVER_IP:13354"
    echo "📡 API地址: http://$SERVER_IP:8888"
    echo ""
    echo "📋 常用管理命令:"
    echo "  查看nginx状态: ssh $SERVER_USER@$SERVER_IP 'systemctl status nginx'"
    echo "  查看nginx日志: ssh $SERVER_USER@$SERVER_IP 'journalctl -u nginx -f'"
    echo "  重启nginx: ssh $SERVER_USER@$SERVER_IP 'systemctl restart nginx'"
    echo "  查看前端文件: ssh $SERVER_USER@$SERVER_IP 'ls -la $FRONTEND_DIR'"
else
    echo "❌ 前端部署失败"
    exit 1
fi
