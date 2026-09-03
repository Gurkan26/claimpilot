import React, { useState } from 'react'
import { X, Building2, User, CheckCircle, Sparkles } from 'lucide-react'
import { useAuth, UserProfile } from '../context/AuthContext'

export const AuthModal: React.FC = () => {
  const { isAuthModalOpen, setIsAuthModalOpen, user, login, t } = useAuth()

  const [accountType, setAccountType] = useState<'b2b' | 'b2c'>(user.accountType)
  const [name, setName] = useState(user.name)
  const [email, setEmail] = useState(user.email)
  const [org, setOrg] = useState(user.organization)
  const [role, setRole] = useState(user.role)

  if (!isAuthModalOpen) return null

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

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    login({
      name,
      email,
      accountType,
      organization: org || (accountType === 'b2b' ? 'Kurumsal Şirket' : 'Bireysel Portföy'),
      role: role || (accountType === 'b2b' ? 'Operasyon Yöneticisi' : 'Kullanıcı'),
      avatar: name.split(' ').map((n) => n[0]).join('').slice(0, 2).toUpperCase() || 'CP',
    })
  }

  return (
    <div className="modal-backdrop">
      <div className="modal-card">
        <div className="modal-header">
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <div className="modal-icon-badge">
              <Sparkles size={18} color="#38bdf8" />
            </div>
            <div>
              <h2 style={{ fontSize: 17, fontWeight: 700, color: 'var(--text-primary)' }}>
                {t.authModalTitle}
              </h2>
              <p style={{ fontSize: 12, color: 'var(--text-secondary)', marginTop: 2 }}>
                {t.authModalSubtitle}
              </p>
            </div>
          </div>
          <button className="win-btn" onClick={() => setIsAuthModalOpen(false)}>
            <X size={16} />
          </button>
        </div>

        {/* Quick Demo Login Cards */}
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, marginBottom: 20 }}>
          <button
            type="button"
            className={`quick-login-card ${accountType === 'b2b' ? 'selected' : ''}`}
            onClick={handleQuickCorporate}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
              <span className="badge badge-medium">{t.corporateBadge}</span>
              <Building2 size={16} color="#38bdf8" />
            </div>
            <div style={{ fontWeight: 700, fontSize: 14, color: 'var(--text-primary)' }}>
              Acme Holding A.Ş.
            </div>
            <div style={{ fontSize: 11, color: 'var(--text-secondary)', marginTop: 4 }}>
              {t.corporateDesc}
            </div>
          </button>

          <button
            type="button"
            className={`quick-login-card ${accountType === 'b2c' ? 'selected' : ''}`}
            onClick={handleQuickPersonal}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
              <span className="badge badge-success">{t.personalBadge}</span>
              <User size={16} color="#10b981" />
            </div>
            <div style={{ fontWeight: 700, fontSize: 14, color: 'var(--text-primary)' }}>
              Gürkan Şentürk
            </div>
            <div style={{ fontSize: 11, color: 'var(--text-secondary)', marginTop: 4 }}>
              {t.personalDesc}
            </div>
          </button>
        </div>

        {/* Custom Form */}
        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div style={{ display: 'flex', gap: 12 }}>
            <div style={{ flex: 1 }}>
              <label className="form-label">{t.fullName}</label>
              <input
                className="form-input"
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
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
                required
              />
            </div>
          </div>

          <div style={{ display: 'flex', gap: 12 }}>
            <div style={{ flex: 1 }}>
              <label className="form-label">
                {accountType === 'b2b' ? t.companyName : 'Hesap Başlığı'}
              </label>
              <input
                className="form-input"
                type="text"
                value={org}
                onChange={(e) => setOrg(e.target.value)}
              />
            </div>
            <div style={{ flex: 1 }}>
              <label className="form-label">
                {accountType === 'b2b' ? t.department : 'Kullanıcı Rolü'}
              </label>
              <input
                className="form-input"
                type="text"
                value={role}
                onChange={(e) => setRole(e.target.value)}
              />
            </div>
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 10, marginTop: 12 }}>
            <button
              type="button"
              className="btn btn-secondary"
              onClick={() => setIsAuthModalOpen(false)}
            >
              {t.cancel}
            </button>
            <button type="submit" className="btn btn-primary">
              <CheckCircle size={14} /> {t.continueButton}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
