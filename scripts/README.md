# Scripts Directory

This directory contains utility scripts for the fabric-x network setup and management.

## Available Scripts

### `setup-hsm.sh`

Sets up SoftHSM for use with cryptogen. This script:

- Reads HSM configuration from `ansible/inventory/fabric-x.yaml`
- Creates SoftHSM configuration file and token directory
- Initializes or verifies the HSM token
- Can be run manually or automatically via `make setup`

**Usage:**

```bash
# Run manually
./scripts/setup-hsm.sh

# Or via make (automatically called during setup)
make setup
```

**Configuration:**
The script reads configuration from:

1. `ansible/inventory/fabric-x.yaml` (hsm_global_config section)
2. Environment variables (HSM_TOKEN_LABEL, HSM_PIN, HSM_SOPIN, etc.)
3. Default values (fallback)

## Adding New Scripts

When adding new scripts to this directory:

1. Make scripts executable: `chmod +x scripts/your-script.sh`
2. Update this README with a description
3. Ensure scripts use relative paths from the project root
4. Use the `PROJECT_ROOT` pattern if needed:
   ```bash
   SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
   PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
   ```
