# Development Guide

## Build & Test

```bash
make build        # Build binary to bin/swisseph-mcp
make test         # Run all tests
make test-cover   # Run tests with coverage report
make clean        # Remove build artifacts
```

## Architecture

The project is a Go MCP server wrapping the Swiss Ephemeris C library via CGO.

- `pkg/sweph/` - CGO bindings (thread-safe via mutex; Swiss Ephemeris is not thread-safe)
- `pkg/chart/` - Chart calculations depend on `sweph` and `internal/aspect`
- `pkg/transit/` - Transit engine depends on `chart`, `progressions`, `sweph`
- `pkg/mcp/` - MCP protocol layer, depends on all `pkg/` packages
- `internal/aspect/` - Pure Go aspect math, no external dependencies

## Key Constraints

- Transit detection results are validated against Solar Fire 9 at 100% match (247/247 events). Changes to `pkg/transit/transit.go` must preserve this accuracy.
- The `pkg/sweph/` package uses a global mutex because the Swiss Ephemeris C library is not thread-safe.
- Ephemeris data files (`.se1`) must be available at runtime. Set `SWISSEPH_EPHE_PATH` or use the default relative path.

## Testing

Run all tests: `make test`

The `pkg/transit/solarfire_test.go` validates against a Solar Fire reference CSV. This is the primary accuracy gate.

Coverage tests (`*_coverage_test.go`) exercise edge cases for coverage metrics.
