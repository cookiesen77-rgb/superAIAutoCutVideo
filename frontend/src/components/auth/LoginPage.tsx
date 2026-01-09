import React, { useState } from "react";
import { Lock, Mail, LogIn, UserPlus } from "lucide-react";
import { authService } from "@/services/authService";

interface LoginPageProps {
  onAuthSuccess: () => void;
}

const LoginPage: React.FC<LoginPageProps> = ({ onAuthSuccess }) => {
  const [mode, setMode] = useState<"login" | "register">("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async () => {
    setError(null);
    if (!email.trim()) {
      setError("请输入邮箱");
      return;
    }
    if (!password.trim()) {
      setError("请输入密码");
      return;
    }
    setSubmitting(true);
    try {
      if (mode === "login") {
        await authService.login(email.trim(), password);
      } else {
        await authService.register(email.trim(), password);
      }
      onAuthSuccess();
    } catch (err) {
      setError(err instanceof Error ? err.message : "认证失败");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
      <div className="w-full max-w-md bg-white rounded-2xl shadow-xl border p-6">
        <div className="flex items-center gap-3 mb-6">
          <div className="w-10 h-10 rounded-full bg-purple-100 flex items-center justify-center">
            <Lock className="w-5 h-5 text-purple-600" />
          </div>
          <div>
            <h1 className="text-xl font-semibold text-gray-900">云端服务登录</h1>
            <p className="text-xs text-gray-500">
              {mode === "login" ? "请输入账号继续" : "注册后即可使用云端功能"}
            </p>
          </div>
        </div>

        <div className="flex gap-2 p-1 bg-gray-100 rounded-lg mb-6">
          <button
            onClick={() => setMode("login")}
            className={`flex-1 py-2 rounded-md text-sm font-medium transition-colors ${
              mode === "login"
                ? "bg-white text-purple-600 shadow-sm"
                : "text-gray-600 hover:text-gray-900"
            }`}
          >
            登录
          </button>
          <button
            onClick={() => setMode("register")}
            className={`flex-1 py-2 rounded-md text-sm font-medium transition-colors ${
              mode === "register"
                ? "bg-white text-purple-600 shadow-sm"
                : "text-gray-600 hover:text-gray-900"
            }`}
          >
            注册
          </button>
        </div>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              邮箱
            </label>
            <div className="flex items-center gap-2 border border-gray-300 rounded-lg px-3 py-2 focus-within:ring-2 focus-within:ring-purple-500">
              <Mail className="w-4 h-4 text-gray-400" />
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="name@example.com"
                className="flex-1 outline-none text-sm"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              密码
            </label>
            <div className="flex items-center gap-2 border border-gray-300 rounded-lg px-3 py-2 focus-within:ring-2 focus-within:ring-purple-500">
              <Lock className="w-4 h-4 text-gray-400" />
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="至少 6 位"
                className="flex-1 outline-none text-sm"
              />
            </div>
          </div>
        </div>

        {error && (
          <div className="mt-4 p-3 bg-red-50 text-red-600 rounded-lg text-sm">
            {error}
          </div>
        )}

        <button
          onClick={handleSubmit}
          disabled={submitting}
          className={`mt-6 w-full flex items-center justify-center gap-2 py-2.5 rounded-lg text-sm font-medium transition-colors ${
            submitting
              ? "bg-gray-200 text-gray-500 cursor-not-allowed"
              : "bg-purple-600 text-white hover:bg-purple-700"
          }`}
        >
          {mode === "login" ? (
            <LogIn className="w-4 h-4" />
          ) : (
            <UserPlus className="w-4 h-4" />
          )}
          {submitting ? "处理中..." : mode === "login" ? "登录" : "注册"}
        </button>
      </div>
    </div>
  );
};

export default LoginPage;
