#!/usr/bin/make -f
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
TIME ?= $(shell date +%Y-%m-%dT%H:%M:%S%z)

# don't override user values
ifeq (,$(VERSION))
  VERSION := $(shell git describe --tags)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif


ldflags = -X oproller/version.BuildVersion=$(VERSION) \
		  -X oproller/version.BuildCommit=$(COMMIT) \
		  -X oproller/version.BuildTime=$(TIME)

BUILD_FLAGS := -ldflags '$(ldflags)'
# ---------------------------------------------------------------------------- #
#                                 Make targets                                 #
# ---------------------------------------------------------------------------- #
.PHONY: install
install: go.sum ## Installs the roller binary
	go mod tidy
	go install -mod=readonly $(BUILD_FLAGS) .


.PHONY: build
build: ## Compiles the roller binary
	go mod tidy
	go build -o build/roller $(BUILD_FLAGS) .