# Copyright 2023 Furqan Software Ltd. All rights reserved.

BUILD_TAG := $(shell git describe --tags)
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

.PHONY: printd
printd:
	go build -ldflags "-X main.version=$(BUILD_TAG:v%=%) -X main.date=$(BUILD_TIME)" -o printd .

.PHONY: lint
lint:
	staticcheck .

.PHONY: lint.tools.install
lint.tools.install:
	go install honnef.co/go/tools/cmd/staticcheck@2023.1.2

.PHONY: test
test:
	go test -v .
