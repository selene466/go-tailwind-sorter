#!/usr/bin/env bash
set -eo pipefail

# ------------------------------------------
# tailwind-sorter Installer
# ------------------------------------------

GITHUB_REPO="selene466/go-tailwind-sorter"
INSTALL_DIR="${TWVM_INSTALL_DIR:-$HOME/.local/bin}"

# Colors
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${CYAN}Installing twvm...${NC}"

detect_arch() {
  local ARCH="$(uname -m)"

  case $ARCH in
  x86_64 | amd64)
    echo "amd64"
    ;;
  aarch64 | arm64)
    echo "arm64"
    ;;
  *)
    echo -e "${RED}Error: Unsupported platform architecture: $ARCH${NC}" >&2
    exit 1
    ;;
  esac
}

detect_os() {
  local OS="$(uname -s | tr '[:upper:]' '[:lower:]')"

  case $OS in
  linux)
    echo "linux"
    ;;
  darwin)
    echo "darwin"
    ;;
  mingw* | msys* | cygwin*)
    echo "windows"
    ;;
  *)
    echo -e "${RED}Error: Unsupported OS: $OS${NC}" >&2
    exit 1
    ;;
  esac
}

fetch_latest_version() {
  local API_RESPONSE=""

  if command -v curl >/dev/null 2>&1; then
    API_RESPONSE=$(curl -sL "https://api.github.com/repos/$GITHUB_REPO/releases/latest")
  elif command -v wget >/dev/null 2>&1; then
    API_RESPONSE=$(wget -qO- "https://api.github.com/repos/$GITHUB_REPO/releases/latest")
  else
    echo -e "${RED}Error: curl or wget is required to download twvm.${NC}" >&2
    exit 1
  fi

  if echo "$API_RESPONSE" | grep -q "API rate limit exceeded"; then
    echo -e "${RED}Error: GitHub API rate limit exceeded. Please try again later.${NC}" >&2
    exit 1
  fi

  local TAG=$(echo "$API_RESPONSE" | grep '"tag_name":' | head -n 1 | sed -E 's/.*"([^"]+)".*/\1/')

  if [ -z "$TAG" ]; then
    echo -e "${RED}Error: Failed to fetch the latest release version.${NC}" >&2
    exit 1
  fi

  echo "$TAG"
}

unquarantine_mac() {
  local DEST="$1"

  echo -e "\n${YELLOW}macOS Gatekeeper might block this binary because it is unsigned.${NC}"

  if command -v xattr >/dev/null 2>&1; then
    if [ -c /dev/tty ]; then
      read -p "Do you want to remove the quarantine attribute to allow execution? [Y/n] " -r unquarantine </dev/tty
      unquarantine=${unquarantine:-Y}

      if [[ "$unquarantine" =~ ^[Yy]$ ]]; then
        xattr -d com.apple.quarantine "$DEST" 2>/dev/null || true
        echo -e "${GREEN}Quarantine attribute removed.${NC}"
      else
        echo -e "${YELLOW}Skipped. If you get an error later, run: xattr -d com.apple.quarantine $DEST${NC}"
      fi
    else
      xattr -d com.apple.quarantine "$DEST" 2>/dev/null || true
    fi
  fi
}

setup_path() {
  if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "\n${YELLOW}Warning: $INSTALL_DIR is not in your PATH.${NC}"

    local USER_SHELL=$(basename "${SHELL:-bash}")
    local PROFILE=""

    if [[ "$USER_SHELL" == "zsh" ]]; then
      PROFILE="$HOME/.zshrc"
    elif [[ "$USER_SHELL" == "bash" ]]; then
      if [ -f "$HOME/.bash_profile" ]; then
        PROFILE="$HOME/.bash_profile"
      else
        PROFILE="$HOME/.bashrc"
      fi
    elif [ -f "$HOME/.profile" ]; then
      PROFILE="$HOME/.profile"
    fi

    if [ -n "$PROFILE" ]; then
      echo "Attempting to add to $PROFILE..."
      echo -e "\n# tailwind-sorter (Tailwind Sorter)" >>"$PROFILE"
      echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >>"$PROFILE"
      echo -e "${GREEN}Added $INSTALL_DIR to $PROFILE.${NC}"
      echo -e "Please run ${CYAN}source $PROFILE${NC} or restart your terminal."
    else
      echo -e "Please manually add ${CYAN}export PATH=\"$INSTALL_DIR:\$PATH\"${NC} to your shell profile."
    fi
  fi
}

echo -e "${CYAN}Installing tailwind-sorter...${NC}"

OS_NAME="$(detect_os)"

ARCH_NAME="$(detect_arch)"

if [[ -z "$VERSION" || "$VERSION" == "latest" ]]; then
  echo "Warning: No VERSION provided or 'latest' requested."
  echo "Resolving..."
  VERSION="$(fetch_latest_version)"
fi

echo -e "Target version: ${YELLOW}$VERSION${NC}"

BINARY_NAME="tailwind-sorter-$OS_NAME-$ARCH_NAME"
DEST_FILE="$INSTALL_DIR/tailwind-sorter"

if [[ "$OS_NAME" == "windows" ]]; then
  BINARY_NAME+=".exe"
  DEST_FILE+=".exe"
fi

DOWNLOAD_URL="https://github.com/$GITHUB_REPO/releases/download/$VERSION/$BINARY_NAME"

mkdir -p "$INSTALL_DIR"

echo "Downloading $BINARY_NAME..."
if command -v curl >/dev/null 2>&1; then
  curl -# -fLo "$DEST_FILE" "$DOWNLOAD_URL"
else
  wget -qO "$DEST_FILE" "$DOWNLOAD_URL" --show-progress
fi

chmod +x "$DEST_FILE"

if [[ "$OS_NAME" == "darwin" ]]; then
  unquarantine_mac "$DEST_FILE"
fi

setup_path

echo -e "\n${GREEN}✅ Successfully installed twvm $VERSION to $INSTALL_DIR!${NC}"
echo -e "Run ${CYAN}tailwind-sorter --help${NC} to get started."
