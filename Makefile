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
DOCKER_DIR := $(CURDIR)/docker

# ============================================================================
# Docker Tool Configuration
# ============================================================================
# cbdc-tool 镜像包含: fxconfig, cryptogen, fabric-ca-client, fabric-ca-server,
# tokengen, libkms_pkcs11.so
# 使用 .env 文件中定义的 DOCKER_TOOLS_IMAGE 变量

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


# Build all the artifacts using cbdc-tool Docker image
.PHONY: setup-fabric
setup-fabric:
	@echo "==> Generating Fabric network configuration using cbdc-tool..."
	$(DOCKER_RUN_BASE) $(DOCKER_TOOLS_IMAGE) \
		config-builder setup -c configs/test-full.yaml -o ./out --use-local-tools
	@echo "==> Network configuration generated in ./out"

# Generate docker-compose.yaml file using cbdc-tool Docker image
.PHONY: gen-compose
gen-compose:
	@echo "==> Generating docker-compose.yaml using cbdc-tool..."
	$(DOCKER_RUN_BASE) $(DOCKER_TOOLS_IMAGE) \
		config-builder gen-compose -c configs/test-full.yaml -o ./out --use-local-tools
	@echo "==> docker-compose.yaml generated in ./out"

# Clean all the artifacts (configs and bins) built on the controller node (e.g. make clean).
.PHONY: clean-fabric
clean-fabric:
	rm -rf $(OUTPUT_DIR)

# Start fabric-x on the targeted hosts.
.PHONY: start-fabric
start-fabric:
	@echo "Starting network using docker compose..."
	@cd $(OUTPUT_DIR) && $(CONTAINER_CLI) compose up -d

# Create a namespace in fabric-x for the tokens.
# 使用 Docker 容器运行 fxconfig 工具
# 注意：路径已转换为容器内路径（/workspace 对应宿主机的 PROJECT_DIR）
create-ns:
	@echo "Creating namespace using Docker..."
	$(DOCKER_RUN_HSM) $(DOCKER_TOOLS_IMAGE) fxconfig namespace create cbdc \
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

# Create a namespace in fabric-x for the tokens.
# 使用 Docker 容器运行 fxconfig 工具
# 注意：路径已转换为容器内路径（/workspace 对应宿主机的 PROJECT_DIR）
create-ns-dev:
	@echo "Creating namespace using Docker..."
	$(DOCKER_RUN_HSM) $(DOCKER_TOOLS_IMAGE) fxconfig namespace create cbdc \
		--channel=arma \
		--orderer=cbdc-dev.sign.global:7050 \
		--mspID=Org1MSP \
		--mspConfigPath=./out/build/config/cryptogen-artifacts/crypto/peerOrganizations/org1.example.com/users/channel_admin@org1.example.com/msp \
		--pk=./out/build/config/cryptogen-artifacts/crypto/peerOrganizations/org1.example.com/users/endorser@org1.example.com/msp/signcerts/endorser@org1.example.com-cert.pem \
		--pkcs11-library=/app/libkms_pkcs11.so \
		--pkcs11-label="$(KMS_TOKEN_LABEL)" \
		--pkcs11-pin=$(KMS_USER_PIN) \
		--connTimeout=60s
	@until $(DOCKER_RUN_NET) $(DOCKER_TOOLS_IMAGE) fxconfig namespace list --endpoint=cbdc-dev.sign.global:5500 | grep -q cbdc; do \
		sleep 2; \
		echo "waiting for namespace to be created..."; \
	done
	$(DOCKER_RUN_NET) $(DOCKER_TOOLS_IMAGE) fxconfig namespace list --endpoint=cbdc-dev.sign.global:5500

# List namespaces
# 使用 Docker 容器运行 fxconfig 工具
.PHONY: list-ns
list-ns:
	$(DOCKER_RUN_NET) $(DOCKER_TOOLS_IMAGE) fxconfig namespace list --endpoint=localhost:5500

# Stop the targeted hosts (e.g. make fabric-x stop).
.PHONY: stop-fabric
stop-fabric:
	@echo "Stopping network using docker compose..."
	@cd $(OUTPUT_DIR) && $(CONTAINER_CLI) compose stop

# Teardown the targeted hosts (e.g. make fabric-x teardown).
.PHONY: teardown-fabric
teardown-fabric:
	@echo "Teardown network using docker compose..."
	@if [ -d "$(OUTPUT_DIR)" ]; then \
		cd $(OUTPUT_DIR) && $(CONTAINER_CLI) compose down -v; \
	else \
		echo "Output directory does not exist, skipping teardown"; \
	fi
	@$(CONTAINER_CLI) network inspect cbdc_net >/dev/null 2>&1 && $(CONTAINER_CLI) network rm cbdc_net || true

# Restart the targeted hosts (e.g. make fabric-x restart).
.PHONY: restart-fabric
restart-fabric: teardown-fabric start-fabric

# Build all the artifacts and binaries, and copy them to the application folders
.PHONY: setup
setup: clean setup-fabric gen-compose

# Start a Fabric and token network.
.PHONY: start
start: start-fabric

# Teardown Fabric and the token network.
.PHONY: teardown
teardown: teardown-fabric

# Stop Fabric and the token network.
.PHONY: stop
stop: stop-fabric

# Remove all generated crypto.
.PHONY: clean
clean: clean-fabric


# Print the list of supported commands.
.PHONY: help
help:
	@echo "CBDC Network Makefile - Available Targets"
	@echo "=========================================="
	@echo ""
	@echo "Network Configuration:"
	@echo "  setup-fabric              - Generate Fabric network config using cbdc-tool"
	@echo "  setup-fabric-kms          - Generate Fabric network config with KMS using cbdc-tool"
	@echo "  gen-compose               - Generate docker-compose.yaml using cbdc-tool"
	@echo "  gen-compose-kms           - Generate docker-compose.yaml with KMS using cbdc-tool"
	@echo ""
	@echo "Network Operations:"
	@echo "  setup                     - Clean and setup network (software mode)"
	@echo "  setup-kms                 - Clean and setup network (KMS mode)"
	@echo "  start-fabric              - Start Fabric network using docker compose"
	@echo "  stop-fabric               - Stop Fabric network"
	@echo "  teardown-fabric           - Teardown Fabric network and remove volumes"
	@echo "  restart-fabric            - Restart Fabric network"
	@echo "  create-ns                 - Create namespace in Fabric-X"
	@echo "  list-ns                   - List namespaces in Fabric-X"
	@echo ""
	@echo "Docker Image Building:"
	@echo "  build-orderer-kms         - Build KMS-enabled orderer Docker image"
	@echo "  build-kms-images          - Build all KMS-enabled images"
	@echo ""
	@echo "Cleanup:"
	@echo "  clean-fabric              - Remove all generated artifacts"
	@echo "  clean                     - Alias for clean-fabric"
	@echo ""
	@echo "Environment Variables:"
	@echo "  DOCKER_TOOLS_IMAGE        - cbdc-tool Docker image (default: from .env)"
	@echo "  SIGN_KMS_ENDPOINT         - KMS endpoint (default: host.docker.internal:9200)"
	@echo "  KMS_TOKEN_LABEL           - KMS token label (default: tk)"
	@echo "  CONTAINER_CLI             - Container CLI to use (default: docker)"
	@echo ""
	@echo "Examples:"
	@echo "  make setup                                    # Setup network (software mode)"
	@echo "  make setup-kms                                # Setup network (KMS mode)"
	@echo "  make setup-fabric                             # Generate network config only"
	@echo "  SIGN_KMS_ENDPOINT=192.168.1.100:9200 make setup-fabric-kms  # Custom KMS endpoint"
	@echo ""
