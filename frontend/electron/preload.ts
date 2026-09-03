import { contextBridge, ipcRenderer } from 'electron'

export interface DesktopAPI {
  minimize: () => Promise<void>
  maximize: () => Promise<void>
  close: () => Promise<void>
  isMaximized: () => Promise<boolean>
  getPlatform: () => Promise<string>
  openFileDialog: () => Promise<string | null>
  notify: (title: string, body: string) => Promise<void>
}

const desktopAPI: DesktopAPI = {
  minimize: () => ipcRenderer.invoke('window-minimize'),
  maximize: () => ipcRenderer.invoke('window-maximize'),
  close: () => ipcRenderer.invoke('window-close'),
  isMaximized: () => ipcRenderer.invoke('is-maximized'),
  getPlatform: () => ipcRenderer.invoke('get-platform'),
  openFileDialog: () => ipcRenderer.invoke('open-file-dialog'),
  notify: (title: string, body: string) => ipcRenderer.invoke('notify', title, body),
}

contextBridge.exposeInMainWorld('claimpilotDesktop', desktopAPI)
