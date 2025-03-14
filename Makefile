GOBUILD=go build
BINARY_NAME=cf-ddns-agent
COV_REPORT=coverage.txt
CGO_ENABLED=0

.DEFAULT_GOAL := local

VERSION=$(shell git describe --tags --always --dirty)
# linker flags for stripping debug info and injecting version info
LD_FLAGS="-s -w -X github.com/headincloud/cf-ddns-agent/cmd.Version=$(VERSION)"

# Used for help output
HELP_SPACING=15
HELP_COLOR=33
HELP_FORMATSTRING="\033[$(HELP_COLOR)m%-$(HELP_SPACING)s \033[00m%s.\n"

GO_FILES?=$(shell find . -name '*.go' | grep -v vendor)

EXTERNAL_TOOLS=\
	golang.org/x/tools/cmd/goimports \
	github.com/client9/misspell/cmd/misspell \
	github.com/mgechev/revive

all-platforms:
	goreleaser build --clean --snapshot

.PHONY: all-platforms local clean fmt check_spelling fix_spelling vet bootstrap help tests lint release

local:
	@echo "*** Building local binary... ***"
	$(GOBUILD) -o $(BINARY_NAME) -v -ldflags=$(LD_FLAGS)  .
	@echo "*** Done ***"

clean:
	@echo "*** Cleaning up object files... ***"
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
		go install $$tool; \
	done
	@echo "*** Done ***"

test:
	@go test -race -coverprofile=$(COV_REPORT) -covermode=atomic ./...

release:
	goreleaser release --clean --auto-snapshot

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
