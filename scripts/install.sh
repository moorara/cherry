#!/usr/bin/env sh

#
# This script will install the latest version of Cherry
#
# USAGE:
#   ./install.sh
#

set -euo pipefail


get_latest_release() {
  os="$1"
  arch="$2"
  release_url="https://github.com/moorara/cherry/releases/latest"
  bin_pattern="/moorara/cherry/releases/download/v.*/cherry-$os-$arch"

  if hash curl 2>/dev/null; then
    bin_path=$(curl -sL $release_url | grep -o "$bin_pattern")
  elif hash wget 2>/dev/null; then
    bin_path=$(wget -qO- $release_url | grep -o "$bin_pattern")
  else
    printf "No command available to get %s\n" "$release_url"
    exit 1
  fi

  download_url="https://github.com$bin_path"
  latest_version=$(echo "$bin_path" | cut -d '/' -f6 | cut -d 'v' -f2)
}

install_cherry() {
  download_url="$1"
  bin_path="/usr/local/bin/cherry"

  if hash curl 2>/dev/null; then
    curl -fsSL -o "$bin_path" "$download_url"
  elif hash wget 2>/dev/null; then
    wget -qO "$bin_path" "$download_url"
  else
    printf "No command available to download %s\n" "$download_url"
    exit 1
  fi

  chmod 755 $bin_path
}

main() {
  printf "Installing the latest release of Cherry ...\n"

  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  arch=$(uname -m)

  if [ "$arch" = "i386" ]; then
    arch="386"
  elif [ "$arch" = "x86_64" ]; then
    arch="amd64"
  fi

  get_latest_release "$os" "$arch"
  install_cherry "$download_url"

  printf "Cherry %s installed successfully.\n" "$latest_version"
}


main
