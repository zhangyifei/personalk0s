include embedded-bins/Makefile.variables

GO_SRCS := $(shell find . -type f -name '*.go' -not -path './build/cache/*' -not -name 'zz_generated*')
GO_DIRS := . ./cmd/... ./pkg/... ./internal/... ./hack/...
BUILD_DIR := build

# EMBEDDED_BINS_BUILDMODE can be either:
#   docker	builds the binaries in docker
#   none	does not embed any binaries

EMBEDDED_BINS_BUILDMODE ?= docker

# eke runs on linux even if its built on mac or windows
TARGET_OS ?= linux
GOARCH ?= $(shell go env GOARCH)
GOPATH ?= $(shell go env GOPATH)
BUILD_UID ?= $(shell id -u)
BUILD_GID ?= $(shell id -g)
BUILD_GO_FLAGS := -tags osusergo
BUILD_GO_CGO_ENABLED ?= 0
BUILD_GO_LDFLAGS_EXTRA :=
DEBUG ?= false

VERSION ?= $(shell git describe --tags)
ifeq ($(DEBUG), false)
LD_FLAGS ?= -w -s
endif

KUBECTL_VERSION = $(shell go mod graph |  grep "eke" |  grep kubectl  | cut -d "@" -f 2 | sed "s/v0\./1./")
KUBECTL_MAJOR= $(shell echo ${KUBECTL_VERSION} | cut -d "." -f 1)
KUBECTL_MINOR= $(shell echo ${KUBECTL_VERSION} | cut -d "." -f 2)

# https://reproducible-builds.org/docs/source-date-epoch/#makefile
# https://reproducible-builds.org/docs/source-date-epoch/#git
# https://stackoverflow.com/a/15103333
BUILD_DATE_FMT = %Y-%m-%dT%H:%M:%SZ
ifdef SOURCE_DATE_EPOCH
	BUILD_DATE ?= $(shell date -u -d "@$(SOURCE_DATE_EPOCH)" "+$(BUILD_DATE_FMT)" 2>/dev/null || date -u -r "$(SOURCE_DATE_EPOCH)" "+$(BUILD_DATE_FMT)" 2>/dev/null || date -u "+$(BUILD_DATE_FMT)")
else
	BUILD_DATE ?= $(shell TZ=UTC git log -1 --pretty=%cd --date='format-local:$(BUILD_DATE_FMT)' || date -u +$(BUILD_DATE_FMT))
endif

LD_FLAGS += -X eke/pkg/build.Version=$(VERSION)
LD_FLAGS += -X eke/pkg/build.RuncVersion=$(runc_version)
LD_FLAGS += -X eke/pkg/build.ContainerdVersion=$(containerd_version)
LD_FLAGS += -X eke/pkg/build.KubernetesVersion=$(kubernetes_version)
LD_FLAGS += -X eke/pkg/build.KineVersion=$(kine_version)
LD_FLAGS += -X eke/pkg/build.EtcdVersion=$(etcd_version)
LD_FLAGS += -X eke/pkg/build.KonnectivityVersion=$(konnectivity_version)
LD_FLAGS += -X eke/pkg/build.EulaNotice=$(EULA_NOTICE)
LD_FLAGS += -X eke/pkg/telemetry.segmentToken=$(SEGMENT_TOKEN)
LD_FLAGS += -X k8s.io/component-base/version.gitVersion=v$(KUBECTL_VERSION)
LD_FLAGS += -X k8s.io/component-base/version.gitMajor=$(KUBECTL_MAJOR)
LD_FLAGS += -X k8s.io/component-base/version.gitMinor=$(KUBECTL_MINOR)
LD_FLAGS += -X k8s.io/component-base/version.buildDate=$(BUILD_DATE)
LD_FLAGS += -X k8s.io/component-base/version.gitCommit="not_available"
LD_FLAGS += $(BUILD_GO_LDFLAGS_EXTRA)

golint := $(shell which golangci-lint 2>/dev/null)
ifeq ($(golint),)
golint := cd hack/ci-deps && go install github.com/golangci/golangci-lint/cmd/golangci-lint && cd ../.. && "${GOPATH}/bin/golangci-lint"
endif

go_clientgen := $(shell which client-gen 2>/dev/null)
ifeq ($(go_clientgen),)
go_clientgen := cd hack/ci-deps && go install k8s.io/code-generator/cmd/client-gen@v0.22.2 && cd ../.. && "${GOPATH}/bin/client-gen"
endif

go_controllergen := $(shell which controller-gen 2>/dev/null)
ifeq ($(go_controllergen),)
go_controllergen := cd hack/ci-deps && go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0 && cd ../.. && "${GOPATH}/bin/controller-gen"
endif

GOLANG_IMAGE = golang:$(go_version)-alpine
GO ?= GOCACHE=/go/src/eke/build/cache/go/build GOMODCACHE=/go/src/eke/build/cache/go/mod docker run --rm \
	-v "$(CURDIR)":/go/src/eke \
	-w /go/src/eke \
	-e GOOS \
	-e CGO_ENABLED \
	-e GOARCH \
	-e GOCACHE \
	-e GOMODCACHE \
	--user $(BUILD_UID):$(BUILD_GID) \
	$(GOLANG_IMAGE) go

.PHONY: build
ifeq ($(TARGET_OS),windows)
build: eke.exe
else
build: eke
endif

.PHONY: all
all: eke eke.exe

go.sum: go.mod
	$(GO) mod tidy

eke: TARGET_OS = linux
eke: BUILD_GO_CGO_ENABLED = 0
eke: GOLANG_IMAGE = golang:1.17-alpine
eke: BUILD_GO_LDFLAGS_EXTRA = -extldflags=-static

eke.exe: TARGET_OS = windows
eke.exe: BUILD_GO_CGO_ENABLED = 0
eke.exe: GOLANG_IMAGE = golang:1.17-alpine

eke.exe eke: $(GO_SRCS) go.sum
	CGO_ENABLED=$(BUILD_GO_CGO_ENABLED) GOOS=$(TARGET_OS) GOARCH=$(GOARCH) $(GO) build $(BUILD_GO_FLAGS) -ldflags='$(LD_FLAGS)' -o $@ main.go
	rm -rf "$(BUILD_DIR)/bin" && mkdir -p "$(BUILD_DIR)/bin" && mv $@ $(BUILD_DIR)/bin
		
.PHONY: lint
lint:
	$(golint) run --verbose $(GO_DIRS)

.PHONY: check-unit
check-unit: go.sum
	echo "unit test"
	$(GO) test -race `$(GO) list $(GO_DIRS)`

check-unit \
clean-gocache: GO = \
  GOCACHE='$(CURDIR)/build/cache/go/build' \
  GOMODCACHE='$(CURDIR)/build/cache/go/mod' \
  go

.PHONY: clean-gocache
clean-gocache:
	$(GO) clean -cache -modcache

clean-docker-image:
	-docker rmi ekebuild.docker-image.eke -f
	-rm -f .ekebuild.docker-image.eke

.PHONY: clean
clean: clean-gocache clean-docker-image

SKIP_GOMOD_LINT ?= false
ifeq ($(SKIP_GOMOD_LINT), false)
GOMODLINT=lint-gomod
endif

GOMODTIDYLINT=sh -c '\
if [ `git diff go.mod go.sum | wc -l` -gt "0" ]; then \
	echo "Run \`go mod tidy\` and commit the result"; \
	exit 1; \
fi ; \
${GO} mod tidy; \
if [ `git diff go.mod go.sum | wc -l` -gt "0" ]; then \
 git checkout go.mod go.sum ; \
 echo "Linter failure: go.mod and go.sum have unused deps. Run \`go mod tidy\` and commit the result"; \
 exit 2; \
fi \
 ; ' GOMODTIDYLINT

lint-gomod:
	@${GOMODTIDYLINT}

generate-APIClient: hack/client-gen/boilerplate.go.txt
	$(go_clientgen) --go-header-file hack/client-gen/boilerplate.go.txt --input "eke/v1beta1" --input-base eke/pkg/apis --clientset-name="clientset" -p eke/pkg/apis/eke/
