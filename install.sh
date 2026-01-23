#!/bin/bash
set -e

REPO="ykushch/ask"
INSTALL_DIR="$HOME/.local/bin"
TMP_DIR=$(mktemp -d)

cleanup() {
    rm -rf "$TMP_DIR"
}
trap cleanup EXIT

DEFAULT_MODEL="qwen2.5-coder:7b"

echo "Installing ask..."
echo ""

# --- Install Ollama if not present ---
if command -v ollama &> /dev/null; then
    echo "✓ Ollama already installed"
else
    echo "Installing Ollama..."
    if [ "$(uname -s)" = "Darwin" ]; then
        # macOS: download the app
        if command -v brew &> /dev/null; then
            brew install --cask ollama
        else
            echo "Downloading Ollama for macOS..."
            curl -fsSL https://ollama.com/download/Ollama-darwin.zip -o "$TMP_DIR/ollama.zip"
            unzip -q "$TMP_DIR/ollama.zip" -d /Applications
            echo "Ollama.app installed to /Applications"
        fi
    else
        # Linux: use the official install script
        curl -fsSL https://ollama.com/install.sh | sh
    fi

    if command -v ollama &> /dev/null; then
        echo "✓ Ollama installed"
    else
        echo "Warning: Ollama install may require a new shell. Continuing..."
    fi
fi

# --- Start Ollama if not running ---
if curl -s http://localhost:11434/ > /dev/null 2>&1; then
    echo "✓ Ollama is running"
else
    echo "Starting Ollama..."
    if [ "$(uname -s)" = "Darwin" ]; then
        open -a Ollama 2>/dev/null || ollama serve &>/dev/null &
    else
        ollama serve &>/dev/null &
    fi
    # Wait for it to be ready
    for i in $(seq 1 15); do
        if curl -s http://localhost:11434/ > /dev/null 2>&1; then
            break
        fi
        sleep 1
    done
    if curl -s http://localhost:11434/ > /dev/null 2>&1; then
        echo "✓ Ollama is running"
    else
        echo "Warning: Could not start Ollama. You may need to start it manually."
    fi
fi

# --- Pull default model ---
if command -v ollama &> /dev/null; then
    if ollama list 2>/dev/null | grep -q "$DEFAULT_MODEL"; then
        echo "✓ Model $DEFAULT_MODEL already available"
    else
        echo "Pulling model $DEFAULT_MODEL (this may take a few minutes on first run)..."
        ollama pull "$DEFAULT_MODEL"
        echo "✓ Model $DEFAULT_MODEL ready"
    fi
fi

echo ""

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    linux|darwin) ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Try downloading pre-built binary from GitHub releases
BINARY_URL="https://github.com/$REPO/releases/latest/download/ask-${OS}-${ARCH}"
echo "Downloading ask for ${OS}/${ARCH}..."

if curl -fsSL -o "$TMP_DIR/ask" "$BINARY_URL" 2>/dev/null; then
    chmod +x "$TMP_DIR/ask"
else
    # Fallback: build from source
    echo "No pre-built binary found. Building from source..."
    if ! command -v go &> /dev/null; then
        echo "Error: Go is required to build from source."
        echo "Install Go: https://go.dev/dl/"
        exit 1
    fi
    git clone --quiet --depth 1 "https://github.com/$REPO.git" "$TMP_DIR/src"
    cd "$TMP_DIR/src"
    go build -o "$TMP_DIR/ask" .
fi

# Install binary
mkdir -p "$INSTALL_DIR"
mv "$TMP_DIR/ask" "$INSTALL_DIR/ask"
echo "Installed to $INSTALL_DIR/ask"

# Ensure ~/.local/bin is in PATH
add_to_path() {
    local rc_file="$1"
    if [ -f "$rc_file" ] && grep -q '.local/bin' "$rc_file" 2>/dev/null; then
        return
    fi
    if [ -f "$rc_file" ] || [ "$2" = "create" ]; then
        echo '' >> "$rc_file"
        echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$rc_file"
    fi
}

PATH_ADDED=""
if ! echo "$PATH" | grep -q "$HOME/.local/bin"; then
    shell_name=$(basename "$SHELL")
    case "$shell_name" in
        zsh)  add_to_path "$HOME/.zshrc" create ; PATH_ADDED="zshrc" ;;
        bash) add_to_path "$HOME/.bashrc" create ; PATH_ADDED="bashrc" ;;
        *)    add_to_path "$HOME/.profile" create ; PATH_ADDED="profile" ;;
    esac
fi

echo ""
echo "✓ ask installed successfully!"
echo ""

if [ -n "$PATH_ADDED" ]; then
    echo "⚠ To start using ask, run this in your terminal:"
    echo ""
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo ""
    echo "  (or open a new terminal window)"
    echo ""
fi

echo "Usage:"
echo "  ask list all go files in this directory"
echo "  ask                  # interactive mode"
echo "  ask --model llama3 show disk usage"
echo ""
echo "Default model: $DEFAULT_MODEL"
echo "Ollama host: ${OLLAMA_HOST:-http://localhost:11434}"
