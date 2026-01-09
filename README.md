<img src="frontend/src/assets/logo.png" alt="SuperAIAutoCutVideo Logo" width="120" />

# SuperAIAutoCutVideo · AI智能视频剪辑

轻量、跨平台的一站式智能视频处理桌面应用，面向内容创作者与团队的高效产出流程。

## 亮点特性

- 多项目管理：项目级配置与切换
- 短剧解说工作流：多集上传 → 自动合并 → 生成解说脚本（当前基于字幕分析） → 生成解说视频
- 自动提取视频字幕
- 支持自定义提示词（高级配置）
- 支持上传字幕文件（高级配置）
- 支持腾讯 TTS、Edge TTS

## 技术栈

- 桌面端：Electron
- 前端：React + Vite
- 后端：Go（云端 API）
- 存储与任务：S3 + Redis/队列 + Worker
- 视频处理：FFmpeg / OpenCV（云端）

## 快速开始

前置要求：`Node.js ≥ 18`、`Go ≥ 1.21`

### 开发（本地连接云端 API）

```bash
# 安装前端依赖
cd frontend
npm install

# 配置云端 API
cp .env.example .env.local
# 编辑 VITE_API_BASE_URL 与 VITE_WS_BASE_URL

# 启动 Electron + 前端
npm run dev:electron
```

## 更新计划（持续优化中）

- 加入更多的大模型集合平台
- 添加 OCR 识别字幕
- 添加 Whisper 提取字幕
- 添加影视解说功能
- 添加视觉分析视频功能
- 打包 Windows 和 macOS 版本（自动更新）

## 文档

- 云端 API 规划：`docs/cloud_api.md`
- 前端说明：`docs/FRONTEND_README.md`
- 使用指南：`USAGE.md`

## 许可证

MIT

## 致谢

- Electron · React · Go · FFmpeg · OpenCV
