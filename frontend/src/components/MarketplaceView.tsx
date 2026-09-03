import React from 'react'
import {
  ArrowRight,
  ShieldCheck,
  TrendingDown,
  RefreshCw,
} from 'lucide-react'
import { MarketplaceOpportunity, MarketplaceMetrics } from '../types'
import { useAuth } from '../context/AuthContext'

interface MarketplaceViewProps {
  opportunities: MarketplaceOpportunity[]
  metrics: MarketplaceMetrics | null
  onTriggerRFQ: (oppId: string) => void
  onAcceptBid: (oppId: string, bidId: string) => void
}

export const MarketplaceView: React.FC<MarketplaceViewProps> = ({
  opportunities,
  metrics,
  onTriggerRFQ,
  onAcceptBid,
}) => {
  const { user, t } = useAuth()
  const isB2B = user.accountType === 'b2b'

  // Personal B2C opportunities when in personal mode
  const personalOpportunities: MarketplaceOpportunity[] = [
    {
      id: 'opp-p-01',
      obligationId: 'b2c-001',
      userId: '00000000-0000-0000-0000-000000000001',
      category: 'insurance',
      currentVendorName: 'Anadolu Sigorta Araç Kaskosu',
      currentVendorCost: { amount: 18500, currency: 'TRY' },
      status: 'MATCHED',
      bidCount: 2,
      bids: [
        {
          id: 'bid-p-01',
          opportunityId: 'opp-p-01',
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
          terms: 'Peşin fiyatına 6 taksit, yetkili servis güvencesi.',
          contractUrl: 'https://www.aksigorta.com.tr',
          status: 'PENDING',
          riskScore: 0.08,
          riskNotes: 'Sektörün en yüksek müşteri memnuniyet skoru, Verifier onaylı.',
          createdAt: new Date().toISOString(),
        },
        {
          id: 'bid-p-02',
          opportunityId: 'opp-p-01',
          vendor: {
            id: 'v-sompo',
            name: 'Sompo Sigorta Kasko Plus',
            category: 'insurance',
            rating: 4.7,
            website: 'https://www.sompojapan.com.tr',
            description: 'Sıfır araç koruma klozu ve 7/24 yol yardım.',
            verified: true,
          },
          price: { amount: 13500, currency: 'TRY' },
          savingsAmount: { amount: 5000, currency: 'TRY' },
          savingsRate: 27.0,
          terms: 'Anlaşmalı özel ve yetkili servis ağı.',
          contractUrl: 'https://www.sompojapan.com.tr',
          status: 'PENDING',
          riskScore: 0.12,
          riskNotes: 'Güvenilir teminat yapısı.',
          createdAt: new Date().toISOString(),
        },
      ],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    },
    {
      id: 'opp-p-02',
      obligationId: 'b2c-003',
      userId: '00000000-0000-0000-0000-000000000001',
      category: 'telecom',
      currentVendorName: 'Turkcell Superonline 1000 Mbps',
      currentVendorCost: { amount: 5880, currency: 'TRY' },
      status: 'MATCHED',
      bidCount: 1,
      bids: [
        {
          id: 'bid-p-03',
          opportunityId: 'opp-p-02',
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
          terms: '12 ay taahhüt, ücretsiz fiber kurulum.',
          contractUrl: 'https://www.turktelekom.com.tr',
          status: 'PENDING',
          riskScore: 0.10,
          riskNotes: 'Geniş kapsama alanı.',
          createdAt: new Date().toISOString(),
        },
      ],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    },
  ]

  const activeOpps = isB2B ? opportunities : personalOpportunities

  return (
    <div>
      <div className="page-header">
        <div className="page-title-group">
          <h1>{t.marketplaceTitle}</h1>
          <p className="page-subtitle">{t.marketplaceSubtitle}</p>
        </div>
      </div>

      {/* Top Marketplace Metrics */}
      {metrics && (
        <div className="card-grid-3">
          <div className="card">
            <div className="card-label">{t.transactedGmv}</div>
            <div className="card-value mono">${metrics.totalGMV.amount.toLocaleString()}</div>
            <div className="card-subtext">{metrics.completedDeals} deals closed</div>
          </div>
          <div className="card">
            <div className="card-label">{t.realizedSavings}</div>
            <div className="card-value mono" style={{ color: '#10b981' }}>
              +${metrics.totalSavings.amount.toLocaleString()}
            </div>
            <div className="card-subtext">Avg {metrics.averageSavingsRate}% cost reduction</div>
          </div>
          <div className="card">
            <div className="card-label">{t.commissionFee}</div>
            <div className="card-value mono" style={{ color: '#38bdf8' }}>
              ${metrics.estimatedCommission.amount.toLocaleString()}
            </div>
            <div className="card-subtext">Zero upfront procurement retainer</div>
          </div>
        </div>
      )}

      {/* Opportunities List with Bids */}
      {activeOpps.length === 0 ? (
        <div className="card" style={{ textAlign: 'center', padding: '48px 24px' }}>
          <p style={{ color: 'var(--text-muted)' }}>{t.noDeals}</p>
        </div>
      ) : (
        activeOpps.map((opp) => (
          <div key={opp.id} className="opportunity-block">
            <div className="opportunity-header">
              <div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <h3 style={{ fontSize: 16, fontWeight: 700 }}>{opp.currentVendorName}</h3>
                  <span className="badge badge-medium" style={{ textTransform: 'uppercase' }}>
                    {opp.category}
                  </span>
                </div>
                <div style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 4 }}>
                  {t.currentCost}:{' '}
                  <strong style={{ color: 'var(--text-primary)' }}>
                    {opp.currentVendorCost ? `${opp.currentVendorCost.amount.toLocaleString()} ${opp.currentVendorCost.currency}/yr` : '—'}
                  </strong>{' '}
                  • Durum: <strong>{opp.status}</strong>
                </div>
              </div>

              <div style={{ display: 'flex', gap: 10 }}>
                <button
                  className="btn btn-secondary"
                  style={{ fontSize: 12 }}
                  onClick={() => onTriggerRFQ(opp.id)}
                >
                  <RefreshCw size={13} /> {t.collectQuotes}
                </button>
              </div>
            </div>

            {/* Bids Comparison Deck */}
            {opp.bids && opp.bids.length > 0 ? (
              <div className="bids-deck">
                {opp.bids.map((bid, idx) => {
                  const isBest = idx === 0

                  return (
                    <div key={bid.id} className={`bid-card ${isBest ? 'best' : ''}`}>
                      <div>
                        <div className="bid-vendor">
                          <span className="bid-vendor-name">{bid.vendor.name}</span>
                          <span className="savings-pill">
                            <TrendingDown size={11} style={{ display: 'inline', marginRight: 2 }} />
                            {t.savePercent} {bid.savingsRate.toFixed(0)}%
                          </span>
                        </div>

                        <div className="bid-price">
                          {bid.price.amount.toLocaleString()} {bid.price.currency}
                          <span style={{ fontSize: 12, fontWeight: 400, color: 'var(--text-muted)', marginLeft: 4 }}>
                            /yıl
                          </span>
                        </div>

                        <div className="bid-terms">{bid.terms}</div>
                      </div>

                      <div>
                        <div className="bid-footer">
                          <div className="risk-pill">
                            <ShieldCheck size={13} color="#10b981" />
                            <span>{t.riskScore}: {bid.riskScore ?? 0.1}</span>
                          </div>

                          <button
                            className={`btn ${isBest ? 'btn-primary' : 'btn-secondary'}`}
                            style={{ fontSize: 12, padding: '5px 12px' }}
                            onClick={() => onAcceptBid(opp.id, bid.id)}
                          >
                            {t.acceptDeal} <ArrowRight size={13} />
                          </button>
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            ) : (
              <div style={{ padding: '20px 0', textAlign: 'center', color: 'var(--text-muted)', fontSize: 13 }}>
                {t.noDeals}
              </div>
            )}
          </div>
        ))
      )}
    </div>
  )
}
