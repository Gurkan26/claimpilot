import React, { useState, useEffect } from 'react'
import {
  Cpu,
  Bot,
  Layers,
  ShieldAlert,
  ShieldCheck,
  Zap,
  CheckCircle2,
  RefreshCw,
  Sliders,
  Play,
  Server,
  BookOpen,
  Mail,
  Calendar,
  MessageSquare,
  Plus,
  ArrowRight,
  Sparkles,
  ToggleLeft,
  ToggleRight,
  AlertCircle,
  ExternalLink,
  Code2,
} from 'lucide-react'
import {
  HarnessConfig,
  LLMConfigItem,
  MCPConfigItem,
  LLMRole,
  SimulationResult,
} from '../types'
import { api } from '../services/api'
import { useAuth } from '../context/AuthContext'

interface AdminViewProps {
  onShowToast: (message: string, type: 'success' | 'error' | 'info') => void
}

type AdminSubTab = 'llm' | 'mcp' | 'guardrails' | 'simulator'

export const AdminView: React.FC<AdminViewProps> = ({ onShowToast }) => {
  const { t } = useAuth()
  const [subTab, setSubTab] = useState<AdminSubTab>('llm')
  const [config, setConfig] = useState<HarnessConfig | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [testingRole, setTestingRole] = useState<LLMRole | null>(null)

  // Simulation state
  const [simSample, setSimSample] = useState(
    'Adobe Creative Cloud Enterprise Agreement: Term ends Nov 30, 2025. 12-month auto-renewal clause applies unless 30-day prior written non-renewal notice is delivered to renewals@adobe.com. Annual commitment: $3,600 USD.'
  )
  const [simulating, setSimulating] = useState(false)
  const [simResult, setSimResult] = useState<SimulationResult | null>(null)

  // New MCP modal state
  const [showAddMcpModal, setShowAddMcpModal] = useState(false)
  const [newMcpName, setNewMcpName] = useState('')
  const [newMcpEndpoint, setNewMcpEndpoint] = useState('')
  const [newMcpCategory, setNewMcpCategory] = useState<'knowledge' | 'email' | 'calendar' | 'chat' | 'custom'>('custom')

  useEffect(() => {
    loadConfig()
  }, [])

  const loadConfig = async () => {
    setLoading(true)
    try {
      const data = await api.getHarnessConfig()
      setConfig(data)
    } catch {
      onShowToast('Harness yapılandırması yüklenemedi', 'error')
    } finally {
      setLoading(false)
    }
  }

  const handleSaveConfig = async () => {
    if (!config) return
    setSaving(true)
    try {
      const updated = await api.updateHarnessConfig(config)
      setConfig(updated)
      onShowToast(t.hotSwapSuccess, 'success')
    } catch {
      onShowToast('Hata: Ayarlar uygulanamadı', 'error')
    } finally {
      setSaving(false)
    }
  }

  const handleTestLLM = async (role: LLMRole) => {
    if (!config) return
    setTestingRole(role)
    const target = role === 'analyst' ? config.analystLLM : config.verifierLLM

    try {
      const res = await api.testLLMConnection(role, target)
      if (res.success) {
        onShowToast(`⚡ ${role.toUpperCase()}: ${res.message} (${res.latencyMs}ms)`, 'success')
        setConfig({
          ...config,
          [role === 'analyst' ? 'analystLLM' : 'verifierLLM']: {
            ...target,
            status: 'connected',
            lastPingMs: res.latencyMs,
          },
        })
      } else {
        onShowToast(`${role.toUpperCase()} bağlantı hatası!`, 'error')
      }
    } catch {
      onShowToast('Bağlantı testi başarısız oldu', 'error')
    } finally {
      setTestingRole(null)
    }
  }

  const handleToggleMCP = async (id: string, currentEnabled: boolean) => {
    if (!config) return
    const newEnabled = !currentEnabled
    try {
      await api.toggleMCPAdapter(id, newEnabled)
      setConfig({
        ...config,
        mcpAdapters: config.mcpAdapters.map((a) =>
          a.id === id ? { ...a, enabled: newEnabled, status: newEnabled ? 'online' : 'disabled' } : a
        ),
      })
      onShowToast(`MCP ${id.toUpperCase()} ${newEnabled ? 'etkinleştirildi' : 'devre dışı bırakıldı'}`, 'info')
    } catch {
      onShowToast('MCP durumu değiştirilemedi', 'error')
    }
  }

  const handleAddCustomMcp = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!config || !newMcpName.trim() || !newMcpEndpoint.trim()) return

    const newMcp = await api.registerMCPServer({
      name: newMcpName.trim(),
      endpointOrCmd: newMcpEndpoint.trim(),
      category: newMcpCategory,
      description: 'Harici MCP Protokol Servisi',
      transport: 'sse',
    })

    setConfig({
      ...config,
      mcpAdapters: [...config.mcpAdapters, newMcp],
    })

    setShowAddMcpModal(false)
    setNewMcpName('')
    setNewMcpEndpoint('')
    onShowToast(`Yeni MCP sunucusu "${newMcp.name}" başarıyla bağlandı!`, 'success')
  }

  const handleRunSimulation = async () => {
    setSimulating(true)
    setSimResult(null)
    try {
      const res = await api.runHarnessSimulation(simSample)
      setSimResult(res)
      onShowToast('Harness simülasyonu başarıyla tamamlandı!', 'success')
    } catch {
      onShowToast('Simülasyon çalıştırılamadı', 'error')
    } finally {
      setSimulating(false)
    }
  }

  if (loading || !config) {
    return (
      <div className="view-content" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: 400 }}>
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 12 }}>
          <RefreshCw className="spin" size={28} color="#a855f7" />
          <span style={{ color: '#94a3b8', fontSize: 14 }}>Agent Harness yükleniyor...</span>
        </div>
      </div>
    )
  }

  const activeMcpCount = config.mcpAdapters.filter((a) => a.enabled).length

  return (
    <div className="view-content admin-view">
      {/* Admin Header & Stats */}
      <div className="view-header" style={{ marginBottom: 20 }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 4 }}>
            <h1 className="view-title" style={{ margin: 0 }}>
              {t.adminWelcome}
            </h1>
            <span
              className="badge"
              style={{
                background: 'linear-gradient(135deg, rgba(168,85,247,0.3) 0%, rgba(99,102,241,0.3) 100%)',
                color: '#e9d5ff',
                borderColor: '#a855f7',
                fontSize: 10,
                fontWeight: 700,
              }}
            >
              AGENT HARNESS v2.4
            </span>
          </div>
          <p className="view-subtitle">{t.adminSubtitle}</p>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <button
            className="btn btn-secondary"
            onClick={loadConfig}
            title="Yenile"
            style={{ padding: '8px 12px' }}
          >
            <RefreshCw size={14} />
          </button>
          <button
            className="btn btn-primary"
            onClick={handleSaveConfig}
            disabled={saving}
            style={{
              background: 'linear-gradient(135deg, #9333ea 0%, #6366f1 100%)',
              border: 'none',
              boxShadow: '0 4px 14px rgba(147, 51, 234, 0.35)',
            }}
          >
            {saving ? <RefreshCw className="spin" size={14} /> : <CheckCircle2 size={14} />}
            <span>{saving ? 'Uygulanıyor...' : t.applyChanges}</span>
          </button>
        </div>
      </div>

      {/* Top Telemetry Pills */}
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
          gap: 14,
          marginBottom: 20,
        }}
      >
        <div className="admin-stat-card">
          <div className="admin-stat-icon" style={{ background: 'rgba(56, 189, 248, 0.12)', color: '#38bdf8' }}>
            <Cpu size={18} />
          </div>
          <div className="admin-stat-info">
            <span className="admin-stat-label">LLM Engine (Ollama)</span>
            <span className="admin-stat-val">2 Aktif Model</span>
            <span className="admin-stat-sub">Analyst + Verifier rolleri ayrık</span>
          </div>
        </div>

        <div className="admin-stat-card">
          <div className="admin-stat-icon" style={{ background: 'rgba(168, 85, 247, 0.12)', color: '#c084fc' }}>
            <BookOpen size={18} />
          </div>
          <div className="admin-stat-info">
            <span className="admin-stat-label">DeepWiki MCP</span>
            <span className="admin-stat-val">Online (SSE)</span>
            <span className="admin-stat-sub">14ms Emsal & Politika Erişimi</span>
          </div>
        </div>

        <div className="admin-stat-card">
          <div className="admin-stat-icon" style={{ background: 'rgba(16, 185, 129, 0.12)', color: '#10b981' }}>
            <Zap size={18} />
          </div>
          <div className="admin-stat-info">
            <span className="admin-stat-label">MCP Araç Bağlantıları</span>
            <span className="admin-stat-val">{activeMcpCount} / {config.mcpAdapters.length} Aktif</span>
            <span className="admin-stat-sub">Gmail, Calendar, Slack, DeepWiki</span>
          </div>
        </div>

        <div className="admin-stat-card">
          <div className="admin-stat-icon" style={{ background: 'rgba(245, 158, 11, 0.12)', color: '#f59e0b' }}>
            <ShieldCheck size={18} />
          </div>
          <div className="admin-stat-info">
            <span className="admin-stat-label">Harness Politikası</span>
            <span className="admin-stat-val">Human-in-the-Loop</span>
            <span className="admin-stat-sub">Onaysız mali işlem yapılmaz</span>
          </div>
        </div>
      </div>

      {/* Admin Sub-Tabs Navigation */}
      <div className="admin-tabs-bar" style={{ marginBottom: 20 }}>
        <button
          className={`admin-tab-btn ${subTab === 'llm' ? 'active' : ''}`}
          onClick={() => setSubTab('llm')}
        >
          <Cpu size={15} />
          <span>{t.tabLlmConfig}</span>
        </button>

        <button
          className={`admin-tab-btn ${subTab === 'mcp' ? 'active' : ''}`}
          onClick={() => setSubTab('mcp')}
        >
          <Layers size={15} />
          <span>{t.tabMcpHarness}</span>
        </button>

        <button
          className={`admin-tab-btn ${subTab === 'guardrails' ? 'active' : ''}`}
          onClick={() => setSubTab('guardrails')}
        >
          <Sliders size={15} />
          <span>{t.tabGuardrails}</span>
        </button>

        <button
          className={`admin-tab-btn ${subTab === 'simulator' ? 'active' : ''}`}
          onClick={() => setSubTab('simulator')}
        >
          <Play size={15} />
          <span>{t.tabSimulator}</span>
        </button>
      </div>

      {/* ========================================================================= */}
      {/* SUB-TAB 1: DUAL-LLM ENGINE (OLLAMA) */}
      {/* ========================================================================= */}
      {subTab === 'llm' && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(440px, 1fr))', gap: 20 }}>
          {/* Card 1: Analyst LLM */}
          <div className="admin-card">
            <div className="admin-card-header">
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div className="agent-avatar" style={{ background: 'linear-gradient(135deg, #0284c7 0%, #38bdf8 100%)' }}>
                  A1
                </div>
                <div>
                  <h3 style={{ margin: 0, fontSize: 16, color: '#f8fafc' }}>{t.analystRoleTitle}</h3>
                  <p style={{ margin: 0, fontSize: 12, color: '#94a3b8' }}>
                    Sözleşme ayrıştırma, taahhüt & bildirim süresi tespiti
                  </p>
                </div>
              </div>

              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <span className="badge badge-success" style={{ fontSize: 11 }}>
                  ● {config.analystLLM.status === 'connected' ? `Online (${config.analystLLM.lastPingMs || 24}ms)` : 'Offline'}
                </span>
                <button
                  className="btn btn-secondary"
                  style={{ padding: '4px 8px', fontSize: 11 }}
                  onClick={() => handleTestLLM('analyst')}
                  disabled={testingRole === 'analyst'}
                >
                  {testingRole === 'analyst' ? <RefreshCw className="spin" size={11} /> : <Zap size={11} />}
                  <span>{t.testConnection}</span>
                </button>
              </div>
            </div>

            <div className="admin-card-body" style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                <div>
                  <label className="form-label">{t.llmProvider}</label>
                  <select
                    className="form-input"
                    value={config.analystLLM.provider}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        analystLLM: { ...config.analystLLM, provider: e.target.value as any },
                      })
                    }
                  >
                    <option value="ollama">Ollama (Local / Dedicated)</option>
                    <option value="openai">OpenAI (GPT-4o / Mini)</option>
                    <option value="anthropic">Anthropic Claude</option>
                    <option value="gemini">Google Gemini 2.5/3.0</option>
                    <option value="custom">Custom OpenAI-Compatible API</option>
                  </select>
                </div>

                <div>
                  <label className="form-label">{t.llmModel}</label>
                  <input
                    type="text"
                    className="form-input"
                    value={config.analystLLM.model}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        analystLLM: { ...config.analystLLM, model: e.target.value },
                      })
                    }
                    placeholder="Örn: gemma2:9b, llama3.2:3b"
                  />
                </div>
              </div>

              <div>
                <label className="form-label">{t.llmEndpoint}</label>
                <input
                  type="text"
                  className="form-input"
                  value={config.analystLLM.endpoint}
                  onChange={(e) =>
                    setConfig({
                      ...config,
                      analystLLM: { ...config.analystLLM, endpoint: e.target.value },
                    })
                  }
                  placeholder="http://localhost:11434"
                />
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                <div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                    <label className="form-label">{t.llmTemperature}</label>
                    <span style={{ fontSize: 11, color: '#38bdf8', fontWeight: 600 }}>
                      {config.analystLLM.temperature}
                    </span>
                  </div>
                  <input
                    type="range"
                    min="0"
                    max="1"
                    step="0.05"
                    value={config.analystLLM.temperature}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        analystLLM: { ...config.analystLLM, temperature: parseFloat(e.target.value) },
                      })
                    }
                    style={{ width: '100%', accentColor: '#38bdf8' }}
                  />
                </div>

                <div>
                  <label className="form-label">{t.llmMaxTokens}</label>
                  <input
                    type="number"
                    className="form-input"
                    value={config.analystLLM.maxTokens}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        analystLLM: { ...config.analystLLM, maxTokens: parseInt(e.target.value) || 2048 },
                      })
                    }
                  />
                </div>
              </div>

              <div>
                <label className="form-label">Analyst Sistem Rol Promptu (Harness System Instruction)</label>
                <textarea
                  className="form-input"
                  rows={3}
                  value={config.analystLLM.systemPrompt}
                  onChange={(e) =>
                    setConfig({
                      ...config,
                      analystLLM: { ...config.analystLLM, systemPrompt: e.target.value },
                    })
                  }
                  style={{ fontSize: 12, resize: 'vertical' }}
                />
              </div>
            </div>
          </div>

          {/* Card 2: Verifier LLM */}
          <div className="admin-card">
            <div className="admin-card-header">
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div className="agent-avatar" style={{ background: 'linear-gradient(135deg, #7c3aed 0%, #c084fc 100%)' }}>
                  V2
                </div>
                <div>
                  <h3 style={{ margin: 0, fontSize: 16, color: '#f8fafc' }}>{t.verifierRoleTitle}</h3>
                  <p style={{ margin: 0, fontSize: 12, color: '#94a3b8' }}>
                    Denetçi, PII kontrolü & sözleşme tutarlılık onayı
                  </p>
                </div>
              </div>

              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <span className="badge badge-success" style={{ fontSize: 11 }}>
                  ● {config.verifierLLM.status === 'connected' ? `Online (${config.verifierLLM.lastPingMs || 19}ms)` : 'Offline'}
                </span>
                <button
                  className="btn btn-secondary"
                  style={{ padding: '4px 8px', fontSize: 11 }}
                  onClick={() => handleTestLLM('verifier')}
                  disabled={testingRole === 'verifier'}
                >
                  {testingRole === 'verifier' ? <RefreshCw className="spin" size={11} /> : <Zap size={11} />}
                  <span>{t.testConnection}</span>
                </button>
              </div>
            </div>

            <div className="admin-card-body" style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                <div>
                  <label className="form-label">{t.llmProvider}</label>
                  <select
                    className="form-input"
                    value={config.verifierLLM.provider}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        verifierLLM: { ...config.verifierLLM, provider: e.target.value as any },
                      })
                    }
                  >
                    <option value="ollama">Ollama (Local / Dedicated)</option>
                    <option value="openai">OpenAI (GPT-4o / Mini)</option>
                    <option value="anthropic">Anthropic Claude</option>
                    <option value="gemini">Google Gemini 2.5/3.0</option>
                    <option value="custom">Custom OpenAI-Compatible API</option>
                  </select>
                </div>

                <div>
                  <label className="form-label">{t.llmModel}</label>
                  <input
                    type="text"
                    className="form-input"
                    value={config.verifierLLM.model}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        verifierLLM: { ...config.verifierLLM, model: e.target.value },
                      })
                    }
                    placeholder="Örn: qwen2.5:7b, mistral:7b"
                  />
                </div>
              </div>

              <div>
                <label className="form-label">{t.llmEndpoint}</label>
                <input
                  type="text"
                  className="form-input"
                  value={config.verifierLLM.endpoint}
                  onChange={(e) =>
                    setConfig({
                      ...config,
                      verifierLLM: { ...config.verifierLLM, endpoint: e.target.value },
                    })
                  }
                  placeholder="http://localhost:11434"
                />
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                <div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                    <label className="form-label">{t.llmTemperature}</label>
                    <span style={{ fontSize: 11, color: '#c084fc', fontWeight: 600 }}>
                      {config.verifierLLM.temperature}
                    </span>
                  </div>
                  <input
                    type="range"
                    min="0"
                    max="1"
                    step="0.05"
                    value={config.verifierLLM.temperature}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        verifierLLM: { ...config.verifierLLM, temperature: parseFloat(e.target.value) },
                      })
                    }
                    style={{ width: '100%', accentColor: '#c084fc' }}
                  />
                </div>

                <div>
                  <label className="form-label">{t.llmMaxTokens}</label>
                  <input
                    type="number"
                    className="form-input"
                    value={config.verifierLLM.maxTokens}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        verifierLLM: { ...config.verifierLLM, maxTokens: parseInt(e.target.value) || 2048 },
                      })
                    }
                  />
                </div>
              </div>

              <div>
                <label className="form-label">Verifier Sistem Rol Promptu (Harness System Instruction)</label>
                <textarea
                  className="form-input"
                  rows={3}
                  value={config.verifierLLM.systemPrompt}
                  onChange={(e) =>
                    setConfig({
                      ...config,
                      verifierLLM: { ...config.verifierLLM, systemPrompt: e.target.value },
                    })
                  }
                  style={{ fontSize: 12, resize: 'vertical' }}
                />
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* SUB-TAB 2: MCP ADAPTER HARNESS & DEEPWIKI */}
      {/* ========================================================================= */}
      {subTab === 'mcp' && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
          {/* DeepWiki Featured Box */}
          <div
            className="admin-card"
            style={{
              background: 'radial-gradient(ellipse at top left, rgba(168, 85, 247, 0.15), rgba(15, 23, 42, 0.95))',
              borderColor: 'rgba(168, 85, 247, 0.35)',
            }}
          >
            <div className="admin-card-header">
              <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <div
                  style={{
                    width: 44,
                    height: 44,
                    borderRadius: 10,
                    background: 'rgba(168, 85, 247, 0.2)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    border: '1px solid rgba(168, 85, 247, 0.4)',
                  }}
                >
                  <BookOpen size={22} color="#c084fc" />
                </div>
                <div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <h3 style={{ margin: 0, fontSize: 17, color: '#f8fafc' }}>
                      DeepWiki Knowledge Base MCP
                    </h3>
                    <span className="badge" style={{ background: 'rgba(168, 85, 247, 0.2)', color: '#d8b4fe', borderColor: '#a855f7' }}>
                      CORE HARNESS TOOL
                    </span>
                  </div>
                  <p style={{ margin: 0, fontSize: 12, color: '#cbd5e1', marginTop: 2 }}>
                    {t.deepWikiDesc}
                  </p>
                </div>
              </div>

              <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={() => handleToggleMCP('deepwiki', true)}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 6,
                    color: '#d8b4fe',
                    borderColor: 'rgba(168, 85, 247, 0.4)',
                  }}
                >
                  <ToggleRight size={18} color="#a855f7" />
                  <span>Aktif (Online)</span>
                </button>
              </div>
            </div>

            <div className="admin-card-body" style={{ marginTop: 12 }}>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))', gap: 12 }}>
                <div style={{ background: 'rgba(0,0,0,0.3)', padding: 12, borderRadius: 8 }}>
                  <span style={{ fontSize: 11, color: '#94a3b8', display: 'block', marginBottom: 4 }}>DeepWiki SSE Endpoint</span>
                  <code style={{ fontSize: 12, color: '#38bdf8' }}>http://localhost:8899/mcp/deepwiki/sse</code>
                </div>

                <div style={{ background: 'rgba(0,0,0,0.3)', padding: 12, borderRadius: 8 }}>
                  <span style={{ fontSize: 11, color: '#94a3b8', display: 'block', marginBottom: 4 }}>Desteklenen Ajan Araçları</span>
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: 5 }}>
                    {['search_wiki', 'query_precedents', 'get_policy_rule', 'verify_clause'].map((tool) => (
                      <span key={tool} className="badge badge-low" style={{ fontSize: 10, fontFamily: 'monospace' }}>
                        {tool}
                      </span>
                    ))}
                  </div>
                </div>

                <div style={{ background: 'rgba(0,0,0,0.3)', padding: 12, borderRadius: 8 }}>
                  <span style={{ fontSize: 11, color: '#94a3b8', display: 'block', marginBottom: 4 }}>Emsal Kapsamı</span>
                  <span style={{ fontSize: 12, color: '#f1f5f9', fontWeight: 500 }}>
                    Sözleşme Fesih Hükümleri, KVKK Maskeleme Listesi, Tedarikçi SLA Standartları
                  </span>
                </div>
              </div>
            </div>
          </div>

          {/* Other MCP Adapters Grid */}
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <h3 style={{ margin: 0, fontSize: 16, color: '#f8fafc' }}>Aktif MCP Adaptörleri</h3>
            <button
              className="btn btn-secondary"
              style={{ fontSize: 12, display: 'flex', alignItems: 'center', gap: 6 }}
              onClick={() => setShowAddMcpModal(true)}
            >
              <Plus size={14} />
              <span>{t.addMcpServer}</span>
            </button>
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: 16 }}>
            {config.mcpAdapters
              .filter((a) => a.id !== 'deepwiki')
              .map((adapter) => {
                let icon = <Server size={18} color="#94a3b8" />
                if (adapter.category === 'email') icon = <Mail size={18} color="#38bdf8" />
                if (adapter.category === 'calendar') icon = <Calendar size={18} color="#10b981" />
                if (adapter.category === 'chat') icon = <MessageSquare size={18} color="#f59e0b" />

                return (
                  <div key={adapter.id} className="admin-card" style={{ padding: 16 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 10 }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                        <div
                          style={{
                            width: 36,
                            height: 36,
                            borderRadius: 8,
                            background: 'rgba(255, 255, 255, 0.05)',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                          }}
                        >
                          {icon}
                        </div>
                        <div>
                          <h4 style={{ margin: 0, fontSize: 14, color: '#f8fafc' }}>{adapter.name}</h4>
                          <span style={{ fontSize: 11, color: '#64748b' }}>Transport: {adapter.transport.toUpperCase()}</span>
                        </div>
                      </div>

                      <button
                        type="button"
                        onClick={() => handleToggleMCP(adapter.id, adapter.enabled)}
                        style={{
                          background: 'none',
                          border: 'none',
                          cursor: 'pointer',
                          padding: 0,
                        }}
                      >
                        {adapter.enabled ? (
                          <ToggleRight size={26} color="#10b981" />
                        ) : (
                          <ToggleLeft size={26} color="#64748b" />
                        )}
                      </button>
                    </div>

                    <p style={{ fontSize: 12, color: '#94a3b8', margin: '0 0 12px 0', minHeight: 34 }}>
                      {adapter.description}
                    </p>

                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4, marginBottom: 10 }}>
                      {adapter.actions.map((act) => (
                        <span key={act} className="badge badge-low" style={{ fontSize: 9, fontFamily: 'monospace' }}>
                          {act}
                        </span>
                      ))}
                    </div>

                    <div
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                        borderTop: '1px solid rgba(255,255,255,0.06)',
                        paddingTop: 8,
                        fontSize: 11,
                        color: '#64748b',
                      }}
                    >
                      <span>Gecikme: {adapter.latencyMs || 25}ms</span>
                      <span style={{ color: adapter.enabled ? '#10b981' : '#64748b' }}>
                        ● {adapter.enabled ? 'Aktif' : 'Devre Dışı'}
                      </span>
                    </div>
                  </div>
                )
              })}
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* SUB-TAB 3: GUARDRAILS & POLICY */}
      {/* ========================================================================= */}
      {subTab === 'guardrails' && (
        <div style={{ maxWidth: 780, display: 'flex', flexDirection: 'column', gap: 16 }}>
          <div className="admin-card">
            <h3 style={{ margin: '0 0 12px 0', fontSize: 16, color: '#f8fafc' }}>
              Agent Harness Güvenlik ve İcra Politikaları
            </h3>
            <p style={{ margin: '0 0 20px 0', fontSize: 13, color: '#94a3b8' }}>
              Otonom ajanların harici araçları çalıştırma ve şirket adına eylem alma sınırlarını belirleyin.
            </p>

            <div style={{ display: 'flex', flexDirection: 'column', gap: 18 }}>
              {/* Policy 1: Human in the loop */}
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  background: 'rgba(255, 255, 255, 0.02)',
                  padding: 14,
                  borderRadius: 10,
                  border: '1px solid rgba(255,255,255,0.06)',
                }}
              >
                <div>
                  <span style={{ fontSize: 14, fontWeight: 600, color: '#f1f5f9', display: 'block' }}>
                    {t.humanApprovalToggle}
                  </span>
                  <span style={{ fontSize: 12, color: '#94a3b8' }}>
                    İptal e-postası gönderme veya teklif kabul etme gibi mali sonuç doğuran tüm eylemler kullanıcı onayına sunulur.
                  </span>
                </div>
                <button
                  type="button"
                  onClick={() =>
                    setConfig({
                      ...config,
                      policy: {
                        ...config.policy,
                        requireHumanApproval: !config.policy.requireHumanApproval,
                      },
                    })
                  }
                  style={{ background: 'none', border: 'none', cursor: 'pointer' }}
                >
                  {config.policy.requireHumanApproval ? (
                    <ToggleRight size={28} color="#38bdf8" />
                  ) : (
                    <ToggleLeft size={28} color="#64748b" />
                  )}
                </button>
              </div>

              {/* Policy 2: PII Masking */}
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  background: 'rgba(255, 255, 255, 0.02)',
                  padding: 14,
                  borderRadius: 10,
                  border: '1px solid rgba(255,255,255,0.06)',
                }}
              >
                <div>
                  <span style={{ fontSize: 14, fontWeight: 600, color: '#f1f5f9', display: 'block' }}>
                    {t.piiStrictness}
                  </span>
                  <span style={{ fontSize: 12, color: '#94a3b8' }}>
                    Belgedeki TCKN, IBAN, telefon, ve isimler LLM istemine gönderilmeden önce maskelenir.
                  </span>
                </div>
                <select
                  className="form-input"
                  style={{ width: 140 }}
                  value={config.policy.piiMaskingLevel}
                  onChange={(e) =>
                    setConfig({
                      ...config,
                      policy: {
                        ...config.policy,
                        piiMaskingLevel: e.target.value as any,
                      },
                    })
                  }
                >
                  <option value="strict">Strict (Tam KVKK)</option>
                  <option value="moderate">Moderate</option>
                  <option value="off">Bypass (Off)</option>
                </select>
              </div>

              {/* Policy 3: Max Loop Iterations */}
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  background: 'rgba(255, 255, 255, 0.02)',
                  padding: 14,
                  borderRadius: 10,
                  border: '1px solid rgba(255,255,255,0.06)',
                }}
              >
                <div>
                  <span style={{ fontSize: 14, fontWeight: 600, color: '#f1f5f9', display: 'block' }}>
                    Maksimum Araç Çağrısı (Guardrail Loop Limit)
                  </span>
                  <span style={{ fontSize: 12, color: '#94a3b8' }}>
                    Bir analiz görevi sırasında ajanın çalıştırabileceği azami ardışık MCP araç döngüsü.
                  </span>
                </div>
                <input
                  type="number"
                  className="form-input"
                  style={{ width: 90 }}
                  min={1}
                  max={20}
                  value={config.policy.maxToolIterations}
                  onChange={(e) =>
                    setConfig({
                      ...config,
                      policy: {
                        ...config.policy,
                        maxToolIterations: parseInt(e.target.value) || 5,
                      },
                    })
                  }
                />
              </div>

              {/* Policy 4: Audit Logging */}
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  background: 'rgba(255, 255, 255, 0.02)',
                  padding: 14,
                  borderRadius: 10,
                  border: '1px solid rgba(255,255,255,0.06)',
                }}
              >
                <div>
                  <span style={{ fontSize: 14, fontWeight: 600, color: '#f1f5f9', display: 'block' }}>
                    Değiştirilemez Denetim İzi (Audit Logging)
                  </span>
                  <span style={{ fontSize: 12, color: '#94a3b8' }}>
                    Her model çıktısı ve MCP araç çağrısı MongoDB audit koleksiyonunda şifrelenerek saklanır.
                  </span>
                </div>
                <button
                  type="button"
                  onClick={() =>
                    setConfig({
                      ...config,
                      policy: {
                        ...config.policy,
                        auditLogging: !config.policy.auditLogging,
                      },
                    })
                  }
                  style={{ background: 'none', border: 'none', cursor: 'pointer' }}
                >
                  {config.policy.auditLogging ? (
                    <ToggleRight size={28} color="#10b981" />
                  ) : (
                    <ToggleLeft size={28} color="#64748b" />
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ========================================================================= */}
      {/* SUB-TAB 4: LIVE AGENT PIPELINE SIMULATOR */}
      {/* ========================================================================= */}
      {subTab === 'simulator' && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
          <div className="admin-card">
            <div className="admin-card-header">
              <div>
                <h3 style={{ margin: 0, fontSize: 16, color: '#f8fafc' }}>
                  Canlı Multi-Agent Harness Pipeline Simülatörü
                </h3>
                <p style={{ margin: 0, fontSize: 12, color: '#94a3b8', marginTop: 2 }}>
                  Ollama çift LLM motoru, DeepWiki ve MCP araçlarının uçtan uca etkileşimini gerçek zamanlı test edin.
                </p>
              </div>

              <button
                className="btn btn-primary"
                onClick={handleRunSimulation}
                disabled={simulating}
                style={{
                  background: 'linear-gradient(135deg, #0284c7 0%, #38bdf8 100%)',
                  border: 'none',
                }}
              >
                {simulating ? <RefreshCw className="spin" size={15} /> : <Play size={15} />}
                <span>{simulating ? 'Pipeline Çalışıyor...' : t.runSimulation}</span>
              </button>
            </div>

            <div className="admin-card-body" style={{ marginTop: 14 }}>
              <label className="form-label">Girdi Sözleşme / Taahhüt Metni</label>
              <textarea
                className="form-input"
                rows={3}
                value={simSample}
                onChange={(e) => setSimSample(e.target.value)}
                style={{ fontSize: 13, marginBottom: 10, resize: 'vertical' }}
              />

              <div style={{ display: 'flex', gap: 8 }}>
                <button
                  type="button"
                  className="btn btn-secondary"
                  style={{ fontSize: 11, padding: '4px 10px' }}
                  onClick={() =>
                    setSimSample(
                      'Adobe Creative Cloud Enterprise Agreement: Term ends Nov 30, 2025. 12-month auto-renewal clause applies unless 30-day prior written non-renewal notice is delivered to renewals@adobe.com. Annual commitment: $3,600 USD.'
                    )
                  }
                >
                  Örnek 1: Adobe SaaS Yenilemesi
                </button>
                <button
                  type="button"
                  className="btn btn-secondary"
                  style={{ fontSize: 11, padding: '4px 10px' }}
                  onClick={() =>
                    setSimSample(
                      'AWS EMEA Cloud Infrastructure: Monthly minimum commitment 2,450 EUR. Net-30 invoice term, failure to object within 14 days constitutes acceptance of charges.'
                    )
                  }
                >
                  Örnek 2: AWS Hosting Faturası
                </button>
              </div>
            </div>
          </div>

          {/* Simulation Results Trace Graph */}
          {simResult && (
            <div className="admin-card" style={{ borderColor: 'rgba(56, 189, 248, 0.4)' }}>
              <div
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  marginBottom: 16,
                  borderBottom: '1px solid rgba(255,255,255,0.06)',
                  paddingBottom: 12,
                }}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <Sparkles size={18} color="#38bdf8" />
                  <h3 style={{ margin: 0, fontSize: 16, color: '#f8fafc' }}>
                    Agent Execution Trace
                  </h3>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: 14, fontSize: 12 }}>
                  <span>Toplam Süre: <strong style={{ color: '#38bdf8' }}>{simResult.totalTimeMs}ms</strong></span>
                  <span>Doğrulama Güveni: <strong style={{ color: '#10b981' }}>{(simResult.verifiedRatio * 100).toFixed(1)}%</strong></span>
                  <span>MCP Çağrısı: <strong style={{ color: '#c084fc' }}>{simResult.mcpCallsCount}</strong></span>
                </div>
              </div>

              {/* Step Timeline */}
              <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                {simResult.steps.map((step) => {
                  let stepBadgeColor = '#38bdf8'
                  if (step.component === 'pii_redactor') stepBadgeColor = '#10b981'
                  if (step.component === 'analyst_llm') stepBadgeColor = '#0284c7'
                  if (step.component === 'mcp_deepwiki') stepBadgeColor = '#c084fc'
                  if (step.component === 'verifier_llm') stepBadgeColor = '#a855f7'
                  if (step.component === 'guardrail') stepBadgeColor = '#f59e0b'

                  return (
                    <div
                      key={step.stepNumber}
                      style={{
                        display: 'flex',
                        gap: 14,
                        background: 'rgba(255, 255, 255, 0.02)',
                        padding: '12px 16px',
                        borderRadius: 8,
                        borderLeft: `4px solid ${stepBadgeColor}`,
                      }}
                    >
                      <div
                        style={{
                          minWidth: 26,
                          height: 26,
                          borderRadius: '50%',
                          background: 'rgba(255,255,255,0.06)',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          fontSize: 12,
                          fontWeight: 700,
                          color: stepBadgeColor,
                        }}
                      >
                        {step.stepNumber}
                      </div>

                      <div style={{ flex: 1 }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 2 }}>
                          <span style={{ fontSize: 13, fontWeight: 600, color: '#f8fafc' }}>
                            {step.title}
                          </span>
                          <span style={{ fontSize: 11, color: '#64748b' }}>{step.durationMs}ms</span>
                        </div>
                        <p style={{ margin: '0 0 6px 0', fontSize: 12, color: '#94a3b8' }}>
                          {step.description}
                        </p>
                        {step.outputData && (
                          <pre
                            style={{
                              margin: 0,
                              background: 'rgba(0, 0, 0, 0.35)',
                              padding: '6px 10px',
                              borderRadius: 6,
                              fontSize: 11,
                              color: '#cbd5e1',
                              overflowX: 'auto',
                            }}
                          >
                            {JSON.stringify(step.outputData, null, 2)}
                          </pre>
                        )}
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Add Custom MCP Modal */}
      {showAddMcpModal && (
        <div className="modal-backdrop" onClick={() => setShowAddMcpModal(false)}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 460 }}>
            <h3 style={{ margin: '0 0 10px 0', fontSize: 17, color: '#f8fafc' }}>
              {t.addMcpServer}
            </h3>
            <p style={{ margin: '0 0 16px 0', fontSize: 12, color: '#94a3b8' }}>
              Mevcut veya harici bir MCP (Model Context Protocol) sunucusunu sisteme kaydedin.
            </p>

            <form onSubmit={handleAddCustomMcp} style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              <div>
                <label className="form-label">MCP Sunucu Adı</label>
                <input
                  type="text"
                  className="form-input"
                  placeholder="Örn: DeepWiki Secondary, Notion MCP, Jira Tool"
                  value={newMcpName}
                  onChange={(e) => setNewMcpName(e.target.value)}
                  required
                />
              </div>

              <div>
                <label className="form-label">Endpoint URL veya SSE Stream</label>
                <input
                  type="text"
                  className="form-input"
                  placeholder="http://localhost:8899/sse"
                  value={newMcpEndpoint}
                  onChange={(e) => setNewMcpEndpoint(e.target.value)}
                  required
                />
              </div>

              <div>
                <label className="form-label">Kategori</label>
                <select
                  className="form-input"
                  value={newMcpCategory}
                  onChange={(e) => setNewMcpCategory(e.target.value as any)}
                >
                  <option value="knowledge">Knowledge / Wiki (DeepWiki vb.)</option>
                  <option value="email">Email / İhtarname</option>
                  <option value="calendar">Calendar / Takvim</option>
                  <option value="chat">Chat / Slack / Webhook</option>
                  <option value="custom">Özel Araç (Custom Tool)</option>
                </select>
              </div>

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 10, marginTop: 10 }}>
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={() => setShowAddMcpModal(false)}
                >
                  İptal
                </button>
                <button type="submit" className="btn btn-primary">
                  Kaydet ve Bağlan
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
