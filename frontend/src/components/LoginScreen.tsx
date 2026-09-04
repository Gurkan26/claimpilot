import React, { useState } from 'react'
import {
  Shield,
  Sparkles,
  Building2,
  User,
  CheckCircle,
  ArrowRight,
  FileText,
  Clock,
  BarChart3,
  ShieldCheck,
  Cpu,
  KeyRound,
  AlertCircle,
} from 'lucide-react'
import { useAuth } from '../context/AuthContext'

export const LoginScreen: React.FC = () => {
  const { login, adminLogin, t, language, setLanguage } = useAuth()

  const [accountType, setAccountType] = useState<'b2b' | 'b2c'>('b2b')
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [showCustomForm, setShowCustomForm] = useState(false)
  const [showAdminLogin, setShowAdminLogin] = useState(false)
  const [adminPassword, setAdminPassword] = useState('admin123')
  const [adminError, setAdminError] = useState('')

  const handleQuickCorporate = () => {
    login({
      name: 'Gürkan Şentürk',
      email: 'gurkan@acmeholding.com.tr',
      accountType: 'b2b',
      organization: 'Acme Holding A.Ş.',
      role: 'Head of Procurement & IT',
      avatar: 'GS',
    })
  }

  const handleQuickPersonal = () => {
    login({
      name: 'Gürkan Şentürk',
      email: 'gurkan.senturk@gmail.com',
      accountType: 'b2c',
      organization: 'Bireysel Portföy',
      role: 'Hesap Sahibi',
      avatar: 'GŞ',
    })
  }

  const handleAdminSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setAdminError('')
    const success = await adminLogin(adminPassword)
    if (!success) {
      setAdminError(t.adminWrongPassword)
    }
  }

  const handleCustomLogin = (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim() || !email.trim()) return
    login({
      name: name.trim(),
      email: email.trim(),
      accountType,
      organization: accountType === 'b2b' ? 'Kurumsal Şirket' : 'Bireysel Portföy',
      role: accountType === 'b2b' ? 'Operasyon Yöneticisi' : 'Kullanıcı',
      avatar: name.split(' ').map((n) => n[0]).join('').slice(0, 2).toUpperCase() || 'CP',
    })
  }

  const features = [
    { icon: <FileText size={18} />, text: t.loginFeature1 },
    { icon: <Clock size={18} />, text: t.loginFeature2 },
    { icon: <BarChart3 size={18} />, text: t.loginFeature3 },
    { icon: <ShieldCheck size={18} />, text: t.loginFeature4 },
  ]

  return (
    <div className="login-screen">
      {/* Animated background elements */}
      <div className="login-bg-orb login-bg-orb-1" />
      <div className="login-bg-orb login-bg-orb-2" />
      <div className="login-bg-orb login-bg-orb-3" />

      {/* Language switcher */}
      <div className="login-lang-switcher">
        <button
          className={`lang-btn ${language === 'tr' ? 'active' : ''}`}
          onClick={() => setLanguage('tr')}
        >
          TR
        </button>
        <span style={{ color: 'rgba(255,255,255,0.2)', fontSize: 10 }}>|</span>
        <button
          className={`lang-btn ${language === 'en' ? 'active' : ''}`}
          onClick={() => setLanguage('en')}
        >
          EN
        </button>
      </div>

      <div className="login-container">
        {/* Left Panel — Branding */}
        <div className="login-branding">
          <div className="login-logo">
            <div className="login-logo-icon">
              <Shield size={32} />
            </div>
            <div className="login-logo-text">
              <h1>{t.loginTitle}</h1>
              <span className="login-logo-badge">AI Employee</span>
            </div>
          </div>

          <p className="login-tagline">{t.loginSubtitle}</p>

          <div className="login-features">
            {features.map((feat, i) => (
              <div key={i} className="login-feature-item" style={{ animationDelay: `${0.3 + i * 0.1}s` }}>
                <div className="login-feature-icon">{feat.icon}</div>
                <span>{feat.text}</span>
              </div>
            ))}
          </div>

          <div className="login-branding-footer">
            <Sparkles size={14} color="#38bdf8" />
            <span>Powered by MCP Protocol & Multi-Agent AI</span>
          </div>
        </div>

        {/* Right Panel — Login Form */}
        <div className="login-form-panel">
          <div className="login-form-card">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
              <h2 className="login-form-title" style={{ margin: 0 }}>
                {showAdminLogin ? t.adminLoginTitle : t.loginWelcome}
              </h2>
              <button
                type="button"
                className={`mode-btn ${showAdminLogin ? 'active' : ''}`}
                style={{
                  fontSize: 11,
                  padding: '4px 10px',
                  borderRadius: 6,
                  display: 'flex',
                  alignItems: 'center',
                  gap: 5,
                  background: showAdminLogin ? 'rgba(168, 85, 247, 0.25)' : 'rgba(255, 255, 255, 0.05)',
                  borderColor: showAdminLogin ? '#a855f7' : 'rgba(255, 255, 255, 0.1)',
                  color: showAdminLogin ? '#d8b4fe' : 'var(--text-secondary)',
                }}
                onClick={() => {
                  setShowAdminLogin(!showAdminLogin)
                  setAdminError('')
                }}
              >
                <Cpu size={13} color={showAdminLogin ? '#c084fc' : '#94a3b8'} />
                <span>{showAdminLogin ? 'Kullanıcı Girişi' : 'Admin & Harness'}</span>
              </button>
            </div>

            <p className="login-form-subtitle">
              {showAdminLogin
                ? 'Dual-LLM (Ollama) & DeepWiki MCP bağlantı orkestrasyonu için yönetici girişi.'
                : t.loginQuickAccess}
            </p>

            {showAdminLogin ? (
              /* Admin Password Form */
              <form onSubmit={handleAdminSubmit} className="admin-login-card" style={{ marginTop: 14 }}>
                <div
                  style={{
                    background: 'rgba(168, 85, 247, 0.07)',
                    border: '1px solid rgba(168, 85, 247, 0.25)',
                    borderRadius: 10,
                    padding: '14px 16px',
                    marginBottom: 16,
                  }}
                >
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
                    <KeyRound size={16} color="#c084fc" />
                    <span style={{ fontWeight: 600, fontSize: 13, color: '#e9d5ff' }}>
                      {t.adminPasswordLabel}
                    </span>
                  </div>
                  <input
                    type="password"
                    className="form-input"
                    value={adminPassword}
                    onChange={(e) => setAdminPassword(e.target.value)}
                    placeholder={t.adminPasswordPlaceholder}
                    required
                    style={{
                      width: '100%',
                      background: 'rgba(15, 23, 42, 0.8)',
                      borderColor: adminError ? '#ef4444' : 'rgba(168, 85, 247, 0.3)',
                    }}
                  />
                  {adminError && (
                    <div style={{ display: 'flex', alignItems: 'center', gap: 5, color: '#f87171', fontSize: 12, marginTop: 6 }}>
                      <AlertCircle size={13} />
                      <span>{adminError}</span>
                    </div>
                  )}
                  <div style={{ fontSize: 11, color: '#94a3b8', marginTop: 8 }}>
                    Varsayılan şifre: <code style={{ color: '#c084fc' }}>admin123</code> (Ollama çift LLM + MCP kontrolü)
                  </div>
                </div>

                <button
                  type="submit"
                  className="btn btn-primary"
                  style={{
                    width: '100%',
                    background: 'linear-gradient(135deg, #9333ea 0%, #6366f1 100%)',
                    boxShadow: '0 4px 14px rgba(147, 51, 234, 0.35)',
                    border: 'none',
                    padding: '11px',
                    fontSize: 14,
                    fontWeight: 600,
                  }}
                >
                  <Cpu size={16} /> {t.adminLoginButton}
                </button>
              </form>
            ) : (
              <>
                {/* Quick Demo Login Cards */}
                <div className="login-quick-cards">
                  <button
                    type="button"
                    className="login-quick-card"
                    onClick={handleQuickCorporate}
                  >
                    <div className="login-quick-card-header">
                      <span className="badge badge-medium">{t.corporateBadge}</span>
                      <Building2 size={18} color="#38bdf8" />
                    </div>
                    <div className="login-quick-card-name">Acme Holding A.Ş.</div>
                    <div className="login-quick-card-desc">{t.corporateDesc}</div>
                    <div className="login-quick-card-action">
                      <ArrowRight size={14} />
                    </div>
                  </button>

                  <button
                    type="button"
                    className="login-quick-card"
                    onClick={handleQuickPersonal}
                  >
                    <div className="login-quick-card-header">
                      <span className="badge badge-success">{t.personalBadge}</span>
                      <User size={18} color="#10b981" />
                    </div>
                    <div className="login-quick-card-name">Gürkan Şentürk</div>
                    <div className="login-quick-card-desc">{t.personalDesc}</div>
                    <div className="login-quick-card-action">
                      <ArrowRight size={14} />
                    </div>
                  </button>
                </div>

                {/* Divider */}
                <div className="login-divider">
                  <span>{t.loginOrCustom}</span>
                </div>

                {/* Custom Login Toggle / Form */}
                {!showCustomForm ? (
                  <button
                    className="btn btn-secondary login-custom-toggle"
                    onClick={() => setShowCustomForm(true)}
                  >
                    {t.continueButton}
                  </button>
                ) : (
                  <form onSubmit={handleCustomLogin} className="login-custom-form">
                    {/* Account type toggle */}
                    <div className="account-mode-toggle" style={{ marginBottom: 14 }}>
                      <button
                        type="button"
                        className={`mode-btn ${accountType === 'b2b' ? 'active' : ''}`}
                        onClick={() => setAccountType('b2b')}
                      >
                        <Building2 size={12} />
                        <span>{t.corporate}</span>
                      </button>
                      <button
                        type="button"
                        className={`mode-btn ${accountType === 'b2c' ? 'active' : ''}`}
                        onClick={() => setAccountType('b2c')}
                      >
                        <User size={12} />
                        <span>{t.personal}</span>
                      </button>
                    </div>

                    <div style={{ display: 'flex', gap: 10 }}>
                      <div style={{ flex: 1 }}>
                        <label className="form-label">{t.fullName}</label>
                        <input
                          className="form-input"
                          type="text"
                          value={name}
                          onChange={(e) => setName(e.target.value)}
                          placeholder="Ad Soyad"
                          required
                        />
                      </div>
                      <div style={{ flex: 1 }}>
                        <label className="form-label">{t.email}</label>
                        <input
                          className="form-input"
                          type="email"
                          value={email}
                          onChange={(e) => setEmail(e.target.value)}
                          placeholder="email@example.com"
                          required
                        />
                      </div>
                    </div>

                    <button type="submit" className="btn btn-primary login-submit-btn">
                      <CheckCircle size={15} /> {t.continueButton}
                    </button>
                  </form>
                )}
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
