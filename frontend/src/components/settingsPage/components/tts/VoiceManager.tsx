import React, { useState, useEffect, useCallback } from "react";
import { Plus, Settings2, Volume2, RefreshCw } from "lucide-react";
import type { CustomVoice, VoiceCategory, VoiceUpdateRequest } from "../../types";
import { voiceService } from "@/services/ttsService";
import { apiClient } from "@/services/clients";
import { VoiceCard } from "./VoiceCard";
import { VoiceUploadModal } from "./VoiceUploadModal";
import { VoiceEditModal } from "./VoiceEditModal";
import { CategoryManager } from "./CategoryManager";

interface VoiceManagerProps {
  activeVoiceId: string | null;
  onSelectVoice: (voiceId: string) => void;
}

export const VoiceManager: React.FC<VoiceManagerProps> = ({
  activeVoiceId,
  onSelectVoice,
}) => {
  // 状态
  const [voices, setVoices] = useState<CustomVoice[]>([]);
  const [categories, setCategories] = useState<VoiceCategory[]>([]);
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 弹窗状态
  const [showUploadModal, setShowUploadModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showCategoryManager, setShowCategoryManager] = useState(false);
  const [editingVoice, setEditingVoice] = useState<CustomVoice | null>(null);

  // 加载数据
  const loadData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [catRes, voiceRes] = await Promise.all([
        voiceService.getCategories(),
        voiceService.getCustomVoices(selectedCategory || undefined),
      ]);

      if (catRes?.success) {
        setCategories(catRes.data || []);
      }
      if (voiceRes?.success) {
        setVoices(voiceRes.data || []);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载失败");
    } finally {
      setLoading(false);
    }
  }, [selectedCategory]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // 处理分类筛选
  const handleCategoryFilter = (categoryId: string | null) => {
    setSelectedCategory(categoryId);
  };

  // 处理上传
  const handleUpload = async (data: {
    name: string;
    category: string;
    description: string;
    audio: File;
  }) => {
    await voiceService.uploadVoice(data);
    await loadData();
  };

  // 处理编辑
  const handleEdit = (voice: CustomVoice) => {
    setEditingVoice(voice);
    setShowEditModal(true);
  };

  const handleSaveEdit = async (voiceId: string, data: VoiceUpdateRequest) => {
    await voiceService.updateVoice(voiceId, data);
    await loadData();
  };

  // 处理删除
  const handleDelete = async (voiceId: string) => {
    try {
      await voiceService.deleteVoice(voiceId);
      await loadData();
    } catch (err) {
      alert(err instanceof Error ? err.message : "删除失败");
    }
  };

  // 处理试听
  const handlePlayPreview = async (voiceId: string): Promise<Blob | null> => {
    try {
      return await voiceService.generatePreview(voiceId);
    } catch (err) {
      console.error("生成试听失败:", err);
      return null;
    }
  };

  // 分类管理
  const handleAddCategory = async (name: string, icon: string) => {
    await voiceService.createCategory(name, icon);
    await loadData();
  };

  const handleUpdateCategory = async (id: string, name: string, icon: string) => {
    await voiceService.updateCategory(id, name, icon);
    await loadData();
  };

  const handleDeleteCategory = async (id: string) => {
    await voiceService.deleteCategory(id);
    await loadData();
  };

  // 过滤后的音色
  const filteredVoices = selectedCategory
    ? voices.filter((v) => v.category === selectedCategory)
    : voices;

  return (
    <div className="space-y-4">
      {/* 头部 */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Volume2 className="w-5 h-5 text-purple-600" />
          <h4 className="font-medium text-gray-900">音色库</h4>
          <span className="text-xs text-gray-500">({voices.length})</span>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => loadData()}
            disabled={loading}
            className="p-2 text-gray-500 hover:text-purple-600 hover:bg-purple-50 rounded-lg transition-colors"
            title="刷新"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
          </button>
          <button
            onClick={() => setShowCategoryManager(!showCategoryManager)}
            className={`p-2 rounded-lg transition-colors ${
              showCategoryManager
                ? "text-purple-600 bg-purple-100"
                : "text-gray-500 hover:text-purple-600 hover:bg-purple-50"
            }`}
            title="分类管理"
          >
            <Settings2 className="w-4 h-4" />
          </button>
          <button
            onClick={() => setShowUploadModal(true)}
            className="flex items-center gap-1.5 px-3 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors text-sm font-medium"
          >
            <Plus className="w-4 h-4" />
            上传音色
          </button>
        </div>
      </div>

      {/* 分类管理面板 */}
      {showCategoryManager && (
        <div className="p-4 bg-white border rounded-xl">
          <CategoryManager
            categories={categories}
            onAdd={handleAddCategory}
            onUpdate={handleUpdateCategory}
            onDelete={handleDeleteCategory}
          />
        </div>
      )}

      {/* 分类筛选 */}
      <div className="flex flex-wrap gap-2">
        <button
          onClick={() => handleCategoryFilter(null)}
          className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
            selectedCategory === null
              ? "bg-purple-500 text-white"
              : "bg-gray-100 text-gray-700 hover:bg-gray-200"
          }`}
        >
          全部
        </button>
        {categories.map((cat) => (
          <button
            key={cat.id}
            onClick={() => handleCategoryFilter(cat.id)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              selectedCategory === cat.id
                ? "bg-purple-500 text-white"
                : "bg-gray-100 text-gray-700 hover:bg-gray-200"
            }`}
          >
            {cat.icon} {cat.name}
          </button>
        ))}
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="p-3 bg-red-50 text-red-600 rounded-lg text-sm">{error}</div>
      )}

      {/* 音色列表 */}
      {loading ? (
        <div className="flex items-center justify-center py-12">
          <div className="w-8 h-8 border-4 border-purple-200 border-t-purple-600 rounded-full animate-spin" />
        </div>
      ) : filteredVoices.length === 0 ? (
        <div className="text-center py-12 text-gray-500">
          <Volume2 className="w-12 h-12 mx-auto mb-3 text-gray-300" />
          <p>暂无音色</p>
          <p className="text-sm mt-1">点击「上传音色」添加自定义音色</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredVoices.map((voice) => (
            <VoiceCard
              key={voice.id}
              voice={voice}
              categories={categories}
              isActive={activeVoiceId === voice.id}
              onSelect={onSelectVoice}
              onEdit={handleEdit}
              onDelete={handleDelete}
              onPlayPreview={handlePlayPreview}
              audioBaseUrl={apiClient.getBaseUrl()}
            />
          ))}
        </div>
      )}

      {/* 上传弹窗 */}
      <VoiceUploadModal
        isOpen={showUploadModal}
        onClose={() => setShowUploadModal(false)}
        onUpload={handleUpload}
        categories={categories}
      />

      {/* 编辑弹窗 */}
      <VoiceEditModal
        isOpen={showEditModal}
        onClose={() => {
          setShowEditModal(false);
          setEditingVoice(null);
        }}
        onSave={handleSaveEdit}
        voice={editingVoice}
        categories={categories}
      />
    </div>
  );
};

export default VoiceManager;

