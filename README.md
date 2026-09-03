# ClaimPilot — Autonomous AI Employee for Corporate & Personal Obligations

> **"AI Employee for Corporate & Personal Obligations"** — sözleşme, fatura, lisans ve yükümlülükleri otonom takip eden; yenileme/iptal anında alternatif tedarikçilerden teklif toplayıp platform üzerinden işlemi bağlayan, hem B2C hem B2B kullanıma açık, self-hosted destekli AI agent platformu.

---

## 🏛 Mimari ve Proje Yapısı

ClaimPilot iki ana katmandan ve tam teşekküllü bir monorepodan oluşur:

```
claimpilot/
├── backend/                   # Go Clean Architecture Monorepo
│   ├── cmd/
│   │   ├── api/               # REST & GraphQL API Gateway (:8080)
│   │   ├── spark/             # Terminal REPL Interactive AI Assistant
│   │   ├── worker/            # 7/24 Proaktif Vade Tarama Daemon'u
│   │   └── server/            # Orijinal PostgreSQL Servisi
│   └── internal/
│       ├── agent/             # Multi-Agent Motoru (Analyst, Verifier, RFQ)
│       ├── application/       # Dashboard, Documents, Obligations, Marketplace Use Cases
│       ├── domain/            # Core Domain Modelleri ve Repository Arayüzleri
│       ├── infrastructure/    # MongoDB, HTTP/GraphQL Handler'lar, Local Storage
│       ├── mcp/               # Gmail, Google Calendar, Slack Adaptörleri
│       └── pii/               # Sıfır-Sızıntı KVKK/GDPR Regex Maskeleme Motoru
│
└── frontend/                  # Çapraz Platform Masaüstü Uygulaması (macOS & Windows)
    ├── electron/              # Electron Main Process & Preload Context Bridge
    ├── src/
    │   ├── components/        # Dashboard, Obligations, Marketplace, Documents, Audit
    │   ├── context/           # AuthContext (B2B Kurumsal / B2C Bireysel Modları)
    │   ├── i18n/              # Çift Dil Desteği (Türkçe & İngilizce)
    │   └── services/          # Go API İstemcisi & Çevrimdışı Masaüstü Modu
    └── package.json           # Electron-builder paketleme yapılandırması
```

---

## 🚀 Hızlı Başlangıç

### 1. Backend Servisini Başlatma (Go)

```bash
cd backend
go run ./cmd/api
```

- REST API Gateway: `http://localhost:8080/api/v1`
- GraphQL Playground: `http://localhost:8080/graphql`
- Health Check: `http://localhost:8080/health`

### 2. Masaüstü Uygulamasını Başlatma (Electron + React)

```bash
cd frontend

# Bağımlılıkları yükleyin (ilk seferde)
npm install

# Geliştirme modunda başlatın
npm run electron:dev

# macOS Uygulama Paketini Derleyin (.dmg & .app)
npm run package:mac

# Windows Uygulama Paketini Derleyin (.exe & NSIS)
npm run package:win
```

### 3. Terminal CLI (Spark) Çalıştırma

```bash
cd backend
go run ./cmd/spark
```

---

## 🔐 Güvenlik ve Gizlilik (KVKK & GDPR)

- **Sıfır-Sızıntı PII Redaksiyonu**: TCKN, IBAN, kredi kartı, telefon numarası ve e-posta adresleri hiçbir zaman harici LLM sağlayıcılarına gönderilmez; belge ayrıştırılırken yerel regex motoruyla maskelenir (`[TCKN_REDACTED]`, `[IBAN_REDACTED]`).
- **Değişmez Denetim İzi (Audit Log)**: Tüm otonom kararlar ve MCP araç çalıştırmaları zaman damgasıyla yerel MongoDB üzerinde kayıt altına alınır.
