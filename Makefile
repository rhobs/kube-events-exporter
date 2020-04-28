GOARCH?=$(shell go env GOARCH)
GOOS?=$(shell uname -s | tr A-Z a-z)
REPO=github.com/rhobs/kube-events-exporter
VERSION?=$(shell cat VERSION)
DOCKER_REPO?=quay.io/dgrisonnet/kube-events-exporter

BIN_DIR?=$(shell pwd)/tmp/bin
GOLANGCI_BIN=$(BIN_DIR)/golangci-lint
TOOLING=$(GOLANGCI_BIN)

GOMOD_DIRS=. scripts
PKGS=$(shell go list ./... | grep -v /test/e2e)

.PHONY: all
all: lint build test

.PHONY: vendor
vendor:
	@for dir in $(GOMOD_DIRS); do \
		cd $$dir; \
		go mod tidy; \
		go mod vendor; \
		go mod verify; \
	done

.PHONY: lint
lint: check-license lint-go

.PHONY: check-license
check-license:
	./scripts/check_license.sh

.PHONY: lint-go
lint-go: $(GOLANGCI_BIN)
	$(GOLANGCI_BIN) run -v

.PHONY: build
build: kube-events-exporter

.PHONY: kube-events-exporter
kube-events-exporter:
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -ldflags "-s -w -X $(REPO)/internal/version.version=$(VERSION)" ./cmd/$@

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
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test -v -race $(PKGS)

.PHONY: test-e2e
test-e2e:
	@echo "FIXME: add e2e tests"

.PHONY: clean
clean:
	git clean -Xfd .

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(TOOLING): $(BIN_DIR)
	@echo Installing tools from scripts/tools.go
	@cd scripts && cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go build -o $(BIN_DIR) %
