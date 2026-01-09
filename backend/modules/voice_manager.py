#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
音色管理器
提供音色的增删改查、分类管理、音频验证等功能
"""

import json
import shutil
import subprocess
import uuid
import logging
from pathlib import Path
from datetime import datetime
from typing import List, Optional, Dict, Any, Tuple
from threading import Lock
import asyncio

logger = logging.getLogger(__name__)

# 数据目录
DATA_DIR = Path(__file__).resolve().parent.parent / "serviceData" / "index_tts"
VOICES_DIR = DATA_DIR / "voices"
PRESET_DIR = VOICES_DIR / "preset"
CUSTOM_DIR = VOICES_DIR / "custom"
PREVIEWS_DIR = DATA_DIR / "previews"
META_FILE = DATA_DIR / "voices_meta.json"

# 确保目录存在
for d in [DATA_DIR, VOICES_DIR, PRESET_DIR, CUSTOM_DIR, PREVIEWS_DIR]:
    d.mkdir(parents=True, exist_ok=True)

# 支持的音频格式
SUPPORTED_FORMATS = {".wav", ".mp3", ".webm", ".ogg", ".m4a"}
MAX_FILE_SIZE = 20 * 1024 * 1024  # 20MB
MIN_DURATION = 3.0  # 秒
MAX_DURATION = 30.0  # 秒


class VoiceManager:
    """音色管理器单例"""
    
    _instance = None
    _lock = Lock()
    
    def __new__(cls):
        if cls._instance is None:
            with cls._lock:
                if cls._instance is None:
                    cls._instance = super().__new__(cls)
                    cls._instance._initialized = False
        return cls._instance
    
    def __init__(self):
        if self._initialized:
            return
        self._initialized = True
        self._file_lock = Lock()
        self._ensure_meta_file()
    
    def _ensure_meta_file(self):
        """确保元数据文件存在"""
        if not META_FILE.exists():
            default_meta = {
                "version": "2.0.0",
                "categories": [
                    {"id": "male", "name": "男声", "icon": "👨", "is_default": True},
                    {"id": "female", "name": "女声", "icon": "👩", "is_default": True},
                    {"id": "child", "name": "童声", "icon": "👶", "is_default": True}
                ],
                "voices": []
            }
            self._save_meta(default_meta)
    
    def _load_meta(self) -> Dict[str, Any]:
        """加载元数据"""
        try:
            with open(META_FILE, "r", encoding="utf-8") as f:
                return json.load(f)
        except Exception as e:
            logger.error(f"加载音色元数据失败: {e}")
            return {"version": "2.0.0", "categories": [], "voices": []}
    
    def _save_meta(self, data: Dict[str, Any]) -> bool:
        """保存元数据（原子操作）"""
        try:
            with self._file_lock:
                temp_file = META_FILE.with_suffix(".tmp")
                with open(temp_file, "w", encoding="utf-8") as f:
                    json.dump(data, f, ensure_ascii=False, indent=2)
                temp_file.replace(META_FILE)
            return True
        except Exception as e:
            logger.error(f"保存音色元数据失败: {e}")
            return False
    
    # ========== 分类管理 ==========
    
    def get_categories(self) -> List[Dict[str, Any]]:
        """获取所有分类"""
        meta = self._load_meta()
        return meta.get("categories", [])
    
    def add_category(self, name: str, icon: str = "🎤") -> Dict[str, Any]:
        """添加分类"""
        meta = self._load_meta()
        
        # 生成唯一ID
        category_id = f"cat_{uuid.uuid4().hex[:8]}"
        
        new_category = {
            "id": category_id,
            "name": name,
            "icon": icon,
            "is_default": False
        }
        
        meta["categories"].append(new_category)
        self._save_meta(meta)
        
        logger.info(f"添加分类: {name} ({category_id})")
        return new_category
    
    def update_category(self, category_id: str, name: str = None, icon: str = None) -> Optional[Dict[str, Any]]:
        """更新分类"""
        meta = self._load_meta()
        
        for cat in meta["categories"]:
            if cat["id"] == category_id:
                if name is not None:
                    cat["name"] = name
                if icon is not None:
                    cat["icon"] = icon
                self._save_meta(meta)
                logger.info(f"更新分类: {category_id}")
                return cat
        
        return None
    
    def delete_category(self, category_id: str) -> Tuple[bool, str]:
        """删除分类"""
        meta = self._load_meta()
        
        # 检查是否为默认分类
        for cat in meta["categories"]:
            if cat["id"] == category_id:
                if cat.get("is_default", False):
                    return False, "默认分类不可删除"
                break
        else:
            return False, "分类不存在"
        
        # 检查是否有音色使用此分类
        voices_in_category = [v for v in meta["voices"] if v.get("category") == category_id]
        if voices_in_category:
            return False, f"该分类下还有 {len(voices_in_category)} 个音色，请先移动或删除"
        
        meta["categories"] = [c for c in meta["categories"] if c["id"] != category_id]
        self._save_meta(meta)
        
        logger.info(f"删除分类: {category_id}")
        return True, "删除成功"
    
    # ========== 音色管理 ==========
    
    def get_voices(self, category: str = None) -> List[Dict[str, Any]]:
        """获取音色列表"""
        meta = self._load_meta()
        voices = meta.get("voices", [])
        
        if category:
            voices = [v for v in voices if v.get("category") == category]
        
        return voices
    
    def get_voice(self, voice_id: str) -> Optional[Dict[str, Any]]:
        """获取单个音色"""
        meta = self._load_meta()
        for v in meta.get("voices", []):
            if v["id"] == voice_id:
                return v
        return None
    
    def add_voice(
        self,
        name: str,
        category: str,
        audio_data: bytes,
        filename: str,
        description: str = ""
    ) -> Tuple[Optional[Dict[str, Any]], str]:
        """添加音色"""
        
        # 验证分类存在
        meta = self._load_meta()
        if not any(c["id"] == category for c in meta["categories"]):
            return None, f"分类 '{category}' 不存在"
        
        # 获取文件扩展名
        ext = Path(filename).suffix.lower()
        if ext not in SUPPORTED_FORMATS:
            return None, f"不支持的音频格式: {ext}"
        
        # 检查文件大小
        if len(audio_data) > MAX_FILE_SIZE:
            return None, f"文件过大，最大支持 {MAX_FILE_SIZE // 1024 // 1024}MB"
        
        # 生成唯一ID和文件名
        voice_id = f"custom_{datetime.now().strftime('%Y%m%d_%H%M%S')}_{uuid.uuid4().hex[:6]}"
        audio_filename = f"{voice_id}{ext}"
        audio_path = CUSTOM_DIR / audio_filename
        
        try:
            # 保存音频文件
            with open(audio_path, "wb") as f:
                f.write(audio_data)
            
            # 获取音频时长
            duration = self._get_audio_duration(audio_path)
            if duration is None:
                audio_path.unlink(missing_ok=True)
                return None, "无法读取音频时长，请检查文件格式"
            
            if duration < MIN_DURATION:
                audio_path.unlink(missing_ok=True)
                return None, f"音频过短，最少需要 {MIN_DURATION} 秒"
            
            if duration > MAX_DURATION:
                audio_path.unlink(missing_ok=True)
                return None, f"音频过长，最大支持 {MAX_DURATION} 秒"
            
            # 创建音色记录
            new_voice = {
                "id": voice_id,
                "name": name,
                "category": category,
                "is_preset": False,
                "audio_file": f"custom/{audio_filename}",
                "description": description,
                "duration": round(duration, 2),
                "created_at": datetime.now().isoformat()
            }
            
            meta["voices"].append(new_voice)
            self._save_meta(meta)
            
            logger.info(f"添加音色: {name} ({voice_id})")
            return new_voice, "添加成功"
            
        except Exception as e:
            audio_path.unlink(missing_ok=True)
            logger.error(f"添加音色失败: {e}")
            return None, f"添加失败: {str(e)}"
    
    def update_voice(self, voice_id: str, updates: Dict[str, Any]) -> Tuple[Optional[Dict[str, Any]], str]:
        """更新音色信息"""
        meta = self._load_meta()
        
        for voice in meta["voices"]:
            if voice["id"] == voice_id:
                # 预置音色只能修改部分字段
                if voice.get("is_preset", False):
                    allowed_fields = {"name", "description"}
                    updates = {k: v for k, v in updates.items() if k in allowed_fields}
                
                # 如果更改分类，检查分类是否存在
                if "category" in updates:
                    if not any(c["id"] == updates["category"] for c in meta["categories"]):
                        return None, f"分类 '{updates['category']}' 不存在"
                
                voice.update(updates)
                self._save_meta(meta)
                
                logger.info(f"更新音色: {voice_id}")
                return voice, "更新成功"
        
        return None, "音色不存在"
    
    def delete_voice(self, voice_id: str) -> Tuple[bool, str]:
        """删除音色"""
        meta = self._load_meta()
        
        for i, voice in enumerate(meta["voices"]):
            if voice["id"] == voice_id:
                # 预置音色不可删除
                if voice.get("is_preset", False):
                    return False, "预置音色不可删除"
                
                # 删除音频文件
                audio_path = VOICES_DIR / voice["audio_file"]
                if audio_path.exists():
                    audio_path.unlink()
                
                # 删除预览缓存
                for preview in PREVIEWS_DIR.glob(f"{voice_id}_*"):
                    preview.unlink(missing_ok=True)
                
                # 从列表中移除
                meta["voices"].pop(i)
                self._save_meta(meta)
                
                logger.info(f"删除音色: {voice_id}")
                return True, "删除成功"
        
        return False, "音色不存在"
    
    # ========== 音频处理 ==========
    
    def get_audio_path(self, voice_id: str) -> Optional[Path]:
        """获取音频文件路径"""
        voice = self.get_voice(voice_id)
        if voice:
            return VOICES_DIR / voice["audio_file"]
        return None
    
    def _get_audio_duration(self, file_path: Path) -> Optional[float]:
        """获取音频时长（秒）"""
        try:
            # 尝试使用 ffprobe
            result = subprocess.run(
                [
                    "ffprobe", "-v", "quiet", "-show_entries",
                    "format=duration", "-of", "csv=p=0", str(file_path)
                ],
                capture_output=True,
                text=True,
                timeout=10
            )
            if result.returncode == 0 and result.stdout.strip():
                return float(result.stdout.strip())
        except Exception as e:
            logger.warning(f"ffprobe 获取时长失败: {e}")
        
        # 回退：使用 librosa
        try:
            import librosa
            duration = librosa.get_duration(path=str(file_path))
            return duration
        except Exception as e:
            logger.warning(f"librosa 获取时长失败: {e}")
        
        return None
    
    # ========== 试听功能 ==========
    
    async def generate_preview(self, voice_id: str, text: str = None) -> Tuple[Optional[Path], str]:
        """生成试听音频"""
        voice = self.get_voice(voice_id)
        if not voice:
            return None, "音色不存在"
        
        # 默认试听文本
        if not text:
            text = "大家好，欢迎使用智能配音系统，这是音色试听效果。"
        
        # 缓存文件名（基于音色ID和文本hash）
        text_hash = hash(text) & 0xFFFFFFFF
        preview_filename = f"{voice_id}_{text_hash}.wav"
        preview_path = PREVIEWS_DIR / preview_filename
        
        # 如果缓存存在，直接返回
        if preview_path.exists():
            return preview_path, "使用缓存"
        
        # 调用 IndexTTS 生成
        try:
            from modules.index_tts_service import index_tts_service
            
            audio_path = self.get_audio_path(voice_id)
            if not audio_path or not audio_path.exists():
                return None, "音色音频文件不存在"
            
            result = await index_tts_service.synthesize(
                text=text,
                out_path=preview_path,
                voice_id=voice_id,
                reference_audio=str(audio_path)
            )
            
            if result.get("success"):
                return preview_path, "生成成功"
            else:
                return None, result.get("error", "生成失败")
                
        except Exception as e:
            logger.error(f"生成试听失败: {e}")
            return None, str(e)


# 单例实例
voice_manager = VoiceManager()

