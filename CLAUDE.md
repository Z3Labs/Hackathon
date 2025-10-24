# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 重要说明

**始终使用中文回答和交流**。本项目团队使用中文作为主要工作语言。

## 项目概述

智能发布系统 Hackathon 项目。这是一个全栈应用：
- **后端**：基于 go-zero 框架的 Go 微服务，使用 MongoDB 数据库
- **前端**：React + TypeScript + Vite 单页应用

## 开发命令

### 后端（Go + go-zero）

```bash
cd backend

# 安装依赖
go mod tidy

# 从 API 定义生成代码（修改 api/hackathon.api 后执行）
goctl api go -api api/hackathon.api -dir .

# 运行服务
go run hackathon.go -f etc/hackathon-api.yaml
```

后端服务运行在 `http://localhost:8888`

### 前端（React + Vite）

前端运行在 `http://localhost:3000`，通过 `/api` 路径代理到后端服务

## 架构说明

### 后端架构（go-zero 结构）

后端遵循 go-zero 标准微服务架构：

- **api/** - API 定义文件（`.api` 格式），定义服务契约
  - 修改 `api/hackathon.api` 后需要使用 `goctl api go` 命令重新生成代码
- **etc/** - 配置文件（YAML 格式）
  - `hackathon-api.yaml` 包含服务器端口（8888）和 MongoDB 连接配置
- **internal/** - 内部应用代码（包含生成代码和自定义代码）
  - `config/` - 配置结构体定义
  - `handler/` - HTTP 处理器（请求的入口点）
  - `logic/` - 业务逻辑层（核心应用逻辑）
  - `svc/` - 服务上下文（依赖注入容器）
  - `types/` - 类型定义（请求/响应结构体）
  - `middleware/` - HTTP 中间件组件

**关键工作流程**：添加新接口时，先在 `api/hackathon.api` 中定义，然后运行 `goctl` 代码生成命令，最后在生成的 logic 文件中实现业务逻辑。

### 前端架构（React 架构）

- **src/components/** - 可复用的 React 组件
- **src/pages/** - 页面级组件（路由目标）
- **src/services/** - API 服务层（基于 Axios 的后端 HTTP 调用）
- **src/utils/** - 工具函数
- **src/App.tsx** - 根组件，包含路由配置
- **vite.config.ts** - 开发服务器配置，运行在 3000 端口，将 `/api` 请求代理到 `http://localhost:8888`

### 数据库

MongoDB 连接配置在 `backend/etc/hackathon-api.yaml` 中：
- 默认 URI：`mongodb://localhost:27017`
- 数据库名：`hackathon`

## 使用 go-zero API 定义

`.api` 文件格式是 go-zero 定义 HTTP 服务的 DSL：

```
type (
    RequestType {
        Field string `json:"field"`
    }
    ResponseType {
        Result string `json:"result"`
    }
)

service hackathon-api {
    @handler HandlerName
    post /path (RequestType) returns (ResponseType)
}
```

修改 API 定义后，务必先运行代码生成，再实现处理器逻辑。
