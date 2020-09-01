#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

# Make sure to use project tooling
PATH="$(pwd)/tmp/bin:${PATH}"

# Make sure to start with a clean 'manifests' dir
rm -rf manifests
mkdir manifests

jsonnet_path="scripts/generate"
# Calling gojsontoyaml is optional, but we would like to generate yaml, not json
jsonnet -J "${jsonnet_path}/vendor" -m manifests "${jsonnet_path}/kube-events-exporter.jsonnet" \
    --ext-str VERSION --ext-str IMAGE_REPO \
    | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {}

# Make sure to remove json files
find manifests -type f ! -name '*.yaml' -delete
