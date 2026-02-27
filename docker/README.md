# Docker 镜像构建说明

## 镜像概述

本目录包含用于构建 Fabric-X 网络组件的 Dockerfile。

### 镜像类型

#### 1. Orderer 镜像 (`Dockerfile.orderer`)

- **用途**：构建 Orderer 节点镜像
- **组件**：Router, Batcher, Consenter, Assembler
- **构建方式**：
  ```bash
  make build-orderer
  ```

#### 2. Committer 镜像

- **用途**：运行 Committer 节点
- **组件**：Validator, Verifier, Coordinator, Sidecar, Query Service
- **镜像来源**：直接使用官方镜像
  ```bash
  docker pull hyperledger/fabric-x-committer:0.1.5
  ```


## 快速开始

1. **配置环境变量**

   编辑配置文件设置镜像：
   ```bash
   # Orderer 镜像
   DOCKER_ORDERER_IMAGE=your-registry/fabric-x-orderer:latest

   # Committer 使用官方镜像
   DOCKER_COMMITTER_IMAGE=hyperledger/fabric-x-committer:0.1.5
   ```

2. **构建 Orderer 镜像**

   ```bash
   make build-orderer
   ```

3. **启动网络**

   ```bash
   make setup
   make start
   ```

## 组件功能对比

| 组件类型 | 需要签名 | 镜像来源 |
|---------|---------|---------|
| **Orderer** | 是 | 自定义构建 |
| Router | 是 | 自定义构建 |
| Batcher | 是 | 自定义构建 |
| Consenter | 是 | 自定义构建 |
| Assembler | 是 | 自定义构建 |
| **Committer** | 否 | 官方镜像 |
| Validator | 否 | 官方镜像 |
| Verifier | 否 | 官方镜像 |
| Coordinator | 否 | 官方镜像 |
| Sidecar | 否 | 官方镜像 |
| Query Service | 否 | 官方镜像 |

## 故障排查

### 问题：Committer 启动失败

**检查镜像配置**：
```bash
# 查看 docker-compose.yaml 中的镜像配置
grep "committer.*image:" docker-compose.yaml
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
