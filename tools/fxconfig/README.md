# FX Config

A command-line tool for managing Fabric-X namespace configurations. This tool provides utilities for creating, listing, and updating namespace configurations in a Fabric-X network.

## Features

- **Namespace Management**: Create, list, and update namespace configurations
- **Multi-Platform Support**: Build binaries for Linux and macOS (amd64 and arm64)
- **Configuration as Code**: Manage network configurations through command-line interface
- **Orderer Integration**: Direct integration with ordering service endpoints

## Prerequisites

- Go 1.24.3 or higher
- Make
- golangci-lint (optional, for linting)

## Installation

### From Source

Clone the repository and build:

```bash
git clone https://github.com/ethsign/fxconfig.git
cd fxconfig
```

Build the binary:

```bash
make build
```

This will create the binary at `build/cli/fxconfig`.

Install to your GOPATH/bin:

```bash
make install
```

### Prebuilt Binaries

Build for multiple platforms:

```bash
make build-all
```

This creates binaries for:

- Linux: amd64, arm64
- macOS: amd64, arm64

## Usage

### Getting Help

```bash
fxconfig --help
fxconfig namespace --help
fxconfig namespace create --help
```

### Namespace Commands

#### Create a Namespace

```bash
fxconfig namespace create \
  --orderer <orderer_endpoint> \
  --cafile <path_to_ca_cert> \
  --channel <channel_name> \
  [additional flags]
```

#### List Namespaces

```bash
fxconfig namespace list \
  --orderer <orderer_endpoint> \
  --cafile <path_to_ca_cert> \
  --channel <channel_name>
```

#### Update a Namespace

```bash
fxconfig namespace update \
  --orderer <orderer_endpoint> \
  --cafile <path_to_ca_cert> \
  --channel <channel_name> \
  [additional flags]
```

### Version

Check the version of fxconfig:

```bash
fxconfig version
```

## Development

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build for specific platform
make build-linux
make build-darwin
```

### Testing

Run all tests:

```bash
make test
```

Generate coverage report:

```bash
make test-coverage
```

This generates `coverage.html` with detailed coverage information.

### Code Quality

Format code:

```bash
make fmt
```

Lint code:

```bash
make lint
```

This uses golangci-lint and will install it if not available.

### Dependency Management

Download and tidy dependencies:

```bash
make deps
```

### Code Generation

Generate code (if applicable):

```bash
make generate
```

## Project Structure

```
.
├── main.go                 # Entry point
├── go.mod                  # Go module definition
├── Makefile               # Build and development tasks
├── cmd/                   # Command-line commands
│   ├── cmd.go            # Root command setup
│   ├── version.go        # Version command
│   ├── version_test.go   # Version command tests
│   └── namespace/        # Namespace-related commands
│       ├── cmd.go
│       ├── create.go
│       ├── create_test.go
│       ├── list.go
│       ├── list_test.go
│       └── update.go
└── internal/             # Internal packages
    └── namespace/        # Namespace logic
        ├── config.go
        ├── config_test.go
        ├── create.go
        ├── list.go
```

## Contributing

1. Ensure tests pass: `make test`
2. Run linters: `make lint`
3. Format code: `make fmt`
4. Submit pull request

## License

This project is licensed under the Apache License, Version 2.0 - see the [LICENSE](LICENSE) file for details.

Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
