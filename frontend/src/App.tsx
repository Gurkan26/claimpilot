import React, { useEffect, useState } from 'react'
import { AuthProvider, useAuth } from './context/AuthContext'
import { TitleBar } from './components/TitleBar'
import { Sidebar, TabType } from './components/Sidebar'
import { DashboardView } from './components/DashboardView'
import { ObligationsView } from './components/ObligationsView'
import { MarketplaceView } from './components/MarketplaceView'
import { DocumentsView } from './components/DocumentsView'
import { AuditLogView } from './components/AuditLogView'
import { AuthModal } from './components/AuthModal'
import { api } from './services/api'
import {
  DashboardSummary,
  Obligation,
  MarketplaceOpportunity,
  MarketplaceMetrics,
  DocumentItem,
  AuditEntry,
} from './types'

const AppContent: React.FC = () => {
  const [activeTab, setActiveTab] = useState<TabType>('dashboard')
  const [isBackendOnline, setIsBackendOnline] = useState(false)
  const { user, t } = useAuth()

  // State
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [obligations, setObligations] = useState<Obligation[]>([])
  const [opportunities, setOpportunities] = useState<MarketplaceOpportunity[]>([])
  const [metrics, setMetrics] = useState<MarketplaceMetrics | null>(null)
  const [documents, setDocuments] = useState<DocumentItem[]>([])
  const [auditLogs, setAuditLogs] = useState<AuditEntry[]>([])

  // Load data
  const loadData = async () => {
    const isOnline = await api.checkHealth()
    setIsBackendOnline(isOnline)

    try {
      const [sumData, oblData, oppData, metData, docData, logData] = await Promise.all([
        api.getDashboardSummary(),
        api.listObligations(30),
        api.listOpportunities(),
        api.getMarketplaceMetrics(),
        api.listDocuments(),
        api.listAuditLogs(),
      ])

      setSummary(sumData)
      setObligations(oblData)
      setOpportunities(oppData)
      setMetrics(metData)
      setDocuments(docData)
      setAuditLogs(logData)
    } catch (e) {
      console.error('Failed to load application data:', e)
    }
  }

  useEffect(() => {
    loadData()
    const interval = setInterval(loadData, 20000)
    return () => clearInterval(interval)
  }, [])

  // Action handlers
  const handleApprove = async (id: string) => {
    try {
      await api.approveObligation(id, true)
      if (window.claimpilotDesktop) {
        window.claimpilotDesktop.notify('ClaimPilot', 'Yükümlülük onaylandı ve MCP eylemi başarıyla çalıştırıldı.')
      }
      loadData()
    } catch (e: any) {
      alert(`Approval error: ${e.message}`)
    }
  }

  const handleDismiss = async (id: string, reason: string) => {
    try {
      await api.dismissObligation(id, reason)
      loadData()
    } catch (e: any) {
      alert(`Dismiss error: ${e.message}`)
    }
  }

  const handleTriggerRFQ = async (oppId: string) => {
    try {
      await api.triggerRFQ(oppId)
      if (window.claimpilotDesktop) {
        window.claimpilotDesktop.notify('ClaimPilot RFQ', 'Tedarikçi teklifleri toplandı ve Verifier risk skoru hesaplandı.')
      }
      loadData()
    } catch (e: any) {
      alert(`RFQ error: ${e.message}`)
    }
  }

  const handleAcceptBid = async (oppId: string, bidId: string) => {
    try {
      const res = await api.acceptBid(oppId, bidId)
      if (window.claimpilotDesktop) {
        window.claimpilotDesktop.notify('Anlaşma Bağlandı!', res.message || 'Yeni tedarikçiye geçiş tamamlandı.')
      }
      loadData()
    } catch (e: any) {
      alert(`Accept bid error: ${e.message}`)
    }
  }

  const handleUploadFile = async (file: File) => {
    try {
      await api.uploadDocument(file)
      if (window.claimpilotDesktop) {
        window.claimpilotDesktop.notify('Belge Analiz Edildi', `${file.name} başarıyla ayrıştırıldı.`)
      }
      loadData()
    } catch (e: any) {
      alert(`Upload error: ${e.message}`)
    }
  }

  const pendingCount = obligations.filter((o) => o.status === 'PENDING_APPROVAL').length
  const dealsCount = opportunities.filter((o) => o.status === 'MATCHED').length

  return (
    <div style={{ height: '100%', width: '100%', display: 'flex', flexDirection: 'column' }}>
      <TitleBar isBackendOnline={isBackendOnline} />

      <div className="app-shell">
        <Sidebar
          activeTab={activeTab}
          onTabChange={setActiveTab}
          pendingCount={pendingCount}
          dealsCount={dealsCount}
        />

        <main className="app-content">
          {activeTab === 'dashboard' && (
            <DashboardView
              summary={summary}
              onApprove={handleApprove}
              onNavigateTab={setActiveTab}
            />
          )}

          {activeTab === 'obligations' && (
            <ObligationsView
              obligations={obligations}
              onApprove={handleApprove}
              onDismiss={handleDismiss}
            />
          )}

          {activeTab === 'marketplace' && (
            <MarketplaceView
              opportunities={opportunities}
              metrics={metrics}
              onTriggerRFQ={handleTriggerRFQ}
              onAcceptBid={handleAcceptBid}
            />
          )}

          {activeTab === 'documents' && (
            <DocumentsView documents={documents} onUploadFile={handleUploadFile} />
          )}

          {activeTab === 'audit' && <AuditLogView logs={auditLogs} />}
        </main>
      </div>

      {/* Auth & Account Switcher Modal */}
      <AuthModal />
    </div>
  )
}

export const App: React.FC = () => {
  return (
    <AuthProvider>
      <AppContent />
    </AuthProvider>
  )
}
