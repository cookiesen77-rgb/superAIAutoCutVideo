const { contextBridge, ipcRenderer } = require("electron");

contextBridge.exposeInMainWorld("electronAPI", {
  notify: (payload) => ipcRenderer.invoke("app:notify", payload),
  openExternal: (url) => ipcRenderer.invoke("app:openExternal", url),
  getAppInfo: () => ipcRenderer.invoke("app:getAppInfo"),
});
