# Hackathon Backend

基于 go-zero 框架和 MongoDB 的后端服务。

## 环境要求

- Go 1.21+
- MongoDB
- 七牛云账号（用于存储应用包）

## 环境变量配置

在运行服务前，需要设置以下环境变量：

### 七牛云配置（必需）

```bash
export QINIU_ACCESS_KEY=your_access_key
export QINIU_SECRET_KEY=your_secret_key
export QINIU_BUCKET=your_bucket_name
```

**获取七牛云密钥：**
1. 登录 [七牛云控制台](https://portal.qiniu.com/)
2. 进入"密钥管理"获取 AccessKey 和 SecretKey
3. 在"对象存储"中创建存储空间（Bucket）

### AI 配置（可选）

```bash
export AI_BASE_URL=https://api.openai.com/v1
export AI_API_KEY=your_api_key
export AI_MODEL=gpt-4
export AI_USE_MCP=true
export PROMETHEUS_URL=http://localhost:9090
```

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 生成代码（修改 API 后需要执行）

```bash
make all
```

### 3. 启动服务

```bash
# 确保已设置环境变量
make run
```

服务将在 `http://localhost:8888` 启动。

## 开发指南

### API 定义

修改 `api/hackathon.api` 文件来定义新的 API 接口，然后运行：

```bash
make gen      # 生成代码
make format   # 格式化 API 文件
```

## 目录结构

```
backend/
├── api/              # API 定义文件
├── etc/              # 配置文件
├── internal/         # 内部代码
│   ├── config/       # 配置结构
│   ├── handler/      # 处理器
│   ├── logic/        # 业务逻辑
│   ├── svc/          # 服务上下文
│   ├── types/        # 类型定义
│   └── middleware/   # 中间件
├── model/            # 数据模型
└── common/           # 公共代码（如七牛云客户端）
```

## 故障排查

### 版本列表加载失败

如果前端无法加载版本列表，请检查：

1. **环境变量是否配置**：
   ```bash
   echo $QINIU_ACCESS_KEY
   echo $QINIU_SECRET_KEY
   echo $QINIU_BUCKET
   ```

2. **七牛云密钥是否正确**：
   - 确认密钥未过期
   - 确认 Bucket 名称正确
   - 确认账号有读取权限

3. **查看服务日志**：
   后端日志会显示详细的错误信息，包括七牛云配置问题。

### 七牛云存储结构

应用包应按以下路径存储：
```
niulink-materials/{app_name}/{version}_{app_name}-linux-amd64.tar.gz
```

例如：
```
niulink-materials/myapp/v1.0.0_myapp-linux-amd64.tar.gz
niulink-materials/myapp/v1.0.1_myapp-linux-amd64.tar.gz
```
