/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary test config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-network.yaml")

	configContent := `
channel_id: test-channel
output_dir: ./test-out

hsm:
  enabled: true
  config_path: /tmp/softhsm2.conf
  token_dir: /tmp/softhsm/tokens
  pin: "1234"
  sopin: "5678"
  multi_token: true

orderer_orgs:
  - name: TestOrdererOrg1
    domain: testordererorg1.example.com
    orderers:
      - name: orderer1
        type: consenter
        port: 7050
        host: localhost

peer_orgs:
  - name: TestOrg1
    domain: testorg1.example.com
    users:
      - name: admin
        meta_namespace_admin: true

docker:
  network: test_net
  orderer_image: hyperledger/fabric-x-orderer:test
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Test loading
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded values
	if cfg.ChannelID != "test-channel" {
		t.Errorf("Expected channel_id 'test-channel', got '%s'", cfg.ChannelID)
	}

	// OutputDir is resolved relative to the detected project directory; ensure it points to test-out
	if filepath.Base(cfg.OutputDir) != "test-out" {
		t.Errorf("Expected output_dir to end with 'test-out', got '%s'", cfg.OutputDir)
	}
	if !filepath.IsAbs(cfg.OutputDir) {
		t.Errorf("Expected output_dir to be absolute, got '%s'", cfg.OutputDir)
	}

	if cfg.HSM == nil {
		t.Fatal("Expected HSM config to be set")
	}

	if !cfg.HSM.Enabled {
		t.Error("Expected HSM to be enabled")
	}

	if cfg.HSM.PIN != "1234" {
		t.Errorf("Expected PIN '1234', got '%s'", cfg.HSM.PIN)
	}

	if len(cfg.OrdererOrgs) != 1 {
		t.Errorf("Expected 1 orderer org, got %d", len(cfg.OrdererOrgs))
	}

	if cfg.OrdererOrgs[0].Name != "TestOrdererOrg1" {
		t.Errorf("Expected orderer org name 'TestOrdererOrg1', got '%s'", cfg.OrdererOrgs[0].Name)
	}

	if len(cfg.OrdererOrgs[0].Orderers) != 1 {
		t.Errorf("Expected 1 orderer, got %d", len(cfg.OrdererOrgs[0].Orderers))
	}

	if len(cfg.PeerOrgs) != 1 {
		t.Errorf("Expected 1 peer org, got %d", len(cfg.PeerOrgs))
	}

	if cfg.Docker.Network != "test_net" {
		t.Errorf("Expected docker network 'test_net', got '%s'", cfg.Docker.Network)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *NetworkConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &NetworkConfig{
				ChannelID: "test",
				OrdererOrgs: []OrdererOrg{
					{
						Name:   "Org1",
						Domain: "org1.example.com",
						Orderers: []Node{
							{Name: "orderer1", Type: "consenter", Port: 7050},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing channel_id",
			config: &NetworkConfig{
				OrdererOrgs: []OrdererOrg{
					{Name: "Org1", Domain: "org1.example.com", Orderers: []Node{{Name: "o1"}}},
				},
			},
			wantErr: true,
			errMsg:  "channel_id is required",
		},
		{
			name: "no orderer orgs",
			config: &NetworkConfig{
				ChannelID: "test",
			},
			wantErr: true,
			errMsg:  "at least one orderer organization is required",
		},
		{
			name: "orderer org missing name",
			config: &NetworkConfig{
				ChannelID: "test",
				OrdererOrgs: []OrdererOrg{
					{Domain: "org1.example.com", Orderers: []Node{{Name: "o1"}}},
				},
			},
			wantErr: true,
			errMsg:  "orderer_orgs[0].name is required",
		},
		{
			name: "hsm enabled but missing config_path",
			config: &NetworkConfig{
				ChannelID: "test",
				OrdererOrgs: []OrdererOrg{
					{Name: "Org1", Domain: "org1.example.com", Orderers: []Node{{Name: "o1"}}},
				},
				HSM: &HSMConfig{
					Enabled:  true,
					TokenDir: "/tmp/tokens",
				},
			},
			wantErr: true,
			errMsg:  "hsm.config_path is required when HSM is enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errMsg)
				} else if err.Error() != tt.errMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got '%v'", err)
				}
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ChannelID != "arma" {
		t.Errorf("Expected default channel_id 'arma', got '%s'", cfg.ChannelID)
	}

	if cfg.HSM == nil {
		t.Fatal("Expected default HSM config to be set")
	}

	if !cfg.HSM.Enabled {
		t.Error("Expected HSM to be enabled by default")
	}

	if cfg.HSM.PIN != "1234" {
		t.Errorf("Expected default PIN '1234', got '%s'", cfg.HSM.PIN)
	}

	if cfg.Docker.Network != "fabric_x_net" {
		t.Errorf("Expected default docker network 'fabric_x_net', got '%s'", cfg.Docker.Network)
	}
}

func TestGetOrdererOrg(t *testing.T) {
	cfg := &NetworkConfig{
		OrdererOrgs: []OrdererOrg{
			{Name: "Org1", Domain: "org1.example.com"},
			{Name: "Org2", Domain: "org2.example.com"},
		},
	}

	org := cfg.GetOrdererOrg("Org1")
	if org == nil {
		t.Fatal("Expected to find Org1")
	}
	if org.Domain != "org1.example.com" {
		t.Errorf("Expected domain 'org1.example.com', got '%s'", org.Domain)
	}

	org = cfg.GetOrdererOrg("NonExistent")
	if org != nil {
		t.Error("Expected nil for non-existent org")
	}
}

func TestAllOrderers(t *testing.T) {
	cfg := &NetworkConfig{
		OrdererOrgs: []OrdererOrg{
			{
				Name: "Org1",
				Orderers: []Node{
					{Name: "orderer1"},
					{Name: "orderer2"},
				},
			},
			{
				Name: "Org2",
				Orderers: []Node{
					{Name: "orderer3"},
				},
			},
		},
	}

	orderers := cfg.AllOrderers()
	if len(orderers) != 3 {
		t.Errorf("Expected 3 orderers, got %d", len(orderers))
	}
}

