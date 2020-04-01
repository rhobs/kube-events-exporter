GOARCH?=$(shell go env GOARCH)
GOOS?=$(shell uname -s | tr A-Z a-z)
GOLANGCI_VERSION?=v1.24.0
VERSION?=$(shell cat VERSION)

.PHONY: all
all: lint build

.PHONY: lint
lint:
	@docker run --rm -v $(shell pwd):/app:ro \
					 -w /app \
					 golangci/golangci-lint:$(GOLANGCI_VERSION) \
					 golangci-lint run -v

.PHONY: build
build: kube-events-exporter

.PHONY: kube-events-exporter
kube-events-exporter:
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -ldflags "-s -w"

.PHONY: clean
clean:
	git clean -Xfd .
