#!/bin/bash
# Copyright (c) 2026 Julius Cameron Hill. All rights reserved.

set -e

echo "=== RHEA SYSTEM STARTUP ==="
echo "Initializing distributed orchestration platform..."

# Create storage directories
mkdir -p storage/{policy_store,blueprints,quarantine,ledger_db}

# Build and start services
echo "Building containers..."
docker-compose build --parallel

echo "Starting services..."
docker-compose up -d

echo ""
echo "=== SERVICE STATUS ==="
docker-compose ps

echo ""
echo "=== ENDPOINTS ==="
echo "Orchestra:   http://localhost:9100"
echo "Brain:       http://localhost:9101"
echo "Cloning:     http://localhost:9102"
echo "Sanctuary:   http://localhost:9103"
echo "Ledger:      http://localhost:9104"
echo "Dashboard:   http://localhost:9190"

echo ""
echo "=== HEALTH CHECK ==="
sleep 5

for service in orchestra brain cloning sanctuary ledger; do
  port=$((9100 + $(echo "orchestra brain cloning sanctuary ledger" | tr ' ' '\n' | grep -n "^$service$" | cut -d: -f1) - 1))
  status=$(curl -s http://localhost:$port/health || echo "FAILED")
  echo "$service: $status"
done

echo ""
echo "Dashboard: http://localhost:9190"
echo "RHEA system operational."
