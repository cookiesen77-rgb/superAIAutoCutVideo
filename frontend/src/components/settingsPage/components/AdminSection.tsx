import React, { useCallback, useEffect, useMemo, useState } from "react";
import { CheckCircle, CreditCard, KeyRound, RefreshCw, Shield, UserCog, UserX } from "lucide-react";
import adminService, { AdminUsageSummary, AdminUser } from "@/services/adminService";

const formatDateTime = (value?: string) => {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString("zh-CN", { hour12: false });
};

const formatNumber = (value?: number) => {
  if (typeof value !== "number") return "0";
  return value.toLocaleString("zh-CN");
};

const percent = (used: number, quota: number) => {
  if (quota <= 0) return 0;
  return Math.min(100, Math.round((used / quota) * 100));
};

const AdminSection: React.FC = () => {
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [offset, setOffset] = useState(0);
  const [filter, setFilter] = useState("");
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);

  const [editEmail, setEditEmail] = useState("");
  const [editStatus, setEditStatus] = useState<"active" | "disabled">("active");
  const [editIsAdmin, setEditIsAdmin] = useState(false);
  const [savingProfile, setSavingProfile] = useState(false);

  const [resetPassword, setResetPassword] = useState("");
  const [resettingPassword, setResettingPassword] = useState(false);

  const [planId, setPlanId] = useState("");
  const [updatingPlan, setUpdatingPlan] = useState(false);

  const [usage, setUsage] = useState<AdminUsageSummary | null>(null);
  const [usageLoading, setUsageLoading] = useState(false);
  const [usageError, setUsageError] = useState<string | null>(null);

  const [notice, setNotice] = useState<{ type: "success" | "error"; text: string } | null>(null);

  const limit = 20;

  const selectedUser = useMemo(
    () => users.find((user) => user.id === selectedUserId) || null,
    [users, selectedUserId]
  );

  const filteredUsers = useMemo(() => {
    const keyword = filter.trim().toLowerCase();
    if (!keyword) return users;
    return users.filter((user) => {
      const email = user.email?.toLowerCase() || "";
      return email.includes(keyword) || user.id.toLowerCase().includes(keyword);
    });
  }, [users, filter]);

  const loadUsers = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await adminService.listUsers(limit, offset);
      const nextUsers = response?.data || [];
      setUsers(nextUsers);
      setSelectedUserId((prev) => {
        if (nextUsers.length === 0) return null;
        if (prev && nextUsers.some((u) => u.id === prev)) return prev;
        return nextUsers[0].id;
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载用户失败");
    } finally {
      setLoading(false);
    }
  }, [limit, offset]);

  const loadUsage = useCallback(
    async (userId: string) => {
      setUsageLoading(true);
      setUsageError(null);
      try {
        const response = await adminService.getUsage(userId);
        setUsage(response?.data || null);
      } catch (err) {
        setUsageError(err instanceof Error ? err.message : "获取用量失败");
      } finally {
        setUsageLoading(false);
      }
    },
    []
  );

  useEffect(() => {
    loadUsers();
  }, [loadUsers]);

  useEffect(() => {
    if (!selectedUser) {
      setEditEmail("");
      setEditStatus("active");
      setEditIsAdmin(false);
      setResetPassword("");
      setPlanId("");
      setUsage(null);
      setUsageError(null);
      return;
    }
    setEditEmail(selectedUser.email);
    setEditStatus(selectedUser.status);
    setEditIsAdmin(selectedUser.is_admin);
    setResetPassword("");
    setNotice(null);
    loadUsage(selectedUser.id);
  }, [selectedUser, loadUsage]);

  useEffect(() => {
    if (usage?.plan?.id) {
      setPlanId(usage.plan.id);
    }
  }, [usage?.plan?.id]);

  const handleSaveProfile = async () => {
    if (!selectedUser) return;
    const payload: { email?: string; status?: string; is_admin?: boolean } = {};
    const trimmedEmail = editEmail.trim();
    if (trimmedEmail && trimmedEmail !== selectedUser.email) {
      payload.email = trimmedEmail;
    }
    if (editStatus !== selectedUser.status) {
      payload.status = editStatus;
    }
    if (editIsAdmin !== selectedUser.is_admin) {
      payload.is_admin = editIsAdmin;
    }
    if (Object.keys(payload).length === 0) {
      setNotice({ type: "success", text: "没有需要更新的内容" });
      return;
    }
    setSavingProfile(true);
    setNotice(null);
    try {
      const response = await adminService.updateUser(selectedUser.id, payload);
      const updatedUser = response?.data;
      if (updatedUser) {
        setUsers((prev) => prev.map((u) => (u.id === updatedUser.id ? updatedUser : u)));
        setSelectedUserId(updatedUser.id);
      }
      setNotice({ type: "success", text: response?.message || "用户信息已更新" });
    } catch (err) {
      setNotice({ type: "error", text: err instanceof Error ? err.message : "更新失败" });
    } finally {
      setSavingProfile(false);
    }
  };

  const handleResetPassword = async () => {
    if (!selectedUser) return;
    const trimmed = resetPassword.trim();
    if (trimmed.length < 6) {
      setNotice({ type: "error", text: "密码至少 6 位" });
      return;
    }
    setResettingPassword(true);
    setNotice(null);
    try {
      const response = await adminService.resetPassword(selectedUser.id, trimmed);
      setResetPassword("");
      setNotice({ type: "success", text: response?.message || "密码已重置" });
    } catch (err) {
      setNotice({ type: "error", text: err instanceof Error ? err.message : "重置失败" });
    } finally {
      setResettingPassword(false);
    }
  };

  const handleUpdatePlan = async () => {
    if (!selectedUser) return;
    if (!planId.trim()) {
      setNotice({ type: "error", text: "请输入套餐 ID" });
      return;
    }
    setUpdatingPlan(true);
    setNotice(null);
    try {
      const response = await adminService.updatePlan(selectedUser.id, planId.trim());
      setNotice({ type: "success", text: response?.message || "套餐已更新" });
      await loadUsage(selectedUser.id);
    } catch (err) {
      setNotice({ type: "error", text: err instanceof Error ? err.message : "更新失败" });
    } finally {
      setUpdatingPlan(false);
    }
  };

  const llmUsed = usage?.used?.llm_chars ?? 0;
  const ttsUsed = usage?.used?.tts_chars ?? 0;
  const llmQuota = usage?.quota?.llm_chars ?? 0;
  const ttsQuota = usage?.quota?.tts_chars ?? 0;

  const profileDirty =
    !!selectedUser &&
    (editEmail.trim() !== selectedUser.email ||
      editStatus !== selectedUser.status ||
      editIsAdmin !== selectedUser.is_admin);

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <Shield className="w-5 h-5 text-blue-600" />
          <h4 className="text-lg font-semibold text-gray-900">管理员控制台</h4>
        </div>
        <button
          onClick={loadUsers}
          disabled={loading}
          className="flex items-center gap-2 px-3 py-2 text-sm rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-50"
        >
          <RefreshCw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
          刷新用户
        </button>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <div className="flex-1 min-w-[240px]">
          <label className="text-xs text-gray-500">搜索用户</label>
          <input
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            placeholder="邮箱或用户 ID"
            className="mt-1 w-full px-3 py-2 rounded-lg border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <div className="text-xs text-gray-500">
          当前显示 {filteredUsers.length} / {users.length} 人
        </div>
      </div>

      {error && (
        <div className="p-3 rounded-lg bg-red-50 text-red-600 text-sm">{error}</div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-[minmax(0,2fr)_minmax(0,1fr)] gap-6">
        <div className="bg-white border rounded-xl overflow-hidden">
          <div className="px-4 py-3 border-b bg-gray-50 text-sm font-medium text-gray-700">
            用户列表
          </div>
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="w-8 h-8 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin" />
            </div>
          ) : filteredUsers.length === 0 ? (
            <div className="py-10 text-center text-sm text-gray-500">暂无用户</div>
          ) : (
            <div className="divide-y">
              {filteredUsers.map((user) => {
                const isActive = user.id === selectedUserId;
                return (
                  <button
                    key={user.id}
                    type="button"
                    onClick={() => setSelectedUserId(user.id)}
                    className={`w-full text-left px-4 py-3 transition-colors ${
                      isActive ? "bg-blue-50" : "hover:bg-gray-50"
                    }`}
                  >
                    <div className="flex items-center justify-between gap-3">
                      <div>
                        <p className="text-sm font-medium text-gray-900">{user.email}</p>
                        <p className="text-xs text-gray-500">{user.id}</p>
                      </div>
                      <div className="text-right">
                        <span
                          className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                            user.status === "active"
                              ? "bg-green-100 text-green-700"
                              : "bg-gray-200 text-gray-600"
                          }`}
                        >
                          {user.status === "active" ? "正常" : "禁用"}
                        </span>
                        <div className="mt-1 text-xs text-gray-500">
                          {user.is_admin ? "管理员" : "普通用户"}
                        </div>
                      </div>
                    </div>
                  </button>
                );
              })}
            </div>
          )}
          <div className="flex items-center justify-between px-4 py-3 border-t bg-gray-50 text-xs text-gray-500">
            <button
              disabled={offset === 0}
              onClick={() => setOffset((prev) => Math.max(0, prev - limit))}
              className={`px-2 py-1 rounded ${
                offset === 0 ? "text-gray-300 cursor-not-allowed" : "hover:bg-gray-200"
              }`}
            >
              上一页
            </button>
            <span>当前偏移 {offset}</span>
            <button
              disabled={users.length < limit}
              onClick={() => setOffset((prev) => prev + limit)}
              className={`px-2 py-1 rounded ${
                users.length < limit ? "text-gray-300 cursor-not-allowed" : "hover:bg-gray-200"
              }`}
            >
              下一页
            </button>
          </div>
        </div>

        <div className="space-y-4">
          <div className="bg-white border rounded-xl p-4 space-y-4">
            <div className="flex items-center gap-2 text-sm font-medium text-gray-700">
              <UserCog className="w-4 h-4 text-blue-600" />
              用户详情
            </div>

            {!selectedUser ? (
              <div className="text-sm text-gray-500">请选择一个用户查看详情</div>
            ) : (
              <>
                <div className="space-y-1 text-sm">
                  <div className="flex items-center justify-between">
                    <span className="text-gray-500">用户 ID</span>
                    <span className="font-medium text-gray-900">{selectedUser.id}</span>
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-gray-500">创建时间</span>
                    <span className="text-gray-900">{formatDateTime(selectedUser.created_at)}</span>
                  </div>
                </div>

                {notice && (
                  <div
                    className={`text-sm rounded-lg px-3 py-2 ${
                      notice.type === "success" ? "bg-green-50 text-green-700" : "bg-red-50 text-red-600"
                    }`}
                  >
                    {notice.text}
                  </div>
                )}

                <div className="space-y-3">
                  <div>
                    <label className="text-xs text-gray-500">邮箱</label>
                    <input
                      value={editEmail}
                      onChange={(e) => setEditEmail(e.target.value)}
                      className="mt-1 w-full px-3 py-2 rounded-lg border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                    />
                  </div>
                  <div className="flex items-center justify-between gap-3">
                    <div className="flex-1">
                      <label className="text-xs text-gray-500">账号状态</label>
                      <select
                        value={editStatus}
                        onChange={(e) => setEditStatus(e.target.value as "active" | "disabled")}
                        className="mt-1 w-full px-3 py-2 rounded-lg border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                      >
                        <option value="active">正常</option>
                        <option value="disabled">禁用</option>
                      </select>
                    </div>
                    <div className="flex-1">
                      <label className="text-xs text-gray-500">管理员</label>
                      <button
                        onClick={() => setEditIsAdmin((prev) => !prev)}
                        className={`mt-1 w-full px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                          editIsAdmin
                            ? "bg-blue-600 text-white"
                            : "bg-gray-100 text-gray-600 hover:bg-gray-200"
                        }`}
                      >
                        {editIsAdmin ? "已启用" : "未启用"}
                      </button>
                    </div>
                  </div>
                  <button
                    onClick={handleSaveProfile}
                    disabled={!profileDirty || savingProfile}
                    className={`w-full flex items-center justify-center gap-2 py-2 rounded-lg text-sm font-medium ${
                      !profileDirty || savingProfile
                        ? "bg-gray-200 text-gray-500 cursor-not-allowed"
                        : "bg-blue-600 text-white hover:bg-blue-700"
                    }`}
                  >
                    {savingProfile ? <RefreshCw className="w-4 h-4 animate-spin" /> : <CheckCircle className="w-4 h-4" />}
                    保存用户信息
                  </button>
                </div>

                <div className="space-y-3 border-t pt-4">
                  <div className="flex items-center gap-2 text-sm font-medium text-gray-700">
                    <KeyRound className="w-4 h-4 text-purple-600" />
                    重置密码
                  </div>
                  <input
                    type="password"
                    value={resetPassword}
                    onChange={(e) => setResetPassword(e.target.value)}
                    placeholder="新密码（至少 6 位）"
                    className="w-full px-3 py-2 rounded-lg border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-purple-500"
                  />
                  <button
                    onClick={handleResetPassword}
                    disabled={resettingPassword || resetPassword.trim().length < 6}
                    className={`w-full flex items-center justify-center gap-2 py-2 rounded-lg text-sm font-medium ${
                      resettingPassword || resetPassword.trim().length < 6
                        ? "bg-gray-200 text-gray-500 cursor-not-allowed"
                        : "bg-purple-600 text-white hover:bg-purple-700"
                    }`}
                  >
                    {resettingPassword ? <RefreshCw className="w-4 h-4 animate-spin" /> : <UserX className="w-4 h-4" />}
                    重置密码并注销登录
                  </button>
                </div>

                <div className="space-y-3 border-t pt-4">
                  <div className="flex items-center gap-2 text-sm font-medium text-gray-700">
                    <CreditCard className="w-4 h-4 text-emerald-600" />
                    套餐与计费
                  </div>
                  <div className="text-xs text-gray-500">
                    当前套餐：{usage?.plan?.name || "-"} ({usage?.plan?.id || "未知"})
                  </div>
                  <input
                    value={planId}
                    onChange={(e) => setPlanId(e.target.value)}
                    placeholder="输入套餐 ID，例如 free"
                    className="w-full px-3 py-2 rounded-lg border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500"
                  />
                  <button
                    onClick={handleUpdatePlan}
                    disabled={updatingPlan || !planId.trim()}
                    className={`w-full flex items-center justify-center gap-2 py-2 rounded-lg text-sm font-medium ${
                      updatingPlan || !planId.trim()
                        ? "bg-gray-200 text-gray-500 cursor-not-allowed"
                        : "bg-emerald-600 text-white hover:bg-emerald-700"
                    }`}
                  >
                    {updatingPlan ? <RefreshCw className="w-4 h-4 animate-spin" /> : <CreditCard className="w-4 h-4" />}
                    更新套餐
                  </button>
                </div>
              </>
            )}
          </div>

          <div className="bg-white border rounded-xl p-4 space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2 text-sm font-medium text-gray-700">
                <CheckCircle className="w-4 h-4 text-green-600" />
                用量统计
              </div>
              {selectedUser && (
                <button
                  onClick={() => loadUsage(selectedUser.id)}
                  disabled={usageLoading}
                  className="text-xs text-gray-500 hover:text-gray-700"
                >
                  {usageLoading ? "刷新中..." : "刷新"}
                </button>
              )}
            </div>

            {usageError && (
              <div className="text-xs text-red-600 bg-red-50 rounded-lg px-3 py-2">{usageError}</div>
            )}

            {!selectedUser ? (
              <div className="text-xs text-gray-500">请先选择用户</div>
            ) : usageLoading ? (
              <div className="flex items-center justify-center py-6">
                <div className="w-6 h-6 border-4 border-green-200 border-t-green-600 rounded-full animate-spin" />
              </div>
            ) : !usage ? (
              <div className="text-xs text-gray-500">暂无用量数据</div>
            ) : (
              <div className="space-y-3 text-xs text-gray-600">
                <div>
                  计费周期：{formatDateTime(usage.period_start)} - {formatDateTime(usage.period_end)}
                </div>
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span>LLM 字符</span>
                    <span>
                      {formatNumber(llmUsed)} / {formatNumber(llmQuota)}
                    </span>
                  </div>
                  <div className="h-2 bg-gray-100 rounded-full">
                    <div
                      className="h-2 bg-blue-500 rounded-full"
                      style={{ width: `${percent(llmUsed, llmQuota)}%` }}
                    />
                  </div>
                </div>
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span>TTS 字符</span>
                    <span>
                      {formatNumber(ttsUsed)} / {formatNumber(ttsQuota)}
                    </span>
                  </div>
                  <div className="h-2 bg-gray-100 rounded-full">
                    <div
                      className="h-2 bg-emerald-500 rounded-full"
                      style={{ width: `${percent(ttsUsed, ttsQuota)}%` }}
                    />
                  </div>
                </div>
                <div className="text-xs text-gray-500">
                  当前订阅状态：{usage.subscription?.status || "-"}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default AdminSection;
