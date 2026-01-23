/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/
package hsm

import (
	"os"
	"testing"

	"github.com/ethsign/fabric-x-network/tools/config-builder/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNewProvider(t *testing.T) {
	// Skip SoftHSM provider tests when softhsm2-util is not available on the system.
	softHSMAvailable := true
	if _, err := findSoftHSMUtil(); err != nil {
		softHSMAvailable = false
	}

	tests := []struct {
		name     string
		provider string
		wantErr  bool
		wantName string
	}{
		{
			name:     "softhsm provider",
			provider: "softhsm",
			wantErr:  false,
			wantName: "softhsm",
		},
		{
			name:     "softhsm2 provider",
			provider: "softhsm2",
			wantErr:  false,
			wantName: "softhsm",
		},
		{
			name:     "empty provider defaults to softhsm",
			provider: "",
			wantErr:  false,
			wantName: "softhsm",
		},
		{
			name:     "unknown provider",
			provider: "unknown",
			wantErr:  true,
		},
		{
			name:     "pkcs11 provider not implemented",
			provider: "pkcs11",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantName == "softhsm" && !softHSMAvailable {
				t.Skip("softhsm2-util not found, skipping SoftHSM provider test")
			}

			cfg := &config.HSMConfig{
				Enabled:    true,
				Provider:   tt.provider,
				ConfigPath: "/tmp/test-softhsm2.conf",
				TokenDir:   t.TempDir(),
				PIN:        "1234",
				SOPIN:      "5678",
				TokenLabel: "TestToken",
			}

			provider, err := NewProvider(cfg)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, provider)
			} else {
				require.NoError(t, err)
				require.NotNil(t, provider)
				require.Equal(t, tt.wantName, provider.Name())
			}
		})
	}
}

func TestGetRequiredTokens(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.NetworkConfig
		expected []string
	}{
		{
			name: "multi-token mode with orderer orgs",
			config: &config.NetworkConfig{
				HSM: &config.HSMConfig{
					Enabled:    true,
					TokenLabel: "FabricToken",
					MultiToken: true,
				},
				OrdererOrgs: []config.OrdererOrg{
					{Name: "OrdererOrg1", HSMTokenLabel: "FabricToken-OrdererOrg1"},
					{Name: "OrdererOrg2", HSMTokenLabel: "FabricToken-OrdererOrg2"},
				},
			},
			expected: []string{"FabricToken-OrdererOrg1", "FabricToken-OrdererOrg2"},
		},
		{
			name: "single token mode",
			config: &config.NetworkConfig{
				HSM: &config.HSMConfig{
					Enabled:    true,
					TokenLabel: "FabricToken",
					MultiToken: false,
				},
				OrdererOrgs: []config.OrdererOrg{
					{Name: "OrdererOrg1"},
				},
			},
			expected: []string{"FabricToken"},
		},
		{
			name: "no HSM enabled",
			config: &config.NetworkConfig{
				HSM: &config.HSMConfig{
					Enabled: false,
				},
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := getRequiredTokens(tt.config)
			require.Equal(t, tt.expected, tokens)
		})
	}
}

func TestGetTokenLabel(t *testing.T) {
	hsmConfig := &config.HSMConfig{
		TokenLabel: "FabricToken",
		MultiToken: true,
	}

	tests := []struct {
		name        string
		orgName     string
		customLabel string
		expected    string
	}{
		{
			name:        "custom label",
			orgName:     "OrdererOrg1",
			customLabel: "CustomToken",
			expected:    "CustomToken",
		},
		{
			name:        "multi-token mode",
			orgName:     "OrdererOrg1",
			customLabel: "",
			expected:    "FabricToken-OrdererOrg1",
		},
		{
			name:        "single token mode",
			orgName:     "OrdererOrg1",
			customLabel: "",
			expected:    "FabricToken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := *hsmConfig
			if tt.name == "single token mode" {
				cfg.MultiToken = false
			}
			label := getTokenLabel(tt.orgName, tt.customLabel, &cfg)
			require.Equal(t, tt.expected, label)
		})
	}
}

// TestSoftHSMProvider_Integration requires SoftHSM2 to be installed
// This is an integration test that should be run manually
func TestSoftHSMProvider_Integration(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test")
	}

	// Check if softhsm2-util is available
	if _, err := findSoftHSMUtil(); err != nil {
		t.Skip("softhsm2-util not found, skipping integration test")
	}

	tempDir := t.TempDir()
	cfg := &config.HSMConfig{
		Enabled:    true,
		Provider:   "softhsm",
		ConfigPath: tempDir + "/softhsm2.conf",
		TokenDir:   tempDir + "/tokens",
		PIN:        "1234",
		SOPIN:      "5678",
		TokenLabel: "TestToken",
	}

	provider, err := NewSoftHSMProvider(cfg)
	require.NoError(t, err)
	require.NotNil(t, provider)

	// Initialize
	err = provider.Initialize()
	require.NoError(t, err)

	// Validate
	err = provider.Validate()
	require.NoError(t, err)

	// Create token
	err = provider.CreateToken("TestToken")
	require.NoError(t, err)

	// Check if token exists
	exists, err := provider.TokenExists("TestToken")
	require.NoError(t, err)
	require.True(t, exists)

	// List tokens
	tokens, err := provider.ListTokens()
	require.NoError(t, err)
	require.Contains(t, tokens, "TestToken")
}
