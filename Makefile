#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# exported vars
CONTAINER_CLI ?= docker
PROJECT_DIR := $(CURDIR)
export PROJECT_DIR

# Makefile vars
OUTPUT_DIR := ./out

# ============================================================================
# Docker Tool Configuration
# ============================================================================
# cbdc-tool 镜像包含: fxconfig, cryptogen, fabric-ca-client, fabric-ca-server, tokengen
# 可通过环境变量 DOCKER_TOOLS_IMAGE 覆盖默认镜像
DOCKER_TOOLS_IMAGE ?= ghcr.io/built-by-sign/fabric-x-tool:v0.0.4

# Docker 运行基础命令
# --rm: 容器退出后自动删除
# --user: 使用当前用户权限，避免生成 root 权限的文件
# -v: 挂载当前项目目录到容器内的 /workspace
# -w: 设置工作目录为 /workspace
# -e HOME: 设置 HOME 目录为 /tmp，避免工具在宿主机创建配置文件
# 传递 Docker 配置环境变量供 config-builder 使用
DOCKER_RUN_BASE = $(CONTAINER_CLI) run --rm \
	--user $(shell id -u):$(shell id -g) \
	-v $(PROJECT_DIR):/workspace \
	-w /workspace \
	-e HOME=/tmp

# 带网络访问的 Docker 运行命令（用于需要访问 orderer 等服务的操作）
# --network host: 使用宿主机网络，可以访问 localhost 上的服务
DOCKER_RUN_NET = $(DOCKER_RUN_BASE) \
	--network host




# ============================================================================
# Config Builder Targets
# ============================================================================


# Generate network configuration and docker-compose.yaml
.PHONY: setup
setup:
	@echo "==> Cleaning old artifacts..."
	@rm -rf $(OUTPUT_DIR)
	@echo "==> Generating Fabric network configuration..."
	$(DOCKER_RUN_BASE) $(DOCKER_TOOLS_IMAGE) \
		config-builder setup -c configs/test-full.yaml -o ./out --use-local-tools
	@echo "==> Generating docker-compose.yaml..."
	$(DOCKER_RUN_BASE) $(DOCKER_TOOLS_IMAGE) \
		config-builder gen-compose -c configs/test-full.yaml -o ./out --use-local-tools
	@echo "==> Network configuration ready in ./out"

# Start the network
.PHONY: start
start:
	@echo "Starting network using docker compose..."
	@cd $(OUTPUT_DIR) && $(CONTAINER_CLI) compose up -d

# Stop the network
.PHONY: stop
stop:
	@echo "Stopping network..."
	@cd $(OUTPUT_DIR) && $(CONTAINER_CLI) compose stop

# Teardown the network and remove volumes
.PHONY: teardown
teardown:
	@echo "Tearing down network..."
	@if [ -d "$(OUTPUT_DIR)" ]; then \
		cd $(OUTPUT_DIR) && $(CONTAINER_CLI) compose down -v; \
	else \
		echo "Output directory does not exist, skipping teardown"; \
	fi
	@$(CONTAINER_CLI) network inspect cbdc_net >/dev/null 2>&1 && $(CONTAINER_CLI) network rm cbdc_net || true

# Remove all generated artifacts
.PHONY: clean
clean:
	@rm -rf $(OUTPUT_DIR)

# Create namespace
.PHONY: create-ns
create-ns:
	@echo "Creating namespace..."
	$(DOCKER_RUN_NET) $(DOCKER_TOOLS_IMAGE) fxconfig namespace create cbdc \
		--channel=arma \
		--orderer=localhost:7050 \
		--mspID=Org1MSP \
		--mspConfigPath=./out/build/config/cryptogen-artifacts/crypto/peerOrganizations/org1.example.com/users/channel_admin@org1.example.com/msp \
		--pk=./out/build/config/cryptogen-artifacts/crypto/peerOrganizations/org1.example.com/users/endorser@org1.example.com/msp/signcerts/endorser@org1.example.com-cert.pem \
		--connTimeout=60s
	@until $(DOCKER_RUN_NET) $(DOCKER_TOOLS_IMAGE) fxconfig namespace list --endpoint=localhost:5500 | grep -q cbdc; do \
		sleep 2; \
		echo "waiting for namespace to be created..."; \
	done
	$(DOCKER_RUN_NET) $(DOCKER_TOOLS_IMAGE) fxconfig namespace list --endpoint=localhost:5500

# List namespaces
.PHONY: list-ns
list-ns:
	$(DOCKER_RUN_NET) $(DOCKER_TOOLS_IMAGE) fxconfig namespace list --endpoint=localhost:5500

# One-click: setup, start network and create namespace (for first time)
.PHONY: quickstart
quickstart: setup start create-ns
	@echo "=========================================="
	@echo "Network is ready!"
	@echo "=========================================="

# Restart: stop and start the network (keep existing config)
.PHONY: restart
restart: stop start
	@echo "=========================================="
	@echo "Network restarted!"
	@echo "=========================================="

# Rebuild: teardown, regenerate config and start fresh
.PHONY: rebuild
rebuild: teardown setup start create-ns
	@echo "=========================================="
	@echo "Network rebuilt successfully!"
	@echo "=========================================="


# Print the list of supported commands.
.PHONY: help
help:
	@echo "Fabric-X Network Makefile"
	@echo "========================="
	@echo ""
	@echo "Quick Start:"
	@echo "  quickstart    - One-click: setup, start network and create namespace (first time)"
	@echo "  restart       - Stop and start the network (keep existing config)"
	@echo "  rebuild       - Teardown, regenerate config and start fresh"
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
	@echo "  DOCKER_TOOLS_IMAGE  - cbdc-tool Docker image (required)"
	@echo "  CONTAINER_CLI       - Container CLI (default: docker)"
	@echo ""
