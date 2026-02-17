#!/bin/bash

# MCP-Obsidian-Go Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/xvierd/mcp-obsidian-go/main/install.sh | bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="xvierd/mcp-obsidian-go"
BINARY_NAME="mcp-obsidian-go"
INSTALL_DIR="${HOME}/.local/bin"
CONFIG_DIR="${HOME}/.config/mcp-obsidian"

# Print functions
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$ARCH" in
        x86_64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    case "$OS" in
        linux)
            PLATFORM="linux-${ARCH}"
            ;;
        darwin)
            PLATFORM="darwin-${ARCH}"
            ;;
        *)
            print_error "Unsupported OS: $OS"
            exit 1
            ;;
    esac
    
    print_info "Detected platform: $PLATFORM"
}

# Install from local build directory (for development/distribution)
install_from_local() {
    # Get the directory where the script is located
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    
    # Check for local binary in script directory
    if [ -f "${SCRIPT_DIR}/build/${BINARY_NAME}" ]; then
        print_info "Found local binary in script directory..."
        cp "${SCRIPT_DIR}/build/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
        return 0
    fi
    
    # Check for local binary in current directory
    if [ -f "build/${BINARY_NAME}" ]; then
        print_info "Found local binary in current directory..."
        cp "build/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
        return 0
    fi
    
    return 1
}

# Check if Go is installed
check_go() {
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        print_info "Go installed: $GO_VERSION"
        return 0
    else
        return 1
    fi
}

# Install from source (if Go is available)
install_from_source() {
    print_info "Installing from source..."
    
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    # Clone repository
    print_info "Cloning repository..."
    git clone --depth 1 "https://github.com/${REPO}.git" 2>/dev/null || {
        print_warning "Could not clone repository."
        return 1
    }
    
    cd mcp-obsidian-go
    
    # Build
    print_info "Building binary..."
    make build
    
    # Copy binary
    cp "build/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    
    # Cleanup
    cd "$HOME"
    rm -rf "$TEMP_DIR"
    
    print_success "Built and installed from source!"
}

# Install from GitHub release
install_from_release() {
    print_info "Downloading from GitHub releases..."
    
    # Get latest release URL
    RELEASE_URL="https://github.com/${REPO}/releases/latest/download/${BINARY_NAME}-${PLATFORM}"
    
    print_info "Downloading: $RELEASE_URL"
    
    if command -v curl &> /dev/null; then
        curl -fsSL "$RELEASE_URL" -o "${INSTALL_DIR}/${BINARY_NAME}" || {
            print_error "Failed to download binary"
            print_info "The repository may not have releases yet."
            return 1
        }
    elif command -v wget &> /dev/null; then
        wget -q "$RELEASE_URL" -O "${INSTALL_DIR}/${BINARY_NAME}" || {
            print_error "Failed to download binary"
            return 1
        }
    else
        print_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
    
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    print_success "Downloaded and installed!"
}

# Manual installation instructions
manual_install_instructions() {
    echo ""
    print_warning "Automatic installation failed."
    echo ""
    print_info "Manual installation instructions:"
    echo ""
    echo "1. Download the binary for your platform:"
    echo "   https://github.com/${REPO}/releases"
    echo ""
    echo "2. Place it in your PATH, for example:"
    echo "   mkdir -p ~/.local/bin"
    echo "   mv mcp-obsidian-go ~/.local/bin/"
    echo "   chmod +x ~/.local/bin/mcp-obsidian-go"
    echo ""
    echo "3. Ensure ~/.local/bin is in your PATH:"
    echo "   export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo "   # Add this to your ~/.zshrc or ~/.bashrc"
    echo ""
}

# Create config directory
setup_config() {
    print_info "Setting up configuration directory..."
    mkdir -p "$CONFIG_DIR"
    
    # Create example config
    cat > "${CONFIG_DIR}/example-config.yaml" << 'EOF'
# MCP-Obsidian-Go Configuration
# Copy this to ~/.config/mcp-obsidian/config.yaml and fill in your API key

obsidian:
  api_key: "your-api-key-here"
  host: "127.0.0.1"
  port: 27124

cache:
  enabled: true
  ttl: 30s
  size: 1000
EOF
    
    print_success "Config directory created: $CONFIG_DIR"
}

# Print Claude Desktop configuration
print_claude_config() {
    BINARY_PATH="${INSTALL_DIR}/${BINARY_NAME}"
    
    echo ""
    echo "========================================"
    print_success "Installation complete!"
    echo "========================================"
    echo ""
    print_info "Add this to your Claude Desktop config:"
    echo ""
    echo -e "${GREEN}macOS:${NC}"
    echo "~/Library/Application Support/Claude/claude_desktop_config.json"
    echo ""
    echo -e "${GREEN}Windows:${NC}"
    echo "%APPDATA%/Claude/claude_desktop_config.json"
    echo ""
    echo -e "${GREEN}Linux:${NC}"
    echo "~/.config/Claude/claude_desktop_config.json"
    echo ""
    echo -e "${YELLOW}Configuration JSON:${NC}"
    cat << EOF
{
  "mcpServers": {
    "obsidian": {
      "command": "${BINARY_PATH}",
      "env": {
        "OBSIDIAN_API_KEY": "your-api-key-here"
      }
    }
  }
}
EOF
    echo ""
    print_warning "IMPORTANT: Replace 'your-api-key-here' with your actual Obsidian API key!"
    echo ""
    print_info "To get your API key:"
    echo "1. Open Obsidian"
    echo "2. Install 'Local REST API' plugin from Community Plugins"
    echo "3. Go to Settings → Local REST API → Copy API Key"
    echo ""
    print_info "Example config saved to: ${CONFIG_DIR}/example-config.yaml"
}

# Main installation
main() {
    echo "========================================"
    echo "  MCP-Obsidian-Go Installer"
    echo "========================================"
    echo ""
    
    # Detect platform
    detect_platform
    echo ""
    
    # Create install directory
    mkdir -p "$INSTALL_DIR"
    
    # First, try to install from local build (for development/testing)
    if install_from_local; then
        print_success "Installed from local build!"
    # Otherwise, try remote methods
    elif check_go; then
        install_from_source || {
            print_warning "Source build failed, trying release..."
            install_from_release || manual_install_instructions
        }
    else
        install_from_release || manual_install_instructions
    fi
    
    # Setup config
    setup_config
    
    # Print instructions
    print_claude_config
}

# Run main
main
