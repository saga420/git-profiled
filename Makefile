# Constants
GREEN := $(shell tput setaf 2)
NORMAL := $(shell tput sgr0)

PKG := $(shell go list -m)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2> /dev/null || echo "unknown")
GIT_COMMIT_TIME := $(shell git show -s --format=%ct "$(GIT_COMMIT)" 2> /dev/null || echo "unknown")
APPNAME := $(shell basename "$(shell git rev-parse --show-toplevel)")
TARGETS := darwin-amd64 linux-amd64

# Flags
LDFLAGS := -s -w -X $(PKG)/version.GitRevision=$(GIT_COMMIT) -X $(PKG)/version.GitCommitAt=$(GIT_COMMIT_TIME)

# Targets
.PHONY: all
all: build

.PHONY: build
build: $(TARGETS)

$(TARGETS):
	$(eval GOOS := $(word 1,$(subst -, ,$@)))
	$(eval GOARCH := $(word 2,$(subst -, ,$@)))
	@echo "$(GREEN)Building for $(GOOS)/$(GOARCH)...$(NORMAL)"
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "$(LDFLAGS)" -o "build/bin/$(APPNAME)_$(GOOS)"

.PHONY: clean
clean:
	@echo "$(GREEN)Cleaning up...$(NORMAL)"
	rm -rf build/
