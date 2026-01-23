/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// Load reads and parses a network configuration file
func Load(path string) (*NetworkConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Auto-detect project_dir if not set or is empty/relative
	// If project_dir is empty or ".", auto-detect from config file location
	// If project_dir is a relative path, resolve it first, then validate
	if config.ProjectDir == "" || config.ProjectDir == "." {
		// Auto-detect: find fabric-x-network root directory
		detectedDir, err := detectProjectDir(path)
		if err != nil {
			return nil, fmt.Errorf("failed to detect project directory: %w", err)
		}
		config.ProjectDir = detectedDir
	} else if !filepath.IsAbs(config.ProjectDir) {
		// Resolve relative path based on config file directory
		configDir := filepath.Dir(path)
		absConfigDir, err := filepath.Abs(configDir)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for config directory: %w", err)
		}
		resolvedDir, err := filepath.Abs(filepath.Join(absConfigDir, config.ProjectDir))
		if err != nil {
			return nil, fmt.Errorf("failed to resolve project directory: %w", err)
		}
		config.ProjectDir = resolvedDir
	}

	// Post-process: auto-detect HSM library path if not set
	if config.HSM != nil && config.HSM.Enabled && config.HSM.LibraryPath == "" {
		config.HSM.LibraryPath = detectHSMLibraryPath()
	}

	// Resolve relative paths
	if !filepath.IsAbs(config.OutputDir) {
		config.OutputDir = filepath.Join(config.ProjectDir, config.OutputDir)
	}

	return config, nil
}

// Save writes a network configuration to a file
func Save(config *NetworkConfig, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// detectProjectDir automatically detects the project directory by looking for
// the fabric-x-network root directory (contains tools/ directory)
func detectProjectDir(configPath string) (string, error) {
	configDir, err := filepath.Abs(filepath.Dir(configPath))
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	// Start from config file directory and walk up to find fabric-x-network root
	// The root directory should contain a "tools" directory
	currentDir := configDir
	for {
		// Check if this directory contains "tools" directory (indicating fabric-x-network root)
		toolsDir := filepath.Join(currentDir, "tools")
		if info, err := os.Stat(toolsDir); err == nil && info.IsDir() {
			return currentDir, nil
		}

		// Check if we've reached the filesystem root
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached filesystem root, use config directory's parent as fallback
			// This handles cases where config is in config-builder/configs/
			if filepath.Base(configDir) == "configs" {
				return filepath.Dir(filepath.Dir(configDir)), nil
			}
			return filepath.Dir(configDir), nil
		}

		currentDir = parentDir
	}
}

// detectHSMLibraryPath auto-detects the SoftHSM library path based on OS
func detectHSMLibraryPath() string {
	switch runtime.GOOS {
	case "darwin":
		// macOS - check common Homebrew paths
		paths := []string{
			"/opt/homebrew/opt/softhsm@2.5/lib/softhsm/libsofthsm2.so",
			"/opt/homebrew/lib/softhsm/libsofthsm2.so",
			"/usr/local/opt/softhsm/lib/softhsm/libsofthsm2.so",
			"/usr/local/lib/softhsm/libsofthsm2.so",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	case "linux":
		// Linux - check common paths based on architecture
		arch := runtime.GOARCH
		var paths []string
		if arch == "amd64" {
			paths = []string{
				"/usr/lib64/softhsm/libsofthsm2.so",
				"/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so",
				"/usr/lib/softhsm/libsofthsm2.so",
			}
		} else {
			paths = []string{
				"/usr/lib/softhsm/libsofthsm2.so",
				"/usr/lib/aarch64-linux-gnu/softhsm/libsofthsm2.so",
			}
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}

	// Fallback
	return "/usr/lib/softhsm/libsofthsm2.so"
}

// Validate checks if the configuration is valid
func (c *NetworkConfig) Validate() error {
	if c.ChannelID == "" {
		return fmt.Errorf("channel_id is required")
	}

	if len(c.OrdererOrgs) == 0 {
		return fmt.Errorf("at least one orderer organization is required")
	}

	for i, org := range c.OrdererOrgs {
		if org.Name == "" {
			return fmt.Errorf("orderer_orgs[%d].name is required", i)
		}
		if org.Domain == "" {
			return fmt.Errorf("orderer_orgs[%d].domain is required", i)
		}
		if len(org.Orderers) == 0 {
			return fmt.Errorf("orderer_orgs[%d] must have at least one orderer", i)
		}
	}

	if c.HSM != nil && c.HSM.Enabled {
		if c.HSM.ConfigPath == "" {
			return fmt.Errorf("hsm.config_path is required when HSM is enabled")
		}
		if c.HSM.TokenDir == "" {
			return fmt.Errorf("hsm.token_dir is required when HSM is enabled")
		}
	}

	return nil
}

// GetOrdererOrg returns the orderer organization by name
func (c *NetworkConfig) GetOrdererOrg(name string) *OrdererOrg {
	for i := range c.OrdererOrgs {
		if c.OrdererOrgs[i].Name == name {
			return &c.OrdererOrgs[i]
		}
	}
	return nil
}

// GetPeerOrg returns the peer organization by name
func (c *NetworkConfig) GetPeerOrg(name string) *PeerOrg {
	for i := range c.PeerOrgs {
		if c.PeerOrgs[i].Name == name {
			return &c.PeerOrgs[i]
		}
	}
	return nil
}

// AllOrderers returns all orderer nodes across all organizations
func (c *NetworkConfig) AllOrderers() []Node {
	var nodes []Node
	for _, org := range c.OrdererOrgs {
		nodes = append(nodes, org.Orderers...)
	}
	return nodes
}

// AllPeers returns all peer nodes across all organizations
func (c *NetworkConfig) AllPeers() []Node {
	var nodes []Node
	for _, org := range c.PeerOrgs {
		nodes = append(nodes, org.Peers...)
	}
	return nodes
}
