#!/bin/bash

# MCP-Obsidian-Go Uninstaller
# Usage: ./uninstall.sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
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

# Main uninstallation
main() {
    echo "========================================"
    echo "  MCP-Obsidian-Go Uninstaller"
    echo "========================================"
    echo ""
    
    # Check if binary exists
    if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        print_info "Removing binary from ${INSTALL_DIR}..."
        rm -f "${INSTALL_DIR}/${BINARY_NAME}"
        print_success "Binary removed!"
    else
        print_warning "Binary not found in ${INSTALL_DIR}"
    fi
    
    # Remove indexer binary if exists
    if [ -f "${INSTALL_DIR}/mcp-obsidian-indexer" ]; then
        print_info "Removing indexer binary..."
        rm -f "${INSTALL_DIR}/mcp-obsidian-indexer"
        print_success "Indexer binary removed!"
    fi
    
    # Ask about config
    echo ""
    if [ -d "$CONFIG_DIR" ]; then
        print_warning "Configuration directory found: $CONFIG_DIR"
        echo "This contains your settings and any cached data."
        read -p "Remove configuration directory? [y/N]: " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$CONFIG_DIR"
            print_success "Configuration directory removed!"
        else
            print_info "Configuration preserved at: $CONFIG_DIR"
        fi
    fi
    
    # Summary
    echo ""
    echo "========================================"
    print_success "Uninstallation complete!"
    echo "========================================"
    echo ""
    print_info "The following have been removed:"
    echo "  • ${INSTALL_DIR}/${BINARY_NAME}"
    echo "  • ${INSTALL_DIR}/mcp-obsidian-indexer (if existed)"
    echo ""
    print_info "To complete uninstallation:"
    echo "1. Remove the 'obsidian' MCP server from your Claude Desktop config"
    echo "2. Restart Claude Desktop"
    echo ""
}

# Run main
main
