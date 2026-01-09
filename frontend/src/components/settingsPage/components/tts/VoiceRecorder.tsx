import React, { useState, useRef, useCallback, useEffect } from "react";
import { Mic, Square, Play, Pause, RotateCcw, Check } from "lucide-react";

interface VoiceRecorderProps {
  onRecordingComplete: (blob: Blob) => void;
  maxDuration?: number; // 最大录音时长（秒）
  minDuration?: number; // 最小录音时长（秒）
}

export const VoiceRecorder: React.FC<VoiceRecorderProps> = ({
  onRecordingComplete,
  maxDuration = 30,
  minDuration = 3,
}) => {
  const [isRecording, setIsRecording] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [duration, setDuration] = useState(0);
  const [audioBlob, setAudioBlob] = useState<Blob | null>(null);
  const [audioUrl, setAudioUrl] = useState<string | null>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const chunksRef = useRef<Blob[]>([]);
  const timerRef = useRef<number | null>(null);
  const audioRef = useRef<HTMLAudioElement | null>(null);

  // 清理函数
  const cleanup = useCallback(() => {
    if (timerRef.current) {
      clearInterval(timerRef.current);
      timerRef.current = null;
    }
    if (streamRef.current) {
      streamRef.current.getTracks().forEach((track) => track.stop());
      streamRef.current = null;
    }
    if (audioUrl) {
      URL.revokeObjectURL(audioUrl);
    }
  }, [audioUrl]);

  useEffect(() => {
    return cleanup;
  }, [cleanup]);

  // 开始录音
  const startRecording = async () => {
    try {
      setError(null);
      chunksRef.current = [];

      const stream = await navigator.mediaDevices.getUserMedia({
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          sampleRate: 44100,
        },
      });

      streamRef.current = stream;

      const mediaRecorder = new MediaRecorder(stream, {
        mimeType: MediaRecorder.isTypeSupported("audio/webm")
          ? "audio/webm"
          : "audio/mp4",
      });

      mediaRecorderRef.current = mediaRecorder;

      mediaRecorder.ondataavailable = (e) => {
        if (e.data.size > 0) {
          chunksRef.current.push(e.data);
        }
      };

      mediaRecorder.onstop = () => {
        const blob = new Blob(chunksRef.current, {
          type: mediaRecorder.mimeType,
        });
        setAudioBlob(blob);
        setAudioUrl(URL.createObjectURL(blob));
      };

      mediaRecorder.start(100); // 每100ms收集一次数据
      setIsRecording(true);
      setDuration(0);

      // 计时器
      timerRef.current = window.setInterval(() => {
        setDuration((prev) => {
          const newDuration = prev + 1;
          // 自动停止
          if (newDuration >= maxDuration) {
            stopRecording();
          }
          return newDuration;
        });
      }, 1000);
    } catch (err) {
      console.error("录音失败:", err);
      setError("无法访问麦克风，请检查浏览器权限设置");
    }
  };

  // 停止录音
  const stopRecording = () => {
    if (mediaRecorderRef.current && isRecording) {
      mediaRecorderRef.current.stop();
      setIsRecording(false);
      setIsPaused(false);

      if (timerRef.current) {
        clearInterval(timerRef.current);
        timerRef.current = null;
      }

      if (streamRef.current) {
        streamRef.current.getTracks().forEach((track) => track.stop());
      }
    }
  };

  // 重新录制
  const resetRecording = () => {
    cleanup();
    setAudioBlob(null);
    setAudioUrl(null);
    setDuration(0);
    setIsPlaying(false);
  };

  // 播放/暂停
  const togglePlayback = () => {
    if (!audioRef.current || !audioUrl) return;

    if (isPlaying) {
      audioRef.current.pause();
      setIsPlaying(false);
    } else {
      audioRef.current.src = audioUrl;
      audioRef.current.play();
      setIsPlaying(true);
    }
  };

  // 确认使用
  const confirmRecording = () => {
    if (audioBlob && duration >= minDuration) {
      onRecordingComplete(audioBlob);
    }
  };

  // 格式化时间
  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, "0")}:${secs.toString().padStart(2, "0")}`;
  };

  const isValidDuration = duration >= minDuration;

  return (
    <div className="space-y-4">
      {/* 错误提示 */}
      {error && (
        <div className="p-3 bg-red-50 text-red-600 rounded-lg text-sm">{error}</div>
      )}

      {/* 录音可视化 */}
      <div className="relative h-24 bg-gray-100 rounded-xl flex items-center justify-center overflow-hidden">
        {isRecording ? (
          <div className="flex items-center gap-1">
            {[...Array(20)].map((_, i) => (
              <div
                key={i}
                className="w-1 bg-purple-500 rounded-full animate-pulse"
                style={{
                  height: `${Math.random() * 40 + 20}px`,
                  animationDelay: `${i * 0.05}s`,
                }}
              />
            ))}
          </div>
        ) : audioUrl ? (
          <div className="flex items-center gap-2 text-gray-600">
            <Volume2Icon className="w-6 h-6" />
            <span>录音完成</span>
          </div>
        ) : (
          <div className="text-gray-400 text-sm">点击下方按钮开始录音</div>
        )}

        {/* 时间显示 */}
        <div className="absolute bottom-2 right-2 px-2 py-1 bg-black/50 text-white text-xs rounded">
          {formatTime(duration)} / {formatTime(maxDuration)}
        </div>

        {/* 录音中指示器 */}
        {isRecording && (
          <div className="absolute top-2 left-2 flex items-center gap-1">
            <div className="w-2 h-2 bg-red-500 rounded-full animate-pulse" />
            <span className="text-xs text-red-500 font-medium">录音中</span>
          </div>
        )}
      </div>

      {/* 控制按钮 */}
      <div className="flex items-center justify-center gap-4">
        {!audioUrl ? (
          // 录音模式
          <>
            {!isRecording ? (
              <button
                onClick={startRecording}
                className="w-16 h-16 rounded-full bg-gradient-to-br from-purple-500 to-pink-500 text-white flex items-center justify-center shadow-lg hover:shadow-xl transition-shadow"
              >
                <Mic className="w-8 h-8" />
              </button>
            ) : (
              <button
                onClick={stopRecording}
                className="w-16 h-16 rounded-full bg-red-500 text-white flex items-center justify-center shadow-lg hover:shadow-xl transition-shadow"
              >
                <Square className="w-6 h-6" />
              </button>
            )}
          </>
        ) : (
          // 预览模式
          <>
            <button
              onClick={resetRecording}
              className="p-3 rounded-full bg-gray-200 text-gray-600 hover:bg-gray-300 transition-colors"
              title="重新录制"
            >
              <RotateCcw className="w-5 h-5" />
            </button>

            <button
              onClick={togglePlayback}
              className="w-14 h-14 rounded-full bg-purple-500 text-white flex items-center justify-center shadow-lg hover:shadow-xl transition-shadow"
            >
              {isPlaying ? (
                <Pause className="w-6 h-6" />
              ) : (
                <Play className="w-6 h-6 ml-1" />
              )}
            </button>

            <button
              onClick={confirmRecording}
              disabled={!isValidDuration}
              className={`p-3 rounded-full transition-colors ${
                isValidDuration
                  ? "bg-green-500 text-white hover:bg-green-600"
                  : "bg-gray-200 text-gray-400 cursor-not-allowed"
              }`}
              title={isValidDuration ? "使用此录音" : `至少需要${minDuration}秒`}
            >
              <Check className="w-5 h-5" />
            </button>
          </>
        )}
      </div>

      {/* 提示文字 */}
      <div className="text-center text-xs text-gray-500">
        {!audioUrl ? (
          isRecording ? (
            <span>点击红色按钮停止录音（{minDuration}-{maxDuration}秒）</span>
          ) : (
            <span>点击麦克风开始录音，建议清晰朗读一段话</span>
          )
        ) : (
          <span>
            {isValidDuration
              ? "点击 ✓ 使用此录音，或点击 ↺ 重新录制"
              : `录音时长不足，至少需要${minDuration}秒`}
          </span>
        )}
      </div>

      {/* 隐藏的音频播放器 */}
      <audio
        ref={audioRef}
        onEnded={() => setIsPlaying(false)}
        style={{ display: "none" }}
      />
    </div>
  );
};

// 简单的音量图标
const Volume2Icon: React.FC<{ className?: string }> = ({ className }) => (
  <svg
    className={className}
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
  >
    <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" />
    <path d="M15.54 8.46a5 5 0 0 1 0 7.07" />
    <path d="M19.07 4.93a10 10 0 0 1 0 14.14" />
  </svg>
);

export default VoiceRecorder;

