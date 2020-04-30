#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

KUBECTL_BIN=${KUBECTL_BIN:-kubectl}

# Inject local image in the cluster
./scripts/inject-image.sh

# Expose Kubernetes master
pkill -f "kubectl proxy" || true
${KUBECTL_BIN} proxy &
