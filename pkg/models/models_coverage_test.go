package models

import (
	"testing"
)

func TestAspectCSVName(t *testing.T) {
	tests := []struct {
		at   AspectType
		want string
	}{
		{AspectConjunction, "Conjunction"},
		{AspectOpposition, "Opposition"},
		{AspectTrine, "Trine"},
		{AspectSquare, "Square"},
		{AspectSextile, "Sextile"},
		{AspectQuincunx, "Quincunx"},
		{AspectSemiSextile, "Semi-Sextile"},
		{AspectSemiSquare, "Semi-Square"},
		{AspectSesquiquadrate, "Sesquiquadrate"},
	}
	for _, tt := range tests {
		got := AspectCSVName(tt.at)
		if got != tt.want {
			t.Errorf("AspectCSVName(%s) = %q, want %q", tt.at, got, tt.want)
		}
	}
}

func TestSignAbbrFromLongitude(t *testing.T) {
	tests := []struct {
		lon  float64
		want string
	}{
		{0, "Ari"}, {35, "Tau"}, {65, "Gem"}, {95, "Can"},
		{125, "Leo"}, {155, "Vir"}, {185, "Lib"}, {215, "Sco"},
		{245, "Sag"}, {275, "Cap"}, {305, "Aqu"}, {335, "Pis"},
		{-1, "Ari"}, {361, "Pis"},
	}
	for _, tt := range tests {
		got := SignAbbrFromLongitude(tt.lon)
		if got != tt.want {
			t.Errorf("SignAbbrFromLongitude(%f) = %q, want %q", tt.lon, got, tt.want)
		}
	}
}

func TestToDMS(t *testing.T) {
	tests := []struct {
		deg  float64
		d, m, s int
		neg  bool
	}{
		{0, 0, 0, 0, false},
		{10.5, 10, 30, 0, false},
		{-10.5, 10, 30, 0, true},
		{23.4381, 23, 26, 17, false},
	}
	for _, tt := range tests {
		got := ToDMS(tt.deg)
		if got.Degrees != tt.d || got.Minutes != tt.m || got.Seconds != tt.s || got.Neg != tt.neg {
			t.Errorf("ToDMS(%f) = %+v, want d=%d m=%d s=%d neg=%v", tt.deg, got, tt.d, tt.m, tt.s, tt.neg)
		}
	}
}

func TestDMSString(t *testing.T) {
	tests := []struct {
		dms  DMS
		want string
	}{
		{DMS{10, 30, 0, false}, "10°30'00\""},
		{DMS{10, 30, 0, true}, "-10°30'00\""},
		{DMS{0, 0, 0, false}, "0°00'00\""},
	}
	for _, tt := range tests {
		got := tt.dms.String()
		if got != tt.want {
			t.Errorf("DMS.String() = %q, want %q", got, tt.want)
		}
	}
}

func TestFormatLonDMS(t *testing.T) {
	got := FormatLonDMS(266.4997)
	if got == "" {
		t.Error("FormatLonDMS returned empty")
	}
}

func TestFormatDMS(t *testing.T) {
	got := FormatDMS(10.5)
	if got != "10°30'00\"" {
		t.Errorf("FormatDMS(10.5) = %q, want 10°30'00\"", got)
	}
}

func TestBodyDisplayName(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		{"SUN", "Sun"}, {"MOON", "Moon"}, {"MERCURY", "Mercury"},
		{"VENUS", "Venus"}, {"MARS", "Mars"}, {"JUPITER", "Jupiter"},
		{"SATURN", "Saturn"}, {"URANUS", "Uranus"}, {"NEPTUNE", "Neptune"},
		{"PLUTO", "Pluto"}, {"CHIRON", "Chiron"},
		{"NORTH_NODE_TRUE", "NorthNode"}, {"NORTH_NODE_MEAN", "NorthNode"},
		{"SOUTH_NODE", "SouthNode"},
		{"LILITH_MEAN", "Lilith"}, {"LILITH_TRUE", "Lilith"},
		{"ASC", "ASC"}, {"MC", "MC"}, {"DSC", "DSC"}, {"IC", "IC"},
		{"VERTEX", "Vertex"}, {"ANTI_VERTEX", "AntiVertex"},
		{"EAST_POINT", "EastPoint"},
		{"LOT_FORTUNE", "LotFortune"}, {"LOT_SPIRIT", "LotSpirit"},
		{"UNKNOWN", "UNKNOWN"},
	}
	for _, tt := range tests {
		got := BodyDisplayName(tt.id)
		if got != tt.want {
			t.Errorf("BodyDisplayName(%q) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestChartTypeShort(t *testing.T) {
	tests := []struct {
		ct   ChartType
		want string
	}{
		{ChartTransit, "Tr"}, {ChartNatal, "Na"},
		{ChartProgressions, "Sp"}, {ChartSolarArc, "Sa"},
		{ChartType("OTHER"), "OTHER"},
	}
	for _, tt := range tests {
		got := ChartTypeShort(tt.ct)
		if got != tt.want {
			t.Errorf("ChartTypeShort(%s) = %q, want %q", tt.ct, got, tt.want)
		}
	}
}

func TestEventTypeCSV(t *testing.T) {
	tests := []struct {
		et   EventType
		st   StationType
		want string
	}{
		{EventAspectBegin, "", "Begin"},
		{EventAspectEnter, "", "Enter"},
		{EventAspectExact, "", "Exact"},
		{EventAspectLeave, "", "Leave"},
		{EventSignIngress, "", "SignIngress"},
		{EventStation, StationRetrograde, "Retrograde"},
		{EventStation, StationDirect, "Direct"},
		{EventVoidOfCourse, "", "Void"},
		{EventHouseIngress, "", "HouseIngress"},
		{EventType("OTHER"), "", "OTHER"},
	}
	for _, tt := range tests {
		got := EventTypeCSV(tt.et, tt.st)
		if got != tt.want {
			t.Errorf("EventTypeCSV(%s, %s) = %q, want %q", tt.et, tt.st, got, tt.want)
		}
	}
}

func TestFormatSignDegreeCSV(t *testing.T) {
	tests := []struct {
		deg  float64
		want float64
	}{
		{0, 0},
		{15.5, 15.5},
		{29.999, 29.983},
	}
	for _, tt := range tests {
		got := FormatSignDegreeCSV(tt.deg)
		if got != tt.want {
			t.Errorf("FormatSignDegreeCSV(%f) = %f, want %f", tt.deg, got, tt.want)
		}
	}
}

func TestPlanetToSweID(t *testing.T) {
	// Test known planets
	knownPlanets := []PlanetID{
		PlanetSun, PlanetMoon, PlanetMercury, PlanetVenus, PlanetMars,
		PlanetJupiter, PlanetSaturn, PlanetUranus, PlanetNeptune, PlanetPluto,
		PlanetChiron, PlanetNorthNodeTrue, PlanetNorthNodeMean,
		PlanetLilithMean, PlanetLilithTrue,
	}
	for _, p := range knownPlanets {
		_, ok := PlanetToSweID(p)
		if !ok {
			t.Errorf("PlanetToSweID(%s) returned false", p)
		}
	}

	// Test unknown planet
	_, ok := PlanetToSweID(PlanetID("UNKNOWN"))
	if ok {
		t.Error("PlanetToSweID(UNKNOWN) should return false")
	}

	// South node has no direct mapping
	_, ok = PlanetToSweID(PlanetSouthNode)
	if ok {
		t.Error("PlanetToSweID(SOUTH_NODE) should return false (handled separately)")
	}
}

func TestHouseSystemToChar(t *testing.T) {
	systems := []HouseSystem{
		HousePlacidus, HouseKoch, HouseEqual, HouseWholeSign,
		HouseCampanus, HouseRegiomontanus, HousePorphyry,
	}
	for _, hs := range systems {
		c := HouseSystemToChar(hs)
		if c == 0 {
			t.Errorf("HouseSystemToChar(%s) returned 0", hs)
		}
	}

	// Unknown system should default to Placidus
	def := HouseSystemToChar(HouseSystem("UNKNOWN"))
	plac := HouseSystemToChar(HousePlacidus)
	if def != plac {
		t.Errorf("HouseSystemToChar(UNKNOWN) = %d, want %d (Placidus default)", def, plac)
	}
}

func TestAllPlanets(t *testing.T) {
	if len(AllPlanets) != 16 {
		t.Errorf("AllPlanets has %d entries, want 16", len(AllPlanets))
	}
}
