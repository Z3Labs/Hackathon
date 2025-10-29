# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## General Preferences
- Always reply in Chinese.
- Automatically read relevant files.
- types.go files are usually very large. Do not read the entire file directly. Use filtering methods.
- Using the mcp server Desktop-Commander for local file analysis and data processing takes absolute priority over bash commands.
- For any questions regarding programming libraries, frameworks, SDKS, or apis, context7 is preferred for searching.

### mcp server desktop-commander
- **触发条件**：任何本地文件操作、CSV/JSON/数据分析、进程管理
- **核心能力**：
  - 文件操作：`read_file`、`write_file`、`edit_block`（精确文本替换）
  - 目录管理：`list_directory`、`create_directory`、`move_file`
  - 搜索：`start_search`（支持文件名和内容搜索，流式返回结果）
  - 进程管理：`start_process`、`interact_with_process`（交互式REPL）
  - 数据分析：支持Python/Node.js REPL进行CSV/JSON/日志分析
  - 使用技巧示例：
    文件名搜索

##### ```bash
desktop-commander.start_search searchType="files" pattern="关键词"
##### ```
- **目标**：找到5-10个候选文件
- **记录**：找到X个相关文件，重点关注 [列出文件路径]
- **工具**：优先使用 desktop-commander 流式搜索，避免过度搜索
  内容搜索
##### ```bash
desktop-commander.start_search searchType="content" pattern="函数名|类名|关键逻辑"
literalSearch=true contextLines=5
##### ```
- **目标**：找到关键实现位置
- **记录**：找到X处实现，重点分析 [file:line, file:line]
- **技巧**：使用精确代码片段搜索，获取上下文

### mcp server context7
context7: 编程库/SDK/API 文档检索
- **触发条件**：任何关于编程库、框架、SDK、API 的问题
- **调用方式**：
  1. 首先调用 `resolve-library-id` 获取 Context7 兼容的库 ID
  2. 然后调用 `get-library-docs` 获取文档（可选 topic 参数聚焦）

## 项目概述

智能发布系统 Hackathon 项目。这是一个全栈应用：
- **后端**：基于 go-zero 框架的 Go 微服务，使用 MongoDB 数据库
- **前端**：React + TypeScript + Vite 单页应用

## 开发命令

### 后端（Go + go-zero）

```bash
cd backend

# 从 API 定义生成代码（修改 api/hackathon.api 后执行）
make all

# 运行服务
make run
```

后端服务运行在 `http://localhost:8888`

### 前端（React + Vite）

前端运行在 `http://localhost:3000`，通过 `/api` 路径代理到后端服务

## 架构说明

### 后端架构（go-zero 结构）

后端遵循 go-zero 标准微服务架构：
后端 api 都不做鉴权处理

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
