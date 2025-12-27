import { Sparkles, Sliders } from "lucide-react";
import React from "react";

interface Props {
  emotionMode: "auto" | "manual" | "disabled";
  selectedEmotion: string;
  emoAlpha: number;
  onModeChange: (mode: "auto" | "manual" | "disabled") => void;
  onEmotionChange: (emotion: string) => void;
  onAlphaChange: (alpha: number) => void;
  disabled?: boolean;
}

const EMOTION_OPTIONS = [
  { id: "happy", name: "开心", icon: "😊" },
  { id: "sad", name: "悲伤", icon: "😢" },
  { id: "angry", name: "愤怒", icon: "😠" },
  { id: "afraid", name: "恐惧", icon: "😨" },
  { id: "calm", name: "平静", icon: "😌" },
  { id: "surprised", name: "惊讶", icon: "😲" },
  { id: "melancholic", name: "忧郁", icon: "😔" },
  { id: "disgusted", name: "厌恶", icon: "🤢" },
];

const MODE_OPTIONS = [
  { id: "disabled", name: "禁用", icon: "🔇", description: "不使用情感控制" },
  { id: "auto", name: "自动", icon: "🎯", description: "根据文本自动推断" },
  { id: "manual", name: "手动", icon: "🎨", description: "手动选择情感" },
];

export const TtsEmotionSelect: React.FC<Props> = ({
  emotionMode,
  selectedEmotion,
  emoAlpha,
  onModeChange,
  onEmotionChange,
  onAlphaChange,
  disabled = false,
}) => {
  return (
    <div className="space-y-4">
      {/* 情感模式选择 */}
      <div>
        <div className="flex items-center gap-2 mb-2">
          <Sparkles className="h-4 w-4 text-purple-600" />
          <span className="text-sm font-medium text-gray-700">情感控制模式</span>
        </div>
        <div className="flex gap-2">
          {MODE_OPTIONS.map((mode) => (
            <button
              key={mode.id}
              onClick={() => onModeChange(mode.id as "auto" | "manual" | "disabled")}
              disabled={disabled}
              className={`flex-1 px-3 py-2 rounded-lg border text-sm transition-all ${
                emotionMode === mode.id
                  ? "border-purple-500 bg-purple-50 text-purple-700"
                  : "border-gray-200 bg-white text-gray-600 hover:border-gray-300"
              } ${disabled ? "opacity-50 cursor-not-allowed" : ""}`}
              title={mode.description}
            >
              <span className="mr-1">{mode.icon}</span>
              {mode.name}
            </button>
          ))}
        </div>
        <p className="mt-1 text-xs text-gray-500">
          {emotionMode === "auto" && "将根据解说文本内容自动推断合适的情感语调"}
          {emotionMode === "manual" && "手动选择固定的情感语调应用到所有配音"}
          {emotionMode === "disabled" && "不使用情感控制，保持中性语调"}
        </p>
      </div>

      {/* 手动选择情感 */}
      {emotionMode === "manual" && (
        <div>
          <div className="flex items-center gap-2 mb-2">
            <span className="text-sm font-medium text-gray-700">选择情感</span>
          </div>
          <div className="grid grid-cols-4 gap-2">
            {EMOTION_OPTIONS.map((emotion) => (
              <button
                key={emotion.id}
                onClick={() => onEmotionChange(emotion.id)}
                disabled={disabled}
                className={`px-2 py-2 rounded-lg border text-sm transition-all ${
                  selectedEmotion === emotion.id
                    ? "border-purple-500 bg-purple-50 text-purple-700"
                    : "border-gray-200 bg-white text-gray-600 hover:border-gray-300"
                } ${disabled ? "opacity-50 cursor-not-allowed" : ""}`}
              >
                <span className="text-lg">{emotion.icon}</span>
                <div className="text-xs mt-1">{emotion.name}</div>
              </button>
            ))}
          </div>
        </div>
      )}

      {/* 情感强度滑块（仅在非禁用模式下显示） */}
      {emotionMode !== "disabled" && (
        <div>
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <Sliders className="h-4 w-4 text-gray-500" />
              <span className="text-sm font-medium text-gray-700">情感强度</span>
            </div>
            <span className="text-sm text-gray-500">{Math.round(emoAlpha * 100)}%</span>
          </div>
          <input
            type="range"
            min="0"
            max="100"
            value={Math.round(emoAlpha * 100)}
            onChange={(e) => onAlphaChange(parseInt(e.target.value) / 100)}
            disabled={disabled}
            className="w-full h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer accent-purple-600 disabled:opacity-50"
          />
          <div className="flex justify-between text-xs text-gray-400 mt-1">
            <span>轻微</span>
            <span>适中</span>
            <span>强烈</span>
          </div>
          <p className="mt-1 text-xs text-gray-500">
            建议设置在 50-70% 之间，过高可能导致语音不自然
          </p>
        </div>
      )}
    </div>
  );
};

export default TtsEmotionSelect;

