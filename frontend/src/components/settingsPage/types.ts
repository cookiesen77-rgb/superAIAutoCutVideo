/**
 * 视频生成模型配置接口
 */
export interface VideoModelConfig {
  provider: string;
  api_key: string;
  base_url: string;
  model_name: string;
  extra_params: Record<string, any>;
  description?: string;
  enabled?: boolean;
}

/**
 * 文案生成模型配置接口
 */
export interface ContentModelConfig {
  provider: string;
  api_key: string;
  base_url: string;
  model_name: string;
  extra_params: Record<string, any>;
  description?: string;
  enabled?: boolean;
}

/**
 * 应用设置接口
 */
export interface AppSettings {
}

/**
 * 测试结果接口
 */
export interface TestResult {
  success: boolean;
  message: string;
}

/**
 * 设置页面栏目接口
 */
export interface SettingsSection {
  id: string;
  label: string;
  icon: any;
}

/**
 * TTS 引擎元信息
 */
export interface TtsEngineMeta {
  provider: string;
  display_name: string;
  description?: string;
  required_fields?: string[];
  optional_fields?: string[];
}

/**
 * TTS 音色信息
 */
export interface TtsVoice {
  id: string;
  name: string;
  description?: string;
  sample_wav_url?: string;
  language?: string;
  gender?: string;
  tags?: string[];
  voice_quality?: string;
  voice_type_tag?: string;
  voice_human_style?: string;
}

/**
 * TTS 引擎配置
 */
export interface TtsEngineConfig {
  provider: string;
  secret_id?: string | null;
  secret_key?: string | null;
  region?: string | null;
  description?: string | null;
  enabled: boolean;
  active_voice_id?: string | null;
  speed_ratio: number;
  extra_params?: {
    // IndexTTS2 情感控制参数
    emotion_mode?: 'auto' | 'manual' | 'disabled';
    default_emotion?: string;
    emo_alpha?: number;
    use_fp16?: boolean;
    // 其他参数
    [key: string]: any;
  };
}

/**
 * TTS 情感类型（IndexTTS2 专用）
 */
export interface TtsEmotion {
  id: string;
  name: string;
  icon: string;
  description?: string;
}

/**
 * IndexTTS 模型状态
 */
export interface IndexTtsStatus {
  loaded: boolean;
  loading: boolean;
  available: boolean;
  error?: string | null;
}

export interface TtsConfigsData {
  configs: Record<string, TtsEngineConfig>;
  active_config_id?: string | null;
}

export interface TtsTestResult {
  success: boolean;
  config_id: string;
  provider: string;
  message: string;
}

