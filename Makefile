PROJECT_NAME := FocusFrame
CMD_DIR := ./cmd
BIN_DIR := ./bin

VERSION := v0.1.0

GO := go
GO_FLAGS := -ldflags "-H=windowsgui -X main.Version=${VERSION}"

all: clean build

.PHONY: build
build:
	$(GO) build $(GO_FLAGS) -o $(BIN_DIR)/$(PROJECT_NAME).exe $(CMD_DIR)

.PHONY: vendor
vendor:
	$(GO) mod tidy
	@GOFLAGS="-mod=readonly" $(GO) mod vendor
	rm go.sum

# Clean build artifacts
clean:
	rm -rf $(BIN_DIR)
	mkdir -p $(BIN_DIR)
