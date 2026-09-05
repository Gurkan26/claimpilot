import React, { useEffect, useState, useCallback } from 'react'
import { AuthProvider, useAuth } from './context/AuthContext'
import { TitleBar } from './components/TitleBar'
import { Sidebar, TabType } from './components/Sidebar'
import { DashboardView } from './components/DashboardView'
import { ObligationsView } from './components/ObligationsView'
import { MarketplaceView } from './components/MarketplaceView'
import { DocumentsView } from './components/DocumentsView'
import { AuditLogView } from './components/AuditLogView'
import { AdminView } from './components/AdminView'
import { LoginScreen } from './components/LoginScreen'
import { ToastContainer, useToast } from './components/Toast'
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
  const { user, isAuthenticated, isAdmin, t } = useAuth()
  const [activeTab, setActiveTab] = useState<TabType>(() => (isAdmin ? 'admin' : 'dashboard'))
  const [isBackendOnline, setIsBackendOnline] = useState(false)
  const { toasts, addToast, removeToast } = useToast()

  // Auto-switch to admin tab if user logs in as admin; redirect away if not admin
  useEffect(() => {
    if (isAdmin) {
      setActiveTab('admin')
    } else if (activeTab === 'admin') {
      setActiveTab('dashboard')
    }
  }, [isAdmin, activeTab])

  // State
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [obligations, setObligations] = useState<Obligation[]>([])
  const [opportunities, setOpportunities] = useState<MarketplaceOpportunity[]>([])
  const [metrics, setMetrics] = useState<MarketplaceMetrics | null>(null)
  const [documents, setDocuments] = useState<DocumentItem[]>([])
  const [auditLogs, setAuditLogs] = useState<AuditEntry[]>([])

  // Load data
  const loadData = useCallback(async () => {
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
  }, [])

  useEffect(() => {
    if (!isAuthenticated) return
    loadData()
    const interval = setInterval(loadData, 20000)
    return () => clearInterval(interval)
  }, [isAuthenticated, loadData])

  // Action handlers with toast notifications
  const handleApprove = async (id: string) => {
    try {
      await api.approveObligation(id, true)
      addToast('success', t.toastApproved)
      if (window.claimpilotDesktop) {
        window.claimpilotDesktop.notify('ClaimPilot', t.toastApproved)
      }
      loadData()
    } catch (e: any) {
      addToast('error', `${t.toastError} ${e.message}`)
    }
  }

  const handleDismiss = async (id: string, reason: string) => {
    try {
      await api.dismissObligation(id, reason)
      addToast('info', t.toastDismissed)
      loadData()
    } catch (e: any) {
      addToast('error', `${t.toastError} ${e.message}`)
    }
  }

  const handleTriggerRFQ = async (oppId: string) => {
    try {
      await api.triggerRFQ(oppId)
      addToast('success', t.toastRfqCollected)
      if (window.claimpilotDesktop) {
        window.claimpilotDesktop.notify('ClaimPilot RFQ', t.toastRfqCollected)
      }
      loadData()
    } catch (e: any) {
      addToast('error', `${t.toastError} ${e.message}`)
    }
  }

  const handleAcceptBid = async (oppId: string, bidId: string) => {
    try {
      await api.acceptBid(oppId, bidId)
      addToast('success', t.toastDealClosed)
      if (window.claimpilotDesktop) {
        window.claimpilotDesktop.notify('Anlaşma Bağlandı!', t.toastDealClosed)
      }
      loadData()
    } catch (e: any) {
      addToast('error', `${t.toastError} ${e.message}`)
    }
  }

  const handleUploadFile = async (file: File) => {
    try {
      await api.uploadDocument(file)
      addToast('success', `${t.toastUploadSuccess} (${file.name})`)
      if (window.claimpilotDesktop) {
        window.claimpilotDesktop.notify('Belge Analiz Edildi', `${file.name} başarıyla ayrıştırıldı.`)
      }
      loadData()
    } catch (e: any) {
      addToast('error', `${t.toastError} ${e.message}`)
    }
  }

  // If not authenticated, show login
  if (!isAuthenticated) {
    return (
      <>
        <LoginScreen />
        <ToastContainer toasts={toasts} onRemove={removeToast} />
      </>
    )
  }

  const pendingCount = obligations.filter((o) => o.status === 'PENDING_APPROVAL').length
  const dealsCount = opportunities.filter((o) => o.status === 'MATCHED').length

  return (
    <div className="app-root-wrapper">
      <div style={{ height: '100%', width: '100%', display: 'flex', flexDirection: 'column' }}>
        <TitleBar isBackendOnline={isBackendOnline} onNavigateTab={setActiveTab} />

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

            {activeTab === 'admin' && isAdmin && (
              <AdminView onShowToast={(msg, type) => addToast(type, msg)} />
            )}
          </main>
        </div>
      </div>

      <ToastContainer toasts={toasts} onRemove={removeToast} />
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
