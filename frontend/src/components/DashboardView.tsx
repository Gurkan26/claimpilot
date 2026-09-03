import React from 'react'
import {
  Sparkles,
  AlertTriangle,
  ArrowUpRight,
  TrendingUp,
  Calendar,
  CheckCircle2,
  Building2,
  User,
} from 'lucide-react'
import { DashboardSummary } from '../types'
import { useAuth } from '../context/AuthContext'

interface DashboardViewProps {
  summary: DashboardSummary | null
  onApprove: (id: string) => void
  onNavigateTab: (tab: any) => void
}

export const DashboardView: React.FC<DashboardViewProps> = ({
  summary,
  onApprove,
  onNavigateTab,
}) => {
  const { user, t } = useAuth()

  if (!summary) {
    return <div className="page-header">Loading Dashboard...</div>
  }

  const isB2B = user.accountType === 'b2b'

  // Dynamic briefing tailored to account type
  const briefingText = isB2B
    ? summary.aiBriefing
    : 'ClaimPilot kişisel portföyünüzde 1 kritik kasko yenilemesi (Anadolu Sigorta - 3 gün kaldı) tespit etti. Pazar yerinde %35 daha uygun alternatif teklifler toplandı. Ev kirası ve Turkcell Fiber sözleşmeniz takvimde güncel.'

  return (
    <div>
      <div className="page-header">
        <div className="page-title-group">
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <h1>{t.greetingMorning}, {user.name}</h1>
            <span
              className="badge"
              style={{
                backgroundColor: isB2B ? 'rgba(56, 189, 248, 0.15)' : 'rgba(16, 185, 129, 0.15)',
                color: isB2B ? '#38bdf8' : '#34d399',
                borderColor: isB2B ? 'rgba(56, 189, 248, 0.3)' : 'rgba(16, 185, 129, 0.3)',
                display: 'inline-flex',
                alignItems: 'center',
                gap: 4,
              }}
            >
              {isB2B ? <Building2 size={11} /> : <User size={11} />}
              {user.organization}
            </span>
          </div>
          <p className="page-subtitle">{t.dashboardSubtitle}</p>
        </div>
        <button className="btn btn-primary" onClick={() => onNavigateTab('documents')}>
          {t.uploadDocButton}
        </button>
      </div>

      {/* AI Briefing Banner */}
      <div className="briefing-box">
        <div className="modal-icon-badge" style={{ width: 36, height: 36 }}>
          <Sparkles size={20} color="#38bdf8" />
        </div>
        <div className="briefing-content">
          <h3>{t.aiBriefingTitle}</h3>
          <p>{briefingText}</p>
        </div>
      </div>

      {/* Top 4 Stat Cards */}
      <div className="card-grid-4">
        <div className="card">
          <div className="card-label">{t.totalObligations}</div>
          <div className="card-value">{summary.totalObligations}</div>
          <div className="card-subtext">{summary.upcoming30Days} {t.dueNext30Days}</div>
        </div>

        <div
          className="card"
          style={{ borderColor: summary.pendingApprovals > 0 ? 'rgba(245, 158, 11, 0.4)' : undefined }}
        >
          <div className="card-label">{t.pendingApprovals}</div>
          <div className="card-value" style={{ color: summary.pendingApprovals > 0 ? '#fbbf24' : undefined }}>
            {summary.pendingApprovals}
          </div>
          <div className="card-subtext">{t.requiresReview}</div>
        </div>

        <div
          className="card"
          style={{ borderColor: summary.overdueCount > 0 ? 'rgba(239, 68, 68, 0.4)' : undefined }}
        >
          <div className="card-label">{t.overdueItems}</div>
          <div className="card-value" style={{ color: summary.overdueCount > 0 ? '#ef4444' : undefined }}>
            {summary.overdueCount}
          </div>
          <div className="card-subtext">{t.urgentAction}</div>
        </div>

        <div className="card">
          <div className="card-label">{t.activeRfqs}</div>
          <div className="card-value">{summary.financialSummary.activeOpportunities}</div>
          <div className="card-subtext">{t.quotesReady}</div>
        </div>
      </div>

      {/* Financial & Risk Distribution Cards */}
      <div className="card-grid-2">
        <div className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
            <span className="card-label" style={{ margin: 0 }}>{t.financialImpact}</span>
            <TrendingUp size={16} color="#10b981" />
          </div>
          <div style={{ display: 'flex', gap: 32 }}>
            <div>
              <span className="card-label">{t.transactedGmv}</span>
              <div className="card-value mono" style={{ fontSize: 22 }}>
                ${summary.financialSummary.totalGMV.amount.toLocaleString()}
              </div>
            </div>
            <div>
              <span className="card-label">{t.realizedSavings}</span>
              <div className="card-value mono" style={{ fontSize: 22, color: '#10b981' }}>
                +${summary.financialSummary.totalSavings.amount.toLocaleString()}
              </div>
            </div>
          </div>
          <div className="card-subtext" style={{ marginTop: 14 }}>
            {t.commissionFee}
          </div>
        </div>

        <div className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
            <span className="card-label" style={{ margin: 0 }}>{t.riskDistribution}</span>
            <AlertTriangle size={16} color="#f59e0b" />
          </div>
          <div style={{ display: 'flex', gap: 12 }}>
            <div style={{ flex: 1, padding: 10, background: 'var(--bg-surface-elevated)', borderRadius: 6, border: '1px solid var(--border-subtle)' }}>
              <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>Critical</div>
              <div style={{ fontSize: 18, fontWeight: 700, color: '#ef4444' }}>{summary.riskBreakdown.critical}</div>
            </div>
            <div style={{ flex: 1, padding: 10, background: 'var(--bg-surface-elevated)', borderRadius: 6, border: '1px solid var(--border-subtle)' }}>
              <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>High</div>
              <div style={{ fontSize: 18, fontWeight: 700, color: '#f59e0b' }}>{summary.riskBreakdown.high}</div>
            </div>
            <div style={{ flex: 1, padding: 10, background: 'var(--bg-surface-elevated)', borderRadius: 6, border: '1px solid var(--border-subtle)' }}>
              <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>Medium</div>
              <div style={{ fontSize: 18, fontWeight: 700, color: '#38bdf8' }}>{summary.riskBreakdown.medium}</div>
            </div>
            <div style={{ flex: 1, padding: 10, background: 'var(--bg-surface-elevated)', borderRadius: 6, border: '1px solid var(--border-subtle)' }}>
              <div style={{ fontSize: 11, color: 'var(--text-muted)' }}>Low</div>
              <div style={{ fontSize: 18, fontWeight: 700, color: '#10b981' }}>{summary.riskBreakdown.low}</div>
            </div>
          </div>
          <div className="card-subtext" style={{ marginTop: 14 }}>
            High & Critical items trigger automatic MCP actions
          </div>
        </div>
      </div>

      {/* Critical Deadlines Table */}
      <div className="page-header" style={{ marginBottom: 12 }}>
        <h2 style={{ fontSize: 16, fontWeight: 700 }}>{t.criticalDeadlines}</h2>
        <button
          className="btn btn-secondary"
          style={{ fontSize: 12, padding: '4px 10px' }}
          onClick={() => onNavigateTab('obligations')}
        >
          {t.viewAll} ({summary.totalObligations}) <ArrowUpRight size={13} />
        </button>
      </div>

      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>{t.obligationCol}</th>
              <th>{t.categoryCol}</th>
              <th>{t.dueDateCol}</th>
              <th>{t.riskCol}</th>
              <th>{t.actionCol}</th>
              <th style={{ textAlign: 'right' }}>{t.actionsCol}</th>
            </tr>
          </thead>
          <tbody>
            {summary.criticalDeadlines.map((item) => (
              <tr key={item.id}>
                <td style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{item.title}</td>
                <td>{item.type}</td>
                <td>
                  <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}>
                    <Calendar size={13} color="var(--text-muted)" />
                    {new Date(item.dueDate).toLocaleDateString()}
                    <span style={{ color: item.daysRemaining <= 3 ? '#ef4444' : '#f59e0b', fontWeight: 600 }}>
                      ({item.daysRemaining} {t.daysLeft})
                    </span>
                  </span>
                </td>
                <td>
                  <span className={`badge badge-${item.riskLevel.toLowerCase()}`}>
                    {item.riskLevel}
                  </span>
                </td>
                <td style={{ color: 'var(--text-secondary)' }}>{item.suggestedAction}</td>
                <td style={{ textAlign: 'right' }}>
                  <button
                    className="btn btn-primary"
                    style={{ fontSize: 12, padding: '5px 10px' }}
                    onClick={() => onApprove(item.id)}
                  >
                    <CheckCircle2 size={13} /> {t.approve}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
