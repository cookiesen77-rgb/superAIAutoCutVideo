#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
IndexTTS2 本地推理服务模块

提供零样本语音克隆和情感控制功能，采用懒加载单例模式管理模型。
"""

import asyncio
import json
import logging
import os
import threading
from pathlib import Path
from typing import Any, Dict, List, Optional

logger = logging.getLogger(__name__)

# 基础路径
BASE_DIR = Path(__file__).resolve().parent.parent  # backend/
SERVICE_DATA_DIR = BASE_DIR / "serviceData" / "index_tts"
VOICES_DIR = SERVICE_DATA_DIR / "voices"
VOICES_META_PATH = SERVICE_DATA_DIR / "voices_meta.json"
CHECKPOINTS_DIR = BASE_DIR / "checkpoints"


# 情感向量映射（8维：happy, angry, sad, afraid, disgusted, melancholic, surprised, calm）
EMOTION_VECTORS: Dict[str, List[float]] = {
    "happy": [1.0, 0, 0, 0, 0, 0, 0, 0],
    "angry": [0, 1.0, 0, 0, 0, 0, 0, 0],
    "sad": [0, 0, 1.0, 0, 0, 0, 0, 0],
    "afraid": [0, 0, 0, 1.0, 0, 0, 0, 0],
    "disgusted": [0, 0, 0, 0, 1.0, 0, 0, 0],
    "melancholic": [0, 0, 0, 0, 0, 1.0, 0, 0],
    "surprised": [0, 0, 0, 0, 0, 0, 1.0, 0],
    "calm": [0, 0, 0, 0, 0, 0, 0, 1.0],
}


async def _ensure_parent_dir(path: Path) -> None:
    """确保父目录存在"""
    path.parent.mkdir(parents=True, exist_ok=True)


async def _ffprobe_duration(path: str) -> Optional[float]:
    """使用 ffprobe 获取音频时长"""
    try:
        cmd = [
            "ffprobe",
            "-v", "error",
            "-select_streams", "a:0",
            "-show_entries", "stream=duration",
            "-of", "default=nk=1:nw=1",
            path,
        ]
        proc = await asyncio.create_subprocess_exec(
            *cmd, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE
        )
        out, _ = await proc.communicate()
        if proc.returncode == 0:
            try:
                return float(out.decode().strip())
            except Exception:
                pass
        # 备用方案：通过 format 获取
        cmd2 = [
            "ffprobe",
            "-v", "error",
            "-show_entries", "format=duration",
            "-of", "default=nk=1:nw=1",
            path,
        ]
        proc2 = await asyncio.create_subprocess_exec(
            *cmd2, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE
        )
        out2, _ = await proc2.communicate()
        if proc2.returncode == 0:
            try:
                return float(out2.decode().strip())
            except Exception:
                return None
        return None
    except Exception:
        return None


class IndexTtsService:
    """IndexTTS2 本地推理服务（懒加载单例模式）"""

    _model: Optional[Any] = None
    _model_lock: threading.Lock = threading.Lock()
    _model_loading: bool = False
    _load_error: Optional[str] = None

    def __init__(
        self,
        model_dir: Optional[str] = None,
        cfg_path: Optional[str] = None,
        use_fp16: bool = True,
    ):
        self.model_dir = Path(model_dir) if model_dir else CHECKPOINTS_DIR
        self.cfg_path = Path(cfg_path) if cfg_path else (self.model_dir / "config.yaml")
        self.use_fp16 = use_fp16
        self._voices_cache: Optional[List[Dict[str, Any]]] = None

    def _check_model_files(self) -> bool:
        """检查模型文件是否存在"""
        if not self.model_dir.exists():
            return False
        if not self.cfg_path.exists():
            return False
        return True

    def _ensure_model_loaded(self) -> Any:
        """确保模型已加载（线程安全）"""
        with self._model_lock:
            if IndexTtsService._model is not None:
                return IndexTtsService._model

            if IndexTtsService._model_loading:
                raise RuntimeError("模型正在加载中，请稍候")

            if not self._check_model_files():
                raise FileNotFoundError(
                    f"IndexTTS2 模型文件未找到，请确保模型已下载到 {self.model_dir}"
                )

            IndexTtsService._model_loading = True
            IndexTtsService._load_error = None

            try:
                logger.info(f"开始加载 IndexTTS2 模型: {self.model_dir}")

                # 动态导入 IndexTTS2
                from indextts.infer_v2 import IndexTTS2

                model = IndexTTS2(
                    cfg_path=str(self.cfg_path),
                    model_dir=str(self.model_dir),
                    use_fp16=self.use_fp16,
                    use_cuda_kernel=False,
                    use_deepspeed=False,
                )

                IndexTtsService._model = model
                logger.info("IndexTTS2 模型加载成功")
                return model

            except Exception as e:
                IndexTtsService._load_error = str(e)
                logger.error(f"IndexTTS2 模型加载失败: {e}")
                raise RuntimeError(f"模型加载失败: {e}")

            finally:
                IndexTtsService._model_loading = False

    def is_model_loaded(self) -> bool:
        """检查模型是否已加载"""
        return IndexTtsService._model is not None

    def is_model_available(self) -> bool:
        """检查模型文件是否可用"""
        return self._check_model_files()

    def get_load_error(self) -> Optional[str]:
        """获取加载错误信息"""
        return IndexTtsService._load_error

    async def preload_model(self) -> Dict[str, Any]:
        """预加载模型到显存"""
        try:
            if IndexTtsService._model is not None:
                return {"success": True, "message": "模型已加载", "loaded": True}

            if IndexTtsService._model_loading:
                return {"success": False, "message": "模型正在加载中", "loading": True}

            # 在线程池中执行加载
            loop = asyncio.get_event_loop()
            await loop.run_in_executor(None, self._ensure_model_loaded)

            return {"success": True, "message": "模型加载成功", "loaded": True}

        except Exception as e:
            return {"success": False, "message": str(e), "error": str(e)}

    def get_available_voices(self) -> List[Dict[str, Any]]:
        """获取预置音色列表"""
        if self._voices_cache is not None:
            return self._voices_cache

        voices: List[Dict[str, Any]] = []

        try:
            if VOICES_META_PATH.exists():
                data = json.loads(VOICES_META_PATH.read_text("utf-8"))
                for item in data.get("voices", []):
                    audio_file = item.get("audio_file", "")
                    audio_path = VOICES_DIR / audio_file
                    voices.append({
                        "id": item.get("id"),
                        "name": item.get("name"),
                        "description": item.get("description"),
                        "gender": item.get("gender"),
                        "tags": item.get("tags", []),
                        "audio_file": audio_file,
                        "audio_path": str(audio_path),
                        "exists": audio_path.exists(),
                        "sample_wav_url": f"/backend/serviceData/index_tts/voices/{audio_file}",
                    })
        except Exception as e:
            logger.error(f"加载音色元数据失败: {e}")

        self._voices_cache = voices
        return voices

    def get_voice_audio_path(self, voice_id: str) -> Optional[Path]:
        """根据音色ID获取参考音频路径"""
        voices = self.get_available_voices()
        for v in voices:
            if v.get("id") == voice_id:
                audio_path = Path(v.get("audio_path", ""))
                if audio_path.exists():
                    return audio_path
                # 尝试直接构造路径
                audio_file = v.get("audio_file")
                if audio_file:
                    direct_path = VOICES_DIR / audio_file
                    if direct_path.exists():
                        return direct_path
        return None

    async def synthesize(
        self,
        text: str,
        out_path: str,
        voice_id: Optional[str] = None,
        emotion: Optional[str] = None,
        emo_alpha: float = 0.6,
        use_emo_text: bool = False,
    ) -> Dict[str, Any]:
        """
        合成语音

        Args:
            text: 要合成的文本
            out_path: 输出音频路径
            voice_id: 音色ID（对应预置音色）
            emotion: 情感类型（happy/sad/angry/calm 等）
            emo_alpha: 情感强度（0-1）
            use_emo_text: 是否根据文本自动推断情感

        Returns:
            合成结果字典
        """
        try:
            # 获取音色参考音频
            voice_id = voice_id or "female_sweet"
            voice_audio_path = self.get_voice_audio_path(voice_id)

            if voice_audio_path is None or not voice_audio_path.exists():
                return {
                    "success": False,
                    "error": f"音色 '{voice_id}' 的参考音频不存在，请确保音色文件已配置",
                }

            # 确保输出目录存在
            out_path_obj = Path(out_path)
            await _ensure_parent_dir(out_path_obj)

            # 确保模型已加载
            loop = asyncio.get_event_loop()
            model = await loop.run_in_executor(None, self._ensure_model_loaded)

            # 构建推理参数
            infer_kwargs: Dict[str, Any] = {
                "spk_audio_prompt": str(voice_audio_path),
                "text": text,
                "output_path": str(out_path_obj),
                "verbose": False,
            }

            # 情感控制
            if use_emo_text:
                # 自动根据文本推断情感
                infer_kwargs["use_emo_text"] = True
                infer_kwargs["emo_alpha"] = emo_alpha
            elif emotion and emotion in EMOTION_VECTORS:
                # 使用指定情感向量
                infer_kwargs["emo_vector"] = EMOTION_VECTORS[emotion]
                infer_kwargs["use_random"] = False
            # 如果 emotion 为 None 或 disabled，则不添加情感控制参数

            # 执行推理（在线程池中）
            def do_infer():
                model.infer(**infer_kwargs)

            await loop.run_in_executor(None, do_infer)

            # 获取音频时长
            duration = await _ffprobe_duration(str(out_path_obj))

            return {
                "success": True,
                "path": str(out_path_obj),
                "duration": duration,
                "codec": "wav",
                "sample_rate": 44100,
                "voice_id": voice_id,
                "emotion": emotion if not use_emo_text else "auto",
            }

        except FileNotFoundError as e:
            return {"success": False, "error": str(e)}
        except RuntimeError as e:
            return {"success": False, "error": str(e)}
        except Exception as e:
            logger.error(f"IndexTTS2 合成失败: {e}")
            return {"success": False, "error": f"合成失败: {e}"}

    async def test_connection(self) -> Dict[str, Any]:
        """测试 IndexTTS2 可用性"""
        try:
            # 检查模型文件
            if not self._check_model_files():
                return {
                    "success": False,
                    "message": "模型文件未找到",
                    "model_dir": str(self.model_dir),
                    "model_available": False,
                }

            # 检查音色文件
            voices = self.get_available_voices()
            available_voices = [v for v in voices if v.get("exists")]

            if not available_voices:
                return {
                    "success": False,
                    "message": "未找到可用的音色文件",
                    "model_available": True,
                    "voices_count": 0,
                }

            # 检查模型是否已加载
            loaded = self.is_model_loaded()

            return {
                "success": True,
                "message": "IndexTTS2 可用" + ("（模型已加载）" if loaded else "（模型未加载）"),
                "model_available": True,
                "model_loaded": loaded,
                "voices_count": len(available_voices),
            }

        except Exception as e:
            return {
                "success": False,
                "message": f"测试失败: {e}",
                "error": str(e),
            }


# 单例实例
index_tts_service = IndexTtsService()

