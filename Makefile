# Copyright 2023 Furqan Software Ltd. All rights reserved.

BUILD_TAG := $(shell git describe --tags)
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

.PHONY: printd
printd:
	go build -ldflags "-X main.buildTag=$(BUILD_TAG) -X main.buildTime=$(BUILD_TIME)" -o printd .
