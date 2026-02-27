# Fabric-X Network Makefile
# SPDX-License-Identifier: Apache-2.0

# Configuration
CONTAINER_CLI ?= docker
PROJECT_DIR := $(CURDIR)
OUTPUT_DIR := ./out
CONFIG ?= configs/test-full.yaml
DOCKER_TOOLS_IMAGE ?= ghcr.io/built-by-sign/fabric-x-tool:v0.0.5

export PROJECT_DIR

# Docker base command
DOCKER_RUN_BASE = $(CONTAINER_CLI) run --rm \
	--user $(shell id -u):$(shell id -g) \
	-v $(PROJECT_DIR):/workspace \
	-w /workspace \
	-e HOME=/tmp

# Docker command with host network access
DOCKER_RUN_NET = $(DOCKER_RUN_BASE) --network host
# ============================================================================
# Network Configuration
# ============================================================================

# Generate network configuration and docker-compose.yaml
.PHONY: setup
setup:
	@echo "==> Cleaning old artifacts..."
	@rm -rf $(OUTPUT_DIR)
	@echo "==> Generating network configuration..."
	$(DOCKER_RUN_BASE) $(DOCKER_TOOLS_IMAGE) \
		config-builder setup -c $(CONFIG) -o ./out --use-local-tools
	@echo "==> Generating docker-compose.yaml..."
	$(DOCKER_RUN_BASE) $(DOCKER_TOOLS_IMAGE) \
		config-builder gen-compose -c $(CONFIG) -o ./out --use-local-tools
	@echo "==> Configuration ready in ./out"

# Start the network
.PHONY: start
start:
	@echo "Starting network..."
	@cd $(OUTPUT_DIR) && $(CONTAINER_CLI) compose up -d

# Stop the network
.PHONY: stop
stop:
	@echo "Stopping network..."
	@cd $(OUTPUT_DIR) && $(CONTAINER_CLI) compose stop

# Teardown network and remove volumes
.PHONY: teardown
teardown:
	@echo "Tearing down network..."
	@if [ -d "$(OUTPUT_DIR)" ]; then \
		cd $(OUTPUT_DIR) && $(CONTAINER_CLI) compose down -v; \
	else \
		echo "Output directory does not exist, skipping teardown"; \
	fi
	@$(CONTAINER_CLI) network inspect fabric-x-test-net >/dev/null 2>&1 && $(CONTAINER_CLI) network rm fabric-x-test-net || true

# Remove all generated artifacts
.PHONY: clean
clean:
	@rm -rf $(OUTPUT_DIR)

# ============================================================================
# Namespace Management
# ============================================================================

# Create namespace
.PHONY: create-ns
create-ns:
	@echo "Creating namespace..."
	$(DOCKER_RUN_NET) $(DOCKER_TOOLS_IMAGE) fxconfig namespace create ns1 \
		--channel=arma \
		--orderer=127.0.0.1:7050 \
		--mspID=Org1MSP \
		--mspConfigPath=./out/build/config/cryptogen-artifacts/crypto/peerOrganizations/org1.example.com/users/channel_admin@org1.example.com/msp \
		--pk=./out/build/config/cryptogen-artifacts/crypto/peerOrganizations/org1.example.com/users/endorser@org1.example.com/msp/signcerts/endorser@org1.example.com-cert.pem \
		--connTimeout=60s
	@until $(DOCKER_RUN_NET) $(DOCKER_TOOLS_IMAGE) fxconfig namespace list --endpoint=127.0.0.1:5500 | grep -q ns1; do \
		sleep 2; \
		echo "Waiting for namespace creation..."; \
	done
	$(DOCKER_RUN_NET) $(DOCKER_TOOLS_IMAGE) fxconfig namespace list --endpoint=127.0.0.1:5500

# List namespaces
.PHONY: list-ns
list-ns:
	$(DOCKER_RUN_NET) $(DOCKER_TOOLS_IMAGE) fxconfig namespace list --endpoint=127.0.0.1:5500

# ============================================================================
# Convenience Commands
# ============================================================================

# One-click: teardown, setup, start and create namespace
.PHONY: quickstart
quickstart: teardown setup start create-ns
	@echo "=========================================="
	@echo "Network is ready!"
	@echo "=========================================="

# Restart: stop and start (keep existing config)
.PHONY: restart
restart: stop start
	@echo "=========================================="
	@echo "Network restarted!"
	@echo "=========================================="

# ============================================================================
# Help
# ============================================================================

.PHONY: help
help:
	@echo "Fabric-X Network Makefile"
	@echo "========================="
	@echo ""
	@echo "Quick Start:"
	@echo "  quickstart    - Teardown, setup, start network and create namespace"
	@echo "  restart       - Stop and start the network (keep existing config)"
	@echo ""
	@echo "Main Commands:"
	@echo "  setup         - Generate network configuration and docker-compose.yaml"
	@echo "  start         - Start the network"
	@echo "  stop          - Stop the network"
	@echo "  teardown      - Stop network and remove volumes"
	@echo "  clean         - Remove all generated artifacts"
	@echo ""
	@echo "Namespace:"
	@echo "  create-ns     - Create namespace in Fabric-X"
	@echo "  list-ns       - List namespaces"
	@echo ""
	@echo "Environment Variables:"
	@echo "  CONFIG              - Network config file (default: configs/test-full.yaml)"
	@echo "  DOCKER_TOOLS_IMAGE  - Tool Docker image"
	@echo "  CONTAINER_CLI       - Container CLI (default: docker)"
	@echo ""
