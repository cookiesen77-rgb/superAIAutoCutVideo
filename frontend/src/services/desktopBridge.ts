type ElectronApi = {
  notify: (payload: { title: string; body: string }) => Promise<void> | void;
  openExternal: (url: string) => Promise<void> | void;
  getAppInfo?: () => Promise<Record<string, string>> | Record<string, string>;
};

declare global {
  interface Window {
    electronAPI?: ElectronApi;
  }
}

export class DesktopCommands {
  static async showNotification(title: string, body: string): Promise<void> {
    try {
      if (window.electronAPI?.notify) {
        await window.electronAPI.notify({ title, body });
        return;
      }
      if ("Notification" in window) {
        if (Notification.permission === "granted") {
          new Notification(title, { body });
          return;
        }
        if (Notification.permission === "default") {
          const perm = await Notification.requestPermission();
          if (perm === "granted") {
            new Notification(title, { body });
            return;
          }
        }
      }
    } catch {
      // fall through to console
    }
    console.log(`[notify] ${title}: ${body}`);
  }

  static async openExternalLink(url: string): Promise<void> {
    try {
      if (window.electronAPI?.openExternal) {
        await window.electronAPI.openExternal(url);
        return;
      }
    } catch {
      // fall through to window.open
    }
    window.open(url, "_blank");
  }

  static async getAppInfo(): Promise<Record<string, string>> {
    try {
      if (window.electronAPI?.getAppInfo) {
        const info = await window.electronAPI.getAppInfo();
        return info || {};
      }
    } catch {
      // ignore
    }
    return {};
  }
}
