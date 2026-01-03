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

- 桌面端：Tauri
- 前端：React + Vite
- 后端：FastAPI
- 视频处理：FFmpeg / OpenCV

## 快速开始

前置要求：`Node.js ≥ 18`、`Python ≥ 3.11`、`Rust`、`FFmpeg`

### Windows（PowerShell）

```powershell
# 安装前端依赖（优先使用 cnpm）
cd frontend
cnpm install

# 创建并使用后端虚拟环境（无需激活）
cd ..
py -3 -m venv backend\.venv
backend\.venv\Scripts\python.exe -m pip install -r backend\requirements.txt
backend\.venv\Scripts\python.exe backend\main.py
```

#### Edge TTS 使用说明（代理与持久化）

- Edge TTS 免凭据，但在中国大陆直连通常会返回 403 或 TLS 连接错误；需配置 HTTP 代理。
- 推荐使用持久化配置：在 `backend/config/tts_config.json` 的 `edge_tts_default.extra_params` 写入 `ProxyUrl`（示例：`http://127.0.0.1:7890`）。
- 后端代理解析优先级：`EDGE_TTS_PROXY` → `HTTPS_PROXY`/`HTTP_PROXY` → `extra_params.ProxyUrl`。
- 每次测试可直接调用接口，无需手动设置环境变量：
  - 无代理（示例，仅在可直连时可用）：
    ```powershell
    Invoke-RestMethod -Method POST -Uri "http://127.0.0.1:8000/api/tts/configs/edge_tts_default/test" -ContentType 'application/json'
    ```
  - 指定临时代理：
    ```powershell
    Invoke-RestMethod -Method POST -Uri "http://127.0.0.1:8000/api/tts/configs/edge_tts_default/test?proxy_url=http://127.0.0.1:7890" -ContentType 'application/json'
    ```
- 若使用虚拟环境 `Activate.ps1`，需在激活后的会话中重新设置会话级环境变量；或继续按上面方式直接调用虚拟环境内的 `python.exe`，避免环境变量丢失。
- 前端会自动发现后端端口（扫描范围 8000–8019），无需手动修改端口。

### macOS

```bash
# 安装前端依赖
cd frontend
cnpm install

# 创建并使用后端虚拟环境（可不激活）
cd ..
python3 -m venv backend/.venv
backend/.venv/bin/python -m pip install -r backend/requirements.txt
backend/.venv/bin/python backend/main.py

# 启动桌面应用
cargo tauri dev
```

## 更新计划（持续优化中）

- 加入更多的大模型集合平台
- 添加 OCR 识别字幕
- 添加 Whisper 提取字幕
- 添加影视解说功能
- 添加视觉分析视频功能
- 打包 Windows 和 macOS 版本

## 文档

- 后端 API 文档：`docs/backend_api_documentation.md`
- 前端说明：`docs/FRONTEND_README.md`
- 使用指南：`USAGE.md`

## 许可证

MIT

## 致谢

- Tauri · React · FastAPI · FFmpeg · OpenCV
