#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

export DOCKER_REPO=${DOCKER_REPO:-"quay.io/dgrisonnet/kube-events-exporter"}

CPU_ARCHS=${CPU_ARCHS:-"amd64 arm64 arm"}
TAG=${TAG:-$(git rev-parse --short HEAD)}

if [ "${TRAVIS-}" == "true" ]; then
        # Workaround for docker bug https://github.com/docker/for-linux/issues/396
        sudo chmod o+x /etc/docker
fi

# Images need to be on remote registry before creating manifests
for arch in $CPU_ARCHS; do
        make --always-make container-push GOARCH="$arch" TAG="${TAG}-$arch"
done

export DOCKER_CLI_EXPERIMENTAL=enabled
# Create manifest to join all images under one virtual tag
docker manifest create -a "${DOCKER_REPO}:${TAG}" \
                          "${DOCKER_REPO}:${TAG}-amd64" \
                          "${DOCKER_REPO}:${TAG}-arm64" \
                          "${DOCKER_REPO}:${TAG}-arm"

# Annotate to set which image is build for which CPU architecture
for arch in $CPU_ARCHS; do
        docker manifest annotate --arch "$arch" "${DOCKER_REPO}:${TAG}" "${DOCKER_REPO}:${TAG}-$arch"
done

docker manifest push "${DOCKER_REPO}:${TAG}"
