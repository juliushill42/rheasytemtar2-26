# Copyright (c) 2026 Julius Cameron Hill. All rights reserved.

.PHONY: build start stop restart logs clean health test

build:
	@echo "Building RHEA services..."
	docker-compose build --parallel

start:
	@echo "Starting RHEA system..."
	docker-compose up -d
	@sleep 3
	@make health

stop:
	@echo "Stopping RHEA system..."
	docker-compose down

restart: stop start

logs:
	docker-compose logs -f

clean:
	@echo "Cleaning RHEA system..."
	docker-compose down -v
	rm -rf storage/*/
	mkdir -p storage/{policy_store,blueprints,quarantine,ledger_db}

health:
	@echo "=== RHEA HEALTH CHECK ==="
	@curl -s http://localhost:9100/health | jq -r '.status' && echo "Orchestra: OK" || echo "Orchestra: FAIL"
	@curl -s http://localhost:9101/health | jq -r '.status' && echo "Brain: OK" || echo "Brain: FAIL"
	@curl -s http://localhost:9102/health | jq -r '.status' && echo "Cloning: OK" || echo "Cloning: FAIL"
	@curl -s http://localhost:9103/health | jq -r '.status' && echo "Sanctuary: OK" || echo "Sanctuary: FAIL"
	@curl -s http://localhost:9104/health | jq -r '.status' && echo "Ledger: OK" || echo "Ledger: FAIL"

test:
	@echo "=== TESTING RHEA SYSTEM ==="
	@echo "Creating test policy..."
	@curl -s -X POST http://localhost:9101/policy/create \
		-H "Content-Type: application/json" \
		-d '{"name":"test","priority":10,"enabled":true,"rules":[{"condition":"test","action":"allow"}]}' | jq
	@echo "\nCreating test blueprint..."
	@curl -s -X POST http://localhost:9102/create \
		-H "Content-Type: application/json" \
		-d '{"name":"test-bp","version":"1.0","template":{"key":"value"}}' | jq
	@echo "\nQuarantining test threat..."
	@curl -s -X POST http://localhost:9103/quarantine \
		-H "Content-Type: application/json" \
		-d '{"source":"test","threat_level":2,"payload":{"data":"test"}}' | jq
	@echo "\nVerifying ledger..."
	@curl -s http://localhost:9104/verify | jq

dashboard:
	@echo "Opening Glass Dashboard..."
	@open http://localhost:9190 || xdg-open http://localhost:9190

all: build start
