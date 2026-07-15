#!/usr/bin/env bash
set -euo pipefail

root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd -P)"
required="$(tr -d '[:space:]' < "$root/.go-version")"

if [[ ! "$required" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  printf 'Invalid Go version in %s/.go-version: %q\n' "$root" "$required" >&2
  exit 2
fi

candidates=()
if [[ -n "${GO_ONYX_GO:-}" ]]; then
  candidates+=("$GO_ONYX_GO")
else
  candidates+=(
    "$HOME/.local/toolchains/go${required}/bin/go"
    "$HOME/.local/go${required}/bin/go"
  )
  if command -v go >/dev/null 2>&1; then
    candidates+=("$(command -v go)")
  fi
fi

for candidate in "${candidates[@]}"; do
  [[ -x "$candidate" ]] || continue
  actual="$("$candidate" version 2>/dev/null | sed -nE 's/^go version go([0-9]+\.[0-9]+\.[0-9]+).*/\1/p')"
  if [[ "$actual" == "$required" ]]; then
    export GOTOOLCHAIN=local
    exec "$candidate" "$@"
  fi
done

printf 'Go %s is required; no matching compiler was found.\n' "$required" >&2
printf 'Install it under ~/.local/toolchains/go%s or set GO_ONYX_GO to its go binary.\n' "$required" >&2
exit 1
