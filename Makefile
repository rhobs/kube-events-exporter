GOARCH?=$(shell go env GOARCH)
GOOS?=$(shell uname -s | tr A-Z a-z)
GOLANGCI_VERSION?=v1.24.0
VERSION?=$(shell cat VERSION)
DOCKER_REPO?=quay.io/dgrisonnet/kube-events-exporter

.PHONY: all
all: lint build test

.PHONY: lint
lint: check-license
	@docker run --rm -v $(shell pwd):/app:ro \
					 -w /app \
					 golangci/golangci-lint:$(GOLANGCI_VERSION) \
					 golangci-lint run -v

.PHONY: check-license
check-license:
	./scripts/check_license.sh

.PHONY: build
build: kube-events-exporter

.PHONY: kube-events-exporter
kube-events-exporter:
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -ldflags "-s -w"

.PHONY: container
container: build
	docker build -t $(DOCKER_REPO):$(VERSION) .

.PHONY: container-push
container-push: container
	docker push $(DOCKER_REPO):$(VERSION)

.PHONY: test
test: test-unit test-e2e

.PHONY: test-unit
test-unit:
	@echo "FIXME: add unit tests"

.PHONY: test-e2e
test-e2e:
	@echo "FIXME: add e2e tests"

.PHONY: clean
clean:
	git clean -Xfd .
