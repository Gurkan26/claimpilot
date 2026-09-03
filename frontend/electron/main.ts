import { app, BrowserWindow, ipcMain, dialog, Notification } from 'electron'
import path from 'path'

process.env.DIST = path.join(__dirname, '../dist')
process.env.VITE_PUBLIC = app.isPackaged ? process.env.DIST : path.join(process.env.DIST, '../public')

let win: BrowserWindow | null = null
const isMac = process.platform === 'darwin'

function createWindow() {
  win = new BrowserWindow({
    width: 1280,
    height: 840,
    minWidth: 1024,
    minHeight: 680,
    title: 'ClaimPilot',
    backgroundColor: '#090d16',
    titleBarStyle: isMac ? 'hiddenInset' : 'hidden',
    titleBarOverlay: !isMac ? {
      color: '#090d16',
      symbolColor: '#94a3b8',
      height: 36,
    } : false,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      nodeIntegration: false,
      contextIsolation: true,
      sandbox: false,
      webSecurity: false,
    },
  })

  // Load Vite Dev Server in development or index.html in production
  if (process.env.VITE_DEV_SERVER_URL) {
    win.loadURL(process.env.VITE_DEV_SERVER_URL)
  } else {
    win.loadFile(path.join(__dirname, '../dist/index.html'))
  }

  win.on('closed', () => {
    win = null
  })
}

// Window control IPC handlers
ipcMain.handle('window-minimize', () => {
  win?.minimize()
})

ipcMain.handle('window-maximize', () => {
  if (win?.isMaximized()) {
    win.unmaximize()
  } else {
    win?.maximize()
  }
})

ipcMain.handle('window-close', () => {
  win?.close()
})

ipcMain.handle('is-maximized', () => {
  return win?.isMaximized() ?? false
})

ipcMain.handle('get-platform', () => {
  return process.platform
})

// Native file picker IPC handler
ipcMain.handle('open-file-dialog', async () => {
  if (!win) return null

  const result = await dialog.showOpenDialog(win, {
    title: 'Select Contract, Invoice, or License Document',
    properties: ['openFile'],
    filters: [
      { name: 'Supported Documents', extensions: ['pdf', 'txt', 'csv', 'md', 'json'] },
      { name: 'PDF Documents', extensions: ['pdf'] },
      { name: 'All Files', extensions: ['*'] },
    ],
  })

  if (result.canceled || result.filePaths.length === 0) {
    return null
  }

  return result.filePaths[0]
})

// Native OS desktop notification
ipcMain.handle('notify', (_event, title: string, body: string) => {
  if (Notification.isSupported()) {
    new Notification({
      title: title || 'ClaimPilot Alert',
      body: body || '',
      silent: false,
    }).show()
  }
})

app.whenReady().then(createWindow)

app.on('window-all-closed', () => {
  if (!isMac) {
    app.quit()
  }
})

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0) {
    createWindow()
  }
})
