# 移除 KMS 相关逻辑实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标:** 完全移除项目中的 KMS（Key Management Service）相关逻辑，只保留本地密钥生成（cryptogen）模式

**架构:** 删除所有 KMS 特定的文件、配置和文档，简化 Makefile 移除 KMS targets，更新所有文档移除 KMS 引用，确保标准模式（cryptogen）正常工作

**技术栈:** Makefile, YAML 配置, Markdown 文档, Docker

---

## Task 1: 删除 KMS 专用文件

**文件:**
- Delete: `docs/KMS_INTEGRATION.md`
- Delete: `configs/test-full-kms.yaml`
- Delete: `docker/Dockerfile.orderer`
- Delete: `.env`
- Delete: `.env.example`

**Step 1: 删除 KMS 集成文档**

```bash
rm docs/KMS_INTEGRATION.md
```

预期: 文件被删除

**Step 2: 删除 KMS 配置文件**

```bash
rm configs/test-full-kms.yaml
```

预期: 文件被删除

**Step 3: 删除包含 KMS 支持的 Dockerfile**

```bash
rm docker/Dockerfile.orderer
```

预期: 文件被删除

**Step 4: 删除环境变量配置文件**

```bash
rm .env .env.example
```

预期: 两个文件被删除

**Step 5: 验证文件已删除**

```bash
ls -la docs/KMS_INTEGRATION.md configs/test-full-kms.yaml docker/Dockerfile.orderer .env .env.example 2>&1
```

预期: 所有文件显示 "No such file or directory"

**Step 6: 提交删除**

```bash
git add -A
git commit -m "chore: remove KMS-specific files"
```

---

## Task 2: 更新 README.md

**文件:**
- Modify: `README.md`

**Step 1: 删除 KMS 特性说明（第 10-11 行）**

在 README.md 中找到：
```markdown
- **KMS 集成**：支持远程 HSM 密钥管理服务（与 cbdc-biz 一致）
- **多种部署模式**：支持 cryptogen 和 KMS 两种证书生成方式
```

删除这两行，保留：
```markdown
- **配置驱动**：通过 YAML 配置文件定义网络拓扑
- **自动化部署**：一键生成配置、证书和 Docker Compose 文件
```

**Step 2: 删除 KMS 模式章节（第 26-51 行）**

删除整个 "### KMS 模式" 章节，包括所有子内容

**Step 3: 删除 KMS 构建命令（第 74-77 行）**

在 "## 可用命令" 章节中，删除：
```markdown
# 构建支持 KMS 的 Docker 镜像（可选）
make build-orderer-kms
make build-peer-kms
```

**Step 4: 删除 KMS 集成章节（第 117-160 行）**

删除整个 "## KMS 集成" 章节及其所有子章节

**Step 5: 验证修改**

```bash
grep -i "kms\|hsm\|pkcs11" README.md
```

预期: 无输出（所有 KMS 引用已删除）

**Step 6: 提交更改**

```bash
git add README.md
git commit -m "docs: remove KMS references from README"
```

---

## Task 3: 更新 Makefile（第一部分：删除环境变量和 HSM 配置）

**文件:**
- Modify: `Makefile`

**Step 1: 删除 .env 文件加载逻辑（第 6-10 行）**

删除：
```makefile
# 读取 .env 文件
ifneq (,$(wildcard .env))
    include .env
    export
endif
```

**Step 2: 删除 Docker 环境变量传递（第 39-41 行）**

在 `DOCKER_RUN_BASE` 定义中，删除：
```makefile
	-e DOCKER_ORDERER_IMAGE \
	-e DOCKER_COMMITTER_IMAGE \
	-e DOCKER_TOOLS_IMAGE
```

修改后的 `DOCKER_RUN_BASE` 应该是：
```makefile
DOCKER_RUN_BASE = $(CONTAINER_CLI) run --rm \
	--user $(shell id -u):$(shell id -g) \
	-v $(PROJECT_DIR):/workspace \
	-w /workspace \
	-e HOME=/tmp
```

**Step 3: 删除 DOCKER_RUN_HSM 定义（第 48-54 行）**

删除整个 `DOCKER_RUN_HSM` 变量定义：
```makefile
# 带 HSM/PKCS11 支持的 Docker 运行命令
# 挂载 HSM 库目录，并设置相关环境变量
DOCKER_RUN_HSM = $(DOCKER_RUN_NET) \
	-e SIGN_KMS_ENDPOINT \
	-e KMS_TOKEN_LABEL \
	-e KMS_USER_PIN \
	-e CA_URL
```

**Step 4: 提交更改**

```bash
git add Makefile
git commit -m "refactor: remove .env loading and HSM variables from Makefile"
```

---

## Task 4: 更新 Makefile（第二部分：删除 KMS targets）

**文件:**
- Modify: `Makefile`

**Step 1: 删除 setup-fabric-kms target（第 72-78 行）**

删除：
```makefile
# Build all the artifacts with KMS configuration using cbdc-tool Docker image
.PHONY: setup-fabric-kms
setup-fabric-kms:
	@echo "==> Generating Fabric network configuration with KMS using cbdc-tool..."
	$(DOCKER_RUN_HSM) $(DOCKER_TOOLS_IMAGE) \
		config-builder setup -c configs/test-full-kms.yaml -o ./out --use-local-tools --log-level=info
	@echo "==> Network configuration with KMS generated in ./out"
```

**Step 2: 删除 gen-compose-kms target（第 88-94 行）**

删除：
```makefile
# Generate docker-compose.yaml file with KMS configuration using cbdc-tool Docker image
.PHONY: gen-compose-kms
gen-compose-kms:
	@echo "==> Generating docker-compose.yaml with KMS using cbdc-tool..."
	$(DOCKER_RUN_HSM) $(DOCKER_TOOLS_IMAGE) \
		config-builder gen-compose -c configs/test-full-kms.yaml -o ./out --use-local-tools
	@echo "==> docker-compose.yaml with KMS generated in ./out"
```

**Step 3: 删除 setup-kms target（第 178-179 行）**

删除：
```makefile
# Build all the artifacts and binaries with KMS configuration
.PHONY: setup-kms
setup-kms: clean setup-fabric-kms gen-compose-kms
```

**Step 4: 删除 build-kms-images 和 build-orderer-kms targets（第 198-215 行）**

删除：
```makefile
# Build all KMS-enabled images
.PHONY: build-kms-images
build-kms-images: build-orderer-kms
	@echo "Building all KMS-enabled images..."

# Build KMS-enabled orderer Docker image
# Orderer 需要 PKCS11/KMS 支持用于签名操作
# 依赖 DOCKER_TOOLS_IMAGE 提供 KMS 运行时库
.PHONY: build-orderer-kms
build-orderer-kms:
	@echo "Building KMS-enabled orderer image..."
	@echo "Checking DOCKER_TOOLS_IMAGE dependency: $(DOCKER_TOOLS_IMAGE)"
	@$(CONTAINER_CLI) image inspect $(DOCKER_TOOLS_IMAGE) >/dev/null 2>&1 || \
		(echo "Error: $(DOCKER_TOOLS_IMAGE) not found. Please build or pull it first." && exit 1)
	@cd $(DOCKER_DIR) && $(CONTAINER_CLI) build --platform=$(PLATFORM) \
		--build-arg DOCKER_TOOLS_IMAGE=$(DOCKER_TOOLS_IMAGE) \
		-t $(IMAGE_PREFIX)/cbdc-orderer-kms:$(TAG) \
		-f Dockerfile.orderer .
	@echo "Orderer KMS image built: $(IMAGE_PREFIX)/cbdc-orderer-kms:$(TAG)"
```

**Step 5: 提交更改**

```bash
git add Makefile
git commit -m "refactor: remove KMS-specific targets from Makefile"
```

---

## Task 5: 更新 Makefile（第三部分：简化 create-ns 命令）

**文件:**
- Modify: `Makefile`

**Step 1: 简化 create-ns target（第 110-123 行）**

将：
```makefile
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
```

修改为（使用 DOCKER_RUN_NET 替代 DOCKER_RUN_HSM）：
```makefile
create-ns:
	@echo "Creating namespace using Docker..."
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
```

**Step 2: 删除 create-ns-dev target（第 128-144 行）**

删除整个 create-ns-dev target（包含 PKCS11 参数）：
```makefile
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
```

**Step 3: 提交更改**

```bash
git add Makefile
git commit -m "refactor: simplify create-ns and remove create-ns-dev"
```

---

## Task 6: 更新 Makefile（第四部分：更新 help 信息）

**文件:**
- Modify: `Makefile`

**Step 1: 更新 help target（第 220-260 行）**

删除所有 KMS 相关的帮助信息，修改为：

```makefile
# Print the list of supported commands.
.PHONY: help
help:
	@echo "CBDC Network Makefile - Available Targets"
	@echo "=========================================="
	@echo ""
	@echo "Network Configuration:"
	@echo "  setup-fabric              - Generate Fabric network config using cbdc-tool"
	@echo "  gen-compose               - Generate docker-compose.yaml using cbdc-tool"
	@echo ""
	@echo "Network Operations:"
	@echo "  setup                     - Clean and setup network"
	@echo "  start-fabric              - Start Fabric network using docker compose"
	@echo "  stop-fabric               - Stop Fabric network"
	@echo "  teardown-fabric           - Teardown Fabric network and remove volumes"
	@echo "  restart-fabric            - Restart Fabric network"
	@echo "  create-ns                 - Create namespace in Fabric-X"
	@echo "  list-ns                   - List namespaces in Fabric-X"
	@echo ""
	@echo "Cleanup:"
	@echo "  clean-fabric              - Remove all generated artifacts"
	@echo "  clean                     - Alias for clean-fabric"
	@echo ""
	@echo "Examples:"
	@echo "  make setup                                    # Setup network"
	@echo "  make start                                    # Start network"
	@echo "  make create-ns                                # Create namespace"
	@echo ""
```

**Step 2: 验证 Makefile 语法**

```bash
make -n help
```

预期: 显示帮助信息，无语法错误

**Step 3: 提交更改**

```bash
git add Makefile
git commit -m "docs: update Makefile help to remove KMS references"
```

---

## Task 7: 更新 docs/CERTIFICATE_DEPLOYMENT_GUIDE.md（第一部分）

**文件:**
- Modify: `docs/CERTIFICATE_DEPLOYMENT_GUIDE.md`

**Step 1: 删除证书生成流程图中的 KMS 部分（第 132-163 行）**

找到 "## 3. 证书依赖关系详细图" 章节，删除整个 mermaid 图表中的 KMS 相关节点：

删除：
```markdown
         KMS["KMS/HSM<br/>私钥存储"]

         KMS -->|1. 生成密钥对| KeyPair["公私钥对"]
```

以及：
```markdown
         KMS -->|PKCS#11| Node
```

保留 CA 和证书生成的基本流程

**Step 2: 删除开发环境 KMS 配置示例（第 336-386 行）**

删除 "### 7.1 开发环境配置 (test-full-kms.yaml 示例)" 整个章节

**Step 3: 提交更改**

```bash
git add docs/CERTIFICATE_DEPLOYMENT_GUIDE.md
git commit -m "docs: remove KMS flow diagrams from certificate guide (part 1)"
```

---

## Task 8: 更新 docs/CERTIFICATE_DEPLOYMENT_GUIDE.md（第二部分）

**文件:**
- Modify: `docs/CERTIFICATE_DEPLOYMENT_GUIDE.md`

**Step 1: 删除生产环境 KMS 配置（第 388-466 行）**

删除 "### 7.2 生产环境配置 (推荐)" 章节中所有 KMS 相关配置示例

**Step 2: 删除 KMS 集群配置（第 468-530 行）**

删除以下章节：
- "### 7.3 KMS 集群配置"
- "### 7.4 CA 服务部署策略"

**Step 3: 清理安全最佳实践章节**

在 "## 8. 安全最佳实践" 中，删除所有 HSM 和 KMS 特定的建议，保留通用的安全实践

**Step 4: 验证文档中没有 KMS 引用**

```bash
grep -i "kms\|hsm\|pkcs11" docs/CERTIFICATE_DEPLOYMENT_GUIDE.md
```

预期: 无输出或仅有历史性提及

**Step 5: 提交更改**

```bash
git add docs/CERTIFICATE_DEPLOYMENT_GUIDE.md
git commit -m "docs: remove KMS configuration examples from certificate guide (part 2)"
```

---

## Task 9: 更新 docker/README.md

**文件:**
- Modify: `docker/README.md`

**Step 1: 删除 Orderer KMS 支持说明（第 10-25 行）**

删除 "#### 1. Orderer 镜像 (`Dockerfile.orderer`)" 章节中的 KMS 相关内容

**Step 2: 删除 KMS 模式部署指南（第 28-57 行）**

删除 "### 标准 KMS 模式部署" 整个章节

**Step 3: 删除镜像依赖关系图（第 59-83 行）**

删除包含 RUNTIME_BASE_IMAGE 和 KMS 的依赖关系图

**Step 4: 简化组件功能对比表（第 85-99 行）**

删除表格中的 "需要 PKCS11" 列，只保留：
- 组件类型
- 需要签名
- 镜像来源

**Step 5: 删除 KMS 优化建议（第 101-145 行）**

删除所有 KMS 相关的优化建议和故障排查章节

**Step 6: 验证文档**

```bash
grep -i "kms\|hsm\|pkcs11" docker/README.md
```

预期: 无输出

**Step 7: 提交更改**

```bash
git add docker/README.md
git commit -m "docs: remove KMS references from docker README"
```

---

## Task 10: 验证和测试

**Step 1: 检查是否有遗留的 KMS 引用**

```bash
grep -r -i "kms\|hsm\|pkcs11" --include="*.md" --include="*.yaml" --include="Makefile" .
```

预期: 仅在 `configs/test-full.yaml` 中可能有注释，或在 git 历史中

**Step 2: 验证 Makefile 语法**

```bash
make -n setup
make -n start
make -n clean
```

预期: 所有命令都能正确解析，无语法错误

**Step 3: 测试标准模式配置生成**

```bash
make clean
make setup
```

预期:
- 成功生成配置文件到 `./out` 目录
- 无 KMS 相关错误
- 生成的配置使用 cryptogen

**Step 4: 验证生成的配置**

```bash
ls -la ./out/build/config/cryptogen-artifacts/crypto/
```

预期: 看到 ordererOrganizations 和 peerOrganizations 目录

**Step 5: 测试网络启动**

```bash
make start
```

预期: Docker Compose 成功启动所有容器

**Step 6: 验证容器运行状态**

```bash
docker ps | grep -E "orderer|committer"
```

预期: 所有 orderer 和 committer 容器都在运行

**Step 7: 清理测试环境**

```bash
make teardown
make clean
```

预期: 成功清理所有容器和生成的文件

**Step 8: 提交验证结果**

```bash
git add -A
git commit -m "test: verify KMS removal and standard mode functionality"
```

---

## Task 11: 最终检查和文档更新

**Step 1: 检查 git 状态**

```bash
git status
```

预期: 所有更改已提交，工作目录干净

**Step 2: 查看提交历史**

```bash
git log --oneline -10
```

预期: 看到所有相关的提交记录

**Step 3: 创建总结提交（如果需要）**

如果有遗漏的小改动：
```bash
git add -A
git commit -m "chore: final cleanup after KMS removal"
```

**Step 4: 验证文档完整性**

检查 README.md 是否清晰描述了当前的功能：
```bash
cat README.md | head -50
```

预期: 只提到标准模式（cryptogen），无 KMS 引用

**Step 5: 更新 CHANGELOG（如果存在）**

如果项目有 CHANGELOG.md，添加条目：
```markdown
## [Unreleased]

### Removed
- KMS (Key Management Service) integration and remote HSM support
- PKCS11 support from Orderer images
- KMS-specific configuration files and environment variables
- Docker image build targets for KMS-enabled components

### Changed
- Simplified to single deployment mode using cryptogen for local key generation
- Updated all documentation to reflect cryptogen-only approach
```

**Step 6: 最终提交**

```bash
git add CHANGELOG.md
git commit -m "docs: update CHANGELOG for KMS removal"
```

---

## 完成标准

- ✅ 所有 KMS 相关文件已删除（5 个文件）
- ✅ README.md 中的 KMS 引用已清除
- ✅ Makefile 中的 KMS targets 已删除
- ✅ docs/CERTIFICATE_DEPLOYMENT_GUIDE.md 中的 KMS 章节已删除
- ✅ docker/README.md 中的 KMS 说明已删除
- ✅ 标准模式（cryptogen）功能正常工作
- ✅ `make setup && make start` 成功运行
- ✅ 没有遗留的 KMS/HSM/PKCS11 引用
- ✅ 所有更改已提交到 git

## 预计时间

- Task 1-2: 15 分钟（文件删除和 README 更新）
- Task 3-6: 30 分钟（Makefile 更新）
- Task 7-9: 25 分钟（文档更新）
- Task 10: 20 分钟（验证和测试）
- Task 11: 10 分钟（最终检查）

**总计: 约 100 分钟（1.5-2 小时）**
