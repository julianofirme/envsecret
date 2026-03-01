#!/usr/bin/env bash
set -euo pipefail

BINARY="envs"
REPO="julianofirme/envsecret"
REPO_URL="https://github.com/$REPO"
INSTALL_DIR="/usr/local/bin"

# ── colors ────────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
RESET='\033[0m'

info()    { echo -e "${BOLD}$*${RESET}"; }
success() { echo -e "${GREEN}✓${RESET} $*"; }
warn()    { echo -e "${YELLOW}! $*${RESET}"; }
die()     { echo -e "${RED}error:${RESET} $*" >&2; exit 1; }

# ── detect platform ───────────────────────────────────────────────────────────
OS=$(uname -s)
ARCH=$(uname -m)

case "$OS" in
  Darwin) GOOS="darwin" ;;
  Linux)  GOOS="linux"  ;;
  *) die "unsupported OS: $OS (only macOS and Linux are supported)" ;;
esac

case "$ARCH" in
  x86_64)          GOARCH="amd64" ;;
  arm64 | aarch64) GOARCH="arm64" ;;
  *) die "unsupported architecture: $ARCH" ;;
esac

ASSET="envs-${GOOS}-${GOARCH}"
info "Platform: ${GOOS}/${GOARCH}"

# ── Linux: check libsecret ────────────────────────────────────────────────────
if [[ "$GOOS" == "linux" ]]; then
  if ! ldconfig -p 2>/dev/null | grep -q libsecret || ! pkg-config --exists libsecret-1 2>/dev/null; then
    warn "libsecret not found. Install it first:"
    warn "  Ubuntu/Debian : sudo apt install libsecret-1-dev"
    warn "  Fedora        : sudo dnf install libsecret-devel"
    warn "  Arch          : sudo pacman -S libsecret"
    die "missing libsecret"
  fi
  success "libsecret found"
fi

# ── resolve version ───────────────────────────────────────────────────────────
VERSION="${ENVS_VERSION:-}"

if [[ -z "$VERSION" ]]; then
  info "Fetching latest release..."
  if command -v curl >/dev/null 2>&1; then
    VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
      | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
  elif command -v wget >/dev/null 2>&1; then
    VERSION=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" \
      | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')
  else
    die "curl or wget is required"
  fi
fi

if [[ -z "$VERSION" ]]; then
  warn "Could not determine latest release version."
  warn "To install from source instead, run:"
  warn "  ENVS_FROM_SOURCE=1 ./install.sh"
  die "no release version found — is the GitHub repo public and does it have releases?"
fi

success "Version: $VERSION"

# ── download ──────────────────────────────────────────────────────────────────
DOWNLOAD_URL="${REPO_URL}/releases/download/${VERSION}/${ASSET}"
TMP_DIR=$(mktemp -d)
TMP_BIN="${TMP_DIR}/${BINARY}"

info "Downloading $ASSET from $DOWNLOAD_URL..."

if command -v curl >/dev/null 2>&1; then
  curl -fsSL --progress-bar "$DOWNLOAD_URL" -o "$TMP_BIN" \
    || die "download failed — check that $DOWNLOAD_URL exists"
else
  wget -q --show-progress "$DOWNLOAD_URL" -O "$TMP_BIN" \
    || die "download failed — check that $DOWNLOAD_URL exists"
fi

# ── verify checksum (optional but recommended) ────────────────────────────────
CHECKSUM_URL="${REPO_URL}/releases/download/${VERSION}/checksums.txt"
TMP_CHECKSUMS="${TMP_DIR}/checksums.txt"

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$CHECKSUM_URL" -o "$TMP_CHECKSUMS" 2>/dev/null || true
else
  wget -qO "$TMP_CHECKSUMS" "$CHECKSUM_URL" 2>/dev/null || true
fi

if [[ -s "$TMP_CHECKSUMS" ]]; then
  if command -v sha256sum >/dev/null 2>&1; then
    EXPECTED=$(grep "$ASSET" "$TMP_CHECKSUMS" | awk '{print $1}')
    ACTUAL=$(sha256sum "$TMP_BIN" | awk '{print $1}')
    if [[ "$EXPECTED" == "$ACTUAL" ]]; then
      success "Checksum verified"
    else
      rm -rf "$TMP_DIR"
      die "checksum mismatch — download may be corrupt"
    fi
  elif command -v shasum >/dev/null 2>&1; then
    EXPECTED=$(grep "$ASSET" "$TMP_CHECKSUMS" | awk '{print $1}')
    ACTUAL=$(shasum -a 256 "$TMP_BIN" | awk '{print $1}')
    if [[ "$EXPECTED" == "$ACTUAL" ]]; then
      success "Checksum verified"
    else
      rm -rf "$TMP_DIR"
      die "checksum mismatch — download may be corrupt"
    fi
  else
    warn "sha256sum/shasum not found — skipping checksum verification"
  fi
else
  warn "checksums.txt not available — skipping verification"
fi

# ── install ───────────────────────────────────────────────────────────────────
chmod +x "$TMP_BIN"
info "Installing to $INSTALL_DIR/$BINARY..."

if [[ -w "$INSTALL_DIR" ]]; then
  mv "$TMP_BIN" "$INSTALL_DIR/$BINARY"
else
  sudo mv "$TMP_BIN" "$INSTALL_DIR/$BINARY"
fi

rm -rf "$TMP_DIR"
success "Installed $INSTALL_DIR/$BINARY ($VERSION)"

# ── verify ────────────────────────────────────────────────────────────────────
if command -v "$BINARY" >/dev/null 2>&1; then
  success "$BINARY is available in PATH"
else
  warn "$INSTALL_DIR is not in your PATH."
  warn "Add this to your shell profile and restart your terminal:"
  warn "  export PATH=\"\$PATH:$INSTALL_DIR\""
fi

echo
echo -e "${BOLD}Done. Quick start:${RESET}"
echo
echo "  cd ~/projects/my-app"
echo "  envs init"
echo "  envs set API_KEY \"sk-...\""
echo "  envs run -- node server.js"
echo
echo "  Full docs: $REPO_URL"
