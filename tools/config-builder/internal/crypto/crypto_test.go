/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/
package crypto

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethsign/fabric-x-network/tools/config-builder/internal/config"
	"gopkg.in/yaml.v3"
)

func TestGenerateCryptoConfig(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.NetworkConfig{
		ProjectDir: tempDir,
		OutputDir:  tempDir,
		ChannelID:  "test-channel",
		HSM: &config.HSMConfig{
			Enabled:    true,
			ConfigPath: "/tmp/softhsm2.conf",
			TokenDir:   "/tmp/softhsm/tokens",
			PIN:        "1234",
			SOPIN:      "5678",
			TokenLabel: "TestToken",
			MultiToken: true,
		},
		OrdererOrgs: []config.OrdererOrg{
			{
				Name:   "OrdererOrg1",
				Domain: "ordererorg1.example.com",
				Orderers: []config.Node{
					{Name: "orderer1", Type: "consenter", Port: 7050},
					{Name: "orderer2", Type: "batcher", Port: 7051},
				},
			},
		},
		PeerOrgs: []config.PeerOrg{
			{
				Name:   "Org1MSP",
				Domain: "org1.example.com",
				Peers: []config.Node{
					{Name: "peer0"},
				},
				Users: []config.User{
					{Name: "Admin"},
					{Name: "User1"},
				},
			},
		},
	}

	generator := NewGenerator(cfg, tempDir, false)
	configPath, err := generator.GenerateCryptoConfigOnly()
	if err != nil {
		t.Fatalf("Failed to generate crypto config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Generated config file does not exist: %s", configPath)
	}

	// Read and verify content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read generated config: %v", err)
	}

	var cc CryptoConfig
	if err := yaml.Unmarshal(data, &cc); err != nil {
		t.Fatalf("Failed to unmarshal generated config: %v", err)
	}

	// Verify HSM config
	if cc.HSM == nil {
		t.Fatal("Expected HSM config to be set")
	}
	if cc.HSM.PIN != "1234" {
		t.Errorf("Expected PIN '1234', got '%s'", cc.HSM.PIN)
	}

	// Verify orderer orgs
	if len(cc.OrdererOrgs) != 1 {
		t.Fatalf("Expected 1 orderer org, got %d", len(cc.OrdererOrgs))
	}
	if cc.OrdererOrgs[0].Name != "OrdererOrg1" {
		t.Errorf("Expected orderer org name 'OrdererOrg1', got '%s'", cc.OrdererOrgs[0].Name)
	}
	if len(cc.OrdererOrgs[0].Specs) != 2 {
		t.Errorf("Expected 2 orderer specs, got %d", len(cc.OrdererOrgs[0].Specs))
	}

	// Verify HSM org config with multi-token
	if cc.OrdererOrgs[0].HSM == nil {
		t.Fatal("Expected HSM org config to be set")
	}
	if !cc.OrdererOrgs[0].HSM.Enabled {
		t.Error("Expected HSM to be enabled for orderer org")
	}
	expectedTokenLabel := "TestToken-OrdererOrg1"
	if cc.OrdererOrgs[0].HSM.TokenLabel != expectedTokenLabel {
		t.Errorf("Expected token label '%s', got '%s'", expectedTokenLabel, cc.OrdererOrgs[0].HSM.TokenLabel)
	}

	// Verify peer orgs
	if len(cc.PeerOrgs) != 1 {
		t.Fatalf("Expected 1 peer org, got %d", len(cc.PeerOrgs))
	}
	if len(cc.PeerOrgs[0].Specs) != 1 {
		t.Errorf("Expected 1 peer spec, got %d", len(cc.PeerOrgs[0].Specs))
	}
}

func TestGetTokenLabel(t *testing.T) {
	tests := []struct {
		name        string
		orgName     string
		customLabel string
		hsm         *config.HSMConfig
		expected    string
	}{
		{
			name:        "custom label",
			orgName:     "Org1",
			customLabel: "CustomToken",
			hsm:         &config.HSMConfig{TokenLabel: "DefaultToken"},
			expected:    "CustomToken",
		},
		{
			name:        "multi-token mode",
			orgName:     "Org1",
			customLabel: "",
			hsm:         &config.HSMConfig{TokenLabel: "FabricToken", MultiToken: true},
			expected:    "FabricToken-Org1",
		},
		{
			name:        "single token mode",
			orgName:     "Org1",
			customLabel: "",
			hsm:         &config.HSMConfig{TokenLabel: "FabricToken", MultiToken: false},
			expected:    "FabricToken",
		},
		{
			name:        "no hsm config",
			orgName:     "Org1",
			customLabel: "",
			hsm:         nil,
			expected:    "FabricToken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.NetworkConfig{HSM: tt.hsm}
			generator := NewGenerator(cfg, "", false)
			result := generator.getTokenLabel(tt.orgName, tt.customLabel)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetHSMLibraryPath(t *testing.T) {
	// Test with explicit path
	cfg := &config.NetworkConfig{
		HSM: &config.HSMConfig{
			LibraryPath: "/custom/path/libsofthsm2.so",
		},
	}
	generator := NewGenerator(cfg, "", false)
	path := generator.getHSMLibraryPath()
	if path != "/custom/path/libsofthsm2.so" {
		t.Errorf("Expected '/custom/path/libsofthsm2.so', got '%s'", path)
	}

	// Test auto-detection (just verify it returns something)
	cfg2 := &config.NetworkConfig{
		HSM: &config.HSMConfig{},
	}
	generator2 := NewGenerator(cfg2, "", false)
	path2 := generator2.getHSMLibraryPath()
	if path2 == "" {
		t.Error("Expected non-empty library path from auto-detection")
	}
	if !strings.HasSuffix(path2, "libsofthsm2.so") {
		t.Errorf("Expected path ending with 'libsofthsm2.so', got '%s'", path2)
	}
}

func TestBuildCryptoConfig(t *testing.T) {
	cfg := &config.NetworkConfig{
		HSM: &config.HSMConfig{
			Enabled:    true,
			ConfigPath: "/tmp/softhsm2.conf",
			PIN:        "1234",
			TokenLabel: "TestToken",
			MultiToken: true,
		},
		OrdererOrgs: []config.OrdererOrg{
			{
				Name:                  "OrdererOrg1",
				Domain:                "orderer1.example.com",
				EnableOrganizationOUs: true,
				Orderers: []config.Node{
					{Name: "orderer1"},
				},
			},
		},
		PeerOrgs: []config.PeerOrg{
			{
				Name:   "Org1",
				Domain: "org1.example.com",
				Peers:  []config.Node{{Name: "peer0"}},
				Users:  []config.User{{Name: "Admin"}, {Name: "User1"}},
			},
		},
	}

	generator := NewGenerator(cfg, "", false)
	cc := generator.buildCryptoConfig()

	// Verify structure
	if cc.HSM == nil {
		t.Fatal("Expected HSM config")
	}
	if len(cc.OrdererOrgs) != 1 {
		t.Errorf("Expected 1 orderer org, got %d", len(cc.OrdererOrgs))
	}
	if len(cc.PeerOrgs) != 1 {
		t.Errorf("Expected 1 peer org, got %d", len(cc.PeerOrgs))
	}

	// Verify orderer org
	ordererOrg := cc.OrdererOrgs[0]
	if !ordererOrg.EnableNodeOUs {
		t.Error("Expected EnableNodeOUs to be true")
	}
	if ordererOrg.HSM == nil || !ordererOrg.HSM.Enabled {
		t.Error("Expected HSM to be enabled for orderer org")
	}

	// Verify peer org users
	peerOrg := cc.PeerOrgs[0]
	if peerOrg.Users == nil {
		t.Fatal("Expected Users to be set")
	}
	// User1 should be in Specs (Admin is added automatically, not in count)
	if peerOrg.Users.Count != 1 {
		t.Errorf("Expected Users.Count to be 1, got %d", peerOrg.Users.Count)
	}
}

func TestFindCryptogen(t *testing.T) {
	tempDir := t.TempDir()

	// Create a fake cryptogen binary
	cliDir := filepath.Join(tempDir, "cli")
	if err := os.MkdirAll(cliDir, 0755); err != nil {
		t.Fatalf("Failed to create cli dir: %v", err)
	}
	cryptogenPath := filepath.Join(cliDir, "cryptogen")
	if err := os.WriteFile(cryptogenPath, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("Failed to create fake cryptogen: %v", err)
	}

	cfg := &config.NetworkConfig{ProjectDir: tempDir}
	generator := NewGenerator(cfg, tempDir, false)

	path, err := generator.findCryptogen()
	if err != nil {
		t.Fatalf("Failed to find cryptogen: %v", err)
	}

	// Check if Docker is available - if so, path will be empty (using container)
	_, dockerAvailable := exec.LookPath("docker")
	if dockerAvailable == nil {
		// Docker is available, so path should be empty (using container)
		if path != "" {
			t.Logf("Docker is available, but path is not empty: '%s'", path)
		}
	} else {
		// Docker is not available, so should use local binary
		if path != cryptogenPath {
			t.Errorf("Expected local binary path '%s', got '%s'", cryptogenPath, path)
		}
	}
}
