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

	// Must have exactly 10 planets (standard set)
	if len(info.Planets) != 10 {
		t.Errorf("NA: expected 10 planets, got %d", len(info.Planets))
	}

	// Spot-check 4 Solar Fire validated planet positions
	wantLon := map[models.PlanetID]float64{
		models.PlanetSun:     266.500,
		models.PlanetMoon:    138.116,
		models.PlanetMars:    300.097,
		models.PlanetSaturn:  13.538,
	}

	wantHouse := map[models.PlanetID]int{
		models.PlanetSun:    6,
		models.PlanetMoon:   2,
		models.PlanetMars:   8,
		models.PlanetSaturn: 10,
	}

	gotMap := make(map[models.PlanetID]models.PlanetPosition)
	for _, p := range info.Planets {
		gotMap[p.PlanetID] = p
	}

	for pid, want := range wantLon {
		got, ok := gotMap[pid]
		if !ok {
			t.Errorf("NA: planet %s not found", pid)
			continue
		}
		if math.Abs(got.Longitude-want) > tolLon {
			t.Errorf("NA %s lon: got %.4f, want %.3f (diff %.4f)", pid, got.Longitude, want, got.Longitude-want)
		}
		// Verify house placement
		if h, ok := wantHouse[pid]; ok {
			if got.House != h {
				t.Errorf("NA %s house: got %d, want %d", pid, got.House, h)
			}
		}
		// Verify all planets have valid positions
		if got.Longitude < 0 || got.Longitude >= 360 {
			t.Errorf("NA %s: longitude out of range %.4f", pid, got.Longitude)
		}
		if got.Sign == "" {
			t.Errorf("NA %s: empty sign", pid)
		}
		if got.House < 1 || got.House > 12 {
			t.Errorf("NA %s: house %d out of range", pid, got.House)
		}
	}

	// Verify 12 house cusps
	if len(info.Houses) != 12 {
		t.Errorf("NA: expected 12 house cusps, got %d", len(info.Houses))
	}

	// Verify angles
	if math.Abs(info.Angles.ASC-96.530) > tolAng {
		t.Errorf("ASC: got %.4f, want 96.530", info.Angles.ASC)
	}
	if math.Abs(info.Angles.MC-351.500) > tolAng {
		t.Errorf("MC: got %.4f, want 351.500", info.Angles.MC)
	}
	// Verify opposite angles (DSC ~= ASC+180, IC ~= MC+180)
	dscWant := math.Mod(info.Angles.ASC+180, 360)
	if math.Abs(normDiff180(info.Angles.DSC, dscWant)) > tolAng {
		t.Errorf("DSC: got %.4f, want %.4f (ASC+180)", info.Angles.DSC, dscWant)
	}
	icWant := math.Mod(info.Angles.MC+180, 360)
	if math.Abs(normDiff180(info.Angles.IC, icWant)) > tolAng {
		t.Errorf("IC: got %.4f, want %.4f (MC+180)", info.Angles.IC, icWant)
	}

	t.Logf("NA: Sun=%.4f Moon=%.4f ASC=%.4f MC=%.4f Houses=%d", gotMap[models.PlanetSun].Longitude, gotMap[models.PlanetMoon].Longitude, info.Angles.ASC, info.Angles.MC, len(info.Houses))
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

	// Calculate age
	age := progressions.Age(jnJDUT, transitJD)
	if age < 27 || age > 29 {
		t.Errorf("SP Age = %.4f, expected ~28", age)
	}

	// Calculate progressed positions for each planet
	for _, pid := range spPlanets {
		lon, speed, err := progressions.CalcProgressedLongitude(pid, jnJDUT, transitJD)
		if err != nil {
			t.Fatalf("SP %s: %v", pid, err)
		}
		if lon < 0 || lon >= 360 {
			t.Errorf("SP %s lon out of range: %.4f", pid, lon)
		}
		if speed < 0 {
			t.Errorf("SP %s speed invalid: %.6f", pid, speed)
		}
		// Speed should be slow (progressed time / real time ÷ 365.25)
		if speed > 0.05 {
			t.Errorf("SP %s speed too high: %.6f (should be <0.05)", pid, speed)
		}
		t.Logf("SP %s: lon=%.4f speed=%.6f", pid, lon, speed)
	}

	// SP Sun correctness check: after ~28 days, Sun should have advanced ~28° from natal
	// Natal Sun = 266.5°, expected SP Sun ≈ 293-296° (Capricorn 23-26°)
	spSun, spSunSpeed, err := progressions.CalcProgressedLongitude(models.PlanetSun, jnJDUT, transitJD)
	if err != nil {
		t.Fatalf("SP Sun: %v", err)
	}
	if spSun < 290 || spSun > 300 {
		t.Errorf("SP Sun = %.4f, expected ~293-296 (Capricorn)", spSun)
	}
	// SP Sun speed should be ~0.00274°/day (1°/year ÷ 365.25)
	if spSunSpeed < 0.002 || spSunSpeed > 0.003 {
		t.Errorf("SP Sun speed = %.6f, expected ~0.00274", spSunSpeed)
	}

	// SA offset sanity check: for age ~28, offset should be ~28°
	saOffset, err := progressions.SolarArcOffset(jnJDUT, transitJD)
	if err != nil {
		t.Fatalf("SA offset: %v", err)
	}
	if saOffset < 26 || saOffset > 30 {
		t.Errorf("SA offset = %.4f, expected ~28° for age 28", saOffset)
	}

	// Verify Solar Arc Sun matches progression (both use same offset)
	saLon, _, err := progressions.CalcSolarArcLongitude(models.PlanetSun, jnJDUT, transitJD)
	if err != nil {
		t.Fatalf("SA Sun: %v", err)
	}
	if math.Abs(saLon-spSun) > 0.5 {
		t.Errorf("SA Sun %.4f vs SP Sun %.4f differ too much", saLon, spSun)
	}

	t.Logf("SP: age=%.3f saOffset=%.4f saLon=%.4f spSun=%.4f", age, saOffset, saLon, spSun)
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

	// Age should be JN's age at time of return (~28 years)
	// This is (returnJD - natalJD) / 365.25
	expectedAge := (rc.ReturnJD - jnJDUT) / 365.25
	if math.Abs(rc.Age-expectedAge) > 0.01 {
		t.Errorf("SR Age = %.4f, expected %.4f", rc.Age, expectedAge)
	}
	if rc.Age < 27 || rc.Age > 29 {
		t.Errorf("SR Age = %.4f, expected ~28 for 2025 return", rc.Age)
	}

	// Solar return chart must have planets
	if len(rc.Chart.Planets) == 0 {
		t.Error("SR Chart has no planets")
	}

	// Solar return must have 12 house cusps
	if len(rc.Chart.Houses) != 12 {
		t.Errorf("SR Chart houses = %d, expected 12", len(rc.Chart.Houses))
	}

	t.Logf("SR 2025: JD=%.2f age=%.3f SunLon=%.4f planets=%d ReturnType=%s", rc.ReturnJD, rc.Age, rc.PlanetLon, len(rc.Chart.Planets), rc.ReturnType)
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

	// Verify all events have valid structure
	eventTypes := make(map[string]int)
	for _, e := range events {
		if e.JD < startJD || e.JD > endJD {
			t.Errorf("TR event JD %.2f outside window [%.2f, %.2f]", e.JD, startJD, endJD)
		}
		if e.EventType == "" {
			t.Error("TR event has empty EventType")
		}
		if e.Planet == "" {
			t.Error("TR event has empty Planet")
		}
		if e.PlanetSign == "" {
			t.Error("TR event has empty PlanetSign")
		}
		if e.PlanetLongitude < 0 || e.PlanetLongitude >= 360 {
			t.Errorf("TR event planet longitude out of range: %.4f", e.PlanetLongitude)
		}
		eventTypes[string(e.EventType)]++
	}

	// Should find mostly TrNa aspects and/or sign ingress events
	aspectCount := eventTypes[string(models.EventAspectExact)] + eventTypes[string(models.EventAspectEnter)] + eventTypes[string(models.EventAspectLeave)]
	ingressCount := eventTypes[string(models.EventSignIngress)]
	if aspectCount == 0 && ingressCount == 0 {
		t.Error("TR: expected aspect or sign ingress events")
	}

	t.Logf("TR: %d events in 7-day window (aspects=%d, ingress=%d)", len(events), aspectCount, ingressCount)
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
	if phase.SunLon < 0 || phase.SunLon >= 360 {
		t.Errorf("SunLon out of range: %.4f", phase.SunLon)
	}

	// Verify phase consistency: phase angle = moon - sun elongation (0-360)
	expectedPhaseAngle := phase.MoonLon - phase.SunLon
	for expectedPhaseAngle < 0 {
		expectedPhaseAngle += 360
	}
	if math.Abs(expectedPhaseAngle-phase.PhaseAngle) > 1.0 {
		t.Errorf("Phase angle inconsistent: moon %.1f - sun %.1f = %.1f, but phase=%.1f",
			phase.MoonLon, phase.SunLon, expectedPhaseAngle, phase.PhaseAngle)
	}

	// Waxing vs waning consistency
	if phase.PhaseAngle < 180 {
		if !phase.IsWaxing {
			t.Errorf("Phase angle %.1f indicates waxing but IsWaxing=%v", phase.PhaseAngle, phase.IsWaxing)
		}
	} else {
		if phase.IsWaxing {
			t.Errorf("Phase angle %.1f indicates waning but IsWaxing=%v", phase.PhaseAngle, phase.IsWaxing)
		}
	}

	t.Logf("Moon 2026-01-01: phase=%s angle=%.2f illum=%.1f%% moonLon=%.4f sunLon=%.4f waxing=%v",
		phase.PhaseName, phase.PhaseAngle, phase.Illumination*100, phase.MoonLon, phase.SunLon, phase.IsWaxing)
}

// =============================================================================
// TestJN_DoubleChart: Biwheel Chart
//
// Constructs a double chart with JN natal (inner ring) vs transit snapshot
// at 2026-01-01 (outer ring). Verifies:
// - Inner chart planets match known natal positions
// - Outer chart is valid and non-empty
// - Cross-aspects exist between inner and outer planets
// - All chart elements (houses, angles, aspects) are valid
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

	// Inner must have exactly 10 planets
	if len(inner.Planets) != 10 {
		t.Errorf("Inner planets: got %d, want 10", len(inner.Planets))
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

	// Inner chart must have houses and angles
	if len(inner.Houses) != 12 {
		t.Errorf("Inner houses: got %d, want 12", len(inner.Houses))
	}
	if inner.Angles.ASC <= 0 || inner.Angles.ASC >= 360 {
		t.Errorf("Inner ASC out of range: %.4f", inner.Angles.ASC)
	}
	if inner.Angles.MC <= 0 || inner.Angles.MC >= 360 {
		t.Errorf("Inner MC out of range: %.4f", inner.Angles.MC)
	}

	// Outer must be a valid chart (same size since we passed same planets)
	if len(outer.Planets) != 10 {
		t.Errorf("Outer planets: got %d, want 10", len(outer.Planets))
	}
	if len(outer.Houses) != 12 {
		t.Errorf("Outer houses: got %d, want 12", len(outer.Houses))
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
	seenAspectTypes := make(map[models.AspectType]bool)
	for _, asp := range crossAspects {
		if asp.InnerBody == "" || asp.OuterBody == "" {
			t.Error("DoubleChart cross-aspect has empty body reference")
		}
		if asp.AspectAngle < 0 || asp.AspectAngle > 360 {
			t.Errorf("DoubleChart cross-aspect angle out of range: %.4f", asp.AspectAngle)
		}
		if asp.AspectType == "" {
			t.Error("DoubleChart cross-aspect has empty type")
		}
		seenAspectTypes[asp.AspectType] = true
	}

	// Should find multiple aspect types (not all same type)
	if len(seenAspectTypes) < 2 {
		t.Logf("DoubleChart found only %d aspect types, expected variety", len(seenAspectTypes))
	}

	t.Logf("DoubleChart: inner=%d outer=%d cross-aspects=%d astypes=%d", len(inner.Planets), len(outer.Planets), len(crossAspects), len(seenAspectTypes))
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
