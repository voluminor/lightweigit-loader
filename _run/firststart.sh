#!/usr/bin/env bash

set -Eeuo pipefail

run_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root_path="$(cd "$run_dir/.." && pwd)"
gometagen="github.com/amazing-generators/gometagen/cmd/gometagen@latest"

cd "$root_path"

mkdir -p target
mkdir -p tmp

go install "$gometagen"

go run "$gometagen" git add-commit-hook -source "$root_path"
go run "$gometagen" git add-push-hook -source "$root_path"

go generate .
go mod tidy
go generate .
