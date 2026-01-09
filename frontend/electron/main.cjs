const { app, BrowserWindow, ipcMain, Notification, shell } = require("electron");
const path = require("path");
const { autoUpdater } = require("electron-updater");

const isDev = !app.isPackaged;
const devUrl = process.env.VITE_DEV_SERVER_URL || "http://localhost:1420";

const createWindow = () => {
  const win = new BrowserWindow({
    width: 1280,
    height: 800,
    minWidth: 1024,
    minHeight: 640,
    backgroundColor: "#f9fafb",
    webPreferences: {
      preload: path.join(__dirname, "preload.cjs"),
      contextIsolation: true,
      nodeIntegration: false,
    },
  });

  if (isDev) {
    win.loadURL(devUrl);
    win.webContents.openDevTools({ mode: "detach" });
  } else {
    win.loadFile(path.join(__dirname, "../dist/index.html"));
  }
};

app.whenReady().then(() => {
  createWindow();

  ipcMain.handle("app:notify", (_event, payload) => {
    if (!payload?.title) return;
    new Notification({ title: payload.title, body: payload.body || "" }).show();
  });

  ipcMain.handle("app:openExternal", (_event, url) => {
    if (typeof url === "string" && url.trim()) {
      return shell.openExternal(url);
    }
  });

  ipcMain.handle("app:getAppInfo", () => ({
    name: app.getName(),
    version: app.getVersion(),
    platform: process.platform,
  }));

  autoUpdater.checkForUpdatesAndNotify();

  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
});
