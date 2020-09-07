#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

# Make sure to use project tooling
PATH="$(pwd)/tmp/bin:${PATH}"

vendor="scripts/generate/vendor"

find examples -name '*.jsonnet' | while read -r jsonnet_path; do
    manifests_dir="${jsonnet_path%.*}"

    # Make sure to start with a clean 'manifests' dir
    rm -rf "$manifests_dir"
    mkdir "$manifests_dir"

    # Calling gojsontoyaml is optional, but we would like to generate yaml, not json
    jsonnet -J "$vendor" -m "$manifests_dir" "$jsonnet_path" \
        --ext-str VERSION --ext-str IMAGE_REPO \
        | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {}

    # Make sure to remove json files
    find "$manifests_dir" -type f ! -name '*.yaml' -delete
done

