#!/usr/bin/env bash
# exit immediately when a command fails
set -e
# only exit with zero if all commands of the pipeline exit successfully
set -o pipefail
# error on unset variables
set -u

KUBECTL_BIN=${KUBECTL_BIN:-kubectl}
KIND_BIN=${KIND_BIN:-kind}

DOCKER_REPO=${DOCKER_REPO:-quay.io/dgrisonnet/kube-events-exporter}
TAG=${TAG:-$(git rev-parse --short HEAD)}

context=$(${KUBECTL_BIN} config view -o json | grep "current-context" | awk -F '"' '{print $4}')

case ${context} in
minikube)
    eval "$(minikube -p minikube docker-env)";
    make container VERSION="${TAG}";;
kind-kind)
    make container VERSION="${TAG}";
    ${KIND_BIN} load docker-image "${DOCKER_REPO}:${TAG}";;
*)
    echo ERROR: cluster context "${context}" not supported, use minikube or kind instead.;
    exit 1;;
esac
