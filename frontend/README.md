# Hackathon Frontend

基于 React + TypeScript + Vite 的前端项目。

## 技术栈

- React 18
- TypeScript
- Vite
- React Router
- Axios

## 目录结构

```
frontend/
├── public/           # 静态资源
├── src/
│   ├── assets/       # 资源文件
│   ├── components/   # 组件
│   ├── pages/        # 页面
│   ├── services/     # API 服务
│   ├── utils/        # 工具函数
│   ├── App.tsx       # 根组件
│   ├── main.tsx      # 入口文件
│   └── index.css     # 全局样式
├── index.html        # HTML 模板
├── package.json      # 依赖配置
├── tsconfig.json     # TypeScript 配置
└── vite.config.ts    # Vite 配置
```

## 快速开始

### 安装依赖

```bash
npm install
```

### 开发模式

```bash
npm run dev
```

访问 http://localhost:3000

### 构建生产版本

```bash
npm run build
```

### 预览生产版本

```bash
npm run preview
```

## API 代理

开发环境下，`/api` 路径会被代理到 `http://localhost:8888`，对应后端服务。
