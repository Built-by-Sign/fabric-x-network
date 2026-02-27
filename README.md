# fabric-network

## 概述

cbdc-network 是一个基于 Hyperledger Fabric 的区块链网络配置和部署工具，支持灵活的网络拓扑配置和企业级密钥管理。

### 主要特性

- **配置驱动**：通过 YAML 配置文件定义网络拓扑
- **自动化部署**：一键生成配置、证书和 Docker Compose 文件

## 快速开始

### 标准模式（cryptogen）

```bash
# 1. Copy fxconfig to this folder: See 如何构建 fxconfig tool

make setup # 2. Setup
make start # 3. Run
make create-ns # 4. Create namespace
```

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
