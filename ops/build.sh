#!/usr/bin/env bash
set -euo pipefail

root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd -P)"
cd "$root"

output="${GO_ONYX_OUTPUT:-$root/.bin/go-away}"
mkdir -p "$(dirname -- "$output")"
temporary="${output}.tmp.$$"
trap 'rm -f -- "$temporary"' EXIT

CGO_ENABLED=0 "$root/ops/go.sh" build \
  -pgo=auto \
  -trimpath \
  -ldflags='-buildid= -bindnow' \
  -buildmode=pie \
  -o "$temporary" \
  ./cmd/go-away

chmod 0755 "$temporary"
mv -f -- "$temporary" "$output"
trap - EXIT

"$root/ops/go.sh" version -m "$output"
