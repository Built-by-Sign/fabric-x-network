# Docker 镜像构建说明

## 镜像概述

本目录包含用于构建 Fabric-X 网络组件的 Dockerfile。

### 镜像类型

#### 1. Orderer 镜像 (`Dockerfile.orderer`)

**需要 PKCS11/KMS 支持** ✅

- **用途**：构建支持 KMS 的 Orderer 节点镜像
- **组件**：Router, Batcher, Consenter, Assembler
- **为什么需要 KMS**：Orderer 节点需要对区块进行签名操作
- **构建方式**：
  ```bash
  make build-orderer-kms
  ```
- **特点**：
  - 从源码构建，启用 `-tags pkcs11` 标志
  - 应用 PKCS11 相关补丁
  - 包含 libkms_pkcs11.so 库
  - 基于 RUNTIME_BASE_IMAGE


## 快速开始

### 标准 KMS 模式部署

1. **配置环境变量**

   编辑 `cbdc-network/.env` 文件：
   ```bash
   # Orderer 需要自定义构建（包含 KMS 支持）
   DOCKER_ORDERER_IMAGE=cbdc-dev.sign.global:5001/cbdc-orderer-kms:latest

   # Committer 直接使用官方镜像（推荐）
   DOCKER_COMMITTER_IMAGE=hyperledger/fabric-x-committer:0.1.5

   # Runtime base 镜像（仅 Orderer 需要）
   RUNTIME_BASE_IMAGE=cbdc-dev.sign.global:5001/cbdc-biz-runtime-base:latest
   ```

2. **构建镜像**

   ```bash
   cd cbdc-network
   make build-kms-images
   ```

3. **启动网络**

   ```bash
   make setup-kms
   make start
   ```

## 镜像依赖关系

```
┌─────────────────────────────────────────────────────────────┐
│                    RUNTIME_BASE_IMAGE                       │
│  (包含 libkms_pkcs11.so + 系统依赖)                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ 仅 Orderer 需要
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   Orderer KMS Image                         │
│  - 从源码构建，启用 PKCS11                                   │
│  - 应用补丁                                                  │
│  - 包含签名功能                                              │
└─────────────────────────────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────┐
│          hyperledger/fabric-x-committer:0.1.5               │
│  (官方镜像，直接使用，无需构建)                              │
│  - 不需要 PKCS11                                             │
│  - 不需要签名功能                                            │
└─────────────────────────────────────────────────────────────┘
```

## 组件功能对比

| 组件类型 | 需要签名 | 需要 PKCS11 | 镜像来源 |
|---------|---------|------------|---------|
| **Orderer** | ✅ 是 | ✅ 是 | 自定义构建 |
| Router | ✅ | ✅ | 自定义构建 |
| Batcher | ✅ | ✅ | 自定义构建 |
| Consenter | ✅ | ✅ | 自定义构建 |
| Assembler | ✅ | ✅ | 自定义构建 |
| **Committer** | ❌ 否 | ❌ 否 | 官方镜像 |
| Validator | ❌ | ❌ | 官方镜像 |
| Verifier | ❌ | ❌ | 官方镜像 |
| Coordinator | ❌ | ❌ | 官方镜像 |
| Sidecar | ❌ | ❌ | 官方镜像 |
| Query Service | ❌ | ❌ | 官方镜像 |

## 优化建议

### ✅ 推荐配置（已优化）

```bash
# .env 文件
DOCKER_ORDERER_IMAGE=cbdc-dev.sign.global:5001/cbdc-orderer-kms:latest
DOCKER_COMMITTER_IMAGE=hyperledger/fabric-x-committer:0.1.5
```

**优点**：
- ✅ 减少构建时间（无需构建 Committer）
- ✅ 减少镜像大小（Committer 不包含不必要的 PKCS11 库）
- ✅ 简化维护（直接使用官方镜像）
- ✅ 更快的部署速度


## 故障排查

### 问题：Committer 启动失败

**检查镜像配置**：
```bash
# 查看 docker-compose.yaml 中的镜像配置
grep "committer.*image:" cbdc-network/out/docker-compose.yaml
```

**确认使用官方镜像**：
```bash
# 应该看到
image: hyperledger/fabric-x-committer:0.1.5
```

### 问题：找不到 Committer 镜像

**拉取官方镜像**：
```bash
docker pull hyperledger/fabric-x-committer:0.1.5
```

## 参考资料

- [Fabric-X Orderer GitHub](https://github.com/hyperledger/fabric-x-orderer)
- [Fabric-X Committer GitHub](https://github.com/hyperledger/fabric-x-committer)
- [KMS Integration Guide](../docs/KMS_INTEGRATION.md)
