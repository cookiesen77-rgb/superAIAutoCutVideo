# SuperAIAutoCutVideo 使用说明

轻量、跨平台的一站式智能视频处理桌面应用。当前版本使用云端 API 完成视频处理、AI脚本与 TTS 配音。

## 环境要求

- Node.js ≥ 18
- Go ≥ 1.21（仅用于本地运行云端 API 调试）

## 安装依赖

```bash
# 前端依赖
cd frontend
npm install

# 可选：本地运行云端 API（调试用）
cd ../server
go run ./cmd/api
```

## 启动应用

### 开发模式（Electron + 云端 API）

```bash
# 启动前端开发服务器
cd frontend
npm run dev

# 启动 Electron 开发窗口
npm run dev:electron
```

访问地址与端口：
- 前端开发：`http://localhost:1420`
- 云端 API：`VITE_API_BASE_URL` 指定
- WebSocket：`VITE_WS_BASE_URL` 指定

## 基本使用流程

1. 导入或拖拽视频到应用
2. 选择处理方式（剪辑、合并、响度标准化等）
3. 可选：生成 AI 脚本与 TTS 配音，并自动与视频片段对齐
4. 开始处理，在状态面板查看实时进度
5. 处理完成后导出成片

## API 快速测试

详细接口说明参考 `docs/cloud_api.md`。常用示例：

```bash
# 健康检查
curl https://api.example.com/api/health

# 注册 / 登录
curl -X POST https://api.example.com/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"your_password"}'

# 创建项目（需要带 Authorization: Bearer <token>）
curl -X POST https://api.example.com/api/projects \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name":"示例项目"}'
```

## 打包（Electron）

```bash
# 生成本地安装包（macOS/Windows）
cd frontend
npm run build:electron
```

自动更新基于 GitHub Releases，打包产物会写入 `frontend/release/`。云端服务部署请参考 `server/README.md`。

## 常见问题

- 云端 API 不可用：检查 `VITE_API_BASE_URL` 和 `VITE_WS_BASE_URL` 配置。
- 前端端口冲突：Vite 端口固定为 `1420`（`vite.config.ts` 中可调整）。

## 参考文档

- 项目概览与快速开始：`README.md`
- 云端 API 文档：`docs/cloud_api.md`
- 前端开发说明：`docs/FRONTEND_README.md`

## 许可证

MIT License
