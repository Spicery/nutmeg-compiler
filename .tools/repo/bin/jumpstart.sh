#!/bin/bash
set -e

# Jumpstart script to install prerequisites for the Nutmeg compiler toolchain.
# Supports Debian/Ubuntu Linux and macOS.

echo "=========================================="
echo "Nutmeg Compiler Toolchain - Jumpstart"
echo "=========================================="
echo ""

# Detect OS.
OS_TYPE="unknown"
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    OS_TYPE="linux"
    echo "Detected: Linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS_TYPE="macos"
    echo "Detected: macOS"
else
    echo "Warning: Unsupported OS type: $OSTYPE"
    echo "This script supports Linux (Debian/Ubuntu) and macOS."
    echo ""
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo ""

# Check for package manager and sudo (Linux only).
if [[ "$OS_TYPE" == "linux" ]]; then
    if ! sudo -v; then
        echo "Error: This script requires sudo access to install system packages."
        exit 1
    fi
    
    if ! command -v apt-get &> /dev/null; then
        echo "Warning: apt-get not found. This script is designed for Debian/Ubuntu."
        echo "Some packages may fail to install."
    fi
elif [[ "$OS_TYPE" == "macos" ]]; then
    if ! command -v brew &> /dev/null; then
        echo "Homebrew not found. Installing Homebrew..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        echo "✓ Homebrew installed"
        echo ""
    else
        echo "✓ Homebrew is already installed"
    fi
fi

echo "Installing prerequisites..."
echo ""

# Install Go if not already installed.
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo "✓ Go is already installed (version $GO_VERSION)"
else
    echo "Installing Go..."
    if [[ "$OS_TYPE" == "linux" ]]; then
        sudo apt-get update
        sudo apt-get install -y golang-go
    elif [[ "$OS_TYPE" == "macos" ]]; then
        brew install go
    fi
    echo "✓ Go installed"
fi

# Install Python3 if not already installed.
if command -v python3 &> /dev/null; then
    PYTHON_VERSION=$(python3 --version | awk '{print $2}')
    echo "✓ Python 3 is already installed (version $PYTHON_VERSION)"
else
    echo "Installing Python 3..."
    if [[ "$OS_TYPE" == "linux" ]]; then
        sudo apt-get install -y python3
    elif [[ "$OS_TYPE" == "macos" ]]; then
        brew install python3
    fi
    echo "✓ Python 3 installed"
fi

# Install uv (Python package manager).
if command -v uv &> /dev/null; then
    UV_VERSION=$(uv --version | awk '{print $2}')
    echo "✓ uv is already installed (version $UV_VERSION)"
else
    echo "Installing uv..."
    curl -LsSf https://astral.sh/uv/install.sh | sh
    echo "✓ uv installed"
    echo "  Note: You may need to restart your shell or run: source $HOME/.cargo/env"
fi

# Install Just.
if command -v just &> /dev/null; then
    JUST_VERSION=$(just --version | awk '{print $2}')
    echo "✓ just is already installed (version $JUST_VERSION)"
else
    echo "Installing just..."
    if [[ "$OS_TYPE" == "linux" ]]; then
        curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ~/.local/bin
        echo "✓ just installed to ~/.local/bin"
        echo "  Note: Make sure ~/.local/bin is in your PATH"
    elif [[ "$OS_TYPE" == "macos" ]]; then
        brew install just
        echo "✓ just installed"
    fi
fi

# Install gosec (optional but recommended).
if command -v gosec &> /dev/null || [ -f "$HOME/go/bin/gosec" ]; then
    if [ -f "$HOME/go/bin/gosec" ]; then
        GOSEC_VERSION=$($HOME/go/bin/gosec --version 2>&1 | head -n1 || echo "unknown")
    else
        GOSEC_VERSION=$(gosec --version 2>&1 | head -n1 || echo "unknown")
    fi
    echo "✓ gosec is already installed ($GOSEC_VERSION)"
else
    echo "Installing gosec..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    echo "✓ gosec installed to ~/go/bin"
    echo "  Note: Make sure ~/go/bin is in your PATH"
fi

# Install golangci-lint (optional but recommended).
if command -v golangci-lint &> /dev/null || [ -f "$HOME/.tools/ext/bin/golangci-lint" ]; then
    if [ -f "$HOME/.tools/ext/bin/golangci-lint" ]; then
        LINT_VERSION=$($HOME/.tools/ext/bin/golangci-lint --version 2>&1 | head -n1 || echo "unknown")
    else
        LINT_VERSION=$(golangci-lint --version 2>&1 | head -n1 || echo "unknown")
    fi
    echo "✓ golangci-lint is already installed ($LINT_VERSION)"
else
    echo "Installing golangci-lint..."
    mkdir -p ~/.tools/ext/bin
    GOBIN=~/.tools/ext/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    echo "✓ golangci-lint installed to ~/.tools/ext/bin"
fi

echo ""
echo "=========================================="
echo "Installation Complete!"
echo "=========================================="
echo ""
echo "Next steps:"

if [[ "$OS_TYPE" == "linux" ]]; then
    echo "  1. Make sure the following directories are in your PATH:"
    echo "     - ~/.local/bin (for just)"
    echo "     - ~/go/bin (for gosec)"
    echo "     - ~/.cargo/bin (for uv, if installed)"
    echo ""
    echo "  2. You may need to restart your shell or run:"
    echo "     export PATH=\"\$HOME/.local/bin:\$HOME/go/bin:\$HOME/.cargo/bin:\$PATH\""
elif [[ "$OS_TYPE" == "macos" ]]; then
    echo "  1. Make sure the following directories are in your PATH:"
    echo "     - ~/go/bin (for gosec)"
    echo "     - ~/.cargo/bin (for uv, if installed)"
    echo ""
    echo "  2. You may need to restart your shell or run:"
    echo "     export PATH=\"\$HOME/go/bin:\$HOME/.cargo/bin:\$PATH\""
fi

echo ""
echo "  3. Verify installation by running:"
echo "     just test"
echo ""
