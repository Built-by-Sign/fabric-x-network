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
CONFIG_BUILDER_DIR := $(PROJECT_DIR)/tools/config-builder
CONFIG_BUILDER_BIN := $(CONFIG_BUILDER_DIR)/build/cli/config-builder
CONFIG_BUILDER_CONFIG ?= $(PROJECT_DIR)/configs/test-simple.yaml
CONFIG ?= $(CONFIG_BUILDER_CONFIG)
FXCONFIG_DIR := $(PROJECT_DIR)/tools/fxconfig
FXCONFIG_BIN := $(FXCONFIG_DIR)/build/cli/fxconfig
CRYPTOGEN_BIN := $(shell go env GOPATH)/bin/cryptogen
TOOLS_IMAGE ?= docker.io/hyperledger/fabric-x-tools:0.0.4
TOOLS_DEPS := build-fxconfig build-config-builder download-cryptogen

# Build config-builder binary from local source
.PHONY: build-config-builder
build-config-builder:
	@echo "Building config-builder from local source..."
	@cd $(CONFIG_BUILDER_DIR) && $(MAKE) build
# Ensure config-builder is built from local source before using it
$(CONFIG_BUILDER_BIN): build-config-builder
# Build fxconfig binary from local source
.PHONY: build-fxconfig
build-fxconfig:
	@echo "Building fxconfig from local source..."
	@cd $(FXCONFIG_DIR) && $(MAKE) build

# Ensure fxconfig is built from local source before using it
$(FXCONFIG_BIN): build-fxconfig

# Build cryptogen binary (external repo)
.PHONY: download-cryptogen
download-cryptogen:
	@if command -v $(CONTAINER_CLI) >/dev/null 2>&1; then \
		if $(CONTAINER_CLI) image inspect $(TOOLS_IMAGE) >/dev/null 2>&1; then \
			echo "Docker tools image available ($(TOOLS_IMAGE)); skipping cryptogen install"; \
		else \
			echo "Docker found; tools image not local. It will be pulled when needed (skipping cryptogen install)"; \
		fi; \
	elif command -v cryptogen >/dev/null 2>&1; then \
		echo "Using existing local cryptogen binary"; \
	else \
		echo "Docker not available; installing cryptogen (Go fallback)..."; \
		go install github.com/ethsign/cryptogen@latest; \
	fi

$(CRYPTOGEN_BIN): download-cryptogen

# Build all the artifacts, the binaries and transfer them to the remote hosts (e.g. make setup).
.PHONY: setup-fabric
setup-fabric: $(TOOLS_DEPS)
	@echo "Using fabric-x-network/tools/config-builder to setup network..."
	@echo "$(CONFIG)..."
	@mkdir -p $(OUTPUT_DIR)
	@$(CONFIG_BUILDER_BIN) setup \
	  -c $(CONFIG) \
	  -o $(OUTPUT_DIR) -v

# Generate docker-compose.yaml file
.PHONY: gen-compose
gen-compose: build-fxconfig build-config-builder
	@echo "Generating docker-compose.yaml..."
	@$(CONFIG_BUILDER_BIN) gen-compose \
	  -c $(CONFIG) \
	  -o $(OUTPUT_DIR) -v

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
create-ns: build-fxconfig
	@echo "Creating namespace..."
	$(FXCONFIG_BIN) namespace create fabric_x \
		--channel=arma \
		--orderer=localhost:7050 \
		--mspID=Org1MSP \
		--mspConfigPath=$(OUTPUT_DIR)/build/config/cryptogen-artifacts/crypto/peerOrganizations/org1.example.com/users/channel_admin@org1.example.com/msp \
		--pk=$(OUTPUT_DIR)/build/config/cryptogen-artifacts/crypto/peerOrganizations/org1.example.com/users/endorser@org1.example.com/msp/signcerts/endorser@org1.example.com-cert.pem \
		--connTimeout=60s
	@until $(FXCONFIG_BIN) namespace list --endpoint=localhost:5500 | grep -q fabric_x; do \
		sleep 2; \
		echo "waiting for namespace to be created..."; \
	done
	$(FXCONFIG_BIN) namespace list --endpoint=localhost:5500

# List namespaces
.PHONY: list-ns
list-ns: build-fxconfig
	$(FXCONFIG_BIN) namespace list --endpoint=localhost:5500

# Stop the targeted hosts (e.g. make fabric-x stop).
.PHONY: stop-fabric
stop-fabric:
	@echo "Stopping network using docker compose..."
	@cd $(OUTPUT_DIR) && $(CONTAINER_CLI) compose stop

# Teardown the targeted hosts (e.g. make fabric-x teardown).
.PHONY: teardown-fabric
teardown-fabric:
	@echo "Teardown network using docker compose..."
	@cd $(OUTPUT_DIR) && $(CONTAINER_CLI) compose down -v
	@$(CONTAINER_CLI) network inspect fabric_x_net >/dev/null 2>&1 && $(CONTAINER_CLI) network rm fabric_x_net || true

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
	@awk ' \
		/^#/ { \
			sub(/^#[ \t]*/, "", $$0); \
			help_msg = $$0; \
		} \
		/^[a-zA-Z0-9][^ :]*:/ { \
			if (help_msg) { \
				split($$1, target, ":"); \
				printf "  %-40s %s\n", target[1], help_msg; \
				help_msg = ""; \
			} \
		} \
	' $(MAKEFILE_LIST)