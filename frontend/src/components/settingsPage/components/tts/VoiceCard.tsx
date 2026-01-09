import React, { useState, useRef } from "react";
import { Play, Pause, Trash2, Edit2, Check, Volume2 } from "lucide-react";
import type { CustomVoice, VoiceCategory } from "../../types";

interface VoiceCardProps {
  voice: CustomVoice;
  categories: VoiceCategory[];
  isActive: boolean;
  onSelect: (voiceId: string) => void;
  onEdit: (voice: CustomVoice) => void;
  onDelete: (voiceId: string) => void;
  onPlayPreview: (voiceId: string) => Promise<Blob | null>;
  audioBaseUrl: string;
}

export const VoiceCard: React.FC<VoiceCardProps> = ({
  voice,
  categories,
  isActive,
  onSelect,
  onEdit,
  onDelete,
  onPlayPreview,
  audioBaseUrl,
}) => {
  const [isPlaying, setIsPlaying] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const objectUrlRef = useRef<string | null>(null);

  const category = categories.find((c) => c.id === voice.category);

  React.useEffect(() => {
    return () => {
      if (audioRef.current) {
        audioRef.current.pause();
        audioRef.current = null;
      }
      if (objectUrlRef.current) {
        URL.revokeObjectURL(objectUrlRef.current);
        objectUrlRef.current = null;
      }
    };
  }, []);

  const handlePlayPause = async (e: React.MouseEvent) => {
    e.stopPropagation();

    if (isPlaying && audioRef.current) {
      audioRef.current.pause();
      audioRef.current.currentTime = 0;
      if (objectUrlRef.current) {
        URL.revokeObjectURL(objectUrlRef.current);
        objectUrlRef.current = null;
      }
      setIsPlaying(false);
      return;
    }

    setIsLoading(true);
    try {
      // 优先使用带鉴权的预览接口
      const previewBlob = await onPlayPreview(voice.id);
      let audioUrl = "";
      if (previewBlob) {
        if (objectUrlRef.current) {
          URL.revokeObjectURL(objectUrlRef.current);
        }
        const objectUrl = URL.createObjectURL(previewBlob);
        objectUrlRef.current = objectUrl;
        audioUrl = objectUrl;
      } else {
        audioUrl = `${audioBaseUrl}/api/voices/${voice.id}/audio`;
      }
      
      if (audioRef.current) {
        audioRef.current.src = audioUrl;
      } else {
        audioRef.current = new Audio(audioUrl);
      }

      audioRef.current.onended = () => {
        setIsPlaying(false);
        if (objectUrlRef.current) {
          URL.revokeObjectURL(objectUrlRef.current);
          objectUrlRef.current = null;
        }
      };
      audioRef.current.onerror = () => {
        setIsPlaying(false);
        if (objectUrlRef.current) {
          URL.revokeObjectURL(objectUrlRef.current);
          objectUrlRef.current = null;
        }
        setIsLoading(false);
      };

      await audioRef.current.play();
      setIsPlaying(true);
    } catch (error) {
      console.error("播放失败:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleEdit = (e: React.MouseEvent) => {
    e.stopPropagation();
    onEdit(voice);
  };

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (window.confirm(`确定要删除音色「${voice.name}」吗？`)) {
      onDelete(voice.id);
    }
  };

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return mins > 0 ? `${mins}分${secs}秒` : `${secs}秒`;
  };

  return (
    <div
      onClick={() => onSelect(voice.id)}
      className={`
        relative p-4 rounded-xl border-2 cursor-pointer transition-all duration-200
        ${
          isActive
            ? "border-purple-500 bg-purple-50 shadow-md"
            : "border-gray-200 bg-white hover:border-purple-300 hover:shadow-sm"
        }
      `}
    >
      {/* 选中标记 */}
      {isActive && (
        <div className="absolute -top-2 -right-2 w-6 h-6 bg-purple-500 rounded-full flex items-center justify-center">
          <Check className="w-4 h-4 text-white" />
        </div>
      )}

      {/* 头部：图标和名称 */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-2">
          <div className="w-10 h-10 rounded-full bg-gradient-to-br from-purple-400 to-pink-400 flex items-center justify-center">
            <Volume2 className="w-5 h-5 text-white" />
          </div>
          <div>
            <h4 className="font-medium text-gray-900 text-sm">{voice.name}</h4>
            <div className="flex items-center gap-1 mt-0.5">
              <span className="text-xs">{category?.icon || "🎤"}</span>
              <span className="text-xs text-gray-500">{category?.name || "未分类"}</span>
              {voice.is_preset && (
                <span className="text-xs px-1.5 py-0.5 bg-blue-100 text-blue-600 rounded">
                  预置
                </span>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* 描述 */}
      {voice.description && (
        <p className="text-xs text-gray-500 mb-3 line-clamp-2">{voice.description}</p>
      )}

      {/* 信息 */}
      <div className="flex items-center gap-3 text-xs text-gray-400 mb-3">
        <span>时长: {formatDuration(voice.duration)}</span>
      </div>

      {/* 操作按钮 */}
      <div className="flex items-center gap-2">
        {/* 播放按钮 */}
        <button
          onClick={handlePlayPause}
          disabled={isLoading}
          className={`
            flex-1 flex items-center justify-center gap-1.5 py-2 rounded-lg text-sm font-medium transition-colors
            ${
              isPlaying
                ? "bg-purple-100 text-purple-700"
                : "bg-gray-100 text-gray-700 hover:bg-gray-200"
            }
            ${isLoading ? "opacity-50 cursor-wait" : ""}
          `}
        >
          {isLoading ? (
            <div className="w-4 h-4 border-2 border-gray-400 border-t-transparent rounded-full animate-spin" />
          ) : isPlaying ? (
            <Pause className="w-4 h-4" />
          ) : (
            <Play className="w-4 h-4" />
          )}
          {isPlaying ? "暂停" : "试听"}
        </button>

        {/* 编辑按钮 */}
        <button
          onClick={handleEdit}
          className="p-2 text-gray-500 hover:text-purple-600 hover:bg-purple-50 rounded-lg transition-colors"
          title="编辑"
        >
          <Edit2 className="w-4 h-4" />
        </button>

        {/* 删除按钮（预置音色不可删除） */}
        {!voice.is_preset && (
          <button
            onClick={handleDelete}
            className="p-2 text-gray-500 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
            title="删除"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        )}
      </div>
    </div>
  );
};

export default VoiceCard;
