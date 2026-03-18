.PHONY: build test test-cover vet clean

BIN_DIR := bin
PACKAGES := ./cmd/... ./pkg/... ./internal/...

build: $(BIN_DIR)/swisseph-mcp

$(BIN_DIR)/swisseph-mcp: $(shell find cmd pkg internal -name '*.go') version.go
	@mkdir -p $(BIN_DIR)
	go build -ldflags="-s -w" -o $(BIN_DIR)/swisseph-mcp ./cmd/server

test:
	go test $(PACKAGES)

test-cover:
	go test -coverprofile=coverage.out $(PACKAGES)
	go tool cover -func=coverage.out | tail -1

vet:
	go vet $(PACKAGES)

clean:
	rm -rf $(BIN_DIR) coverage.out
