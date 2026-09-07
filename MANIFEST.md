# Agentic AI Security Manifest — ClaimPilot

> Aligned with the [Agentic AI Security Protocol](https://github.com/gurkanfikretgunak/cursor-security) by Gürkan Fikret Günak.

## Purpose

Agentic AI systems do not merely answer queries or summarize text. In ClaimPilot, autonomous agents **plan, extract financial data, evaluate risks, negotiate bids, and trigger real-world actions across external tools** (Gmail, Google Calendar, Slack) via the Model Context Protocol (MCP).

Security for agentic systems must govern **intent, capability, memory, and side-effects** — not merely model prompt engineering.

---

## The 8 Principles of Agentic AI Security

### 1. Least Agency (Enforced Capabilities)
* Grant only the tools, scopes, and side-effects an agent strictly requires for its assigned lifecycle step.
* **Read over Write:** Analysis agents (Analyst) are restricted to read-only document processing.
* **Draft over Commit:** MCP email adapters produce drafts or request confirmation prior to irrevocable dispatch.
* **Human Approval for High Impact:** Irreversible or high-risk obligations (contract cancellation, supplier switching) mandate explicit Human-in-the-Loop authorization.

### 2. Identity Before Action
* Every agent execution lifecycle is anchored to an authenticated principal (`user_id`, `organization_id`, and verified session context).
* Agent runs establish distinct, non-forgeable boundary scopes. Cross-tenant or unauthenticated agent actions are rejected at the gateway level.

### 3. Tool Trust is Zero by Default (Zero-Trust MCP)
* Treat all external tools, MCP servers, adapters, and third-party APIs as untrusted boundaries.
* Strictly validate input schemas and sanitize return values prior to downstream consumption.
* Isolate network egress and file system paths.

### 4. Prompt is Not a Policy
* Natural language prompt instructions are advisory, never authoritative security controls.
* Authorization, rate limiting, and business validation rules are strictly enforced in deterministic Go middleware, gateway interceptors, and database query filters.

### 5. Human Control for High Impact (Auto-Approve Guardrails)
* Financial, contractual, or privacy-critical actions require explicit human sign-off.
* Automated approval (`Auto-Approve`) is restricted to obligations where confidence score meets strict mathematical thresholds (`confidence >= 0.90`) and independently verified risk is `LOW`.

### 6. Observable by Design (Immutable Audit Trail)
* Maintain an append-only, tamper-evident audit log in MongoDB (`audit_logs`) recording:
  - Agent identity, model designation (`gemma2:2b`), and inference latency.
  - PII masking verification tokens.
  - Exact tool inputs, parameters, and invocation results.
  - Human approval timestamps and cryptographic user identifiers.
* Secrets, raw TCKN/IBAN numbers, and access tokens must never appear in log payloads.

### 7. Memory & Context is Sensitive Data (Zero-Leak Boundary)
* Long-term memory, vector stores, and prompt contexts are confidential assets.
* All incoming contract text must transit the local **Zero-Leak PII Redaction Engine** before reaching any model socket.
* Apply strict time-to-live (TTL) and data retention parameters.

### 8. Containment and Egress Isolation
* Enforce strict Cross-Origin Resource Sharing (CORS) limits.
* Restrict local desktop and backend egress channels, prohibiting unvetted external HTTP proxying or credential exfiltration.
