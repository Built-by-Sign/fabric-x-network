/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/
package setup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethsign/fabric-x-network/tools/config-builder/internal/armageddon"
	"github.com/ethsign/fabric-x-network/tools/config-builder/internal/config"
	"github.com/ethsign/fabric-x-network/tools/config-builder/internal/crypto"
	"github.com/ethsign/fabric-x-network/tools/config-builder/internal/genesis"
	"github.com/ethsign/fabric-x-network/tools/config-builder/internal/hsm"
	"github.com/ethsign/fabric-x-network/tools/config-builder/internal/template"
)

// Options contains setup command options
type Options struct {
	ConfigFile string
	OutputDir  string
	Verbose    bool
}

// Runner handles the setup process
type Runner struct {
	opts   *Options
	config *config.NetworkConfig
}

// NewRunner creates a new setup runner
func NewRunner(opts *Options) *Runner {
	return &Runner{opts: opts}
}

// Run executes the full setup process
func (r *Runner) Run() error {
	// Step 1: Load and validate configuration
	if err := r.loadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Step 2: Create output directories
	if err := r.createDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Step 3: Setup HSM (if enabled)
	if r.config.HSM != nil && r.config.HSM.Enabled {
		if err := r.setupHSM(); err != nil {
			return fmt.Errorf("failed to setup HSM: %w", err)
		}
	}

	// Step 4: Generate crypto materials
	if err := r.generateCryptoMaterials(); err != nil {
		return fmt.Errorf("failed to generate crypto materials: %w", err)
	}

	// Step 5: Generate shared_config.binpb (armageddon)
	if err := r.generateSharedConfig(); err != nil {
		return fmt.Errorf("failed to generate shared config: %w", err)
	}

	// Step 6: Generate genesis block
	if err := r.generateGenesisBlock(); err != nil {
		return fmt.Errorf("failed to generate genesis block: %w", err)
	}

	// Step 7: Generate node configurations
	if err := r.generateNodeConfigs(); err != nil {
		return fmt.Errorf("failed to generate node configurations: %w", err)
	}

	r.log("Setup completed successfully!")
	return nil
}

// loadConfig loads and validates the network configuration
func (r *Runner) loadConfig() error {
	r.log("Loading configuration from %s...", r.opts.ConfigFile)

	cfg, err := config.Load(r.opts.ConfigFile)
	if err != nil {
		return err
	}

	// Override output directory if specified via command line
	if r.opts.OutputDir != "" {
		cfg.OutputDir = r.opts.OutputDir
	}

	// Configuration validation is done during Load()

	r.config = cfg
	r.logVerbose("Configuration loaded successfully")
	r.logVerbose("  Channel ID: %s", cfg.ChannelID)
	r.logVerbose("  Output Dir: %s", cfg.OutputDir)
	r.logVerbose("  Orderer Orgs: %d", len(cfg.OrdererOrgs))
	r.logVerbose("  Peer Orgs: %d", len(cfg.PeerOrgs))
	if cfg.HSM != nil {
		r.logVerbose("  HSM Enabled: %v", cfg.HSM.Enabled)
	}

	return nil
}

// createDirectories creates the required output directories
func (r *Runner) createDirectories() error {
	r.log("Creating output directories...")

	dirs := []string{
		r.config.OutputDir,
		filepath.Join(r.config.OutputDir, "build", "config", "cryptogen-artifacts"),
		filepath.Join(r.config.OutputDir, "build", "config", "configtxgen-artifacts"),
		filepath.Join(r.config.OutputDir, "build", "config", "armageddon-artifacts"),
		filepath.Join(r.config.OutputDir, "local-deployment"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		r.logVerbose("  Created: %s", dir)
	}

	return nil
}

// setupHSM initializes HSM tokens
func (r *Runner) setupHSM() error {
	r.log("Setting up HSM tokens...")

	opts := hsm.SetupOptions{
		Verbose: r.opts.Verbose,
		DryRun:  false,
	}

	result, err := hsm.Setup(r.config, opts)
	if err != nil {
		return err
	}

	if r.opts.Verbose {
		r.logVerbose("  Provider: %s", result.Provider)
		r.logVerbose("  Config: %s", result.ConfigPath)
		r.logVerbose("  Library: %s", result.LibraryPath)
		r.logVerbose("  Tokens created: %v", result.TokensCreated)
	}

	r.log("HSM setup completed successfully")
	return nil
}

// generateCryptoMaterials generates certificates and keys using cryptogen
func (r *Runner) generateCryptoMaterials() error {
	r.log("Generating crypto materials...")

	generator := crypto.NewGenerator(r.config, r.config.OutputDir, r.opts.Verbose)
	if err := generator.Generate(); err != nil {
		return err
	}

	r.log("Crypto materials generated successfully")
	return nil
}

// generateSharedConfig generates shared_config.binpb using armageddon
func (r *Runner) generateSharedConfig() error {
	r.log("Generating shared config...")

	generator := armageddon.NewGenerator(r.config, r.config.OutputDir, r.opts.Verbose)
	if err := generator.Generate(); err != nil {
		return err
	}

	r.log("Shared config generated successfully")
	return nil
}

// generateGenesisBlock generates the genesis block using configtxgen
func (r *Runner) generateGenesisBlock() error {
	r.log("Generating genesis block...")

	generator := genesis.NewGenerator(r.config, r.config.OutputDir, r.opts.Verbose)
	if err := generator.Generate(); err != nil {
		return err
	}

	r.log("Genesis block generated successfully")
	return nil
}

// generateNodeConfigs generates configuration files for all nodes
func (r *Runner) generateNodeConfigs() error {
	r.log("Generating node configurations...")
	engine := template.NewEngine(r.config, r.config.OutputDir, r.opts.Verbose)
	if err := engine.GenerateNodeConfigs(); err != nil {
		return err
	}
	r.log("Node configurations generated successfully")
	return nil
}

// log prints a message
func (r *Runner) log(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// logVerbose prints a message only in verbose mode
func (r *Runner) logVerbose(format string, args ...interface{}) {
	if r.opts.Verbose {
		fmt.Printf(format+"\n", args...)
	}
}

// GetConfig returns the loaded configuration (for testing)
func (r *Runner) GetConfig() *config.NetworkConfig {
	return r.config
}
