# RHEA Distributed System
**Copyright (c) 2026 Julius Cameron Hill. All rights reserved.**

## Architecture

RHEA is a distributed orchestration platform with five core services:

### Services

1. **Orchestra (Port 9100)** - Go traffic cop
   - Routes requests between services
   - Health monitoring with self-healing
   - Service discovery and coordination

2. **Brain (Port 9101)** - Policy engine
   - Decision making core
   - Policy evaluation and enforcement
   - Context window management

3. **Cloning (Port 9102)** - Blueprint manager
   - Template replication
   - Version control
   - SHA-256 integrity verification

4. **Sanctuary (Port 9103)** - Isolation layer
   - Threat quarantine
   - Security enforcement
   - Threat level classification

5. **Ledger (Port 9104)** - Audit log
   - Immutable event tracking
   - Blockchain-style verification
   - Queryable audit trail

6. **Glass Dashboard (Port 9190)** - Monitoring UI
   - Real-time service health
   - TypeScript + Tailwind CSS
   - 5-second auto-refresh

## Tech Stack

- **Backend**: Go (zero dependencies, scratch containers)
- **Dashboard**: TypeScript + Deno + Tailwind CSS
- **Container**: Docker + Docker Compose
- **Network**: Bridge networking with service discovery

## Deployment

```bash
# Quick start
./start.sh

# Manual
docker-compose up -d

# Rebuild
docker-compose build --parallel

# Logs
docker-compose logs -f

# Shutdown
docker-compose down
```

## API Endpoints

### Orchestra
```bash
POST /route - Route request to target service
GET  /health - Health status
```

### Brain
```bash
POST /decision - Policy evaluation
POST /policy/create - Create policy
GET  /policy/list - List policies
GET  /health - Health status
```

### Cloning
```bash
POST /create - Create blueprint
POST /clone - Clone from blueprint
GET  /list - List blueprints
GET  /health - Health status
```

### Sanctuary
```bash
POST /quarantine - Isolate threat
POST /release - Release from quarantine
GET  /list - List quarantined items
GET  /health - Health status
```

### Ledger
```bash
POST /audit - Log event
POST /query - Query audit log
GET  /verify - Verify chain integrity
GET  /health - Health status
```

## Performance

- **Build time**: < 30 seconds
- **Container size**: < 10MB per service (Go scratch images)
- **Memory**: < 50MB per service
- **Startup**: < 5 seconds
- **Latency**: < 5ms inter-service

## Security

- All events audited in immutable ledger
- SHA-256 integrity verification
- Threat isolation in Sanctuary
- Policy-based access control
- Zero external dependencies in Go services

## Monitoring

Access Glass Dashboard at http://localhost:9190

Features:
- Real-time service health
- Latency tracking
- Auto-refresh every 5s
- Color-coded status indicators

## License

Copyright (c) 2026 Julius Cameron Hill. All rights reserved.
Proprietary and confidential.
