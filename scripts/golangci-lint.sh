#!/usr/bin/env bash                                                                          

# This script provide a hacky way to skip directories listed by go list ./...
# when running golangci-lint.
# https://github.com/golangci/golangci-lint/issues/301

set -e
set -x
# Only exit with zero if all commands of the pipeline exit successfully.
set -o pipefail

# Make sure to use project tooling
PATH="$(pwd)/tmp/bin:${PATH}"

# Directories we want to skip when running golangci-lint.
SKIP_DIRS="scripts"

go list -f '{{.Dir}}' ./...  | fgrep -v ${SKIP_DIRS} | xargs realpath --relative-to=. | xargs golangci-lint run -v
