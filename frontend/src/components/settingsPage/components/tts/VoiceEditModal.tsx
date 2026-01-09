import React, { useState, useEffect } from "react";
import { createPortal } from "react-dom";
import { X } from "lucide-react";
import type { CustomVoice, VoiceCategory, VoiceUpdateRequest } from "../../types";

interface VoiceEditModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (voiceId: string, data: VoiceUpdateRequest) => Promise<void>;
  voice: CustomVoice | null;
  categories: VoiceCategory[];
}

export const VoiceEditModal: React.FC<VoiceEditModalProps> = ({
  isOpen,
  onClose,
  onSave,
  voice,
  categories,
}) => {
  const [name, setName] = useState("");
  const [category, setCategory] = useState("");
  const [description, setDescription] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 当 voice 变化时，重置表单
  useEffect(() => {
    if (voice) {
      setName(voice.name);
      setCategory(voice.category);
      setDescription(voice.description || "");
      setError(null);
    }
  }, [voice]);

  const handleClose = () => {
    setError(null);
    onClose();
  };

  const handleSubmit = async () => {
    if (!voice) return;

    if (!name.trim()) {
      setError("请输入音色名称");
      return;
    }

    setSaving(true);
    setError(null);

    try {
      const updates: VoiceUpdateRequest = {};

      // 只提交有变化的字段
      if (name.trim() !== voice.name) {
        updates.name = name.trim();
      }
      if (category !== voice.category && !voice.is_preset) {
        updates.category = category;
      }
      if (description.trim() !== (voice.description || "")) {
        updates.description = description.trim();
      }

      if (Object.keys(updates).length === 0) {
        handleClose();
        return;
      }

      await onSave(voice.id, updates);
      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存失败");
    } finally {
      setSaving(false);
    }
  };

  if (!isOpen || !voice) return null;

  const modal = (
    <div className="fixed inset-0 z-[60] flex items-center justify-center">
      {/* 背景遮罩 */}
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={handleClose}
      />

      {/* 弹窗内容 */}
      <div className="relative bg-white rounded-2xl shadow-2xl w-full max-w-md mx-4 max-h-[90vh] flex flex-col">
        {/* 头部 */}
        <div className="flex items-center justify-between p-5 border-b">
          <h3 className="text-lg font-semibold text-gray-900">编辑音色</h3>
          <button
            onClick={handleClose}
            className="p-1 text-gray-400 hover:text-gray-600 rounded-lg hover:bg-gray-100"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* 内容 */}
        <div className="p-5 space-y-4 overflow-y-auto flex-1 min-h-0">
          {/* 预置音色提示 */}
          {voice.is_preset && (
            <div className="p-3 bg-blue-50 text-blue-600 rounded-lg text-sm">
              预置音色仅可编辑名称和描述，不可更改分类
            </div>
          )}

          {/* 名称 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              音色名称 <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="为音色起个名字"
              maxLength={50}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
            />
          </div>

          {/* 分类（预置音色不可修改） */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              分类
            </label>
            <div className="flex flex-wrap gap-2">
              {categories.map((cat) => (
                <button
                  key={cat.id}
                  onClick={() => !voice.is_preset && setCategory(cat.id)}
                  disabled={voice.is_preset}
                  className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                    category === cat.id
                      ? "bg-purple-500 text-white"
                      : voice.is_preset
                      ? "bg-gray-100 text-gray-400 cursor-not-allowed"
                      : "bg-gray-100 text-gray-700 hover:bg-gray-200"
                  }`}
                >
                  {cat.icon} {cat.name}
                </button>
              ))}
            </div>
          </div>

          {/* 描述 */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              描述（可选）
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="描述这个音色的特点..."
              maxLength={200}
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent resize-none"
            />
            <div className="text-xs text-gray-400 mt-1 text-right">
              {description.length}/200
            </div>
          </div>

          {/* 错误提示 */}
          {error && (
            <div className="p-3 bg-red-50 text-red-600 rounded-lg text-sm">
              {error}
            </div>
          )}
        </div>

        {/* 底部按钮 */}
        <div className="flex justify-end gap-3 p-5 border-t bg-gray-50 rounded-b-2xl shrink-0">
          <button
            onClick={handleClose}
            className="px-4 py-2 text-gray-700 hover:bg-gray-200 rounded-lg transition-colors"
          >
            取消
          </button>
          <button
            onClick={handleSubmit}
            disabled={saving || !name.trim()}
            className={`px-6 py-2 rounded-lg font-medium transition-colors ${
              saving || !name.trim()
                ? "bg-gray-300 text-gray-500 cursor-not-allowed"
                : "bg-purple-600 text-white hover:bg-purple-700"
            }`}
          >
            {saving ? (
              <span className="flex items-center gap-2">
                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                保存中...
              </span>
            ) : (
              "保存"
            )}
          </button>
        </div>
      </div>
    </div>
  );

  return createPortal(modal, document.body);
};

export default VoiceEditModal;
