# ClaimPilot Desktop — Cross-Platform AI Employee (macOS & Windows)

ClaimPilot Desktop is an autonomous AI employee application for corporate and personal obligations, contract renewals, deadline management, and automated procurement marketplace switching.

Built with **Electron + Vite + React + TypeScript** adhering strictly to the **Frontend Design** principles (restraint, gravitas, deliberate typography, and high operational clarity).

---

## 🖥 Features

1. **Frameless Desktop Shell**:
   - Native macOS traffic light integration (`titleBarStyle: 'hiddenInset'`).
   - Custom Windows titlebar and window controls (minimize, maximize/restore, close).
   - Real-time Go API backend connection status badge.

2. **Executive "Good Morning" Dashboard**:
   - Daily AI synthesized briefing.
   - 4-tier risk distribution (Critical, High, Medium, Low).
   - Imminent deadline countdown with One-Click Approve action.
   - Financial counters (Closed GMV, Realized Net Savings).

3. **Autonomous Obligation Tracking**:
   - Filters by urgency (7, 14, 30, 90 days ahead) and category.
   - One-Click Approve with automated MCP tool execution (Gmail cancellation notice, Calendar event scheduling, Slack alerts).
   - Dismiss obligation with auditable reason tracking.

4. **Procurement Marketplace & RFQ**:
   - Side-by-side alternative vendor quotes (Hetzner, GCP, Canva, Figma).
   - Realized savings percentage pill (e.g. Save 52%).
   - Verifier Agent risk scores & notes.
   - One-Click "Accept Deal & Switch Vendor" with 4% platform take-rate commission calculation.

5. **Document Intake & Zero-Retention PII Masking**:
   - Native desktop drag-and-drop zone.
   - Native OS file picker dialog (`window.claimpilotDesktop.openFileDialog`).
   - KVKK/GDPR automatic PII redaction (TCKN, IBAN, Phone, Email).

6. **Compliance Audit Trail**:
   - Immutable log of all agent decisions, manual approvals, and MCP dispatches.

---

## 🚀 Quick Start

### 1. Prerequisites
- Node.js 18+ and npm.
- ClaimPilot Go Backend running on `http://localhost:8080` (optional; desktop includes high-fidelity fallback demo mode).

### 2. Development

```bash
# Run web preview with hot-reload
npm run dev

# Run desktop Electron app in development mode
npm run electron:dev
```

### 3. Build & Packaging

```bash
# Compile TypeScript and bundle with Vite
npm run build

# Package standalone desktop executable for macOS (.dmg, .zip)
npm run package:mac

# Package standalone desktop executable for Windows (.exe, NSIS)
npm run package:win
```
