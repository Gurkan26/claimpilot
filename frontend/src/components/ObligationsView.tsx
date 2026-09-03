import React, { useState } from 'react'
import {
  Clock,
  Calendar,
  Mail,
  MessageSquare,
  CheckCircle,
  XCircle,
  Search,
} from 'lucide-react'
import { Obligation } from '../types'
import { useAuth } from '../context/AuthContext'

interface ObligationsViewProps {
  obligations: Obligation[]
  onApprove: (id: string) => void
  onDismiss: (id: string, reason: string) => void
}

export const ObligationsView: React.FC<ObligationsViewProps> = ({
  obligations,
  onApprove,
  onDismiss,
}) => {
  const { user, t } = useAuth()
  const [filterDays, setFilterDays] = useState<number>(30)
  const [searchQuery, setSearchQuery] = useState('')

  const isB2B = user.accountType === 'b2b'

  // Personal B2C demo items if user switched to personal
  const personalObligations: Obligation[] = [
    {
      id: 'b2c-001',
      documentId: 'doc-p-01',
      userId: '00000000-0000-0000-0000-000000000001',
      type: 'WARRANTY',
      title: 'Anadolu Sigorta Araç Kaskosu Yenileme',
      description: 'Poliçe vadesi 3 gün içinde doluyor. Hasarsızlık indirimi hakkı mevcuttur.',
      dueDate: new Date(Date.now() + 3 * 86400000).toISOString(),
      amount: { amount: 18500, currency: 'TRY' },
      status: 'PENDING_APPROVAL',
      riskLevel: 'CRITICAL',
      suggestedAction: {
        type: 'send_email',
        description: 'Aksigorta ve Sompo alternatif tekliflerini karşılaştır',
        mcpAdapter: 'gmail',
        suggestedAt: new Date().toISOString(),
      },
      autoApprove: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    },
    {
      id: 'b2c-002',
      documentId: 'doc-p-02',
      userId: '00000000-0000-0000-0000-000000000001',
      type: 'RENEWAL',
      title: 'Kadıköy Konut Kira Sözleşmesi TÜFE Artışı',
      description: 'Yıllık kira artış oranı bildirimi ve TÜFE tavan oranı denetimi.',
      dueDate: new Date(Date.now() + 18 * 86400000).toISOString(),
      amount: { amount: 32000, currency: 'TRY' },
      status: 'PENDING_APPROVAL',
      riskLevel: 'HIGH',
      suggestedAction: {
        type: 'create_event',
        description: 'Ev sahibi ile TÜFE oranında yenileme görüşmesi takvime ekle',
        mcpAdapter: 'calendar',
        suggestedAt: new Date().toISOString(),
      },
      autoApprove: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    },
    {
      id: 'b2c-003',
      documentId: 'doc-p-03',
      userId: '00000000-0000-0000-0000-000000000001',
      type: 'PAYMENT',
      title: 'Turkcell Superonline 1000 Mbps Fiber Taahhüt Sonu',
      description: '24 aylık kampanya sonu; taahhütsüz tarifeye geçmeden yenileme yapılmalı.',
      dueDate: new Date(Date.now() + 28 * 86400000).toISOString(),
      amount: { amount: 490, currency: 'TRY' },
      status: 'IN_PROGRESS',
      riskLevel: 'MEDIUM',
      suggestedAction: {
        type: 'send_slack',
        description: 'Alternatif Türk Telekom ve Vodafone tekliflerini listele',
        mcpAdapter: 'slack',
        suggestedAt: new Date().toISOString(),
      },
      autoApprove: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    },
  ]

  const activeList = isB2B ? obligations : personalObligations

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
            style={{
              width: '100%',
              padding: '9px 12px 9px 36px',
              backgroundColor: 'var(--bg-surface)',
              border: '1px solid var(--border-subtle)',
              borderRadius: 'var(--radius-sm)',
              color: 'var(--text-primary)',
              fontSize: 13,
              outline: 'none',
              fontFamily: 'inherit',
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
                  <tr key={o.id}>
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
                      <span style={{ fontSize: 12, color: o.status === 'RESOLVED' ? '#10b981' : 'var(--text-secondary)' }}>
                        {o.status}
                      </span>
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
                      <div style={{ display: 'inline-flex', gap: 6 }}>
                        {o.status === 'PENDING_APPROVAL' && (
                          <>
                            <button
                              className="btn btn-primary"
                              style={{ fontSize: 12, padding: '4px 8px' }}
                              onClick={() => onApprove(o.id)}
                            >
                              <CheckCircle size={13} /> {t.approve}
                            </button>
                            <button
                              className="btn btn-secondary"
                              style={{ fontSize: 12, padding: '4px 8px' }}
                              onClick={() => onDismiss(o.id, 'Dismissed')}
                            >
                              <XCircle size={13} />
                            </button>
                          </>
                        )}
                        {o.status === 'RESOLVED' && (
                          <span style={{ color: '#10b981', fontSize: 12, display: 'inline-flex', alignItems: 'center', gap: 4 }}>
                            <CheckCircle size={13} /> {t.approved}
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
