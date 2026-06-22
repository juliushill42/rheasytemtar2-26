# RHEA Distributed System

### 🔒 CONFIDENTIAL & PROPRIETARY DOCUMENTATION
**Copyright (c) 2026 Julius Cameron Hill. All rights reserved.** *This document and the underlying architecture, source code substrates, and compilation blueprints are strictly proprietary and confidential. Unauthorized distribution, inspection, or duplication via external networks or corporate entities is prohibited and protected under trade secret law.*

---

## 🏛️ System Architecture

RHEA is a highly optimized, distributed orchestration platform operating via five decoupled, zero-dependency microservices. To ensure absolute data isolation and zero runtime cross-contamination, each service runs on a dedicated, hard-locked port boundary.

### Core Services Matrix

| Service | Port | Substrate Layer | Operational Responsibility |
| :--- | :--- | :--- | :--- |
| **Orchestra** | `9100` | Go | Central traffic router, service discovery, and self-healing state coordination. |
| **Brain** | `9101` | Go | Deterministic decision-making engine and active policy window management. |
| **Cloning** | `9102` | Go | Blueprint template replication engine utilizing SHA-256 integrity loops. |
| **Sanctuary** | `9103` | Go | Ephemeral threat quarantine layer and isolation zone boundary control. |
| **Ledger** | `9104` | Go | Cryptographically immutable event tracking via blockchain-style verification structures. |
| **Glass Dashboard** | `9190` | TypeScript | Real-time observability interface featuring a deterministic 5-second polling loop. |

---

## 💻 Technical Stack & Runtime Dependencies

* **Backend Substrate:** Pure Go (Compiled natively from `scratch` images to optimize isolation).
* **Observability UI:** TypeScript + Deno + Tailwind CSS (Zero heavy Node modules overhead).
* **Virtualization Layer:** Docker + Docker Compose V2 Engine API.
* **Network Topography:** Completely isolated bridge networking featuring localized internal service registration.

---

## 🚀 Deployment & Operations

Manage the internal cluster architecture exclusively using local system commands from administrative terminal contexts:

```powershell
# Rapid Cluster Initialization
./start.sh

# Background Daemon Instantiation
docker-compose up -d

# Parallel Multi-Core Recompilation
docker-compose build --parallel

# Live Substrate Event Logging
docker-compose logs -f

# Absolute Environment Teardown
docker-compose down
