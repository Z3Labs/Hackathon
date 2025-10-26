# 智能发布系统

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.19+-00ADD8.svg)](https://golang.org/)
[![React](https://img.shields.io/badge/react-18+-61dafb.svg)](https://reactjs.org/)

> 一个集成了 AI 智能诊断能力的现代化发布部署平台，支持灰度发布、自动回滚、实时监控与智能异常分析。

## 项目简介

这是一个用于简化应用部署流程的发布系统，支持灰度发布、自动回滚和监控告警。通过集成 AI 能力，当部署出现异常时能够自动分析日志和监控指标，快速定位问题根因，减少运维成本。

## 文档

### 项目启动
- [项目启动会议记录](doc/project_kickoff.md) - 议题讨论、AI 工具选择和团队分工

### MVP 设计
- [MVP 设计与实现计划](doc/design.md) - 对如何 3 天实现基本能力进行规划

### 架构设计
- [运行时架构描述](doc/system_runtime_architecture.txt)
```
 ┌─────────────────────────────────────────────────────────────────┐
 │                          主服务程序                               │
 │                    (Go-Zero API Server)                          │
 │                      Port: 8888                                  │
 │  ┌──────────────┐         ┌──────────────┐                      │
 │  │ 发布管理器     │  <-->   │ 发布执行器    │                      │
 │  │              │         │  (Ansible)   │                      │
 │  └──────────────┘         └───────┬───────┘                      │
 └───────────────────────────────────┼───────────────────────────────┘
                                     │ SSH
                                     ▼
 ┌───────────────────────────┐       ┌──────────────────────────┐
 │    AI 模型服务             │       │      发布设备 1           │
 │                           │       │       (物理机)           │
 │  ┌─────────────────────┐  │       │  ┌────────────────────┐  │
 │  │  AI 分析引擎         │  │       │  │ 业务服务进程 1      │  │
 │  │  (Claude/GPT)       │  │       │  │  (App Instance)    │  │
 │  └─────────────────────┘  │       │  └────────────────────┘  │
 └──────────┬─────────┬──────┘       └──────────┬───────────────┘
            │        MCP                         │ 指标上报
            │         │                         ▼
            │         │            ┌──────────────────────────┐
            │         │            │  指标采集系统              │
            │         │───────────>│  (Prometheus)            │
            │                      │  Port: 9090              │
            │                      └──────────┬───────────────┘
            │                                  │
            │                                  ▼
            │                      ┌──────────────────────────┐
            │                      │  告警系统                 │
            │                      │  (AlertManager)          │
            │                      └──────────────────────────┘
            │                                  │ Webhook
            │                                  ▼
            │                      ┌──────────────────────────┐
            │                      │      主服务程序            │
            └──────────────────────┴──────────────────────────┘
```

### mockServer 设计
- [MockServer 设计文档](doc/mockserver_design.md) - MockExecutor 测试平台实现与技术方案

### 快速开始
- [部署](doc/PRODUCTION_DEPLOYMENT.md)

### 模块设计文档
- [部署执行器设计](doc/deploy_executor.md)
- [发布计划设计](doc/deploy_plan.md)
- [部署设计文档](doc/deploy.md)
- [AI 诊断实现](doc/diagnosis_implementation.md)
- [Prometheus MCP 集成](doc/prometheus-mcp-server.md)

### 集成环境
http://118.89.104.170:13354/

### demo视频
https://meeting.tencent.com/crm/29e64vLxd8

## 贡献者

本项目由 Z³Labs 团队开发。

- [郑祥林](https://github.com/Lewinz)
- [李正乾](https://github.com/lzh2nix)
- [赵英杰](https://github.com/spongehah)
