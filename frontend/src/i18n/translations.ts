export type Language = 'tr' | 'en'

export interface TranslationDict {
  // Navigation & Shell
  appTitle: string
  dashboard: string
  obligations: string
  marketplace: string
  documents: string
  auditTrail: string
  settings: string
  connected: string
  demoMode: string
  login: string
  logout: string
  switchAccount: string
  corporate: string
  personal: string
  corporateBadge: string
  personalBadge: string
  activeAiEmployee: string
  mcpReady: string
  signOut: string

  // Login Screen
  loginTitle: string
  loginSubtitle: string
  loginWelcome: string
  loginFeature1: string
  loginFeature2: string
  loginFeature3: string
  loginFeature4: string
  loginQuickAccess: string
  loginOrCustom: string

  // Dashboard
  greetingMorning: string
  dashboardSubtitle: string
  uploadDocButton: string
  aiBriefingTitle: string
  totalObligations: string
  dueNext30Days: string
  pendingApprovals: string
  requiresReview: string
  overdueItems: string
  urgentAction: string
  activeRfqs: string
  quotesReady: string
  financialImpact: string
  transactedGmv: string
  realizedSavings: string
  commissionFee: string
  riskDistribution: string
  criticalDeadlines: string
  viewAll: string
  approve: string
  approved: string
  dismiss: string
  daysLeft: string

  // Obligations
  obligationsTitle: string
  obligationsSubtitle: string
  searchPlaceholder: string
  all: string
  dueInDays: string
  obligationCol: string
  categoryCol: string
  dueDateCol: string
  amountCol: string
  riskCol: string
  statusCol: string
  actionCol: string
  actionsCol: string
  noObligations: string
  dismissPrompt: string
  dismissReason: string
  enterReason: string
  confirm: string
  cancel: string

  // Marketplace
  marketplaceTitle: string
  marketplaceSubtitle: string
  collectQuotes: string
  savePercent: string
  acceptDeal: string
  dealClosed: string
  currentCost: string
  riskScore: string
  noDeals: string

  // Documents
  documentsTitle: string
  documentsSubtitle: string
  dragDropTitle: string
  dragDropSubtitle: string
  browseFiles: string
  analyzingWithAi: string
  fileNameCol: string
  typeCol: string
  sizeCol: string
  summaryCol: string
  dateCol: string

  // Audit
  auditTitle: string
  auditSubtitle: string
  eventCol: string
  approvalTypeCol: string
  toolCol: string
  detailsCol: string
  timeCol: string

  // Auth & Onboarding Modal
  authModalTitle: string
  authModalSubtitle: string
  selectAccountType: string
  corporateDesc: string
  personalDesc: string
  fullName: string
  companyName: string
  department: string
  email: string
  quickDemoLoginCorporate: string
  quickDemoLoginPersonal: string
  continueButton: string

  // Toast Notifications
  toastApproved: string
  toastDismissed: string
  toastRfqCollected: string
  toastDealClosed: string
  toastUploadSuccess: string
  toastError: string

  // Admin & Agent Harness
  adminPanel: string
  adminBadge: string
  adminLoginTitle: string
  adminPasswordLabel: string
  adminPasswordPlaceholder: string
  adminLoginButton: string
  adminWrongPassword: string
  adminWelcome: string
  adminSubtitle: string
  tabLlmConfig: string
  tabMcpHarness: string
  tabGuardrails: string
  tabSimulator: string
  tabAdminProfile: string
  adminProfileTitle: string
  adminProfileSubtitle: string
  adminSaveProfile: string
  adminProfileSaved: string
  adminPasswordSectionTitle: string
  adminPasswordSectionSubtitle: string
  adminCurrentPassword: string
  adminNewPassword: string
  adminConfirmPassword: string
  adminChangePasswordBtn: string
  adminPasswordMismatch: string
  adminPasswordChanged: string
  adminAccessDenied: string
  adminAccessDeniedDesc: string
  editAdminProfile: string
  analystRoleTitle: string
  verifierRoleTitle: string
  llmProvider: string
  llmEndpoint: string
  llmModel: string
  llmApiKey: string
  llmTemperature: string
  llmMaxTokens: string
  testConnection: string
  applyChanges: string
  hotSwapSuccess: string
  deepWikiDesc: string
  addMcpServer: string
  runSimulation: string
  humanApprovalToggle: string
  piiStrictness: string
}

export const translations: Record<Language, TranslationDict> = {
  tr: {
    // Navigation & Shell
    appTitle: 'ClaimPilot Masaüstü',
    dashboard: 'Yönetici Paneli',
    obligations: 'Yükümlülükler & Vadeler',
    marketplace: 'Pazar Yeri & RFQ',
    documents: 'Belge Alımı',
    auditTrail: 'Denetim İzi (Audit)',
    settings: 'Ayarlar',
    connected: 'Sistem Canlı (API)',
    demoMode: 'Masaüstü Modu',
    login: 'Giriş Yap',
    logout: 'Çıkış Yap',
    switchAccount: 'Hesap / Mod Değiştir',
    corporate: 'Kurumsal (B2B)',
    personal: 'Bireysel (B2C)',
    corporateBadge: 'KURUMSAL',
    personalBadge: 'BİREYSEL',
    activeAiEmployee: 'Otonom AI Çalışanı',
    mcpReady: 'MCP Aktif & Hazır',
    signOut: 'Oturumu Kapat',

    // Login Screen
    loginTitle: 'ClaimPilot',
    loginSubtitle: 'Otonom AI Çalışanınız — Sözleşme, yükümlülük ve tedarikçi yönetimi tek ekranda.',
    loginWelcome: 'Hoş Geldiniz',
    loginFeature1: 'Sözleşme ve fatura otomatik analizi',
    loginFeature2: 'Kritik vade ve yükümlülük takibi',
    loginFeature3: 'Tedarikçi pazar yeri ve otomatik RFQ',
    loginFeature4: 'KVKK/GDPR uyumlu denetim izi',
    loginQuickAccess: 'Hızlı Demo Erişimi',
    loginOrCustom: 'veya kendi bilgilerinizle giriş yapın',

    // Dashboard
    greetingMorning: 'Günaydın',
    dashboardSubtitle: 'Sözleşme, yükümlülük ve pazar yeri fırsatlarının otonom durum özeti.',
    uploadDocButton: '+ Sözleşme / Fatura Yükle',
    aiBriefingTitle: 'Günlük Otonom Yönetici Brifingi',
    totalObligations: 'Takip Edilen Yükümlülük',
    dueNext30Days: 'önümüzdeki 30 günde vadeli',
    pendingApprovals: 'Onayınızı Bekleyenler',
    requiresReview: 'Tek tıkla onay gerektirir',
    overdueItems: 'Gecikmiş / Kritik Vadeler',
    urgentAction: 'Acil aksiyon planlandı',
    activeRfqs: 'Aktif Tedarikçi Teklifleri',
    quotesReady: 'Alternatif teklifler hazır',
    financialImpact: 'Finansal Etki & Ticaret Hacmi (GMV)',
    transactedGmv: 'İşlem Hacmi (GMV)',
    realizedSavings: 'Gerçekleşen Net Tasarruf',
    commissionFee: '%4 platform başarı komisyonu',
    riskDistribution: 'Risk Seviyesi Dağılımı',
    criticalDeadlines: 'Yaklaşan Kritik Vadeler ve Aksiyonlar',
    viewAll: 'Tümünü Gör',
    approve: 'Onayla',
    approved: 'Onaylandı',
    dismiss: 'Reddet',
    daysLeft: 'gün kaldı',

    // Obligations
    obligationsTitle: 'Yükümlülükler ve Kritik Tarihler',
    obligationsSubtitle: 'Otomatik fesih, bildirim, yenileme ve ceza maddelerinin otonom takibi.',
    searchPlaceholder: 'Yükümlülük, sözleşme veya sağlayıcı ara...',
    all: 'Tümü',
    dueInDays: 'Gün Vade',
    obligationCol: 'YÜKÜMLÜLÜK / SAĞLAYICI',
    categoryCol: 'KATEGORİ',
    dueDateCol: 'SON TARİH',
    amountCol: 'TUTAR',
    riskCol: 'RİSK',
    statusCol: 'DURUM',
    actionCol: 'OTONOM AKSİYON',
    actionsCol: 'İŞLEM',
    noObligations: 'Kriterlere uygun yükümlülük bulunamadı.',
    dismissPrompt: 'Reddetme nedeni belirtin:',
    dismissReason: 'Reddetme Nedeni',
    enterReason: 'Neden belirtiniz...',
    confirm: 'Tamam',
    cancel: 'İptal',

    // Marketplace
    marketplaceTitle: 'Tedarikçi Pazar Yeri & Otomatik Teklif Toplama (RFQ)',
    marketplaceSubtitle: 'Yenileme ve fiyat artışlarında pasif kalmayın; alternatif tedarikçilerden otomatik teklif toplayıp tasarrufla geçin.',
    collectQuotes: 'Teklif Topla (RFQ)',
    savePercent: 'Tasarruf',
    acceptDeal: 'Anlaşmayı Bağla & Geç',
    dealClosed: 'Anlaşma Başarıyla Bağlandı!',
    currentCost: 'Mevcut Yıllık Maliyet',
    riskScore: 'Verifier Güven Skoru',
    noDeals: 'Henüz aktif bir teklif fırsatı bulunmuyor.',

    // Documents
    documentsTitle: 'Belge Alımı & Yapay Zeka Analizi',
    documentsSubtitle: 'Çok formatlı sözleşme/fatura okuma ve sıfır-sızıntı KVKK/GDPR PII maskeleme.',
    dragDropTitle: 'Sözleşme, fatura veya poliçeleri buraya sürükleyip bırakın',
    dragDropSubtitle: 'PDF, TXT, CSV, MD desteklenir. TCKN, IBAN, telefon ve isimler LLM öncesi maskelenir.',
    browseFiles: 'Bilgisayardan Dosya Seç...',
    analyzingWithAi: 'Analyst ve Verifier ajanları belgeyi inceliyor...',
    fileNameCol: 'DOSYA ADI',
    typeCol: 'TÜR',
    sizeCol: 'BOYUT',
    summaryCol: 'AI ÖZETİ',
    dateCol: 'YÜKLENME',

    // Audit
    auditTitle: 'Denetim İzi & KVKK Uyumluluk Kaydı',
    auditSubtitle: 'Tüm otonom ajan kararları, kullanıcı onayları ve MCP araç çalıştırma kayıtları.',
    eventCol: 'OLAY / EYLEM',
    approvalTypeCol: 'ONAY TÜRÜ',
    toolCol: 'ARAÇ / ADAPTÖR',
    detailsCol: 'AYRINTILAR',
    timeCol: 'ZAMAN',

    // Auth & Onboarding Modal
    authModalTitle: 'ClaimPilot Hesabı Seçin / Giriş Yapın',
    authModalSubtitle: 'ClaimPilot hem şirketler (B2B) hem şahıslar (B2C) için otonom yükümlülük takibi yapar.',
    selectAccountType: 'Hesap Türü:',
    corporateDesc: 'Şirket sözleşmeleri, SaaS lisansları, bulut faturaları ve tedarikçi pazarlığı.',
    personalDesc: 'Kira sözleşmesi, araç kaskosu, internet/fiber ve kişisel abonelikler.',
    fullName: 'Ad Soyad',
    companyName: 'Şirket / Organizasyon Adı',
    department: 'Departman (Satın Alma, Hukuk, BT)',
    email: 'E-posta Adresi',
    quickDemoLoginCorporate: '🏢 Kurumsal Giriş Yap (Acme Holding A.Ş.)',
    quickDemoLoginPersonal: '👤 Bireysel Giriş Yap (Gürkan Şentürk)',
    continueButton: 'Giriş Yap ve Başla',

    // Toast Notifications
    toastApproved: 'Yükümlülük başarıyla onaylandı ve MCP eylemi tetiklendi.',
    toastDismissed: 'Yükümlülük reddedildi.',
    toastRfqCollected: 'Tedarikçi teklifleri başarıyla toplandı!',
    toastDealClosed: 'Anlaşma başarıyla bağlandı! Geçiş süreci başlatıldı.',
    toastUploadSuccess: 'Belge başarıyla yüklendi ve analiz ediliyor.',
    toastError: 'İşlem sırasında bir hata oluştu.',

    // Admin & Agent Harness
    adminPanel: 'Agent Harness & Yönetici',
    adminBadge: 'ADMİN',
    adminLoginTitle: 'Agent Harness Yönetici Girişi',
    adminPasswordLabel: 'Yönetici Şifresi',
    adminPasswordPlaceholder: 'Şifrenizi girin (varsayılan: admin123)',
    adminLoginButton: 'Yönetici Konsoluna Giriş',
    adminWrongPassword: 'Hatalı şifre! (Varsayılan: admin123)',
    adminWelcome: 'Agent Harness & LLM Orkestrasyonu',
    adminSubtitle: 'Ollama üzerinde çift-rol LLM motoru, DeepWiki ve harici MCP araç bağlantıları yönetimi.',
    tabLlmConfig: 'Çift-LLM Motoru (Ollama)',
    tabMcpHarness: 'MCP Bağlantıları & DeepWiki',
    tabGuardrails: 'Harness İlkeleri & Güvenlik',
    tabSimulator: 'Canlı Agent Pipeline Simülatörü',
    tabAdminProfile: 'Yönetici Bilgileri & Güvenlik',
    adminProfileTitle: 'Yönetici Profil Bilgileri',
    adminProfileSubtitle: 'Sistem yöneticisi kimlik, kurum ve unvan bilgilerini güncelleyin.',
    adminSaveProfile: 'Profil Bilgilerini Güncelle',
    adminProfileSaved: 'Yönetici profil bilgileri başarıyla güncellendi.',
    adminPasswordSectionTitle: 'Yönetici Şifresini Değiştir',
    adminPasswordSectionSubtitle: 'Agent Harness konsolu yönetici giriş şifresini güncelleyin.',
    adminCurrentPassword: 'Mevcut Yönetici Şifresi',
    adminNewPassword: 'Yeni Şifre',
    adminConfirmPassword: 'Yeni Şifre (Tekrar)',
    adminChangePasswordBtn: 'Şifreyi Güncelle',
    adminPasswordMismatch: 'Yeni şifreler birbiriyle eşleşmiyor!',
    adminPasswordChanged: 'Yönetici şifresi başarıyla güncellendi.',
    adminAccessDenied: 'Yetkisiz Erişim Engellendi',
    adminAccessDeniedDesc: 'Bu panel sadece sistem yöneticilerine açıktır. Lütfen yönetici hesabınızla giriş yapınız.',
    editAdminProfile: 'Yönetici Bilgilerini Düzenle',
    analystRoleTitle: 'Rol 1: Analist Ajanı (Analyst Agent)',
    verifierRoleTitle: 'Rol 2: Doğrulayıcı Ajanı (Verifier Agent)',
    llmProvider: 'LLM Sağlayıcı',
    llmEndpoint: 'Endpoint URL',
    llmModel: 'Model Adı',
    llmApiKey: 'API Anahtarı (Opsiyonel)',
    llmTemperature: 'Sıcaklık (Temperature)',
    llmMaxTokens: 'Maksimum Token',
    testConnection: 'Bağlantıyı Test Et',
    applyChanges: 'Değişiklikleri Kaydet & Hot-Swap',
    hotSwapSuccess: 'LLM ve Harness ayarları başarıyla güncellendi ve uygulandı!',
    deepWikiDesc: 'Kurumsal bilgi tabanı, mevzuat, sözleşme emsalleri ve şirket içi yönerge arama aracı.',
    addMcpServer: '+ Yeni MCP Sunucusu Ekle',
    runSimulation: 'Harness Pipeline Simülasyonu Çalıştır',
    humanApprovalToggle: 'MCP Eylemlerinde İnsan Onayı Zorunlu (Human-in-the-Loop)',
    piiStrictness: 'KVKK/GDPR PII Maskeleme Düzeyi',
  },
  en: {
    // Navigation & Shell
    appTitle: 'ClaimPilot Desktop',
    dashboard: 'Executive Dashboard',
    obligations: 'Obligations & Deadlines',
    marketplace: 'Marketplace & RFQ',
    documents: 'Document Intake',
    auditTrail: 'Compliance Audit Trail',
    settings: 'Settings',
    connected: 'API Live',
    demoMode: 'Desktop Mode',
    login: 'Sign In',
    logout: 'Sign Out',
    switchAccount: 'Switch Mode / Account',
    corporate: 'Corporate (B2B)',
    personal: 'Personal (B2C)',
    corporateBadge: 'ENTERPRISE',
    personalBadge: 'PERSONAL',
    activeAiEmployee: 'Autonomous AI Employee',
    mcpReady: 'MCP Tools Active',
    signOut: 'Sign Out',

    // Login Screen
    loginTitle: 'ClaimPilot',
    loginSubtitle: 'Your Autonomous AI Employee — Contract, obligation, and supplier management in one place.',
    loginWelcome: 'Welcome',
    loginFeature1: 'Automatic contract & invoice analysis',
    loginFeature2: 'Critical deadline & obligation tracking',
    loginFeature3: 'Supplier marketplace & automated RFQ',
    loginFeature4: 'KVKK/GDPR compliant audit trail',
    loginQuickAccess: 'Quick Demo Access',
    loginOrCustom: 'or sign in with your credentials',

    // Dashboard
    greetingMorning: 'Good morning',
    dashboardSubtitle: 'Autonomous tracking of obligations, contractual deadlines, and procurement deals.',
    uploadDocButton: '+ Upload Contract / Invoice',
    aiBriefingTitle: 'Daily Autonomous Executive Briefing',
    totalObligations: 'Tracked Obligations',
    dueNext30Days: 'due within next 30 days',
    pendingApprovals: 'Pending Your Approval',
    requiresReview: 'One-click decision required',
    overdueItems: 'Overdue & Urgent Items',
    urgentAction: 'Immediate action scheduled',
    activeRfqs: 'Active Supplier Quotes',
    quotesReady: 'Alternative bids available',
    financialImpact: 'Financial Impact & Deal Volume (GMV)',
    transactedGmv: 'Transacted Volume (GMV)',
    realizedSavings: 'Realized Net Savings',
    commissionFee: '4% platform success fee',
    riskDistribution: 'Obligation Risk Distribution',
    criticalDeadlines: 'Critical Deadlines & Autonomous Actions',
    viewAll: 'View All',
    approve: 'Approve',
    approved: 'Approved',
    dismiss: 'Dismiss',
    daysLeft: 'days left',

    // Obligations
    obligationsTitle: 'Obligations & Critical Deadlines',
    obligationsSubtitle: 'Autonomous monitoring of renewal clauses, penalties, and contractual terms.',
    searchPlaceholder: 'Search obligations, vendors, or clauses...',
    all: 'All',
    dueInDays: 'Days Ahead',
    obligationCol: 'OBLIGATION / VENDOR',
    categoryCol: 'CATEGORY',
    dueDateCol: 'DUE DATE',
    amountCol: 'AMOUNT',
    riskCol: 'RISK',
    statusCol: 'STATUS',
    actionCol: 'AUTONOMOUS ACTION',
    actionsCol: 'ACTION',
    noObligations: 'No obligations found matching your criteria.',
    dismissPrompt: 'Specify reason for dismissal:',
    dismissReason: 'Dismissal Reason',
    enterReason: 'Enter reason...',
    confirm: 'Confirm',
    cancel: 'Cancel',

    // Marketplace
    marketplaceTitle: 'Procurement Marketplace & Automated RFQ',
    marketplaceSubtitle: 'Never renew passively; automatically collect quotes from alternative suppliers and switch with net savings.',
    collectQuotes: 'Collect Quotes (RFQ)',
    savePercent: 'Save',
    acceptDeal: 'Close Deal & Switch',
    dealClosed: 'Deal Successfully Closed!',
    currentCost: 'Current Annual Cost',
    riskScore: 'Verifier Trust Score',
    noDeals: 'No active quote opportunities at the moment.',

    // Documents
    documentsTitle: 'Document Intake & AI Analysis',
    documentsSubtitle: 'Multi-format contract/invoice extraction with zero-retention KVKK/GDPR PII masking.',
    dragDropTitle: 'Drag and drop contracts, invoices, or policies here',
    dragDropSubtitle: 'Supports PDF, TXT, CSV, MD. TCKN, IBAN, phones and names redacted before LLM.',
    browseFiles: 'Browse from Computer...',
    analyzingWithAi: 'Analyst and Verifier agents analyzing document...',
    fileNameCol: 'FILE NAME',
    typeCol: 'TYPE',
    sizeCol: 'SIZE',
    summaryCol: 'AI SUMMARY',
    dateCol: 'INGESTED',

    // Audit
    auditTitle: 'Compliance & Agent Audit Trail',
    auditSubtitle: 'Immutable log of all autonomous agent decisions, approvals, and MCP tool dispatches.',
    eventCol: 'EVENT / ACTION',
    approvalTypeCol: 'APPROVAL TYPE',
    toolCol: 'TOOL / ADAPTER',
    detailsCol: 'DETAILS',
    timeCol: 'TIMESTAMP',

    // Auth & Onboarding Modal
    authModalTitle: 'Select Account / Sign In to ClaimPilot',
    authModalSubtitle: 'ClaimPilot operates autonomously for both corporate enterprises (B2B) and individuals (B2C).',
    selectAccountType: 'Account Type:',
    corporateDesc: 'Corporate MSAs, SaaS tools, cloud infrastructure, and procurement bargaining.',
    personalDesc: 'Lease agreements, car insurance, home fiber, and personal subscriptions.',
    fullName: 'Full Name',
    companyName: 'Company / Organization Name',
    department: 'Department (Procurement, Legal, IT)',
    email: 'Email Address',
    quickDemoLoginCorporate: '🏢 Sign In as Corporate (Acme Holding A.Ş.)',
    quickDemoLoginPersonal: '👤 Sign In as Personal (Gürkan Şentürk)',
    continueButton: 'Sign In & Enter Dashboard',

    // Toast Notifications
    toastApproved: 'Obligation approved and MCP action triggered successfully.',
    toastDismissed: 'Obligation dismissed.',
    toastRfqCollected: 'Supplier quotes collected successfully!',
    toastDealClosed: 'Deal closed successfully! Vendor switch initiated.',
    toastUploadSuccess: 'Document uploaded and analysis in progress.',
    toastError: 'An error occurred during the operation.',

    // Admin & Agent Harness
    adminPanel: 'Agent Harness & Admin',
    adminBadge: 'ADMIN',
    adminLoginTitle: 'Agent Harness Admin Login',
    adminPasswordLabel: 'Admin Password',
    adminPasswordPlaceholder: 'Enter password (default: admin123)',
    adminLoginButton: 'Enter Admin Console',
    adminWrongPassword: 'Invalid password! (Default: admin123)',
    adminWelcome: 'Agent Harness & LLM Orchestration',
    adminSubtitle: 'Manage dual-role Ollama LLMs, DeepWiki knowledge base, and external MCP tool adapters.',
    tabLlmConfig: 'Dual-LLM Engine (Ollama)',
    tabMcpHarness: 'MCP Adapters & DeepWiki',
    tabGuardrails: 'Harness Guardrails & Policies',
    tabSimulator: 'Live Agent Pipeline Simulator',
    tabAdminProfile: 'Admin Profile & Security',
    adminProfileTitle: 'Administrator Profile',
    adminProfileSubtitle: 'Update system administrator identity, organization, and role details.',
    adminSaveProfile: 'Update Profile Details',
    adminProfileSaved: 'Administrator profile successfully updated.',
    adminPasswordSectionTitle: 'Change Admin Password',
    adminPasswordSectionSubtitle: 'Update administrator console login password.',
    adminCurrentPassword: 'Current Admin Password',
    adminNewPassword: 'New Password',
    adminConfirmPassword: 'Confirm New Password',
    adminChangePasswordBtn: 'Update Password',
    adminPasswordMismatch: 'New passwords do not match!',
    adminPasswordChanged: 'Administrator password successfully updated.',
    adminAccessDenied: 'Access Denied',
    adminAccessDeniedDesc: 'This panel is restricted to system administrators only. Please log in with administrator credentials.',
    editAdminProfile: 'Edit Admin Profile',
    analystRoleTitle: 'Role 1: Analyst Agent',
    verifierRoleTitle: 'Role 2: Verifier Agent',
    llmProvider: 'LLM Provider',
    llmEndpoint: 'Endpoint URL',
    llmModel: 'Model Name',
    llmApiKey: 'API Key (Optional)',
    llmTemperature: 'Temperature',
    llmMaxTokens: 'Max Tokens',
    testConnection: 'Test Connection',
    applyChanges: 'Save & Hot-Swap',
    hotSwapSuccess: 'LLM & Harness configuration updated and hot-swapped!',
    deepWikiDesc: 'Enterprise knowledge base, compliance documentation, and contract precedent retrieval tool.',
    addMcpServer: '+ Add New MCP Server',
    runSimulation: 'Execute Harness Pipeline',
    humanApprovalToggle: 'Require Human Approval for MCP Actions (Human-in-the-Loop)',
    piiStrictness: 'KVKK/GDPR PII Masking Strictness',
  },
}
