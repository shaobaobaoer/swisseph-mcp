package solarsage

import (
	"math"
	"testing"

	"github.com/shaobaobaoer/solarsage-mcp/pkg/chart"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/lunar"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/progressions"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/returns"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/sweph"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/transit"
)

// =============================================================================
// JN Precision Test Suite
//
// Reference: JN (Male, b. 1997-12-18 09:36 UTC, Jinshan China, 30.9°N 121.15°E)
// This test covers: NA (Natal), TR (Transit), SP (Secondary Progressions),
// SR (Solar Return), Moon Phase, and Double Chart (Biwheel).
//
// All tests use validated Solar Fire reference data and run in <1 second total.
// =============================================================================

const (
	jnJDUT    = 2450800.900009  // 1997-12-18 09:36 UTC
	jnLat     = 30.9            // Jinshan, China
	jnLon     = 121.15
	transitJD = 2461041.5       // 2026-01-01 00:00 UTC (approximately)
	tolLon    = 0.01            // longitude tolerance (degrees)
	tolAng    = 0.01            // angle/cusp tolerance (degrees)
)

var jnPlanets = []models.PlanetID{
	models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
	models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
	models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune, models.PlanetPluto,
}

// =============================================================================
// TestJN_NA: Natal Chart Precision
//
// Verifies JN natal chart against Solar Fire validated values.
// Known values:
//   Sun:     266.500° (Sagittarius 26.5°), house 6
//   Moon:    138.116° (Leo 18.1°), house 2
//   Mars:    300.097° (Aquarius 0.1°), house 8
//   Saturn:  13.538° (Aries 13.5°), house 10
//   ASC:     96.530°
//   MC:      351.500°
// =============================================================================
func TestJN_NA(t *testing.T) {
	orbs := models.DefaultOrbConfig()
	info, err := chart.CalcSingleChart(jnLat, jnLon, jnJDUT, jnPlanets, orbs, models.HousePlacidus)
	if err != nil {
		t.Fatalf("CalcSingleChart: %v", err)
	}

	// Spot-check 4 Solar Fire validated planet positions
	wantLon := map[models.PlanetID]float64{
		models.PlanetSun:     266.500,
		models.PlanetMoon:    138.116,
		models.PlanetMars:    300.097,
		models.PlanetSaturn:  13.538,
	}

	gotMap := make(map[models.PlanetID]float64)
	for _, p := range info.Planets {
		gotMap[p.PlanetID] = p.Longitude
	}

	for pid, want := range wantLon {
		got, ok := gotMap[pid]
		if !ok {
			t.Errorf("NA: planet %s not found", pid)
			continue
		}
		if math.Abs(got-want) > tolLon {
			t.Errorf("NA %s lon: got %.4f, want %.3f (diff %.4f)", pid, got, want, got-want)
		}
	}

	// Verify angles
	if math.Abs(info.Angles.ASC-96.530) > tolAng {
		t.Errorf("ASC: got %.4f, want 96.530", info.Angles.ASC)
	}
	if math.Abs(info.Angles.MC-351.500) > tolAng {
		t.Errorf("MC: got %.4f, want 351.500", info.Angles.MC)
	}

	t.Logf("NA: Sun=%.4f Moon=%.4f ASC=%.4f MC=%.4f", gotMap[models.PlanetSun], gotMap[models.PlanetMoon], info.Angles.ASC, info.Angles.MC)
}

// =============================================================================
// TestJN_SP: Secondary Progressions
//
// Verifies progressed planet positions at transit date 2026-01-01.
// Age ~28.03 years → progressed offset ~28.03 days after natal.
// Expected SP Sun ≈ 293-296° (Capricorn 23-26°), progressed from natal 266.5°.
// =============================================================================
func TestJN_SP(t *testing.T) {
	spPlanets := []models.PlanetID{models.PlanetSun, models.PlanetMoon, models.PlanetMercury}

	// Calculate progressed positions for each planet
	for _, pid := range spPlanets {
		lon, speed, err := progressions.CalcProgressedLongitude(pid, jnJDUT, transitJD)
		if err != nil {
			t.Fatalf("SP %s: %v", pid, err)
		}
		if lon < 0 || lon >= 360 {
			t.Errorf("SP %s lon out of range: %.4f", pid, lon)
		}
		if speed <= 0 {
			t.Errorf("SP %s speed non-positive: %.6f", pid, speed)
		}
		t.Logf("SP %s: lon=%.4f speed=%.6f", pid, lon, speed)
	}

	// SP Sun correctness check: after ~28 days, Sun should have advanced ~28° from natal
	// Natal Sun = 266.5°, expected SP Sun ≈ 293-296° (Capricorn 23-26°)
	spSun, _, err := progressions.CalcProgressedLongitude(models.PlanetSun, jnJDUT, transitJD)
	if err != nil {
		t.Fatalf("SP Sun: %v", err)
	}
	if spSun < 290 || spSun > 300 {
		t.Errorf("SP Sun = %.4f, expected ~293-296 (Capricorn)", spSun)
	}

	// SA offset sanity check: for age ~28, offset should be ~28°
	saOffset, err := progressions.SolarArcOffset(jnJDUT, transitJD)
	if err != nil {
		t.Fatalf("SA offset: %v", err)
	}
	if saOffset < 26 || saOffset > 30 {
		t.Errorf("SA offset = %.4f, expected ~28° for age 28", saOffset)
	}
	t.Logf("SP SA offset: %.4f", saOffset)
}

// =============================================================================
// TestJN_SR: Solar Return
//
// Finds JN's 2025 solar return and verifies Sun accuracy.
// Expected: return date around 2025-12-18, Sun position ≈ 266.5° (natal value).
// Tolerance: Sun position within 0.01° of natal position.
// =============================================================================
func TestJN_SR(t *testing.T) {
	searchJD := sweph.JulDay(2025, 11, 1, 0, true) // search from Nov 1

	rc, err := returns.CalcSolarReturn(returns.ReturnInput{
		NatalJD:     jnJDUT,
		NatalLat:    jnLat,
		NatalLon:    jnLon,
		SearchJD:    searchJD,
		Planets:     jnPlanets,
		OrbConfig:   models.DefaultOrbConfig(),
		HouseSystem: models.HousePlacidus,
	})
	if err != nil {
		t.Fatalf("CalcSolarReturn: %v", err)
	}

	// Sun at return must match natal Sun within 0.01°
	if math.Abs(normDiff180(rc.PlanetLon, 266.500)) > 0.01 {
		t.Errorf("SR Sun accuracy: got %.4f, natal 266.500 (diff %.4f)", rc.PlanetLon, rc.PlanetLon-266.500)
	}

	// Return must be in Dec 2025 (JD 2460999–2461031)
	decStart := sweph.JulDay(2025, 12, 1, 0, true)
	decEnd := sweph.JulDay(2026, 1, 1, 0, true)
	if rc.ReturnJD < decStart || rc.ReturnJD > decEnd {
		t.Errorf("SR JD = %.1f not in Dec 2025 window", rc.ReturnJD)
	}

	if rc.ReturnType != "solar" {
		t.Errorf("ReturnType = %s, want solar", rc.ReturnType)
	}
	if rc.Chart == nil {
		t.Error("SR Chart is nil")
	}

	t.Logf("SR 2025: JD=%.2f age=%.3f SunLon=%.4f ReturnType=%s", rc.ReturnJD, rc.Age, rc.PlanetLon, rc.ReturnType)
}

// =============================================================================
// TestJN_TR: Transit Events (7-day narrow window)
//
// Scans a narrow 7-day window (2026-01-08 to 2026-01-15) with 3 fast-moving
// planets (Sun, Venus, Mars) vs natal chart to minimize ephemeris calls and
// keep execution time under 100ms.
// Expected: at least 1 transit event (aspects or sign ingresses) found.
// =============================================================================
func TestJN_TR(t *testing.T) {
	startJD := sweph.JulDay(2026, 1, 8, 0, true)
	endJD := sweph.JulDay(2026, 1, 15, 0, true)

	events, err := transit.CalcTransitEvents(transit.TransitCalcInput{
		NatalChart: transit.NatalChartConfig{
			Lat:     jnLat,
			Lon:     jnLon,
			JD:      jnJDUT,
			Planets: jnPlanets,
		},
		TimeRange: transit.TimeRangeConfig{StartJD: startJD, EndJD: endJD},
		Charts: transit.ChartSetConfig{
			Transit: &transit.TransitChartConfig{
				Lat:         jnLat,
				Lon:         jnLon,
				Planets:     []models.PlanetID{models.PlanetSun, models.PlanetVenus, models.PlanetMars},
				Orbs:        models.DefaultOrbConfig(),
				HouseSystem: models.HousePlacidus,
			},
		},
		EventFilter: transit.EventFilterConfig{TrNa: true, SignIngress: true},
	})
	if err != nil {
		t.Fatalf("CalcTransitEvents: %v", err)
	}

	// Must find at least some events in 7 days with 3 fast planets vs 10 natal
	if len(events) == 0 {
		t.Error("TR: expected at least 1 transit event in 7-day window, got none")
	}

	// All events must have valid JD within window
	for _, e := range events {
		if e.JD < startJD || e.JD > endJD {
			t.Errorf("TR event JD %.2f outside window [%.2f, %.2f]", e.JD, startJD, endJD)
		}
	}

	t.Logf("TR: %d events in 7-day window Jan 8-15", len(events))
}

// =============================================================================
// TestJN_Moon: Lunar Phase
//
// Verifies Moon phase at transit date 2026-01-01.
// Expected: valid phase name, illumination 0-1, phase angle 0-360°.
// =============================================================================
func TestJN_Moon(t *testing.T) {
	phase, err := lunar.CalcLunarPhase(transitJD)
	if err != nil {
		t.Fatalf("CalcLunarPhase: %v", err)
	}

	if phase.PhaseAngle < 0 || phase.PhaseAngle >= 360 {
		t.Errorf("PhaseAngle out of range: %.4f", phase.PhaseAngle)
	}
	if phase.Illumination < 0 || phase.Illumination > 1 {
		t.Errorf("Illumination out of range: %.4f", phase.Illumination)
	}
	if phase.MoonLon < 0 || phase.MoonLon >= 360 {
		t.Errorf("MoonLon out of range: %.4f", phase.MoonLon)
	}
	if phase.PhaseName == "" {
		t.Error("PhaseName empty")
	}

	t.Logf("Moon 2026-01-01: phase=%s angle=%.2f illum=%.1f%% moonLon=%.4f",
		phase.PhaseName, phase.PhaseAngle, phase.Illumination*100, phase.MoonLon)
}

// =============================================================================
// TestJN_DoubleChart: Biwheel Chart
//
// Constructs a double chart with JN natal (inner ring) vs transit snapshot
// at 2026-01-01 (outer ring). Verifies:
// - Inner chart planets match known natal positions
// - Outer chart is valid and non-empty
// - Cross-aspects exist between inner and outer planets
// =============================================================================
func TestJN_DoubleChart(t *testing.T) {
	orbs := models.DefaultOrbConfig()

	// Inner: JN natal.  Outer: transit snapshot on 2026-01-01
	inner, outer, crossAspects, err := chart.CalcDoubleChart(
		jnLat, jnLon, jnJDUT, jnPlanets,        // natal (inner ring)
		jnLat, jnLon, transitJD, jnPlanets,     // transit snapshot (outer ring)
		nil, orbs, models.HousePlacidus,
	)
	if err != nil {
		t.Fatalf("CalcDoubleChart: %v", err)
	}

	// Inner must have natal planets with known positions
	innerMap := make(map[models.PlanetID]float64)
	for _, p := range inner.Planets {
		innerMap[p.PlanetID] = p.Longitude
	}

	if math.Abs(innerMap[models.PlanetSun]-266.500) > tolLon {
		t.Errorf("Inner Sun: got %.4f, want 266.500", innerMap[models.PlanetSun])
	}
	if math.Abs(innerMap[models.PlanetMoon]-138.116) > tolLon {
		t.Errorf("Inner Moon: got %.4f, want 138.116", innerMap[models.PlanetMoon])
	}

	// Outer must be a valid chart
	if len(outer.Planets) == 0 {
		t.Error("DoubleChart outer has no planets")
	}
	for _, p := range outer.Planets {
		if p.Longitude < 0 || p.Longitude >= 360 {
			t.Errorf("Outer %s lon out of range: %.4f", p.PlanetID, p.Longitude)
		}
	}

	// Cross aspects must exist between two different charts (10 vs 10 planets)
	if len(crossAspects) == 0 {
		t.Error("DoubleChart: expected cross-aspects between natal and transit, got none")
	}

	// Verify cross-aspects have valid structure
	for _, asp := range crossAspects {
		if asp.InnerBody == "" || asp.OuterBody == "" {
			t.Error("DoubleChart cross-aspect has empty body reference")
		}
		if asp.AspectAngle < 0 || asp.AspectAngle > 360 {
			t.Errorf("DoubleChart cross-aspect angle out of range: %.4f", asp.AspectAngle)
		}
	}

	t.Logf("DoubleChart: inner=%d outer=%d cross-aspects=%d", len(inner.Planets), len(outer.Planets), len(crossAspects))
}

// =============================================================================
// Helper Functions
// =============================================================================

// normDiff180 returns the smallest signed difference between two angles,
// handling the 0/360° wrap-around.
func normDiff180(a, b float64) float64 {
	d := a - b
	for d > 180 {
		d -= 360
	}
	for d < -180 {
		d += 360
	}
	return d
}
