// API客户端 - 处理与后端API的通信

const RAW_API_BASE = (import.meta.env.VITE_API_BASE_URL as string | undefined) || "http://localhost:8000";
const RAW_WS_BASE = (import.meta.env.VITE_WS_BASE_URL as string | undefined) || "";
const API_BASE_URL = RAW_API_BASE.replace(/\/+$/, "");
const WS_BASE_URL = (RAW_WS_BASE || API_BASE_URL.replace(/^http/, "ws")).replace(/\/+$/, "");
const WS_ENDPOINT = WS_BASE_URL ? `${WS_BASE_URL}/ws` : "";

const AUTH_STORAGE_KEY = "superai.auth.tokens";

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
}

export interface AuthTokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type?: string;
}

export interface AuthUserInfo {
  user_id: string;
  email?: string;
  is_admin?: boolean;
  status?: string;
}

export function getStoredTokens(): AuthTokens | null {
  try {
    const raw = localStorage.getItem(AUTH_STORAGE_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as AuthTokens;
    if (!parsed?.accessToken || !parsed?.refreshToken || !parsed?.expiresAt) return null;
    return parsed;
  } catch {
    return null;
  }
}

export function setStoredTokens(tokens: AuthTokens | null): void {
  try {
    if (!tokens) {
      localStorage.removeItem(AUTH_STORAGE_KEY);
      return;
    }
    localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(tokens));
  } catch {
    // ignore storage errors
  }
}

export function isTokenExpired(tokens: AuthTokens | null): boolean {
  if (!tokens) return true;
  return Date.now() >= tokens.expiresAt - 30_000;
}

// 类型定义
export interface ApiResponse<T = any> {
  message: string;
  data?: T;
  timestamp: string;
}

export interface TaskStatus {
  task_id: string;
  status: "pending" | "processing" | "completed" | "failed";
  progress: number;
  message: string;
}

export interface VideoProcessRequest {
  video_path: string;
  output_path: string;
  settings?: Record<string, any>;
}

export interface WebSocketMessage {
  type: "progress" | "completed" | "error" | "heartbeat" | "pong";
  task_id?: string;
  progress?: number;
  message?: string;
  timestamp: string;
  [key: string]: any;
}

// HTTP客户端类
export class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  setBaseUrl(url: string) {
    this.baseUrl = url;
  }

  getBaseUrl(): string {
    return this.baseUrl;
  }

  // 通用请求方法
  private async request<T>(
    endpoint: string,
    options: (RequestInit & { skipAuth?: boolean; _retry?: boolean }) = {}
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;

    const { skipAuth, _retry, ...fetchOptions } = options;
    const tokens = getStoredTokens();
    const authHeader = !skipAuth && tokens?.accessToken
      ? { Authorization: `Bearer ${tokens.accessToken}` }
      : {};

    const defaultOptions: RequestInit = {
      headers: {
        "Content-Type": "application/json",
        ...authHeader,
        ...(fetchOptions.headers || {}),
      },
      ...fetchOptions,
    };

    try {
      const response = await fetch(url, defaultOptions);

      if (response.status === 401 && !skipAuth && !_retry && tokens?.refreshToken) {
        const refreshed = await this.refreshTokens(tokens.refreshToken);
        if (refreshed) {
          setStoredTokens(refreshed);
          return this.request<T>(endpoint, { ...options, _retry: true });
        }
      }

      if (!response.ok) {
        // 尝试从后端错误响应中提取更明确的提示信息（detail 或 message）
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
        try {
          const contentType = response.headers.get("content-type") || "";
          if (contentType.includes("application/json")) {
            const errJson = await response.json();
            if (typeof errJson === "string") {
              errorMessage = errJson;
            } else if (errJson?.detail) {
              errorMessage = errJson.detail;
            } else if (errJson?.message) {
              errorMessage = errJson.message;
            }
          } else {
            const text = await response.text();
            if (text) errorMessage = text;
          }
        } catch {
          // 忽略解析错误，保留默认错误信息
        }
        throw new Error(errorMessage);
      }

      const contentType = response.headers.get("content-type") || "";
      const data = contentType.includes("application/json")
        ? await response.json()
        : ((await response.text()) as any);
      return data;
    } catch (error) {
      console.error(`API请求失败 [${endpoint}]:`, error);
      throw error;
    }
  }

  private async refreshTokens(refreshToken: string): Promise<AuthTokens | null> {
    try {
      const response = await fetch(`${this.baseUrl}/api/auth/refresh`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
      if (!response.ok) return null;
      const data = (await response.json()) as AuthTokenResponse;
      if (!data?.access_token || !data?.refresh_token || !data?.expires_in) {
        return null;
      }
      return {
        accessToken: data.access_token,
        refreshToken: data.refresh_token,
        expiresAt: Date.now() + data.expires_in * 1000,
      };
    } catch {
      return null;
    }
  }

  // GET请求
  async get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: "GET" });
  }

  // POST请求
  async post<T>(endpoint: string, data?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: "POST",
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  // PUT请求
  async put<T>(endpoint: string, data?: any): Promise<T> {
    return this.request<T>(endpoint, {
      method: "PUT",
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  // DELETE请求
  async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, { method: "DELETE" });
  }

  private normalizeTokenResponse(data: AuthTokenResponse): AuthTokens {
    return {
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
      expiresAt: Date.now() + data.expires_in * 1000,
    };
  }

  async register(email: string, password: string): Promise<AuthTokens> {
    const data = await this.request<AuthTokenResponse>("/api/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password }),
      skipAuth: true,
    });
    const tokens = this.normalizeTokenResponse(data);
    setStoredTokens(tokens);
    return tokens;
  }

  async login(email: string, password: string): Promise<AuthTokens> {
    const data = await this.request<AuthTokenResponse>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
      skipAuth: true,
    });
    const tokens = this.normalizeTokenResponse(data);
    setStoredTokens(tokens);
    return tokens;
  }

  async logout(): Promise<void> {
    const tokens = getStoredTokens();
    if (tokens?.refreshToken) {
      await this.request("/api/auth/logout", {
        method: "POST",
        body: JSON.stringify({ refresh_token: tokens.refreshToken }),
        skipAuth: true,
      });
    }
    setStoredTokens(null);
  }

  async me(): Promise<AuthUserInfo> {
    return this.get<AuthUserInfo>("/api/auth/me");
  }

  async getAdminUsers(limit = 20, offset = 0): Promise<any> {
    const params = new URLSearchParams();
    if (limit > 0) params.set("limit", String(limit));
    if (offset > 0) params.set("offset", String(offset));
    const query = params.toString();
    return this.get(`/api/admin/users${query ? `?${query}` : ""}`);
  }

  async getAdminUser(userId: string): Promise<any> {
    return this.get(`/api/admin/users/${encodeURIComponent(userId)}`);
  }

  async updateAdminUser(
    userId: string,
    payload: { email?: string; is_admin?: boolean; status?: string }
  ): Promise<any> {
    return this.request(`/api/admin/users/${encodeURIComponent(userId)}`, {
      method: "PATCH",
      body: JSON.stringify(payload),
    });
  }

  async resetAdminPassword(userId: string, password: string): Promise<any> {
    return this.post(`/api/admin/users/${encodeURIComponent(userId)}/reset-password`, {
      password,
    });
  }

  async getAdminUserUsage(userId: string): Promise<any> {
    return this.get(`/api/admin/users/${encodeURIComponent(userId)}/usage`);
  }

  async updateAdminUserPlan(userId: string, planId: string): Promise<any> {
    return this.post(`/api/admin/users/${encodeURIComponent(userId)}/plan`, {
      plan_id: planId,
    });
  }

  // 测试连接
  async testConnection(): Promise<boolean> {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 1000); // 1秒超时

      const url = `${this.baseUrl}/api/health`;
      const response = await fetch(url, {
        method: "GET",
        headers: { "Content-Type": "application/json" },
        signal: controller.signal,
      });

      clearTimeout(timeoutId);
      return response.ok;
    } catch {
      return false;
    }
  }

  // 获取Hello消息
  async getHello(): Promise<ApiResponse> {
    return this.get<ApiResponse>("/api/health");
  }

  // 获取服务状态
  async getStatus(): Promise<any> {
    return this.get("/api/health");
  }

  // 获取视频信息
  async getVideoInfo(
    videoPath: string
  ): Promise<{ duration?: number; format?: string }> {
    return this.post("/api/video/info", { video_path: videoPath });
  }

  // 处理视频
  async processVideo(
    request: VideoProcessRequest
  ): Promise<{ task_id: string }> {
    return this.post<{ task_id: string }>("/api/video/process", request);
  }

  // 获取任务状态
  async getTaskStatus(taskId: string): Promise<TaskStatus> {
    return this.get<TaskStatus>(`/api/task/${taskId}`);
  }

  // 获取视频分析模型配置
  async getVideoAnalysisConfigs(): Promise<any> {
    return this.get("/api/models/video-analysis/configs");
  }

  // 更新视频分析模型配置
  async updateVideoAnalysisConfig(configId: string, config: any): Promise<any> {
    return this.request(`/api/models/video-analysis/configs/${configId}`, {
      method: "PUT",
      body: JSON.stringify({ config }),
    });
  }

  // 激活视频分析模型配置
  async activateVideoAnalysisConfig(configId: string): Promise<any> {
    return this.post(`/api/models/video-analysis/configs/${configId}/activate`);
  }

  // 测试视频分析模型配置
  async testVideoAnalysisConfig(configId: string): Promise<any> {
    return this.post(`/api/models/video-analysis/test/${configId}`);
  }

  // 获取文案生成模型配置
  async getContentGenerationConfigs(): Promise<any> {
    return this.get("/api/models/content-generation/configs");
  }

  // 更新文案生成模型配置
  async updateContentGenerationConfig(
    configId: string,
    config: any
  ): Promise<any> {
    return this.request(`/api/models/content-generation/configs/${configId}`, {
      method: "PUT",
      body: JSON.stringify({ config }),
    });
  }

  // 测试文案生成模型配置
  async testContentGenerationConfig(configId: string): Promise<any> {
    return this.post(`/api/models/content-generation/test/${configId}`);
  }

  // ===== TTS（音色设置）相关 API =====
  // 获取TTS引擎列表
  async getTtsEngines(): Promise<any> {
    return this.get(`/api/tts/engines`);
  }

  // 获取音色列表
  async getTtsVoices(provider: string): Promise<any> {
    const p = encodeURIComponent(provider);
    return this.get(`/api/tts/voices?provider=${p}`);
  }

  // 获取TTS配置
  async getTtsConfigs(): Promise<any> {
    return this.get(`/api/tts/configs`);
  }

  // 更新/创建TTS配置（实时保存，局部更新）
  async patchTtsConfig(configId: string, partial: any): Promise<any> {
    return this.request(`/api/tts/configs/${encodeURIComponent(configId)}`, {
      method: "PATCH",
      body: JSON.stringify(partial),
    });
  }

  // 激活指定TTS配置（保证唯一启用）
  async activateTtsConfig(configId: string): Promise<any> {
    return this.post(`/api/tts/configs/${encodeURIComponent(configId)}/activate`);
  }

  // 测试TTS引擎连通性
  async testTtsConnection(configId: string): Promise<any> {
    return this.post(`/api/tts/configs/${encodeURIComponent(configId)}/test`);
  }

  // 测试所有集成组件（一键自检）
  async testIntegrations(): Promise<any> {
    return this.post("/api/health/test-integrations");
  }

  // 音色试听（优先使用凭据生成，其次回退 sample_wav_url）
  async previewTtsVoice(voiceId: string, req?: { text?: string; provider?: string; config_id?: string }): Promise<any> {
    return this.post(`/api/tts/voices/${encodeURIComponent(voiceId)}/preview`, req || {});
  }

  // ===== IndexTTS2 专属 API =====
  // 获取情感类型列表
  async getTtsEmotions(): Promise<any> {
    return this.get('/api/tts/emotions');
  }

  // 获取 IndexTTS 模型状态
  async getIndexTtsStatus(): Promise<any> {
    return this.get('/api/tts/index-tts/status');
  }

  // 预加载 IndexTTS 模型
  async preloadIndexTts(): Promise<any> {
    return this.post('/api/tts/index-tts/preload');
  }

  // 测试 IndexTTS 可用性
  async testIndexTts(): Promise<any> {
    return this.post('/api/tts/index-tts/test');
  }

  // ===== 音色管理 API =====
  // 获取音色分类列表
  async getVoiceCategories(): Promise<any> {
    return this.get('/api/voices/categories');
  }

  // 创建音色分类
  async createVoiceCategory(name: string, icon: string): Promise<any> {
    return this.post('/api/voices/categories', { name, icon });
  }

  // 更新音色分类
  async updateVoiceCategory(id: string, name?: string, icon?: string): Promise<any> {
    return this.request(`/api/voices/categories/${encodeURIComponent(id)}`, {
      method: 'PATCH',
      body: JSON.stringify({ name, icon }),
    });
  }

  // 删除音色分类
  async deleteVoiceCategory(id: string): Promise<any> {
    return this.delete(`/api/voices/categories/${encodeURIComponent(id)}`);
  }

  // 获取自定义音色列表
  async getCustomVoices(category?: string): Promise<any> {
    const params = category ? `?category=${encodeURIComponent(category)}` : '';
    return this.get(`/api/voices${params}`);
  }

  // 获取单个音色详情
  async getCustomVoice(id: string): Promise<any> {
    return this.get(`/api/voices/${encodeURIComponent(id)}`);
  }

  // 上传音色
  async uploadVoice(data: { name: string; category: string; description?: string; audio: File }): Promise<any> {
    const formData = new FormData();
    formData.append('name', data.name);
    formData.append('category', data.category);
    formData.append('description', data.description || '');
    formData.append('audio', data.audio);

    const url = `${this.baseUrl}/api/voices/upload`;
    const tokens = getStoredTokens();
    const response = await fetch(url, {
      method: 'POST',
      headers: tokens?.accessToken ? { Authorization: `Bearer ${tokens.accessToken}` } : undefined,
      body: formData,
    });

    if (!response.ok) {
      const errJson = await response.json().catch(() => ({}));
      throw new Error(errJson.detail || errJson.message || `HTTP ${response.status}`);
    }

    return response.json();
  }

  // 上传录音
  async uploadRecording(data: { name: string; category: string; description?: string; audio: File }): Promise<any> {
    const formData = new FormData();
    formData.append('name', data.name);
    formData.append('category', data.category);
    formData.append('description', data.description || '');
    formData.append('audio', data.audio);

    const url = `${this.baseUrl}/api/voices/record`;
    const tokens = getStoredTokens();
    const response = await fetch(url, {
      method: 'POST',
      headers: tokens?.accessToken ? { Authorization: `Bearer ${tokens.accessToken}` } : undefined,
      body: formData,
    });

    if (!response.ok) {
      const errJson = await response.json().catch(() => ({}));
      throw new Error(errJson.detail || errJson.message || `HTTP ${response.status}`);
    }

    return response.json();
  }

  // 更新音色信息
  async updateVoice(id: string, data: { name?: string; category?: string; description?: string }): Promise<any> {
    return this.request(`/api/voices/${encodeURIComponent(id)}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  // 删除音色
  async deleteVoice(id: string): Promise<any> {
    return this.delete(`/api/voices/${encodeURIComponent(id)}`);
  }

  // 生成音色试听
  async generateVoicePreview(id: string, text?: string): Promise<Blob> {
    const url = `${this.baseUrl}/api/voices/${encodeURIComponent(id)}/preview`;
    const tokens = getStoredTokens();
    const response = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(tokens?.accessToken ? { Authorization: `Bearer ${tokens.accessToken}` } : {}),
      },
      body: JSON.stringify({ text }),
    });

    if (!response.ok) {
      const errJson = await response.json().catch(() => ({}));
      throw new Error(errJson.detail || errJson.message || `HTTP ${response.status}`);
    }

    return response.blob();
  }
}

// WebSocket客户端类
export class WebSocketClient {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;
  private listeners: Map<string, Set<(data: any) => void>> = new Map();

  constructor(url: string = WS_ENDPOINT) {
    this.url = url;
  }

  setUrl(url: string) {
    this.url = url;
  }

  // 连接WebSocket
  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        if (!this.url) {
          reject(new Error("WebSocket URL 未配置"));
          return;
        }
        console.log("正在连接WebSocket:", this.url);
        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          console.log("WebSocket连接已建立");
          this.reconnectAttempts = 0;
          // 触发open事件
          this.emit("open", { connected: true });
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message: WebSocketMessage = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (error) {
            console.error("解析WebSocket消息失败:", error);
          }
        };

        this.ws.onclose = (event) => {
          console.log("WebSocket连接已关闭:", event.code, event.reason);
          // 触发close事件
          this.emit("close", {
            connected: false,
            code: event.code,
            reason: event.reason,
          });
          this.handleReconnect();
        };

        this.ws.onerror = (error) => {
          console.error("WebSocket错误:", error);
          // 触发error事件
          this.emit("error", { error });
          reject(error);
        };
      } catch (error) {
        console.error("创建WebSocket连接失败:", error);
        reject(error);
      }
    });
  }

  // 处理消息
  private handleMessage(message: WebSocketMessage) {
    // 触发对应类型的监听器
    const typeListeners = this.listeners.get(message.type);
    if (typeListeners) {
      typeListeners.forEach((listener) => listener(message));
    }

    // 触发通用监听器
    const allListeners = this.listeners.get("*");
    if (allListeners) {
      allListeners.forEach((listener) => listener(message));
    }
  }

  // 处理重连
  private handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      console.log(
        `尝试重连WebSocket (${this.reconnectAttempts}/${this.maxReconnectAttempts})`
      );

      setTimeout(() => {
        this.connect().catch((error) => {
          console.error("WebSocket重连失败:", error);
        });
      }, this.reconnectDelay * this.reconnectAttempts);
    } else {
      console.error("WebSocket重连次数已达上限");
    }
  }

  // 发送消息
  send(message: any): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn("WebSocket未连接，无法发送消息");
    }
  }

  // 发送ping消息
  ping(): void {
    this.send({ type: "ping", timestamp: new Date().toISOString() });
  }

  // 添加事件监听器
  on(type: string, listener: (data: any) => void): void {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, new Set());
    }
    this.listeners.get(type)!.add(listener);
  }

  // 移除事件监听器
  off(type: string, listener: (data: any) => void): void {
    const typeListeners = this.listeners.get(type);
    if (typeListeners) {
      typeListeners.delete(listener);
    }
  }

  // 触发事件
  private emit(type: string, data: any): void {
    const typeListeners = this.listeners.get(type);
    if (typeListeners) {
      typeListeners.forEach((listener) => listener(data));
    }
  }

  // 断开连接
  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  // 获取连接状态
  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

// Tauri命令包装器
// 导出单例实例
export const apiClient = new ApiClient();
export const wsClient = new WebSocketClient();

// 工具函数
export const utils = {
  // 格式化文件大小
  formatFileSize(bytes: number): string {
    const sizes = ["B", "KB", "MB", "GB", "TB"];
    if (bytes === 0) return "0 B";
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round((bytes / Math.pow(1024, i)) * 100) / 100 + " " + sizes[i];
  },

  // 格式化时长
  formatDuration(seconds: number): string {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = Math.floor(seconds % 60);

    if (hours > 0) {
      return `${hours}:${minutes.toString().padStart(2, "0")}:${secs
        .toString()
        .padStart(2, "0")}`;
    }
    return `${minutes}:${secs.toString().padStart(2, "0")}`;
  },

  // 延迟函数
  delay(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  },

  // 重试函数
  async retry<T>(
    fn: () => Promise<T>,
    maxAttempts: number = 3,
    delay: number = 1000
  ): Promise<T> {
    let lastError: Error;

    for (let attempt = 1; attempt <= maxAttempts; attempt++) {
      try {
        return await fn();
      } catch (error) {
        lastError = error as Error;
        if (attempt < maxAttempts) {
          await this.delay(delay * attempt);
        }
      }
    }

    throw lastError!;
  },
};
