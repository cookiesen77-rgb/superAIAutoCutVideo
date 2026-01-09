import { apiClient } from "./clients";
import type { VoiceUploadRequest, VoiceUpdateRequest } from "@/components/settingsPage/types";

export const ttsService = {
  getEngines: () => apiClient.getTtsEngines(),
  getConfigs: () => apiClient.getTtsConfigs(),
  patchConfig: (configId: string, data: any) => apiClient.patchTtsConfig(configId, data),
  activateConfig: (configId: string) => apiClient.activateTtsConfig(configId),
  getVoices: (provider: string) => apiClient.getTtsVoices(provider),
  previewVoice: (voiceId: string, data: any) => apiClient.previewTtsVoice(voiceId, data),
  testConnection: (configId: string) => apiClient.testTtsConnection(configId),
  // IndexTTS2 专属 API
  getEmotions: () => apiClient.getTtsEmotions(),
  getIndexTtsStatus: () => apiClient.getIndexTtsStatus(),
  preloadIndexTts: () => apiClient.preloadIndexTts(),
  testIndexTts: () => apiClient.testIndexTts(),
};

// 音色管理服务
export const voiceService = {
  // 分类管理
  getCategories: () => apiClient.getVoiceCategories(),
  createCategory: (name: string, icon: string) => apiClient.createVoiceCategory(name, icon),
  updateCategory: (id: string, name?: string, icon?: string) => apiClient.updateVoiceCategory(id, name, icon),
  deleteCategory: (id: string) => apiClient.deleteVoiceCategory(id),
  
  // 音色管理
  getCustomVoices: (category?: string) => apiClient.getCustomVoices(category),
  getVoice: (id: string) => apiClient.getCustomVoice(id),
  uploadVoice: (data: VoiceUploadRequest) => apiClient.uploadVoice(data),
  uploadRecording: (data: VoiceUploadRequest) => apiClient.uploadRecording(data),
  updateVoice: (id: string, data: VoiceUpdateRequest) => apiClient.updateVoice(id, data),
  deleteVoice: (id: string) => apiClient.deleteVoice(id),
  
  // 音频
  getVoiceAudioUrl: (id: string) => `${apiClient.getBaseUrl()}/api/voices/${id}/audio`,
  generatePreview: (id: string, text?: string) => apiClient.generateVoicePreview(id, text),
};

export default ttsService;