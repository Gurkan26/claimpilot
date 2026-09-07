# ClaimPilot Security Policy & Agentic AI Protocol

> Aligned with the [Cursor Security Protocol](https://github.com/gurkanfikretgunak/cursor-security) by Gürkan Fikret Günak, SOC 2, ISO 27001, and OWASP ASVS standards.

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.0   | :white_check_mark: |

---

## 🛡 Agentic AI Security Protocol

Autonomous AI employees in ClaimPilot interact with external tools, parse sensitive legal documents, and recommend procurement decisions. To prevent autonomous compromise, the platform enforces 5 strict security pillars:

### 1. Least Agency & Sandboxed Tool Execution
* **Zero Direct Execution:** AI agents never execute shell commands or unvetted scripts directly.
* **Typed MCP Adapters:** External tools (Gmail, Google Calendar, Slack) implement the `mcp.Adapter` interface with strict parameter schemas.
* **Auto-Approve Guardrails:** Autonomous execution is restricted to low-risk actions where confidence >= 0.90. Contract terminations and financial transactions mandate Human-in-the-Loop authorization.

### 2. Zero-Leak PII Pre-Condition (KVKK & GDPR)
* Before text reaches any model (Ollama Gemma 2:2B or external fallback), it must pass through `internal/pii`.
* TCKN, IBAN, credit card, phone, and email information are masked in-process.

### 3. Identity Before Action
* All API endpoints and WebSocket streams require cryptographic JWT authentication.
* Tenant isolation is strictly enforced via `X-Organization-ID` and PostgreSQL/MongoDB query scoping.

### 4. Deterministic Authorization (Prompt is NOT a Policy)
* Access control (RBAC), quota limits, and input validations are enforced in deterministic Go middleware and interceptors — never in LLM prompts.

### 5. Immutable Audit Trail
* Every autonomous decision, model invocation, and MCP action is written to an append-only MongoDB `audit_logs` collection for full auditability and non-repudiation.

---

## 🔐 Security Controls Registry

| ID | Domain | Control | Implementation |
| -- | ------ | ------- | -------------- |
| **AC-01** | AI Security | Zero-Leak PII Sanitization | `internal/pii/patterns.go` (Pre-compiled Regex for TCKN, IBAN, Cards, Phones, Emails) |
| **AC-02** | AI Security | Dual-Agent Verification | `internal/agent/verifier` (Actor-Critic cross-check preventing hallucinations) |
| **AC-03** | AI Security | Human-in-the-Loop Thresholds | `internal/agent/orchestrator` (High/Critical risks forced to `PENDING_APPROVAL`) |
| **AC-04** | AI Security | MCP Tool Isolation | `internal/mcp` (Typed interfaces, strict schema validation, OAuth/Token bounds) |
| **SC-01** | Authentication | Cryptographic JWT Validation | `internal/infrastructure/auth/jwt_service.go` |
| **SC-02** | Authorization | Hierarchical RBAC | `internal/infrastructure/auth/rbac_service.go` (Wildcard-aware permission matching) |
| **SC-03** | Network | Strict CORS Boundaries | `CORS_ALLOWED_ORIGINS` with credentials disabled on wildcards |
| **SC-04** | Input Handling | Body Size Caps & Schema Validation | `middleware.MaxBodyBytes` (1 MiB default limit) and JSON schema interceptors |
| **SC-05** | Observability | Non-leaking Health Probes | `/health/ready` & `/health/live` strip internal error details from public output |
| **SC-06** | Secrets | Zero Secrets in Repositories | Protected by GitHub Push Protection and `.cursor-securityignore` |

---

## 🚨 Reporting a Vulnerability

**Please do NOT report security vulnerabilities through public GitHub issues.**

To report a vulnerability:
1. Open a **Private Security Advisory** on GitHub: [Security Advisory](https://github.com/Gurkan26/claimpilot/security/advisories/new)
2. Or contact the security team directly: `security@claimpilot.internal`

### Response SLAs
* **Initial Acknowledgment:** Within 24 hours.
* **Triage & Assessment:** Within 72 hours.
* **Remediation & Patch:** Within 14 to 30 days depending on severity.
