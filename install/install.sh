#!/bin/bash
# Katsh Installer for Unix/Linux/macOS
# This script builds and installs Katsh to your PATH

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${CYAN}  Katsh Installer for Unix/Linux/macOS${NC}"
echo -e "${CYAN}═══════════════════════════════════════════════════════════${NC}"
echo ""

# Detect install directory
DEFAULT_INSTALL_DIR="$HOME/.local/bin"
INSTALL_DIR="${INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${YELLOW}[1/4] Checking Go installation...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${RED}ERROR: Go is not installed or not in PATH${NC}"
    echo -e "${RED}Please install Go from: https://go.dev/dl/${NC}"
    exit 1
fi
GO_VERSION=$(go version)
echo -e "  Found: ${GREEN}$GO_VERSION${NC}"

echo -e "${YELLOW}[2/4] Building Katsh...${NC}"
cd "$PROJECT_ROOT"

# Build for current platform
go build -o katsh .
if [ $? -ne 0 ]; then
    echo -e "${RED}ERROR: Build failed!${NC}"
    exit 1
fi
echo -e "  ${GREEN}Build successful!${NC}"

echo -e "${YELLOW}[3/4] Installing to $INSTALL_DIR...${NC}"

# Create installation directory
mkdir -p "$INSTALL_DIR"

# Copy binary
cp katsh "$INSTALL_DIR/katsh"
chmod +x "$INSTALL_DIR/katsh"

# Clean up build artifact
rm -f katsh

echo -e "  Installed to: ${GREEN}$INSTALL_DIR${NC}"

echo -e "${YELLOW}[4/4] Adding to PATH...${NC}"

# Check if already in PATH
if [[ ":$PATH:" == *":$INSTALL_DIR:"* ]]; then
    echo -e "  Already in PATH${NC}"
else
    # Add to PATH in shell rc files
    export PATH="$INSTALL_DIR:$PATH"
    
    # Detect shell and add to appropriate rc file
    SHELL_RC=""
    if [ -n "$ZSH_VERSION" ]; then
        SHELL_RC="$HOME/.zshrc"
    elif [ -n "$BASH_VERSION" ]; then
        SHELL_RC="$HOME/.bashrc"
    elif [ -n "$FISH_VERSION" ]; then
        SHELL_RC="$HOME/.config/fish/config.fish"
    fi
    
    if [ -n "$SHELL_RC" ]; then
        # Check if already added
        if ! grep -q "$INSTALL_DIR" "$SHELL_RC" 2>/dev/null; then
            echo "" >> "$SHELL_RC"
            echo "# Katsh" >> "$SHELL_RC"
            echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$SHELL_RC"
            echo -e "  Added to ${GREEN}$SHELL_RC${NC} (persistent)"
        else
            echo -e "  Already configured in shell config${NC}"
        fi
    else
        echo -e "  ${YELLOW}Could not detect shell - please add to your PATH manually:${NC}"
        echo -e "    export PATH=\"$INSTALL_DIR:\$PATH\""
    fi
fi

echo ""
echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}  Installation complete!${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "Run '${CYAN}katsh${NC}' to start the shell."
echo ""
echo -e "${YELLOW}Note: Run 'source ~/.zshrc' or 'source ~/.bashrc' for PATH changes to take effect.${NC}"
