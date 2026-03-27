.PHONY: build build-api test test-race test-cover cover-html bench vet check clean docs

BIN_DIR := bin
PACKAGES := ./cmd/... ./pkg/... ./internal/...
LIB_PACKAGES := ./pkg/... ./internal/...

build: $(BIN_DIR)/solarsage-mcp

$(BIN_DIR)/solarsage-mcp: $(shell find cmd pkg internal -name '*.go') version.go
	@mkdir -p $(BIN_DIR)
	go build -ldflags="-s -w" -o $(BIN_DIR)/solarsage-mcp ./cmd/server

build-api: $(BIN_DIR)/solarsage-api

$(BIN_DIR)/solarsage-api: $(shell find cmd pkg internal -name '*.go') version.go
	@mkdir -p $(BIN_DIR)
	go build -ldflags="-s -w" -o $(BIN_DIR)/solarsage-api ./cmd/api

test:
	go test $(PACKAGES)

test-race:
	go test -race $(LIB_PACKAGES)

test-cover:
	go test -coverprofile=coverage.out $(LIB_PACKAGES)
	go tool cover -func=coverage.out | tail -1

cover-html: test-cover
	go tool cover -html=coverage.out -o coverage.html
	@echo "Open coverage.html in your browser"

bench:
	go test -bench=. -benchmem -run=^$$ ./pkg/chart/ ./pkg/transit/ ./pkg/api/

vet:
	go vet $(PACKAGES)

check: vet test
	@echo "All checks passed"

clean:
	rm -rf $(BIN_DIR) coverage.out coverage.html

docs:
	@which gomarkdoc > /dev/null 2>&1 || go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
	@mkdir -p doc
	@for pkg in antiscia api bounds chart composite dignity dispositor \
	            export firdaria fixedstars geo harmonic heliacal julian \
	            lots lunar mcp midpoint models planetary primary profection progressions \
	            render report returns solarsage sweph symbolic synastry transit; do \
	    gomarkdoc --output doc/pkg-$${pkg}.md ./pkg/$${pkg}/; \
	done
	@gomarkdoc --output doc/internal-aspect.md ./internal/aspect/
	@echo "API docs generated in doc/"
