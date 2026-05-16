# RHEA System Architecture

## Directory Structure
```
rhea-system/
├── docker-compose.yml          # Container orchestration
├── Makefile                    # Build automation
├── start.sh                    # Quick start script
├── test.sh                     # Integration tests
├── README.md                   # Documentation
│
├── rhea-orchestra/             # Traffic coordinator (Port 9100)
│   ├── main.go                 # Go service implementation
│   └── Dockerfile              # Multi-stage build
│
├── rhea-brain/                 # Policy engine (Port 9101)
│   ├── main.go                 # Decision making core
│   └── Dockerfile              # Multi-stage build
│
├── rhea-cloning/               # Blueprint manager (Port 9102)
│   ├── main.go                 # Template replication
│   └── Dockerfile              # Multi-stage build
│
├── rhea-sanctuary/             # Isolation layer (Port 9103)
│   ├── main.go                 # Threat quarantine
│   └── Dockerfile              # Multi-stage build
│
├── rhea-ledger/                # Audit log (Port 9104)
│   ├── main.go                 # Immutable event tracker
│   └── Dockerfile              # Multi-stage build
│
├── glass-dashboard/            # Monitoring UI (Port 9190)
│   ├── main.ts                 # TypeScript + Deno
│   └── Dockerfile              # Deno runtime
│
└── storage/                    # Persistent data
    ├── policy_store/           # Brain policies
    ├── blueprints/             # Cloning templates
    ├── quarantine/             # Sanctuary isolation
    └── ledger_db/              # Audit logs
```

## Service Communication

```
                    ┌─────────────────┐
                    │  Glass Dashboard │
                    │  (TypeScript)    │
                    │  Port 9190       │
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │   Orchestra      │
                    │   (Go Traffic    │
                    │    Cop)          │
                    │   Port 9100      │
                    └────────┬─────────┘
                             │
          ┌──────────────────┼──────────────────┐
          │                  │                  │
    ┌─────▼──────┐  ┌───────▼────┐  ┌─────────▼─────┐
    │   Brain    │  │  Cloning   │  │  Sanctuary    │
    │  (Policy)  │  │ (Blueprint)│  │ (Quarantine)  │
    │ Port 9101  │  │ Port 9102  │  │  Port 9103    │
    └────────────┘  └─────┬──────┘  └───────┬───────┘
                          │                 │
                    ┌─────▼─────────────────▼─┐
                    │       Ledger            │
                    │   (Audit Trail)         │
                    │     Port 9104           │
                    └─────────────────────────┘
```

## Technology Stack

### Backend Services (Go)
- **Zero external dependencies**
- **Scratch containers** (< 10MB each)
- **Sub-5ms latency**
- **Concurrent request handling**
- **SHA-256 integrity**

### Dashboard (TypeScript + Deno)
- **Tailwind CSS** for styling
- **Auto-refresh** every 5s
- **Real-time health monitoring**
- **Color-coded status**

### Infrastructure
- **Docker multi-stage builds**
- **Bridge networking**
- **Volume persistence**
- **Health checks**
- **Auto-restart policies**

## Performance Characteristics

| Metric              | Value         |
|---------------------|---------------|
| Build Time          | < 30s         |
| Container Size      | < 10MB/svc    |
| Memory Per Service  | < 50MB        |
| Startup Time        | < 5s          |
| Inter-Service Latency| < 5ms        |
| Health Check Interval| 10s          |
| Dashboard Refresh   | 5s            |

## API Flow Example

1. **Client** → Dashboard (9190)
2. **Dashboard** → Orchestra (9100) `/route`
3. **Orchestra** → Brain (9101) `/decision`
4. **Brain** → Ledger (9104) `/audit` (async)
5. **Response** flows back through chain

## Security Features

- Immutable audit trail (blockchain-style)
- SHA-256 content verification
- Threat isolation chamber
- Policy-based access control
- Container network isolation
- Volume-mounted persistence

## Quick Commands

```bash
# Start system
make start

# Run tests
make test

# View logs
make logs

# Health check
make health

# Open dashboard
make dashboard

# Full rebuild
make clean build start
```

## Copyright

Copyright (c) 2026 Julius Cameron Hill. All rights reserved.
