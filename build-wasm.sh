#!/bin/bash

set -e
set -o pipefail

root="$(cd -P -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
cd "$root"

# Keep nested TinyGo build commands on the repository-pinned Go toolchain.
go_root="$("$root/ops/go.sh" env GOROOT)"
export PATH="$go_root/bin:$PATH"

mkdir -p .bin/ 2>/dev/null

# Setup tinygo first
if [[ ! -d .bin/tinygo ]]; then
  git clone --depth=1 --branch v0.38.0 https://github.com/tinygo-org/tinygo.git .bin/tinygo
  pushd .bin/tinygo
  git submodule update --init --recursive

  go mod download -x && go mod verify

  make llvm-source
  make llvm-build

  make binaryen STATIC=1

  make build/release
else
  pushd .bin/tinygo
fi

TINYGOROOT="$(realpath ./build/release/tinygo/)"
export TINYGOROOT
PATH="$PATH:$(realpath ./build/release/tinygo/bin/)"
export PATH

popd

go generate ./...
