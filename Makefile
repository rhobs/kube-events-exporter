GOARCH?=$(shell go env GOARCH)
GOOS?=$(shell uname -s | tr A-Z a-z)

VERSION?=$(shell cat VERSION)

.PHONY: all
all: build

.PHONY: build
build: kube-events-exporter

.PHONY: kube-events-exporter
kube-events-exporter:
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -ldflags "-s -w"

.PHONY: clean
clean:
	git clean -Xfd .
