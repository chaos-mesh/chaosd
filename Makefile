.DEFAULT_GOAL := build

LDFLAGS = $(if $(IMG_LDFLAGS),$(IMG_LDFLAGS),$(if $(DEBUGGER),,-s -w) $(shell ./hack/version.sh))

# Enable GO111MODULE=on explicitly, disable it with GO111MODULE=off when necessary.
export GO111MODULE := on
GOOS ?=
GOARCH ?=
GOENV  := CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH)
CGOENV := CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH)
GO     := $(GOENV) go
CGO    := $(CGOENV) go
GOTEST := TEST_USE_EXISTING_CLUSTER=false NO_PROXY="${NO_PROXY},testhost" go test -count=1
SHELL    := /usr/bin/env bash
TOOLS_BIN ?= $(CURDIR)/bin/dev-tools/$(shell go env GOHOSTOS)-$(shell go env GOHOSTARCH)
GOLANGCI_LINT_VERSION := v2.15.0
GOLANGCI_LINT := $(TOOLS_BIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)

IMAGE_TAG := $(if $(IMAGE_TAG),$(IMAGE_TAG),latest)
IMAGE_NAME ?= chaosd

BUILD_TAGS ?=

ifeq ($(SWAGGER),1)
	BUILD_TAGS += swagger_server
endif

TARGET_GOOS := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
TARGET_GOARCH := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))

$(GOLANGCI_LINT):
	mkdir -p "$(TOOLS_BIN)"
	set -o pipefail; curl -sSfL "https://raw.githubusercontent.com/golangci/golangci-lint/$(GOLANGCI_LINT_VERSION)/install.sh" | sh -s -- -b "$(TOOLS_BIN)" "$(GOLANGCI_LINT_VERSION)"
	mv "$(TOOLS_BIN)/golangci-lint" "$@"

build: binary

binary: chaosd chaos-tools

image-binary:
	DOCKER_BUILDKIT=1 docker build --target binary --build-arg LDFLAGS="$(LDFLAGS)" --build-arg SWAGGER="$(SWAGGER)" -t chaosd-binary:local ${DOCKER_BUILD_ARGS} .

chaosd: swagger_spec
	$(CGOENV) go build -ldflags '$(LDFLAGS)' -tags "${BUILD_TAGS}" -o bin/chaosd ./cmd/main.go


chaos-tools:
	$(CGOENV) go build -ldflags '-s -w' -o bin/tools/PortOccupyTool ./tools
	$(CGOENV) go build -ldflags '-s -w' -o bin/tools/FileTool ./tools/file
	bash hack/install-tools.sh $(TARGET_GOOS) $(TARGET_GOARCH) bin/tools

swagger_spec:
ifeq ($(SWAGGER),1)
	hack/generate_swagger_spec.sh
endif

image-build-base:
	DOCKER_BUILDKIT=1 docker build --target builder --ulimit nofile=65536:65536 -t chaosd-build-base:local ${DOCKER_BUILD_ARGS} .

image-chaosd:
	DOCKER_BUILDKIT=1 docker build --target runtime --build-arg LDFLAGS="$(LDFLAGS)" --build-arg SWAGGER="$(SWAGGER)" -t ${DOCKER_REGISTRY_PREFIX}${IMAGE_NAME}:${IMAGE_TAG} ${DOCKER_BUILD_ARGS} .

check: fmt lint tidy

# gofmt fails on syntax errors; golangci-lint fmt only warns.
fmt: $(GOLANGCI_LINT)
	find . -type f -name '*.go' \
		-not -path './.git/*' \
		-not -path './.cache/*' \
		-not -path './bin/*' \
		-not -path './vendor/*' \
		-exec gofmt -l {} + \
		-exec "$(GOLANGCI_LINT)" fmt --config .golangci.yml -- {} +

lint: $(GOLANGCI_LINT)
	$(CGOENV) "$(GOLANGCI_LINT)" run --config .golangci.yml ./...

tidy:
	@echo "go mod tidy"
	go mod tidy -diff

unit-test:
	rm -rf cover.* cover
	$(GOTEST) ./... -coverprofile cover.out.tmp
	cat cover.out.tmp | grep -v "_generated.deepcopy.go" > cover.out

dummy:
	$(CGOENV) go build -ldflags '$(LDFLAGS)' -tags "${BUILD_TAGS}" -o bin/dummy ./test/utilities/dummy.go

integration-test: build dummy
	bash test/integration_test/run.sh

.PHONY: build check fmt lint tidy binary chaosd chaos-tools swagger_spec unit-test dummy integration-test image-binary image-build-base image-chaosd
