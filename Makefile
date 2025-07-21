PROJECT_NAME := FocusFrame
CMD_DIR := ./cmd
BIN_DIR := ./bin

VERSION := v0.1.0

GO := go
GO_FLAGS := -ldflags "-H=windowsgui -X github.com/skryvvara/focusframe/config.Version=${VERSION}"

all: clean build

.PHONY: build
build:
	$(GO) build $(GO_FLAGS) -o $(BIN_DIR)/$(PROJECT_NAME).exe $(CMD_DIR)

.PHONY: dev
dev:
	${GO} run ./cmd

.PHONY: test
test:
	$(GO) test -v ./...

.PHONY: syso
syso:
	rsrc -ico ./cmd/monitor.ico -o ./cmd/focusframe.syso

.PHONY: vendor
vendor:
	$(GO) mod tidy
	@GOFLAGS="-mod=readonly" $(GO) mod vendor
	rm go.sum

# Clean build artifacts
clean:
	rm -rf $(BIN_DIR)
	mkdir -p $(BIN_DIR)
