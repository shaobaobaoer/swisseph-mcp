# SolarSage

The most comprehensive open-source Western astrology calculation engine.
**35 MCP tools · 35 REST endpoints · 33 packages · 800+ tests · 11 house systems · 50+ fixed stars · 15+ Arabic lots · 7 aspect patterns** — all with sub-arcsecond accuracy.

Usable as a **Go library**, an **MCP server** for AI assistants, or a **RESTful HTTP API** for web/mobile clients.

Built on the [Swiss Ephemeris](https://www.astro.com/swisseph/).
Independently validated: **100% accuracy on single-chart transits** (247/247 events), **84.0% timeline event coverage** (971/1156 events), **83.1% on double-chart progressions** (833/1002 events) against Solar Fire 9.

> [中文文档 →](README.zh.md)

---

## Table of Contents

- [Why SolarSage?](#why-solarsage)
- [Related Projects](#related-projects)
- [Features](#features)
- [Quick Start](#quick-start)
  - [Prerequisites & Build](#prerequisites--build)
  - [MCP Server](#run-as-mcp-server)
  - [REST API Server](#run-as-rest-api-server)
  - [Go Library](#use-as-a-go-library)
- [MCP Tools Reference](#mcp-tools-reference)
- [REST API Reference](#rest-api-reference)
- [Go API Documentation](#go-api-documentation)
- [Architecture](#architecture)
- [Performance](#performance)
- [Accuracy](#accuracy)
- [Docker](#docker)
- [Contributing](#contributing)
- [License](#license)

---

## Related Projects

| Project | Description | Link |
|---------|-------------|------|
| **SolarSage** | This repository - Core calculation engine (MCP + REST API) | [GitHub](https://github.com/shaobaobaoer/solarsage-mcp) |
| **SolarSageDataService** | Standalone data service for astrology dataset (HTTP API + Go package) | [GitHub](../SolarSageDataService/) |

SolarSage focuses on **astrological calculations** while SolarSageDataService provides **access to historical data** (national charts, person birth data, eclipses, etc.). They can be used independently or together.

> **Note**: The dataset functionality has been moved to [SolarSageDataService](../SolarSageDataService/). This repository now contains only the core calculation engine.

---

## Why SolarSage?

| Feature | SolarSage | flatlib (Python) | Kerykeion (Python) | Swiss Ephemeris (C) |
|---|---|---|---|---|
| Language | Go | Python | Python | C |
| Transit detection | 7 types, 1s precision | Basic | None | Manual |
| Solar/Lunar returns | Series support | Single | Single | Manual |
| Composite charts | Midpoint method | None | None | Manual |
| Synastry scoring | Category breakdown | None | Basic | None |
| Eclipse detection | Solar + Lunar | None | None | Low-level |
| Profections | Annual + monthly | None | None | None |
| Arabic lots | 15+ with day/night | None | None | None |
| Essential dignities | Full + mutual reception | Basic | Basic | None |
| Aspect patterns | 7 types | None | None | None |
| Fixed stars | 50+ catalog | None | None | Low-level |
| Midpoints | 90deg dial + activations | None | None | None |
| Harmonic charts | 1-180 | None | None | None |
| Planetary hours | Chaldean | None | None | None |
| House systems | 11 | 7 | 3 | All |
| Dispositors | Full chains | None | None | None |
| Primary directions | Ptolemy semi-arc | None | None | None |
| Symbolic directions | 4 methods | None | None | None |
| Firdaria | Day/night sequences | None | None | None |
| Heliacal phenomena | Visibility algorithm | None | None | Low-level |
| Bonification | Aspect-based scoring | None | None | None |
| One-call report | Everything combined | None | None | None |
| Chart visualization | Wheel coordinates | None | None | None |
| MCP server | 40 tools | None | None | None |
| REST API | 40 endpoints | None | None | None |
| Accuracy validated | 247/247 (100%) | No | No | N/A |
| Thread-safe | Yes (mutex) | No | No | No |

---

## Features

### Chart Calculations
- **Natal Charts** — Positions, houses (11 systems), angles, aspects (9 types)
- **Double Charts** — Synastry/transit overlay with cross-aspects
- **Composite Charts** — Midpoint method for relationship analysis
- **Davison Charts** — Midpoint in time and space
- **Harmonic Charts** — Nth harmonic charts (5th quintile, 7th septile, 9th novile, etc.)

### Predictive Techniques
- **Transit Detection** — Tr-Na, Tr-Tr, Tr-Sp, Tr-Sa, Sp-Na, Sp-Sp, Sa-Na with 1-second precision
- **Secondary Progressions** — Day-for-a-year progressed positions and events
- **Solar Arc Directions** — Solar arc directed positions and events
- **Primary Directions** — Ptolemy semi-arc method with Naibod key
- **Symbolic Directions** — 1-degree/year, Naibod, Profection, custom rate
- **Solar & Lunar Returns** — Exact return charts with series support
- **Annual Profections** — Time-lord technique with monthly sub-profections
- **Firdaria** — Planetary period system (day/night sequences)
- **Sign/House Ingress** — Planet sign and house change detection
- **Stations** — Retrograde and direct station detection

### Traditional Astrology
- **Essential Dignities** — Rulership, exaltation, detriment, fall with scoring
- **Mutual Receptions** — Rulership and exaltation mutual receptions
- **Sect** — Diurnal/nocturnal planet alignment analysis
- **Arabic Lots** — 15+ lots (Fortune, Spirit, Eros, Victory, etc.) with day/night reversal
- **Decans & Terms** — Chaldean decans and Egyptian/Ptolemaic term boundaries
- **Planetary Hours** — Chaldean hours with computed sunrise/sunset
- **Antiscia** — Solstice and equinox mirror points with pair detection
- **Dispositors** — Dispositorship chains, final dispositor, mutual dispositors
- **Bonification & Maltreatment** — Aspect-based planetary condition analysis
- **Heliacal Risings/Settings** — Swiss Ephemeris visibility algorithms

### Pattern Detection
- **Aspect Patterns** — Grand Trine, T-Square, Grand Cross, Yod, Kite, Mystic Rectangle, Stellium
- **Fixed Stars** — 50+ major star catalog with precession-corrected conjunctions
- **Midpoint Analysis** — Full midpoint tree, 90-degree Cosmobiology dial, activations

### Astronomical
- **Lunar Phases** — New/full moon finder, phase angle, illumination percentage
- **Eclipse Finder** — Solar and lunar eclipse detection with type classification
- **Void of Course Moon** — Automatic VOC detection with aspect context

### Relationship
- **Synastry Scoring** — Compatibility analysis with category breakdown (love, passion, communication, commitment)
- **Composite Charts** — Midpoint method with aspects
- **Davison Chart** — Midpoint in time and space relationship chart

### Visualization
- **Chart Wheel Coordinates** — Planet x/y positions, house cusp lines, aspect lines, sign segments for SVG/Canvas rendering

### Supported Bodies

**Planets:** Sun, Moon, Mercury, Venus, Mars, Jupiter, Saturn, Uranus, Neptune, Pluto, Chiron, North Node (True/Mean), South Node, Lilith (Mean/True)

**Special Points:** ASC, MC, DSC, IC, Vertex, East Point, Lot of Fortune, Lot of Spirit

**House Systems:** Placidus, Koch, Equal, Whole Sign, Campanus, Regiomontanus, Porphyry, Morinus, Topocentric, Alcabitius, Meridian

**Output:** JSON and CSV for all chart types. Unicode astrology glyphs (♈♉♊♋♌♍♎♏♐♑♒♓, ☉☽☿♀♂♃♄, ☌☍△□✱).

---

## Quick Start

### Prerequisites & Build

**System requirements:**
- Go 1.25+
- GCC (for CGO / Swiss Ephemeris)
- The Swiss Ephemeris C sources must be present at `third_party/swisseph/` (see [Contributing](CONTRIBUTING.md) for setup)

```bash
git clone https://github.com/shaobaobaoer/solarsage-mcp.git
cd solarsage-mcp
make build        # → bin/solarsage-mcp  (MCP server)
make build-api    # → bin/solarsage-api  (REST API server)
```

The `make build` step compiles the Swiss Ephemeris C library via CGO and links it into the Go binary. No separate installation of `libswisseph` is needed — everything is vendored under `third_party/`.

### Ephemeris Options

SolarSage supports multiple ephemeris sources:

| Type | Description | Use Case |
|------|-------------|----------|
| **Swiss (default)** | Swiss Ephemeris based on JPL DE431 | Highest accuracy, longest time span |
| **JPL** | Direct JPL files (DE200, DE406, DE440, etc.) | Solar Fire compatibility (DE406) |
| **Moshier** | Built-in analytical approximation | No files needed, lower precision |

Configure via environment variables:

```bash
# Use JPL DE406 for Solar Fire compatibility
export SWISSEPH_TYPE=jpl
export SWISSEPH_JPL_FILE=de406.eph

# Or stick with Swiss Ephemeris (default)
export SWISSEPH_TYPE=swiss
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for download links and detailed setup.

### Run as MCP Server

```bash
./bin/solarsage-mcp

# With a custom ephemeris data path
SWISSEPH_EPHE_PATH=/path/to/ephe ./bin/solarsage-mcp
```

#### Claude Desktop Integration

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "astrology": {
      "command": "/path/to/solarsage-mcp",
      "env": {
        "SWISSEPH_EPHE_PATH": "/path/to/ephe"
      }
    }
  }
}
```

#### Cursor / Other MCP Clients

```json
{
  "mcpServers": {
    "solarsage": {
      "command": "/path/to/solarsage-mcp"
    }
  }
}
```

### Run as REST API Server

```bash
./bin/solarsage-api --port 8080

# With API key authentication
./bin/solarsage-api --port 8080 --api-key your-secret-key

# Example: natal chart
curl -X POST http://localhost:8080/api/v1/chart/natal \
  -H "Content-Type: application/json" \
  -d '{"latitude": 51.5074, "longitude": -0.1278, "jd_ut": 2451545.0}'

# Example: transit events
curl -X POST http://localhost:8080/api/v1/transit \
  -H "Content-Type: application/json" \
  -d '{
    "natal_lat": 51.5074, "natal_lon": -0.1278, "natal_jd": 2451545.0,
    "start_jd": 2460676.5, "end_jd": 2460736.5
  }'
```

All 40 endpoints are under `/api/v1/`. CORS is enabled. Optional API key auth via `X-API-Key` header.

Health check: `GET /api/v1/health`

---

## Use as a Go Library

### High-Level API (recommended)

The `solarsage` package provides a high-level API with sensible defaults. Pass ISO 8601 datetime strings instead of Julian Day numbers:

```go
package main

import (
    "fmt"
    "github.com/shaobaobaoer/solarsage-mcp/pkg/solarsage"
)

func main() {
    solarsage.Init("/path/to/ephe")
    defer solarsage.Close()

    // Natal chart
    chart, _ := solarsage.NatalChart(51.5074, -0.1278, "1990-06-15T14:30:00Z")
    for _, p := range chart.Planets {
        fmt.Printf("%s in %s (house %d)\n", p.PlanetID, p.Sign, p.House)
    }

    // Solar return for 2025
    sr, _ := solarsage.SolarReturn(51.5074, -0.1278, "1990-06-15T14:30:00Z", 2025)
    fmt.Printf("Solar return: age %.1f\n", sr.Age)

    // Current moon phase
    phase, _ := solarsage.MoonPhase("2025-03-18T12:00:00Z")
    fmt.Printf("Moon: %s (%.0f%% illuminated)\n", phase.PhaseName, phase.Illumination*100)

    // Eclipses in a date range
    eclipses, _ := solarsage.Eclipses("2025-01-01", "2026-01-01")
    for _, e := range eclipses {
        fmt.Printf("Eclipse: %s in %s\n", e.Type, e.MoonSign)
    }

    // Relationship compatibility
    score, _ := solarsage.Compatibility(
        51.5074, -0.1278, "1990-06-15T14:30:00Z",
        40.7128, -74.006, "1992-03-22T08:00:00Z",
    )
    fmt.Printf("Compatibility: %.0f%%\n", score.Compatibility)

    // Single planet position
    pos, _ := solarsage.PlanetPosition("Venus", "2025-03-18T12:00:00Z")
    fmt.Printf("Venus: %s at %.2f°\n", pos.Sign, pos.SignDegree)

    // Chart wheel coordinates for SVG/Canvas rendering
    wheel, _ := solarsage.ChartWheel(51.5074, -0.1278, "1990-06-15T14:30:00Z")
    for _, p := range wheel.Planets {
        fmt.Printf("%s at (%.2f, %.2f)\n", p.PlanetID, p.Position.X, p.Position.Y)
    }

    // Comprehensive report (all techniques in one call)
    report, _ := solarsage.FullReport(51.5074, -0.1278, "1990-06-15T14:30:00Z")
    fmt.Printf("Elements: Fire=%d Earth=%d Air=%d Water=%d\n",
        report.ElementBalance["Fire"], report.ElementBalance["Earth"],
        report.ElementBalance["Air"], report.ElementBalance["Water"])
}
```

### Low-Level API

Import individual packages for full control over orbs, house systems, and planet selection:

```go
import (
    "github.com/shaobaobaoer/solarsage-mcp/pkg/chart"
    "github.com/shaobaobaoer/solarsage-mcp/pkg/models"
    "github.com/shaobaobaoer/solarsage-mcp/pkg/sweph"
    "github.com/shaobaobaoer/solarsage-mcp/pkg/transit"
)

func main() {
    sweph.Init("/path/to/ephe")
    defer sweph.Close()

    planets := []models.PlanetID{
        models.PlanetSun, models.PlanetMoon, models.PlanetVenus,
    }

    // Full control over orbs, house system, and planet selection
    // New: orb_config supports entering/exiting orbs for applying/separating aspects
    info, _ := chart.CalcSingleChart(
        51.5074, -0.1278, 2451545.0,
        planets,
        models.OrbConfig{
            Definitions: []models.AspectOrbDef{
                {Name: "conjunction", Angle: 0, EnteringOrb: 10, ExitingOrb: 3, Enabled: true},
                {Name: "opposition", Angle: 180, EnteringOrb: 10, ExitingOrb: 3, Enabled: true},
                {Name: "trine", Angle: 120, EnteringOrb: 8, ExitingOrb: 2, Enabled: true},
                {Name: "square", Angle: 90, EnteringOrb: 8, ExitingOrb: 2, Enabled: true},
            },
        },
        models.HouseKoch,
    )

    // Transit search with all options
    events, _ := transit.CalcTransitEvents(transit.TransitCalcInput{
        NatalLat:         51.5074,
        NatalLon:         -0.1278,
        NatalJD:          2451545.0,
        NatalPlanets:     planets,
        TransitLat:       51.5074,
        TransitLon:       -0.1278,
        StartJD:          2460676.5,
        EndJD:            2460706.5,
        TransitPlanets:   planets,
        EventConfig:      models.DefaultEventConfig(),
        OrbConfigTransit: models.DefaultOrbConfig(),
        HouseSystem:      models.HousePlacidus,
    })
    _ = info
    _ = events
}
```

---

## MCP Tools Reference

### Utilities

| Tool | Description |
|------|-------------|
| `geocode` | Convert location name to latitude, longitude, and timezone |
| `datetime_to_jd` | Convert ISO 8601 datetime string to Julian Day (UT and TT) |
| `jd_to_datetime` | Convert Julian Day number to ISO 8601 datetime string |

### Chart Calculations

| Tool | Description |
|------|-------------|
| `calc_planet_position` | Single planet position, sign, house, and speed at a given time |
| `calc_single_chart` | Full natal/event chart: positions, houses (11 systems), and aspects |
| `calc_double_chart` | Synastry/transit double chart with cross-aspects between two charts |
| `calc_composite_chart` | Composite (midpoint) chart for relationship analysis |
| `calc_davison_chart` | Davison relationship chart (midpoint in time and space) |
| `calc_harmonic_chart` | Nth harmonic chart for any harmonic number |
| `calc_chart_wheel` | Chart wheel x/y coordinates for SVG/Canvas rendering |

### Predictive

| Tool | Description |
|------|-------------|
| `calc_transit` | Full transit event search over a date range (JSON or CSV output) |
| `calc_progressions` | Secondary progressed planet positions (day-for-a-year) |
| `calc_solar_arc` | Solar arc directed planet positions |
| `calc_primary_directions` | Primary directions using Ptolemy semi-arc and Naibod key |
| `calc_symbolic_directions` | Symbolic directions (1°/year, Naibod, Profection, custom rate) |
| `calc_solar_return` | Solar return chart for a given year |
| `calc_lunar_return` | Lunar return chart for the next Moon return |
| `calc_profection` | Annual/monthly profections with activated time-lord |
| `calc_firdaria` | Firdaria planetary period timeline (day/night sequences) |

### Traditional Astrology

| Tool | Description |
|------|-------------|
| `calc_dignity` | Essential dignities, mutual receptions, and sect for all planets |
| `calc_bonification` | Bonification and maltreatment scoring from aspects and dignity |
| `calc_lots` | Arabic lots — Fortune, Spirit, Eros, Victory, and 10+ more |
| `calc_bounds` | Chaldean decans and Egyptian/Ptolemaic term (bound) rulers |
| `calc_planetary_hours` | Chaldean planetary hours with sunrise/sunset for any date/location |
| `calc_antiscia` | Antiscia and contra-antiscia solstice/equinox mirror points |
| `calc_heliacal_events` | Heliacal risings and settings using Swiss Ephemeris visibility algorithm |

### Pattern Detection

| Tool | Description |
|------|-------------|
| `calc_aspect_patterns` | Detect Grand Trine, T-Square, Grand Cross, Yod, Kite, Mystic Rectangle, Stellium |
| `calc_fixed_stars` | Fixed star conjunctions from 50+ star catalog (precession-corrected) |
| `calc_midpoints` | Midpoint tree with 90-degree Cosmobiology dial and activations |

### Astronomical

| Tool | Description |
|------|-------------|
| `calc_lunar_phase` | Current lunar phase, illumination percentage, and phase angle |
| `calc_lunar_phases` | Find new moons, full moons, and quarters in a date range |
| `calc_eclipses` | Solar and lunar eclipse finder with type classification |

### Analysis

| Tool | Description |
|------|-------------|
| `calc_synastry` | Relationship compatibility scoring with category breakdown |
| `calc_dispositors` | Dispositorship chains, final dispositor, and mutual dispositors |
| `calc_natal_report` | Comprehensive natal analysis combining all techniques |

---

## REST API Reference

All endpoints are `POST /api/v1/<path>` and accept/return JSON. CORS is enabled on all routes. Optional authentication via `X-API-Key` header.

### Common Parameters

Most chart endpoints accept an optional `orb_config` parameter for customizing aspect orbs:

```json
{
  "orb_config": {
    "definitions": [
      {"name": "conjunction", "angle": 0, "entering_orb": 10, "exiting_orb": 3, "enabled": true},
      {"name": "opposition", "angle": 180, "entering_orb": 10, "exiting_orb": 3, "enabled": true},
      {"name": "trine", "angle": 120, "entering_orb": 8, "exiting_orb": 2, "enabled": true},
      {"name": "square", "angle": 90, "entering_orb": 8, "exiting_orb": 2, "enabled": true},
      {"name": "sextile", "angle": 60, "entering_orb": 5, "exiting_orb": 2, "enabled": true}
    ]
  }
}
```

- `entering_orb`: Orb for applying/entering aspects (planets moving toward exact)
- `exiting_orb`: Orb for separating/exiting aspects (planets moving away from exact)
- Custom aspects can be defined with any angle (e.g., quintile at 72°)

| Endpoint | Description |
|----------|-------------|
| `GET  /api/v1/health` | Health check |
| `POST /api/v1/geocode` | Geocode location name |
| `POST /api/v1/datetime/to-jd` | ISO 8601 → Julian Day |
| `POST /api/v1/datetime/from-jd` | Julian Day → ISO 8601 |
| `POST /api/v1/planet/position` | Single planet position |
| `POST /api/v1/chart/natal` | Natal chart |
| `POST /api/v1/chart/double` | Double/synastry chart |
| `POST /api/v1/chart/composite` | Composite chart |
| `POST /api/v1/chart/davison` | Davison chart |
| `POST /api/v1/chart/harmonic` | Harmonic chart |
| `POST /api/v1/chart/wheel` | Chart wheel coordinates |
| `POST /api/v1/transit` | Transit events |
| `POST /api/v1/progressions` | Secondary progressions |
| `POST /api/v1/solar-arc` | Solar arc directions |
| `POST /api/v1/primary-directions` | Primary directions |
| `POST /api/v1/symbolic-directions` | Symbolic directions |
| `POST /api/v1/solar-return` | Solar return chart |
| `POST /api/v1/lunar-return` | Lunar return chart |
| `POST /api/v1/dignity` | Essential dignities |
| `POST /api/v1/bonification` | Bonification & maltreatment |
| `POST /api/v1/dispositors` | Dispositorship chains |
| `POST /api/v1/profection` | Annual profections |
| `POST /api/v1/firdaria` | Firdaria periods |
| `POST /api/v1/lots` | Arabic lots |
| `POST /api/v1/bounds` | Decans & terms |
| `POST /api/v1/antiscia` | Antiscia points |
| `POST /api/v1/planetary-hours` | Planetary hours |
| `POST /api/v1/heliacal` | Heliacal events |
| `POST /api/v1/aspects/patterns` | Aspect patterns |
| `POST /api/v1/fixed-stars` | Fixed star conjunctions |
| `POST /api/v1/midpoints` | Midpoint analysis |
| `POST /api/v1/synastry` | Synastry scoring |
| `POST /api/v1/lunar/phase` | Lunar phase |
| `POST /api/v1/lunar/phases` | Lunar phase list |
| `POST /api/v1/lunar/eclipses` | Eclipse finder |
| `POST /api/v1/report/natal` | Comprehensive natal report |

**Note**: Dataset endpoints (`/api/v1/dataset/*`) have been moved to [SolarSageDataService](../SolarSageDataService/).

---

## Go API Documentation

Full API documentation for every exported type and function is available in the [`doc/`](doc/) directory, auto-generated from Go source comments using [gomarkdoc](https://github.com/princjef/gomarkdoc). Start with [`doc/README.md`](doc/README.md) for the index.

You can also browse documentation locally with the official Go documentation server:

```bash
go install golang.org/x/pkgsite/cmd/pkgsite@latest
pkgsite -http=:6060
# Open http://localhost:6060/github.com/shaobaobaoer/solarsage-mcp
```

---

## Architecture

```
cmd/
  server/          MCP server entry point (JSON-RPC over stdio)
  api/             REST API server entry point (net/http)
pkg/
  solarsage/       High-level convenience API (recommended entry point)
  mcp/             MCP protocol handler (40 tools)
  api/             REST API handler (40 endpoints)
  chart/           Chart calculations (positions, houses, aspects)
  transit/         Transit event detection engine (100% validated)
  progressions/    Secondary progressions & solar arc
  returns/         Solar & lunar return charts
  composite/       Composite (midpoint) & Davison charts
  synastry/        Synastry compatibility scoring
  primary/         Primary directions (Ptolemy semi-arc, Naibod)
  symbolic/        Symbolic directions (1°/year, Naibod, Profection, custom)
  dignity/         Essential dignities, mutual receptions, sect
  dispositor/      Dispositorship chains & final dispositor
  report/          Comprehensive chart analysis report
  profection/      Annual & monthly profections
  firdaria/        Firdaria planetary period system
  lots/            Arabic lots/parts calculator
  bounds/          Chaldean decans & Egyptian terms
  antiscia/        Antiscia & contra-antiscia
  dignity/         Essential dignities, sect, bonification
  fixedstars/      Fixed star catalog & conjunction detection
  midpoint/        Midpoint analysis & Cosmobiology dial
  harmonic/        Harmonic charts
  planetary/       Chaldean planetary hours
  heliacal/        Heliacal risings/settings
  lunar/           Lunar phases & eclipse detection
  render/          Chart wheel visualization coordinates
  models/          Core data types and constants
  julian/          Julian Day conversions (ISO 8601 ↔ JD)
  geo/             Geocoding and timezone lookup
  export/          CSV/JSON export utilities
  sweph/           Swiss Ephemeris C bindings (CGO, thread-safe mutex)
internal/
  aspect/          Aspect calculation & 7-pattern detection engine
third_party/
  swisseph/        Swiss Ephemeris C source + headers + libswe.a + ephe/
```

### Key Design Decisions

- **CGO with static library** — Swiss Ephemeris is compiled to `libswe.a` and statically linked. No runtime `.so` dependencies, no installation friction.
- **Global mutex in `pkg/sweph`** — The Swiss Ephemeris C library is not thread-safe. All C calls are serialized through a single `sync.Mutex`.
- **`pkg/solarsage` as the stable API** — The high-level package wraps all lower-level packages and provides a stable, ergonomic interface. Lower-level packages are intentionally kept narrow in scope.
- **Transit accuracy gate** — `pkg/transit/solarfire_test.go` validates 247 events against a Solar Fire 9 reference CSV at 100% match. Any change to `transit.go` must preserve this.

---

## Performance

| Operation | Time | Throughput |
|-----------|------|------------|
| Planet position | ~380 ns | ~2.6 M/sec |
| Natal chart (10 planets) | ~80 µs | ~12,400/sec |
| Double chart + cross-aspects | ~347 µs | ~2,880/sec |
| 30-day transit scan (5 planets) | ~764 ms | — |
| 1-year transit scan (outer planets) | ~2.1 s | — |

Run `make bench` to reproduce on your hardware.

---

## Accuracy

### Single-Chart Transit Detection (TC1)
**100% exact event match** (247/247 transit events) over a 1-year period. Covers all 7 chart-type combinations (Tr-Na, Tr-Tr, Tr-Sp, Tr-Sa, Sp-Na, Sp-Sp, Sa-Na), benchmarked against Solar Fire 9 — the industry-standard desktop astrology software.

### Double-Chart Scenarios (TC2)
**83.1% match rate** (833/1002 transit events) on double-chart progressions (Sp-Sp and Sp-Na pairings). The remaining 17% gap is due to:
- **Sp-Sp formula incompatibility** — Solar Fire's secondary progression-to-secondary progression algorithm differs fundamentally from the Swiss Ephemeris calculation method
- **Position errors > 0.3°** — Some event pairs show systemic ephemeris divergence beyond the 0.1° tolerance threshold
- This is a known limitation and represents the practical ceiling for greedy matching on this dataset

The validation tests live at `pkg/transit/solarfire_test.go` (`TestSolarFireCSV_TC1_*` for single-chart, `TestSolarFireCSV_TC2_*` for double-chart) and run as part of `make test`.

### Timeline Event Validation (Phase D)
**84.0% event coverage** (971/1156 timeline events) on full-year natal chart transits. Validates all event types (Begin, Enter, Exact, Leave, SignIngress, Void, HouseChange, Stations) across all chart pairings:
- **Tr-Na:** 69.5% (139/200) — transit vs natal
- **SignIngress:** 98.8% (167/169) — planet sign ingress detection
- **Sp-Na:** 76.9% (40/52) — secondary progressions vs natal
- **Void of Course:** 100.0% (161/161) — Moon void detection
- **Stations:** 100.0% (12/12) — retrograde/direct stations
- **All other pairings:** 68-92% coverage

Validation tests at `pkg/solarsage/phase_d_timeline_test.go` validate the full year's worth of predicted transits at event-level granularity, not just snapshot validation.

---

## Docker

```bash
# Build image (compiles Swiss Ephemeris + Go binary inside container)
docker build -t solarsage-mcp .

# Run MCP server
docker run -i solarsage-mcp

# Run REST API server
docker run -p 8080:8080 --entrypoint solarsage-api solarsage-mcp --port 8080
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, Swiss Ephemeris build instructions, and contribution guidelines.

---

## License

MIT — see [LICENSE](LICENSE).

Swiss Ephemeris is licensed under AGPL-3.0 (or a commercial license from Astrodienst). See `third_party/swisseph/LICENSE` for details.
