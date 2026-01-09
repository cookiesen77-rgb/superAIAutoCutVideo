import { RefreshCw } from "lucide-react";
import React, { useEffect, useState } from "react";
import Navigation from "./components/Navigation";
import SettingsPage from "./components/settingsPage";
import ProjectEditPage from "./pages/ProjectEditPage";
import ProjectManagementPage from "./pages/ProjectManagementPage";
import LoginPage from "./components/auth/LoginPage";
import {
  WebSocketMessage,
  apiClient,
  wsClient,
} from "./services/clients";
import { authService } from "./services/authService";

interface BackendStatus {
  running: boolean;
  baseUrl: string;
}

const App: React.FC = () => {
  // 状态管理
  const [backendStatus, setBackendStatus] = useState<BackendStatus>({
    running: false,
    baseUrl: apiClient.getBaseUrl(),
  });
  const [connectionStatus, setConnectionStatus] = useState({
    backend: false,
    api: false,
    websocket: false,
  });
  const [messages, setMessages] = useState<WebSocketMessage[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [activeTab, setActiveTab] = useState("home");
  const [currentProjectId, setCurrentProjectId] = useState<string | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(authService.isAuthenticated());

  // 初始化应用
  useEffect(() => {
    const handleWsMessage = (message: WebSocketMessage) => {
      setMessages((prev) => [...prev, message]);
    };
    const handleWsOpen = () => {
      console.log("WebSocket连接状态更新: 已连接");
      setConnectionStatus((prev) => ({ ...prev, websocket: true }));
    };
    const handleWsClose = () => {
      console.log("WebSocket连接状态更新: 已断开");
      setConnectionStatus((prev) => ({ ...prev, websocket: false }));
    };
    const handleWsError = (error: any) => {
      console.error("WebSocket连接错误:", error);
      setConnectionStatus((prev) => ({ ...prev, websocket: false }));
    };

    wsClient.on("*", handleWsMessage);
    wsClient.on("open", handleWsOpen);
    wsClient.on("close", handleWsClose);
    wsClient.on("error", handleWsError);

    if (isAuthenticated) {
      initializeApp();
    } else {
      setIsLoading(false);
    }

    return () => {
      wsClient.off("*", handleWsMessage);
      wsClient.off("open", handleWsOpen);
      wsClient.off("close", handleWsClose);
      wsClient.off("error", handleWsError);
      wsClient.disconnect();
    };
  }, [isAuthenticated]);

  const initializeApp = async () => {
    try {
      setIsLoading(true);

      const apiOk = await testApiConnection();
      setBackendStatus({ running: apiOk, baseUrl: apiClient.getBaseUrl() });
      setConnectionStatus((prev) => ({ ...prev, backend: apiOk }));

      if (apiOk) {
        console.log("API可用，尝试连接WebSocket...");
        const wsTimeout = new Promise((_, reject) =>
          setTimeout(() => reject(new Error("WebSocket连接超时")), 2000)
        );
        try {
          await Promise.race([wsClient.connect(), wsTimeout]);
          console.log("WebSocket连接成功");
        } catch (error) {
          console.error("WebSocket连接失败:", error);
        }
      }
    } catch (error) {
      console.error("初始化应用失败:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const testApiConnection = async (): Promise<boolean> => {
    try {
      const response = await apiClient.testConnection();
      setConnectionStatus((prev) => ({ ...prev, api: response }));
      return response;
    } catch (error) {
      setConnectionStatus((prev) => ({ ...prev, api: false }));
      return false;
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <RefreshCw className="h-8 w-8 animate-spin text-blue-600 mx-auto mb-4" />
          <p className="text-gray-600">正在初始化应用...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <LoginPage onAuthSuccess={() => setIsAuthenticated(true)} />;
  }


  return (
    <div className="min-h-screen bg-gray-50">
      {/* 导航栏 */}
      <Navigation activeTab={activeTab} onTabChange={setActiveTab} />

      {/* 主要内容区域 */}
      <main className="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
        {/* 标签页内容 */}
        {activeTab === "home" && !currentProjectId && (
          <ProjectManagementPage
            onEditProject={(projectId) => setCurrentProjectId(projectId)}
          />
        )}

        {activeTab === "home" && currentProjectId && (
          <ProjectEditPage
            projectId={currentProjectId}
            onBack={() => setCurrentProjectId(null)}
          />
        )}

        

        {activeTab === "settings" && (
          <SettingsPage
            messages={messages}
            backendStatus={backendStatus}
            connections={connectionStatus}
          />
        )}

        
      </main>
    </div>
  );
};

export default App;
