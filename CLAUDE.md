# Development Guide

## Build & Test

```bash
make build        # Build MCP server to bin/solarsage-mcp
make build-api    # Build REST API server to bin/solarsage-api
make test         # Run all tests (33 packages, 800+ tests)
make test-race    # Run tests with race detector
make test-cover   # Run tests with coverage report
make cover-html   # Generate HTML coverage report
make bench        # Run benchmarks
make vet          # Run go vet
make check        # Run vet + test
make clean        # Remove build artifacts
```

## Architecture

The project is a Go MCP server wrapping the Swiss Ephemeris C library via CGO. It's also usable as a standalone Go library.

### Package Layers

**High-level API:**
- `pkg/solarsage/` - **Convenience API** (33 functions, datetime strings, sensible defaults)
- `pkg/mcp/` - MCP protocol layer (35 tools)
- `pkg/api/` - **RESTful HTTP API** (35 endpoints, JSON, CORS, API key auth)
- `pkg/report/` - Comprehensive chart analysis report

**Chart Calculations:**
- `pkg/chart/` - Natal/synastry charts (11 house systems)
- `pkg/transit/` - Transit event detection (7 types, 100% validated)
- `pkg/progressions/` - Secondary progressions & solar arc
- `pkg/returns/` - Solar/lunar return charts
- `pkg/composite/` - Composite midpoint & Davison charts
- `pkg/synastry/` - Compatibility scoring

**Predictive Techniques:**
- `pkg/primary/` - Primary Directions (Ptolemy semi-arc, Naibod key)
- `pkg/symbolic/` - Symbolic Directions (1°/year, Naibod, Profection, custom rate)

**Traditional Astrology:**
- `pkg/dignity/` - Essential dignities, mutual receptions, sect, bonification & maltreatment
- `pkg/heliacal/` - Heliacal risings/settings (Swiss Ephemeris visibility algorithms)
- `pkg/dispositor/` - Dispositorship chains, final dispositor
- `pkg/lots/` - 15+ Arabic lots with day/night reversal
- `pkg/bounds/` - Chaldean decans, Egyptian terms
- `pkg/profection/` - Annual/monthly profections
- `pkg/firdaria/` - Firdaria planetary period system (day/night sequences)
- `pkg/antiscia/` - Solstice/equinox mirror points
- `pkg/planetary/` - Chaldean planetary hours

**Analysis:**
- `pkg/fixedstars/` - 50+ star catalog with conjunctions
- `pkg/midpoint/` - Midpoints, 90deg dial, activations
- `pkg/harmonic/` - Harmonic charts (1-180)
- `internal/aspect/` - Aspect math + 7 pattern types

**Astronomical:**
- `pkg/lunar/` - Phases, eclipse detection

**Visualization:**
- `pkg/render/` - Chart wheel coordinates for SVG/Canvas

**Infrastructure:**
- `pkg/sweph/` - Swiss Ephemeris CGO bindings (thread-safe)
- `pkg/models/` - Core types with Stringer interfaces
- `pkg/julian/` - Julian Day / ISO 8601 conversions
- `pkg/geo/` - Geocoding and timezone
- `pkg/export/` - CSV/JSON export

## Key Constraints

- Transit detection results are validated against Solar Fire 9 at 100% match (247/247 events). Changes to `pkg/transit/transit.go` must preserve this accuracy.
- The `pkg/sweph/` package uses a global mutex because the Swiss Ephemeris C library is not thread-safe.
- Ephemeris data files (`.se1`) must be available at runtime. Set `SWISSEPH_EPHE_PATH` or use the default relative path.

## Testing

Run all tests: `make test`

The `pkg/transit/solarfire_test.go` validates against a Solar Fire reference CSV. This is the primary accuracy gate.

Coverage tests (`*_coverage_test.go`) exercise edge cases for coverage metrics.

Current: 38 packages, 824+ tests, race-free.
