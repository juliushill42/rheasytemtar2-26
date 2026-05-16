# RHEA System Deployment Checklist
**Copyright (c) 2026 Julius Cameron Hill**

## Pre-Deployment

- [ ] Docker installed (version 20.10+)
- [ ] Docker Compose installed (version 2.0+)
- [ ] Ports 9100-9104, 9190 available
- [ ] Minimum 2GB RAM available
- [ ] Minimum 10GB disk space

## Deployment Steps

1. **Extract System**
   ```bash
   cd rhea-system
   ```

2. **Initialize Storage**
   ```bash
   mkdir -p storage/{policy_store,blueprints,quarantine,ledger_db}
   ```

3. **Build Services**
   ```bash
   make build
   # OR
   docker-compose build --parallel
   ```

4. **Start System**
   ```bash
   make start
   # OR
   ./start.sh
   ```

5. **Verify Health**
   ```bash
   make health
   ```

6. **Run Tests**
   ```bash
   make test
   # OR
   ./test.sh
   ```

7. **Access Dashboard**
   ```bash
   open http://localhost:9190
   ```

## Production Configuration

### Environment Variables

Create `.env` file:
```env
# Orchestra
ORCHESTRA_PORT=9100
SELF_HEAL_ENABLED=true

# Brain
BRAIN_PORT=9101
GLOBAL_CONTEXT_WINDOW=large

# Cloning
CLONING_PORT=9102

# Sanctuary
SANCTUARY_PORT=9103

# Ledger
LEDGER_PORT=9104

# Dashboard
DASHBOARD_PORT=9190
```

### Resource Limits

Update `docker-compose.yml`:
```yaml
services:
  rhea-orchestra:
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 128M
        reservations:
          cpus: '0.25'
          memory: 64M
```

### Logging

```yaml
services:
  rhea-orchestra:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

## Security Hardening

1. **Network Isolation**
   - Keep services on internal network
   - Expose only dashboard externally
   - Use reverse proxy (nginx/traefik)

2. **Secrets Management**
   - Use Docker secrets
   - Never commit .env files
   - Rotate keys regularly

3. **Volume Permissions**
   ```bash
   chmod 700 storage/*
   chown -R 1000:1000 storage/
   ```

## Monitoring

1. **Health Endpoints**
   - Orchestra: http://localhost:9100/health
   - Brain: http://localhost:9101/health
   - Cloning: http://localhost:9102/health
   - Sanctuary: http://localhost:9103/health
   - Ledger: http://localhost:9104/health

2. **Metrics**
   - Container stats: `docker stats`
   - Service logs: `docker-compose logs -f`
   - Dashboard: http://localhost:9190

3. **Alerts**
   - Set up monitoring (Prometheus/Grafana)
   - Configure alerting rules
   - Monitor disk usage for ledger

## Backup Strategy

1. **Critical Data**
   ```bash
   tar -czf rhea-backup-$(date +%Y%m%d).tar.gz storage/
   ```

2. **Restore**
   ```bash
   tar -xzf rhea-backup-YYYYMMDD.tar.gz
   ```

3. **Automated Backups**
   ```bash
   # Add to crontab
   0 2 * * * cd /path/to/rhea-system && tar -czf /backups/rhea-$(date +\%Y\%m\%d).tar.gz storage/
   ```

## Troubleshooting

### Services Won't Start
```bash
# Check logs
docker-compose logs

# Rebuild
make clean build start
```

### Port Conflicts
```bash
# Check ports
netstat -tulpn | grep 910

# Update ports in docker-compose.yml
```

### Performance Issues
```bash
# Check resources
docker stats

# Increase limits in docker-compose.yml
```

### Data Corruption
```bash
# Verify ledger
curl http://localhost:9104/verify

# Restore from backup
make clean
tar -xzf rhea-backup-YYYYMMDD.tar.gz
make start
```

## Maintenance

### Regular Tasks

**Daily**
- [ ] Check service health
- [ ] Monitor disk usage
- [ ] Review audit logs

**Weekly**
- [ ] Backup data
- [ ] Check ledger integrity
- [ ] Review quarantined items
- [ ] Update blueprints

**Monthly**
- [ ] Security updates
- [ ] Performance optimization
- [ ] Log rotation
- [ ] Test disaster recovery

### Updates

```bash
# Pull new code
git pull

# Rebuild
make clean build

# Deploy
make start

# Verify
make test
```

## Support

For issues or questions:
- Review ARCHITECTURE.md
- Check logs: `make logs`
- Run tests: `make test`
- Verify health: `make health`

---
Copyright (c) 2026 Julius Cameron Hill. All rights reserved.
