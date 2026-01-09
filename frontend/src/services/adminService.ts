import { apiClient } from "./clients";

export interface AdminUser {
  id: string;
  email: string;
  is_admin: boolean;
  status: "active" | "disabled";
  created_at?: string;
}

export interface AdminPlanInfo {
  id: string;
  name: string;
  monthly_llm_quota: number;
  monthly_tts_quota: number;
  price_cents: number;
  currency: string;
}

export interface AdminSubscriptionInfo {
  plan_id: string;
  status: string;
  current_period_start: string;
  current_period_end: string;
  cancel_at_period_end: boolean;
}

export interface AdminUsageSummary {
  period_start: string;
  period_end: string;
  used: Record<string, number>;
  quota: Record<string, number>;
  plan: AdminPlanInfo;
  subscription: AdminSubscriptionInfo;
}

export interface AdminListResponse<T> {
  data: T;
  message?: string;
  timestamp?: string;
}

export const adminService = {
  listUsers: (limit: number, offset: number) =>
    apiClient.getAdminUsers(limit, offset) as Promise<AdminListResponse<AdminUser[]>>,
  getUser: (userId: string) =>
    apiClient.getAdminUser(userId) as Promise<AdminListResponse<AdminUser>>,
  updateUser: (
    userId: string,
    payload: { email?: string; is_admin?: boolean; status?: string }
  ) =>
    apiClient.updateAdminUser(userId, payload) as Promise<AdminListResponse<AdminUser>>,
  resetPassword: (userId: string, password: string) =>
    apiClient.resetAdminPassword(userId, password) as Promise<AdminListResponse<null>>,
  getUsage: (userId: string) =>
    apiClient.getAdminUserUsage(userId) as Promise<AdminListResponse<AdminUsageSummary>>,
  updatePlan: (userId: string, planId: string) =>
    apiClient.updateAdminUserPlan(userId, planId) as Promise<AdminListResponse<null>>,
};

export default adminService;
