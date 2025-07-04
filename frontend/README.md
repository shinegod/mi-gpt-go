# MiGPT Go 前端管理面板

基于 Vue3 + Element Plus 构建的 MiGPT Go 系统管理界面。

## 功能特性

- 📊 **实时仪表盘** - 系统状态监控和快捷操作
- ⚙️ **配置管理** - 在线修改所有系统配置
- 🎤 **音箱控制** - TTS测试和音箱状态管理  
- ⚡ **并发处理** - 并发配置和性能监控
- 📝 **系统日志** - 实时日志查看和过滤

## 技术栈

- **Vue 3** - 渐进式JavaScript框架
- **Element Plus** - Vue 3 组件库
- **Pinia** - Vue 3 状态管理
- **Vite** - 现代前端构建工具
- **Axios** - HTTP客户端

## 开发环境搭建

### 1. 安装依赖

```bash
cd frontend
npm install
# 或
yarn install
# 或
pnpm install
```

### 2. 启动开发服务器

```bash
npm run dev
```

前端服务将运行在 `http://localhost:3000`

### 3. 构建生产版本

```bash
npm run build
```

构建产物将输出到 `../internal/web/static` 目录，可直接被Go后端服务。

## 项目结构

```
frontend/
├── src/
│   ├── api/              # API请求封装
│   ├── components/       # 公共组件
│   │   └── Layout.vue    # 布局组件
│   ├── stores/           # Pinia状态管理
│   ├── views/            # 页面组件
│   │   ├── Dashboard.vue # 仪表盘
│   │   ├── Config.vue    # 配置管理
│   │   ├── Speaker.vue   # 音箱控制
│   │   ├── Concurrent.vue# 并发处理
│   │   └── Logs.vue      # 系统日志
│   ├── router/           # 路由配置
│   ├── App.vue           # 根组件
│   └── main.js           # 入口文件
├── index.html            # HTML模板
├── package.json          # 项目配置
├── vite.config.js        # Vite配置
└── README.md             # 说明文档
```

## 配置说明

### API代理配置

Vite开发服务器已配置API代理，将 `/api` 请求转发到后端 `http://localhost:8080`。

### 自动导入配置

项目配置了Element Plus组件和Vue API的自动导入，无需手动引入。

### 构建配置

构建输出目录配置为 `../internal/web/static`，方便Go后端直接提供静态文件服务。

## 使用说明

### 1. 启动后端服务

```bash
cd ..
go run main.go
```

### 2. 启动前端开发服务器

```bash
npm run dev
```

### 3. 访问管理面板

- 开发环境: `http://localhost:3000`
- 生产环境: `http://localhost:8080` (Go后端提供)

## 主要功能

### 仪表盘
- 系统运行状态监控
- 音箱连接状态显示
- 并发处理性能统计
- 快捷操作按钮

### 配置管理
- AI服务配置（OpenAI/Azure/DeepSeek）
- 机器人人设配置
- 音箱行为配置
- 小米设备配置
- 并发处理配置
- 数据库配置

### 音箱控制
- TTS语音测试
- 音箱状态查看
- 预设文本快捷测试

### 并发处理
- 实时性能监控
- 配置参数调整
- 性能优化建议

### 系统日志
- 实时日志查看
- 日志级别过滤
- 关键词搜索
- 日志统计分析

## 开发指南

### 添加新页面

1. 在 `src/views/` 下创建Vue组件
2. 在 `src/router/index.js` 中添加路由配置
3. 在Layout组件的菜单中添加导航项

### 添加新API

1. 在 `src/api/index.js` 中添加API方法
2. 在对应的Pinia store中添加状态管理
3. 在组件中使用store方法

### 样式自定义

项目使用Element Plus默认主题，可通过CSS变量自定义样式。

## 部署说明

### 开发部署

前端开发服务器和Go后端分别运行，通过代理进行通信。

### 生产部署

执行 `npm run build` 构建前端资源，Go后端直接提供静态文件服务。

## 注意事项

- 确保Go后端服务正常运行
- API请求依赖后端Web服务器
- 生产环境下前后端使用同一端口（8080）
- 配置修改需要后端重启才能生效

## 常见问题

### Q: 前端无法连接后端
A: 检查Go后端是否正常启动，确认端口8080未被占用

### Q: 配置保存失败
A: 检查后端API服务是否正常，查看浏览器网络请求

### Q: 页面样式异常
A: 清除浏览器缓存，重新构建前端项目

## 更新日志

- v1.0.0 - 初始版本，包含基础管理功能 