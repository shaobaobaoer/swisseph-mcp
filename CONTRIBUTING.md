# Contributing

Contributions are welcome! This guide covers how to set up the development environment, build the project, and submit changes.

## Table of Contents

- [Development Setup](#development-setup)
  - [Prerequisites](#prerequisites)
  - [Swiss Ephemeris Setup](#swiss-ephemeris-setup)
  - [Build](#build)
- [Running Tests](#running-tests)
- [Make Targets](#make-targets)
- [Guidelines](#guidelines)
- [Architecture](#architecture)

---

## Development Setup

### Prerequisites

- **Go 1.25+** — [golang.org/dl](https://golang.org/dl/)
- **GCC** — Required for CGO compilation of the Swiss Ephemeris C library
  - Linux: `sudo apt install build-essential` (Debian/Ubuntu) or `sudo yum install gcc` (RHEL/CentOS)
  - macOS: `xcode-select --install`
- **Git**

### Swiss Ephemeris Setup

SolarSage embeds the Swiss Ephemeris C library via CGO. The source code and ephemeris data files must be present at `third_party/swisseph/` before building.

**Step 1 — Download the Swiss Ephemeris source:**

```bash
# Clone the official Swiss Ephemeris repository
git clone https://github.com/aloistr/swisseph.git /tmp/swisseph
```

**Step 2 — Build the static library:**

```bash
cd /tmp/swisseph
make libswe.a
```

**Step 3 — Copy files into the project:**

```bash
# Create the directory
mkdir -p third_party/swisseph/ephe

# Copy C source files and headers
cp /tmp/swisseph/*.c third_party/swisseph/
cp /tmp/swisseph/*.h third_party/swisseph/

# Copy the compiled static library
cp /tmp/swisseph/libswe.a third_party/swisseph/

# Copy ephemeris data files (.se1)
cp /tmp/swisseph/ephe/*.se1 third_party/swisseph/ephe/
```

After this, `third_party/swisseph/` should contain:
```
third_party/swisseph/
  *.c           Swiss Ephemeris C source files
  *.h           Header files (including swephexp.h, swehouse.h)
  libswe.a      Pre-compiled static library (~1.6 MB)
  ephe/
    *.se1       Ephemeris data files (~150 files)
```

### JPL Ephemeris (Optional)

SolarSage supports JPL ephemeris files (DE200, DE406, DE431, DE440, DE441) as an alternative to the Swiss Ephemeris. Use JPL ephemeris when:

- **Matching Solar Fire results** — Solar Fire uses DE200/DE406; using DE406 gives identical planetary positions
- **Maximum precision** — DE440 (2021) is the most accurate modern ephemeris for 1550–2650

#### Download JPL Ephemeris Files

Download from JPL directly (rename after download):

```bash
cd third_party/swisseph/ephe

# DE406 (recommended for Solar Fire compatibility, 190 MB)
wget -O de406.eph https://ssd.jpl.nasa.gov/ftp/eph/planets/Linux/de406/lnxm3000p3000.406

# DE440 (highest precision for modern dates, 2.6 GB)
wget -O de440.eph https://ssd.jpl.nasa.gov/ftp/eph/planets/Linux/de441/linux_m13000p17000.441

# DE200 (legacy, 41 MB)
wget -O de200.eph https://ssd.jpl.nasa.gov/ftp/eph/planets/Linux/de200/lnxm1600p2170.200
```

#### Ephemeris Comparison

| Ephemeris | Year Range | Size | Precision | Use Case |
|-----------|------------|------|-----------|----------|
| DE440 | 1550–2650 | 2.6 GB | Highest | Modern high-precision work |
| DE431 | -13000 ~ +17000 | 2.6 GB | Very High | Swiss Ephemeris default |
| DE406 | -3000 ~ +3000 | 190 MB | High | Solar Fire compatibility |
| DE200 | 1599–2169 | 41 MB | Moderate | Legacy compatibility |

#### Configure Ephemeris Type

Set environment variables before running:

```bash
# Use JPL ephemeris (e.g., DE406 for Solar Fire compatibility)
export SWISSEPH_TYPE=jpl
export SWISSEPH_JPL_FILE=de406.eph

# Or use Swiss Ephemeris (default)
export SWISSEPH_TYPE=swiss

# Or use Moshier (built-in, no files needed, lower precision)
export SWISSEPH_TYPE=moshier
```

These can also be set programmatically:

```go
import "github.com/shaobaobaoer/solarsage-mcp/pkg/sweph"

// Switch to JPL DE406
sweph.SetEphemerisType(sweph.EphemerisJPL)
sweph.SetJPLFile("de406.eph")

// Switch back to Swiss Ephemeris
sweph.SetEphemerisType(sweph.EphemerisSwiss)
```

### Build

```bash
git clone https://github.com/shaobaobaoer/solarsage-mcp.git
cd solarsage-mcp

# (After Swiss Ephemeris setup above)
make build        # → bin/solarsage-mcp  (MCP server)
make build-api    # → bin/solarsage-api  (REST API server)
```

---

## Running Tests

```bash
make test          # Run all 824+ tests across 38 packages
make test-race     # Run with race detector (recommended before submitting PRs)
make test-cover    # Show line coverage summary
make cover-html    # Generate coverage.html report
make vet           # Run go vet
make check         # vet + test (full quality gate)
```

The `pkg/transit/solarfire_test.go` test validates 247 transit events against a Solar Fire 9 reference CSV. It is the primary accuracy gate and must always pass at 100% (247/247).

**Set the ephemeris path for tests:**

```bash
export SWISSEPH_EPHE_PATH=/path/to/solarsage-mcp/third_party/swisseph/ephe
make test
```

If `SWISSEPH_EPHE_PATH` is not set, the packages fall back to `third_party/swisseph/ephe` relative to the repository root.

---

## Make Targets

| Target | Description |
|--------|-------------|
| `make build` | Build MCP server binary to `bin/solarsage-mcp` |
| `make build-api` | Build REST API server binary to `bin/solarsage-api` |
| `make test` | Run all tests (38 packages, 824+ tests) |
| `make test-race` | Run tests with race detector |
| `make test-cover` | Run tests with coverage report (summary) |
| `make cover-html` | Generate `coverage.html` for browser viewing |
| `make bench` | Run benchmarks for chart, transit, and API packages |
| `make vet` | Run `go vet` on all packages |
| `make check` | Run `vet` + `test` (full quality gate) |
| `make docs` | Regenerate `doc/` API documentation via gomarkdoc |
| `make clean` | Remove `bin/`, `coverage.out`, `coverage.html` |

---

## Guidelines

### Code Quality

- Keep changes focused and minimal — avoid refactoring unrelated code
- Add tests for new functionality; target 80%+ coverage for new packages
- All packages must pass race detection (`make test-race`)
- Run `make check` before submitting a PR

### Critical Constraints

- **Transit accuracy** — Changes to `pkg/transit/transit.go` must preserve 247/247 event accuracy. The `solarfire_test.go` validation test is non-negotiable.
- **Thread safety** — All Swiss Ephemeris C calls must go through `pkg/sweph`, which serializes them via a global `sync.Mutex`. Never call the C library directly.
- **`pkg/solarsage` API stability** — This is the public high-level API. Avoid breaking changes to exported function signatures.
- **MCP tool API stability** — Tool names and input schemas in `pkg/mcp/server.go` are used by AI clients. Avoid renaming or removing tools.

### Ephemeris Data

- The `.se1` files in `third_party/swisseph/ephe/` are data files from the Swiss Ephemeris project
- They are licensed under AGPL-3.0 (or commercial license). See `third_party/swisseph/LICENSE`
- Do not commit ephemeris data files to git — add them to `.gitignore` or manage them separately in your fork

### Submitting Changes

1. Fork the repository and create a feature branch
2. Make your changes
3. Run `make check` and `make test-race`
4. Ensure the Solar Fire accuracy test passes: `go test ./pkg/transit/ -v -run SolarFire`
5. Submit a pull request with a clear description of the change

---

## Architecture

See [CLAUDE.md](CLAUDE.md) for a full architecture overview, package layer descriptions, and key constraints.
