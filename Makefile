.DEFAULT_GOAL := build

APP_NAME := needs-retitle

BINDIR := bin
RELEASEDIR := release

LDFLAGS := -extldflags "-static"

BUILD_PATH = github.com/ouzi-dev/needs-retitle

GOLANGCI_LINT_VERSION := v1.23.6
GOLANG_VERSION := 1.14.2

HAS_GOX := $(shell command -v gox;)
HAS_GO_IMPORTS := $(shell command -v goimports;)
HAS_GO_MOCKGEN := $(shell command -v mockgen;)
HAS_GOLANGCI_LINT := $(shell command -v golangci-lint;)
GOLANGCI_VERSION_CHECK := $(shell golangci-lint --version | grep -oh $(GOLANGCI_LINT_VERSION);)

DOCKER_REGISTRY ?= quay.io
DOCKER_REPO ?= $(DOCKER_REGISTRY)/ouzi/$(APP_NAME)
DOCKER_IMAGE ?= $(APP_NAME)

SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

GIT_SHORT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_TAG    := $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_DIRTY  = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")

TMP_VERSION := canary

BINARY_VERSION := ""

ifndef VERSION
ifeq ($(GIT_DIRTY), clean)
ifdef GIT_TAG
	TMP_VERSION = $(GIT_TAG)
	BINARY_VERSION = $(GIT_TAG)
endif
endif
else
  BINARY_VERSION = $(VERSION)
endif

VERSION ?= $(TMP_VERSION)

DIST_DIR := _dist
TARGETS   ?= darwin/amd64 linux/amd64 windows/amd64
TARGET_DIRS = find * -type d -exec

# Only set Version if building a tag or VERSION is set
ifneq ($(BINARY_VERSION),"")
	LDFLAGS += -X $(BUILD_PATH)/pkg/version.Version=$(VERSION)
	CHART_VERSION = $(VERSION)
endif

LDFLAGS += -X $(BUILD_PATH)/pkg/version.GitCommit=$(GIT_SHORT_COMMIT)

.PHONY: info
info:
	@echo "How are you:       $(GIT_DIRTY)"
	@echo "Version:           $(VERSION)"
	@echo "Git Tag:           $(GIT_TAG)"
	@echo "Git Commit:        $(GIT_SHORT_COMMIT)"
	@echo "binary:            $(BINARY_VERSION)"

build: clean-bin info bootstrap tidy generate fmt 
	@echo "build target..."
	@CGO_ENABLED=0 GOARCH=amd64 go build -o $(BINDIR)/$(APP_NAME) -ldflags '$(LDFLAGS)' cmd/needs-retitle/main.go 

.PHONY: clean-bin
clean-bin: 
	@rm -rf $(BINDIR)

.PHONY: clean
clean: 
	@rm -rf $(DIST_DIR)

.PHONY: tidy
tidy:
	@echo "tidy target..."
	@go mod tidy

.PHONY: generate
generate:
	@echo "generate target..."
	go generate ./...

.PHONY: vendor
vendor: tidy
	@echo "vendor target..."
	@go mod vendor

.PHONY: test
test: generate build
	@echo "test target..."
	@go test ./... -v -count=1

.PHONY: lint
lint: bootstrap bootstrap-lint build
	@echo "lint target..."
	@golangci-lint run --enable-all --disable lll,nakedret,funlen,gochecknoglobals,gomnd,wsl,errcheck ./...

.PHONY: bootstrap-lint
bootstrap-lint:
	@echo "bootstrap lint..."
ifndef HAS_GOLANGCI_LINT
	@echo "golangci-lint $(GOLANGCI_LINT_VERSION) not found..."
	@GOPROXY=direct GOSUMDB=off go get -u github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
else
	@echo "golangci-lint found, checking version..."
ifeq ($(GOLANGCI_VERSION_CHECK), )
	@echo "found different version, installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@GOPROXY=direct GOSUMDB=off go get -u github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
else
	@echo "golangci-lint version $(GOLANGCI_VERSION_CHECK) found!"
endif
endif

.PHONY: bootstrap
bootstrap: 
	@echo "bootstrap target..."
ifndef HAS_GO_IMPORTS
	@go get golang.org/x/tools/cmd/goimports
endif
ifndef HAS_GO_MOCKGEN
	@go get -u github.com/golang/mock/gomock
	@go install github.com/golang/mock/mockgen
endif
ifndef HAS_GOX
	@go get -u github.com/mitchellh/gox
endif

.PHONY: fmt
fmt:
	@echo "fmt target..."
	@gofmt -l -w -s $(SRC)

.PHONY: semantic-release
semantic-release:
	@npm ci
	@npx semantic-release

.PHONY: semantic-release-dry-run
semantic-release-dry-run:
	@npm ci
	@npx semantic-release -d

.PHONY: install-npm-check-updates
install-npm-check-updates:
	npm install npm-check-updates

.PHONY: update-dependencies
update-dependencies: install-npm-check-updates
	ncu -u
	npm install

.PHONY: docker-login
docker-login:
	@docker login $(DOCKER_REGISTRY) --username $(DOCKER_USERNAME) --password $(DOCKER_PASSWORD)

.PHONY: docker-build
docker-build:
	@docker build -t $(DOCKER_IMAGE):$(VERSION) -f hack/Dockerfile --build-arg GOLANG_VERSION=$(GOLANG_VERSION) --build-arg VERSION=$(VERSION) . 
	@docker tag $(DOCKER_IMAGE):$(VERSION) ${DOCKER_REPO}/$(DOCKER_IMAGE):$(VERSION)

.PHONY: docker-push
docker-push: 
	@docker push ${DOCKER_REPO}/$(DOCKER_IMAGE):$(VERSION)

.PHONY: docker-release
docker-release: docker-login docker-build docker-push
