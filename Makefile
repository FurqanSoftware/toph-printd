# Copyright 2023 Furqan Software Ltd. All rights reserved.

BUILD_TAG := $(shell git describe --tags)
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

PRINTD_EXT := 
ifeq ($(shell go env GOOS),windows)
	PRINTD_EXT := .exe
endif

.PHONY: printd
printd:
	go generate
	go build -ldflags "-X main.version=$(BUILD_TAG:v%=%) -X main.date=$(BUILD_TIME)" -o printd$(PRINTD_EXT) .

.PHONY: lint
lint:
	staticcheck .

.PHONY: lint.tools.install
lint.tools.install:
	go install honnef.co/go/tools/cmd/staticcheck@2023.1.2

.PHONY: test
test:
	go test -race -v .

.PHONY: goversioninfo
goversioninfo:
ifeq ($(shell go env GOOS),windows)
	${GOPATH}/bin/goversioninfo \
		-o="rsrc_$(shell go env GOOS)_$(shell go env GOARCH).syso" \
		-copyright="© 2015-$(shell date +'%Y') Furqan Software Ltd." \
		-file-version=v$(BUILD_TAG:v%=%) \
		-product-version=v$(BUILD_TAG:v%=%) \
		-ver-major=$(word 1,$(subst ., ,$(BUILD_TAG:v%=%))) \
		-ver-minor=$(word 2,$(subst ., ,$(BUILD_TAG:v%=%))) \
		-ver-patch=$(word 3,$(subst ., ,$(word 1,$(subst -, ,$(BUILD_TAG:v%=%))))) \
		-ver-build=0 \
		-product-ver-major=$(word 1,$(subst ., ,$(BUILD_TAG:v%=%))) \
		-product-ver-minor=$(word 2,$(subst ., ,$(BUILD_TAG:v%=%))) \
		-product-ver-patch=$(word 3,$(subst ., ,$(word 1,$(subst -, ,$(BUILD_TAG:v%=%))))) \
		-product-ver-build=0 \
		-arm=$(if $(filter arm64,$(shell go env GOARCH)),true,false)
endif

.PHONY: goversioninfo.install
goversioninfo.install:
	go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@v1.4.0
