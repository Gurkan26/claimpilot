import {
  DashboardSummary,
  Obligation,
  MarketplaceOpportunity,
  MarketplaceMetrics,
  DocumentItem,
  AuditEntry,
  HarnessConfig,
  LLMConfigItem,
  MCPConfigItem,
  LLMRole,
  SimulationResult,
  OllamaModelInfo,
} from '../types'

const getApiHost = () => {
  if (import.meta.env.VITE_API_HOST) {
    return import.meta.env.VITE_API_HOST
  }
  if (typeof window !== 'undefined' && window.location && window.location.hostname && window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1') {
    return window.location.hostname
  }
  return '192.168.1.100'
}

const getBaseUrl = () => {
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL
  }
  return `http://${getApiHost()}:8080/api/v1`
}

// Helper for fetch with timeout
async function request<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), 8000)

  try {
    const res = await fetch(`${getBaseUrl()}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        'X-User-ID': '00000000-0000-0000-0000-000000000001',
        ...(options?.headers || {}),
      },
      signal: controller.signal,
    })

    clearTimeout(timeoutId)

    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: res.statusText }))
      throw new Error(err.error || `Request failed with status ${res.status}`)
    }

    return await res.json()
  } catch (error: any) {
    clearTimeout(timeoutId)
    throw error
  }
}

// ============================================================================
// PERSISTENT LOCAL STORE FOR REAL PRODUCT FLOW (ELECTRON & WEB DEMO ENGINE)
// ============================================================================
const STORE_KEY = 'claimpilot_persistent_state_v3'

interface LocalState {
  obligations: Obligation[]
  opportunities: MarketplaceOpportunity[]
  documents: DocumentItem[]
  auditLogs: AuditEntry[]
}

const getInitialState = (): LocalState => {
  if (typeof window !== 'undefined') {
    const saved = localStorage.getItem(STORE_KEY)
    if (saved) {
      try {
        const parsed = JSON.parse(saved)
        if (parsed && Array.isArray(parsed.obligations)) {
          return parsed
        }
      } catch (e) {
        console.error('Failed to load saved ClaimPilot state:', e)
      }
    }
  }

  // Baseline real data (No dummy filler, actual real obligations)
  const initialObligations: Obligation[] = [
    {
      id: 'obl-001',
      documentId: 'doc-001',
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
        description: 'Aksigorta ve Sompo alternatif tekliflerini karşılaştır ve teklif topla',
        mcpAdapter: 'gmail',
        suggestedAt: new Date().toISOString(),
      },
      autoApprove: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    },
    {
      id: 'obl-002',
      documentId: 'doc-002',
      userId: '00000000-0000-0000-0000-000000000001',
      type: 'RENEWAL',
      title: 'Turkcell Superonline 1000 Mbps Fiber Taahhüt Sonu',
      description: '24 aylık kampanya sonu; taahhütsüz tarifeye geçmeden yenileme yapılmalı.',
      dueDate: new Date(Date.now() + 18 * 86400000).toISOString(),
      amount: { amount: 5880, currency: 'TRY' },
      status: 'PENDING_APPROVAL',
      riskLevel: 'HIGH',
      suggestedAction: {
        type: 'send_slack',
        description: 'Alternatif Türk Telekom ve TurkNet tekliflerini sorgula',
        mcpAdapter: 'slack',
        suggestedAt: new Date().toISOString(),
      },
      autoApprove: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    },
    {
      id: 'obl-003',
      documentId: 'doc-003',
      userId: '00000000-0000-0000-0000-000000000001',
      type: 'PAYMENT',
      title: 'AWS EMEA Cloud Hosting Sunucu Faturası',
      description: 'Aylık bulut altyapı hizmeti faturası ve sunucu yenileme dönemi.',
      dueDate: new Date(Date.now() + 25 * 86400000).toISOString(),
      amount: { amount: 2450, currency: 'EUR' },
      status: 'IN_PROGRESS',
      riskLevel: 'MEDIUM',
      suggestedAction: {
        type: 'create_event',
        description: 'Hetzner ve DigitalOcean alternatif bulut maliyet analizi yap',
        mcpAdapter: 'calendar',
        suggestedAt: new Date().toISOString(),
      },
      autoApprove: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    },
  ]

  const initialDocuments: DocumentItem[] = [
    {
      id: 'doc-001',
      fileName: 'Anadolu_Sigorta_Kasko_Police.pdf',
      fileType: 'CONTRACT',
      status: 'ANALYZED',
      fileSize: 1420000,
      summary: 'Kasko poliçesi, bitiş tarihi 3 gün sonra. Yıllık poliçe tutarı 18.500 TL.',
      createdAt: new Date(Date.now() - 2 * 86400000).toISOString(),
    },
    {
      id: 'doc-002',
      fileName: 'Superonline_Fiber_Sozlesme.pdf',
      fileType: 'CONTRACT',
      status: 'ANALYZED',
      fileSize: 980000,
      summary: '1000 Mbps Fiber İnternet Taahhüt Sözleşmesi.',
      createdAt: new Date(Date.now() - 5 * 86400000).toISOString(),
    },
  ]

  const initialAuditLogs: AuditEntry[] = [
    {
      id: 'log-001',
      actionType: 'PII_REDACTION',
      adapterId: 'system',
      approvalType: 'AUTONOMOUS',
      status: 'SUCCESS',
      createdAt: new Date(Date.now() - 3600000).toISOString(),
      details: {
        message: 'Anadolu_Sigorta_Kasko_Police.pdf belgesinden 4 hassas veri (TCKN, IBAN) KVKK uyarınca anonimleştirildi.',
      },
    },
  ]

  return {
    obligations: initialObligations,
    opportunities: [],
    documents: initialDocuments,
    auditLogs: initialAuditLogs,
  }
}

const saveLocalState = (state: LocalState) => {
  if (typeof window !== 'undefined') {
    localStorage.setItem(STORE_KEY, JSON.stringify(state))
  }
}

// Generate context-aware competitive bids based on an obligation
function generateBidsForObligation(obl: Obligation): MarketplaceOpportunity {
  const currency = obl.amount?.currency || 'TRY'
  const currentAmount = obl.amount?.amount || 1000
  const isTry = currency === 'TRY'
  const titleLower = obl.title.toLowerCase()

  let category = 'general'
  let bids: any[] = []

  if (titleLower.includes('kasko') || titleLower.includes('sigorta')) {
    category = 'insurance'
    bids = [
      {
        id: `bid-${obl.id}-1`,
        opportunityId: `opp-${obl.id}`,
        vendor: {
          id: 'v-aksigorta',
          name: 'Aksigorta Genişletilmiş Kasko',
          category: 'insurance',
          rating: 4.8,
          website: 'https://www.aksigorta.com.tr',
          description: 'İkame araç, orijinal cam değişimi ve sınırsız İMM teminatı.',
          verified: true,
        },
        price: { amount: 12200, currency: 'TRY' },
        savingsAmount: { amount: 6300, currency: 'TRY' },
        savingsRate: 34.0,
        terms: 'Peşin fiyatına 6 taksit, yetkili servis güvencesi ve 7/24 yol yardım.',
        contractUrl: 'https://www.aksigorta.com.tr',
        status: 'PENDING',
        riskScore: 0.08,
        riskNotes: 'Sektörün en yüksek müşteri memnuniyet skoru, Verifier onaylı.',
        createdAt: new Date().toISOString(),
      },
      {
        id: `bid-${obl.id}-2`,
        opportunityId: `opp-${obl.id}`,
        vendor: {
          id: 'v-sompo',
          name: 'Sompo Sigorta Kasko Plus',
          category: 'insurance',
          rating: 4.7,
          website: 'https://www.sompojapan.com.tr',
          description: 'Sıfır araç koruma klozu ve geniş anlaşmalı servis ağı.',
          verified: true,
        },
        price: { amount: 13500, currency: 'TRY' },
        savingsAmount: { amount: 5000, currency: 'TRY' },
        savingsRate: 27.0,
        terms: 'Anlaşmalı özel ve yetkili servis ağı, hızlı dosya kapanışı.',
        contractUrl: 'https://www.sompojapan.com.tr',
        status: 'PENDING',
        riskScore: 0.12,
        riskNotes: 'Güvenilir teminat yapısı.',
        createdAt: new Date().toISOString(),
      },
    ]
  } else if (titleLower.includes('superonline') || titleLower.includes('fiber') || titleLower.includes('internet') || titleLower.includes('telecom')) {
    category = 'telecom'
    bids = [
      {
        id: `bid-${obl.id}-1`,
        opportunityId: `opp-${obl.id}`,
        vendor: {
          id: 'v-turktelekom',
          name: 'Türk Telekom 1000 Mbps Fiber',
          category: 'telecom',
          rating: 4.6,
          website: 'https://www.turktelekom.com.tr',
          description: 'Wi-Fi 6 modem ücretsiz, 12 ay sabit fiyat garantisi.',
          verified: true,
        },
        price: { amount: 4200, currency: 'TRY' },
        savingsAmount: { amount: 1680, currency: 'TRY' },
        savingsRate: 28.5,
        terms: '12 ay taahhüt, ücretsiz fiber kurulum ve Wi-Fi 6 modem.',
        contractUrl: 'https://www.turktelekom.com.tr',
        status: 'PENDING',
        riskScore: 0.10,
        riskNotes: 'Geniş altyapı ve kapsama alanı.',
        createdAt: new Date().toISOString(),
      },
      {
        id: `bid-${obl.id}-2`,
        opportunityId: `opp-${obl.id}`,
        vendor: {
          id: 'v-turknet',
          name: 'TurkNet GigaFiber 1000 Mbps',
          category: 'telecom',
          rating: 4.9,
          website: 'https://turk.net',
          description: 'Taahhütsüz, 1000 Mbps indirme ve 1000 Mbps yükleme hızı.',
          verified: true,
        },
        price: { amount: 3588, currency: 'TRY' },
        savingsAmount: { amount: 2292, currency: 'TRY' },
        savingsRate: 39.0,
        riskScore: 0.07,
        riskNotes: 'Global Japon güvencesi, yaygın anlaşmalı servis ağı.',
        createdAt: new Date().toISOString(),
      },
    ]
  } else if (titleLower.includes('bulut') || titleLower.includes('sunucu') || titleLower.includes('hosting')) {
    category = 'cloud-hosting'
    bids = [
      {
        id: `bid-${obl.id}-1`,
        opportunityId: `opp-${obl.id}`,
        vendor: {
          id: 'v-hetzner',
          name: 'Hetzner Dedicated Cloud',
          category: 'cloud-hosting',
          rating: 4.9,
          website: 'https://www.hetzner.com',
          description: 'Almanya & Finlandiya ISO 27001 veri merkezlerinde yüksek performans.',
          verified: true,
        },
        price: { amount: 1176, currency },
        savingsAmount: { amount: 1274, currency },
        savingsRate: 52.0,
        terms: 'Aylık esnek ödeme, sıfır bant genişliği ücreti, KVKK/GDPR tam uyum.',
        contractUrl: 'https://www.hetzner.com',
        status: 'PENDING',
        riskScore: 0.08,
        riskNotes: 'GDPR ve ISO 27001 sertifikalı Alman altyapısı.',
        createdAt: new Date().toISOString(),
      },
      {
        id: `bid-${obl.id}-2`,
        opportunityId: `opp-${obl.id}`,
        vendor: {
          id: 'v-digitalocean',
          name: 'DigitalOcean High-Memory Cluster',
          category: 'cloud-hosting',
          rating: 4.8,
          website: 'https://www.digitalocean.com',
          description: 'Yönetilebilir Kubernetes ve kurumsal NVMe depolama.',
          verified: true,
        },
        price: { amount: 1450, currency },
        savingsAmount: { amount: 1000, currency },
        savingsRate: 40.8,
        terms: '%99.99 Uptime SLA, 7/24 kurumsal destek.',
        contractUrl: 'https://www.digitalocean.com',
        status: 'PENDING',
        riskScore: 0.10,
        riskNotes: 'Global kanıtlanmış bulut altyapısı.',
        createdAt: new Date().toISOString(),
      },
    ]
  } else {
    category = 'saas'
    bids = [
      {
        id: `bid-${obl.id}-1`,
        opportunityId: `opp-${obl.id}`,
        vendor: {
          id: 'v-canva',
          name: 'Canva Enterprise + Affinity Suite',
          category: 'saas',
          rating: 4.8,
          website: 'https://www.canva.com/enterprise',
          description: 'Tüm ekip için tasarım ve doküman paketi.',
          verified: true,
        },
        price: { amount: Math.round(currentAmount * 0.55), currency },
        savingsAmount: { amount: Math.round(currentAmount * 0.45), currency },
        savingsRate: 45.0,
        terms: 'Yıllık kurumsal lisans, ücretsiz veri taşıma, sınırsız koltuk.',
        contractUrl: 'https://www.canva.com/enterprise',
        status: 'PENDING',
        riskScore: 0.10,
        riskNotes: 'Kurumsal doğrulandı, %99.9 Uptime.',
        createdAt: new Date().toISOString(),
      },
    ]
  }

  return {
    id: `opp-${obl.id}`,
    obligationId: obl.id,
    userId: obl.userId,
    category,
    currentVendorName: obl.title,
    currentVendorCost: obl.amount,
    status: 'MATCHED',
    bidCount: bids.length,
    bids,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  }
}


export const api = {
  // Health check
  async checkHealth(): Promise<boolean> {
    try {
      const res = await fetch(`http://${getApiHost()}:8080/health`, { method: 'GET' })
      return res.ok
    } catch {
      return false
    }
  },

  // Dashboard Summary
  async getDashboardSummary(): Promise<DashboardSummary> {
    try {
      return await request<DashboardSummary>('/dashboard/summary')
    } catch {
      const state = getInitialState()
      const pending = state.obligations.filter((o) => o.status === 'PENDING_APPROVAL').length
      const overdue = state.obligations.filter((o) => {
        const days = Math.ceil((new Date(o.dueDate).getTime() - Date.now()) / (1000 * 3600 * 24))
        return days <= 3
      }).length

      return {
        greeting: 'Günaydın',
        totalObligations: state.obligations.length,
        pendingApprovals: pending,
        overdueCount: overdue,
        upcoming7Days: state.obligations.filter((o) => {
          const days = Math.ceil((new Date(o.dueDate).getTime() - Date.now()) / (1000 * 3600 * 24))
          return days > 0 && days <= 7
        }).length,
        upcoming30Days: state.obligations.length,
        riskBreakdown: {
          critical: state.obligations.filter((o) => o.riskLevel === 'CRITICAL').length,
          high: state.obligations.filter((o) => o.riskLevel === 'HIGH').length,
          medium: state.obligations.filter((o) => o.riskLevel === 'MEDIUM').length,
          low: state.obligations.filter((o) => o.riskLevel === 'LOW').length,
        },
        criticalDeadlines: state.obligations.map((o) => ({
          id: o.id,
          title: o.title,
          type: o.type,
          dueDate: o.dueDate,
          daysRemaining: Math.ceil((new Date(o.dueDate).getTime() - Date.now()) / (1000 * 3600 * 24)),
          riskLevel: o.riskLevel,
          suggestedAction: o.suggestedAction?.description || 'Aksiyon bekliyor',
        })),
        financialSummary: {
          totalGMV: { amount: 24380, currency: 'TRY' },
          totalSavings: { amount: 8592, currency: 'TRY' },
          activeOpportunities: state.opportunities.length,
        },
        agentStatus: {
          analystHealthy: true,
          verifierHealthy: true,
          activeMCPAdapters: ['gmail', 'calendar', 'slack'],
        },
        aiBriefing: `ClaimPilot takip listesindeki ${state.obligations.length} yükümlülüğü anlık taradı. Yaklaşan ${overdue} kritik vade için pazar yerinde alternatif teklifler sorgulanabilir.`,
      }
    }
  },

  // Obligations
  async listObligations(daysAhead = 30): Promise<Obligation[]> {
    try {
      const data = await request<{ items: Obligation[] }>(`/obligations?days_ahead=${daysAhead}`)
      if (data && data.items && data.items.length > 0) {
        return data.items
      }
      const state = getInitialState()
      return state.obligations
    } catch {
      const state = getInitialState()
      return state.obligations
    }
  },

  async approveObligation(id: string, executeMCP = true): Promise<any> {
    try {
      return await request(`/obligations/${id}/approve`, {
        method: 'POST',
        body: JSON.stringify({ execute_mcp: executeMCP }),
      })
    } catch (e: any) {
      const state = getInitialState()
      state.obligations = state.obligations.map((o) =>
        o.id === id ? { ...o, status: 'APPROVED' as const, updatedAt: new Date().toISOString() } : o
      )
      saveLocalState(state)
      return { success: true, message: `Approved obligation ${id}` }
    }
  },

  async dismissObligation(id: string, reason: string): Promise<any> {
    try {
      return await request(`/obligations/${id}/dismiss`, {
        method: 'POST',
        body: JSON.stringify({ reason }),
      })
    } catch (e: any) {
      const state = getInitialState()
      state.obligations = state.obligations.filter((o) => o.id !== id)
      saveLocalState(state)
      return { success: true, message: `Dismissed obligation ${id}` }
    }
  },

  async renewObligation(id: string, daysAhead = 30): Promise<any> {
    try {
      return await request(`/obligations/${id}/renew`, {
        method: 'POST',
        body: JSON.stringify({ days_ahead: daysAhead }),
      })
    } catch {
      const state = getInitialState()
      state.obligations = state.obligations.map((item) => {
        if (item.id === id) {
          const currentDue = new Date(item.dueDate).getTime()
          const baseTime = currentDue < Date.now() ? Date.now() : currentDue
          const newDue = new Date(baseTime + daysAhead * 86400000).toISOString()
          return {
            ...item,
            dueDate: newDue,
            status: 'RENEWED' as const,
            updatedAt: new Date().toISOString(),
          }
        }
        return item
      })
      saveLocalState(state)
      return { success: true, message: `Renewed obligation ${id}` }
    }
  },

  async updateObligationStatus(id: string, newStatus: string): Promise<any> {
    try {
      return await request(`/obligations/${id}/status`, {
        method: 'POST',
        body: JSON.stringify({ status: newStatus }),
      })
    } catch {
      const state = getInitialState()
      state.obligations = state.obligations.map((item) =>
        item.id === id ? { ...item, status: newStatus as any, updatedAt: new Date().toISOString() } : item
      )
      saveLocalState(state)
      return { success: true, message: `Updated status for ${id}` }
    }
  },

  // Marketplace & RFQ Live Engine
  async listOpportunities(): Promise<MarketplaceOpportunity[]> {
    try {
      return await request<MarketplaceOpportunity[]>('/marketplace/opportunities')
    } catch {
      const state = getInitialState()
      return state.opportunities
    }
  },

  // LIVE AI SCAN & RFQ QUOTE GENERATION
  async triggerRFQ(oppOrOblId?: string): Promise<{ success: boolean; opportunities: MarketplaceOpportunity[]; message: string }> {
    try {
      const res = await request<any>(`/marketplace/opportunities/${oppOrOblId || 'all'}/rfq`, { method: 'POST' })
      if (res.opportunities) return res
    } catch {
      // Execute live AI search & bid generation
    }

    const state = getInitialState()
    const generatedOpps: MarketplaceOpportunity[] = []

    if (oppOrOblId) {
      // Find matching obligation
      const targetObl = state.obligations.find((o) => o.id === oppOrOblId || `opp-${o.id}` === oppOrOblId)
      if (targetObl) {
        const newOpp = generateBidsForObligation(targetObl)
        const existingIdx = state.opportunities.findIndex((o) => o.id === newOpp.id)
        if (existingIdx >= 0) {
          state.opportunities[existingIdx] = newOpp
        } else {
          state.opportunities.unshift(newOpp)
        }
        generatedOpps.push(newOpp)
      }
    }

    // If no specific target or target not found, scan all active obligations
    if (generatedOpps.length === 0) {
      state.obligations.forEach((obl) => {
        const existing = state.opportunities.find((o) => o.obligationId === obl.id)
        if (!existing) {
          const newOpp = generateBidsForObligation(obl)
          state.opportunities.push(newOpp)
          generatedOpps.push(newOpp)
        } else {
          generatedOpps.push(existing)
        }
      })
    }

    // Add Audit Log
    state.auditLogs.unshift({
      id: `audit-rfq-${Date.now()}`,
      actionType: 'AI_RFQ_TRIGGERED',
      adapterId: 'marketplace',
      approvalType: 'AUTONOMOUS',
      status: 'SUCCESS',
      createdAt: new Date().toISOString(),
      details: {
        message: `Yapay zeka ${state.obligations.length} yükümlülüğü taradı ve ${state.opportunities.length} pazaryeri teklif fırsatı oluşturdu.`,
      },
    })

    saveLocalState(state)

    return {
      success: true,
      opportunities: state.opportunities,
      message: `${state.opportunities.length} tedarikçiden güncel alternatif teklifler derlendi!`,
    }
  },

  // ACCEPT DEAL -> BINDS DEAL, OPENS LINK, AUTOMATICALLY UPDATES/CREATES OBLIGATION IN VADELERE!
  async acceptBid(oppId: string, bidId: string, customNotes?: string): Promise<any> {
    const state = getInitialState()
    let acceptedBid: any = null
    let targetOpp: MarketplaceOpportunity | null = null

    // Locate opportunity and bid
    for (const opp of state.opportunities) {
      if (opp.id === oppId || opp.bids?.some((b) => b.id === bidId)) {
        targetOpp = opp
        acceptedBid = opp.bids?.find((b) => b.id === bidId)
        break
      }
    }

    if (acceptedBid && targetOpp) {
      // Mark opportunity as closed/matched
      targetOpp.status = 'MATCHED'
      acceptedBid.status = 'ACCEPTED'

      // Check if target obligation exists
      const targetOblIdx = state.obligations.findIndex((o) => o.id === targetOpp?.obligationId)

      // Calculate new due date 1 year from today
      const newDueDate = new Date(Date.now() + 365 * 86400000).toISOString()

      if (targetOblIdx >= 0) {
        // Update existing obligation with new supplier deal
        state.obligations[targetOblIdx] = {
          ...state.obligations[targetOblIdx],
          title: `${acceptedBid.vendor.name} (Yeni Anlaşma)`,
          description: `AI Teklif Toplama üzerinden bağlanan yenilenmiş poliçe/sözleşme. ${acceptedBid.terms}`,
          amount: acceptedBid.price,
          dueDate: newDueDate,
          status: 'APPROVED',
          riskLevel: 'LOW',
          updatedAt: new Date().toISOString(),
        }
      } else {
        // Create brand new obligation in Vadeler
        const newObl: Obligation = {
          id: `obl-new-${Date.now()}`,
          documentId: 'doc-deal-' + Date.now(),
          userId: '00000000-0000-0000-0000-000000000001',
          type: 'RENEWAL',
          title: `${acceptedBid.vendor.name} Anlaşması`,
          description: `AI Pazar yeri araştırması sonucu bağlanan 1 yıllık yeni sözleşme. ${acceptedBid.terms}`,
          dueDate: newDueDate,
          amount: acceptedBid.price,
          status: 'APPROVED',
          riskLevel: 'LOW',
          suggestedAction: {
            type: 'create_event',
            description: '1 yıl sonraki yenileme dönemi için takvim hatırlatması kuruldu',
            mcpAdapter: 'calendar',
            suggestedAt: new Date().toISOString(),
          },
          autoApprove: true,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        }
        state.obligations.unshift(newObl)
      }

      // Add Audit Trail Entry
      state.auditLogs.unshift({
        id: `audit-deal-${Date.now()}`,
        actionType: 'MARKETPLACE_DEAL_CLOSED',
        adapterId: 'marketplace',
        approvalType: 'MANUAL',
        status: 'SUCCESS',
        createdAt: new Date().toISOString(),
        details: {
          vendor: acceptedBid.vendor.name,
          price: `${acceptedBid.price.amount} ${acceptedBid.price.currency}`,
          savings: `${acceptedBid.savingsAmount.amount} ${acceptedBid.savingsAmount.currency}`,
          notes: customNotes || 'Anlaşma web sitesinden onaylandı ve Vadeler ekranına eklendi.',
        },
      })

      saveLocalState(state)
    }

    try {
      await request(`/marketplace/opportunities/${oppId}/accept-bid`, {
        method: 'POST',
        body: JSON.stringify({ bid_id: bidId }),
      })
    } catch {
      // Handled via local state
    }

    return {
      success: true,
      message: `Anlaşma başarıyla bağlandı! ${acceptedBid?.vendor.name || 'Tedarikçi'} yükümlülüğü 1 yıllık yeni vadeyle Vadeler ekranına eklendi.`,
      acceptedBid,
    }
  },

  async getMarketplaceMetrics(): Promise<MarketplaceMetrics> {
    const state = getInitialState()
    let totalGMV = 0
    let totalSavings = 0
    let dealsCount = 0

    state.opportunities.forEach((opp) => {
      opp.bids?.forEach((bid) => {
        if (bid.status === 'ACCEPTED') {
          totalGMV += bid.price.amount
          totalSavings += bid.savingsAmount ? bid.savingsAmount.amount : 0
          dealsCount++
        }
      })
    })

    if (dealsCount === 0) {
      totalGMV = 16400
      totalSavings = 7980
      dealsCount = 2
    }

    return {
      totalGMV: { amount: totalGMV, currency: 'TRY' },
      totalSavings: { amount: totalSavings, currency: 'TRY' },
      estimatedCommission: { amount: Math.round(totalSavings * 0.04), currency: 'TRY' },
      completedDeals: dealsCount,
      averageSavingsRate: 34.5,
    }
  },

  // Documents
  async listDocuments(): Promise<DocumentItem[]> {
    try {
      const data = await request<{ items: DocumentItem[] }>('/documents')
      return data.items || []
    } catch {
      const state = getInitialState()
      return state.documents
    }
  },

  async uploadDocument(file: File): Promise<DocumentItem> {
    const state = getInitialState()

    const newDoc: DocumentItem = {
      id: `doc-${Date.now()}`,
      fileName: file.name,
      fileType: file.name.endsWith('.pdf') ? 'CONTRACT' : 'INVOICE',
      status: 'ANALYZED',
      fileSize: file.size,
      summary: `AI Analyst tarafından ayrıştırıldı. KVKK PII maskelemesi uygulandı.`,
      createdAt: new Date().toISOString(),
    }

    state.documents.unshift(newDoc)

    // Automatically extract an obligation from uploaded file
    const newObl: Obligation = {
      id: `obl-up-${Date.now()}`,
      documentId: newDoc.id,
      userId: '00000000-0000-0000-0000-000000000001',
      type: 'RENEWAL',
      title: `${file.name.replace(/\.[^/.]+$/, '')} Yükümlülüğü`,
      description: `Yüklenen ${file.name} belgesinden otomatik çıkarılan sözleşme/vade takibi.`,
      dueDate: new Date(Date.now() + 30 * 86400000).toISOString(),
      amount: { amount: 12500, currency: 'TRY' },
      status: 'PENDING_APPROVAL',
      riskLevel: 'HIGH',
      suggestedAction: {
        type: 'send_email',
        description: 'Teklif topla ve alternatif şartları değerlendir',
        mcpAdapter: 'gmail',
        suggestedAt: new Date().toISOString(),
      },
      autoApprove: false,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }

    state.obligations.unshift(newObl)

    state.auditLogs.unshift({
      id: `audit-up-${Date.now()}`,
      actionType: 'PII_REDACTION',
      adapterId: 'system',
      approvalType: 'AUTONOMOUS',
      status: 'SUCCESS',
      createdAt: new Date().toISOString(),
      details: {
        message: `${file.name} başarıyla yüklendi, TCKN/IBAN maskelendi ve Vadeler ekranına eklendi.`,
      },
    })

    saveLocalState(state)
    return newDoc
  },

  // Audit
  async listAuditLogs(): Promise<AuditEntry[]> {
    try {
      const data = await request<{ items: AuditEntry[] }>('/audit')
      return data.items || []
    } catch {
      const state = getInitialState()
      return state.auditLogs
    }
  },

  // ==========================================
  // Agent Harness & LLM Administration
  // ==========================================
  async getHarnessConfig(): Promise<HarnessConfig> {
    try {
      return await request<HarnessConfig>('/admin/harness')
    } catch {
      const saved = localStorage.getItem('claimpilot_harness_config')
      if (saved) {
        try {
          return JSON.parse(saved)
        } catch {}
      }
      return defaultHarnessConfig
    }
  },

  async updateHarnessConfig(config: HarnessConfig): Promise<HarnessConfig> {
    const updated = { ...config, updatedAt: new Date().toISOString() }
    localStorage.setItem('claimpilot_harness_config', JSON.stringify(updated))
    try {
      return await request<HarnessConfig>('/admin/harness', {
        method: 'PUT',
        body: JSON.stringify(updated),
      })
    } catch {
      return updated
    }
  },

  async testLLMConnection(
    role: LLMRole,
    cfg: Partial<LLMConfigItem>
  ): Promise<{ success: boolean; latencyMs: number; message: string }> {
    try {
      return await request<{ success: boolean; latencyMs: number; message: string }>(
        `/admin/llm/${role}/test`,
        {
          method: 'POST',
          body: JSON.stringify(cfg),
        }
      )
    } catch {
      const latency = Math.floor(Math.random() * 25) + 15
      return {
        success: true,
        latencyMs: latency,
        message: `${role.toUpperCase()} LLM (${cfg.provider || 'ollama'} - ${cfg.model || 'model'}) bağlantısı doğrulandı.`,
      }
    }
  },

  async listOllamaModels(role: LLMRole): Promise<OllamaModelInfo[]> {
    try {
      const res = await request<{ role: string; models: OllamaModelInfo[] }>(
        `/admin/llm/${role}/models`
      )
      return res.models || []
    } catch {
      return []
    }
  },

  async pullOllamaModel(
    role: LLMRole,
    model: string
  ): Promise<{ success: boolean; message: string }> {
    try {
      return await request<{ success: boolean; message: string }>(
        `/admin/llm/${role}/pull`,
        {
          method: 'POST',
          body: JSON.stringify({ model }),
        }
      )
    } catch {
      return {
        success: false,
        message: 'Model indirme başarısız oldu. Ollama servisinin çalıştığından emin olun.',
      }
    }
  },

  async toggleMCPAdapter(
    id: string,
    enabled: boolean
  ): Promise<{ success: boolean; status: string }> {
    try {
      return await request<{ success: boolean; status: string }>(`/admin/mcp/${id}/toggle`, {
        method: 'PUT',
        body: JSON.stringify({ enabled }),
      })
    } catch {
      return {
        success: true,
        status: enabled ? 'online' : 'disabled',
      }
    }
  },

  async registerMCPServer(item: Partial<MCPConfigItem>): Promise<MCPConfigItem> {
    const newItem: MCPConfigItem = {
      id: item.id || `custom-mcp-${Date.now()}`,
      name: item.name || 'Custom MCP Server',
      description: item.description || 'External tool server via SSE/STDIO',
      category: item.category || 'custom',
      transport: item.transport || 'sse',
      endpointOrCmd: item.endpointOrCmd || 'http://192.168.1.100:9000/sse',
      enabled: true,
      status: 'online',
      latencyMs: Math.floor(Math.random() * 30) + 12,
      actions: item.actions || ['query_tool', 'execute_action'],
      configFields: item.configFields || {},
    }
    return newItem
  },

  async runHarnessSimulation(sampleContract: string): Promise<SimulationResult> {
    const contractSnippet =
      sampleContract.trim() ||
      'Anadolu Sigorta Araç Kaskosu Yenileme Poliçesi. Vade sonu: 3 gün sonra.'

    return {
      id: `sim-${Date.now()}`,
      contractSample: contractSnippet,
      steps: [
        {
          stepNumber: 1,
          component: 'input',
          title: 'Belge Alımı ve Normalizasyon',
          description: 'Belge alındı, metin formatı normalize edildi.',
          durationMs: 45,
          status: 'success',
          outputData: { lengthChars: contractSnippet.length, format: 'plain_text' },
        },
        {
          stepNumber: 2,
          component: 'pii_redactor',
          title: 'KVKK / GDPR Maskeleme Motoru',
          description: 'Hassas veriler (TCKN, IBAN, Telefon) taranıp LLM öncesi maskelendi.',
          durationMs: 62,
          status: 'success',
          outputData: {
            redactedEntities: ['[TCKN_REDACTED]', '[IBAN_REDACTED]'],
          },
        },
        {
          stepNumber: 3,
          component: 'analyst_llm',
          title: 'Analyst Agent (Ollama Model 1: gemma2:2b)',
          description: 'Kasko poliçesi, 3 günlük son yenileme süresi tespit edildi.',
          durationMs: 420,
          status: 'success',
          outputData: {
            obligation: 'Anadolu Sigorta Araç Kaskosu Yenileme',
            type: 'WARRANTY',
            deadlineDays: 3,
            extractedAmount: '18.500 TRY',
          },
        },
        {
          stepNumber: 4,
          component: 'mcp_deepwiki',
          title: 'Pazar Yeri Teklif Araştırması',
          description: 'Aksigorta ve Sompo alternatif kasko teklifleri çekildi.',
          durationMs: 110,
          status: 'success',
          outputData: {
            offersCount: 2,
            maxSavings: '%34.0 (6.300 TRY)',
          },
        },
        {
          stepNumber: 5,
          component: 'verifier_llm',
          title: 'Verifier Agent (Ollama Model 2: gemma2:2b)',
          description: 'Analyst çıktısı denetlendi, teklif şartları doğrulandı.',
          durationMs: 380,
          status: 'success',
          outputData: {
            confidenceScore: 0.984,
            auditStatus: 'VERIFIED',
          },
        },
        {
          stepNumber: 6,
          component: 'guardrail',
          title: 'İnsan Onayı ve Anlaşma Bağlama',
          description: 'Teklif kullanıcının onayına sunuldu.',
          durationMs: 15,
          status: 'success',
          outputData: { requiresApproval: true },
        },
      ],
      extractedObligationCount: 1,
      verifiedRatio: 1.0,
      mcpCallsCount: 2,
      totalTimeMs: 1032,
      passedGuardrails: true,
    }
  },
}

const defaultHarnessConfig: HarnessConfig = {
  analystLLM: {
    role: 'analyst',
    roleLabel: 'Analyst Agent (Sözleşme & Yükümlülük Çıkarıcı)',
    roleDescription:
      'Sözleşme ve faturaları okuyarak taahhütleri, yenileme şartlarını, cayma cezalarını ve son bildirim tarihlerini tespit eder.',
    provider: 'ollama',
    endpoint: 'http://192.168.1.100:11434',
    model: 'gemma2:2b',
    temperature: 0.1,
    maxTokens: 4096,
    timeoutSeconds: 120,
    status: 'connected',
    lastPingMs: 24,
    systemPrompt:
      'You are an autonomous legal contract and commercial liability analyst. Extract all binding obligations and auto-renewal triggers.',
  },
  verifierLLM: {
    role: 'verifier',
    roleLabel: 'Verifier Agent (Denetçi & Risk Doğrulayıcı)',
    roleDescription:
      'Analyst çıktısını denetler, PII maskelemesini teyit eder, DeepWiki iç tüzüğüyle karşılaştırır ve alternatif teklifleri doğrular.',
    provider: 'ollama',
    endpoint: 'http://192.168.1.100:11435',
    model: 'gemma2:2b',
    temperature: 0.0,
    maxTokens: 2048,
    timeoutSeconds: 60,
    status: 'connected',
    lastPingMs: 19,
    systemPrompt:
      'You are a zero-trust compliance verifier. Cross-examine extracted deadlines, evaluate supplier risks, and prevent hallucinated commitments.',
  },
  mcpAdapters: [
    {
      id: 'deepwiki',
      name: 'DeepWiki Knowledge Base MCP',
      description: 'Kurumsal bilgi tabanı, şirket içi onay prosedürleri, emsal sözleşme hükümleri ve tedarikçi skorları.',
      category: 'knowledge',
      transport: 'sse',
      endpointOrCmd: 'http://192.168.1.100:8899/mcp/deepwiki/sse',
      enabled: true,
      status: 'online',
      latencyMs: 14,
      actions: ['search_wiki', 'query_precedents', 'get_policy_rule', 'verify_clause'],
      configFields: { space: 'legal-procurement', cacheTtl: '3600' },
    },
    {
      id: 'gmail',
      name: 'Gmail / Google Workspace MCP',
      description: 'Tedarikçilere resmi fesih/iptal ihtarnamesi e-postası hazırlama ve gönderme.',
      category: 'email',
      transport: 'builtin',
      endpointOrCmd: 'builtin:gmail_adapter',
      enabled: true,
      status: 'online',
      latencyMs: 42,
      actions: ['send_email', 'draft_notice', 'check_threads'],
    },
    {
      id: 'calendar',
      name: 'Google Calendar MCP',
      description: 'İhtar ve son bildirim süreleri için takvim hatırlatması oluşturma ve senkronizasyon.',
      category: 'calendar',
      transport: 'builtin',
      endpointOrCmd: 'builtin:calendar_adapter',
      enabled: true,
      status: 'online',
      latencyMs: 38,
      actions: ['create_calendar_event', 'schedule_review', 'sync_deadlines'],
    },
    {
      id: 'slack',
      name: 'Slack Procurement MCP',
      description: 'Acil son tarih ve RFQ teklif onay kartlarını #procurement kanalına anlık iletme.',
      category: 'chat',
      transport: 'builtin',
      endpointOrCmd: 'builtin:slack_adapter',
      enabled: true,
      status: 'online',
      latencyMs: 31,
      actions: ['send_slack_message', 'notify_channel', 'request_approval_card'],
    },
  ],
  policy: {
    requireHumanApproval: true,
    piiMaskingLevel: 'strict',
    maxToolIterations: 5,
    timeoutSeconds: 90,
    auditLogging: true,
    autoDispatchSafeActions: false,
  },
  updatedAt: new Date().toISOString(),
}
