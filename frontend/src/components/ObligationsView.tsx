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
  RefreshCw,
} from 'lucide-react'
import { Obligation } from '../types'
import { useAuth } from '../context/AuthContext'

interface ObligationsViewProps {
  obligations: Obligation[]
  onApprove: (id: string) => void
  onDismiss: (id: string, reason: string) => void
  onRenew?: (id: string) => void
  onUpdateStatus?: (id: string, status: string) => void
}

export const ObligationsView: React.FC<ObligationsViewProps> = ({
  obligations,
  onApprove,
  onDismiss,
  onRenew,
  onUpdateStatus,
}) => {
  const { user, t } = useAuth()
  const [filterDays, setFilterDays] = useState<number>(30)
  const [searchQuery, setSearchQuery] = useState('')
  const [dismissingId, setDismissingId] = useState<string | null>(null)
  const [dismissReason, setDismissReason] = useState('')

  const activeList = obligations

  const filtered = activeList.filter((o) => {
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

  return (
    <div>
      <div className="page-header">
        <div className="page-title-group">
          <h1>{t.obligationsTitle}</h1>
          <p className="page-subtitle">{t.obligationsSubtitle}</p>
        </div>
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

                return (
                  <tr key={o.id} className={dismissingId === o.id ? 'row-highlight' : ''}>
                    <td>
                      <div style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{o.title}</div>
                      <div style={{ fontSize: 11, color: 'var(--text-muted)', marginTop: 2 }}>
                        {o.description.slice(0, 52)}...
                      </div>
                    </td>
                    <td>{o.type}</td>
                    <td>
                      <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}>
                        {new Date(o.dueDate).toLocaleDateString()}
                        <span style={{ color: daysRemaining <= 3 ? '#ef4444' : '#f59e0b', fontWeight: 600 }}>
                          ({daysRemaining} {t.daysLeft})
                        </span>
                      </span>
                    </td>
                    <td className="card-value mono" style={{ fontSize: 13 }}>
                      {o.amount ? `${o.amount.amount.toLocaleString()} ${o.amount.currency}` : '—'}
                    </td>
                    <td>
                      <span className={`badge badge-${o.riskLevel.toLowerCase()}`}>{o.riskLevel}</span>
                    </td>
                    <td>
                      <select
                        value={o.status}
                        onChange={(e) => onUpdateStatus && onUpdateStatus(o.id, e.target.value)}
                        style={{
                          background:
                            o.status === 'APPROVED' || o.status === 'RESOLVED' ? 'rgba(16, 185, 129, 0.15)' :
                            o.status === 'RENEWED' ? 'rgba(6, 182, 212, 0.15)' :
                            o.status === 'PENDING_APPROVAL' ? 'rgba(245, 158, 11, 0.15)' :
                            o.status === 'IN_PROGRESS' ? 'rgba(59, 130, 246, 0.15)' :
                            'rgba(148, 163, 184, 0.15)',
                          color:
                            o.status === 'APPROVED' || o.status === 'RESOLVED' ? '#34d399' :
                            o.status === 'RENEWED' ? '#38bdf8' :
                            o.status === 'PENDING_APPROVAL' ? '#fbbf24' :
                            o.status === 'IN_PROGRESS' ? '#60a5fa' :
                            '#94a3b8',
                          border: '1px solid rgba(255, 255, 255, 0.12)',
                          borderRadius: 6,
                          padding: '3px 8px',
                          fontSize: 11,
                          fontWeight: 600,
                          cursor: 'pointer',
                          outline: 'none',
                        }}
                        title="Durumu güncellemek için tıklayın"
                      >
                        <option value="PENDING_APPROVAL" style={{ background: '#090d16', color: '#fbbf24' }}>PENDING_APPROVAL</option>
                        <option value="APPROVED" style={{ background: '#090d16', color: '#34d399' }}>APPROVED</option>
                        <option value="RENEWED" style={{ background: '#090d16', color: '#38bdf8' }}>RENEWED</option>
                        <option value="IN_PROGRESS" style={{ background: '#090d16', color: '#60a5fa' }}>IN_PROGRESS</option>
                        <option value="RESOLVED" style={{ background: '#090d16', color: '#34d399' }}>RESOLVED</option>
                        <option value="DISMISSED" style={{ background: '#090d16', color: '#94a3b8' }}>DISMISSED</option>
                      </select>
                    </td>
                    <td>
                      {o.suggestedAction ? (
                        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5, fontSize: 12 }}>
                          {getAdapterIcon(o.suggestedAction.mcpAdapter)}
                          <span style={{ textTransform: 'capitalize' }}>{o.suggestedAction.mcpAdapter}</span>
                        </span>
                      ) : (
                        '—'
                      )}
                    </td>
                    <td style={{ textAlign: 'right' }}>
                      <div style={{ display: 'inline-flex', gap: 6, alignItems: 'center' }}>
                        {/* Yenileme (Renew) Button */}
                        <button
                          className="btn btn-secondary"
                          style={{
                            fontSize: 11,
                            padding: '4px 9px',
                            borderColor: 'rgba(6, 182, 212, 0.4)',
                            color: '#38bdf8',
                            display: 'inline-flex',
                            alignItems: 'center',
                            gap: 4,
                          }}
                          title={t.renew || 'Yenile'}
                          onClick={() => onRenew && onRenew(o.id)}
                        >
                          <RefreshCw size={12} /> {t.renew || 'Yenile'}
                        </button>

                        {o.status === 'PENDING_APPROVAL' && (
                          <>
                            <button
                              className="btn btn-primary"
                              style={{ fontSize: 11, padding: '4px 8px' }}
                              onClick={() => onApprove(o.id)}
                            >
                              <CheckCircle size={12} /> {t.approve}
                            </button>
                            <button
                              className="btn btn-secondary"
                              style={{ fontSize: 11, padding: '4px 8px' }}
                              onClick={() => handleDismissClick(o.id)}
                            >
                              <XCircle size={12} />
                            </button>
                          </>
                        )}
                        {(o.status === 'APPROVED' || o.status === 'RESOLVED') && (
                          <span style={{ color: '#10b981', fontSize: 11, display: 'inline-flex', alignItems: 'center', gap: 3 }}>
                            <CheckCircle size={12} /> {t.approved}
                          </span>
                        )}
                        {o.status === 'RENEWED' && (
                          <span style={{ color: '#38bdf8', fontSize: 11, display: 'inline-flex', alignItems: 'center', gap: 3 }}>
                            <CheckCircle size={12} /> {t.renewed || 'Yenilendi'}
                          </span>
                        )}
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
