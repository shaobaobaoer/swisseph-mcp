// Package swisseph is the root of the solarsage-mcp module.
//
// SolarSage is a comprehensive astrology calculation engine built on the
// Swiss Ephemeris. It exposes 40 MCP tools, 40 REST endpoints, and a clean
// Go library API spanning 38 packages.
//
// # Recommended Entry Point
//
// For most applications, import the high-level pkg/solarsage package:
//
//	import "github.com/shaobaobaoer/solarsage-mcp/pkg/solarsage"
//
//	solarsage.Init("/path/to/ephe")
//	defer solarsage.Close()
//
//	chart, _ := solarsage.NatalChart(51.5, -0.1, "2000-01-01T12:00:00Z")
//	phase, _ := solarsage.MoonPhase("2025-03-18T12:00:00Z")
//	score, _ := solarsage.Compatibility(lat1, lon1, dt1, lat2, lon2, dt2)
//	report, _ := solarsage.FullReport(lat, lon, datetime)
//
// # Package Overview
//
// Chart calculations:
//   - pkg/chart        — natal/event charts, double charts, house systems, aspects
//   - pkg/composite    — composite (midpoint) and Davison charts
//   - pkg/harmonic     — harmonic divisional charts (1-180)
//   - pkg/render       — chart wheel x/y coordinates for SVG/Canvas
//
// Predictive techniques:
//   - pkg/transit      — 7-type transit event detection (100% validated)
//   - pkg/progressions — secondary progressions and solar arc
//   - pkg/returns      — solar and lunar return charts with series support
//   - pkg/primary      — primary directions (Ptolemy semi-arc, Naibod)
//   - pkg/symbolic     — symbolic directions (1°/year, Naibod, profection)
//   - pkg/profection   — annual and monthly profections
//   - pkg/firdaria     — Firdaria planetary period system
//
// Traditional astrology:
//   - pkg/dignity      — essential dignities, mutual receptions, sect
//   - pkg/dispositor   — dispositorship chains and final dispositor
//   - pkg/lots         — 15+ Arabic lots with day/night reversal
//   - pkg/bounds       — Chaldean decans and Egyptian terms
//   - pkg/antiscia     — antiscia and contra-antiscia mirror points
//   - pkg/planetary    — Chaldean planetary hours
//   - pkg/heliacal     — heliacal risings and settings
//
// Pattern detection:
//   - pkg/fixedstars   — 50+ fixed star catalog with conjunction detection
//   - pkg/midpoint     — midpoint tree, 90-degree Cosmobiology dial
//   - internal/aspect  — aspect math and 7 geometric pattern types
//
// Vedic astrology:
//   - pkg/vedic        — sidereal charts, 5 ayanamsas, Nakshatras, Vimshottari Dasha
//   - pkg/divisional   — 16 Varga charts (D1-D60)
//   - pkg/ashtakavarga — Ashtakavarga bindu tables
//   - pkg/yoga         — Vedic yoga detection
//
// Analysis and reporting:
//   - pkg/synastry     — relationship compatibility scoring
//   - pkg/report       — comprehensive one-call natal report
//   - pkg/lunar        — lunar phases and eclipse detection
//
// Infrastructure:
//   - pkg/sweph        — thread-safe CGO bindings to Swiss Ephemeris
//   - pkg/models       — core types, constants, and zodiac helpers
//   - pkg/julian       — Julian Day / ISO 8601 conversions
//   - pkg/geo          — geocoding and timezone lookup
//   - pkg/export       — CSV/JSON export utilities
//
// # Server Interfaces
//
// For MCP server usage, build and run cmd/server (JSON-RPC over stdio).
// For REST API usage, build and run cmd/api (net/http, 40 POST endpoints).
// Both can be built with: make build && make build-api
package swisseph
