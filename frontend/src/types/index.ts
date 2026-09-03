export interface Money {
  amount: number
  currency: string
}

export interface RiskBreakdown {
  critical: number
  high: number
  medium: number
  low: number
}

export interface CriticalDeadlines {
  id: string
  title: string
  type: string
  dueDate: string
  daysRemaining: number
  riskLevel: 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW'
  suggestedAction: string
}

export interface DashboardSummary {
  greeting: string
  totalObligations: number
  pendingApprovals: number
  overdueCount: number
  upcoming7Days: number
  upcoming30Days: number
  riskBreakdown: RiskBreakdown
  criticalDeadlines: CriticalDeadlines[]
  financialSummary: {
    totalGMV: Money
    totalSavings: Money
    activeOpportunities: number
  }
  agentStatus: {
    analystHealthy: boolean
    verifierHealthy: boolean
    activeMCPAdapters: string[]
  }
  aiBriefing: string
}

export interface ObligationAction {
  type: string
  description: string
  mcpAdapter: string
  suggestedAt: string
}

export interface Obligation {
  id: string
  documentId: string
  userId: string
  type: 'RENEWAL' | 'PAYMENT' | 'REPORT' | 'WARRANTY' | 'CANCELLATION' | 'COMPLIANCE'
  title: string
  description: string
  dueDate: string
  amount?: Money
  status: 'DETECTED' | 'PENDING_APPROVAL' | 'IN_PROGRESS' | 'RESOLVED' | 'DISMISSED' | 'EXPIRED'
  riskLevel: 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW'
  suggestedAction?: ObligationAction
  approvedAction?: ObligationAction
  autoApprove: boolean
  marketplaceOpportunityId?: string
  createdAt: string
  updatedAt: string
}

export interface Vendor {
  id: string
  name: string
  category: string
  rating: number
  website: string
  description: string
  verified: boolean
}

export interface Bid {
  id: string
  opportunityId: string
  vendor: Vendor
  price: Money
  savingsAmount?: Money
  savingsRate: number
  terms: string
  contractUrl: string
  status: 'PENDING' | 'ACCEPTED' | 'REJECTED' | 'EXPIRED'
  riskScore?: number
  riskNotes?: string
  validUntil?: string
  createdAt: string
}

export interface MarketplaceOpportunity {
  id: string
  obligationId: string
  userId: string
  category: string
  currentVendorName: string
  currentVendorCost?: Money
  status: 'OPEN' | 'BIDDING' | 'MATCHED' | 'TRANSACTED' | 'EXPIRED' | 'CANCELLED'
  bidCount: number
  bestBidId?: string
  bids?: Bid[]
  createdAt: string
  updatedAt: string
}

export interface Transaction {
  id: string
  opportunityId: string
  bidId: string
  userId: string
  amount: Money
  commission: Money
  commissionRate: number
  savingsRealized?: Money
  status: string
  vendorName: string
  notes: string
  createdAt: string
}

export interface MarketplaceMetrics {
  totalGMV: Money
  totalSavings: Money
  estimatedCommission: Money
  completedDeals: number
  averageSavingsRate: number
}

export interface DocumentItem {
  id: string
  fileName: string
  fileType: string
  status: 'UPLOADED' | 'PROCESSING' | 'ANALYZED' | 'FAILED'
  fileSize: number
  summary?: string
  createdAt: string
}

export interface AuditEntry {
  id: string
  actionType: string
  adapterId?: string
  approvalType: string
  status: string
  createdAt: string
  details?: Record<string, any>
}
