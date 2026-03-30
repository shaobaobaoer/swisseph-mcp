package solarsage

import (
	"testing"
)

// TestFullPipeline exercises the complete analysis pipeline end-to-end
func TestFullPipeline(t *testing.T) {
	birthLat, birthLon := 51.5074, -0.1278
	birthDT := "1990-06-15T14:30:00Z"

	// 1. Natal chart
	natal, err := NatalChart(birthLat, birthLon, birthDT)
	if err != nil {
		t.Fatalf("NatalChart: %v", err)
	}
	if len(natal.Planets) != 10 {
		t.Errorf("Expected 10 planets, got %d", len(natal.Planets))
	}
	t.Logf("Natal: %s", natal.Angles)

	// 2. Full chart with all bodies
	full, err := NatalChartFull(birthLat, birthLon, birthDT)
	if err != nil {
		t.Fatalf("NatalChartFull: %v", err)
	}
	if len(full.Planets) < len(natal.Planets) {
		t.Error("Full chart should have more planets than default")
	}

	// 3. Solar return for current year
	sr, err := SolarReturn(birthLat, birthLon, birthDT, 2025)
	if err != nil {
		t.Fatalf("SolarReturn: %v", err)
	}
	t.Logf("Solar return age: %.1f", sr.Age)

	// 4. Lunar return
	lr, err := LunarReturn(birthLat, birthLon, birthDT, "2025-03-01")
	if err != nil {
		t.Fatalf("LunarReturn: %v", err)
	}
	t.Logf("Lunar return type: %s", lr.ReturnType)

	// 5. Moon phase
	phase, err := MoonPhase("2025-03-18T12:00:00Z")
	if err != nil {
		t.Fatalf("MoonPhase: %v", err)
	}
	t.Logf("Moon: %s (%.0f%%)", phase.PhaseName, phase.Illumination*100)

	// 6. Eclipses
	eclipses, err := Eclipses("2025-01-01", "2026-01-01")
	if err != nil {
		t.Fatalf("Eclipses: %v", err)
	}
	t.Logf("Eclipses in 2025: %d", len(eclipses))

	// 7. Essential dignities
	dignities, err := Dignities(birthLat, birthLon, birthDT)
	if err != nil {
		t.Fatalf("Dignities: %v", err)
	}
	for _, d := range dignities {
		if d.Score != 0 {
			t.Logf("Dignity: %s in %s = %d %v", d.PlanetID, d.Sign, d.Score, d.Dignities)
		}
	}

	// 8. Aspect patterns
	patterns, err := AspectPatterns(birthLat, birthLon, birthDT)
	if err != nil {
		t.Fatalf("AspectPatterns: %v", err)
	}
	t.Logf("Patterns found: %d", len(patterns))


	// 10. Compatibility
	partner := "1992-03-22T08:00:00Z"
	score, err := Compatibility(
		birthLat, birthLon, birthDT,
		40.7128, -74.006, partner,
	)
	if err != nil {
		t.Fatalf("Compatibility: %v", err)
	}
	t.Logf("Compatibility: %.0f%% (harmony=%.1f, tension=%.1f)",
		score.Compatibility, score.Harmony, score.Tension)

	// 11. Composite chart
	comp, err := CompositeChart(
		birthLat, birthLon, birthDT,
		40.7128, -74.006, partner,
	)
	if err != nil {
		t.Fatalf("CompositeChart: %v", err)
	}
	if len(comp.Planets) != 10 {
		t.Errorf("Composite planets: %d", len(comp.Planets))
	}

	// 12. Transit search (1 month)
	events, err := Transits(birthLat, birthLon, birthDT,
		"2025-03-01", "2025-04-01")
	if err != nil {
		t.Fatalf("Transits: %v", err)
	}
	t.Logf("Transit events (1 month): %d", len(events))

	// 13. Planet position
	pos, err := PlanetPosition("Venus", "2025-03-18T12:00:00Z")
	if err != nil {
		t.Fatalf("PlanetPosition: %v", err)
	}
	t.Logf("Venus: %s", pos)

}
