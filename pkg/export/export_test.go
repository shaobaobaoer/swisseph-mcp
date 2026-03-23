package export

import (
	"strings"
	"testing"

	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
)

func TestCSVHeader(t *testing.T) {
	h := CSVHeader()
	if !strings.Contains(h, "P1") || !strings.Contains(h, "Aspect") {
		t.Error("CSV header missing expected columns")
	}
}

func TestDirection(t *testing.T) {
	if direction(true) != "Rx" {
		t.Error("direction(true) should be Rx")
	}
	if direction(false) != "Dir" {
		t.Error("direction(false) should be Dir")
	}
}

func TestChartPairType(t *testing.T) {
	tests := []struct {
		evt  models.TransitEvent
		want string
	}{
		{models.TransitEvent{EventType: models.EventStation, ChartType: models.ChartTransit}, "Tr"},
		{models.TransitEvent{EventType: models.EventStation, ChartType: models.ChartProgressions}, "Sp"},
		{models.TransitEvent{EventType: models.EventSignIngress, ChartType: models.ChartTransit}, "Tr-Tr"},
		{models.TransitEvent{EventType: models.EventSignIngress, ChartType: models.ChartProgressions}, "Sp-Na"},
		{models.TransitEvent{EventType: models.EventVoidOfCourse, ChartType: models.ChartTransit}, "Tr-Tr"},
		{models.TransitEvent{EventType: models.EventAspectExact, ChartType: models.ChartTransit, TargetChartType: models.ChartNatal}, "Tr-Na"},
		{models.TransitEvent{EventType: models.EventAspectExact, ChartType: models.ChartProgressions, TargetChartType: models.ChartNatal}, "Sp-Na"},
	}
	for _, tt := range tests {
		got := chartPairType(tt.evt)
		if got != tt.want {
			t.Errorf("chartPairType(%v) = %q, want %q", tt.evt.EventType, got, tt.want)
		}
	}
}

func TestFormatDeg(t *testing.T) {
	tests := []struct {
		d    float64
		want string
	}{
		{0, "0.0"},
		{10.5, "10.5"},
		{10.123, "10.123"},
		{10.100, "10.1"},
		{10.000, "10.0"},
	}
	for _, tt := range tests {
		got := formatDeg(tt.d)
		if got != tt.want {
			t.Errorf("formatDeg(%f) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestEventToCSVRow_Station(t *testing.T) {
	evt := models.TransitEvent{
		EventType:       models.EventStation,
		ChartType:       models.ChartTransit,
		Planet:          models.PlanetMercury,
		JD:              2460310.667,
		PlanetLongitude: 120.5,
		PlanetSign:      "Leo",
		StationType:     models.StationRetrograde,
		IsRetrograde:    true,
	}
	row := EventToCSVRow(evt, "UTC")
	if row.P1 != "Mercury" {
		t.Errorf("Station P1 = %q, want Mercury", row.P1)
	}
	if row.Aspect != "Station" {
		t.Errorf("Station Aspect = %q, want Station", row.Aspect)
	}
	if row.EventType != "Retrograde" {
		t.Errorf("Station EventType = %q, want Retrograde", row.EventType)
	}
}

func TestEventToCSVRow_SignIngress(t *testing.T) {
	evt := models.TransitEvent{
		EventType:       models.EventSignIngress,
		ChartType:       models.ChartTransit,
		Planet:          models.PlanetSun,
		JD:              2460310.667,
		PlanetLongitude: 0,
		ToSign:          "Aries",
	}
	row := EventToCSVRow(evt, "Asia/Shanghai")
	if row.P2 != "Aries" {
		t.Errorf("SignIngress P2 = %q, want Aries", row.P2)
	}
	if row.EventType != "SignIngress" {
		t.Errorf("SignIngress EventType = %q, want SignIngress", row.EventType)
	}
}

func TestEventToCSVRow_VoidOfCourse(t *testing.T) {
	evt := models.TransitEvent{
		EventType:       models.EventVoidOfCourse,
		ChartType:       models.ChartTransit,
		Planet:          models.PlanetMoon,
		JD:              2460310.667,
		PlanetLongitude: 45.5,
		PlanetSign:      "Taurus",
		LastAspectType:  "Trine",
		LastAspectTarget: "JUPITER",
		TargetLongitude: 120.0,
		TargetSign:      "Leo",
	}
	row := EventToCSVRow(evt, "UTC")
	if row.Aspect != "Trine" {
		t.Errorf("VOC Aspect = %q, want Trine", row.Aspect)
	}
	if row.EventType != "Void" {
		t.Errorf("VOC EventType = %q, want Void", row.EventType)
	}
}

func TestEventToCSVRow_HouseIngress(t *testing.T) {
	evt := models.TransitEvent{
		EventType:       models.EventHouseIngress,
		ChartType:       models.ChartTransit,
		Planet:          models.PlanetMoon,
		JD:              2460310.667,
		PlanetLongitude: 90.0,
		PlanetSign:      "Cancer",
		ToHouse:         7,
	}
	row := EventToCSVRow(evt, "UTC")
	if row.P2 != "House7" {
		t.Errorf("HouseIngress P2 = %q, want House7", row.P2)
	}
}

func TestEventToCSVRow_AspectExact(t *testing.T) {
	evt := models.TransitEvent{
		EventType:       models.EventAspectExact,
		ChartType:       models.ChartTransit,
		TargetChartType: models.ChartNatal,
		Planet:          models.PlanetSun,
		Target:          "MOON",
		JD:              2460310.667,
		PlanetLongitude: 120.0,
		PlanetSign:      "Leo",
		TargetLongitude: 0.0,
		TargetSign:      "Aries",
		AspectType:      models.AspectTrine,
	}
	row := EventToCSVRow(evt, "UTC")
	if row.Aspect != "Trine" {
		t.Errorf("AspectExact CSV Aspect = %q, want Trine", row.Aspect)
	}
	if row.EventType != "Exact" {
		t.Errorf("AspectExact EventType = %q, want Exact", row.EventType)
	}
}

func TestCSVRowToString_Station(t *testing.T) {
	row := CSVRow{
		P1: "Mercury", P1House: 5, Aspect: "Station",
		EventType: "Retrograde", Type: "Tr", Date: "2024-01-01",
		Time: "12:00:00", Timezone: "UTC", Age: 33.5,
		Pos1Deg: 15.5, Pos1Sign: "Leo", Pos1Dir: "Rx",
	}
	s := CSVRowToString(row)
	if !strings.Contains(s, "Mercury") || !strings.Contains(s, "Station") {
		t.Errorf("Station CSV row missing expected content: %s", s)
	}
}

func TestCSVRowToString_Aspect(t *testing.T) {
	row := CSVRow{
		P1: "Sun", P1House: 5, Aspect: "Conjunction",
		P2: "Moon", P2House: 7,
		EventType: "Exact", Type: "Tr-Na", Date: "2024-01-01",
		Time: "12:00:00", Timezone: "UTC", Age: 33.5,
		Pos1Deg: 15.5, Pos1Sign: "Leo", Pos1Dir: "Dir",
		Pos2Deg: 15.3, Pos2Sign: "Leo", Pos2Dir: "Dir",
	}
	s := CSVRowToString(row)
	if !strings.Contains(s, "Sun") || !strings.Contains(s, "Moon") {
		t.Errorf("Aspect CSV row missing expected content: %s", s)
	}
}

func TestEventsToCSV(t *testing.T) {
	events := []models.TransitEvent{
		{
			EventType:       models.EventStation,
			ChartType:       models.ChartTransit,
			Planet:          models.PlanetMercury,
			JD:              2460310.667,
			PlanetLongitude: 120.5,
			StationType:     models.StationRetrograde,
			IsRetrograde:    true,
		},
	}
	csv := EventsToCSV(events, "UTC", "")
	if !strings.Contains(csv, "P1") {
		t.Error("EventsToCSV should include header")
	}
	if !strings.Contains(csv, "Mercury") {
		t.Error("EventsToCSV should include event data")
	}

	// With custom tz label
	csv2 := EventsToCSV(events, "UTC", "AWST")
	if !strings.Contains(csv2, "AWST") {
		t.Error("EventsToCSV should use custom tz label")
	}
}

func TestEventsToJSON(t *testing.T) {
	events := []models.TransitEvent{
		{
			EventType: models.EventStation,
			ChartType: models.ChartTransit,
			Planet:    models.PlanetMercury,
		},
	}
	jsonStr, err := EventsToJSON(events)
	if err != nil {
		t.Fatalf("EventsToJSON error: %v", err)
	}
	if !strings.Contains(jsonStr, "STATION") {
		t.Error("JSON should contain event type")
	}
}

// === Chart export tests ===

func TestChartToCSV(t *testing.T) {
	chart := &models.ChartInfo{
		Planets: []models.PlanetPosition{
			{PlanetID: models.PlanetSun, Longitude: 280.5, Sign: "Capricorn", SignDegree: 10.5, House: 10, Speed: 1.0},
			{PlanetID: models.PlanetMoon, Longitude: 120.3, Sign: "Leo", SignDegree: 0.3, House: 5, Speed: 13.0},
		},
	}
	csv := ChartToCSV(chart)
	if !strings.Contains(csv, "Planet,Longitude") {
		t.Error("Missing header")
	}
	if !strings.Contains(csv, "Sun") {
		t.Error("Missing Sun")
	}
	if !strings.Contains(csv, "Moon") {
		t.Error("Missing Moon")
	}
	if !strings.Contains(csv, "Capricorn") {
		t.Error("Missing sign")
	}
	// Should contain glyph column
	if !strings.Contains(csv, "\u2609") { // Sun glyph
		t.Error("Missing Sun glyph")
	}
}

func TestChartToJSON(t *testing.T) {
	chart := &models.ChartInfo{
		Planets: []models.PlanetPosition{
			{PlanetID: models.PlanetSun, Longitude: 280.5, Sign: "Capricorn"},
		},
	}
	jsonStr, err := ChartToJSON(chart)
	if err != nil {
		t.Fatalf("ChartToJSON error: %v", err)
	}
	if !strings.Contains(jsonStr, "Capricorn") {
		t.Error("Missing sign in JSON")
	}
}

func TestAspectsToCSV(t *testing.T) {
	aspects := []models.AspectInfo{
		{PlanetA: "SUN", PlanetB: "MOON", AspectType: models.AspectTrine, AspectAngle: 120, ActualAngle: 119.5, Orb: 0.5, IsApplying: true},
	}
	csv := AspectsToCSV(aspects)
	if !strings.Contains(csv, "PlanetA,PlanetB") {
		t.Error("Missing header")
	}
	if !strings.Contains(csv, "Trine") {
		t.Error("Missing aspect type")
	}
	if !strings.Contains(csv, "\u25B3") { // △ trine glyph
		t.Error("Missing trine glyph")
	}
}

func TestCrossAspectsToCSV(t *testing.T) {
	aspects := []models.CrossAspectInfo{
		{InnerBody: "SUN", OuterBody: "VENUS", AspectType: models.AspectConjunction, Orb: 1.0},
	}
	csv := CrossAspectsToCSV(aspects)
	if !strings.Contains(csv, "InnerBody,OuterBody") {
		t.Error("Missing header")
	}
	if !strings.Contains(csv, "Conjunction") {
		t.Error("Missing aspect type")
	}
}

func TestHousesToCSV(t *testing.T) {
	houses := make([]float64, 12)
	for i := range houses {
		houses[i] = float64(i) * 30
	}
	angles := models.AnglesInfo{ASC: 15, MC: 285}
	csv := HousesToCSV(houses, angles)
	if !strings.Contains(csv, "House,Longitude") {
		t.Error("Missing header")
	}
	if !strings.Contains(csv, "ASC") {
		t.Error("Missing ASC")
	}
	if !strings.Contains(csv, "MC") {
		t.Error("Missing MC")
	}
}

func TestPositionsToCSV(t *testing.T) {
	positions := []models.PlanetPosition{
		{PlanetID: models.PlanetSun, Longitude: 280.5, Sign: "Capricorn", SignDegree: 10.5, Speed: 1.0},
		{PlanetID: models.PlanetMoon, Longitude: 120.3, Sign: "Leo", SignDegree: 0.3, Speed: 13.0, IsRetrograde: false},
	}
	csv := PositionsToCSV(positions)
	if !strings.Contains(csv, "Planet,Longitude") {
		t.Error("Missing header")
	}
	if !strings.Contains(csv, "Sun") {
		t.Error("Missing Sun")
	}
	if !strings.Contains(csv, "\u2609") { // Sun glyph
		t.Error("Missing Sun glyph")
	}
}

func TestDignityToCSV(t *testing.T) {
	// Use the actual dignity type via JSON round-trip
	dignities := []map[string]interface{}{
		{"planet_id": "SUN", "sign": "Leo", "score": 5, "ruler": "SUN", "exalted": false, "in_detriment": false, "in_fall": false, "dignities": []string{"RULERSHIP"}},
	}
	csv := DignityToCSV(dignities)
	if !strings.Contains(csv, "Planet,Sign,Score") {
		t.Error("Missing header")
	}
	if !strings.Contains(csv, "Sun") {
		t.Error("Missing Sun")
	}
}

func TestLotsToCSV(t *testing.T) {
	lots := []map[string]interface{}{
		{"name": "Lot of Fortune", "longitude": 215.5, "sign": "Scorpio", "sign_degree": 5.5, "formula": "ASC + Moon - Sun"},
	}
	csv := LotsToCSV(lots)
	if !strings.Contains(csv, "Lot,Longitude") {
		t.Error("Missing header")
	}
	if !strings.Contains(csv, "Fortune") {
		t.Error("Missing Fortune")
	}
}

func TestEclipsesToCSV(t *testing.T) {
	eclipses := []map[string]interface{}{
		{"type": "SOLAR_TOTAL", "jd": 2451545.0, "moon_longitude": 280.0, "moon_sign": "Capricorn", "sun_sign": "Capricorn", "moon_latitude": 0.1},
	}
	csv := EclipsesToCSV(eclipses)
	if !strings.Contains(csv, "Type,JD") {
		t.Error("Missing header")
	}
	if !strings.Contains(csv, "SOLAR_TOTAL") {
		t.Error("Missing eclipse type")
	}
}

func TestLunarPhasesToCSV(t *testing.T) {
	phases := []map[string]interface{}{
		{"phase": "NEW_MOON", "jd": 2451545.0, "moon_longitude": 280.0, "moon_sign": "Capricorn", "sun_sign": "Capricorn"},
	}
	csv := LunarPhasesToCSV(phases)
	if !strings.Contains(csv, "Phase,JD") {
		t.Error("Missing header")
	}
	if !strings.Contains(csv, "NEW_MOON") {
		t.Error("Missing phase")
	}
}

func TestToJSON(t *testing.T) {
	data := map[string]int{"a": 1, "b": 2}
	jsonStr, err := ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON error: %v", err)
	}
	if !strings.Contains(jsonStr, "\"a\": 1") {
		t.Error("Missing key in JSON")
	}
}
