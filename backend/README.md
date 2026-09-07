# ClaimPilot Backend Engine — Autonomous AI & Obligation Gateway

<div align="center">

![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)
![Chi Router](https://img.shields.io/badge/Router-Chi%20v5-000000?logo=go&logoColor=white)
![MongoDB](https://img.shields.io/badge/Storage-MongoDB%20v2-47A248?logo=mongodb&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/Storage-PostgreSQL%2016-336791?logo=postgresql&logoColor=white)
![Ollama](https://img.shields.io/badge/AI-Ollama%20%7C%20Gemma%202%3A2B-FF6F00?logo=google&logoColor=white)
![MCP](https://img.shields.io/badge/Protocol-Model%20Context%20Protocol-8A2BE2)
![Architecture](https://img.shields.io/badge/Architecture-Clean%20%2F%20Hexagonal%20DDD-brightgreen)

**Kurumsal ve bireysel sözleşme, fatura, taahhüt ve lisansların yaşam döngüsünü otonom yöneten; Dual-Agent (Gemma 2:2B) mimarisi ve Model Context Protocol (MCP) ile donatılmış yüksek performanslı Go backend motoru.**

[🏛 Mimari](#-mimari-ve-katmanlar) • [🚀 Başlangıç](#-kurulum-ve-hızlı-başlangıç) • [🤖 Ajan Motoru](#-multi-agent--orkestrasyon-motoru) • [📡 API Uç Noktaları](#-api-ve-uç-noktalar) • [🧪 Testler](#-test-ve-kalite-güvence) • [🔐 Güvenlik](#-güvenlik-ve-pii-kalkanı)

</div>

---

## 🏛 Mimari ve Katmanlar (Clean / Hexagonal Architecture)

Backend, **Domain-Driven Design (DDD)** ve **Hexagonal (Port & Adapter) Mimari** prensipleriyle geliştirilmiştir. Domain katmanı saf Go ile yazılmış olup hiçbir dış kütüphane veya veritabanı bağımlılığı barındırmaz.

```
backend/
├── cmd/
│   ├── api/                   # REST & GraphQL API Gateway Sunucusu (:8080)
│   ├── worker/                # 7/24 Proaktif Vade Tarama ve Tetikleme Daemon'u
│   ├── spark/                 # Terminal İçi İnteraktif CLI Asistanı (REPL)
│   └── server/                # İlişkisel Çok Kiracılı PostgreSQL Servisi
├── internal/
│   ├── domain/                # Saf Domain Modelleri ve Repository Arayüzleri
│   │   ├── document/          # Belge varlıkları, depolama ve ayrıştırma sözleşmeleri
│   │   ├── obligation/        # Yükümlülükler, vade tipleri, risk seviyeleri
│   │   ├── marketplace/       # RFQ fırsatları, tedarikçi teklifleri, Money modeli
│   │   └── audit/             # Değişmez (append-only) denetim izi modelleri
│   ├── application/           # İş Akışları (Use Cases) ve DTO Yapıları
│   │   ├── document/usecase/  # Belge yükleme, işleme ve içerik çekme akışları
│   │   ├── obligation/usecase/# Onaylama (Approve), reddetme (Dismiss) ve listeleme
│   │   ├── marketplace/usecase/# RFQ tetikleme, teklif değerlendirme ve kabul etme
│   │   ├── dashboard/usecase/ # KPI metrikleri, risk dağılımı ve tasarruf özetleri
│   │   └── admin/usecase/     # Agent Harness runtime yönetimi ve simülasyon
│   ├── agent/                 # Otonom Yapay Zeka Ajan Motoru
│   │   ├── analyst/           # Belge anlamsal çıkarım ajanı (Gemma 2:2B / Port 11434)
│   │   ├── verifier/          # Bağımsız çapraz teyit ve risk puanlama ajanı (Port 11435)
│   │   ├── marketplace/       # Alternatif tedarikçi teklif motoru (RFQ Engine)
│   │   ├── orchestrator/      # PII -> Analyst -> Verifier -> RFQ boru hattı koordinatörü
│   │   └── llmclient/         # Çalışma anında değiştirilebilir (DynamicProvider) arayüzü
│   ├── infrastructure/        # Adaptörler ve Dış Servis Entegrasyonları
│   │   ├── http/handler/      # Chi REST HTTP handler'ları
│   │   ├── graphql/           # GraphQL şeması, sorgu ve mutasyon resolver'ları
│   │   ├── mongodb/           # MongoDB repository implementasyonları
│   │   └── websocket/         # Canlı olay akışı için WebSocket Hub
│   ├── mcp/                   # Model Context Protocol (MCP) Motoru
│   │   ├── gmail/             # Fesih / itiraz e-postası hazırlama adaptörü
│   │   ├── calendar/          # Kritik vadeleri Google Calendar'a işleme adaptörü
│   │   ├── slack/             # Satın alma kanallarına anlık uyarı gönderme adaptörü
│   │   └── deepwiki/          # SSE tabanlı kurumsal bilgi tabanı adaptörü
│   ├── pii/                   # Sıfır-Sızıntı (Zero-Leak) Regex Maskeleme Motoru
│   └── shared/                # Konfigürasyon, slog loglama, in-memory event bus
└── deployments/               # Docker Compose ve ortam dosyaları
```

---

## 🚀 Kurulum ve Hızlı Başlangıç

### 1. Sistem Gereksinimleri
*   **Go**: `1.26+`
*   **Docker & Docker Compose**
*   **Ollama**: `gemma2:2b` modeli indirilmiş olmalıdır (`ollama run gemma2:2b`)

### 2. Ortam Değişkenleri (`.env`)
Kök dizindeki `.env` dosyasını yapılandırın:

```ini
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:8080

# AI Models (Ollama Gemma 2:2B Dual Ports)
LLM_ANALYST_PROVIDER=ollama
LLM_ANALYST_ENDPOINT=http://localhost:11434
LLM_ANALYST_MODEL=gemma2:2b

LLM_VERIFIER_PROVIDER=ollama
LLM_VERIFIER_ENDPOINT=http://localhost:11435
LLM_VERIFIER_MODEL=gemma2:2b

# Admin & Security
ADMIN_PASSWORD=admin123
```

### 3. Altyapıyı Başlatma (Docker)
```bash
# MongoDB, PostgreSQL ve Redis servislerini başlatın
docker compose -f deployments/docker-compose.yml up -d
```

### 4. Servisleri Çalıştırma

```bash
# 1. API Gateway (REST + GraphQL + WebSocket)
go run ./cmd/api

# 2. 7/24 Proaktif Vade Tarama Daemon'u (Ayrı terminalde)
go run ./cmd/worker

# 3. Terminal REPL Spark Asistanı (Ayrı terminalde)
go run ./cmd/spark
```

---

## 🤖 Multi-Agent & Orkestrasyon Motoru

ClaimPilot boru hattı, insan müdahalesine gerek kalmadan belgeleri işleyen 4 aşamalı bir mimariye sahiptir:

```
[Ham Belge Yükleme] 
        │
        ▼
[1. PII Redactor (internal/pii)]
   └─> TCKN, IBAN, Kredi Kartı, Telefon ve E-postalar bellek içinde [REDACTED] yapılır.
        │
        ▼
[2. Analyst Agent (Port 11434)]
   └─> Gemma 2:2B sözleşmeyi analiz eder: Bitiş tarihi, ihbar süresi, cayma bedeli ve ceza maddeleri çıkarılır.
        │
        ▼
[3. Verifier Agent (Port 11435)]
   └─> Gemma 2:2B bağımsız denetim yapar: Analyst'in çıkarımlarını kaynak metinle teyit eder ve Risk Skoru üretir.
        │
        ├─> Güven >= 0.90 ve Risk LOW ise: Durum -> IN_PROGRESS (Otonom Onay)
        └─> Risk HIGH veya Güven Düşükse: Durum -> PENDING_APPROVAL (İnsan Onayı)
        │
        ▼
[4. RFQ Engine & MCP Execution]
   └─> Alternatif tedarikçi teklifleri toplanır; Gmail, Calendar ve Slack üzerinden bildirimler tetiklenir.
```

---

## 📡 API ve Uç Noktalar

### 1. Sağlık ve Gözlemlenebilirlik (Health & Observability)
*   `GET /health` — Servis canlılık ve veritabanı durum kontrolü.
*   `GET /metrics` — Prometheus metrikleri.

### 2. Belgeler (Documents — `/api/v1/documents`)
*   `POST /` — Çok parçalı (multipart/form-data) PDF/resim/metin sözleşme yükleme.
*   `GET /` — Yüklenen belgelerin sayfalanmış listesi.
*   `GET /{id}` — Belgenin maskelenmiş içeriği ve çıkarılan alan detayları.

### 3. Yükümlülükler (Obligations — `/api/v1/obligations`)
*   `GET /` — Çıkarılan taahhüt ve yükümlülüklerin listesi (`?status=`, `?risk=`).
*   `POST /{id}/approve` — Kullanıcı onayı vererek MCP aksiyonlarını tetikleme.
*   `POST /{id}/dismiss` — Yükümlülüğü arşivleme veya iptal etme.

### 4. Pazar Yeri ve Teklifler (Marketplace — `/api/v1/marketplace`)
*   `POST /rfq/generate` — Gemma 2:2B ile yükümlülük için canlı alternatif teklif (bids) üretimi.
*   `GET /opportunities` — Açık pazar yeri fırsatlarının listesi.
*   `POST /bids/{id}/accept` — Alternatif tedarikçi teklifini kabul etme ve eski sözleşmeyi fesih sürecine alma.
*   `GET /metrics` — Toplam tasarruf miktarı, ortalama indirim oranı ve risk analizi.

### 5. Admin & Agent Harness (`/api/v1/admin`)
*   `GET /harness/config` — Çalışan LLM modelleri ve kayıtlı MCP adaptörlerinin canlı durumu.
*   `PUT /harness/llm` — Çalışma anında sistemi kapatmadan model değiştirme (**Hot-Swapping**).
*   `POST /harness/simulate` — Örnek sözleşme metni üzerinde baştan sona boru hattı simülasyonu koşturma.

### 6. Canlı Akış ve GraphQL
*   `GET /api/v1/ws?token=<jwt>` — WebSocket üzerinden gerçek zamanlı olay akışı (`obligation.verified`, vb.).
*   `POST /graphql` — GraphQL sorgu ve mutasyonları.
*   `GET /playground` — İnteraktif GraphQL test arayüzü.

---

## 🧪 Test ve Kalite Güvence

Backend katmanında birim, entegrasyon ve simülasyon testleri `testify` kütüphanesiyle icra edilir:

```bash
# Tüm backend testlerini koşun
go test -v ./internal/...

# Yalnızca Ajan ve Orkestrasyon testleri
go test -v ./internal/agent/orchestrator ./internal/agent/analyst ./internal/agent/verifier

# PII Maskeleme testleri
go test -v ./internal/pii

# Test kapsamını (coverage) HTML olarak raporlayın
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out
```

### Otomatik Test Kapsamı:
*   **Birim Testleri:** PII regex sınır testleri, DTO doğrulama kuralları, şema interceptor'ları.
*   **Mock Tabanlı Testler:** Harici veritabanı veya LLM bağlantısı olmadan çalışan `mockLLM` ve `mockDocRepo` testleri.
*   **Harness Simülasyonu:** Admin paneli üzerinden uçtan uca milisaniye bazlı trace testi.
*   **Postman Test Koleksiyonu:** `postman/masterfabric-go.postman_collection.json` dosyasında 37 adet hazır otomatik test.

---

## ⚡ Performans Optimizasyonları

1.  **Go Concurrency:** Dosya ayrıştırma, LLM çıkarımı ve MCP eylemleri ana HTTP iş parçacığını bloklamayan asenkron Goroutine'ler ile yönetilir.
2.  **İki Portlu Ollama Dağıtımı:** Analyst (11434) ve Verifier (11435) iki ayrı Ollama portuna paylaştırılarak model kuyruk kilitlenmeleri (throttling) engellenmiştir.
3.  **Hafıza İçi EventBus:** Mikroservis karmaşası olmadan, bellek içi Go kanalları (`chan`) ile mikro-saniyeler içinde olay dağıtımı yapılır.
4.  **Akışlı Yanıtlar (Streaming):** LLM çıktısı `StreamComplete` kanallarıyla arayüze token akışı şeklinde iletilir.
5.  **B-Tree Veritabanı İndeksleri:** `user_id`, `status`, `due_date` alanlarında indeksleme yapılarak gecikme önlenmiştir.

---

## 🔐 Güvenlik ve PII Kalkanı

*   **Sıfır-Sızıntı (Zero-Data-Leak) Prensibi:** Belgelerdeki TCKN, IBAN, kredi kartı ve telefon bilgileri model soketine erişmeden önce `internal/pii` tarafından yerel olarak maskelenir.
*   **Değişmez Denetim İzi (Audit Log):** Alınan her otonom karar, kullanılan model adı (`gemma2:2b`), güven skoru ve zaman damgası MongoDB `audit_logs` koleksiyonuna silinemez olarak mühürlenir.
*   **Prompt Politika Değildir:** Yetkilendirme ve güvenlik kuralları LLM sistem prompt'larına değil; Go middleware ve interceptor katmanlarına emanet edilmiştir.
