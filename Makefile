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
REVIVE_VERSION := v1.17.1
GOIMPORTS_VERSION := v0.52.0
REVIVE := $(TOOLS_BIN)/revive-$(REVIVE_VERSION)
GOIMPORTS := $(TOOLS_BIN)/goimports-$(GOIMPORTS_VERSION)

IMAGE_TAG := $(if $(IMAGE_TAG),$(IMAGE_TAG),latest)
IMAGE_NAME ?= chaosd

BUILD_TAGS ?=

ifeq ($(SWAGGER),1)
	BUILD_TAGS += swagger_server
endif

PACKAGE_LIST := go list ./... | grep -vE "chaos-daemon/test|pkg/ptrace|zz_generated|vendor"
PACKAGE_DIRECTORIES := $(PACKAGE_LIST) | sed 's|github.com/chaos-mesh/chaosd/||'

TARGET_GOOS := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
TARGET_GOARCH := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))

$(REVIVE):
	GOBIN=$(TOOLS_BIN) CGO_ENABLED=0 GOOS= GOARCH= go install github.com/mgechev/revive@$(REVIVE_VERSION)
	mv $(TOOLS_BIN)/revive $@

$(GOIMPORTS):
	GOBIN=$(TOOLS_BIN) CGO_ENABLED=0 GOOS= GOARCH= go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	mv $(TOOLS_BIN)/goimports $@

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

check: fmt-check vet boilerplate lint tidy

# Run go fmt against code
fmt: groupimports
	$(CGOENV) go fmt ./...

groupimports: $(GOIMPORTS)
	@set -o pipefail; packages=$$($(PACKAGE_DIRECTORIES)) || exit $$?; \
	$< -w -l -local github.com/chaos-mesh/chaosd $$packages

fmt-check: $(GOIMPORTS)
	@set -o pipefail; packages=$$($(PACKAGE_DIRECTORIES)) || exit $$?; \
	files=$$($< -l -local github.com/chaos-mesh/chaosd $$packages) || exit $$?; \
	if [ -n "$$files" ]; then echo "Run make fmt to format:"; echo "$$files"; exit 1; fi

# Run go vet against code
vet:
	$(CGOENV) go vet ./...

lint: $(REVIVE)
	@echo "linting"
	@set -o pipefail; packages=$$($(PACKAGE_LIST)) || exit $$?; \
	$< -formatter friendly -config revive.toml $$packages

boilerplate:
	./hack/verify-boilerplate.sh

tidy:
	@echo "go mod tidy"
	go mod tidy -diff

unit-test:
	rm -rf cover.* cover
	@set -o pipefail; packages=$$($(PACKAGE_LIST)) || exit $$?; \
	$(GOTEST) $$packages -coverprofile cover.out.tmp
	cat cover.out.tmp | grep -v "_generated.deepcopy.go" > cover.out

dummy:
	$(CGOENV) go build -ldflags '$(LDFLAGS)' -tags "${BUILD_TAGS}" -o bin/dummy ./test/utilities/dummy.go

integration-test: build dummy
	bash test/integration_test/run.sh

.PHONY: all chaos build check fmt fmt-check groupimports vet boilerplate lint tidy binary chaosd chaos-tools swagger_spec unit-test dummy integration-test image-binary image-build-base image-chaosd
