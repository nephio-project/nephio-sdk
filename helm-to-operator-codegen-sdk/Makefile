GO_VERSION ?= 1.20.2
GOLANG_CI_VER ?= v1.52
GOSEC_VER ?= 2.15.0
TEST_COVERAGE_FILE=lcov.info
TEST_COVERAGE_HTML_FILE=coverage_unit.html
TEST_COVERAGE_FUNC_FILE=func_coverage.out

# CONTAINER_RUNNABLE checks if tests and lint check can be run inside container.
PODMAN ?= $(shell podman -v > /dev/null 2>&1; echo $$?)
ifeq ($(PODMAN), 0)
CONTAINER_RUNTIME=podman
else
CONTAINER_RUNTIME=docker
endif
CONTAINER_RUNNABLE ?= $(shell $(CONTAINER_RUNTIME) -v > /dev/null 2>&1; echo $$?)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


.PHONY: unit_clean
unit_clean: ## clean up the unit test artifacts created
ifeq ($(CONTAINER_RUNNABLE), 0)
		$(CONTAINER_RUNTIME) system prune -f
endif
		rm -f ${TEST_COVERAGE_FILE} ${TEST_COVERAGE_HTML_FILE} ${TEST_COVERAGE_FUNC_FILE}

.PHONY: unit
unit: ## Run unit tests against code. Installing Helm also as a pre-requisite
ifeq ($(CONTAINER_RUNNABLE), 0)
		$(CONTAINER_RUNTIME) run -it -v ${PWD}:/go/src -w /go/src docker.io/library/golang:${GO_VERSION}-alpine3.17 \
         /bin/sh -c "wget https://get.helm.sh/helm-v3.9.3-linux-amd64.tar.gz; \
		tar xvf helm-v3.9.3-linux-amd64.tar.gz; \
		mv linux-amd64/helm /usr/local/bin; \
		rm -rf linux-amd64 helm-v3.9.3-linux-amd64.tar.gz; \
		go test ./... -v -coverprofile ${TEST_COVERAGE_FILE}; \
		go tool cover -html=${TEST_COVERAGE_FILE} -o ${TEST_COVERAGE_HTML_FILE}; \
		go tool cover -func=${TEST_COVERAGE_FILE} -o ${TEST_COVERAGE_FUNC_FILE}"
else
		go test ./... -v -coverprofile ${TEST_COVERAGE_FILE}
		go tool cover -html=${TEST_COVERAGE_FILE} -o ${TEST_COVERAGE_HTML_FILE}
		go tool cover -func=${TEST_COVERAGE_FILE} -o ${TEST_COVERAGE_FUNC_FILE}
endif

# Install link at https://golangci-lint.run/usage/install/ if not running inside a container
.PHONY: lint
lint: ## Run lint  against code.
ifeq ($(CONTAINER_RUNNABLE), 0)
		$(CONTAINER_RUNTIME) run -it -v ${PWD}:/go/src -w /go/src docker.io/golangci/golangci-lint:${GOLANG_CI_VER}-alpine golangci-lint run ./... -v
else
		golangci-lint run ./... -v --timeout 10m
endif

# Install link at https://github.com/securego/gosec#install if not running inside a container
.PHONY: gosec
gosec: ## inspects source code for security problem by scanning the Go Abstract Syntax Tree
ifeq ($(CONTAINER_RUNNABLE), 0)
		$(CONTAINER_RUNTIME) run -it -v ${PWD}:/go/src -w /go/src docker.io/securego/gosec:${GOSEC_VER} ./...
else
		gosec ./...
endif
