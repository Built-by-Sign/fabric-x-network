# HSM Provider Abstraction

This package provides an abstraction layer for HSM (Hardware Security Module) providers, allowing the network builder to support different HSM implementations.

## Architecture

The HSM module uses the **Provider Pattern** to abstract different HSM implementations:

```
Provider Interface
    ├── SoftHSMProvider (current implementation)
    ├── PKCS11Provider (future: real hardware HSM)
    ├── AWSCloudHSMProvider (future: AWS CloudHSM)
    └── AzureKeyVaultProvider (future: Azure Key Vault HSM)
```

## Current Implementation

### SoftHSM2 Provider

The `SoftHSMProvider` implements the `Provider` interface for SoftHSM2 (software HSM):

- **Initialize()**: Creates SoftHSM2 configuration file and token directories
- **CreateToken()**: Creates a new HSM token using `softhsm2-util`
- **TokenExists()**: Checks if a token with the given label exists
- **ListTokens()**: Lists all available tokens
- **GetLibraryPath()**: Returns the PKCS11 library path
- **GetConfigPath()**: Returns the configuration file path
- **GetEnvironment()**: Returns required environment variables
- **Validate()**: Validates that SoftHSM2 is properly installed

## Adding a New HSM Provider

To add support for a new HSM provider (e.g., real PKCS11 hardware HSM):

1. **Implement the Provider interface** in a new file (e.g., `pkcs11.go`):

```go
type PKCS11Provider struct {
    config *config.HSMConfig
    // Add provider-specific fields
}

func NewPKCS11Provider(cfg *config.HSMConfig) (*PKCS11Provider, error) {
    // Initialize provider
}

func (p *PKCS11Provider) Name() string {
    return "pkcs11"
}

func (p *PKCS11Provider) Initialize() error {
    // Setup PKCS11 environment
}

func (p *PKCS11Provider) CreateToken(label string) error {
    // Create token using PKCS11 API
}

// Implement other interface methods...
```

2. **Register the provider** in `factory.go`:

```go
case "pkcs11":
    return NewPKCS11Provider(cfg)
```

3. **Update configuration** to use the new provider:

```yaml
hsm:
  enabled: true
  provider: "pkcs11"  # Use PKCS11 instead of softhsm
  library_path: "/usr/lib/pkcs11/libpkcs11.so"
  provider_config:
    slot_id: 0
    # Provider-specific configuration
```

## Configuration

### SoftHSM2 Configuration

```yaml
hsm:
  enabled: true
  provider: "softhsm"  # or "softhsm2"
  library_path: ""     # Auto-detect if empty
  config_path: "/tmp/softhsm2.conf"
  token_dir: "/tmp/softhsm/tokens"
  pin: "1234"
  sopin: "5678"
  token_label: "FabricToken"
  multi_token: true    # One token per organization
```

### Future: PKCS11 Hardware HSM Configuration

```yaml
hsm:
  enabled: true
  provider: "pkcs11"
  library_path: "/usr/lib/pkcs11/libpkcs11.so"
  config_path: "/etc/pkcs11/pkcs11.conf"
  pin: "user-pin"
  sopin: "so-pin"
  token_label: "FabricToken"
  multi_token: true
  provider_config:
    slot_id: 0
    token_label_prefix: "FabricToken"
```

## Usage

The HSM module is automatically used during `network-builder setup`:

```bash
network-builder setup -c network.yaml -o ./out
```

If HSM is enabled in the configuration, the setup process will:

1. Create the appropriate HSM provider
2. Initialize the HSM environment
3. Create required tokens (one per organization in multi-token mode)
4. Generate crypto materials using the HSM tokens

## Testing

Run tests:

```bash
go test ./internal/hsm/...
```

## Future Enhancements

- [ ] PKCS11 hardware HSM provider
- [ ] AWS CloudHSM provider
- [ ] Azure Key Vault HSM provider
- [ ] Google Cloud HSM provider
- [ ] Token rotation support
- [ ] Multi-slot support for hardware HSMs

