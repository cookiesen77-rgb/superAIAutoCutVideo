import { apiClient, getStoredTokens, isTokenExpired, setStoredTokens } from "./clients";

export const authService = {
  isAuthenticated(): boolean {
    const tokens = getStoredTokens();
    return !!tokens && !isTokenExpired(tokens);
  },

  async login(email: string, password: string) {
    const tokens = await apiClient.login(email, password);
    setStoredTokens(tokens);
    return tokens;
  },

  async register(email: string, password: string) {
    const tokens = await apiClient.register(email, password);
    setStoredTokens(tokens);
    return tokens;
  },

  async logout() {
    await apiClient.logout();
  },
};
