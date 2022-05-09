# Name of output binary.
BIN_NAME := $(or $(BIN_NAME),refeedmutator)

# Makefile variables.
MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_DIR := $(dir $(MAKEFILE_PATH))

# Artifacts output directory.
ARTIFACTS_DIR := $(or $(ARTIFACTS_DIR),bin)

# Artifacts output directory abspath.
ARTIFACTS_ABS_PATH := $(addprefix $(PROJECT_DIR),$(ARTIFACTS_DIR))

# Version for go mod tidy -compat flag.
GO_MOD_COMPAT_VERSION := 1.18

# Version of linter.
#  - https://github.com/golangci/golangci-lint/releases/tag/v1.45.2
#  - https://github.com/mgechev/revive/tree/v1.1.4
GOLANGCI_LINT_VERSION := $(or $(GOLANGCI_LINT_VERSION),v1.45.2)

# Set common utilities environs.
DATE_BIN := $(or $(DATE_BIN),date)
ENV_BIN := $(or $(ENV_BIN),env)
GIT_BIN := $(or $(GIT_BIN),git)
GO_BIN := $(or $(GO_BIN),go)
MKDIR_BIN := $(or $(MKDIR_BIN),mkdir)
RM_BIN := $(or $(RM_BIN),rm)
SH_BIN := $(or $(SH_BIN),sh)
TEST_BIN := $(or $(TEST_BIN),test)
WGET_BIN := $(or $(WGET_BIN),wget)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set).
ifeq (,$(shell $(GO_BIN) env GOBIN))
	GOBIN = $(shell $(GO_BIN) env GOPATH)/bin
else
	GOBIN = $(shell $(GO_BIN) env GOBIN)
endif

# Build flags.
BUILD_ENV = CGO_ENABLED=0 GOOS=linux GOARCH=amd64
# https://github.com/golang/go/issues/26492
BUILD_ARGS = -buildvcs=true -ldflags "-extldflags \"-static\""

# Set golangci-lint binary path.
GOLANGCI_LINT_BIN=$(GOBIN)/golangci-lint

all: clean tidy verify check-git-clean lint test build

# Download golangci-lint if needed.
.PHONY: golangci-lint
golangci-lint:
ifeq ("$(wildcard $(GOLANGCI_LINT_BIN))","")
	$(WGET_BIN) -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | $(SH_BIN) -s -- -b $(GOBIN) $(GOLANGCI_LINT_VERSION)
endif

# Run linter.
.PHONY: lint
lint: golangci-lint
	$(GOLANGCI_LINT_BIN) config path
	$(GOLANGCI_LINT_BIN) -v run --sort-results ./...

# Run tests.
.PHONY: test
test:
	$(ENV_BIN) $(GO_BUILD_ENV) $(GO_BIN) test -v -failfast ./...

# Update golang dependencies.
.PHONY: tidy
tidy:
	$(TEST_BIN) -d vendor || $(ENV_BIN) $(BUILD_ENV) $(GO_BIN) mod tidy -v -compat=$(GO_MOD_COMPAT_VERSION)

# Update vendor directory.
.PHONY: vendor
vendor:
	$(GO_BIN) mod vendor -v && $(GIT_BIN) status -s

# Verify golang dependencies.
.PHONY: verify
verify:
	$(TEST_BIN) -d vendor || $(GO_BIN) mod verify

# Create artifacts directory.
.PHONY: artifacts-dir
artifacts-dir:
	$(MKDIR_BIN) -vp "$(ARTIFACTS_ABS_PATH)"

# Clean artifacts directory and cache.
.PHONY: clean
clean:
	$(GOLANGCI_LINT_BIN) cache clean -v
	$(GO_BIN) clean -cache -testcache -fuzzcache -x
	$(RM_BIN) -vrf "$(ARTIFACTS_ABS_PATH)"

# Build application.
.PHONY: build
build: artifacts-dir
	$(ENV_BIN) $(BUILD_ENV) $(GO_BIN) build $(BUILD_ARGS) -v -a -o "$(ARTIFACTS_ABS_PATH)/$(BIN_NAME)"

# Fail when directory tree is dirty.
.PHONY: check-git-clean
check-git-clean:
	@status=$$($(GIT_BIN) status --porcelain=v1); \
	if [ ! -z "$${status}" ]; then \
		echo "Error: working directory tree is dirty."; \
		exit 1; \
	fi

# Show git diff helper (with excludes).
.PHONY: diff
diff:
	$(GIT_BIN) diff --diff-algorithm=minimal --ignore-all-space -- ":(exclude)vendor/*" ":(exclude)go.sum"
