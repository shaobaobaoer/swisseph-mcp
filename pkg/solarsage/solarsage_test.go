package solarsage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	ephePath, _ := filepath.Abs("../../third_party/swisseph/ephe")
	Init(ephePath)
	defer Close()
	os.Exit(m.Run())
}

func TestNatalChart(t *testing.T) {
	chart, err := NatalChart(51.5074, -0.1278, "2000-01-01T12:00:00Z")
	if err != nil {
		t.Fatalf("NatalChart: %v", err)
	}
	if len(chart.Planets) != 10 {
		t.Errorf("Expected 10 planets, got %d", len(chart.Planets))
	}
	if len(chart.Houses) != 12 {
		t.Errorf("Expected 12 houses, got %d", len(chart.Houses))
	}
	if len(chart.Aspects) == 0 {
		t.Error("Expected some aspects")
	}
}

func TestNatalChartFull(t *testing.T) {
	chart, err := NatalChartFull(51.5074, -0.1278, "2000-01-01T12:00:00Z")
	if err != nil {
		t.Fatalf("NatalChartFull: %v", err)
	}
	if len(chart.Planets) != 14 {
		t.Errorf("Expected 14 planets (full set), got %d", len(chart.Planets))
	}
}

func TestNatalChart_InvalidDatetime(t *testing.T) {
	_, err := NatalChart(51.5074, -0.1278, "not-a-date")
	if err == nil {
		t.Error("Expected error for invalid datetime")
	}
}

func TestNatalChart_InvalidCoords(t *testing.T) {
	_, err := NatalChart(999, -0.1278, "2000-01-01T12:00:00Z")
	if err == nil {
		t.Error("Expected error for invalid latitude")
	}
	_, err = NatalChart(51.5, 999, "2000-01-01T12:00:00Z")
	if err == nil {
		t.Error("Expected error for invalid longitude")
	}
}

func TestValidateCoords(t *testing.T) {
	tests := []struct {
		lat, lon float64
		valid    bool
	}{
		{0, 0, true},
		{90, 180, true},
		{-90, -180, true},
		{91, 0, false},
		{0, 181, false},
		{-91, 0, false},
	}
	for _, tt := range tests {
		err := ValidateCoords(tt.lat, tt.lon)
		if tt.valid && err != nil {
			t.Errorf("ValidateCoords(%v, %v) unexpected error: %v", tt.lat, tt.lon, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("ValidateCoords(%v, %v) expected error", tt.lat, tt.lon)
		}
	}
}

func TestTransits(t *testing.T) {
	events, err := Transits(51.5074, -0.1278, "2000-01-01T12:00:00Z",
		"2000-02-01T00:00:00Z", "2000-03-01T00:00:00Z")
	if err != nil {
		t.Fatalf("Transits: %v", err)
	}
	if len(events) == 0 {
		t.Error("Expected some transit events")
	}
}

func TestSolarReturn(t *testing.T) {
	sr, err := SolarReturn(51.5074, -0.1278, "1990-06-15T14:00:00Z", 2025)
	if err != nil {
		t.Fatalf("SolarReturn: %v", err)
	}
	if sr.ReturnType != "solar" {
		t.Errorf("ReturnType = %s, want solar", sr.ReturnType)
	}
	if sr.Age < 34 || sr.Age > 36 {
		t.Errorf("Age = %.2f, expected ~35", sr.Age)
	}
}

func TestLunarReturn(t *testing.T) {
	lr, err := LunarReturn(51.5074, -0.1278, "2000-01-01T12:00:00Z", "2000-01-25T00:00:00Z")
	if err != nil {
		t.Fatalf("LunarReturn: %v", err)
	}
	if lr.ReturnType != "lunar" {
		t.Errorf("ReturnType = %s, want lunar", lr.ReturnType)
	}
}

func TestMoonPhase(t *testing.T) {
	phase, err := MoonPhase("2000-01-01T12:00:00Z")
	if err != nil {
		t.Fatalf("MoonPhase: %v", err)
	}
	if phase.PhaseName == "" {
		t.Error("Phase name is empty")
	}
	if phase.Illumination < 0 || phase.Illumination > 1 {
		t.Errorf("Illumination out of range: %f", phase.Illumination)
	}
}

func TestEclipses(t *testing.T) {
	eclipses, err := Eclipses("2000-01-01T00:00:00Z", "2001-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("Eclipses: %v", err)
	}
	if len(eclipses) < 2 {
		t.Errorf("Expected at least 2 eclipses in 2000, got %d", len(eclipses))
	}
}

func TestCompatibility(t *testing.T) {
	score, err := Compatibility(
		51.5074, -0.1278, "2000-01-01T12:00:00Z",
		40.7128, -74.006, "2001-01-01T12:00:00Z",
	)
	if err != nil {
		t.Fatalf("Compatibility: %v", err)
	}
	if score.Compatibility < 0 || score.Compatibility > 100 {
		t.Errorf("Compatibility out of range: %.2f", score.Compatibility)
	}
}

func TestCompositeChart(t *testing.T) {
	cc, err := CompositeChart(
		51.5074, -0.1278, "2000-01-01T12:00:00Z",
		40.7128, -74.006, "2001-01-01T12:00:00Z",
	)
	if err != nil {
		t.Fatalf("CompositeChart: %v", err)
	}
	if len(cc.Planets) != 10 {
		t.Errorf("Expected 10 planets, got %d", len(cc.Planets))
	}
}

func TestDignities(t *testing.T) {
	dignities, err := Dignities(51.5074, -0.1278, "2000-01-01T12:00:00Z")
	if err != nil {
		t.Fatalf("Dignities: %v", err)
	}
	if len(dignities) != 10 {
		t.Errorf("Expected 10 dignities, got %d", len(dignities))
	}
}

func TestAspectPatterns(t *testing.T) {
	patterns, err := AspectPatterns(51.5074, -0.1278, "2000-01-01T12:00:00Z")
	if err != nil {
		t.Fatalf("AspectPatterns: %v", err)
	}
	// May or may not find patterns depending on the chart
	_ = patterns
}


func TestPlanetPosition(t *testing.T) {
	pos, err := PlanetPosition("Sun", "2000-01-01T12:00:00Z")
	if err != nil {
		t.Fatalf("PlanetPosition: %v", err)
	}
	if pos.Sign != "Capricorn" && pos.Sign != "Sagittarius" {
		// Sun at J2000 should be in Capricorn (~280°)
		t.Errorf("Sun sign = %s", pos.Sign)
	}
}

func TestPlanetPosition_CaseInsensitive(t *testing.T) {
	tests := []string{"sun", "Sun", "SUN", "moon", "VENUS", "Mars"}
	for _, name := range tests {
		_, err := PlanetPosition(name, "2000-01-01T12:00:00Z")
		if err != nil {
			t.Errorf("PlanetPosition(%q): %v", name, err)
		}
	}
}

func TestPlanetPosition_InvalidPlanet(t *testing.T) {
	_, err := PlanetPosition("invalid", "2000-01-01T12:00:00Z")
	if err == nil {
		t.Error("Expected error for invalid planet")
	}
}

func TestNatalChartWithOptions(t *testing.T) {
	opts := DefaultOptions()
	opts.HouseSystem = "KOCH"
	chart, err := NatalChartWithOptions(51.5074, -0.1278, "2000-01-01T12:00:00Z", opts)
	if err != nil {
		t.Fatalf("NatalChartWithOptions: %v", err)
	}
	if len(chart.Planets) != 10 {
		t.Errorf("Expected 10 planets, got %d", len(chart.Planets))
	}
}

func TestBatchNatalCharts(t *testing.T) {
	people := []Person{
		{Lat: 51.5074, Lon: -0.1278, Datetime: "2000-01-01T12:00:00Z"},
		{Lat: 40.7128, Lon: -74.006, Datetime: "1990-06-15T14:30:00Z"},
		{Lat: 35.6762, Lon: 139.6503, Datetime: "1985-12-25T08:00:00Z"},
	}
	charts, errs := BatchNatalCharts(people)
	if len(charts) != 3 {
		t.Fatalf("Expected 3 charts, got %d", len(charts))
	}
	for i, err := range errs {
		if err != nil {
			t.Errorf("Person %d error: %v", i, err)
		}
	}
	for i, c := range charts {
		if c == nil {
			t.Errorf("Person %d chart is nil", i)
		}
	}
}

func TestParseDatetime_Timezone(t *testing.T) {
	// "2000-01-01T20:00:00+08:00" = 2000-01-01T12:00:00Z = J2000.0
	jd, err := ParseDatetime("2000-01-01T20:00:00+08:00")
	if err != nil {
		t.Fatalf("ParseDatetime with tz: %v", err)
	}
	// Should equal J2000.0 (2451545.0)
	if jd < 2451544.9 || jd > 2451545.1 {
		t.Errorf("JD = %.6f, expected ~2451545.0", jd)
	}
}

func TestParseDatetime(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"2000-01-01T12:00:00Z", true},
		{"2000-01-01T12:00:00", true},
		{"2000-01-01 12:00:00", true},
		{"2000-01-01 12:00", true},
		{"2000-01-01", true},
		{"not-a-date", false},
		{"", false},
	}
	for _, tt := range tests {
		_, err := ParseDatetime(tt.input)
		if tt.valid && err != nil {
			t.Errorf("ParseDatetime(%q) unexpected error: %v", tt.input, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("ParseDatetime(%q) expected error", tt.input)
		}
	}
}

func TestParsePlanet(t *testing.T) {
	tests := map[string]bool{
		"Sun":             true,
		"sun":             true,
		"SUN":             true,
		"NORTH_NODE":      true,
		"NORTHNODE":       true,
		"Chiron":          true, // case-insensitive
		"CHIRON":          true,
		"invalid":         false,
		"LILITH":          true,
		"LILITH_TRUE":     true,
		"":                false,
		"  ":              false,
	}
	for name, shouldWork := range tests {
		_, err := ParsePlanet(name)
		if shouldWork && err != nil {
			t.Errorf("ParsePlanet(%q) unexpected error: %v", name, err)
		}
		if !shouldWork && err == nil {
			t.Errorf("ParsePlanet(%q) expected error", name)
		}
	}
}
