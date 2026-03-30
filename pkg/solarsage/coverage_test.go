package solarsage

import (
	"testing"

	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
)

// Error path coverage for all convenience functions

func TestTransits_InvalidNatal(t *testing.T) {
	_, err := Transits(51.5, -0.1, "bad", "2000-01-01", "2000-02-01")
	if err == nil {
		t.Error("expected error for bad natal datetime")
	}
}

func TestTransits_InvalidStart(t *testing.T) {
	_, err := Transits(51.5, -0.1, "2000-01-01T12:00:00Z", "bad", "2000-02-01")
	if err == nil {
		t.Error("expected error for bad start datetime")
	}
}

func TestTransits_InvalidEnd(t *testing.T) {
	_, err := Transits(51.5, -0.1, "2000-01-01T12:00:00Z", "2000-01-01", "bad")
	if err == nil {
		t.Error("expected error for bad end datetime")
	}
}

func TestSolarReturn_InvalidDatetime(t *testing.T) {
	_, err := SolarReturn(51.5, -0.1, "bad", 2025)
	if err == nil {
		t.Error("expected error")
	}
}

func TestLunarReturn_InvalidNatal(t *testing.T) {
	_, err := LunarReturn(51.5, -0.1, "bad", "2000-01-25")
	if err == nil {
		t.Error("expected error for bad natal datetime")
	}
}

func TestLunarReturn_InvalidSearch(t *testing.T) {
	_, err := LunarReturn(51.5, -0.1, "2000-01-01T12:00:00Z", "bad")
	if err == nil {
		t.Error("expected error for bad search datetime")
	}
}

func TestMoonPhase_InvalidDatetime(t *testing.T) {
	_, err := MoonPhase("bad")
	if err == nil {
		t.Error("expected error")
	}
}

func TestEclipses_InvalidStart(t *testing.T) {
	_, err := Eclipses("bad", "2001-01-01")
	if err == nil {
		t.Error("expected error")
	}
}

func TestEclipses_InvalidEnd(t *testing.T) {
	_, err := Eclipses("2000-01-01", "bad")
	if err == nil {
		t.Error("expected error")
	}
}

func TestCompatibility_InvalidDatetime1(t *testing.T) {
	_, err := Compatibility(51.5, -0.1, "bad", 40.7, -74, "2001-01-01")
	if err == nil {
		t.Error("expected error for bad person 1 datetime")
	}
}

func TestCompatibility_InvalidDatetime2(t *testing.T) {
	_, err := Compatibility(51.5, -0.1, "2000-01-01", 40.7, -74, "bad")
	if err == nil {
		t.Error("expected error for bad person 2 datetime")
	}
}

func TestCompositeChart_InvalidDatetime1(t *testing.T) {
	_, err := CompositeChart(51.5, -0.1, "bad", 40.7, -74, "2001-01-01")
	if err == nil {
		t.Error("expected error")
	}
}

func TestCompositeChart_InvalidDatetime2(t *testing.T) {
	_, err := CompositeChart(51.5, -0.1, "2000-01-01", 40.7, -74, "bad")
	if err == nil {
		t.Error("expected error")
	}
}

func TestDignities_InvalidDatetime(t *testing.T) {
	_, err := Dignities(51.5, -0.1, "bad")
	if err == nil {
		t.Error("expected error")
	}
}

func TestAspectPatterns_InvalidDatetime(t *testing.T) {
	_, err := AspectPatterns(51.5, -0.1, "bad")
	if err == nil {
		t.Error("expected error")
	}
}


func TestPlanetPosition_InvalidDatetime(t *testing.T) {
	_, err := PlanetPosition("Sun", "bad")
	if err == nil {
		t.Error("expected error")
	}
}

func TestNatalChartWithOptions_InvalidDatetime(t *testing.T) {
	_, err := NatalChartWithOptions(51.5, -0.1, "bad", DefaultOptions())
	if err == nil {
		t.Error("expected error")
	}
}

func TestNatalChartWithOptions_InvalidCoords(t *testing.T) {
	_, err := NatalChartWithOptions(999, -0.1, "2000-01-01", DefaultOptions())
	if err == nil {
		t.Error("expected error")
	}
}

func TestBatchGroupCompatibility_InvalidDatetime(t *testing.T) {
	people := []Person{
		{Lat: 51.5, Lon: -0.1, Datetime: "2000-01-01"},
		{Lat: 40.7, Lon: -74, Datetime: "bad"},
	}
	_, err := BatchGroupCompatibility(people)
	if err == nil {
		t.Error("expected error")
	}
}

func TestNatalChartFull_InvalidDatetime(t *testing.T) {
	_, err := NatalChartFull(51.5, -0.1, "bad")
	if err == nil {
		t.Error("expected error")
	}
}

func TestNatalChartWithOptions_AllHouseSystems(t *testing.T) {
	houseSystems := []string{
		"PLACIDUS", "KOCH", "EQUAL", "WHOLE_SIGN", "CAMPANUS",
		"REGIOMONTANUS", "PORPHYRY", "MORINUS", "TOPOCENTRIC",
		"ALCABITIUS", "MERIDIAN",
	}
	for _, hs := range houseSystems {
		opts := DefaultOptions()
		opts.HouseSystem = models.HouseSystem(hs)
		chart, err := NatalChartWithOptions(51.5074, -0.1278, "2000-01-01T12:00:00Z", opts)
		if err != nil {
			t.Errorf("%s: %v", hs, err)
			continue
		}
		if len(chart.Houses) != 12 {
			t.Errorf("%s: expected 12 houses, got %d", hs, len(chart.Houses))
		}
	}
}

func TestParseHouseSystem(t *testing.T) {
	tests := map[string]bool{
		"Placidus":      true,
		"placidus":      true,
		"PLACIDUS":      true,
		"Koch":          true,
		"Whole Sign":    true,
		"WHOLE_SIGN":    true,
		"wholesign":     true,
		"Topocentric":   true,
		"Polich-Page":   true,
		"Morinus":       true,
		"Alcabitius":    true,
		"Meridian":      true,
		"invalid":       false,
		"":              false,
	}
	for name, shouldWork := range tests {
		_, err := ParseHouseSystem(name)
		if shouldWork && err != nil {
			t.Errorf("ParseHouseSystem(%q) unexpected error: %v", name, err)
		}
		if !shouldWork && err == nil {
			t.Errorf("ParseHouseSystem(%q) expected error", name)
		}
	}
}

