# IndexTTS2 本地推理集成指南

## 功能概述

本次更新集成了 IndexTTS2 零样本语音克隆系统，支持：
- 本地 GPU 推理（无需云端 API）
- 6 种预置音色可选
- 8 种情感控制模式
- 自动情感推断

---

## 一、系统要求

| 项目 | 最低要求 | 推荐配置 |
|------|----------|----------|
| 操作系统 | Windows 10/11 | Windows 11 |
| GPU | NVIDIA 8GB 显存 | NVIDIA 12GB+ 显存 |
| 内存 | 32GB | 64GB |
| 硬盘 | 10GB 可用空间 | SSD |
| CUDA | 11.8+ | 12.1 |

---

## 二、安装步骤

### 步骤 1：安装 Python 依赖

```bash
cd backend
pip install -r requirements.txt
```

### 步骤 2：安装 IndexTTS2

```bash
pip install git+https://github.com/index-tts/index-tts.git
```

或者从 PyPI 安装（如果已发布）：
```bash
pip install indextts
```

### 步骤 3：下载模型文件

**方法 A：使用 ModelScope（推荐国内用户）**
```bash
pip install modelscope
python -c "from modelscope import snapshot_download; snapshot_download('IndexTeam/IndexTTS-1.5', local_dir='./backend/checkpoints')"
```

**方法 B：使用 Hugging Face**
```bash
pip install huggingface_hub
python -c "from huggingface_hub import snapshot_download; snapshot_download('IndexTeam/IndexTTS-1.5', local_dir='./backend/checkpoints')"
```

下载后目录结构：
```
backend/checkpoints/
├── config.yaml
├── bigvgan_discriminator.safetensors
├── bigvgan_generator.safetensors
├── bpe.model
├── dvae_encoder.safetensors
├── gpt.safetensors
└── ...
```

### 步骤 4：准备预置音色

将 6 个参考音频文件放入 `backend/serviceData/index_tts/voices/` 目录：

| 文件名 | 描述 | 要求 |
|--------|------|------|
| `male_youth.wav` | 青年男声 | WAV格式，5-15秒，清晰无噪音 |
| `male_mature.wav` | 成熟男声 | 同上 |
| `female_sweet.wav` | 甜美女声 | 同上 |
| `female_elegant.wav` | 知性女声 | 同上 |
| `female_lively.wav` | 活泼女声 | 同上 |
| `child.wav` | 童声 | 同上 |

**音频录制建议：**
- 采样率：44.1kHz 或更高
- 位深：16bit 或 24bit
- 环境：安静无回声
- 内容：正常语速的中文句子

---

## 三、使用方法

### 1. 启动应用

```bash
# 终端 1：启动后端
cd backend
python main.py

# 终端 2：启动前端
cd frontend
npm run dev
```

### 2. 配置 TTS 引擎

1. 打开应用 → 设置 → TTS 设置
2. 选择引擎：**IndexTTS2 本地推理**
3. 检查模型状态：
   - 「模型文件」应显示「已就绪」
   - 点击「预加载模型」（首次约 30-60 秒）
4. 选择音色：在音色库中点击「设为当前」
5. 配置情感：
   - **自动**：根据文本智能推断
   - **手动**：固定情感类型
   - **禁用**：中性语调

### 3. 生成配音

正常使用视频生成功能，系统会自动调用 IndexTTS2 进行配音。

---

## 四、情感控制说明

| 情感 | 适用场景 |
|------|----------|
| 开心 | 娱乐、搞笑内容 |
| 悲伤 | 感人、煽情片段 |
| 愤怒 | 激烈、冲突场景 |
| 平静 | 解说、旁白 |
| 惊讶 | 反转、意外情节 |
| 恐惧 | 悬疑、恐怖内容 |

**情感强度建议：** 50-70%（过高可能不自然）

---

## 五、常见问题

### Q1: 提示「模型文件未找到」
检查 `backend/checkpoints/` 目录是否包含 `config.yaml` 和模型文件。

### Q2: 显存不足 (CUDA OOM)
- 关闭其他占用显存的程序
- 确保 GPU 至少有 8GB 显存
- 可在配置中启用 FP16 模式节省显存

### Q3: 音色效果不理想
- 更换更清晰的参考音频
- 参考音频时长建议 5-15 秒
- 确保参考音频无背景音乐/噪音

### Q4: 首次合成很慢
正常现象，首次加载模型需要 30-60 秒，之后会常驻显存。

### Q5: 如何切换回云端 TTS
在 TTS 设置中选择「Edge TTS」或「腾讯云 TTS」即可。

---

## 六、技术支持

如有问题，请联系开发者或查阅：
- IndexTTS 官方文档：https://github.com/index-tts/index-tts
- 项目 Issues：[项目仓库地址]

---

*文档版本：1.0.0 | 更新日期：2025-12-27*

