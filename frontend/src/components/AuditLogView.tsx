import React from 'react'
import { ShieldCheck, CheckCircle, Bot, UserCheck } from 'lucide-react'
import { AuditEntry } from '../types'
import { useAuth } from '../context/AuthContext'

interface AuditLogViewProps {
  logs: AuditEntry[]
}

export const AuditLogView: React.FC<AuditLogViewProps> = ({ logs }) => {
  const { t } = useAuth()

  return (
    <div>
      <div className="page-header">
        <div className="page-title-group">
          <h1>{t.auditTitle}</h1>
          <p className="page-subtitle">{t.auditSubtitle}</p>
        </div>
      </div>

      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>{t.eventCol}</th>
              <th>{t.approvalTypeCol}</th>
              <th>{t.toolCol}</th>
              <th>{t.statusCol}</th>
              <th>{t.detailsCol}</th>
              <th>{t.timeCol}</th>
            </tr>
          </thead>
          <tbody>
            {logs.map((log) => (
              <tr key={log.id}>
                <td>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontWeight: 600, color: 'var(--text-primary)' }}>
                    <ShieldCheck size={14} color="#38bdf8" />
                    {log.actionType}
                  </div>
                </td>
                <td>
                  <span
                    style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: 4,
                      fontSize: 11,
                      padding: '2px 6px',
                      borderRadius: 4,
                      backgroundColor: log.approvalType === 'AUTONOMOUS' ? 'rgba(56, 189, 248, 0.12)' : 'rgba(245, 158, 11, 0.12)',
                      color: log.approvalType === 'AUTONOMOUS' ? '#38bdf8' : '#fbbf24',
                    }}
                  >
                    {log.approvalType === 'AUTONOMOUS' ? <Bot size={11} /> : <UserCheck size={11} />}
                    {log.approvalType}
                  </span>
                </td>
                <td>
                  <span className="badge badge-medium" style={{ textTransform: 'capitalize' }}>
                    {log.adapterId || 'system'}
                  </span>
                </td>
                <td>
                  <span className="badge badge-success" style={{ gap: 4 }}>
                    <CheckCircle size={10} /> {log.status}
                  </span>
                </td>
                <td style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--text-secondary)' }}>
                  {log.details ? JSON.stringify(log.details) : '—'}
                </td>
                <td>{new Date(log.createdAt).toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
