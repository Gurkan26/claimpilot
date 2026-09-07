# ClaimPilot — Autonomous AI Employee for Corporate & Personal Obligations

<div align="center">

![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-18.3-61DAFB?logo=react&logoColor=black)
![TypeScript](https://img.shields.io/badge/TypeScript-5.7-3178C6?logo=typescript&logoColor=white)
![Electron](https://img.shields.io/badge/Electron-34.2-47848F?logo=electron&logoColor=white)
![Ollama](https://img.shields.io/badge/AI-Ollama%20%7C%20Gemma%202%3A2B-FF6F00?logo=google&logoColor=white)
![Databases](https://img.shields.io/badge/Databases-MongoDB%20%26%20PostgreSQL-47A248?logo=mongodb&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-green.svg)

**"Sözleşme, fatura, lisans ve yükümlülükleri otonom takip eden; yenileme/iptal anında alternatif tedarikçilerden teklif toplayıp platform üzerinden işlemi bağlayan, B2B & B2C kullanıma açık, self-hosted destekli AI Agent platformu."**

[🚀 Hızlı Başlangıç](#-hızlı-başlangıç) • [🏛 Mimari & Akış](#-mimari-ve-sistem-akışı) • [🤖 Yapay Zeka & Gemma 2](#-yapay-zeka-ve-ajan-mimarisi) • [🔌 MCP Araçları](#-model-context-protocol-mcp) • [📚 LLA Gist](#-detaylı-mimari-dokümantasyonu) • [🔐 Güvenlik](#-güvenlik-ve-kvkkgdpr)

</div>

---

## 🎯 Projenin Amacı ve Hedef Kitle

Geleneksel sözleşme yönetimi Excel tablolarına ve pasif takvim hatırlatıcılarına mahkumdur. Fark edilmeyen **otomatik yenileme maddeleri (auto-renewal clauses)** ve kaçırılan **fesih ihbar süreleri (notice periods)** şirketlere ve bireylere her yıl yüz binlerce lira gereksiz maliyet çıkarmaktadır.

**ClaimPilot**, bu süreci insan müdahalesine muhtaç olmaktan çıkarıp bir **Otonom Yapay Zeka Çalışanı (Autonomous AI Employee)** haline getirir:
*   **B2B Kurumsal:** Hukuk, Satın Alma (Procurement) ve Finans ekipleri için yüzlerce SaaS aboneliğini, ofis sözleşmesini ve SLA taahhüdünü proaktif yönetir.
*   **B2C Bireysel:** Bireysel kullanıcıların spor salonu, sigorta, kasko, GSM taahhüdü ve dijital aboneliklerini takip eder, cayma bedeli risklerini sıfırlar.
*   **Gizlilik Hassasiyeti Olan Kurumlar:** Verileri buluta göndermeden, %100 yerel **Ollama + Google Gemma 2:2B** modeliyle sunucu içinde (on-premise / air-gapped) çalışır.

---

## 🏛 Mimari ve Sistem Akışı

ClaimPilot, **Clean / Hexagonal Architecture** prensipleriyle tasarlanmış bir Go monoreposu ve **Electron + React 18 + Vite** masaüstü istemcisinden oluşur.

```
+-----------------------------------------------------------------------------------+
|                         KULLANICI ARAYÜZ KATMANI                                  |
|   • Electron Desktop App (React 18, TypeScript, Tailwind/Custom CSS)             |
|   • Spark CLI (cmd/spark — Terminal İçi İnteraktif REPL Asistanı)                 |
+-----------------------------------------┬-----------------------------------------+
                                          │ HTTP / REST / GraphQL (:8080)
+-----------------------------------------▼-----------------------------------------+
|                         GATEWAY & UYGULAMA KATMANI                                |
|   • Chi Router & Middleware (CORS, JWT, Rate-Limit, Body-Limit)                  |
|   • In-Process EventBus (Gecikmesiz Olay Dağıtımı) & Real-time WebSocket Hub      |
|   • Use Cases (Document, Obligation, Marketplace, Dashboard, Admin Harness)       |
+-----------------------------------------┬-----------------------------------------+
                                          │ Orkestrasyon
+-----------------------------------------▼-----------------------------------------+
|                         AJAN VE GÜVENLİK KATMANI                                  |
|   1. [ PII Redactor ]    -> TCKN, IBAN, Telefon, Kart regex maskeleme (Yerel)     |
|   2. [ Analyst Agent ]   -> Gemma 2:2B (Port 11434): Belgeden yükümlülük çıkarımı |
|   3. [ Verifier Agent ]  -> Gemma 2:2B (Port 11435): Bağımsız çapraz denetim     |
|   4. [ Eşik Denetimi ]   -> Auto-Approve vs. Human-in-the-Loop Onay Mekanizması   |
|   5. [ RFQ Engine ]      -> Alternatif tedarikçi pazar yeri teklif toplama        |
+-----------------------------------------┬-----------------------------------------+
                                          │ Araç Çağrımı & Kalıcılık
+-----------------------------------------▼-----------------------------------------+
|                         DIŞ DÜNYA VE VERİ KATMANI                                 |
|   • Model Context Protocol (MCP) : Gmail, Google Calendar, Slack, DeepWiki        |
|   • NoSQL Depolama (MongoDB)     : Belgeler, Yükümlülükler, Teklifler, Audit Log  |
|   • SQL Depolama (PostgreSQL)    : Çok kiracılı kurumsal kullanıcılar & RBAC      |
+-----------------------------------------------------------------------------------+
```

---

## 🤖 Yapay Zeka ve Ajan Mimarisi

ClaimPilot, halüsinasyonu engellemek ve veriyi korumak için **Dual-Agent (Actor-Critic)** felsefesini kullanır:

1.  **Sıfır-Sızıntı (Zero-Leak PII):** Metin LLM'e gitmeden önce yerel regex motorundan geçer. TCKN, IBAN, kart no ve telefon bilgileri `[REDACTED]` ile maskelenir.
2.  **Analyst Ajanı (`Ollama - gemma2:2b` / Port 11434):** Belge içeriğini analiz eder, bitiş tarihini, otomatik yenileme şartını ve ceza koşullarını yapılandırılmış JSON formatında üretir.
3.  **Verifier Ajanı (`Ollama - gemma2:2b` / Port 11435):** Analyst'in çıktısını belgenin orijinal haliyle satır satır karşılaştırır. Sayfa ve madde referanslarını teyit eder, bağımsız risk skoru (`LOW`, `MEDIUM`, `HIGH`, `CRITICAL`) atar.
4.  **Kontrollü Otonomi:** Risk düşükse ve model güveni tam ise işlem onaylanır (`Auto-Approve`). Risk yüksekse insan onayına (`Pending Approval`) sunulur.

---

## 🔌 Model Context Protocol (MCP)

ClaimPilot, sadece ekrana metin yazdıran bir araç değildir. Anthropic tarafından standartlaştırılan **Model Context Protocol (MCP)** adaptörleriyle doğrudan aksiyon alır:

*   **📧 Gmail MCP:** Fesih ihbarı veya fiyat itirazı için şirket adına e-posta taslağı oluşturur ve gönderir.
*   **📅 Google Calendar MCP:** Kritik fesih bildirim günlerini Google Takvim'e otomatik etkinlik olarak kaydeder.
*   **💬 Slack MCP:** Yaklaşan vadeleri ve onay bekleyen yükümlülükleri ilgili satın alma kanalına fırlatır.
*   **📖 DeepWiki MCP:** Kurumsal sözleşme standartları ve şirket bilgi tabanını SSE üzerinden bağlar.

---

## 🚀 Hızlı Başlangıç

### 1. Ön Koşullar

*   **Go 1.26+**
*   **Node.js 20+** & **npm**
*   **Docker & Docker Compose**
*   **Ollama** (`ollama run gemma2:2b`)

### 2. Altyapı ve Veritabanlarını Başlatma

```bash
# Docker servislerini ayağa kaldırın (MongoDB, PostgreSQL, Redis)
cd backend
docker compose -f deployments/docker-compose.yml up -d
```

### 3. Backend API Gateway & Daemon'ları Çalıştırma

```bash
cd backend

# 1. Ana REST & GraphQL API Gateway (:8080)
go run ./cmd/api

# 2. (İsteğe bağlı) 7/24 Arka Plan Vade Tarama Daemon'u
go run ./cmd/worker

# 3. (İsteğe bağlı) Terminal İçi Spark CLI Asistanı
go run ./cmd/spark
```

*   **REST API:** `http://localhost:8080/api/v1`
*   **GraphQL Playground:** `http://localhost:8080/graphql`
*   **Health Check:** `http://localhost:8080/health`

### 4. Masaüstü Uygulamasını Başlatma (Electron + React)

```bash
cd frontend

# Bağımlılıkları yükleyin
npm install

# Geliştirme modunda başlatın (Vite + Electron)
npm run electron:dev

# macOS Uygulama Paketi (.dmg / .app)
npm run package:mac

# Windows Uygulama Paketi (.exe / NSIS)
npm run package:win
```

---

## 🧪 Test Stratejisi ve Kalite Güvence

Projede çok katmanlı test piramidi uygulanmıştır:

```bash
# Backend birim ve entegrasyon testlerini çalıştırın
cd backend
go test -v ./internal/...

# PII Maskeleme testlerini çalıştırın
go test -v ./internal/pii/...

# Orkestrasyon ve Ajan testlerini çalıştırın
go test -v ./internal/agent/...

# Coverage raporu alın
./scripts/test.sh -cover
```

*   **Unit Tests:** Domain modelleri, DTO validasyonları ve PII regex desenleri `testify` ile %100 kapsamda test edilir.
*   **Mock Repositories:** Harici MongoDB ve LLM servislerine bağımlı olmadan izole orkestrasyon testleri (`mockLLM`, `mockDocRepo`).
*   **Admin Simulation Harness:** Canlı sözleşme metinleri üzerinde PII maskeleme, Analyst çıkarımı ve Verifier onay oranlarını görsel adımlarla test eden uçtan uca simülatör.
*   **Postman Collection:** 37 otomatik senaryo içeren Postman test paketi (`postman/masterfabric-go.postman_collection.json`).

---

## ⚡ Performans ve Optimizasyonlar

*   **Go Eşzamanlılığı (Goroutines):** Doküman okuma, analiz ve MCP bildirimleri ana HTTP iş parçacığını kitlemeden asenkron yürütülür.
*   **Ollama Yük Paylaşımı:** Analyst ve Verifier iki ayrı Ollama portuna (`11434` & `11435`) bölünerek CPU/çıkarım darboğazı önlenmiştir.
*   **Streaming Yanıtlar (SSE / Channels):** Ajan yanıtları token akış kanalları üzerinden arayüze anında aktarılır (Düşük TTFT).
*   **MongoDB İndeksleme & Projeksiyon:** `due_date`, `user_id`, `status` indeksleri ve gereksiz ham doküman metinlerini çekmeyen seçici projeksiyonlar.
*   **Sıfır-Tahsisli (Pre-Compiled) Regex:** PII desenleri tek seferde derlenerek bellek tahsisatı (heap allocation) minimumda tutulur.
*   **Vite Tree-Shaking & Minification:** Masaüstü bundle boyutu optimize edilmiş, lazy-loading ile açılış süresi kısaltılmıştır.

---

## 🔐 Güvenlik ve KVKK/GDPR Uyumu

*   **Sıfır-Sızıntı (Zero-Data-Leak) Kuralı:** TCKN, IBAN, kredi kartı numaraları, telefon ve e-postalar yerel regex filtresi tarafından maskelenir. Harici LLM veya ağ soketlerine hiçbir zaman ham kişisel veri aktarılmaz.
*   **Değişmez Denetim İzi (Audit Log):** Her ajan kararının, güven puanının, model adının (`gemma2:2b`) ve MCP araç çalıştırmasının silinemez kaydı MongoDB `audit_logs` koleksiyonuna yazılır.

---

## 📚 Detaylı Mimari Dokümantasyonu

Projenin tüm sınıf diyagramlarını, domain modellerini, sequence akışlarını ve veri şemalarını içeren **Low Level Architecture (LLA)** dokümanına aşağıdaki bağlantılardan ulaşabilirsiniz:

*   📄 **Yerel Doküman:** [`LOW_LEVEL_ARCHITECTURE.md`](LOW_LEVEL_ARCHITECTURE.md)
*   🌐 **GitHub Public Gist:** [ClaimPilot Low Level Architecture Gist Linki](https://gist.github.com/Gurkan26/ce4fefc3a3858bb5b79dff153315f867)

---

## 👥 Katkıda Bulunma ve Lisans

Bu proje **MIT Lisansı** ile lisanslanmıştır. Katkıda bulunmak için lütfen bir issue açın veya PR gönderin.
