GOBUILD=go build
BINARY_NAME=cf-ddns-agent
COV_REPORT=coverage.txt
CGO_ENABLED=0
RELEASE_DIR=release
BIN_DIR=bin

.DEFAULT_GOAL := local

VERSION=$(shell git describe --tags --always --dirty)
# linker flags for stripping debug info and injecting version info
LD_FLAGS="-s -w -X main.Version=$(VERSION)"

# Targets we want to build
PLATFORMS=linux/386 linux/amd64 linux/arm linux/arm64 linux/mips linux/mips64 linux/ppc64 linux/riscv64 \
freebsd/386 freebsd/amd64 freebsd/arm freebsd/arm64 \
openbsd/386 openbsd/amd64 openbsd/arm openbsd/arm64 openbsd/mips64 \
netbsd/386 netbsd/amd64 netbsd/arm netbsd/arm64 \
darwin/amd64 windows/amd64 darwin/arm64 \
solaris/amd64

# Used for help output
HELP_SPACING=15
HELP_COLOR=33
HELP_FORMATSTRING="\033[$(HELP_COLOR)m%-$(HELP_SPACING)s \033[00m%s.\n"

GO_FILES?=$(shell find . -name '*.go' | grep -v vendor)

EXTERNAL_TOOLS=\
	golang.org/x/tools/cmd/goimports \
	github.com/client9/misspell/cmd/misspell \
	github.com/mgechev/revive


# template rules which are used to construct the targets
# parameter 1: OS, parameter 2: architecture, parameter 3: target name
define TARGETRULE
$3: $(GO_FILES)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$1 GOARCH=$2 $(GOBUILD) -v -ldflags=$(LD_FLAGS) -o $3 ./cmd/cf-ddns-agent
endef

# some stuff to split the os/arch string to their own variables
get_goos = $(firstword $(subst /, ,$(1)))
get_goarch = $(word 2,$(subst /, ,$(1)))
# aux functions called by get_target_output
get_unix_target = $(BIN_DIR)/$(BINARY_NAME)_$(VERSION)_$(call get_goos,$1)_$(call get_goarch,$1)
get_win_target = $(BIN_DIR)/$(BINARY_NAME)_$(VERSION)_$(call get_goos,$1)_$(call get_goarch,$1).exe
# determine output filename for a platform
get_target_output = $(if $(findstring windows,$(call get_goos,$1)),$(call get_win_target,$1),$(call get_unix_target,$1))

# construct our targets
$(foreach platform,$(PLATFORMS),$(eval $(call TARGETRULE,$(call get_goos,$(platform)),$(call get_goarch,$(platform)),$(call get_target_output,$(platform)))))

# construct string of all target names
ALL_PLATFORMS = $(foreach platform,$(PLATFORMS),$(call get_target_output,$(platform)))

all-platforms: $(ALL_PLATFORMS)
	@ls -l $(BIN_DIR)

.PHONY: all-platforms local clean fmt check_spelling fix_spelling vet bootstrap help tests lint release

local:
	@echo "*** Building local binary... ***"
	$(GOBUILD) -o $(BINARY_NAME) -v -ldflags=$(LD_FLAGS)  ./cmd/cf-ddns-agent
	@echo "*** Done ***"

clean:
	@echo "*** Cleaning up object files... ***"
	rm -r -f $(RELEASE_DIR)/*
	rm -r -f $(BIN_DIR)/*
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe
	rm -f coverage.txt
	go clean
	@echo "*** Done ***"

fmt:
	@echo "*** Applying gofmt on all .go files (excluding vendor)... ***"
	@goimports -w $(GO_FILES)
	@echo "*** Done ***"

lint:
	@revive $(GO_FILES)

check-spelling:
	@echo "*** Check for common spelling mistakes in .go files... ***"
	@misspell -error $(GO_FILES)
	@echo "*** Done ***"

fix-spelling:
	@echo "*** Fix any encountered spelling mistakes in .go files... ***"
	@misspell -w $(GO_FILES)
	@echo "*** Done ***"

vet:
	@echo "*** Running vet on package directories... ***"
	@go list ./... | grep -v /vendor/ | xargs go vet
	@echo "*** Done ***"

bootstrap:
	@echo "*** Installing required tools for building... ***"
	@for tool in  $(EXTERNAL_TOOLS) ; do \
		echo "Installing/Updating $$tool" ; \
		GO111MODULE=off go get -u $$tool; \
	done
	@echo "*** Done ***"

test:
	@go test -race -coverprofile=$(COV_REPORT) -covermode=atomic ./...

release: all-platforms
	build/package/make_release.sh $(BIN_DIR) $(RELEASE_DIR) $(VERSION)

help:
	@printf "\n*** Available make targets ***\n\n"
	@printf $(HELP_FORMATSTRING) "local" "Build executable for your OS, for testing purposes (this is the default)"
	@printf $(HELP_FORMATSTRING) "help" "This message"
	@printf $(HELP_FORMATSTRING) "bootstrap" "Install tools needed for build"
	@printf $(HELP_FORMATSTRING) "all-platforms" "Compile for all platforms"
	@printf $(HELP_FORMATSTRING) "vet" "Checks code for common mistakes"
	@printf $(HELP_FORMATSTRING) "lint" "Perform lint/revive check"
	@printf $(HELP_FORMATSTRING) "test" "Run tests"
	@printf $(HELP_FORMATSTRING) "fmt" "Fix formatting on .go files"
	@printf $(HELP_FORMATSTRING) "check-spelling" "Show potential spelling mistakes"
	@printf $(HELP_FORMATSTRING) "fix-spelling" "Correct detected spelling mistakes"
	@printf $(HELP_FORMATSTRING) "clean" "Clean your working directory"
	@printf "\n*** End ***\n\n"
