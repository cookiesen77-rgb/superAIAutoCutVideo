#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
IndexTTS2 模型下载与环境配置脚本

运行方式：
    python scripts/setup_indextts.py
"""

import os
import sys
from pathlib import Path

# 项目根目录
ROOT_DIR = Path(__file__).resolve().parent.parent
BACKEND_DIR = ROOT_DIR / "backend"
CHECKPOINTS_DIR = BACKEND_DIR / "checkpoints"
VOICES_DIR = BACKEND_DIR / "serviceData" / "index_tts" / "voices"


def print_header(title: str):
    print("\n" + "=" * 60)
    print(f"  {title}")
    print("=" * 60 + "\n")


def check_gpu():
    """检查 GPU 可用性"""
    print_header("检查 GPU 环境")
    try:
        import torch
        if torch.cuda.is_available():
            gpu_name = torch.cuda.get_device_name(0)
            gpu_mem = torch.cuda.get_device_properties(0).total_memory / (1024**3)
            print(f"✓ GPU 可用: {gpu_name}")
            print(f"✓ 显存大小: {gpu_mem:.1f} GB")
            if gpu_mem < 8:
                print("⚠ 警告: 显存小于 8GB，可能影响推理性能")
            return True
        else:
            print("✗ 未检测到 CUDA GPU")
            print("  IndexTTS2 需要 NVIDIA GPU 才能运行")
            return False
    except ImportError:
        print("✗ PyTorch 未安装，请先运行: pip install torch")
        return False


def download_model():
    """下载 IndexTTS2 模型"""
    print_header("下载 IndexTTS2 模型")
    
    if CHECKPOINTS_DIR.exists() and (CHECKPOINTS_DIR / "config.yaml").exists():
        print("✓ 模型文件已存在，跳过下载")
        return True
    
    CHECKPOINTS_DIR.mkdir(parents=True, exist_ok=True)
    
    # 尝试 ModelScope
    print("尝试从 ModelScope 下载...")
    try:
        from modelscope import snapshot_download
        snapshot_download(
            "IndexTeam/IndexTTS-1.5",
            local_dir=str(CHECKPOINTS_DIR),
        )
        print("✓ 模型下载成功 (ModelScope)")
        return True
    except Exception as e:
        print(f"  ModelScope 下载失败: {e}")
    
    # 尝试 Hugging Face
    print("尝试从 Hugging Face 下载...")
    try:
        from huggingface_hub import snapshot_download
        snapshot_download(
            "IndexTeam/IndexTTS-1.5",
            local_dir=str(CHECKPOINTS_DIR),
        )
        print("✓ 模型下载成功 (Hugging Face)")
        return True
    except Exception as e:
        print(f"  Hugging Face 下载失败: {e}")
    
    print("\n✗ 自动下载失败，请手动下载模型")
    print("  1. 访问 https://modelscope.cn/models/IndexTeam/IndexTTS-1.5")
    print("  2. 下载所有文件到 backend/checkpoints/ 目录")
    return False


def setup_voices_dir():
    """创建音色目录"""
    print_header("设置音色目录")
    
    VOICES_DIR.mkdir(parents=True, exist_ok=True)
    
    required_voices = [
        "male_youth.wav",
        "male_mature.wav", 
        "female_sweet.wav",
        "female_elegant.wav",
        "female_lively.wav",
        "child.wav",
    ]
    
    existing = [f.name for f in VOICES_DIR.glob("*.wav")]
    missing = [v for v in required_voices if v not in existing]
    
    if not missing:
        print("✓ 所有预置音色文件已就绪")
    else:
        print(f"⚠ 缺少 {len(missing)} 个音色文件:")
        for v in missing:
            print(f"    - {v}")
        print(f"\n请将音色文件放入: {VOICES_DIR}")
        print("音频要求: WAV格式，5-15秒，44.1kHz，清晰无噪音")


def install_indextts():
    """安装 indextts 包"""
    print_header("安装 IndexTTS2 包")
    
    try:
        import indextts
        print("✓ indextts 已安装")
        return True
    except ImportError:
        print("正在安装 indextts...")
        ret = os.system("pip install git+https://github.com/index-tts/index-tts.git")
        if ret == 0:
            print("✓ indextts 安装成功")
            return True
        else:
            print("✗ indextts 安装失败")
            print("  请手动运行: pip install git+https://github.com/index-tts/index-tts.git")
            return False


def main():
    print("\n" + "=" * 60)
    print("       IndexTTS2 环境配置脚本")
    print("=" * 60)
    
    # 1. 检查 GPU
    gpu_ok = check_gpu()
    
    # 2. 安装 indextts
    pkg_ok = install_indextts()
    
    # 3. 下载模型
    model_ok = download_model()
    
    # 4. 设置音色目录
    setup_voices_dir()
    
    # 总结
    print_header("配置完成")
    print(f"GPU 检测:    {'✓' if gpu_ok else '✗'}")
    print(f"包安装:      {'✓' if pkg_ok else '✗'}")
    print(f"模型下载:    {'✓' if model_ok else '✗'}")
    print(f"\n音色目录: {VOICES_DIR}")
    print(f"模型目录: {CHECKPOINTS_DIR}")
    
    if gpu_ok and pkg_ok and model_ok:
        print("\n✓ IndexTTS2 配置完成！可以开始使用了。")
    else:
        print("\n⚠ 部分配置未完成，请根据提示手动处理。")


if __name__ == "__main__":
    main()

