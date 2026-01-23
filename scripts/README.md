# Scripts Directory

Utility scripts for fabric-x network setup and day-one checks.

## Quickstart

- Install everything: `./scripts/install-dependencies.sh`
- Verify tooling: `./scripts/check-prerequisites.sh`
- Configure SoftHSM tokens: `./scripts/setup-hsm-tokens.sh`
- Detect PKCS11 library path: `./scripts/detect-hsm-library.sh`

## Script Details

### install-dependencies.sh

- Installs Go, Docker, Git, Python3, pip, Ansible, SoftHSM2, Python packages, and the hyperledger.fabricx Ansible collection (macOS and common Linux distros).
- Adds helpful reminders for PATH updates and Docker group membership.
- Run before `make setup` on a fresh machine.

### check-prerequisites.sh

- Confirms required CLI tools, Python modules, the hyperledger.fabricx Ansible collection, SoftHSM library presence, Go version (>=1.24.3 recommended), Docker daemon status, and key project files.
- Optional checks: `fxconfig` and `cryptogen` if available on PATH.
- Exits non-zero when mandatory tooling is missing; use after installing dependencies.

### setup-hsm-tokens.sh

- Generates a SoftHSM2 config and creates required tokens based on `ansible/inventory/fabric-x.yaml`; falls back to common orderer token labels when inventory values are templated or missing.
- Honors environment overrides: `SOFTHSM2_CONF`, `SOFTHSM_TOKEN_DIR`, `HSM_PIN`, `HSM_SOPIN`.
- Prints a summary of created tokens and current slots.

### detect-hsm-library.sh

- Auto-detects the SoftHSM2 PKCS11 library path for macOS and Linux.
- Writes the discovered path to stdout; exits non-zero if not found so it can be used in scripts, e.g. `LIB_PATH=$(./scripts/detect-hsm-library.sh)`.

## Adding New Scripts

- Make scripts executable: `chmod +x scripts/your-script.sh`
- Update this README with a short description and usage tips
- Use project-relative paths and, when needed, the pattern below:
  ```bash
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
  ```
