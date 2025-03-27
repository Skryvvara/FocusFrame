PROJECT_NAME := FocusFrame
CMD_DIR := ./cmd
BIN_DIR := .\bin

RM_CMD := rm -rf
MKDIR_CMD := mkdir -p

ifeq ($(OS),Windows_NT)
	RM_CMD = rmdir /s /q
	MKDIR_CMD = mkdir
endif

VERSION := v0.1.0

GO := go
GO_FLAGS := -ldflags "-H=windowsgui -X main.Version=${VERSION}"

all: clean build

.PHONY: build
build:
	$(GO) build $(GO_FLAGS) -o $(BIN_DIR)/$(PROJECT_NAME).exe $(CMD_DIR)

.PHONY: test
test:
	$(GO) test -v ./...

.PHONY: vendor
vendor:
	$(GO) mod tidy
	@GOFLAGS="-mod=readonly" $(GO) mod vendor
	rm go.sum

# Clean build artifacts
clean:
	$(RM_CMD) $(BIN_DIR)
	$(MKDIR_CMD) $(BIN_DIR)
