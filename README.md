# Fabric-X Network

> Enterprise-grade Hyperledger Fabric network deployment toolkit with configuration-driven automation

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Fabric](https://img.shields.io/badge/Hyperledger%20Fabric-2.5+-green.svg)](https://www.hyperledger.org/use/fabric)

## Overview

Fabric-X Network is a production-ready deployment framework for Hyperledger Fabric that eliminates the complexity of blockchain network setup. Define your network topology in YAML, and let the toolkit handle certificate generation, configuration management, and Docker orchestration.

### Key Features

- **Configuration-Driven** - Define entire network topology in declarative YAML
- **One-Command Deployment** - From zero to running network in seconds
- **Production-Ready** - Enterprise-grade certificate management with cryptogen
- **Docker Native** - Seamless integration with Docker Compose
- **Namespace Management** - Built-in support for Fabric-X namespace operations

## Quick Start

### Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- 4GB RAM minimum

### Launch Your Network

```bash
# One-click deployment
make quickstart
```

That's it! Your Fabric network is now running with:
- Multi-organization setup
- Orderer nodes (Router, Batcher, Consenter, Assembler)
- Peer nodes with committer sidecars
- Namespace created and ready

### Available Commands

```bash
make quickstart    # Full deployment: teardown → setup → start → create namespace
make restart       # Quick restart: stop → start (keeps config)
make setup         # Generate network configuration
make start         # Start the network
make stop          # Stop the network
make teardown      # Stop and remove all volumes
make clean         # Remove generated artifacts
make create-ns     # Create namespace
make list-ns       # List namespaces
make help          # Show all commands
```

## Architecture

### Network Topology

```
┌─────────────────────────────────────────────────────────┐
│                    Fabric-X Network                      │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Orderer    │  │   Orderer    │  │   Orderer    │  │
│  │   Org 1-4    │  │   Org 1-4    │  │   Org 1-4    │  │
│  │              │  │              │  │              │  │
│  │  • Router    │  │  • Batcher   │  │  • Consenter │  │
│  │  • Assembler │  │              │  │              │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
│                                                           │
│  ┌──────────────────────────────────────────────────┐   │
│  │              Peer Organizations                   │   │
│  │                                                    │   │
│  │  ┌─────────────┐                                 │   │
│  │  │   Org1MSP   │                                 │   │
│  │  │             │                                 │   │
│  │  │  • Peer0    │                                 │   │
│  │  │  • Committer│                                 │   │
│  │  └─────────────┘                                 │   │
│  └──────────────────────────────────────────────────┘   │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

### Components

- **Orderer Nodes**: Consensus and transaction ordering
  - Router: Message routing and distribution
  - Batcher: Transaction batching
  - Consenter: Consensus participation
  - Assembler: Block assembly
- **Peer Nodes**: Ledger maintenance and chaincode execution
- **Committer**: Transaction validation and ledger updates
- **Namespace**: Fabric-X namespace for token operations

## Configuration

### Network Definition

Edit `configs/test-full.yaml` to customize your network:

```yaml
# Network configuration
channel_id: arma
output_dir: ./out

# Docker images
docker:
  orderer_image: hyperledger/fabric-x-orderer:0.0.19
  committer_image: hyperledger/fabric-x-committer:0.1.5

# Orderer organizations
orderer_orgs:
  - name: OrdererOrg1
    domain: ordererorg1.example.com
    orderers:
      - name: orderer-router-1
        type: router
        port: 7050

# Peer organizations
peer_orgs:
  - name: Org1MSP
    domain: org1.example.com
    peers:
      - name: peer0
    users:
      - name: Admin
      - name: User1
```

### Environment Variables

```bash
# Override default Docker image
export DOCKER_TOOLS_IMAGE=ghcr.io/built-by-sign/fabric-x-tool:v0.0.5

# Use podman instead of docker
export CONTAINER_CLI=podman
```

## Advanced Usage

### Custom Network Topology

Create your own configuration file:

```bash
cp configs/test-full.yaml configs/my-network.yaml
# Edit configs/my-network.yaml
make setup CONFIG=configs/my-network.yaml
```

### Namespace Operations

```bash
# Create namespace
make create-ns

# List all namespaces
make list-ns

# Custom namespace creation
docker run --rm --network host \
  ghcr.io/built-by-sign/fabric-x-tool:v0.0.4 \
  fxconfig namespace create my-namespace \
  --channel=arma \
  --orderer=localhost:7050 \
  --mspID=Org1MSP
```

## Troubleshooting

### Committer State Mismatch

**Symptom**: Sidecar logs show state mismatch errors

**Solution**:
```bash
# Stop committer services
docker compose -f ./out/docker-compose.yaml stop committer-*

# Clear sidecar ledger
rm -rf ./out/local-deployment/committer-sidecar/config/ledger/*

# Truncate database
docker exec committer-db psql -U sc_user -d sc_db \
  -c "TRUNCATE TABLE tx_status, ns_cbdc, ns__config, ns__meta, metadata CASCADE;"

# Restart services
docker compose -f ./out/docker-compose.yaml start committer-*
```

### Network Reset

For a complete clean slate:

```bash
make teardown  # Removes all containers and volumes
make clean     # Removes generated configurations
make quickstart  # Fresh deployment
```

### Port Conflicts

If ports are already in use, modify the port mappings in `configs/test-full.yaml`:

```yaml
orderers:
  - name: orderer-router-1
    port: 8050  # Changed from 7050
```

## Project Structure

```
fabric-x-network/
├── configs/              # Network configuration files
│   └── test-full.yaml   # Default network topology
├── docs/                # Documentation
│   ├── plans/          # Implementation plans
│   └── *.md            # Guides and references
├── out/                 # Generated artifacts (gitignored)
│   ├── build/          # Certificates and configs
│   └── docker-compose.yaml
├── Makefile            # Build automation
└── README.md           # This file
```

## Contributing

We welcome contributions! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Setup

```bash
# Clone the repository
git clone https://github.com/Built-by-Sign/fabric-x-network.git
cd fabric-x-network

# Run tests
make quickstart
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built on [Hyperledger Fabric](https://www.hyperledger.org/use/fabric)
- Powered by [Fabric-X](https://github.com/built-by-sign/fabric-x)
- Certificate management via cryptogen

## Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/Built-by-Sign/fabric-x-network/issues)
- 💬 [Discussions](https://github.com/Built-by-Sign/fabric-x-network/discussions)

---

**Made with ❤️ for the Hyperledger community**
