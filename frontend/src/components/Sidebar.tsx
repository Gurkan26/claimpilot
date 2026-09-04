import React from 'react'
import {
  LayoutDashboard,
  Clock,
  ArrowLeftRight,
  FileText,
  ShieldCheck,
  Zap,
  Building2,
  User,
  LogOut,
  Cpu,
} from 'lucide-react'
import { useAuth } from '../context/AuthContext'

export type TabType = 'dashboard' | 'obligations' | 'marketplace' | 'documents' | 'audit' | 'admin'

interface SidebarProps {
  activeTab: TabType
  onTabChange: (tab: TabType) => void
  pendingCount: number
  dealsCount: number
}

export const Sidebar: React.FC<SidebarProps> = ({
  activeTab,
  onTabChange,
  pendingCount,
  dealsCount,
}) => {
  const { user, switchAccountType, logout, isAdmin, t } = useAuth()

  return (
    <aside className="app-sidebar">
      <div>
        {/* Account Mode Segmented Toggle */}
        <div className="account-mode-toggle">
          <button
            className={`mode-btn ${user.accountType === 'b2b' ? 'active' : ''}`}
            onClick={() => switchAccountType('b2b')}
          >
            <Building2 size={12} />
            <span>{t.corporate}</span>
          </button>
          <button
            className={`mode-btn ${user.accountType === 'b2c' ? 'active' : ''}`}
            onClick={() => switchAccountType('b2c')}
          >
            <User size={12} />
            <span>{t.personal}</span>
          </button>
        </div>

        {/* Navigation Items */}
        <div className="sidebar-nav">
          <button
            className={`nav-item ${activeTab === 'dashboard' ? 'active' : ''}`}
            onClick={() => onTabChange('dashboard')}
          >
            <div className="nav-item-content">
              <LayoutDashboard size={16} />
              <span>{t.dashboard}</span>
            </div>
          </button>

          <button
            className={`nav-item ${activeTab === 'obligations' ? 'active' : ''}`}
            onClick={() => onTabChange('obligations')}
          >
            <div className="nav-item-content">
              <Clock size={16} />
              <span>{t.obligations}</span>
            </div>
            {pendingCount > 0 && <span className="nav-badge">{pendingCount}</span>}
          </button>

          <button
            className={`nav-item ${activeTab === 'marketplace' ? 'active' : ''}`}
            onClick={() => onTabChange('marketplace')}
          >
            <div className="nav-item-content">
              <ArrowLeftRight size={16} />
              <span>{t.marketplace}</span>
            </div>
            {dealsCount > 0 && (
              <span className="nav-badge" style={{ backgroundColor: 'rgba(16, 185, 129, 0.12)', color: '#10b981', borderColor: 'rgba(16, 185, 129, 0.3)' }}>
                {dealsCount}
              </span>
            )}
          </button>

          <button
            className={`nav-item ${activeTab === 'documents' ? 'active' : ''}`}
            onClick={() => onTabChange('documents')}
          >
            <div className="nav-item-content">
              <FileText size={16} />
              <span>{t.documents}</span>
            </div>
          </button>

          <button
            className={`nav-item ${activeTab === 'audit' ? 'active' : ''}`}
            onClick={() => onTabChange('audit')}
          >
            <div className="nav-item-content">
              <ShieldCheck size={16} />
              <span>{t.auditTrail}</span>
            </div>
          </button>

          {/* Admin & Agent Harness Tab */}
          <button
            className={`nav-item ${activeTab === 'admin' ? 'active admin-active' : ''}`}
            onClick={() => onTabChange('admin')}
            style={{
              marginTop: 10,
              background: activeTab === 'admin' ? 'rgba(168, 85, 247, 0.18)' : 'rgba(168, 85, 247, 0.05)',
              borderColor: activeTab === 'admin' ? '#a855f7' : 'rgba(168, 85, 247, 0.2)',
            }}
          >
            <div className="nav-item-content">
              <Cpu size={16} color={activeTab === 'admin' ? '#c084fc' : '#a855f7'} />
              <span style={{ color: activeTab === 'admin' ? '#f3e8ff' : '#d8b4fe', fontWeight: 600 }}>
                {t.adminPanel}
              </span>
            </div>
            <span
              className="nav-badge"
              style={{
                backgroundColor: 'rgba(168, 85, 247, 0.25)',
                color: '#e9d5ff',
                borderColor: 'rgba(168, 85, 247, 0.4)',
                fontSize: 9,
                fontWeight: 700,
                letterSpacing: 0.5,
              }}
            >
              {isAdmin ? 'ADMIN' : 'HARNESS'}
            </span>
          </button>
        </div>
      </div>

      <div className="sidebar-footer">
        <div className="agent-indicator">
          <div className="agent-avatar">CP</div>
          <div className="agent-info">
            <span className="agent-name">{t.activeAiEmployee}</span>
            <span className="agent-sub">
              <Zap size={10} style={{ display: 'inline', marginRight: 2 }} color="#38bdf8" />
              {t.mcpReady}
            </span>
          </div>
        </div>

        <button className="sidebar-logout-btn" onClick={logout}>
          <LogOut size={14} />
          <span>{t.signOut}</span>
        </button>
      </div>
    </aside>
  )
}
