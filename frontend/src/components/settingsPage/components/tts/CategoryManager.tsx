import React, { useState } from "react";
import { Plus, Trash2, Edit2, Check, X, FolderOpen } from "lucide-react";
import type { VoiceCategory } from "../../types";

interface CategoryManagerProps {
  categories: VoiceCategory[];
  onAdd: (name: string, icon: string) => Promise<void>;
  onUpdate: (id: string, name: string, icon: string) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
}

const EMOJI_OPTIONS = ["🎤", "👨", "👩", "👶", "🎭", "🎵", "🎧", "📻", "🔊", "💬", "🗣️", "🎙️"];

export const CategoryManager: React.FC<CategoryManagerProps> = ({
  categories,
  onAdd,
  onUpdate,
  onDelete,
}) => {
  const [isAdding, setIsAdding] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [newName, setNewName] = useState("");
  const [newIcon, setNewIcon] = useState("🎤");
  const [editName, setEditName] = useState("");
  const [editIcon, setEditIcon] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleAdd = async () => {
    if (!newName.trim()) {
      setError("请输入分类名称");
      return;
    }

    setLoading(true);
    setError(null);
    try {
      await onAdd(newName.trim(), newIcon);
      setIsAdding(false);
      setNewName("");
      setNewIcon("🎤");
    } catch (err) {
      setError(err instanceof Error ? err.message : "添加失败");
    } finally {
      setLoading(false);
    }
  };

  const handleStartEdit = (category: VoiceCategory) => {
    setEditingId(category.id);
    setEditName(category.name);
    setEditIcon(category.icon);
    setError(null);
  };

  const handleSaveEdit = async () => {
    if (!editingId || !editName.trim()) {
      setError("请输入分类名称");
      return;
    }

    setLoading(true);
    setError(null);
    try {
      await onUpdate(editingId, editName.trim(), editIcon);
      setEditingId(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "更新失败");
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string, name: string) => {
    if (!window.confirm(`确定要删除分类「${name}」吗？`)) return;

    setLoading(true);
    setError(null);
    try {
      await onDelete(id);
    } catch (err) {
      setError(err instanceof Error ? err.message : "删除失败");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-4">
      {/* 标题 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <FolderOpen className="w-5 h-5 text-purple-600" />
          <h4 className="font-medium text-gray-900">分类管理</h4>
        </div>
        {!isAdding && (
          <button
            onClick={() => setIsAdding(true)}
            className="flex items-center gap-1 px-3 py-1.5 text-sm text-purple-600 hover:bg-purple-50 rounded-lg transition-colors"
          >
            <Plus className="w-4 h-4" />
            添加分类
          </button>
        )}
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="p-2 bg-red-50 text-red-600 rounded-lg text-sm">{error}</div>
      )}

      {/* 添加新分类 */}
      {isAdding && (
        <div className="p-4 bg-purple-50 rounded-lg space-y-3">
          <div className="flex gap-2">
            {/* 图标选择 */}
            <div className="relative">
              <button
                className="w-10 h-10 text-xl bg-white border rounded-lg hover:bg-gray-50"
                title="选择图标"
              >
                {newIcon}
              </button>
              <div className="absolute top-full left-0 mt-1 p-2 bg-white border rounded-lg shadow-lg grid grid-cols-4 gap-1 z-10 opacity-0 invisible hover:opacity-100 hover:visible group-hover:opacity-100 group-hover:visible transition-all">
                {EMOJI_OPTIONS.map((emoji) => (
                  <button
                    key={emoji}
                    onClick={() => setNewIcon(emoji)}
                    className={`w-8 h-8 text-lg rounded hover:bg-gray-100 ${
                      newIcon === emoji ? "bg-purple-100" : ""
                    }`}
                  >
                    {emoji}
                  </button>
                ))}
              </div>
            </div>
            {/* 名称输入 */}
            <input
              type="text"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder="分类名称"
              maxLength={20}
              className="flex-1 px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500"
            />
          </div>
          {/* 图标选择区 */}
          <div className="flex flex-wrap gap-1">
            {EMOJI_OPTIONS.map((emoji) => (
              <button
                key={emoji}
                onClick={() => setNewIcon(emoji)}
                className={`w-8 h-8 text-lg rounded hover:bg-white ${
                  newIcon === emoji ? "bg-white ring-2 ring-purple-500" : ""
                }`}
              >
                {emoji}
              </button>
            ))}
          </div>
          {/* 操作按钮 */}
          <div className="flex justify-end gap-2">
            <button
              onClick={() => {
                setIsAdding(false);
                setNewName("");
                setError(null);
              }}
              className="px-3 py-1.5 text-sm text-gray-600 hover:bg-white rounded-lg"
            >
              取消
            </button>
            <button
              onClick={handleAdd}
              disabled={loading || !newName.trim()}
              className="px-3 py-1.5 text-sm bg-purple-600 text-white rounded-lg hover:bg-purple-700 disabled:opacity-50"
            >
              {loading ? "添加中..." : "添加"}
            </button>
          </div>
        </div>
      )}

      {/* 分类列表 */}
      <div className="space-y-2">
        {categories.map((category) => (
          <div
            key={category.id}
            className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
          >
            {editingId === category.id ? (
              // 编辑模式
              <div className="flex-1 flex items-center gap-2">
                <div className="flex flex-wrap gap-1">
                  {EMOJI_OPTIONS.slice(0, 6).map((emoji) => (
                    <button
                      key={emoji}
                      onClick={() => setEditIcon(emoji)}
                      className={`w-7 h-7 text-sm rounded hover:bg-white ${
                        editIcon === emoji ? "bg-white ring-2 ring-purple-500" : ""
                      }`}
                    >
                      {emoji}
                    </button>
                  ))}
                </div>
                <input
                  type="text"
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  className="flex-1 px-2 py-1 border rounded focus:outline-none focus:ring-2 focus:ring-purple-500 text-sm"
                  maxLength={20}
                />
                <button
                  onClick={handleSaveEdit}
                  disabled={loading}
                  className="p-1.5 text-green-600 hover:bg-green-50 rounded"
                >
                  <Check className="w-4 h-4" />
                </button>
                <button
                  onClick={() => setEditingId(null)}
                  className="p-1.5 text-gray-500 hover:bg-gray-200 rounded"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            ) : (
              // 显示模式
              <>
                <div className="flex items-center gap-2">
                  <span className="text-lg">{category.icon}</span>
                  <span className="text-sm font-medium text-gray-900">
                    {category.name}
                  </span>
                  {category.is_default && (
                    <span className="text-xs px-1.5 py-0.5 bg-blue-100 text-blue-600 rounded">
                      默认
                    </span>
                  )}
                </div>
                <div className="flex items-center gap-1">
                  <button
                    onClick={() => handleStartEdit(category)}
                    className="p-1.5 text-gray-500 hover:text-purple-600 hover:bg-purple-50 rounded"
                    title="编辑"
                  >
                    <Edit2 className="w-4 h-4" />
                  </button>
                  {!category.is_default && (
                    <button
                      onClick={() => handleDelete(category.id, category.name)}
                      disabled={loading}
                      className="p-1.5 text-gray-500 hover:text-red-600 hover:bg-red-50 rounded"
                      title="删除"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  )}
                </div>
              </>
            )}
          </div>
        ))}

        {categories.length === 0 && (
          <div className="text-center text-gray-500 py-8">
            暂无分类，点击上方按钮添加
          </div>
        )}
      </div>
    </div>
  );
};

export default CategoryManager;

