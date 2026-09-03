import React, { useEffect, useState } from 'react'
import { Shield, Minus, Square, X, Copy, Globe, Building2, User } from 'lucide-react'
import { useAuth } from '../context/AuthContext'

interface TitleBarProps {
  isBackendOnline: boolean
}

export const TitleBar: React.FC<TitleBarProps> = ({ isBackendOnline }) => {
  const [isMac, setIsMac] = useState(true)
  const [isMaximized, setIsMaximized] = useState(false)
  const { user, language, setLanguage, setIsAuthModalOpen, t } = useAuth()

  useEffect(() => {
    if (window.claimpilotDesktop) {
      window.claimpilotDesktop.getPlatform().then((p) => {
        setIsMac(p === 'darwin')
      })
      window.claimpilotDesktop.isMaximized().then(setIsMaximized)
    }
  }, [])

  const handleMinimize = () => window.claimpilotDesktop?.minimize()
  const handleMaximize = () => {
    window.claimpilotDesktop?.maximize()
    setIsMaximized(!isMaximized)
  }
  const handleClose = () => window.claimpilotDesktop?.close()

  return (
    <header className={`desktop-titlebar ${isMac ? 'mac' : 'win'}`}>
      <div className="titlebar-brand">
        <Shield size={14} color="#38bdf8" />
        <span><strong>ClaimPilot</strong> Desktop</span>
        <span
          className="badge"
          style={{
            fontSize: 9,
            padding: '1px 5px',
            backgroundColor: user.accountType === 'b2b' ? 'rgba(56, 189, 248, 0.15)' : 'rgba(16, 185, 129, 0.15)',
            color: user.accountType === 'b2b' ? '#38bdf8' : '#34d399',
            borderColor: user.accountType === 'b2b' ? 'rgba(56, 189, 248, 0.3)' : 'rgba(16, 185, 129, 0.3)',
          }}
        >
          {user.accountType === 'b2b' ? t.corporateBadge : t.personalBadge}
        </span>
      </div>

      <div className="titlebar-actions">
        {/* Language Selector (TR / EN) */}
        <div className="lang-switcher">
          <Globe size={12} style={{ color: 'var(--text-muted)' }} />
          <button
            className={`lang-btn ${language === 'tr' ? 'active' : ''}`}
            onClick={() => setLanguage('tr')}
          >
            TR
          </button>
          <span style={{ color: 'var(--border-subtle)', fontSize: 10 }}>|</span>
          <button
            className={`lang-btn ${language === 'en' ? 'active' : ''}`}
            onClick={() => setLanguage('en')}
          >
            EN
          </button>
        </div>

        {/* Backend Connectivity Status */}
        <div className={`status-pill ${isBackendOnline ? 'online' : 'offline'}`}>
          <span className="status-dot" />
          <span>{isBackendOnline ? t.connected : t.demoMode}</span>
        </div>

        {/* User Profile / Account Switcher Pill */}
        <button
          className="user-pill-btn"
          onClick={() => setIsAuthModalOpen(true)}
          title={t.switchAccount}
        >
          <div className="user-avatar-mini">{user.avatar}</div>
          <div className="user-pill-text">
            <span className="user-pill-name">{user.name}</span>
            <span className="user-pill-org">
              {user.accountType === 'b2b' ? <Building2 size={9} style={{ display: 'inline', marginRight: 2 }} /> : <User size={9} style={{ display: 'inline', marginRight: 2 }} />}
              {user.organization}
            </span>
          </div>
        </button>

        {/* Windows / Linux control buttons */}
        {!isMac && (
          <div className="win-controls">
            <button className="win-btn" onClick={handleMinimize} title="Minimize">
              <Minus size={13} />
            </button>
            <button className="win-btn" onClick={handleMaximize} title={isMaximized ? 'Restore' : 'Maximize'}>
              {isMaximized ? <Copy size={12} /> : <Square size={12} />}
            </button>
            <button className="win-btn close" onClick={handleClose} title="Close">
              <X size={14} />
            </button>
          </div>
        )}
      </div>
    </header>
  )
}
