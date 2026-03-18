# swisseph-mcp

High-precision astrology calculation engine exposed as a [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server. Built on the [Swiss Ephemeris](https://www.astro.com/swisseph/) library with sub-arcsecond accuracy.

## Features

- **Natal Charts** - Planet positions, house cusps (7 systems), angles, and aspects
- **Transit Detection** - All major transit types: Tr-Na, Tr-Tr, Tr-Sp, Tr-Sa, Sp-Na, Sp-Sp, Sa-Na
- **Secondary Progressions** - Day-for-a-year progressed positions and events
- **Solar Arc Directions** - Solar arc directed positions and events
- **Sign & House Ingress** - Detect when planets enter new signs or houses
- **Stations** - Retrograde and direct station detection
- **Void of Course Moon** - Automatic VOC detection with aspect context
- **Geocoding** - Location name to coordinates via OpenStreetMap Nominatim
- **CSV Export** - Solar Fire compatible output format
- **1-second precision** - Bisection algorithm for exact event timing

### Supported Bodies

Sun, Moon, Mercury, Venus, Mars, Jupiter, Saturn, Uranus, Neptune, Pluto, Chiron, North Node (True/Mean), South Node, Lilith (Mean/True)

**Special Points:** ASC, MC, DSC, IC, Vertex, East Point, Lot of Fortune, Lot of Spirit

### House Systems

Placidus, Koch, Equal, Whole Sign, Campanus, Regiomontanus, Porphyry

## Quick Start

### Prerequisites

- Go 1.21+
- GCC (for CGO / Swiss Ephemeris compilation)

### Build

```bash
git clone https://github.com/anthropic/swisseph-mcp.git
cd swisseph-mcp
make build
```

### Run as MCP Server

```bash
./bin/swisseph-mcp

# Or with a custom ephemeris path
SWISSEPH_EPHE_PATH=/path/to/ephe ./bin/swisseph-mcp
```

### Claude Desktop Integration

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "astrology": {
      "command": "/path/to/swisseph-mcp",
      "env": {
        "SWISSEPH_EPHE_PATH": "/path/to/ephe"
      }
    }
  }
}
```

## Use as a Go Library

The calculation packages can be imported directly:

```go
import (
    "github.com/anthropic/swisseph-mcp/pkg/chart"
    "github.com/anthropic/swisseph-mcp/pkg/models"
    "github.com/anthropic/swisseph-mcp/pkg/sweph"
    "github.com/anthropic/swisseph-mcp/pkg/transit"
)

func main() {
    // Initialize Swiss Ephemeris
    sweph.Init("/path/to/ephe")
    defer sweph.Close()

    // Calculate a natal chart
    planets := []models.PlanetID{
        models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
        models.PlanetVenus, models.PlanetMars,
    }
    info, _ := chart.CalcSingleChart(
        51.5074, -0.1278, 2451545.0,  // London, J2000.0
        planets, models.DefaultOrbConfig(), models.HousePlacidus,
    )

    // Search for transit events over 30 days
    events, _ := transit.CalcTransitEvents(transit.TransitCalcInput{
        NatalLat: 51.5074, NatalLon: -0.1278,
        NatalJD:  2451545.0, NatalPlanets: planets,
        TransitLat: 51.5074, TransitLon: -0.1278,
        StartJD: 2460676.5, EndJD: 2460706.5,
        TransitPlanets: planets,
        EventConfig:    models.DefaultEventConfig(),
        OrbConfigTransit: models.DefaultOrbConfig(),
        HouseSystem:    models.HousePlacidus,
    })
}
```

## MCP Tools

| Tool | Description |
|------|-------------|
| `geocode` | Location name to coordinates and timezone |
| `datetime_to_jd` | ISO 8601 datetime to Julian Day (UT/TT) |
| `jd_to_datetime` | Julian Day to ISO 8601 datetime |
| `calc_planet_position` | Single planet position at a given time |
| `calc_single_chart` | Full natal/event chart with positions, houses, and aspects |
| `calc_double_chart` | Synastry/transit double chart with cross-aspects |
| `calc_progressions` | Secondary progressed planet positions |
| `calc_solar_arc` | Solar arc directed planet positions |
| `calc_transit` | Full transit event search over a time range (JSON or CSV) |

### Example: Natal Chart

```json
{
  "jsonrpc": "2.0", "id": 1,
  "method": "tools/call",
  "params": {
    "name": "calc_single_chart",
    "arguments": {
      "latitude": 51.5074,
      "longitude": -0.1278,
      "jd_ut": 2451545.0,
      "house_system": "PLACIDUS"
    }
  }
}
```

### Example: Transit Search

```json
{
  "jsonrpc": "2.0", "id": 2,
  "method": "tools/call",
  "params": {
    "name": "calc_transit",
    "arguments": {
      "natal_latitude": 51.5074,
      "natal_longitude": -0.1278,
      "natal_jd_ut": 2451545.0,
      "transit_latitude": 51.5074,
      "transit_longitude": -0.1278,
      "start_jd_ut": 2460676.5,
      "end_jd_ut": 2460706.5,
      "format": "csv",
      "timezone": "Europe/London"
    }
  }
}
```

## Architecture

```
cmd/server/        MCP server entry point (JSON-RPC over stdio)
pkg/mcp/           MCP protocol handler
pkg/chart/         Chart calculations (positions, houses, aspects)
pkg/transit/       Transit event detection engine
pkg/progressions/  Secondary progressions & solar arc
pkg/models/        Core data types and constants
pkg/julian/        Julian Day conversions
pkg/geo/           Geocoding and timezone lookup
pkg/export/        CSV/JSON export
pkg/sweph/         Swiss Ephemeris C bindings (CGO)
internal/aspect/   Aspect calculation engine
```

## Performance

| Operation | Time | Throughput |
|-----------|------|------------|
| Planet position | 380ns | 2.6M/sec |
| Natal chart (10 planets) | 80μs | 12,400/sec |
| Double chart + cross-aspects | 347μs | 2,880/sec |
| 30-day transit scan (5 planets) | 764ms | - |
| 1-year transit scan (outer planets) | 2.1s | - |

Run `make bench` to reproduce.

## Accuracy

Validated against Solar Fire 9 with **100% exact event match** (247/247 events) over a 1-year transit period including all 7 chart type combinations.

## Docker

```bash
docker build -t swisseph-mcp .
docker run -i swisseph-mcp
```

## License

MIT
