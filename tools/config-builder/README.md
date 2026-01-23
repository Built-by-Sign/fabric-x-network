# Fabric-X Network Configuration Builder

A CLI tool to build, configure, and manage Fabric-X networks.

## Installation

```bash
go install github.com/ethsign/fabric-x-network/tools/config-builder@latest
```

## Quick Start

### 1. Configure the Network

Edit configuration files (e.g.`configs/test-simple.yaml`) and move them into your project's repository.

### 2. Generate Network Configurations

```bash
fabric-x-network/tools/config-builder setup -c config.yaml -o ./out
```

This command will:

- Initialize HSM tokens
- Generate certificate material (cryptogen)
- Generate shared configuration (armageddon)
- Generate genesis block (configtxgen)
- Generate node configuration file

### 3. Generate Docker Compose Configuration

```bash
fabric-x-network/tools/config-builder gen-compose -c config.yaml -o ./out
```

## Command Description

### Setup

Generate all network configuration files and docker-compose.yaml.

```bash
fabric-x-network/tools/config-builder setup -c <config-file> -o <output-dir>
```

Parameters:

- `-c, --config`: Network configuration file path (default: `network.yaml`)
- `-o, --output`: Output directory (default:`./out`)
- `-v, --verbose`: Enable verbose output

## Things to Note:

1. **First Start**: Need to ensure that the Docker image is available. If using `:local` tag, you need to build a local image first.

2. **HSM Configuration**: Make sure SoftHSM2 is installed and configured correctly.

3. **Port Conflict**: Ensure that the configured port is not occupied by other services.

4. **Output directory**: `-o` The directory specified by the parameter must match `setup`. The directory used is the same.

## Troubleshooting

### HSM Errors

Make sure SoftHSM2 is configured correctly:

```bash
cat /tmp/softhsm2.conf
softhsm2-util --show-slots
```
