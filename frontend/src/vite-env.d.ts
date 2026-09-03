/// <reference types="vite/client" />

import { DesktopAPI } from '../electron/preload'

declare global {
  interface Window {
    claimpilotDesktop?: DesktopAPI
  }
}
