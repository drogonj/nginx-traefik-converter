include makefiles/security.mk

GOFMT_FILES?=$(shell find . -not -path "./vendor/*" -type f -name '*.go')
APP_NAME?=nginx-traefik-converter
APP_DIR?=$(shell git rev-parse --show-toplevel)
VERSION?=0.1.0
REVISION?=$(shell git rev-parse --verify HEAD)
DATE?=$(shell date)
PLATFORM?=$(shell go env GOOS)
ARCHITECTURE?=$(shell go env GOARCH)
GOVERSION?=$(shell go version | awk '{printf $$3}')
BUILD_WITH_FLAGS=-s -w \
	-X 'github.com/nikhilsbhat/nginx-traefik-converter/version.Version=$(VERSION)' \
	-X 'github.com/nikhilsbhat/nginx-traefik-converter/version.Env=$(BUILD_ENVIRONMENT)' \
	-X 'github.com/nikhilsbhat/nginx-traefik-converter/version.BuildDate=$(DATE)' \
	-X 'github.com/nikhilsbhat/nginx-traefik-converter/version.Revision=$(REVISION)' \
	-X 'github.com/nikhilsbhat/nginx-traefik-converter/version.Platform=$(PLATFORM)/$(ARCHITECTURE)' \
	-X 'github.com/nikhilsbhat/nginx-traefik-converter/version.GoVersion=$(GOVERSION)'

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: help
help: ## Prints help (only for targets with comments)
	@grep -E '^[a-zA-Z0-9._-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

local/fmt: ## Formats all the go code in the application.
	@gofmt -w $(GOFMT_FILES)
	$(GOBIN)/goimports -w $(GOFMT_FILES)
	$(GOBIN)/gofumpt -l -w $(GOFMT_FILES)
	$(GOBIN)/gci write $(GOFMT_FILES) --skip-generated

local/check: local/fmt ## Formats code and syncs vendor directory
	@go mod vendor
	@go mod tidy

local/build: local/check ## Builds the application binary with version info
	@go build -o $(APP_NAME)_v$(VERSION) -ldflags="$(BUILD_WITH_FLAGS)"

local/staticbuild: local/check ## Builds a static binary
	@CGO_ENABLED=0 go build -o $(APP_NAME)_v$(VERSION) -ldflags="$(BUILD_WITH_FLAGS)" -a

lint: ## Runs golangci-lint
	@golangci-lint run --color always

test: ## Runs test cases with coverage report
	@go test ./... -mod=vendor -coverprofile cover.out
	@go tool cover -html=cover.out -o cover.html

generate/document: ## Generates CLI documentation
	@go generate github.com/nikhilsbhat/nginx-traefik-converter/docs
