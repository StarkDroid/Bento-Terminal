#!/bin/sh
# Install the bento binary. This is the macOS and Linux entrypoint; Windows
# installs with scripts/install.ps1.
#
#   curl -fsSL https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.sh | sh
#   curl -fsSL https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.sh | sh -s -- --with-zsh
#   curl -fsSL https://raw.githubusercontent.com/StarkDroid/Bento-Terminal/main/scripts/install.sh | sh -s -- --no-zsh
#   sh scripts/install.sh [--with-zsh|--no-zsh] [--from-source]
#
# Default and --no-zsh install the binary only. --with-zsh appends PATH and a
# bento function to ~/.zshrc and does not change the theme or prompt.
# --from-source builds the checkout this script lives in, or a clone, even when
# a release asset exists; that is how a branch is installed before a release.
#
# A published release asset is the install path that does not need a Go
# toolchain. Asset names are bento_<os>_<arch>, for example
# bento_darwin_arm64 and bento_darwin_amd64. When that asset is missing, the
# script builds from source and requires git and go on PATH.

set -eu

REPO="${BENTO_REPO:-StarkDroid/Bento-Terminal}"
REPO_URL="${BENTO_REPO_URL:-https://github.com/${REPO}.git}"
INSTALL_DIR="${BENTO_INSTALL_DIR:-${HOME:-}/.local/bin}"

WITH_ZSH=0
NO_ZSH=0
FROM_SOURCE=0

for arg in "$@"; do
  case "$arg" in
    --with-zsh) WITH_ZSH=1 ;;
    --no-zsh) NO_ZSH=1 ;;
    --from-source) FROM_SOURCE=1 ;;
    -h|--help)
      echo "Usage: install.sh [--with-zsh|--no-zsh] [--from-source]"
      echo "  (default)      install the binary only"
      echo "  --no-zsh       install the binary only"
      echo "  --with-zsh     also append PATH and a bento function to ~/.zshrc"
      echo "  --from-source  build the local checkout instead of the release binary"
      exit 0
      ;;
    *)
      echo "Unknown option: $arg" >&2
      exit 1
      ;;
  esac
done

if [ "$WITH_ZSH" = 1 ] && [ "$NO_ZSH" = 1 ]; then
  echo "Choose only one of --with-zsh or --no-zsh" >&2
  exit 1
fi

if [ -z "${HOME:-}" ]; then
  echo "HOME is not set" >&2
  exit 1
fi

os=$(uname -s)
arch=$(uname -m)
case "$os" in
  Darwin) goos=darwin ;;
  Linux) goos=linux ;;
  *)
    echo "Unsupported operating system: $os" >&2
    echo "Windows installs with scripts/install.ps1 in PowerShell." >&2
    exit 1
    ;;
esac
case "$arch" in
  arm64|aarch64) goarch=arm64 ;;
  x86_64|amd64) goarch=amd64 ;;
  *)
    echo "Unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

if [ "$goos" = "linux" ]; then
  echo "Linux uses the same binary as macOS; the supported terminals are macOS Terminal and Windows Terminal."
fi

asset="bento_${goos}_${goarch}"
dest="${INSTALL_DIR}/bento"
mkdir -p "$INSTALL_DIR"

is_binary() {
  sig=$(od -An -t x1 -N 4 "$1" 2>/dev/null | tr -d ' \n')
  case "$sig" in
    7f454c46|cffaedfe|feedfacf|cefaedfe|feedface|cafebabe) return 0 ;;
    *) return 1 ;;
  esac
}

download_release() {
  url="https://github.com/${REPO}/releases/latest/download/${asset}"
  tmp="${dest}.partial"
  echo "Looking for a release binary: ${asset}"
  if curl -fsSL --connect-timeout 8 --max-time 60 "$url" -o "$tmp"; then
    if is_binary "$tmp"; then
      mv "$tmp" "$dest"
      chmod +x "$dest"
      return 0
    fi
    rm -f "$tmp"
  fi
  rm -f "$tmp"
  return 1
}

build_tree() {
  tree=$1
  if ! command -v go >/dev/null 2>&1; then
    echo "No release binary is available for ${goos}/${goarch}." >&2
    echo "Building from source requires Go on PATH." >&2
    echo "The install path without a toolchain is a published release binary (${asset})." >&2
    exit 1
  fi
  echo "Building bento from ${tree}"
  (cd "$tree" && go build -o "$dest" ./cmd/bento)
  chmod +x "$dest"
}

checkout_root() {
  case "$0" in
    /*|*/*)
      if [ -f "$0" ]; then
        base=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
        if [ -f "${base}/go.mod" ] && [ -d "${base}/cmd/bento" ]; then
          echo "$base"
          return 0
        fi
      fi
      ;;
  esac
  return 1
}

clone_and_build() {
  if ! command -v git >/dev/null 2>&1; then
    echo "git is required to clone ${REPO_URL} and build from source" >&2
    exit 1
  fi
  tmp=$(mktemp -d)
  echo "Cloning ${REPO_URL}"
  git clone --depth 1 "$REPO_URL" "$tmp/src"
  build_tree "$tmp/src"
  rm -rf "$tmp"
}

if [ "$FROM_SOURCE" = 1 ]; then
  echo "Building from source (--from-source)."
  src=""
  if src=$(checkout_root); then
    build_tree "$src"
  else
    clone_and_build
  fi
elif download_release; then
  echo "Installed the published release binary."
else
  echo "No release binary named ${asset} is published."
  echo "A published release binary is the install path that does not need Go."
  src=""
  if src=$(checkout_root); then
    build_tree "$src"
  else
    clone_and_build
  fi
fi

if [ ! -x "$dest" ]; then
  echo "Install failed: ${dest} is not executable" >&2
  exit 1
fi

echo "Installed bento to ${dest}"

if [ "$WITH_ZSH" = 1 ]; then
  zshrc="${HOME}/.zshrc"
  marker="# >>> bento >>>"
  touch "$zshrc"
  if grep -q "$marker" "$zshrc" 2>/dev/null; then
    echo "zsh snippet already present in ${zshrc}"
  else
    cat >> "$zshrc" <<EOF

${marker}
export PATH="${INSTALL_DIR}:\$PATH"
bento() {
  command bento "\$@"
}
# <<< bento <<<
EOF
    echo "Appended PATH and a bento function to ${zshrc}"
    echo "The snippet does not change your theme or prompt."
    echo "Open a new terminal, or run: . \"${zshrc}\""
  fi
else
  echo "Binary only. Shell startup files were not changed."
  echo "Run: ${dest}"
  echo "Or add ${INSTALL_DIR} to PATH."
fi
