#!/bin/bash
# Automatically detect the PKCS11 library path for the current platform
# Supports cross-platform (macOS/Linux) detection

set -e

if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS - check common Homebrew install paths
    POSSIBLE_PATHS=(
        "/opt/homebrew/opt/softhsm@2.5/lib/softhsm/libsofthsm2.so"
        "/usr/local/opt/softhsm@2.5/lib/softhsm/libsofthsm2.so"
        "/opt/homebrew/lib/softhsm/libsofthsm2.so"
        "/usr/local/lib/softhsm/libsofthsm2.so"
    )
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux - check common system paths
    POSSIBLE_PATHS=(
        "/usr/lib64/softhsm/libsofthsm2.so"
        "/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so"
        "/usr/lib/aarch64-linux-gnu/softhsm/libsofthsm2.so"
        "/usr/lib/softhsm/libsofthsm2.so"
    )
else
    echo "ERROR: Unsupported operating system: $OSTYPE" >&2
    exit 1
fi

for path in "${POSSIBLE_PATHS[@]}"; do
    if [ -f "$path" ]; then
        echo "$path"
        exit 0
    fi
done

echo "ERROR: SoftHSM2 library not found in any of the expected paths:" >&2
for path in "${POSSIBLE_PATHS[@]}"; do
    echo "  - $path" >&2
done
exit 1

