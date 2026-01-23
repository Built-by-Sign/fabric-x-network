/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/
package hsm

import (
	"fmt"

	"github.com/ethsign/fabric-x-network/tools/config-builder/internal/config"
)

// NewProvider creates a new HSM provider based on the configuration
func NewProvider(cfg *config.HSMConfig) (Provider, error) {
	if cfg == nil || !cfg.Enabled {
		return nil, fmt.Errorf("HSM is not enabled")
	}

	providerType := cfg.Provider
	if providerType == "" {
		providerType = "softhsm" // Default to SoftHSM
	}

	switch providerType {
	case "softhsm", "softhsm2":
		return NewSoftHSMProvider(cfg)
	case "pkcs11":
		// TODO: Implement PKCS11 provider for real hardware HSM
		return nil, fmt.Errorf("PKCS11 provider not yet implemented")
	case "aws-cloudhsm":
		// TODO: Implement AWS CloudHSM provider
		return nil, fmt.Errorf("AWS CloudHSM provider not yet implemented")
	case "azure-keyvault":
		// TODO: Implement Azure Key Vault HSM provider
		return nil, fmt.Errorf("Azure Key Vault HSM provider not yet implemented")
	default:
		return nil, fmt.Errorf("unknown HSM provider: %s", providerType)
	}
}
