GO := go

PACKAGES := $(shell $(GO) list ./... | grep -v /vendor/)

all: clean format build

clean:
	@echo ">> cleaning"
	@rm -rf cronetheus

format:
	@echo ">> formatting code"
	@$(GO) fmt $(PACKAGES)

build:
	@echo ">> building binaries"
	@GOOS=linux GOARCH=amd64 $(GO) build github.com/serhatck/cronetheus/cmd/cronetheus

