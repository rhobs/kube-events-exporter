GOARCH?=$(shell go env GOARCH)
ifeq ($(GOARCH),arm)
	ARCH=armv7
else
	ARCH=$(GOARCH)
endif
GOOS?=$(shell uname -s | tr A-Z a-z)
REPO=github.com/rhobs/kube-events-exporter
VERSION?=$(shell cat VERSION)
TAG?=$(shell git rev-parse --short HEAD)
IMAGE_REPO?=quay.io/dgrisonnet/kube-events-exporter

BIN_DIR?=$(shell pwd)/tmp/bin
GOLANGCI_BIN=$(BIN_DIR)/golangci-lint
GOJSONTOYAML_BIN=$(BIN_DIR)/gojsontoyaml
JSONNET_BIN=$(BIN_DIR)/jsonnet
JSONNETFMT_BIN=$(BIN_DIR)/jsonnetfmt
JB_BIN=$(BIN_DIR)/jb
TOOLING=$(GOLANGCI_BIN) $(GOJSONTOYAML_BIN) $(JSONNET_BIN) $(JSONNETFMT_BIN) $(JB_BIN)

KUBECONFIG?=$(HOME)/.kube/config

GO_VENDORS=. scripts
JSONNET_VENDORS=jsonnet/kube-events-exporter scripts/generate
PKGS=$(shell go list ./... | grep -v /test/e2e)

.PHONY: all
all: generate lint build test

.PHONY: vendor
vendor: vendor-go

.PHONY: vendor-go
vendor-go:
	@for dir in $(GO_VENDORS); do \
		cd $$dir; \
		go mod tidy; \
		go mod vendor; \
		go mod verify; \
		cd -; \
	done

.PHONY: vendor-jsonnet
vendor-jsonnet: $(JB_BIN)
	@for dir in $(JSONNET_VENDORS); do \
		cd $$dir; \
		rm -rf vendor; \
		$(JB_BIN) install; \
		cd -; \
	done

.PHONY: generate
generate: manifests

.PHONY: manifests
manifests: vendor-jsonnet $(GOJSONTOYAML_BIN) $(JSONNET_BIN)
	VERSION=$(VERSION) IMAGE_REPO=$(IMAGE_REPO) ./scripts/generate/generate-manifests.sh

.PHONY: lint
lint: check-license shellcheck lint-go lint-jsonnet

.PHONY: check-license
check-license:
	./scripts/check_license.sh

.PHONY: lint-go
lint-go: $(GOLANGCI_BIN)
	$(GOLANGCI_BIN) run -v


.PHONY: lint-jsonnet
lint-jsonnet: JSONNET_FILES:=$(shell find . -type f \( -name "*.libsonnet" -o -name "*.jsonnet" \) -not -path "*/vendor/*")
lint-jsonnet: $(JSONNETFMT_BIN)
	$(JSONNETFMT_BIN) -i $(JSONNET_FILES)

.PHONY: shellcheck
shellcheck: SHELL_FILES:=$(shell find . -type f -name "*.sh" -not -path "*/vendor/*")
shellcheck:
	docker run -v "$(PWD):/mnt" koalaman/shellcheck:stable $(SHELL_FILES)

.PHONY: build
build: kube-events-exporter

.PHONY: kube-events-exporter
kube-events-exporter:
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -ldflags "-s -w -X $(REPO)/internal/version.version=$(VERSION)" ./cmd/$@

.PHONY: container
container: build
	docker build --build-arg ARCH=$(ARCH) --build-arg OS=$(GOOS) -t $(IMAGE_REPO):$(TAG) .

.PHONY: container-push
container-push: container
	docker push $(IMAGE_REPO):$(TAG)

.PHONY: test
test: test-unit test-e2e

.PHONY: test-unit
test-unit:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test -v -race -count=1 $(PKGS)

.PHONY: test-e2e
test-e2e:
	./scripts/setup-e2e.sh
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test -v -race -count=1 ./test/e2e/ --kubeconfig=$(KUBECONFIG) --exporter-image=$(IMAGE_REPO):$(TAG)

.PHONY: clean
clean:
	git clean -Xfd .

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(TOOLING): $(BIN_DIR)
	@echo Installing tools from scripts/tools.go
	@cd scripts && cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go build -o $(BIN_DIR) %
