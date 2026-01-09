#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
音色管理 API 路由
提供音色上传、编辑、删除、分类管理等接口
"""

from typing import Optional, List
from pathlib import Path
from fastapi import APIRouter, HTTPException, UploadFile, File, Form, Query
from fastapi.responses import FileResponse
from pydantic import BaseModel, Field
import logging

from modules.voice_manager import voice_manager, VOICES_DIR, PREVIEWS_DIR

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/voices", tags=["音色管理"])

_AUDIO_MEDIA_TYPES = {
    ".wav": "audio/wav",
    ".mp3": "audio/mpeg",
    ".m4a": "audio/mp4",
    ".ogg": "audio/ogg",
    ".webm": "audio/webm",
}

def _guess_audio_media_type(path: Path) -> str:
    return _AUDIO_MEDIA_TYPES.get(path.suffix.lower(), "application/octet-stream")


# ========== 请求/响应模型 ==========

class CategoryCreateRequest(BaseModel):
    name: str = Field(..., min_length=1, max_length=20, description="分类名称")
    icon: str = Field(default="🎤", max_length=10, description="分类图标")


class CategoryUpdateRequest(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=20)
    icon: Optional[str] = Field(None, max_length=10)


class VoiceUpdateRequest(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=50)
    category: Optional[str] = Field(None)
    description: Optional[str] = Field(None, max_length=200)


class PreviewRequest(BaseModel):
    text: Optional[str] = Field(None, max_length=200, description="试听文本")


# ========== 分类接口 ==========

@router.get("/categories", summary="获取分类列表")
async def get_categories():
    """获取所有音色分类"""
    try:
        categories = voice_manager.get_categories()
        return {
            "success": True,
            "data": categories,
            "message": f"获取到 {len(categories)} 个分类"
        }
    except Exception as e:
        logger.error(f"获取分类列表失败: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/categories", summary="创建分类")
async def create_category(req: CategoryCreateRequest):
    """创建新的音色分类"""
    try:
        category = voice_manager.add_category(req.name, req.icon)
        return {
            "success": True,
            "data": category,
            "message": "分类创建成功"
        }
    except Exception as e:
        logger.error(f"创建分类失败: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.patch("/categories/{category_id}", summary="更新分类")
async def update_category(category_id: str, req: CategoryUpdateRequest):
    """更新分类信息"""
    try:
        category = voice_manager.update_category(
            category_id,
            name=req.name,
            icon=req.icon
        )
        if category:
            return {
                "success": True,
                "data": category,
                "message": "分类更新成功"
            }
        else:
            raise HTTPException(status_code=404, detail="分类不存在")
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"更新分类失败: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.delete("/categories/{category_id}", summary="删除分类")
async def delete_category(category_id: str):
    """删除分类（默认分类和有音色的分类不可删除）"""
    try:
        success, message = voice_manager.delete_category(category_id)
        if success:
            return {"success": True, "message": message}
        else:
            raise HTTPException(status_code=400, detail=message)
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"删除分类失败: {e}")
        raise HTTPException(status_code=500, detail=str(e))


# ========== 音色接口 ==========

@router.get("", summary="获取音色列表")
async def get_voices(category: Optional[str] = Query(None, description="按分类筛选")):
    """获取音色列表，可按分类筛选"""
    try:
        voices = voice_manager.get_voices(category)
        return {
            "success": True,
            "data": voices,
            "message": f"获取到 {len(voices)} 个音色"
        }
    except Exception as e:
        logger.error(f"获取音色列表失败: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/{voice_id}", summary="获取音色详情")
async def get_voice(voice_id: str):
    """获取单个音色详情"""
    try:
        voice = voice_manager.get_voice(voice_id)
        if voice:
            return {"success": True, "data": voice}
        else:
            raise HTTPException(status_code=404, detail="音色不存在")
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"获取音色详情失败: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/upload", summary="上传音色")
async def upload_voice(
    name: str = Form(..., min_length=1, max_length=50, description="音色名称"),
    category: str = Form(..., description="分类ID"),
    description: str = Form(default="", max_length=200, description="音色描述"),
    audio: UploadFile = File(..., description="音频文件")
):
    """
    上传自定义音色
    
    - 支持格式：WAV, MP3, WebM, OGG, M4A
    - 文件大小：最大 20MB
    - 音频时长：3-30 秒
    """
    try:
        # 读取文件内容
        audio_data = await audio.read()
        
        voice, message = voice_manager.add_voice(
            name=name,
            category=category,
            audio_data=audio_data,
            filename=audio.filename or "audio.wav",
            description=description
        )
        
        if voice:
            return {
                "success": True,
                "data": voice,
                "message": message
            }
        else:
            raise HTTPException(status_code=400, detail=message)
            
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"上传音色失败: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/record", summary="上传录音")
async def upload_recording(
    name: str = Form(..., min_length=1, max_length=50),
    category: str = Form(...),
    description: str = Form(default=""),
    audio: UploadFile = File(...)
):
    """
    上传浏览器录制的音频
    
    与 /upload 相同，但语义上用于录音功能
    """
    return await upload_voice(name, category, description, audio)


@router.patch("/{voice_id}", summary="更新音色")
async def update_voice(voice_id: str, req: VoiceUpdateRequest):
    """更新音色信息（名称、分类、描述）"""
    try:
        updates = req.dict(exclude_none=True)
        if not updates:
            raise HTTPException(status_code=400, detail="没有要更新的字段")
        
        voice, message = voice_manager.update_voice(voice_id, updates)
        
        if voice:
            return {
                "success": True,
                "data": voice,
                "message": message
            }
        else:
            raise HTTPException(status_code=400, detail=message)
            
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"更新音色失败: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.delete("/{voice_id}", summary="删除音色")
async def delete_voice(voice_id: str):
    """删除自定义音色（预置音色不可删除）"""
    try:
        success, message = voice_manager.delete_voice(voice_id)
        if success:
            return {"success": True, "message": message}
        else:
            raise HTTPException(status_code=400, detail=message)
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"删除音色失败: {e}")
        raise HTTPException(status_code=500, detail=str(e))


# ========== 音频文件接口 ==========

@router.get("/{voice_id}/audio", summary="获取音色音频")
async def get_voice_audio(voice_id: str):
    """获取音色的参考音频文件"""
    try:
        audio_path = voice_manager.get_audio_path(voice_id)
        if audio_path and audio_path.exists():
            return FileResponse(
                audio_path,
                media_type=_guess_audio_media_type(audio_path),
                filename=audio_path.name
            )
        else:
            raise HTTPException(status_code=404, detail="音频文件不存在")
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"获取音频文件失败: {e}")
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/{voice_id}/preview", summary="生成试听")
async def generate_preview(voice_id: str, req: PreviewRequest = None):
    """
    生成音色试听音频
    
    使用 IndexTTS 合成指定文本，返回音频文件
    """
    try:
        text = req.text if req else None
        preview_path, message = await voice_manager.generate_preview(voice_id, text)
        
        if preview_path and preview_path.exists():
            return FileResponse(
                preview_path,
                media_type="audio/wav",
                filename=f"preview_{voice_id}.wav"
            )
        else:
            raise HTTPException(status_code=500, detail=message)
            
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"生成试听失败: {e}")
        raise HTTPException(status_code=500, detail=str(e))


# ========== 静态文件服务 ==========

@router.get("/audio/{subpath:path}", summary="音频静态文件")
async def serve_audio(subpath: str):
    """提供音色音频文件的静态访问"""
    try:
        file_path = VOICES_DIR / subpath
        if file_path.exists() and file_path.is_file():
            # 安全检查：确保路径在 VOICES_DIR 内
            if VOICES_DIR in file_path.resolve().parents or file_path.resolve().parent == VOICES_DIR:
                return FileResponse(file_path)
        raise HTTPException(status_code=404, detail="文件不存在")
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"提供音频文件失败: {e}")
        raise HTTPException(status_code=500, detail=str(e))
