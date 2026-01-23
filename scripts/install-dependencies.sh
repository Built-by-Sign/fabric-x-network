#!/bin/bash
#
# Dependency installation script
# Auto-detect OS and install all dependencies required by make setup
# Supports macOS and Linux (Ubuntu/Debian/CentOS/RHEL)
#

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Output helpers
print_info() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}❌${NC} $1"
}

print_section() {
    echo -e "${BLUE}==>${NC} $1"
}

# Detect operating system
detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
        print_info "Detected OS: macOS"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if [ -f /etc/os-release ]; then
            . /etc/os-release
            OS="linux"
            OS_ID="$ID"
            print_info "Detected OS: Linux ($OS_ID)"
        else
            print_error "Unable to detect Linux distribution"
            exit 1
        fi
    else
        print_error "Unsupported OS: $OSTYPE"
        exit 1
    fi
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check and install Homebrew (macOS)
check_homebrew() {
    if [ "$OS" != "macos" ]; then
        return 0
    fi
    
    if command_exists brew; then
        print_info "Homebrew already installed"
        return 0
    else
        print_warn "Homebrew not found, installing..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        if command_exists brew; then
            print_info "Homebrew installation succeeded"
        else
            print_error "Homebrew installation failed, please install manually"
            exit 1
        fi
    fi
}

# Install Go
install_go() {
    if command_exists go; then
        local version=$(go version | awk '{print $3}')
        print_info "Go already installed ($version)"
        return 0
    fi
    
    print_warn "Go not found, installing..."
    
    if [ "$OS" == "macos" ]; then
        brew install go
    elif [ "$OS_ID" == "ubuntu" ] || [ "$OS_ID" == "debian" ]; then
        sudo apt-get update
        sudo apt-get install -y golang-go
        
        # Check whether Go is in standard paths
        if [ -f /usr/bin/go ]; then
            print_info "Go installed at /usr/bin/go"
        elif [ -f /usr/local/go/bin/go ]; then
            print_warn "Go installed at /usr/local/go/bin/go; add it to PATH"
            print_warn "Run: export PATH=\$PATH:/usr/local/go/bin"
            print_warn "Or add to ~/.bashrc or ~/.profile:"
            print_warn "  echo 'export PATH=\$PATH:/usr/local/go/bin' >> ~/.bashrc"
        fi
    elif [ "$OS_ID" == "centos" ] || [ "$OS_ID" == "rhel" ] || [ "$OS_ID" == "fedora" ]; then
        if command_exists dnf; then
            sudo dnf install -y golang
        else
            sudo yum install -y golang
        fi
        
        # Check whether Go is in standard paths
        if [ -f /usr/bin/go ]; then
            print_info "Go installed at /usr/bin/go"
        elif [ -f /usr/local/go/bin/go ]; then
            print_warn "Go installed at /usr/local/go/bin/go; add it to PATH"
            print_warn "Run: export PATH=\$PATH:/usr/local/go/bin"
            print_warn "Or add to ~/.bashrc or ~/.profile:"
            print_warn "  echo 'export PATH=\$PATH:/usr/local/go/bin' >> ~/.bashrc"
        fi
    fi
    
    # Wait briefly to allow PATH updates
    sleep 1
    
    # Re-check Go availability
    if command_exists go; then
        local version=$(go version | awk '{print $3}')
        print_info "Go installation succeeded ($version)"
        return 0
    else
        # Try to locate the Go installation
        local go_paths=(
            "/usr/bin/go"
            "/usr/local/go/bin/go"
            "/usr/lib/go/bin/go"
            "$HOME/go/bin/go"
        )
        
        local found_go=""
        for path in "${go_paths[@]}"; do
            if [ -f "$path" ]; then
                found_go="$path"
                break
            fi
        done
        
        if [ -n "$found_go" ]; then
            print_error "Go is installed but not on PATH"
            print_error "Found Go at: $found_go"
            print_warn "Add Go to PATH:"
            if [[ "$found_go" == *"/usr/local/go/bin/go"* ]]; then
                print_warn "  export PATH=\$PATH:/usr/local/go/bin"
                print_warn "  Or add to ~/.bashrc: echo 'export PATH=\$PATH:/usr/local/go/bin' >> ~/.bashrc"
            else
                local go_dir=$(dirname "$found_go")
                print_warn "  export PATH=\$PATH:$go_dir"
                print_warn "  Or add to ~/.bashrc: echo 'export PATH=\$PATH:$go_dir' >> ~/.bashrc"
            fi
            print_warn ""
            print_warn "Then rerun this script or re-login"
            exit 1
        else
            print_error "Go installation failed; go executable not found"
            print_warn ""
            print_warn "Suggest installing Go manually:"
            print_warn "  1. Visit https://go.dev/dl/ and download the latest version"
            print_warn "  2. Or reinstall via package manager:"
            if [ "$OS_ID" == "ubuntu" ] || [ "$OS_ID" == "debian" ]; then
                print_warn "     sudo apt-get install -y golang-go"
            else
                print_warn "     sudo dnf install -y golang  # or sudo yum install -y golang"
            fi
            exit 1
        fi
    fi
}

# Install Docker
install_docker() {
    if command_exists docker; then
        local version=$(docker --version | awk '{print $3}' | sed 's/,//')
        print_info "Docker already installed ($version)"
        
        # Check whether Docker daemon is running
        if docker info >/dev/null 2>&1; then
            print_info "Docker daemon is running"
        else
            print_warn "Docker daemon is not running; start Docker Desktop or the Docker service"
        fi
        return 0
    fi
    
    print_warn "Docker not found, installing..."
    
    if [ "$OS" == "macos" ]; then
        if command_exists brew; then
            print_info "Installing Docker Desktop via Homebrew..."
            brew install --cask docker
            print_warn "Docker Desktop installed; please launch the Docker Desktop app"
        else
            print_warn "On macOS, install Docker Desktop manually: https://www.docker.com/products/docker-desktop"
            print_warn "Or use: brew install --cask docker"
        fi
    elif [ "$OS_ID" == "ubuntu" ] || [ "$OS_ID" == "debian" ]; then
        sudo apt-get update
        sudo apt-get install -y docker.io
        sudo systemctl enable docker
        sudo systemctl start docker
        sudo usermod -aG docker $USER
        print_warn "Added current user to docker group; re-login or run: newgrp docker"
    elif [ "$OS_ID" == "centos" ] || [ "$OS_ID" == "rhel" ] || [ "$OS_ID" == "fedora" ]; then
        if command_exists dnf; then
            sudo dnf install -y docker
        else
            sudo yum install -y docker
        fi
        sudo systemctl enable docker
        sudo systemctl start docker
        sudo usermod -aG docker $USER
        print_warn "Added current user to docker group; re-login or run: newgrp docker"
    fi
    
    if command_exists docker; then
        print_info "Docker installation succeeded"
    else
        print_error "Docker installation failed; please install manually"
    fi
}

# Install Git
install_git() {
    if command_exists git; then
        local version=$(git --version | awk '{print $3}')
        print_info "Git already installed ($version)"
        return 0
    fi
    
    print_warn "Git not found, installing..."
    
    if [ "$OS" == "macos" ]; then
        brew install git
    elif [ "$OS_ID" == "ubuntu" ] || [ "$OS_ID" == "debian" ]; then
        sudo apt-get update
        sudo apt-get install -y git
    elif [ "$OS_ID" == "centos" ] || [ "$OS_ID" == "rhel" ] || [ "$OS_ID" == "fedora" ]; then
        if command_exists dnf; then
            sudo dnf install -y git
        else
            sudo yum install -y git
        fi
    fi
    
    if command_exists git; then
        print_info "Git installation succeeded"
    else
        print_error "Git installation failed"
        exit 1
    fi
}

# Install Python3
install_python3() {
    if command_exists python3; then
        local version=$(python3 --version | awk '{print $2}')
        print_info "Python3 already installed ($version)"
        return 0
    fi
    
    print_warn "Python3 not found, installing..."
    
    if [ "$OS" == "macos" ]; then
        brew install python3
    elif [ "$OS_ID" == "ubuntu" ] || [ "$OS_ID" == "debian" ]; then
        sudo apt-get update
        sudo apt-get install -y python3 python3-pip
    elif [ "$OS_ID" == "centos" ] || [ "$OS_ID" == "rhel" ] || [ "$OS_ID" == "fedora" ]; then
        if command_exists dnf; then
            sudo dnf install -y python3 python3-pip
        else
            sudo yum install -y python3 python3-pip
        fi
    fi
    
    if command_exists python3; then
        print_info "Python3 installation succeeded"
    else
        print_error "Python3 installation failed"
        exit 1
    fi
}

# Install pip
install_pip() {
    if command_exists pip3; then
        print_info "pip3 already installed"
        return 0
    fi
    
    print_warn "pip3 not found, installing..."
    
    if [ "$OS" == "macos" ]; then
        python3 -m ensurepip --upgrade
    elif [ "$OS_ID" == "ubuntu" ] || [ "$OS_ID" == "debian" ]; then
        sudo apt-get install -y python3-pip
    elif [ "$OS_ID" == "centos" ] || [ "$OS_ID" == "rhel" ] || [ "$OS_ID" == "fedora" ]; then
        if command_exists dnf; then
            sudo dnf install -y python3-pip
        else
            sudo yum install -y python3-pip
        fi
    fi
    
    if command_exists pip3; then
        print_info "pip3 installation succeeded"
    else
        print_error "pip3 installation failed"
        exit 1
    fi
}

# Install Ansible
install_ansible() {
    if command_exists ansible-playbook; then
        local version=$(ansible-playbook --version | head -1 | awk '{print $2}')
        print_info "Ansible already installed ($version)"
        return 0
    fi
    
    print_warn "Ansible not found, installing..."
    
    pip3 install --user ansible
    
    # Ensure PATH includes the user-local bin directory
    if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        print_warn "Add $HOME/.local/bin to PATH"
        print_warn "Run: export PATH=\"$HOME/.local/bin:$PATH\""
        print_warn "Or add to ~/.bashrc or ~/.zshrc"
    fi
    
    if command_exists ansible-playbook; then
        print_info "Ansible installation succeeded"
    else
        # Try using the full path
        if [ -f "$HOME/.local/bin/ansible-playbook" ]; then
            print_warn "Ansible is installed but not on PATH"
            print_warn "Run: export PATH=\"$HOME/.local/bin:$PATH\""
        else
            print_error "Ansible installation failed"
            exit 1
        fi
    fi
}

# Install Python packages
install_python_packages() {
    print_section "Install Python packages"
    
    local packages=("ansible" "requests")
    
    for package in "${packages[@]}"; do
        if python3 -c "import ${package//-/_}" 2>/dev/null; then
            print_info "Python package $package already installed"
        else
            print_warn "Installing Python package: $package"
            pip3 install --user "$package"
            print_info "Python package $package installed"
        fi
    done
}

# Install SoftHSM2
install_softhsm() {
    if command_exists softhsm2-util; then
        print_info "SoftHSM2 already installed"
        return 0
    fi
    
    print_warn "SoftHSM2 not found, installing..."
    
    if [ "$OS" == "macos" ]; then
        brew install softhsm
    elif [ "$OS_ID" == "ubuntu" ] || [ "$OS_ID" == "debian" ]; then
        sudo apt-get update
        sudo apt-get install -y softhsm2
    elif [ "$OS_ID" == "centos" ] || [ "$OS_ID" == "rhel" ] || [ "$OS_ID" == "fedora" ]; then
        if command_exists dnf; then
            sudo dnf install -y softhsm
        else
            sudo yum install -y softhsm
        fi
    fi
    
    if command_exists softhsm2-util; then
        print_info "SoftHSM2 installation succeeded"
    else
        print_warn "SoftHSM2 installation failed (required if using HSM)"
    fi
}

# Install Ansible Collection
install_ansible_collection() {
    print_section "Install Ansible Collection"
    
    # Get project root (fabric-x-network)
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
    REQUIREMENTS_FILE="$PROJECT_ROOT/requirements.yml"
    
    if [ ! -f "$REQUIREMENTS_FILE" ]; then
        print_warn "requirements.yml not found; skipping Ansible Collection install"
        return 0
    fi
    
    # Ensure ansible-galaxy is on PATH
    if ! command_exists ansible-galaxy; then
        if [ -f "$HOME/.local/bin/ansible-galaxy" ]; then
            export PATH="$HOME/.local/bin:$PATH"
        else
            print_error "ansible-galaxy not found; install Ansible first"
            return 1
        fi
    fi
    
    print_info "Installing Ansible Collection (from $REQUIREMENTS_FILE)..."
    ansible-galaxy collection install -r "$REQUIREMENTS_FILE" -p "$HOME/.ansible/collections"
    
    if ansible-galaxy collection list 2>/dev/null | grep -q "hyperledger.fabricx"; then
        local version=$(ansible-galaxy collection list 2>/dev/null | grep "hyperledger.fabricx" | awk '{print $2}')
        print_info "Ansible Collection hyperledger.fabricx installed ($version)"
    else
        print_warn "Ansible Collection may not have installed correctly"
    fi
}

# Main
main() {
    echo "=========================================="
    echo "  Dependency installation"
    echo "=========================================="
    echo ""
    
    detect_os
    
    print_section "Check and install system dependencies"
    
    if [ "$OS" == "macos" ]; then
        check_homebrew
    fi
    
    install_go
    install_docker
    install_git
    install_python3
    install_pip
    
    print_section "Install developer tools"
    install_ansible
    install_softhsm
    
    print_section "Install Python dependencies"
    install_python_packages
    
    print_section "Install Ansible Collection"
    install_ansible_collection
    
    echo ""
    print_section "Installation complete"
    echo ""
    print_info "All dependencies installed!"
    echo ""
    echo "Next steps:"
    echo "  1. If Go is not on PATH, run:"
    echo "     export PATH=\$PATH:/usr/local/go/bin"
    echo "     Or add to ~/.bashrc: echo 'export PATH=\$PATH:/usr/local/go/bin' >> ~/.bashrc"
    echo "     Then re-login or run: source ~/.bashrc"
    echo ""
    echo "  2. If Docker group changes were made, re-login or run: newgrp docker"
    echo ""
    echo "  3. If Ansible is not on PATH, run: export PATH=\"$HOME/.local/bin:\$PATH\""
    echo ""
    echo "  4. Run dependency check: ./scripts/check-prerequisites.sh"
    echo ""
    echo "  5. Run setup: make setup"
    echo ""
}

# 运行主函数
main
