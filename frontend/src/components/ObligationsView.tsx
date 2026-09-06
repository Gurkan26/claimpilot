import React, { useState } from 'react'
import {
  Clock,
  Calendar,
  Mail,
  MessageSquare,
  CheckCircle,
  XCircle,
  Search,
  X,
  Sparkles,
  RefreshCw,
  Tag,
} from 'lucide-react'
import { Obligation } from '../types'
import { useAuth } from '../context/AuthContext'

interface ObligationsViewProps {
  obligations: Obligation[]
  onApprove: (id: string) => void
  onDismiss: (id: string, reason: string) => void
  onTriggerRFQ?: (id?: string) => Promise<any>
  onNavigateToMarketplace?: () => void
}

export const ObligationsView: React.FC<ObligationsViewProps> = ({
  obligations,
  onApprove,
  onDismiss,
  onTriggerRFQ,
  onNavigateToMarketplace,
}) => {
  const { user, t } = useAuth()
  const [filterDays, setFilterDays] = useState<number>(30)
  const [searchQuery, setSearchQuery] = useState('')
  const [dismissingId, setDismissingId] = useState<string | null>(null)
  const [dismissReason, setDismissReason] = useState('')
  const [loadingRfqId, setLoadingRfqId] = useState<string | null>(null)

  const isB2B = user.accountType === 'b2b'

  const filtered = obligations.filter((o) => {
    return (
      o.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      o.description.toLowerCase().includes(searchQuery.toLowerCase())
    )
  })

  const getAdapterIcon = (adapter?: string) => {
    switch (adapter) {
      case 'gmail':
        return <Mail size={13} color="#f87171" />
      case 'calendar':
        return <Calendar size={13} color="#38bdf8" />
      case 'slack':
        return <MessageSquare size={13} color="#fbbf24" />
      default:
        return <Clock size={13} color="#94a3b8" />
    }
  }

  const handleDismissClick = (id: string) => {
    setDismissingId(id)
    setDismissReason('')
  }

  const handleDismissConfirm = () => {
    if (dismissingId && dismissReason.trim()) {
      onDismiss(dismissingId, dismissReason.trim())
      setDismissingId(null)
      setDismissReason('')
    }
  }

  const handleDismissCancel = () => {
    setDismissingId(null)
    setDismissReason('')
  }

  const handleCollectQuotesForObligation = async (id?: string) => {
    if (id) setLoadingRfqId(id)
    try {
      if (onTriggerRFQ) {
        await onTriggerRFQ(id)
      }
      if (onNavigateToMarketplace) {
        onNavigateToMarketplace()
      }
    } finally {
      setLoadingRfqId(null)
    }
  }

  return (
    <div>
      <div className="page-header">
        <div className="page-title-group">
          <h1>{t.obligationsTitle}</h1>
          <p className="page-subtitle">{t.obligationsSubtitle}</p>
        </div>

        <button
          className="btn btn-secondary"
          onClick={() => handleCollectQuotesForObligation()}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: 6,
            fontSize: 12,
            padding: '8px 14px',
          }}
        >
          <Sparkles size={14} color="#38bdf8" />
          <span>🤖 AI ile Tüm Teklifleri Sorgula</span>
        </button>
      </div>

      {/* Filter and Search Bar */}
      <div style={{ display: 'flex', gap: 12, marginBottom: 20, alignItems: 'center' }}>
        <div style={{ position: 'relative', flex: 1 }}>
          <Search size={15} style={{ position: 'absolute', left: 12, top: 11, color: 'var(--text-muted)' }} />
          <input
            type="text"
            placeholder={t.searchPlaceholder}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="form-input"
            style={{
              width: '100%',
              paddingLeft: 36,
            }}
          />
        </div>

        <div style={{ display: 'flex', gap: 6 }}>
          {[7, 14, 30, 90].map((days) => (
            <button
              key={days}
              className={`btn ${filterDays === days ? 'btn-primary' : 'btn-secondary'}`}
              style={{ fontSize: 12, padding: '7px 12px' }}
              onClick={() => setFilterDays(days)}
            >
              {days} {t.dueInDays}
            </button>
          ))}
        </div>
      </div>

      {/* Dismiss Reason Inline Modal */}
      {dismissingId && (
        <div className="dismiss-reason-bar">
          <div className="dismiss-reason-content">
            <span className="dismiss-reason-label">{t.dismissReason}:</span>
            <input
              type="text"
              className="form-input dismiss-reason-input"
              placeholder={t.enterReason}
              value={dismissReason}
              onChange={(e) => setDismissReason(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleDismissConfirm()
                if (e.key === 'Escape') handleDismissCancel()
              }}
              autoFocus
            />
            <button
              className="btn btn-primary"
              style={{ fontSize: 12, padding: '5px 12px' }}
              onClick={handleDismissConfirm}
              disabled={!dismissReason.trim()}
            >
              {t.confirm}
            </button>
            <button
              className="btn btn-secondary"
              style={{ fontSize: 12, padding: '5px 10px' }}
              onClick={handleDismissCancel}
            >
              <X size={13} />
            </button>
          </div>
        </div>
      )}

      {/* Obligations Table */}
      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>{t.obligationCol}</th>
              <th>{t.categoryCol}</th>
              <th>{t.dueDateCol}</th>
              <th>{t.amountCol}</th>
              <th>{t.riskCol}</th>
              <th>{t.statusCol}</th>
              <th>{t.actionCol}</th>
              <th style={{ textAlign: 'right' }}>{t.actionsCol}</th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 ? (
              <tr>
                <td colSpan={8} style={{ textAlign: 'center', padding: '32px 0', color: 'var(--text-muted)' }}>
                  {t.noObligations}
                </td>
              </tr>
            ) : (
              filtered.map((o) => {
                const daysRemaining = Math.ceil(
                  (new Date(o.dueDate).getTime() - Date.now()) / (1000 * 3600 * 24)
                )

                const isCritical = daysRemaining <= 3 || o.riskLevel === 'CRITICAL'

                return (
                  <tr key={o.id} className={dismissingId === o.id ? 'row-highlight' : ''}>
                    <td>
                      <div style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{o.title}</div>
                      <div style={{ fontSize: 11, color: 'var(--text-muted)', marginTop: 2 }}>
                        {o.description.length > 60 ? `${o.description.slice(0, 60)}...` : o.description}
                      </div>
                    </td>

                    <td>
                      <span className="badge badge-medium" style={{ textTransform: 'uppercase', fontSize: 10 }}>
                        {o.type}
                      </span>
                    </td>

                    <td>
                      <div style={{ fontSize: 12, fontWeight: 600, color: isCritical ? '#ef4444' : 'var(--text-primary)' }}>
                        {new Date(o.dueDate).toLocaleDateString('tr-TR')}
                      </div>
                      <div style={{ fontSize: 11, color: isCritical ? '#ef4444' : 'var(--text-muted)' }}>
                        {daysRemaining <= 0 ? 'Bugün doluyor' : `${daysRemaining} ${t.daysLeft}`}
                      </div>
                    </td>

                    <td className="mono" style={{ fontWeight: 600 }}>
                      {o.amount ? `${o.amount.amount.toLocaleString()} ${o.amount.currency}` : '—'}
                    </td>

                    <td>
                      <span
                        className={`badge ${
                          o.riskLevel === 'CRITICAL'
                            ? 'badge-critical'
                            : o.riskLevel === 'HIGH'
                            ? 'badge-high'
                            : o.riskLevel === 'MEDIUM'
                            ? 'badge-medium'
                            : 'badge-low'
                        }`}
                      >
                        {o.riskLevel}
                      </span>
                    </td>

                    <td>
                      <span
                        className="badge"
                        style={{
                          backgroundColor:
                            o.status === 'APPROVED'
                              ? 'rgba(16, 185, 129, 0.15)'
                              : o.status === 'PENDING_APPROVAL'
                              ? 'rgba(245, 158, 11, 0.15)'
                              : 'rgba(56, 189, 248, 0.15)',
                          color:
                            o.status === 'APPROVED'
                              ? '#34d399'
                              : o.status === 'PENDING_APPROVAL'
                              ? '#fbbf24'
                              : '#38bdf8',
                        }}
                      >
                        {o.status === 'APPROVED' ? 'Aktif Takipte' : o.status === 'PENDING_APPROVAL' ? 'Onay Bekliyor' : o.status}
                      </span>
                    </td>

                    <td>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 12 }}>
                        {getAdapterIcon(o.suggestedAction?.mcpAdapter)}
                        <span>{o.suggestedAction?.description || 'Aksiyon tanımlı'}</span>
                      </div>
                    </td>

                    <td style={{ textAlign: 'right' }}>
                      <div style={{ display: 'flex', gap: 6, justifyContent: 'flex-end' }}>
                        <button
                          className="btn btn-secondary"
                          style={{ fontSize: 11, padding: '4px 8px', display: 'inline-flex', alignItems: 'center', gap: 4 }}
                          onClick={() => handleCollectQuotesForObligation(o.id)}
                          disabled={loadingRfqId === o.id}
                          title="Alternatif Tedarikçi Tekliflerini Sorgula"
                        >
                          <RefreshCw size={11} className={loadingRfqId === o.id ? 'spin' : ''} />
                          <span>Teklif Topla</span>
                        </button>

                        {o.status === 'PENDING_APPROVAL' && (
                          <button
                            className="btn btn-primary"
                            style={{ fontSize: 11, padding: '4px 10px' }}
                            onClick={() => onApprove(o.id)}
                          >
                            <CheckCircle size={12} /> {t.approve}
                          </button>
                        )}

                        <button
                          className="btn btn-secondary"
                          style={{ fontSize: 11, padding: '4px 8px', color: '#f87171' }}
                          onClick={() => handleDismissClick(o.id)}
                        >
                          <XCircle size={12} />
                        </button>
                      </div>
                    </td>
                  </tr>
                )
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
