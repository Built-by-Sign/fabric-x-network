/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/
package hsm

// Provider defines the interface for HSM providers
// This abstraction allows supporting different HSM implementations:
// - SoftHSM2 (software HSM for development/testing)
// - PKCS11 (standard hardware HSM)
// - AWS CloudHSM
// - Azure Key Vault HSM
// - etc.
type Provider interface {
	// Name returns the provider name (e.g., "softhsm", "pkcs11")
	Name() string

	// Initialize sets up the HSM environment (config files, directories, etc.)
	Initialize() error

	// CreateToken creates a single HSM token with the given label
	CreateToken(label string) error

	// TokenExists checks if a token with the given label exists
	TokenExists(label string) (bool, error)

	// ListTokens returns a list of all available token labels
	ListTokens() ([]string, error)

	// GetLibraryPath returns the PKCS11 library path for this provider
	GetLibraryPath() string

	// GetConfigPath returns the configuration file path
	GetConfigPath() string

	// GetEnvironment returns environment variables needed for this provider
	GetEnvironment() map[string]string

	// Validate verifies that the provider is properly configured and available
	Validate() error
}

// TokenInfo represents information about an HSM token
type TokenInfo struct {
	Label string
	Slot  int
	PIN   string
}

// SetupResult contains the result of HSM setup
type SetupResult struct {
	Provider      string
	ConfigPath    string
	LibraryPath   string
	TokensCreated []string
	Environment   map[string]string
}

// SetupOptions contains options for HSM setup
type SetupOptions struct {
	Verbose bool
	DryRun  bool // If true, only validate but don't create tokens
}
