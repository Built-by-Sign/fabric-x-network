/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/
package hsm

import (
	"fmt"
	"log"

	"github.com/ethsign/fabric-x-network/tools/config-builder/internal/config"
)

// Setup initializes HSM and creates required tokens
func Setup(cfg *config.NetworkConfig, opts SetupOptions) (*SetupResult, error) {
	if cfg.HSM == nil || !cfg.HSM.Enabled {
		return nil, fmt.Errorf("HSM is not enabled")
	}

	// Create provider
	provider, err := NewProvider(cfg.HSM)
	if err != nil {
		return nil, fmt.Errorf("failed to create HSM provider: %w", err)
	}

	if opts.Verbose {
		log.Printf("Using HSM provider: %s", provider.Name())
	}

	// Validate provider
	if err := provider.Validate(); err != nil {
		return nil, fmt.Errorf("HSM provider validation failed: %w", err)
	}

	// Initialize provider environment
	if err := provider.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize HSM: %w", err)
	}

	if opts.Verbose {
		log.Printf("HSM environment initialized")
		log.Printf("  Config: %s", provider.GetConfigPath())
		log.Printf("  Library: %s", provider.GetLibraryPath())
	}

	// Determine which tokens need to be created
	tokenLabels := getRequiredTokens(cfg)

	if opts.Verbose {
		log.Printf("Required tokens: %v", tokenLabels)
	}

	// Create tokens
	var createdTokens []string
	for _, label := range tokenLabels {
		if opts.DryRun {
			exists, err := provider.TokenExists(label)
			if err != nil {
				return nil, fmt.Errorf("failed to check token '%s': %w", label, err)
			}
			if opts.Verbose {
				if exists {
					log.Printf("  Token '%s' already exists", label)
				} else {
					log.Printf("  Token '%s' would be created", label)
				}
			}
			if !exists {
				createdTokens = append(createdTokens, label)
			}
		} else {
			if err := provider.CreateToken(label); err != nil {
				return nil, fmt.Errorf("failed to create token '%s': %w", label, err)
			}
			createdTokens = append(createdTokens, label)
			if opts.Verbose {
				log.Printf("  Created token: %s", label)
			}
		}
	}

	// List all tokens for verification
	allTokens, err := provider.ListTokens()
	if err != nil {
		if opts.Verbose {
			log.Printf("Warning: failed to list tokens: %v", err)
		}
	} else if opts.Verbose {
		log.Printf("Available tokens: %v", allTokens)
	}

	result := &SetupResult{
		Provider:      provider.Name(),
		ConfigPath:    provider.GetConfigPath(),
		LibraryPath:   provider.GetLibraryPath(),
		TokensCreated: createdTokens,
		Environment:   provider.GetEnvironment(),
	}

	return result, nil
}

// getRequiredTokens determines which tokens need to be created based on network configuration
func getRequiredTokens(cfg *config.NetworkConfig) []string {
	tokens := []string{}

	if cfg.HSM == nil || !cfg.HSM.Enabled {
		return tokens
	}

	// Multi-token mode: one token per orderer organization
	if cfg.HSM.MultiToken {
		for _, org := range cfg.OrdererOrgs {
			label := getTokenLabel(org.Name, org.HSMTokenLabel, cfg.HSM)
			if label != "" {
				tokens = append(tokens, label)
			}
		}
	} else {
		// Single token mode: use base token label
		if cfg.HSM.TokenLabel != "" {
			tokens = append(tokens, cfg.HSM.TokenLabel)
		}
	}

	return tokens
}

// getTokenLabel returns the token label for an organization
func getTokenLabel(orgName, customLabel string, hsmConfig *config.HSMConfig) string {
	if customLabel != "" {
		return customLabel
	}
	if hsmConfig.MultiToken {
		return fmt.Sprintf("%s-%s", hsmConfig.TokenLabel, orgName)
	}
	return hsmConfig.TokenLabel
}
