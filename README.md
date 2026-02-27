# fabric-network

## 概述

cbdc-network 是一个基于 Hyperledger Fabric 的区块链网络配置和部署工具，支持灵活的网络拓扑配置和企业级密钥管理。

### 主要特性

- **配置驱动**：通过 YAML 配置文件定义网络拓扑
- **KMS 集成**：支持远程 HSM 密钥管理服务（与 cbdc-biz 一致）
- **多种部署模式**：支持 cryptogen 和 KMS 两种证书生成方式
- **自动化部署**：一键生成配置、证书和 Docker Compose 文件

## 快速开始

### 标准模式（cryptogen）

```bash
# 1. Copy fxconfig to this folder: See 如何构建 fxconfig tool

make setup # 2. Setup
make start # 3. Run
make create-ns # 4. Create namespace
```

### KMS 模式

使用 KMS 进行密钥管理，提供企业级安全性：

```bash
# 1. 准备 KMS 服务（参考 cbdc-biz 项目）
cd ../cbdc-biz
make build-kms
make run-kms

# 2. 配置网络（使用 KMS 配置）
cd ../cbdc-network
cp config-builder/configs/test-full-kms.yaml config-builder/configs/my-network.yaml
# 编辑 my-network.yaml，配置 KMS endpoint 等参数

# 3. 生成配置和证书
make setup CONFIG=my-network.yaml

# 4. 启动网络
make start

# 5. 创建命名空间
make create-ns
```

**详细文档**：[KMS 集成指南](docs/KMS_INTEGRATION.md)

## 可用命令

```bash
# 设置 fabric 环境，生成配置、证书等，输出到 out 目录
make setup

# 运行网络
make start

# Close network and remove images
make teardown

# Remove certifications（Don't do this if you want reuse network）
make clean

# create namespace - 请参考如何构建 fxconfig，先将 fxconfig 生成并拷贝到当前目录
make create-ns

# list namepace
make list-ns

# 构建支持 KMS 的 Docker 镜像（可选）
make build-orderer-kms
make build-peer-kms
```

## Troubleshooting

### Committer Sidecar State Mismatch

**Symptom**: Sidecar logs show error:
```
failed to recover the ledger store: committer should have the status of txID [...] but it does not
```

**Cause**: The sidecar's file-based ledger and committer-db are out of sync. This happens when:
- Running `make setup` after a previous run (deletes sidecar ledger but not committer-db volume)
- Manually resetting only one component

**Fix**:
```bash
# Stop committer services
docker compose -f ./out/docker-compose.yaml stop committer-sidecar committer-coordinator committer-verifier committer-validator committer-query-service

# Clear sidecar ledger
rm -rf ./out/local-deployment/committer-sidecar/config/ledger/*

# Truncate committer-db
docker exec committer-db psql -U sc_user -d sc_db -c "TRUNCATE TABLE tx_status, ns_cbdc, ns__config, ns__meta, metadata CASCADE;"

# Restart
docker compose -f ./out/docker-compose.yaml start committer-validator committer-verifier committer-query-service committer-coordinator committer-sidecar
```

### Clean Restart (Full Reset)

To completely reset the network including all persistent data:
```bash
docker compose -f ./out/docker-compose.yaml down -v  # Removes containers AND volumes
make setup
make start
make create-ns
```

## KMS 集成

cbdc-network 支持通过 KMS（Key Management Service）进行企业级密钥管理，与 cbdc-biz 项目保持一致的架构设计。

### KMS 模式优势

- **安全性增强**：私钥存储在远程 HSM 中，永不离开安全边界
- **集中管理**：统一的密钥管理服务，便于审计和合规
- **高可用性**：KMS 服务可以提供冗余和备份机制
- **与 cbdc-biz 一致**：使用相同的 libkms_pkcs11.so 库和 runtime-base 镜像

### 配置示例

```yaml
# KMS Configuration
kms:
  enabled: true
  endpoint: 'kms.example.com:9443'
  token_label: 'FabricToken'
  ca_url: 'https://ca.example.com:7054'

# 组织配置
orderer_orgs:
  - name: OrdererOrg1
    domain: ordererorg1.example.com
    hsm_token_label: 'FabricToken-OrdererOrg1'
    orderers:
      - name: orderer-router-1
        type: router
        port: 7050
        user_pin: '1001'
```

### 详细文档

完整的 KMS 集成文档请参考：[KMS_INTEGRATION.md](docs/KMS_INTEGRATION.md)

包含以下内容：
- 架构说明和证书生成流程
- 详细的配置指南
- 使用步骤和验证方法
- 故障排查和最佳实践
- 与 cryptogen 模式的对比

## 其他

### 如何构建 fxconfig tool

1. 进入 fabric-x/tools/fxconfig 目录
2. 执行命令生成

```bash
# Mac 上生成 Linux amd64(x86) 可执行文件
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o fxconfig ./main.go

# Mac 上生成本地可用可执行文件
go build -o fxconfig ./main.go
```
