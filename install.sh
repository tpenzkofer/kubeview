#!/usr/bin/env bash
#
# Install kubeview from GitHub Releases.
#
# Because the repo is private, this uses the GitHub CLI (`gh`) so it can
# authenticate. Install gh (https://cli.github.com) and run `gh auth login`
# first, or use `go install` (see README) if you have Go + git credentials.
#
# Usage:
#   ./install.sh                 # latest release
#   ./install.sh v0.1.0          # a specific tag
#   KUBEVIEW_DEST=~/bin ./install.sh
#
set -euo pipefail

REPO="${KUBEVIEW_REPO:-tpenzkofer/kubeview}"
DEST="${KUBEVIEW_DEST:-/usr/local/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64 | amd64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *) echo "unsupported architecture: $arch" >&2; exit 1 ;;
esac

if ! command -v gh >/dev/null 2>&1; then
  echo "error: this installer needs the GitHub CLI (gh) for the private repo." >&2
  echo "install gh and run 'gh auth login', or use: go install $REPO@latest" >&2
  exit 1
fi

tag="${1:-$(gh release view -R "$REPO" --json tagName -q .tagName)}"
echo "installing kubeview $tag ($os/$arch) from $REPO"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

gh release download "$tag" -R "$REPO" \
  --pattern "*_${os}_${arch}.tar.gz" --dir "$tmp"
tar -xzf "$tmp"/*.tar.gz -C "$tmp"

if install -m 0755 "$tmp/kubeview" "$DEST/kubeview" 2>/dev/null; then
  :
else
  echo "elevating with sudo to write $DEST"
  sudo install -m 0755 "$tmp/kubeview" "$DEST/kubeview"
fi

echo "installed: $("$DEST/kubeview" --version)"
