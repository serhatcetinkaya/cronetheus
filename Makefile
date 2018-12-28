GO := go
GOLINT := golint

PACKAGES := $(shell $(GO) list ./... | grep -v /vendor/)

all: clean format lint build

clean:
	@echo ">> cleaning"
	@rm -rf cronetheus

format:
	@echo ">> formatting code"
	@$(GO) fmt $(PACKAGES)

lint:
	@echo ">> running golint"
	@$(GOLINT) $(PACKAGES)

build:
	@echo ">> building binaries"
	@GOOS=linux GOARCH=amd64 $(GO) build github.com/serhatcetinkaya/cronetheus/cmd/cronetheus

