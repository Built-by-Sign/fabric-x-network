/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/
package config

// NetworkConfig represents the complete network configuration
type NetworkConfig struct {
	// Global settings
	ProjectDir string `yaml:"project_dir"`
	OutputDir  string `yaml:"output_dir"`
	ChannelID  string `yaml:"channel_id"`
	CliVersion string `yaml:"cli_version"`

	// HSM configuration
	HSM *HSMConfig `yaml:"hsm"`

	// TLS configuration
	TLS *TLSConfig `yaml:"tls,omitempty"`

	// Organizations
	OrdererOrgs []OrdererOrg `yaml:"orderer_orgs"`
	PeerOrgs    []PeerOrg    `yaml:"peer_orgs"`

	// Committer configuration
	Committer *CommitterConfig `yaml:"committer"`

	// Docker settings
	Docker DockerConfig `yaml:"docker"`
}

// TLSConfig represents TLS configuration for orderer nodes
type TLSConfig struct {
	Enabled            bool `yaml:"enabled"`              // Enable TLS for orderer nodes
	ClientAuthRequired bool `yaml:"client_auth_required"` // Require client authentication (mTLS)
}

// HSMConfig represents HSM configuration
type HSMConfig struct {
	Enabled           bool   `yaml:"enabled"`
	Provider          string `yaml:"provider"`               // HSM provider type: "softhsm", "pkcs11", "aws-cloudhsm", etc.
	LibraryPath       string `yaml:"library_path"`           // For local tools (cryptogen)
	ContainerLibPath  string `yaml:"container_library_path"` // For Docker containers
	ConfigPath        string `yaml:"config_path"`            // Provider-specific config file path
	TokenDir          string `yaml:"token_dir"`              // Token data directory (for SoftHSM)
	ContainerTokenDir string `yaml:"container_token_dir"`    // Token directory in containers
	PIN               string `yaml:"pin"`                    // User PIN
	SOPIN             string `yaml:"sopin"`                  // Security Officer PIN
	TokenLabel        string `yaml:"token_label"`            // Base token label
	MultiToken        bool   `yaml:"multi_token"`            // Multi-token mode (one per org)

	// Provider-specific configuration (for future real HSM providers)
	ProviderConfig map[string]interface{} `yaml:"provider_config,omitempty"`
}

// OrdererOrg represents an orderer organization
type OrdererOrg struct {
	Name                  string `yaml:"name"`
	Domain                string `yaml:"domain"`
	EnableOrganizationOUs bool   `yaml:"enable_organizational_units"`
	Orderers              []Node `yaml:"orderers"`
	HSMTokenLabel         string `yaml:"hsm_token_label"` // Per-org HSM token
}

// PeerOrg represents a peer organization
type PeerOrg struct {
	Name                  string `yaml:"name"`
	Domain                string `yaml:"domain"`
	EnableOrganizationOUs bool   `yaml:"enable_organizational_units"`
	Peers                 []Node `yaml:"peers"`
	Users                 []User `yaml:"users"`
}

// Node represents a network node (orderer or peer)
type Node struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"` // router, batcher, consenter, assembler (for orderer)
	Port    int    `yaml:"port"`
	ShardID int    `yaml:"shard_id,omitempty"`
	Host    string `yaml:"host"`
}

// User represents a user identity
type User struct {
	Name               string `yaml:"name"`
	MetaNamespaceAdmin bool   `yaml:"meta_namespace_admin,omitempty"`
}

// CommitterConfig represents committer component configuration
type CommitterConfig struct {
	UsePostgres bool            `yaml:"use_postgres"`
	Components  []CommitterNode `yaml:"components"`
}

// CommitterNode represents a committer component
type CommitterNode struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // db, validator, verifier, coordinator, sidecar, query-service
	Port int    `yaml:"port"`
	Host string `yaml:"host"`

	// Database specific
	PostgresUser     string `yaml:"postgres_user,omitempty"`
	PostgresPassword string `yaml:"postgres_password,omitempty"`
	PostgresDB       string `yaml:"postgres_db,omitempty"`
}

// DockerConfig represents Docker-related settings
type DockerConfig struct {
	Name            string `yaml:"name"`
	Network         string `yaml:"network"`
	NetworkDriver   string `yaml:"network_driver"`
	NetworkExternal bool   `yaml:"network_external"`

	// Image settings
	OrdererImage   string `yaml:"orderer_image"`
	CommitterImage string `yaml:"committer_image"`

	// Tools image (for cryptogen, configtxgen, etc.)
	// Defaults to docker.io/hyperledger/fabric-x-tools:0.0.4 (matching Ansible)
	ToolsImage string `yaml:"tools_image"`
}

// DefaultConfig returns a default network configuration
func DefaultConfig() *NetworkConfig {
	return &NetworkConfig{
		ProjectDir: ".",
		OutputDir:  "./out",
		ChannelID:  "arma",
		CliVersion: "latest",
		HSM: &HSMConfig{
			Enabled:           true,
			Provider:          "softhsm", // Default to SoftHSM
			LibraryPath:       "",        // Auto-detect
			ContainerLibPath:  "/usr/lib64/softhsm/libsofthsm2.so",
			ConfigPath:        "/tmp/softhsm2.conf",
			TokenDir:          "/tmp/softhsm/tokens",
			ContainerTokenDir: "/tmp/softhsm/tokens",
			PIN:               "1234",
			SOPIN:             "5678",
			TokenLabel:        "FabricToken",
			MultiToken:        true,
		},
		Docker: DockerConfig{
			Name:          "fabric_x",
			Network:       "fabric_x_net",
			NetworkDriver: "bridge",
			OrdererImage:  "hyperledger/fabric-x-orderer:local",
			ToolsImage:    "docker.io/hyperledger/fabric-x-tools:0.0.4", // Match Ansible default
		},
	}
}
