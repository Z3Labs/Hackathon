# 生产环境部署指南

本文档描述如何在生产环境中部署智能发布系统。

## 快速开始（使用 Docker Compose）

如果你想要快速部署整个系统，可以使用 Docker Compose：

```bash
# 1. 配置环境变量
cp .env.example .env
# 编辑 .env 文件，填入七牛云密钥等配置

# 2. 启动所有服务
docker-compose up -d

# 3. 查看服务状态
docker-compose ps

# 4. 查看日志
docker-compose logs -f backend
```

服务访问地址：
- 前端: http://localhost
- API: http://localhost/api
- Prometheus: http://localhost:9090
- AlertManager: http://localhost:9093

如需完全自定义的生产部署，请参考以下详细步骤。

## 架构拓扑

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Nginx     │────▶│  Backend    │────▶│  MongoDB    │
│  (HTTP/80)  │     │  (8888)     │     │  (27017)    │
│  (HTTPS/443)│     └─────────────┘     └─────────────┘
└─────────────┘            │
     │                     │ SSH
     │                     │
┌────┼────┐       ┌────────▼────────┐
│ Frontend │       │   Target Hosts  │
│ (static) │       │  (Physical/VMs)│
└─────────┘       └─────────────────┘
```

## 一、环境准备

### 1.1 服务器要求

最低配置：
- **Frontend服务器**: 1核 1GB，10GB磁盘
- **Backend服务器**: 2核 4GB，50GB磁盘
- **MongoDB服务器**: 4核 8GB，100GB磁盘（推荐 SSD）
- **目标机器**: 根据实际应用需求

推荐配置：
- 所有服务器使用云主机（阿里云/腾讯云/AWS等）
- 启用公网 IP + 弹性带宽
- 配置安全组规则（仅开放必要端口）

### 1.2 网络配置

必需端口：
- `80/443`: Nginx 前端/API
- `8888`: 后端 API 服务
- `27017`: MongoDB
- `9090`: Prometheus
- `9093`: AlertManager
- `22`: SSH（目标机器）

生产环境建议：
- 使用内网 IP 连接 MongoDB
- 前端通过 Nginx 暴露到公网
- 后端 API 仅允许内网或配置 IP 白名单

### 1.3 系统依赖

```bash
# 所有服务器执行
yum update -y  # CentOS/RHEL
# 或 apt update && apt upgrade -y  # Ubuntu/Debian

# 安装基础工具
yum install -y git wget curl vim net-tools
```

## 二、数据库部署

### 2.1 MongoDB 安装（生产环境推荐）

```bash
# 创建 MongoDB 用户
useradd -r -s /bin/false mongodb

# 下载并安装 MongoDB 6.0
cat > /etc/yum.repos.d/mongodb-org-6.0.repo <<EOF
[mongodb-org-6.0]
name=MongoDB 6.0 Repository
baseurl=https://repo.mongodb.org/yum/redhat/7/mongodb-org/6.0/x86_64/
gpgcheck=1
enabled=1
gpgkey=https://www.mongodb.org/static/pgp/server-6.0.asc
EOF

yum install -y mongodb-org
```

### 2.2 配置 MongoDB

```bash
# 编辑配置文件
vim /etc/mongod.conf

# 关键配置
storage:
  dbPath: /data/mongodb
  journal:
    enabled: true
security:
  authorization: enabled

net:
  bindIp: 127.0.0.1,10.0.0.10  # 内网 IP
  port: 27017

# 创建数据目录
mkdir -p /data/mongodb
chown mongodb:mongodb /data/mongodb

# 启动服务
systemctl enable mongod
systemctl start mongod
```

### 2.3 创建数据库用户

```bash
mongo <<EOF
use admin
db.createUser({
  user: "hackathon_admin",
  pwd: "your_strong_password_here",
  roles: [{role: "readWriteAnyDatabase", db: "admin"}]
})

use hackathon
db.createUser({
  user: "hackathon_user",
  pwd: "your_app_password_here",
  roles: [{role: "readWrite", db: "hackathon"}]
})
EOF
```

### 2.4 MongoDB 备份配置

```bash
# 创建备份脚本
cat > /usr/local/bin/mongodb-backup.sh <<'EOF'
#!/bin/bash
BACKUP_DIR="/backup/mongodb"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

mongodump --host localhost:27017 \
  --username hackathon_admin \
  --password "your_strong_password_here" \
  --authenticationDatabase admin \
  --db hackathon \
  --out "$BACKUP_DIR/$DATE"

# 保留最近7天的备份
find $BACKUP_DIR -type d -mtime +7 -exec rm -rf {} \;
EOF

chmod +x /usr/local/bin/mongodb-backup.sh

# 添加到 crontab（每天凌晨2点备份）
echo "0 2 * * * /usr/local/bin/mongodb-backup.sh" | crontab -
```

## 三、后端服务部署

### 3.1 编译和部署

```bash
# 克隆代码到 /opt/hackathon
cd /opt
git clone https://github.com/your-org/hackathon.git
cd hackathon/backend

# 编译
go build -o hackathon-api hackathon.go

# 创建运行目录
mkdir -p /opt/hackathon/run/{logs,config}
cp etc/hackathon-api.yaml /opt/hackathon/run/config/
```

### 3.2 配置文件

```yaml
# /opt/hackathon/run/config/hackathon-api.yaml
Name: hackathon-api
Host: 0.0.0.0
Port: 8888

MongoDB:
  Host: 10.0.0.20:27017  # MongoDB 内网地址
  Database: hackathon
  Username: hackathon_user
  Password: your_app_password_here

Qiniu:
  AccessKey: your_access_key
  SecretKey: your_secret_key
  Bucket: your-bucket-name
  Domain: https://your-domain.com

Log:
  ServiceName: hackathon
  Mode: file
  Path: /opt/hackathon/run/logs
  Level: info
  Compress: true
  KeepDays: 7
```

### 3.3 Systemd 服务配置

```bash
cat > /etc/systemd/system/hackathon-api.service <<'EOF'
[Unit]
Description=Hackathon API Service
After=network.target

[Service]
Type=simple
User=hackathon
Group=hackathon
WorkingDirectory=/opt/hackathon/run
ExecStart=/opt/hackathon/backend/hackathon-api -f /opt/hackathon/run/config/hackathon-api.yaml
Restart=always
RestartSec=5
StandardOutput=append:/opt/hackathon/run/logs/service.log
StandardError=append:/opt/hackathon/run/logs/error.log

[Install]
WantedBy=multi-user.target
EOF

# 创建运行用户
useradd -r -s /bin/false -d /opt/hackathon/run hackathon
chown -R hackathon:hackathon /opt/hackathon

# 启动服务
systemctl daemon-reload
systemctl enable hackathon-api
systemctl start hackathon-api
```

### 3.4 监控和日志

```bash
# 查看服务状态
systemctl status hackathon-api

# 查看日志
journalctl -u hackathon-api -f
tail -f /opt/hackathon/run/logs/service.log

# 检查端口
netstat -tlnp | grep 8888
```

## 四、前端部署

### 4.1 构建生产版本

```bash
cd /opt/hackathon/frontend

# 安装依赖
npm ci --production

# 构建
npm run build
# 生成 dist/ 目录
```

### 4.2 Nginx 配置

```bash
# 安装 Nginx
yum install -y nginx

# 配置虚拟主机
cat > /etc/nginx/conf.d/hackathon.conf <<'EOF'
upstream backend {
    server 127.0.0.1:8888;
}

server {
    listen 80;
    server_name your-domain.com;

    # 访问日志
    access_log /var/log/nginx/hackathon-access.log;
    error_log /var/log/nginx/hackathon-error.log;

    # 前端静态文件
    location / {
        root /opt/hackathon/frontend/dist;
        index index.html;
        try_files $uri $uri/ /index.html;
    }

    # API 代理
    location /api/ {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # 文件上传大小限制
    client_max_body_size 100M;
}
EOF

# 启动 Nginx
systemctl enable nginx
systemctl start nginx
```

### 4.3 SSL/HTTPS 配置

```bash
# 使用 Let's Encrypt 免费证书
yum install -y certbot python3-certbot-nginx

# 申请证书
certbot --nginx -d your-domain.com

# 自动续期
echo "0 3 * * * certbot renew --quiet" | crontab -
```

## 五、监控和告警

### 5.1 Prometheus 配置

```bash
# 部署 Prometheus
wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz
tar xvf prometheus-2.45.0.linux-amd64.tar.gz
mv prometheus-2.45.0.linux-amd64 /opt/prometheus

# 配置文件
cat > /opt/prometheus/prometheus.yml <<'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'hackathon-backend'
    static_configs:
      - targets: ['localhost:8888']
  
  - job_name: 'target-hosts'
    static_configs:
      - targets: ['192.168.1.10:9100', '192.168.1.11:9100']
EOF

# 创建 systemd 服务
cat > /etc/systemd/system/prometheus.service <<'EOF'
[Unit]
Description=Prometheus
After=network.target

[Service]
Type=simple
User=prometheus
Group=prometheus
ExecStart=/opt/prometheus/prometheus --config.file=/opt/prometheus/prometheus.yml
Restart=always

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable prometheus
systemctl start prometheus
```

### 5.2 AlertManager 配置

```bash
wget https://github.com/prometheus/alertmanager/releases/download/v0.26.0/alertmanager-0.26.0.linux-amd64.tar.gz
tar xvf alertmanager-0.26.0.linux-amd64.tar.gz
mv alertmanager-0.26.0.linux-amd64 /opt/alertmanager

# 配置文件
cat > /opt/alertmanager/alertmanager.yml <<'EOF'
global:
  resolve_timeout: 5m

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'webhook'

receivers:
- name: 'webhook'
  webhook_configs:
  - url: 'http://10.0.0.10:8888/api/alerts/callback'
    send_resolved: true
EOF

# 启动服务
systemctl enable alertmanager
systemctl start alertmanager
```

## 六、安全加固

### 6.1 防火墙配置

```bash
# 安装 firewalld
yum install -y firewalld
systemctl enable firewalld
systemctl start firewalld

# 开放必要端口
firewall-cmd --permanent --add-service=http
firewall-cmd --permanent --add-service=https
firewall-cmd --permanent --add-port=8888/tcp
firewall-cmd --reload
```

### 6.2 SSH 安全

```bash
# 禁用密码登录，仅允许密钥
vim /etc/ssh/sshd_config
# 设置
PasswordAuthentication no
PermitRootLogin no

systemctl restart sshd
```

### 6.3 密钥管理

```bash
# 使用加密存储敏感信息
# 推荐使用 Vault 或 KMS

# 在环境变量中设置
export QINIU_ACCESS_KEY=$(vault kv get -field=access_key secret/qiniu)
export QINIU_SECRET_KEY=$(vault kv get -field=secret_key secret/qiniu)
```

## 七、高可用配置（可选）

### 7.1 负载均衡

```bash
# 部署多个后端实例
# 在 Nginx 中配置：
upstream backend {
    server 10.0.0.10:8888;
    server 10.0.0.11:8888;
    server 10.0.0.12:8888;
}
```

### 7.2 MongoDB 副本集

```bash
# 配置主从复制
# 在 /etc/mongod.conf 中：
replication:
  replSetName: "rs0"

# 初始化副本集
mongo --eval "rs.initiate({
  _id: 'rs0',
  members: [
    {_id: 0, host: '10.0.0.20:27017'},
    {_id: 1, host: '10.0.0.21:27017'},
    {_id: 2, host: '10.0.0.22:27017', arbiterOnly: true}
  ]
})"
```

## 八、性能调优

### 8.1 后端优化

```bash
# 在 go-zero 配置中
CacheRedis:
  - Host: localhost:6379
      Pass:
      Db: 0
      Type: node

# 连接池配置
MaxOpenConnections: 100
MaxIdleConnections: 10
ConnMaxLifetime: 3600
```

### 8.2 Nginx 优化

```bash
# /etc/nginx/nginx.conf
worker_processes auto;
worker_connections 2048;

# 启用缓存
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=api_cache:10m max_size=1g;

location /api/ {
    proxy_cache api_cache;
    proxy_cache_valid 200 5m;
}
```

## 九、故障排查

### 9.1 检查清单

```bash
# 服务状态
systemctl status hackathon-api nginx mongod

# 端口监听
netstat -tlnp | grep -E '8888|80|443|27017'

# 磁盘空间
df -h

# 内存使用
free -h

# 日志
tail -f /opt/hackathon/run/logs/service.log
tail -f /var/log/nginx/hackathon-error.log
```

### 9.2 常见问题

**问题：后端无法连接 MongoDB**
```bash
# 检查 MongoDB 状态
systemctl status mongod
# 测试连接
mongo "mongodb://hackathon_user:password@10.0.0.20:27017/hackathon"
```

**问题：Nginx 502 错误**
```bash
# 检查后端是否运行
curl http://127.0.0.1:8888/api/ping
# 检查 Nginx 错误日志
tail -f /var/log/nginx/hackathon-error.log
```

## 十、升级和维护

### 10.1 滚动更新

```bash
# 1. 拉取最新代码
cd /opt/hackathon
git pull origin main

# 2. 重新编译
cd backend
go build -o hackathon-api hackathon.go

# 3. 平滑重启
systemctl reload hackathon-api
```

### 10.2 数据迁移

```bash
# 备份数据
mongodump --host localhost:27017 -d hackathon -o /backup/

# 迁移到新环境
mongorestore --host new-host:27017 -d hackathon /backup/hackathon/
```

### 10.3 回滚

```bash
# 使用 Git 回滚代码
cd /opt/hackathon
git checkout <previous-commit>

# 重新编译和部署
cd backend && go build
systemctl restart hackathon-api
```
