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

BUILD_TAGS ?=

ifeq ($(SWAGGER),1)
	BUILD_TAGS += swagger_server
endif

PACKAGE_LIST := go list ./... | grep -vE "chaos-daemon/test|pkg/ptrace|zz_generated|vendor"
PACKAGE_DIRECTORIES := $(PACKAGE_LIST) | sed 's|github.com/chaos-mesh/chaosd/||'

TARGET_GOOS := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
TARGET_GOARCH := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
CHAOSD_TOOLS_CACHE ?= $(if $(XDG_CACHE_HOME),$(XDG_CACHE_HOME),$(HOME)/.cache)/chaosd

$(REVIVE):
	GOBIN=$(TOOLS_BIN) CGO_ENABLED=0 GOOS= GOARCH= go install github.com/mgechev/revive@$(REVIVE_VERSION)
	mv $(TOOLS_BIN)/revive $@

$(GOIMPORTS):
	GOBIN=$(TOOLS_BIN) CGO_ENABLED=0 GOOS= GOARCH= go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	mv $(TOOLS_BIN)/goimports $@

build: binary

binary: chaosd chaos-tools

taily-build: image-build-base
	@builder_identity=$$(docker image inspect --format '{{json .RootFS.Layers}} {{json .Config}}' pingcap/chaos-build-base) || exit $$?; \
	ids=$$(docker ps -aq --filter 'name=^/taily-build$$') || exit $$?; \
	if [ -n "$$ids" ]; then \
		test "$$(docker inspect --format '{{.State.Running}}' taily-build)" = true && \
		test "$$(docker inspect --format '{{index .Config.Labels "io.chaos-mesh.chaosd.builder"}}' taily-build)" = "$$builder_identity" && \
		test "$$(docker inspect --format '{{range .Mounts}}{{if eq .Destination "/src"}}{{.Source}}{{end}}{{end}}' taily-build)" = "$(CURDIR)" || \
		{ echo "Existing taily-build uses another checkout or builder; inspect it before running make taily-build-clean" >&2; exit 1; }; \
	else \
		mkdir -p "$(shell go env GOMODCACHE)" "$(shell go env GOCACHE)" "$(CHAOSD_TOOLS_CACHE)"; \
		docker run --rm --mount "type=bind,source=$(CURDIR),target=/src" \
			--mount "type=bind,source=$(shell go env GOMODCACHE),target=/go/pkg/mod" \
			--mount "type=bind,source=$(shell go env GOCACHE),target=/root/.cache/go-build" \
			--mount "type=bind,source=$(CHAOSD_TOOLS_CACHE),target=/root/.cache/chaosd" \
			--workdir /src --name taily-build --label "io.chaos-mesh.chaosd.builder=$$builder_identity" -d pingcap/chaos-build-base tail -f /dev/null; \
	fi

taily-build-clean:
	@ids=$$(docker ps -aq --filter 'name=^/taily-build$$') || exit $$?; \
	if [ -n "$$ids" ]; then docker rm -f taily-build; fi

ifneq ($(TAILY_BUILD),)
image-binary: taily-build
	docker exec -e SWAGGER="$(SWAGGER)" -e IMG_LDFLAGS="$(LDFLAGS)" taily-build make binary
	printf 'FROM scratch\nCOPY chaosd /bin/chaosd\nCOPY tools /bin/tools\n' | docker build -t pingcap/chaos-binary -f - ./bin
else
image-binary: image-build-base
	DOCKER_BUILDKIT=1 docker build --build-arg LDFLAGS="$(LDFLAGS)" --build-arg SWAGGER="$(SWAGGER)" -t pingcap/chaos-binary ${DOCKER_BUILD_ARGS} .
endif

chaosd: swagger_spec
	$(CGOENV) go build -ldflags '$(LDFLAGS)' -tags "${BUILD_TAGS}" -o bin/chaosd ./cmd/main.go


chaos-tools:
	$(CGOENV) go build -o bin/tools/PortOccupyTool ./tools
	$(CGOENV) go build -o bin/tools/FileTool ./tools/file
	bash hack/install-tools.sh $(TARGET_GOOS) $(TARGET_GOARCH) bin/tools

swagger_spec:
ifeq ($(SWAGGER),1)
	hack/generate_swagger_spec.sh
endif

image-build-base:
	DOCKER_BUILDKIT=1 docker build --ulimit nofile=65536:65536 -t pingcap/chaos-build-base ${DOCKER_BUILD_ARGS} images/build-base

image-chaosd: image-binary
	docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/chaosd:${IMAGE_TAG} ${DOCKER_BUILD_ARGS} images/chaosd

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

.PHONY: all chaos taily-build taily-build-clean build check fmt fmt-check groupimports vet boilerplate lint tidy binary chaosd chaos-tools swagger_spec unit-test dummy integration-test image-binary image-build-base image-chaosd
