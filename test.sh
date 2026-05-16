#!/bin/bash
# Copyright (c) 2026 Julius Cameron Hill. All rights reserved.
# RHEA System Integration Tests

set -e

BASE_URL="http://localhost"
ORCHESTRA_PORT=9100
BRAIN_PORT=9101
CLONING_PORT=9102
SANCTUARY_PORT=9103
LEDGER_PORT=9104

echo "=== RHEA INTEGRATION TESTS ==="
echo ""

# Test 1: Health checks
echo "[TEST 1] Health Checks"
for service in "Orchestra:$ORCHESTRA_PORT" "Brain:$BRAIN_PORT" "Cloning:$CLONING_PORT" "Sanctuary:$SANCTUARY_PORT" "Ledger:$LEDGER_PORT"; do
  name=$(echo $service | cut -d: -f1)
  port=$(echo $service | cut -d: -f2)
  response=$(curl -s $BASE_URL:$port/health)
  if echo $response | grep -q "healthy"; then
    echo "✓ $name health check passed"
  else
    echo "✗ $name health check failed"
    exit 1
  fi
done
echo ""

# Test 2: Policy creation and listing
echo "[TEST 2] Brain - Policy Management"
policy_response=$(curl -s -X POST $BASE_URL:$BRAIN_PORT/policy/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "integration_test_policy",
    "priority": 100,
    "enabled": true,
    "rules": [
      {"condition": "integration_test", "action": "allow"}
    ]
  }')

if echo $policy_response | grep -q "success"; then
  echo "✓ Policy created successfully"
  policy_id=$(echo $policy_response | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
  echo "  Policy ID: $policy_id"
else
  echo "✗ Policy creation failed"
  exit 1
fi

list_response=$(curl -s $BASE_URL:$BRAIN_PORT/policy/list)
if echo $list_response | grep -q "integration_test_policy"; then
  echo "✓ Policy listing verified"
else
  echo "✗ Policy listing failed"
  exit 1
fi
echo ""

# Test 3: Blueprint creation and cloning
echo "[TEST 3] Cloning - Blueprint Management"
blueprint_response=$(curl -s -X POST $BASE_URL:$CLONING_PORT/create \
  -H "Content-Type: application/json" \
  -d '{
    "name": "integration_test_blueprint",
    "version": "1.0.0",
    "template": {
      "type": "test",
      "config": {"param1": "value1"}
    }
  }')

if echo $blueprint_response | grep -q "success"; then
  echo "✓ Blueprint created successfully"
  bp_id=$(echo $blueprint_response | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
  bp_hash=$(echo $blueprint_response | grep -o '"hash":"[^"]*"' | cut -d'"' -f4)
  echo "  Blueprint ID: $bp_id"
  echo "  Hash: $bp_hash"
else
  echo "✗ Blueprint creation failed"
  exit 1
fi

clone_response=$(curl -s -X POST $BASE_URL:$CLONING_PORT/clone \
  -H "Content-Type: application/json" \
  -d "{\"blueprint_id\": \"$bp_id\"}")

if echo $clone_response | grep -q "replica_id"; then
  echo "✓ Blueprint cloned successfully"
  replica_id=$(echo $clone_response | grep -o '"replica_id":"[^"]*"' | cut -d'"' -f4)
  echo "  Replica ID: $replica_id"
else
  echo "✗ Blueprint cloning failed"
  exit 1
fi
echo ""

# Test 4: Quarantine operations
echo "[TEST 4] Sanctuary - Quarantine Management"
quarantine_response=$(curl -s -X POST $BASE_URL:$SANCTUARY_PORT/quarantine \
  -H "Content-Type: application/json" \
  -d '{
    "source": "integration_test",
    "threat_level": 2,
    "payload": {"suspicious": "data", "test": true},
    "notes": "Integration test quarantine"
  }')

if echo $quarantine_response | grep -q "success"; then
  echo "✓ Threat quarantined successfully"
  quar_id=$(echo $quarantine_response | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
  quar_hash=$(echo $quarantine_response | grep -o '"hash":"[^"]*"' | cut -d'"' -f4)
  echo "  Quarantine ID: $quar_id"
  echo "  Hash: $quar_hash"
else
  echo "✗ Quarantine failed"
  exit 1
fi

list_quar=$(curl -s $BASE_URL:$SANCTUARY_PORT/list)
if echo $list_quar | grep -q "$quar_id"; then
  echo "✓ Quarantine listing verified"
else
  echo "✗ Quarantine listing failed"
  exit 1
fi
echo ""

# Test 5: Ledger audit and verification
echo "[TEST 5] Ledger - Audit Trail"
verify_response=$(curl -s $BASE_URL:$LEDGER_PORT/verify)
if echo $verify_response | grep -q '"valid":true'; then
  echo "✓ Ledger integrity verified"
  entry_count=$(echo $verify_response | grep -o '"entries":[0-9]*' | cut -d':' -f2)
  echo "  Audit entries: $entry_count"
else
  echo "✗ Ledger verification failed"
  exit 1
fi

query_response=$(curl -s -X POST $BASE_URL:$LEDGER_PORT/query \
  -H "Content-Type: application/json" \
  -d '{"limit": 10}')

if echo $query_response | grep -q "entries"; then
  echo "✓ Audit query successful"
else
  echo "✗ Audit query failed"
  exit 1
fi
echo ""

# Test 6: Orchestra routing
echo "[TEST 6] Orchestra - Service Routing"
route_response=$(curl -s -X POST $BASE_URL:$ORCHESTRA_PORT/route \
  -H "Content-Type: application/json" \
  -d '{
    "action": "health",
    "target": "brain",
    "payload": {}
  }')

if echo $route_response | grep -q "success"; then
  echo "✓ Orchestra routing functional"
  latency=$(echo $route_response | grep -o '"latency_ms":[0-9]*' | cut -d':' -f2)
  echo "  Routing latency: ${latency}ms"
else
  echo "✗ Orchestra routing failed"
  exit 1
fi
echo ""

echo "=== ALL TESTS PASSED ==="
echo "RHEA system operational and verified."
