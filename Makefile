LDFLAGS = $(if $(IMG_LDFLAGS),$(IMG_LDFLAGS),$(if $(DEBUGGER),,-s -w) $(shell ./hack/version.sh))

# Enable GO111MODULE=on explicitly, disable it with GO111MODULE=off when necessary.
export GO111MODULE := on
GOOS := $(if $(GOOS),$(GOOS),"")
GOARCH := $(if $(GOARCH),$(GOARCH),"")
GOENV  := GO15VENDOREXPERIMENT="1" CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH)
CGOENV := GO15VENDOREXPERIMENT="1" CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH)
GO     := $(GOENV) go
CGO    := $(CGOENV) go
GOTEST := TEST_USE_EXISTING_CLUSTER=false NO_PROXY="${NO_PROXY},testhost" go test
SHELL    := /usr/bin/env bash
BYTEMAN_DIR := byteman-chaos-mesh-download-v4.0.18-0.9

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

IMAGE_TAG := $(if $(IMAGE_TAG),$(IMAGE_TAG),latest)

BUILD_TAGS ?=

ifeq ($(SWAGGER),1)
	BUILD_TAGS += swagger_server
endif

PACKAGE_LIST := go list ./... | grep -vE "chaos-daemon/test|pkg/ptrace|zz_generated|vendor"
PACKAGE_DIRECTORIES := $(PACKAGE_LIST) | sed 's|github.com/chaos-mesh/chaosd/||'

$(GOBIN)/revive:
	$(GO) install github.com/mgechev/revive@v1.0.2-0.20200225072153-6219ca02fffb

$(GOBIN)/goimports:
	$(GO) install golang.org/x/tools/cmd/goimports@v0.1.1

build: binary

binary: swagger_spec chaosd chaos-tools

taily-build:
	if [ "$(shell docker ps --filter=name=$@ -q)" = "" ]; then \
		docker build -t pingcap/chaos-binary ${DOCKER_BUILD_ARGS} .; \
		docker run --rm --mount type=bind,source=$(shell pwd),target=/src \
			--name $@ -d pingcap/chaos-binary tail -f /dev/null; \
	fi;

taily-build-clean:
	docker kill taily-build && docker rm taily-build || exit 0

ifneq ($(TAILY_BUILD),)
image-binary: taily-build image-build-base
	docker exec -it taily-build make binary
	echo -e "FROM scratch\n COPY . /src/bin\n" | docker build -t pingcap/chaos-binary -f - ./bin
else
image-binary: image-build-base
	DOCKER_BUILDKIT=1 docker build -t pingcap/chaos-binary ${DOCKER_BUILD_ARGS} .
endif

chaosd:
	$(CGOENV) go build -ldflags '$(LDFLAGS)' -tags "${BUILD_TAGS}" -o bin/chaosd ./cmd/main.go


chaos-tools:
	$(CGOENV) go build -o bin/tools/PortOccupyTool tools/PortOccupyTool.go
	$(CGOENV) go build -o bin/tools/FileTool tools/file/*.go
ifeq (,$(wildcard bin/tools/stress-ng))
	curl -fsSL -o ./bin/tools/stress-ng https://mirrors.chaos-mesh.org/latest/stress-ng
	chmod +x ./bin/tools/stress-ng
endif
ifeq (,$(wildcard bin/tools/byteman))
	curl -fsSL -o ${BYTEMAN_DIR}.tar.gz https://mirrors.chaos-mesh.org/${BYTEMAN_DIR}.tar.gz
	tar zxvf ${BYTEMAN_DIR}.tar.gz
	mv ${BYTEMAN_DIR} ./bin/tools/byteman
endif

swagger_spec:
ifeq ($(SWAGGER),1)
	hack/generate_swagger_spec.sh
endif

image-build-base:
	DOCKER_BUILDKIT=0 docker build --ulimit nofile=65536:65536 -t pingcap/chaos-build-base ${DOCKER_BUILD_ARGS} images/build-base

image-chaosd: image-binary
	docker build -t ${DOCKER_REGISTRY_PREFIX}pingcap/chaosd:${IMAGE_TAG} ${DOCKER_BUILD_ARGS} images/chaosd

check: fmt vet boilerplate lint tidy

# Run go fmt against code
fmt: groupimports
	$(CGOENV) go fmt ./...

groupimports: $(GOBIN)/goimports
	$< -w -l -local github.com/chaos-mesh/chaosd $$($(PACKAGE_DIRECTORIES))

# Run go vet against code
vet:
	$(CGOENV) go vet ./...

lint: $(GOBIN)/revive
	@echo "linting"
	$< -formatter friendly -config revive.toml $$($(PACKAGE_LIST))

boilerplate:
	./hack/verify-boilerplate.sh

tidy:
	@echo "go mod tidy"
	GO111MODULE=on go mod tidy
	git diff -U --exit-code go.mod go.sum

unit-test:
	rm -rf cover.* cover
	$(GOTEST) $$($(PACKAGE_LIST)) -coverprofile cover.out.tmp
	cat cover.out.tmp | grep -v "_generated.deepcopy.go" > cover.out

dummy:
	$(CGOENV) go build -ldflags '$(LDFLAGS)' -tags "${BUILD_TAGS}" -o bin/dummy ./test/utilities/dummy.go

integration-test: build dummy
	bash test/integration_test/run.sh

.PHONY: all build check fmt vet lint tidy binary chaosd chaos image-binary image-chaosd
