# Fabric-X Network

A Fabric-X network deployment configuration for CBDC applications.

## Prerequisites

- **Go 1.24.3+** — builds local tools (config-builder, fxconfig) when you run Make targets
- **Docker & Docker Compose** — runs the network and the tooling image `docker.io/hyperledger/fabric-x-tools:0.0.4`
- **Optional**: If Docker is unavailable, `cryptogen` can fall back to a local Go build

## Quick Start

```bash
make setup                        # Build local tools, generate configs & certificates
make start                        # Start the network
make create-ns                    # Create the fabric-x namespace
# Override the config file if desired
make setup CONFIG=./configs/your-config.yaml
```

## Available Commands

| Command          | Description                                                                                 |
| ---------------- | ------------------------------------------------------------------------------------------- |
| `make setup`     | Install config-builder, generate network configuration and certificates (output to `./out`) |
| `make start`     | Start the Fabric-X network using docker compose                                             |
| `make stop`      | Stop the network (preserves data)                                                           |
| `make teardown`  | Stop network and remove all containers and volumes                                          |
| `make clean`     | Remove all generated artifacts in `./out` directory                                         |
| `make create-ns` | Create the `fabric_x` namespace in the network                                              |
| `make list-ns`   | List all namespaces                                                                         |
| `make restart`   | Teardown and restart the network                                                            |
| `make help`      | Show all available commands                                                                 |

## Configuration

The network configuration is defined in [configs/test-simple.yaml](configs/test-simple.yaml). This file specifies:

- Organization structure
- Peer and orderer configuration
- Channel settings
- HSM settings (if enabled)

To modify the network topology, edit this file before running `make setup`.

### Custom Config (Make Variable)

You can override the default config at runtime using the `CONFIG` variable. If not provided, it defaults to `configs/test-simple.yaml`.

Examples:

```bash
# Use a different config file for full setup
make setup CONFIG=./configs/your-config.yaml

# Generate docker-compose for a specific config
make gen-compose CONFIG=./configs/your-config.yaml
```

## Dependencies

### config-builder (local)

Built from the local source at `tools/config-builder` by the Make targets (no external download). Produces:

- Crypto materials
- Docker Compose configs
- Genesis block and channel configs

### fxconfig (local)

Built from the local source at `tools/fxconfig` by the Make targets. Used for namespace management commands like `make create-ns`.

### cryptogen (container-first)

By default, `make setup` uses the Docker image `docker.io/hyperledger/fabric-x-tools:0.0.4` to run `cryptogen` and related tooling. If Docker is unavailable, it falls back to building `cryptogen` locally via Go.

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
docker exec committer-db psql -U sc_user -d sc_db -c "TRUNCATE TABLE tx_status, ns_fabric_x, ns__config, ns__meta, metadata CASCADE;"

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

## Other

## Directory Structure

```
fabric-x-network/
├── Makefile              # Build and deployment commands
├── README.md
├── configs/              # Network configuration files
│   └── test-simple.yaml
├── out/                  # Generated artifacts (gitignored)
│   ├── docker-compose.yaml
│   ├── build/
│   │   └── config/       # Certificates and configs
│   └── cli/
├── scripts/              # Setup and utility scripts
└── tools/                # Custom tooling
```
