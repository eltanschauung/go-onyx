#!/usr/bin/env bash
set -euo pipefail

root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd -P)"
cd "$root"

go_cmd="$root/ops/go.sh"
gofmt="$($go_cmd env GOROOT)/bin/gofmt"
mapfile -d '' go_files < <(git ls-files -z -- '*.go')

unformatted="$($gofmt -l "${go_files[@]}")"
if [[ -n "$unformatted" ]]; then
  printf 'gofmt is required for:\n%s\n' "$unformatted" >&2
  exit 1
fi

git diff --check
"$go_cmd" vet ./...
"$go_cmd" test ./...
"$go_cmd" test -race ./lib/challenge

if command -v shellcheck >/dev/null 2>&1; then
  mapfile -d '' shell_files < <(git ls-files -z -- '*.sh')
  shellcheck "${shell_files[@]}"
else
  printf 'shellcheck not found; skipping shell-script lint\n' >&2
fi
