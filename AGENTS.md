# ClaimPilot — AI Agent Security & Development Protocol

> This file defines authoritative instructions and security constraints for all AI coding agents (Cursor, Claude, Copilot, Antigravity) operating in the `claimpilot` repository.
> Reference: [Cursor Security Protocol](https://github.com/gurkanfikretgunak/cursor-security)

---

## 🛡 Mandatory Security Rules (Never Violate)

### 1. Zero Secrets in Source Code
* **NEVER** write or commit real API keys, tokens, passwords, private keys, or credentials into source code, documentation, or frontend assets.
* All configuration must pass via environment variables (`.env`, loaded via `internal/shared/config/config.go`).
* Maintain test fixtures with clearly synthetic placeholders (e.g. `mock-token-123`).

### 2. Zero-Leak PII Pre-Condition
* All document text and raw user files MUST be processed by the in-process **Pattern Redactor** (`internal/pii`) before passing to any LLM provider (Ollama, Gemma, etc.).
* Masks must replace sensitive information with standard redaction tokens:
  - `[TCKN_REDACTED]`
  - `[IBAN_REDACTED]`
  - `[CARD_REDACTED]`
  - `[PHONE_REDACTED]`
  - `[EMAIL_REDACTED]`

### 3. Prompt is NOT a Policy
* **NEVER** rely on LLM system prompts to enforce security, user authorization, multi-tenant isolation, or data validation.
* Security rules, access control (RBAC), and mathematical boundaries must be deterministically implemented in Go middleware, interceptors, and database query filters.

### 4. Least Agency & Tool Safety (MCP)
* External tool executions (Gmail, Calendar, Slack) must go through the typed `mcp.Adapter` interface.
* Validate all arguments against strict JSON schemas before invoking tools.
* External actions with financial, contractual, or irreversible impacts must require `Human-in-the-Loop` confirmation unless risk is mathematically verified as `LOW` with `confidence >= 0.90`.

### 5. Architectural Invariants (Clean / Hexagonal)
* The `internal/domain` layer must remain pure Go: **ZERO external framework or database dependencies**.
* Use Cases (`internal/application`) must only depend on Domain interfaces, never directly on infrastructure packages (`mongodb`, `postgres`, `http`).
* All persistence operations must be performed through Repository interfaces.

### 6. Audit Logging Integrity
* Every autonomous decision, obligation state transition, and tool call must produce an immutable audit log entry in MongoDB (`audit_logs`).
* Logs must include: `agent_type`, `model_used`, `confidence_score`, `timestamp`, `user_id`, and `action_result`.
* Clear-text credentials or unmasked PII must **never** be logged.
