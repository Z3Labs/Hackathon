# Hackathon Backend

基于 go-zero 框架和 MongoDB 的后端服务。

## 技术栈

- go-zero: 微服务框架
- MongoDB: 数据库

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
└── model/            # 数据模型
```

## 快速开始

### 安装依赖

```bash
go mod tidy
```

### 生成代码

```bash
goctl api go -api api/hackathon.api -dir .
```

### 运行服务

```bash
go run hackathon.go -f etc/hackathon-api.yaml
```

## 数据库配置

在 `etc/hackathon-api.yaml` 中配置 MongoDB 连接信息。
