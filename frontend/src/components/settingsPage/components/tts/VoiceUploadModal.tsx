import React, { useState, useRef, useCallback } from "react";
import { createPortal } from "react-dom";
import { X, Upload, Mic, FileAudio } from "lucide-react";
import type { VoiceCategory } from "../../types";
import { VoiceRecorder } from "./VoiceRecorder";

interface VoiceUploadModalProps {
  isOpen: boolean;
  onClose: () => void;
  onUpload: (data: {
    name: string;
    category: string;
    description: string;
    audio: File;
  }) => Promise<void>;
  categories: VoiceCategory[];
}

type UploadMode = "file" | "record";

export const VoiceUploadModal: React.FC<VoiceUploadModalProps> = ({
  isOpen,
  onClose,
  onUpload,
  categories,
}) => {
  const [mode, setMode] = useState<UploadMode>("file");
  const [name, setName] = useState("");
  const [category, setCategory] = useState(categories[0]?.id || "male");
  const [description, setDescription] = useState("");
  const [audioFile, setAudioFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [dragActive, setDragActive] = useState(false);

  const fileInputRef = useRef<HTMLInputElement>(null);

  const resetForm = () => {
    setName("");
    setCategory(categories[0]?.id || "male");
    setDescription("");
    setAudioFile(null);
    setError(null);
    setMode("file");
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  const handleFileChange = (file: File | null) => {
    if (!file) return;

    // 验证文件类型
    const allowedMimeTypes = new Set([
      "audio/wav",
      "audio/x-wav",
      "audio/mp3",
      "audio/mpeg",
      "audio/webm",
      "audio/ogg",
      "audio/m4a",
      "audio/x-m4a",
      "audio/mp4",
    ]);
    const allowedExts = new Set(["wav", "mp3", "webm", "ogg", "m4a"]);
    const mime = (file.type || "").toLowerCase();
    const ext = file.name.split(".").pop()?.toLowerCase() || "";
    const mimeOk =
      (mime && allowedMimeTypes.has(mime)) ||
      (mime.startsWith("audio/") && allowedExts.has(ext));
    const extOk = !mime && allowedExts.has(ext);
    if (!mimeOk && !extOk) {
      setError("不支持的文件格式，请上传 WAV、MP3、WebM、OGG 或 M4A 格式");
      return;
    }

    // 验证文件大小 (20MB)
    if (file.size > 20 * 1024 * 1024) {
      setError("文件过大，最大支持 20MB");
      return;
    }

    setAudioFile(file);
    setError(null);

    // 自动填充名称
    if (!name) {
      const fileName = file.name.replace(/\.[^/.]+$/, "");
      setName(fileName.slice(0, 50));
    }
  };

  const handleDrag = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === "dragenter" || e.type === "dragover") {
      setDragActive(true);
    } else if (e.type === "dragleave") {
      setDragActive(false);
    }
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleFileChange(e.dataTransfer.files[0]);
    }
  }, []);

  const handleRecordingComplete = (blob: Blob) => {
    // 将 Blob 转换为 File
    const file = new File([blob], `recording_${Date.now()}.webm`, {
      type: blob.type,
    });
    setAudioFile(file);
  };

  const handleSubmit = async () => {
    if (!name.trim()) {
      setError("请输入音色名称");
      return;
    }

    if (!audioFile) {
      setError("请上传音频文件或录制音频");
      return;
    }

    setUploading(true);
    setError(null);

    try {
      await onUpload({
        name: name.trim(),
        category,
        description: description.trim(),
        audio: audioFile,
      });
      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "上传失败");
    } finally {
      setUploading(false);
    }
  };

  if (!isOpen) return null;

  const modal = (
    <div className="fixed inset-0 z-[60] flex items-center justify-center">
      {/* 背景遮罩 */}
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={handleClose}
      />

      {/* 弹窗内容 */}
      <div className="relative bg-white rounded-2xl shadow-2xl w-full max-w-lg mx-4 max-h-[90vh] flex flex-col">
        {/* 头部 */}
        <div className="flex items-center justify-between p-5 border-b">
          <h3 className="text-lg font-semibold text-gray-900">上传自定义音色</h3>
          <button
            onClick={handleClose}
            className="p-1 text-gray-400 hover:text-gray-600 rounded-lg hover:bg-gray-100"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* 内容 */}
        <div className="p-5 space-y-5 overflow-y-auto flex-1 min-h-0">
          {/* 模式切换 */}
          <div className="flex gap-2 p-1 bg-gray-100 rounded-lg">
            <button
              onClick={() => setMode("file")}
              className={`flex-1 flex items-center justify-center gap-2 py-2 rounded-md text-sm font-medium transition-colors ${
                mode === "file"
                  ? "bg-white text-purple-600 shadow-sm"
                  : "text-gray-600 hover:text-gray-900"
              }`}
            >
              <Upload className="w-4 h-4" />
              上传文件
            </button>
            <button
              onClick={() => setMode("record")}
              className={`flex-1 flex items-center justify-center gap-2 py-2 rounded-md text-sm font-medium transition-colors ${
                mode === "record"
                  ? "bg-white text-purple-600 shadow-sm"
                  : "text-gray-600 hover:text-gray-900"
              }`}
            >
              <Mic className="w-4 h-4" />
              录制音频
            </button>
          </div>

          {/* 上传区域 */}
          {mode === "file" ? (
            <div
              onDragEnter={handleDrag}
              onDragLeave={handleDrag}
              onDragOver={handleDrag}
              onDrop={handleDrop}
              onClick={() => fileInputRef.current?.click()}
              className={`
                relative border-2 border-dashed rounded-xl p-8 text-center cursor-pointer transition-colors
                ${
                  dragActive
                    ? "border-purple-500 bg-purple-50"
                    : audioFile
                    ? "border-green-300 bg-green-50"
                    : "border-gray-300 hover:border-purple-400 hover:bg-purple-50/50"
                }
              `}
            >
              <input
                ref={fileInputRef}
                type="file"
                accept="audio/*"
                onChange={(e) => handleFileChange(e.target.files?.[0] || null)}
                className="hidden"
              />

              {audioFile ? (
                <div className="flex flex-col items-center gap-2">
                  <FileAudio className="w-12 h-12 text-green-500" />
                  <p className="font-medium text-gray-900">{audioFile.name}</p>
                  <p className="text-sm text-gray-500">
                    {(audioFile.size / 1024 / 1024).toFixed(2)} MB
                  </p>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      setAudioFile(null);
                    }}
                    className="text-sm text-red-500 hover:text-red-600"
                  >
                    移除文件
                  </button>
                </div>
              ) : (
                <div className="flex flex-col items-center gap-2">
                  <Upload className="w-12 h-12 text-gray-400" />
                  <p className="font-medium text-gray-700">
                    拖拽音频文件到这里，或点击选择
                  </p>
                  <p className="text-sm text-gray-500">
                    支持 WAV、MP3、WebM、OGG、M4A，最大 20MB
                  </p>
                </div>
              )}
            </div>
          ) : (
            <VoiceRecorder
              onRecordingComplete={handleRecordingComplete}
              maxDuration={30}
              minDuration={3}
            />
          )}

          {/* 录音完成显示 */}
          {mode === "record" && audioFile && (
            <div className="flex items-center gap-2 p-3 bg-green-50 rounded-lg">
              <FileAudio className="w-5 h-5 text-green-500" />
              <span className="text-sm text-green-700">
                录音已准备就绪：{audioFile.name}
              </span>
            </div>
          )}

          {/* 表单字段 */}
          <div className="space-y-4">
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

            {/* 分类 */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                分类
              </label>
              <div className="flex flex-wrap gap-2">
                {categories.map((cat) => (
                  <button
                    key={cat.id}
                    onClick={() => setCategory(cat.id)}
                    className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                      category === cat.id
                        ? "bg-purple-500 text-white"
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
                rows={2}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent resize-none"
              />
            </div>
          </div>

          {/* 错误提示 */}
          {error && (
            <div className="p-3 bg-red-50 text-red-600 rounded-lg text-sm">
              {error}
            </div>
          )}

          {/* 录制建议 */}
          <div className="p-4 bg-gray-50 rounded-lg">
            <h4 className="text-sm font-medium text-gray-700 mb-2">录制建议</h4>
            <ul className="text-xs text-gray-500 space-y-1">
              <li>• 朗读一段话，时长 5-15 秒效果最佳</li>
              <li>• 保持声音清晰，避免背景噪音</li>
              <li>• 语调自然，包含抑扬顿挫</li>
              <li>• 避免咳嗽、停顿等杂音</li>
            </ul>
          </div>
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
            disabled={uploading || !audioFile || !name.trim()}
            className={`px-6 py-2 rounded-lg font-medium transition-colors ${
              uploading || !audioFile || !name.trim()
                ? "bg-gray-300 text-gray-500 cursor-not-allowed"
                : "bg-purple-600 text-white hover:bg-purple-700"
            }`}
          >
            {uploading ? (
              <span className="flex items-center gap-2">
                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                上传中...
              </span>
            ) : (
              "上传并保存"
            )}
          </button>
        </div>
      </div>
    </div>
  );

  return createPortal(modal, document.body);
};

export default VoiceUploadModal;
