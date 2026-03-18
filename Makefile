.PHONY: build test clean

# Binary output directory
BIN_DIR := bin

# Main targets
build: $(BIN_DIR)/swisseph-mcp

$(BIN_DIR)/swisseph-mcp: $(shell find cmd pkg internal -name '*.go')
	@mkdir -p $(BIN_DIR)
	go build -ldflags="-s -w" -o $(BIN_DIR)/swisseph-mcp ./cmd/server

test:
	go test ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

clean:
	rm -rf $(BIN_DIR) coverage.out
