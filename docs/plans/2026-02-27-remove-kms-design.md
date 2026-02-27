# 移除 KMS 相关逻辑设计文档

**日期**: 2026-02-27
**作者**: Claude
**状态**: 已批准

## 1. 概述

### 1.1 目标

完全移除项目中的 KMS（Key Management Service）相关逻辑，只保留本地密钥生成（cryptogen）模式。

### 1.2 动机

- 简化项目架构，降低维护成本
- 移除对远程 HSM 和 PKCS11 的依赖
- 统一部署模式，避免配置复杂性
- 减少构建步骤，使用标准公开镜像

### 1.3 影响范围

- 文档：删除 KMS 集成文档和相关章节
- 配置：删除 KMS 配置文件和环境变量
- 构建：删除 KMS 镜像构建逻辑
- 部署：简化为单一标准模式

## 2. 架构变更

### 2.1 变更前

```
支持两种密钥生成模式：
├── 标准模式（cryptogen）
│   └── 本地文件系统生成和存储密钥
└── KMS 模式
    ├── 远程 HSM 密钥管理
    ├── PKCS11 接口
    └── 自定义 Docker 镜像构建
```

### 2.2 变更后

```
仅支持标准模式：
└── cryptogen 模式
    ├── 本地文件系统生成和存储密钥
    ├── 使用公开标准镜像
    └── 无需自定义构建
```

## 3. 详细变更清单

### 3.1 完全删除的文件

| 文件路径 | 说明 | 行数 |
|---------|------|------|
| `docs/KMS_INTEGRATION.md` | KMS 集成完整文档 | 705 |
| `configs/test-full-kms.yaml` | KMS 模式配置文件 | 204 |
| `docker/Dockerfile.orderer` | 包含 PKCS11/KMS 支持的 Dockerfile | 77 |
| `.env` | 环境变量配置（包含 KMS 配置） | 30 |
| `.env.example` | 环境变量示例文件 | 30 |

**删除原因**：
- `.env` 和 `.env.example`：配置已直接写在 `test-full.yaml` 中，不再需要环境变量
- 其他文件：完全属于 KMS 功能，删除后不影响标准模式

### 3.2 需要修改的文件

#### 3.2.1 [README.md](README.md)

**删除内容**：
- 第 10 行：KMS 集成特性说明
- 第 11 行：多种部署模式说明
- 第 26-51 行：整个 "KMS 模式" 使用章节
- 第 74-77 行：KMS 镜像构建命令
- 第 117-160 行：整个 "KMS 集成" 章节

**保留内容**：
- 标准模式（cryptogen）的所有说明
- 快速开始指南
- Troubleshooting 章节
- 其他非 KMS 相关内容

#### 3.2.2 [Makefile](Makefile)

**删除内容**：
- 第 6-10 行：`.env` 文件加载逻辑
- 第 39-41 行：Docker 环境变量传递（DOCKER_ORDERER_IMAGE 等）
- 第 48-54 行：`DOCKER_RUN_HSM` 变量定义
- 第 72-78 行：`setup-fabric-kms` target
- 第 88-94 行：`gen-compose-kms` target
- 第 112、130-144 行：`create-ns` 和 `create-ns-dev` 中的 PKCS11 参数
- 第 178-179 行：`setup-kms` target
- 第 198-215 行：`build-kms-images` 和 `build-orderer-kms` targets
- help 信息中的所有 KMS 相关说明

**简化内容**：
- `DOCKER_RUN_BASE`：移除环境变量传递
- `create-ns`：移除 `--pkcs11-library`、`--pkcs11-label`、`--pkcs11-pin` 参数

#### 3.2.3 [docs/CERTIFICATE_DEPLOYMENT_GUIDE.md](docs/CERTIFICATE_DEPLOYMENT_GUIDE.md)

**删除内容**：
- 第 132-163 行：证书生成流程图中的 KMS 部分
- 第 336-386 行：开发环境 KMS 配置示例
- 第 388-466 行：生产环境 KMS 配置示例
- 第 468-530 行：KMS 集群配置和 CA 部署策略
- 所有提到 KMS/HSM/PKCS11 的段落和图表

**保留内容**：
- 证书依赖关系基本说明
- cryptogen 模式的配置
- 通用的安全最佳实践

#### 3.2.4 [docker/README.md](docker/README.md)

**删除内容**：
- 第 10-25 行：Orderer 镜像的 KMS 支持说明
- 第 28-57 行：KMS 模式部署指南
- 第 59-83 行：镜像依赖关系图（KMS 部分）
- 第 85-99 行：组件功能对比表中的 KMS 列
- 第 101-145 行：所有 KMS 相关的优化建议和故障排查

**保留内容**：
- Committer 镜像说明
- 基本的 Docker 使用说明

## 4. 配置变更

### 4.1 镜像配置

**变更前**（KMS 模式）：
```yaml
docker:
  orderer_image: '${DOCKER_ORDERER_IMAGE}'  # 自定义 KMS 镜像
  committer_image: '${DOCKER_COMMITTER_IMAGE}'
  tools_image: '${DOCKER_TOOLS_IMAGE}'
```

**变更后**（标准模式）：
```yaml
docker:
  orderer_image: 'hyperledger/fabric-x-orderer:0.0.19'  # 公开标准镜像
  committer_image: 'hyperledger/fabric-x-committer:0.1.5'
  tools_image: 'ghcr.io/built-by-sign/fabric-x-tool:v0.0.4'
```

### 4.2 组织配置

**变更前**（包含 KMS 配置）：
```yaml
orderer_orgs:
  - name: OrdererOrg1
    domain: ordererorg1.example.com
    kms_token_label: ${KMS_TOKEN_LABEL}
    kms_user_pin: '${KMS_USER_PIN}'
```

**变更后**（纯 cryptogen）：
```yaml
orderer_orgs:
  - name: OrdererOrg1
    domain: ordererorg1.example.com
    enable_organizational_units: false
```

## 5. 构建流程变更

### 5.1 变更前

```bash
# 需要构建自定义镜像
make build-orderer-kms
make setup-kms
make start
```

### 5.2 变更后

```bash
# 直接使用公开镜像
make setup
make start
```

## 6. 影响分析

### 6.1 正面影响

- ✅ **简化部署**：无需构建自定义镜像，直接使用公开镜像
- ✅ **降低复杂度**：移除 PKCS11/HSM 依赖和配置
- ✅ **减少维护成本**：只维护一种部署模式
- ✅ **加快启动速度**：无需镜像构建步骤
- ✅ **统一配置**：所有配置集中在 YAML 文件中

### 6.2 功能变更

- ⚠️ **不再支持远程 HSM**：私钥只能存储在本地文件系统
- ⚠️ **不再支持 PKCS11**：无法使用硬件安全模块
- ⚠️ **安全级别降低**：适合开发和测试环境，生产环境需评估

### 6.3 兼容性

- ✅ **向后兼容**：标准模式（cryptogen）保持不变
- ❌ **不兼容 KMS 模式**：现有 KMS 部署需要迁移到标准模式

## 7. 迁移指南

### 7.1 从 KMS 模式迁移到标准模式

如果现有部署使用 KMS 模式，需要：

1. **备份现有配置**
   ```bash
   cp configs/test-full-kms.yaml configs/test-full-kms.yaml.backup
   ```

2. **使用标准配置**
   ```bash
   cp configs/test-full.yaml configs/my-network.yaml
   # 根据需要调整配置
   ```

3. **清理旧环境**
   ```bash
   make clean
   make teardown
   ```

4. **重新部署**
   ```bash
   make setup
   make start
   ```

### 7.2 注意事项

- 标准模式生成的证书与 KMS 模式不兼容
- 需要重新生成所有证书和配置
- 私钥将存储在本地文件系统，需要妥善保管

## 8. 测试计划

### 8.1 功能测试

- [ ] 验证 `make setup` 正常生成配置和证书
- [ ] 验证 `make start` 正常启动网络
- [ ] 验证 `make create-ns` 正常创建命名空间
- [ ] 验证所有节点正常运行
- [ ] 验证交易提交和查询功能

### 8.2 文档测试

- [ ] 验证 README 中的快速开始指南可用
- [ ] 验证所有命令示例正确
- [ ] 确认没有遗留的 KMS 引用

### 8.3 回归测试

- [ ] 确认标准模式功能未受影响
- [ ] 确认 Makefile 所有 target 正常工作
- [ ] 确认 Docker Compose 正常生成和运行

## 9. 风险评估

### 9.1 低风险

- 标准模式代码路径保持不变
- 只删除 KMS 特定代码，不影响核心功能
- 可以从 Git 历史恢复删除的代码

### 9.2 中风险

- 文档修改可能遗漏某些 KMS 引用
- Makefile 修改可能影响现有工作流

### 9.3 缓解措施

- 全面的文本搜索确保没有遗留引用
- 完整的功能测试验证
- Git 提交前仔细审查变更

## 10. 实施步骤

1. **删除文件**（5 个文件）
2. **修改 README.md**（删除 KMS 章节）
3. **修改 Makefile**（删除 KMS targets）
4. **修改 docs/CERTIFICATE_DEPLOYMENT_GUIDE.md**（删除 KMS 章节）
5. **修改 docker/README.md**（删除 KMS 说明）
6. **验证测试**（运行完整测试套件）
7. **提交变更**（创建 commit）

## 11. 验收标准

- ✅ 所有 KMS 相关文件已删除
- ✅ 所有文档中的 KMS 引用已清除
- ✅ Makefile 中的 KMS targets 已删除
- ✅ 标准模式功能正常工作
- ✅ `make setup && make start` 成功运行
- ✅ 没有遗留的环境变量或配置引用
- ✅ 文档清晰准确，没有死链接

## 12. 后续工作

- 更新项目 CHANGELOG
- 通知相关团队成员
- 更新部署文档和运维手册
