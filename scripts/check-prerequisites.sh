#!/bin/bash
#
# Prerequisites Check Script
# This script checks if all required dependencies are installed before running make setup
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Track if any checks fail
FAILED=0

echo "Checking prerequisites for fabric-x-network setup..."
echo ""

# Check command availability
check_command() {
    local cmd=$1
    local name=${2:-$cmd}
    local required=${3:-true}
    
    if command -v "$cmd" &> /dev/null; then
        local version=""
        case "$cmd" in
            go)
                version=$(go version | awk '{print $3}')
                print_info "$name is installed ($version)"
                ;;
            docker)
                version=$(docker --version | awk '{print $3}' | sed 's/,//')
                print_info "$name is installed ($version)"
                ;;
            git)
                version=$(git --version | awk '{print $3}')
                print_info "$name is installed ($version)"
                ;;
            python3)
                version=$(python3 --version | awk '{print $2}')
                print_info "$name is installed ($version)"
                ;;
            ansible-playbook)
                version=$(ansible-playbook --version | head -1 | awk '{print $2}')
                print_info "$name is installed ($version)"
                ;;
            softhsm2-util)
                print_info "$name is installed"
                ;;
            *)
                print_info "$name is installed"
                ;;
        esac
        return 0
    else
        if [ "$required" = "true" ]; then
            print_error "$name is not installed"
            FAILED=1
            return 1
        else
            print_warn "$name is not installed (optional)"
            return 0
        fi
    fi
}

# Check Python module
check_python_module() {
    local module=$1
    local name=${2:-$module}
    
    if python3 -c "import $module" 2>/dev/null; then
        print_info "Python module $name is installed"
        return 0
    else
        print_error "Python module $name is not installed"
        echo "  Install with: pip install $module"
        FAILED=1
        return 1
    fi
}

# Check Ansible collection
check_ansible_collection() {
    if ansible-galaxy collection list 2>/dev/null | grep -q "hyperledger.fabricx"; then
        local version=$(ansible-galaxy collection list 2>/dev/null | grep "hyperledger.fabricx" | awk '{print $2}')
        print_info "Ansible Collection hyperledger.fabricx is installed ($version)"
        return 0
    else
        print_error "Ansible Collection hyperledger.fabricx is not installed"
        echo "  Install with: ansible-galaxy collection install hyperledger.fabricx"
        FAILED=1
        return 1
    fi
}

# Check SoftHSM library
check_softhsm_library() {
    local found=false
    local paths=(
        "/opt/homebrew/opt/softhsm@2.5/lib/softhsm/libsofthsm2.so"
        "/opt/homebrew/lib/softhsm2/libsofthsm2.so"
        "/usr/local/opt/softhsm/lib/softhsm/libsofthsm2.so"
        "/usr/lib/softhsm/libsofthsm2.so"
        "/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so"
        "/usr/lib64/softhsm/libsofthsm2.so"
    )
    
    for path in "${paths[@]}"; do
        if [ -f "$path" ]; then
            print_info "SoftHSM library file found: $path"
            found=true
            break
        fi
    done
    
    if [ "$found" = "false" ]; then
        print_warn "SoftHSM library file not found (required when using HSM)"
        echo "  Common paths:"
        echo "    macOS: /opt/homebrew/opt/softhsm@2.5/lib/softhsm/libsofthsm2.so"
        echo "    Linux: /usr/lib/softhsm/libsofthsm2.so"
    fi
}

# Check Go version
check_go_version() {
    if command -v go &> /dev/null; then
        local go_version=$(go version | awk '{print $3}' | sed 's/go//')
        local major=$(echo "$go_version" | cut -d. -f1)
        local minor=$(echo "$go_version" | cut -d. -f2)
        
        if [ "$major" -gt 1 ] || ([ "$major" -eq 1 ] && [ "$minor" -ge 21 ]); then
            print_info "Go version meets requirement ($go_version)"
        else
            print_warn "Go version may be too low ($go_version); use Go 1.21 or newer"
        fi
    fi
}

# Check Docker daemon
check_docker_daemon() {
    if command -v docker &> /dev/null; then
        if docker info &> /dev/null; then
            print_info "Docker daemon is running"
        else
            print_error "Docker daemon is not running"
            echo "  Start Docker Desktop or the Docker daemon"
            FAILED=1
        fi
    fi
}

# Check project files
check_project_files() {
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local project_root="$(cd "$script_dir/.." && pwd)"
    
    local files=(
        "Makefile"
    )
    
    local missing=0
    for file in "${files[@]}"; do
        if [ -f "$project_root/$file" ]; then
            print_info "Project file present: $file"
        else
            print_error "Project file missing: $file"
            missing=1
        fi
    done
    
    if [ $missing -eq 1 ]; then
        FAILED=1
    fi
}

# Main checks
print_section "Check required CLI tools"
check_command go "Go"
check_command docker "Docker"
check_command git "Git"
check_command python3 "Python3"
check_command ansible-playbook "Ansible"
check_command softhsm2-util "SoftHSM2 Util"
check_command fxconfig "fxconfig (external)" false
check_command cryptogen "cryptogen (external)" false

echo ""
print_section "Check Python modules"
check_python_module requests "requests"

echo ""
print_section "Check Ansible Collection"
check_ansible_collection

echo ""
print_section "Check SoftHSM library file"
check_softhsm_library

echo ""
print_section "Check Go version"
check_go_version

echo ""
print_section "Check Docker daemon"
check_docker_daemon

echo ""
print_section "Check project files"
check_project_files

echo ""
if [ $FAILED -eq 0 ]; then
    print_info "All prerequisite checks passed!"
    echo ""
    echo "You can now run:"
    echo "  make setup"
    exit 0
else
    print_error "Some prerequisites are missing; please install them first"
    echo ""
    echo "Install guide:"
    echo "  macOS:"
    echo "    brew install go docker git python3 softhsm"
    echo "    pip3 install ansible requests"
    echo "    ansible-galaxy collection install hyperledger.fabricx"
    echo ""
    echo "  Ubuntu/Debian:"
    echo "    sudo apt-get install golang-go docker.io git python3 python3-pip softhsm2"
    echo "    pip3 install ansible requests"
    echo "    ansible-galaxy collection install hyperledger.fabricx"
    exit 1
fi

