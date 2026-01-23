/*
Copyright IBM Corp. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/
package crypto

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/ethsign/fabric-x-network/tools/config-builder/internal/config"
	"gopkg.in/yaml.v3"
)

// Generator handles crypto material generation
type Generator struct {
	config       *config.NetworkConfig
	outputDir    string
	verbose      bool
	cryptogen    string // path to cryptogen binary (empty if using container)
	useContainer bool   // whether to use Docker container instead of local binary
}

// NewGenerator creates a new crypto generator
func NewGenerator(cfg *config.NetworkConfig, outputDir string, verbose bool) *Generator {
	return &Generator{
		config:    cfg,
		outputDir: outputDir,
		verbose:   verbose,
	}
}

// Generate generates all crypto materials
func (g *Generator) Generate() error {
	// Step 1: Find cryptogen (prefer Docker container, fallback to local binary)
	_, err := g.findCryptogen()
	if err != nil {
		return fmt.Errorf("failed to find cryptogen: %w", err)
	}

	// Step 2: Generate crypto-config.yaml
	configPath, err := g.generateCryptoConfig()
	if err != nil {
		return fmt.Errorf("failed to generate crypto-config.yaml: %w", err)
	}

	// Step 3: Run cryptogen
	if err := g.runCryptogen(configPath); err != nil {
		return fmt.Errorf("failed to run cryptogen: %w", err)
	}

	return nil
}

// GenerateCryptoConfigOnly generates only the crypto-config.yaml file
func (g *Generator) GenerateCryptoConfigOnly() (string, error) {
	return g.generateCryptoConfig()
}

// findCryptogen locates cryptogen, preferring Docker container (matching Ansible behavior)
// Falls back to local binary if container is not available
// Returns the path to cryptogen (empty string if using Docker container) and an error
func (g *Generator) findCryptogen() (string, error) {
	// Priority 1: Try to use Docker container (matching Ansible's default behavior)
	if g.tryDockerCryptogen() {
		g.useContainer = true
		g.log("Using cryptogen from Docker container: %s", g.config.Docker.ToolsImage)
		return "", nil
	}

	// Priority 2: Use local binary (build if necessary)
	g.useContainer = false
	absOutputDir, err := filepath.Abs(g.outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute output path: %w", err)
	}

	// Target location for built cryptogen
	cliDir := filepath.Join(absOutputDir, "cli")
	targetPath := filepath.Join(cliDir, "cryptogen")

	// Check if already built
	if _, err := os.Stat(targetPath); err == nil {
		g.log("Found cryptogen at: %s", targetPath)
		g.cryptogen = targetPath
		return targetPath, nil
	}

	// Check source directory
	sourceDir := filepath.Join(g.config.ProjectDir, "tools", "cryptogen")
	if _, err := os.Stat(filepath.Join(sourceDir, "main.go")); os.IsNotExist(err) {
		return "", fmt.Errorf("cryptogen source not found at %s and Docker container unavailable", sourceDir)
	}

	// Build cryptogen
	g.log("Building cryptogen from source: %s", sourceDir)
	if err := g.buildCryptogen(sourceDir, targetPath); err != nil {
		return "", fmt.Errorf("failed to build cryptogen: %w", err)
	}

	g.cryptogen = targetPath
	return targetPath, nil
}

// tryDockerCryptogen checks if Docker and the tools image are available
func (g *Generator) tryDockerCryptogen() bool {
	// Check if docker command is available
	if _, err := exec.LookPath("docker"); err != nil {
		g.log("Docker not found in PATH, will use local binary")
		return false
	}

	// Check if the image exists or can be pulled
	// We'll try a simple docker run --help to verify docker works
	// and check if image exists
	image := g.config.Docker.ToolsImage
	if image == "" {
		image = "docker.io/hyperledger/fabric-x-tools:0.0.4" // Default from Ansible
	}

	// Try to inspect the image to see if it exists
	cmd := exec.Command("docker", "image", "inspect", image)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err == nil {
		return true // Image exists
	}

	// Image doesn't exist, but Docker is available
	// We'll try to use it anyway and let docker pull it if needed
	g.log("Docker image %s not found locally, will attempt to pull when needed", image)
	return true
}

// buildCryptogen builds the cryptogen binary from source
func (g *Generator) buildCryptogen(sourceDir, targetPath string) error {
	// Create output directory
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build command
	cmd := exec.Command("go", "build", "-o", targetPath, ".")
	cmd.Dir = sourceDir
	cmd.Env = os.Environ()

	g.log("Running: go build -o %s .", targetPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build failed: %w\nOutput: %s", err, string(output))
	}

	g.log("Successfully built cryptogen at: %s", targetPath)
	return nil
}

// generateCryptoConfig generates the crypto-config.yaml file
func (g *Generator) generateCryptoConfig() (string, error) {
	absOutputDir, err := filepath.Abs(g.outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute output path: %w", err)
	}

	configDir := filepath.Join(absOutputDir, "build", "config", "cryptogen-artifacts")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "crypto-config.yaml")

	// Build crypto-config structure
	cryptoConfig := g.buildCryptoConfig()

	// Marshal to YAML
	data, err := yaml.Marshal(cryptoConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal crypto config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write crypto config: %w", err)
	}

	g.log("Generated crypto-config.yaml at: %s", configPath)
	return configPath, nil
}

// CryptoConfig represents the structure for crypto-config.yaml
type CryptoConfig struct {
	HSM         *HSMGlobalConfig `yaml:"HSM,omitempty"`
	OrdererOrgs []OrgSpec        `yaml:"OrdererOrgs"`
	PeerOrgs    []OrgSpec        `yaml:"PeerOrgs,omitempty"`
}

// HSMGlobalConfig matches cryptogen's HSMGlobalConfig
type HSMGlobalConfig struct {
	LibraryPath string `yaml:"LibraryPath"`
	ConfigPath  string `yaml:"ConfigPath"`
	TokenLabel  string `yaml:"TokenLabel"`
	PIN         string `yaml:"PIN"`
	SOPIN       string `yaml:"SOPIN,omitempty"`
}

// HSMOrgConfig matches cryptogen's HSMOrgConfig
type HSMOrgConfig struct {
	Enabled    bool   `yaml:"Enabled"`
	CA         *bool  `yaml:"CA,omitempty"`
	Nodes      *bool  `yaml:"Nodes,omitempty"`
	Users      *bool  `yaml:"Users,omitempty"`
	TokenLabel string `yaml:"TokenLabel,omitempty"`
}

// OrgSpec matches cryptogen's OrgSpec
type OrgSpec struct {
	Name          string        `yaml:"Name"`
	Domain        string        `yaml:"Domain"`
	EnableNodeOUs bool          `yaml:"EnableNodeOUs"`
	HSM           *HSMOrgConfig `yaml:"HSM,omitempty"`
	CA            *NodeSpec     `yaml:"CA,omitempty"`
	Template      *NodeTemplate `yaml:"Template,omitempty"`
	Specs         []NodeSpec    `yaml:"Specs,omitempty"`
	Users         *UsersSpec    `yaml:"Users,omitempty"`
}

// NodeSpec matches cryptogen's NodeSpec
type NodeSpec struct {
	Hostname           string   `yaml:"Hostname,omitempty"`
	CommonName         string   `yaml:"CommonName,omitempty"`
	SANS               []string `yaml:"SANS,omitempty"`
	PublicKeyAlgorithm string   `yaml:"PublicKeyAlgorithm,omitempty"`
}

// NodeTemplate matches cryptogen's NodeTemplate
type NodeTemplate struct {
	Count              int      `yaml:"Count"`
	Start              int      `yaml:"Start,omitempty"`
	Hostname           string   `yaml:"Hostname,omitempty"`
	SANS               []string `yaml:"SANS,omitempty"`
	PublicKeyAlgorithm string   `yaml:"PublicKeyAlgorithm,omitempty"`
}

// UsersSpec matches cryptogen's UsersSpec
type UsersSpec struct {
	Count              int        `yaml:"Count"`
	PublicKeyAlgorithm string     `yaml:"PublicKeyAlgorithm,omitempty"`
	Specs              []UserSpec `yaml:"Specs,omitempty"`
}

// UserSpec matches cryptogen's UserSpec
type UserSpec struct {
	Name string `yaml:"Name"`
	HSM  *bool  `yaml:"HSM,omitempty"`
}

// buildCryptoConfig builds the crypto configuration from network config
func (g *Generator) buildCryptoConfig() *CryptoConfig {
	cc := &CryptoConfig{
		OrdererOrgs: make([]OrgSpec, 0, len(g.config.OrdererOrgs)),
		PeerOrgs:    make([]OrgSpec, 0, len(g.config.PeerOrgs)),
	}

	// Add HSM global config if enabled
	if g.config.HSM != nil && g.config.HSM.Enabled {
		libraryPath := g.getHSMLibraryPath()
		cc.HSM = &HSMGlobalConfig{
			LibraryPath: libraryPath,
			ConfigPath:  g.config.HSM.ConfigPath,
			TokenLabel:  g.config.HSM.TokenLabel,
			PIN:         g.config.HSM.PIN,
			SOPIN:       g.config.HSM.SOPIN,
		}
	}

	// Convert orderer orgs
	for _, org := range g.config.OrdererOrgs {
		orgSpec := g.convertOrdererOrg(&org)
		cc.OrdererOrgs = append(cc.OrdererOrgs, orgSpec)
	}

	// Convert peer orgs
	for _, org := range g.config.PeerOrgs {
		orgSpec := g.convertPeerOrg(&org)
		cc.PeerOrgs = append(cc.PeerOrgs, orgSpec)
	}

	return cc
}

// convertOrdererOrg converts network config orderer org to crypto config format
func (g *Generator) convertOrdererOrg(org *config.OrdererOrg) OrgSpec {
	// Ansible defaults to false if enable_organizational_units is not set
	// We match this behavior: if not explicitly set, default to false
	enableNodeOUs := org.EnableOrganizationOUs
	// Note: If the config explicitly sets it to true, we respect that
	// But to match Ansible's default behavior, we should check if it was explicitly set
	// For now, we use the value as-is, but this may need adjustment based on actual Ansible inventory

	spec := OrgSpec{
		Name:          org.Name,
		Domain:        org.Domain,
		EnableNodeOUs: enableNodeOUs,
		Specs:         make([]NodeSpec, 0, len(org.Orderers)),
	}

	// Add HSM config if enabled
	if g.config.HSM != nil && g.config.HSM.Enabled {
		nodes := true
		users := false // Users typically don't use HSM for orderers
		tokenLabel := g.getTokenLabel(org.Name, org.HSMTokenLabel)
		spec.HSM = &HSMOrgConfig{
			Enabled:    true,
			Nodes:      &nodes,
			Users:      &users,
			TokenLabel: tokenLabel,
		}
	}

	// Convert orderer nodes to specs
	for _, node := range org.Orderers {
		nodeSpec := NodeSpec{
			Hostname: node.Name,
			// Match Ansible's SANS: host.docker.internal, 0.0.0.0, localhost, 127.0.0.1, ::1
			SANS: []string{
				"host.docker.internal",
				"0.0.0.0",
				"localhost",
				"127.0.0.1",
				"::1",
			},
		}
		spec.Specs = append(spec.Specs, nodeSpec)
	}

	// Don't add Users field for orderer orgs (matching Ansible behavior)
	// Admin user is created automatically by cryptogen

	return spec
}

// convertPeerOrg converts network config peer org to crypto config format
func (g *Generator) convertPeerOrg(org *config.PeerOrg) OrgSpec {
	spec := OrgSpec{
		Name:          org.Name,
		Domain:        org.Domain,
		EnableNodeOUs: org.EnableOrganizationOUs,
		Specs:         make([]NodeSpec, 0, len(org.Peers)),
	}

	// Note: PeerOrgs typically don't use HSM in Fabric-X
	// Only add HSM config if org has a specific token label configured
	// Otherwise, skip HSM to avoid "token not found" errors

	// Convert peer nodes to specs
	// Only add Specs if there are peers defined (matching Ansible behavior)
	if len(org.Peers) > 0 {
		for _, node := range org.Peers {
			nodeSpec := NodeSpec{
				Hostname: node.Name,
				// Match Ansible's SANS: host.docker.internal, 0.0.0.0, localhost, 127.0.0.1, ::1
				SANS: []string{
					"host.docker.internal",
					"0.0.0.0",
					"localhost",
					"127.0.0.1",
					"::1",
				},
			}
			spec.Specs = append(spec.Specs, nodeSpec)
		}
	}
	// If no peers, don't set Specs field (matching Ansible behavior)

	// Convert users - with HSM: false to indicate software mode
	// Count is the number of users in addition to Admin (matching Ansible behavior)
	userCount := 0
	userSpecs := make([]UserSpec, 0)
	hsmFalse := false
	for _, user := range org.Users {
		if user.Name != "Admin" {
			// Match Ansible: add HSM: false for each user
			userSpecs = append(userSpecs, UserSpec{
				Name: user.Name,
				HSM:  &hsmFalse,
			})
			userCount++
		}
	}
	// Only add Users field if there are non-Admin users
	// Count should be the number of additional users (excluding Admin)
	if userCount > 0 {
		spec.Users = &UsersSpec{
			Count: userCount, // Count is in addition to Admin
			Specs: userSpecs,
		}
	}

	return spec
}

// getTokenLabel returns the token label for an organization
func (g *Generator) getTokenLabel(orgName, customLabel string) string {
	if customLabel != "" {
		return customLabel
	}
	if g.config.HSM != nil && g.config.HSM.MultiToken {
		return fmt.Sprintf("%s-%s", g.config.HSM.TokenLabel, orgName)
	}
	if g.config.HSM != nil {
		return g.config.HSM.TokenLabel
	}
	return "FabricToken"
}

// getHSMLibraryPath returns the HSM library path for the current OS
func (g *Generator) getHSMLibraryPath() string {
	if g.config.HSM != nil && g.config.HSM.LibraryPath != "" {
		return g.config.HSM.LibraryPath
	}

	// Auto-detect based on OS
	switch runtime.GOOS {
	case "darwin":
		// Check common macOS paths - prefer versioned homebrew path first
		paths := []string{
			"/opt/homebrew/opt/softhsm@2.5/lib/softhsm/libsofthsm2.so",
			"/opt/homebrew/opt/softhsm/lib/softhsm/libsofthsm2.so",
			"/opt/homebrew/lib/softhsm/libsofthsm2.so",
			"/usr/local/opt/softhsm/lib/softhsm/libsofthsm2.so",
			"/usr/local/lib/softhsm/libsofthsm2.so",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		return "/opt/homebrew/opt/softhsm@2.5/lib/softhsm/libsofthsm2.so"
	case "linux":
		// Check common Linux paths
		paths := []string{
			"/usr/lib64/softhsm/libsofthsm2.so",
			"/usr/lib/softhsm/libsofthsm2.so",
			"/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
		return "/usr/lib64/softhsm/libsofthsm2.so"
	default:
		return "/usr/lib/softhsm/libsofthsm2.so"
	}
}

// runCryptogen executes the cryptogen tool (via Docker container or local binary)
func (g *Generator) runCryptogen(configPath string) error {
	absOutputDir, _ := filepath.Abs(g.outputDir)
	// Cryptogen generates files directly in the output directory (peerOrganizations, ordererOrganizations)
	// But we need them in a "crypto" subdirectory to match the expected structure
	// So we point cryptogen to a temp directory, then move files to crypto/
	baseDir := filepath.Join(absOutputDir, "build", "config", "cryptogen-artifacts")
	tempOutputDir := filepath.Join(baseDir, "temp-crypto")
	cryptoDir := filepath.Join(baseDir, "crypto")

	if g.useContainer {
		return g.runCryptogenContainer(configPath, baseDir, tempOutputDir, cryptoDir)
	}

	return g.runCryptogenLocal(configPath, tempOutputDir, cryptoDir)
}

// runCryptogenContainer runs cryptogen using Docker container (matching Ansible behavior)
func (g *Generator) runCryptogenContainer(configPath, baseDir, tempOutputDir, cryptoDir string) error {
	image := g.config.Docker.ToolsImage
	if image == "" {
		image = "docker.io/hyperledger/fabric-x-tools:0.0.4" // Default from Ansible
	}

	// Docker paths (matching Ansible's cryptogen_docker_config_dir and cryptogen_docker_output_dir)
	dockerConfigDir := "/tmp/cryptogen-artifacts"
	dockerOutputDir := filepath.Join(dockerConfigDir, "crypto")

	// Ensure directories exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create base directory: %w", err)
	}
	if err := os.MkdirAll(tempOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp output directory: %w", err)
	}

	// Build docker run command (matching Ansible's container/start.yaml)
	args := []string{
		"run",
		"--rm",                                               // Auto-remove container (matches container_autoremove: true)
		"-v", fmt.Sprintf("%s:%s", baseDir, dockerConfigDir), // Mount config directory
	}

	// Set user to current user (matches container_run_as_host_user: true)
	if uid := os.Getuid(); uid != 0 {
		if gid := os.Getgid(); gid != 0 {
			args = append(args, "-u", fmt.Sprintf("%d:%d", uid, gid))
		}
	}

	// Add HSM mounts if enabled
	if g.config.HSM != nil && g.config.HSM.Enabled {
		// Mount HSM config
		if g.config.HSM.ConfigPath != "" {
			args = append(args, "-v", fmt.Sprintf("%s:%s", g.config.HSM.ConfigPath, g.config.HSM.ConfigPath))
		}
		// Mount token directory
		if g.config.HSM.TokenDir != "" {
			args = append(args, "-v", fmt.Sprintf("%s:%s", g.config.HSM.TokenDir, g.config.HSM.ContainerTokenDir))
		}
		// Set SOFTHSM2_CONF environment variable
		if g.config.HSM.ConfigPath != "" {
			args = append(args, "-e", fmt.Sprintf("SOFTHSM2_CONF=%s", g.config.HSM.ConfigPath))
		}
	}

	// Copy config file to baseDir so it's accessible in container
	configFile := filepath.Base(configPath)
	configInBaseDir := filepath.Join(baseDir, configFile)
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	if err := os.WriteFile(configInBaseDir, configData, 0644); err != nil {
		return fmt.Errorf("failed to copy config file to container directory: %w", err)
	}

	// Container command (matching Ansible's container_command)
	containerCmd := fmt.Sprintf("cryptogen generate --config=%s/%s --output=%s",
		dockerConfigDir, configFile, dockerOutputDir)
	args = append(args, image, "sh", "-c", containerCmd)

	g.log("Running cryptogen via Docker: docker %v", args)

	cmd := exec.Command("docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cryptogen container failed: %w\nOutput: %s", err, string(output))
	}

	if g.verbose {
		fmt.Println(string(output))
	}

	// Move generated files from temp directory to crypto/ subdirectory
	return g.moveCryptoFiles(tempOutputDir, cryptoDir)
}

// runCryptogenLocal runs cryptogen using local binary
func (g *Generator) runCryptogenLocal(configPath, tempOutputDir, cryptoDir string) error {
	args := []string{
		"generate",
		"--config", configPath,
		"--output", tempOutputDir,
	}

	cmd := exec.Command(g.cryptogen, args...)
	cmd.Dir = g.config.ProjectDir

	// Set environment variables for HSM
	// Only set SOFTHSM2_CONF - other parameters are read from crypto-config.yaml
	cmd.Env = os.Environ()
	if g.config.HSM != nil && g.config.HSM.Enabled {
		cmd.Env = append(cmd.Env,
			fmt.Sprintf("SOFTHSM2_CONF=%s", g.config.HSM.ConfigPath),
		)
	}

	g.log("Running cryptogen: %s %v", g.cryptogen, args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Clean up temp directory on error
		os.RemoveAll(tempOutputDir)
		return fmt.Errorf("cryptogen failed: %w\nOutput: %s", err, string(output))
	}

	if g.verbose {
		fmt.Println(string(output))
	}

	// Move generated files from temp directory to crypto/ subdirectory
	return g.moveCryptoFiles(tempOutputDir, cryptoDir)
}

// moveCryptoFiles moves generated crypto files to the final location
func (g *Generator) moveCryptoFiles(tempOutputDir, cryptoDir string) error {
	// Move generated files from temp directory to crypto/ subdirectory
	// Cryptogen generates peerOrganizations and ordererOrganizations directly in output dir
	// We need them in crypto/ subdirectory
	if err := os.MkdirAll(cryptoDir, 0755); err != nil {
		os.RemoveAll(tempOutputDir)
		return fmt.Errorf("failed to create crypto directory: %w", err)
	}

	// Move peerOrganizations if it exists
	peerOrgsSrc := filepath.Join(tempOutputDir, "peerOrganizations")
	if _, err := os.Stat(peerOrgsSrc); err == nil {
		peerOrgsDst := filepath.Join(cryptoDir, "peerOrganizations")
		if err := os.Rename(peerOrgsSrc, peerOrgsDst); err != nil {
			os.RemoveAll(tempOutputDir)
			return fmt.Errorf("failed to move peerOrganizations: %w", err)
		}
	}

	// Move ordererOrganizations if it exists
	ordererOrgsSrc := filepath.Join(tempOutputDir, "ordererOrganizations")
	if _, err := os.Stat(ordererOrgsSrc); err == nil {
		ordererOrgsDst := filepath.Join(cryptoDir, "ordererOrganizations")
		if err := os.Rename(ordererOrgsSrc, ordererOrgsDst); err != nil {
			os.RemoveAll(tempOutputDir)
			return fmt.Errorf("failed to move ordererOrganizations: %w", err)
		}
	}

	// Clean up temp directory
	if err := os.RemoveAll(tempOutputDir); err != nil {
		g.log("Warning: failed to remove temp directory: %v", err)
	}

	g.log("Crypto materials generated successfully")
	return nil
}

// log prints a message if verbose mode is enabled
func (g *Generator) log(format string, args ...interface{}) {
	if g.verbose {
		fmt.Printf("  [crypto] "+format+"\n", args...)
	}
}
