# SolarSage API Documentation

Auto-generated from Go source using [gomarkdoc](https://github.com/princjef/gomarkdoc).

Regenerate with:
```bash
gomarkdoc --output doc/pkg-<name>.md ./pkg/<name>/
```

Or start a local godoc server:
```bash
go install golang.org/x/pkgsite/cmd/pkgsite@latest
pkgsite -http=:6060
# Open http://localhost:6060/github.com/shaobaobaoer/solarsage-mcp
```

---

## High-Level API

| Package | Description |
|---------|-------------|
| [pkg/solarsage](pkg-solarsage.md) | **Recommended entry point.** High-level convenience API with ISO 8601 strings and sensible defaults. |

## Server Interfaces

| Package | Description |
|---------|-------------|
| [pkg/mcp](pkg-mcp.md) | MCP protocol server (40 tools, JSON-RPC over stdio) |
| [pkg/api](pkg-api.md) | RESTful HTTP API server (40 endpoints, JSON, CORS, API key auth) |

## Chart Calculations

| Package | Description |
|---------|-------------|
| [pkg/chart](pkg-chart.md) | Natal/event charts, double charts, house systems, aspects |
| [pkg/composite](pkg-composite.md) | Composite (midpoint) and Davison relationship charts |
| [pkg/harmonic](pkg-harmonic.md) | Harmonic divisional charts (1-180) |
| [pkg/render](pkg-render.md) | Chart wheel x/y coordinates for SVG/Canvas |

## Predictive Techniques

| Package | Description |
|---------|-------------|
| [pkg/transit](pkg-transit.md) | Transit event detection engine (7 types, 100% validated) |
| [pkg/progressions](pkg-progressions.md) | Secondary progressions and solar arc directions |
| [pkg/returns](pkg-returns.md) | Solar and lunar return charts with series support |
| [pkg/primary](pkg-primary.md) | Primary directions (Ptolemy semi-arc, Naibod key) |
| [pkg/symbolic](pkg-symbolic.md) | Symbolic directions (1 deg/year, Naibod, Profection, custom) |
| [pkg/profection](pkg-profection.md) | Annual and monthly profections |
| [pkg/firdaria](pkg-firdaria.md) | Firdaria planetary period system (day/night sequences) |

## Traditional Astrology

| Package | Description |
|---------|-------------|
| [pkg/dignity](pkg-dignity.md) | Essential dignities, mutual receptions, sect, bonification |
| [pkg/dispositor](pkg-dispositor.md) | Dispositorship chains and final dispositor |
| [pkg/lots](pkg-lots.md) | 15+ Arabic lots with day/night reversal |
| [pkg/bounds](pkg-bounds.md) | Chaldean decans and Egyptian/Ptolemaic terms |
| [pkg/antiscia](pkg-antiscia.md) | Antiscia and contra-antiscia mirror points |
| [pkg/planetary](pkg-planetary.md) | Chaldean planetary hours |
| [pkg/heliacal](pkg-heliacal.md) | Heliacal risings and settings |

## Pattern Detection

| Package | Description |
|---------|-------------|
| [pkg/fixedstars](pkg-fixedstars.md) | 50+ fixed star catalog with conjunction detection |
| [pkg/midpoint](pkg-midpoint.md) | Midpoint tree, 90-degree Cosmobiology dial, activations |
| [internal/aspect](internal-aspect.md) | Aspect calculation engine and 7 geometric pattern types |

## Vedic Astrology

| Package | Description |
|---------|-------------|
| [pkg/vedic](pkg-vedic.md) | Sidereal charts, 5 ayanamsas, Nakshatras, Vimshottari Dasha |
| [pkg/divisional](pkg-divisional.md) | 16 Varga charts (D1-D60) |
| [pkg/ashtakavarga](pkg-ashtakavarga.md) | Ashtakavarga bindu tables and Sarvashtakavarga |
| [pkg/yoga](pkg-yoga.md) | Vedic yoga detection (Mahapurusha, Raja, Dhana, etc.) |

## Analysis

| Package | Description |
|---------|-------------|
| [pkg/synastry](pkg-synastry.md) | Relationship compatibility scoring |
| [pkg/report](pkg-report.md) | Comprehensive one-call natal report |
| [pkg/lunar](pkg-lunar.md) | Lunar phases and eclipse detection |

## Infrastructure

| Package | Description |
|---------|-------------|
| [pkg/sweph](pkg-sweph.md) | Swiss Ephemeris CGO bindings (thread-safe, global mutex) |
| [pkg/models](pkg-models.md) | Core types, constants, and zodiac helpers |
| [pkg/julian](pkg-julian.md) | Julian Day / ISO 8601 conversions |
| [pkg/geo](pkg-geo.md) | Geocoding and timezone lookup |
| [pkg/export](pkg-export.md) | CSV/JSON export utilities |
