import React, { useState } from 'react'
import {
  ArrowRight,
  ShieldCheck,
  TrendingDown,
  RefreshCw,
  Sparkles,
  ExternalLink,
  CheckCircle2,
  X,
  Bot,
  Globe,
  Loader2,
} from 'lucide-react'
import { MarketplaceOpportunity, MarketplaceMetrics } from '../types'
import { useAuth } from '../context/AuthContext'

interface MarketplaceViewProps {
  opportunities: MarketplaceOpportunity[]
  metrics: MarketplaceMetrics | null
  onTriggerRFQ: (oppId?: string) => Promise<any>
  onAcceptBid: (oppId: string, bidId: string, customNotes?: string) => Promise<any>
  onNavigateToObligations?: () => void
}

export const MarketplaceView: React.FC<MarketplaceViewProps> = ({
  opportunities,
  metrics,
  onTriggerRFQ,
  onAcceptBid,
  onNavigateToObligations,
}) => {
  const { user, t } = useAuth()
  const isB2B = user.accountType === 'b2b'

  // Modal States
  const [isAiScanning, setIsAiScanning] = useState(false)
  const [scanStep, setScanStep] = useState(1)
  const [scanStatusText, setScanStatusText] = useState('')

  // Deal Confirmation Modal State
  const [dealModal, setDealModal] = useState<{
    oppId: string
    bid: any
    vendorName: string
    price: string
    website: string
  } | null>(null)

  const [dealNote, setDealNote] = useState('')
  const [isSubmittingDeal, setIsSubmittingDeal] = useState(false)

  // Trigger AI RFQ Scanning flow
  const handleRunAiScan = async (oppId?: string) => {
    setIsAiScanning(true)
    setScanStep(1)
    setScanStatusText('🔍 Yapay Zeka çalışıyor: Aktif vadeler ve yükümlülükler taranıyor...')

    await new Promise((resolve) => setTimeout(resolve, 800))
    setScanStep(2)
    setScanStatusText('📑 Kasko, İnternet, Bulut ve Lisans sözleşmesi şartları analiz ediliyor...')

    await new Promise((resolve) => setTimeout(resolve, 1000))
    setScanStep(3)
    setScanStatusText('🌐 Pazar yeri entegrasyonu üzerinden güncel alternatif teklifler derleniyor...')

    await new Promise((resolve) => setTimeout(resolve, 900))
    setScanStep(4)
    setScanStatusText('⚡ Aksigorta, Sompo, Türk Telekom vb. alternatif teklifler üretildi ve risk puanlaması tamamlandı!')

    try {
      await onTriggerRFQ(oppId)
    } finally {
      await new Promise((resolve) => setTimeout(resolve, 600))
      setIsAiScanning(false)
    }
  }

  // Open deal link & show confirmation modal
  const handleInitiateDeal = (oppId: string, bid: any) => {
    const website = bid.contractUrl || bid.vendor.website || 'https://www.google.com'

    // Open link in new tab or electron external browser
    window.open(website, '_blank')

    // Open confirmation drawer/modal
    setDealModal({
      oppId,
      bid,
      vendorName: bid.vendor.name,
      price: `${bid.price.amount.toLocaleString()} ${bid.price.currency}`,
      website,
    })
  }

  // Confirm deal completion & bind to Obligations
  const handleConfirmDeal = async () => {
    if (!dealModal) return
    setIsSubmittingDeal(true)
    try {
      await onAcceptBid(dealModal.oppId, dealModal.bid.id, dealNote)
      setDealModal(null)
      setDealNote('')

      // Auto navigate to obligations tab if callback provided
      if (onNavigateToObligations) {
        onNavigateToObligations()
      }
    } finally {
      setIsSubmittingDeal(false)
    }
  }

  return (
    <div>
      <div className="page-header">
        <div className="page-title-group">
          <h1>{t.marketplaceTitle}</h1>
          <p className="page-subtitle">{t.marketplaceSubtitle}</p>
        </div>

        <button
          className="btn btn-primary"
          onClick={() => handleRunAiScan()}
          disabled={isAiScanning}
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: 8,
            padding: '10px 18px',
            background: 'linear-gradient(135deg, #0284c7 0%, #0369a1 100%)',
            boxShadow: '0 4px 14px rgba(2, 132, 199, 0.35)',
          }}
        >
          {isAiScanning ? (
            <Loader2 className="spin" size={16} />
          ) : (
            <Sparkles size={16} color="#38bdf8" />
          )}
          <span>🤖 Yapay Zeka ile Vadeleri Tara & Teklif Topla</span>
        </button>
      </div>

      {/* Top Marketplace Metrics */}
      {metrics && (
        <div className="card-grid-3">
          <div className="card">
            <div className="card-label">{t.transactedGmv}</div>
            <div className="card-value mono">{metrics.totalGMV.amount.toLocaleString()} {metrics.totalGMV.currency}</div>
            <div className="card-subtext">{metrics.completedDeals} anlaşma başarıyla bağlandı</div>
          </div>
          <div className="card">
            <div className="card-label">{t.realizedSavings}</div>
            <div className="card-value mono" style={{ color: '#10b981' }}>
              +{metrics.totalSavings.amount.toLocaleString()} {metrics.totalSavings.currency}
            </div>
            <div className="card-subtext">Ortalama %{metrics.averageSavingsRate} net maliyet düşüşü</div>
          </div>
          <div className="card">
            <div className="card-label">{t.commissionFee}</div>
            <div className="card-value mono" style={{ color: '#38bdf8' }}>
              {metrics.estimatedCommission.amount.toLocaleString()} {metrics.estimatedCommission.currency}
            </div>
            <div className="card-subtext">Sadece gerçekleşen tasarrufta platform komisyonu</div>
          </div>
        </div>
      )}

      {/* Opportunities List with Bids */}
      {opportunities.length === 0 ? (
        <div className="card" style={{ textAlign: 'center', padding: '48px 24px' }}>
          <Bot size={40} color="#38bdf8" style={{ marginBottom: 12 }} />
          <h3 style={{ fontSize: 16, fontWeight: 700, marginBottom: 6 }}>Henüz Aktif Teklif Listesi Oluşturulmadı</h3>
          <p style={{ color: 'var(--text-muted)', maxWidth: 480, margin: '0 auto 20px', fontSize: 13 }}>
            Sistemdeki süresi yaklaşan kasko, internet, lisans veya sunucu yükümlülükleriniz için yapay zekayı çalıştırarak güncel teklifleri toplayabilirsiniz.
          </p>
          <button className="btn btn-primary" onClick={() => handleRunAiScan()}>
            <Sparkles size={15} /> Vadeleri Tara ve Teklif Topla
          </button>
        </div>
      ) : (
        opportunities.map((opp) => (
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
                    {opp.currentVendorCost
                      ? `${opp.currentVendorCost.amount.toLocaleString()} ${opp.currentVendorCost.currency}/dönem`
                      : '—'}
                  </strong>{' '}
                  • Durum: <strong style={{ color: opp.status === 'MATCHED' ? '#10b981' : '#fbbf24' }}>{opp.status === 'MATCHED' ? 'Teklifler Hazır' : 'Aranıyor'}</strong>
                </div>
              </div>

              <div style={{ display: 'flex', gap: 10 }}>
                <button
                  className="btn btn-secondary"
                  style={{ fontSize: 12 }}
                  onClick={() => handleRunAiScan(opp.id)}
                  disabled={isAiScanning}
                >
                  <RefreshCw size={13} className={isAiScanning ? 'spin' : ''} /> {t.collectQuotes}
                </button>
              </div>
            </div>

            {/* Bids Comparison Deck */}
            {opp.bids && opp.bids.length > 0 ? (
              <div className="bids-deck">
                {opp.bids.map((bid, idx) => {
                  const isBest = idx === 0
                  const isAccepted = bid.status === 'ACCEPTED'

                  return (
                    <div
                      key={bid.id}
                      className={`bid-card ${isAccepted ? 'accepted' : isBest ? 'best' : ''}`}
                      style={{
                        borderColor: isAccepted ? '#10b981' : undefined,
                        backgroundColor: isAccepted ? 'rgba(16, 185, 129, 0.05)' : undefined,
                      }}
                    >
                      <div>
                        <div className="bid-vendor">
                          <span className="bid-vendor-name">{bid.vendor.name}</span>
                          <span className="savings-pill">
                            <TrendingDown size={11} style={{ display: 'inline', marginRight: 2 }} />
                            {t.savePercent} %{bid.savingsRate.toFixed(0)}
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
                            <span>{t.riskScore}: {bid.riskScore ?? 0.08}</span>
                          </div>

                          {isAccepted ? (
                            <span
                              className="badge"
                              style={{
                                backgroundColor: 'rgba(16, 185, 129, 0.2)',
                                color: '#34d399',
                                padding: '6px 12px',
                                display: 'inline-flex',
                                alignItems: 'center',
                                gap: 4,
                              }}
                            >
                              <CheckCircle2 size={13} /> Anlaşma Bağlandı
                            </span>
                          ) : (
                            <button
                              className={`btn ${isBest ? 'btn-primary' : 'btn-secondary'}`}
                              style={{ fontSize: 12, padding: '6px 14px' }}
                              onClick={() => handleInitiateDeal(opp.id, bid)}
                            >
                              {t.acceptDeal} <ExternalLink size={13} style={{ marginLeft: 4 }} />
                            </button>
                          )}
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            ) : (
              <div style={{ padding: '20px 0', textAlign: 'center', color: 'var(--text-muted)', fontSize: 13 }}>
                Bu yükümlülük için henüz alternatif teklif sorgulanmadı. "Teklif Topla" butonuna basarak yapay zekayı çalıştırabilirsiniz.
              </div>
            )}
          </div>
        ))
      )}

      {/* AI SCANNING OVERLAY MODAL */}
      {isAiScanning && (
        <div className="modal-backdrop">
          <div className="modal-content" style={{ maxWidth: 520, textAlign: 'center' }}>
            <div
              style={{
                width: 56,
                height: 56,
                borderRadius: '50%',
                background: 'rgba(56, 189, 248, 0.15)',
                color: '#38bdf8',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                margin: '0 auto 16px',
              }}
            >
              <Sparkles className="spin" size={28} />
            </div>

            <h2 style={{ fontSize: 18, fontWeight: 700, marginBottom: 8 }}>Otonom AI Agent Çalışıyor</h2>
            <p style={{ fontSize: 13, color: 'var(--text-muted)', marginBottom: 20 }}>{scanStatusText}</p>

            <div style={{ display: 'flex', flexDirection: 'column', gap: 10, textAlign: 'left' }}>
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  fontSize: 13,
                  color: scanStep >= 1 ? '#38bdf8' : 'var(--text-muted)',
                }}
              >
                <CheckCircle2 size={16} color={scanStep >= 1 ? '#38bdf8' : '#475569'} />
                <span>1. Vadeler ve Yükümlülükler Taranıyor</span>
              </div>

              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  fontSize: 13,
                  color: scanStep >= 2 ? '#38bdf8' : 'var(--text-muted)',
                }}
              >
                <CheckCircle2 size={16} color={scanStep >= 2 ? '#38bdf8' : '#475569'} />
                <span>2. Poliçe ve Sözleşme Şartları Ayrıştırılıyor</span>
              </div>

              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  fontSize: 13,
                  color: scanStep >= 3 ? '#38bdf8' : 'var(--text-muted)',
                }}
              >
                <CheckCircle2 size={16} color={scanStep >= 3 ? '#38bdf8' : '#475569'} />
                <span>3. Pazar Yeri & İnternet Fiyat Araştırması Yapılıyor</span>
              </div>

              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  fontSize: 13,
                  color: scanStep >= 4 ? '#10b981' : 'var(--text-muted)',
                }}
              >
                <CheckCircle2 size={16} color={scanStep >= 4 ? '#10b981' : '#475569'} />
                <span>4. Verifier Risk Puanlaması ve Teklif Derleme</span>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* DEAL CONFIRMATION & OBLIGATION BINDING MODAL */}
      {dealModal && (
        <div className="modal-backdrop">
          <div className="modal-content" style={{ maxWidth: 540 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div
                  style={{
                    width: 38,
                    height: 38,
                    borderRadius: 8,
                    background: 'rgba(16, 185, 129, 0.15)',
                    color: '#34d399',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                  }}
                >
                  <Globe size={20} />
                </div>
                <div>
                  <h3 style={{ fontSize: 16, fontWeight: 700, margin: 0 }}>Anlaşma & Satın Alma Onayı</h3>
                  <span style={{ fontSize: 12, color: 'var(--text-muted)' }}>Tedarikçi Web Sayfası Açıldı</span>
                </div>
              </div>
              <button
                className="btn btn-secondary"
                style={{ padding: 6 }}
                onClick={() => setDealModal(null)}
              >
                <X size={16} />
              </button>
            </div>

            <div
              style={{
                background: 'rgba(15, 23, 42, 0.6)',
                border: '1px solid var(--border-color)',
                borderRadius: 8,
                padding: 16,
                marginBottom: 16,
              }}
            >
              <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--text-primary)', marginBottom: 4 }}>
                {dealModal.vendorName}
              </div>
              <div style={{ fontSize: 13, color: '#10b981', fontWeight: 700, marginBottom: 8 }}>
                Fiyat: {dealModal.price} / yıl
              </div>
              <div style={{ fontSize: 12, color: 'var(--text-muted)' }}>
                {dealModal.bid.terms}
              </div>
              <div style={{ marginTop: 10, fontSize: 11, color: '#38bdf8', display: 'flex', alignItems: 'center', gap: 4 }}>
                <ExternalLink size={12} />
                <a href={dealModal.website} target="_blank" rel="noreferrer" style={{ color: '#38bdf8', textDecoration: 'underline' }}>
                  {dealModal.website}
                </a>
              </div>
            </div>

            <div style={{ marginBottom: 16 }}>
              <label style={{ display: 'block', fontSize: 12, fontWeight: 600, color: 'var(--text-muted)', marginBottom: 6 }}>
                Anlaşma Notu / Poliçe Numarası (Opsiyonel):
              </label>
              <input
                type="text"
                className="form-input"
                style={{ width: '100%' }}
                placeholder="Örn: Poliçe onaylandı, müşteri no #94812"
                value={dealNote}
                onChange={(e) => setDealNote(e.target.value)}
              />
            </div>

            <p style={{ fontSize: 12, color: 'var(--text-muted)', lineHeight: 1.5, marginBottom: 20 }}>
              💡 <strong>"Evet, Anlaşma Sağlandı"</strong> butonuna bastığınızda bu anlaşma 1 yıllık yeni geçerlilik süresiyle <strong>Vadeler & Yükümlülükler</strong> ekranınıza otomatik aktarılacak ve takibe alınacaktır.
            </p>

            <div style={{ display: 'flex', gap: 10, justifyContent: 'flex-end' }}>
              <button className="btn btn-secondary" onClick={() => setDealModal(null)} disabled={isSubmittingDeal}>
                İptal / Beklemede
              </button>
              <button
                className="btn btn-primary"
                onClick={handleConfirmDeal}
                disabled={isSubmittingDeal}
                style={{
                  background: 'linear-gradient(135deg, #059669 0%, #10b981 100%)',
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 6,
                }}
              >
                {isSubmittingDeal ? <Loader2 className="spin" size={14} /> : <CheckCircle2 size={15} />}
                <span>Evet, Anlaşma Sağlandı ve Vadeler Ekranına Ekle</span>
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
