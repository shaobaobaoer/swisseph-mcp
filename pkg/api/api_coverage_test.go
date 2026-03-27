package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestDoubleChart tests the /api/v1/chart/double endpoint.
func TestDoubleChart(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/chart/double", map[string]interface{}{
		"inner_latitude":  51.5,
		"inner_longitude": -0.1,
		"inner_jd_ut":     2451545.0,
		"outer_latitude":  48.85,
		"outer_longitude": 2.35,
		"outer_jd_ut":     2451600.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if _, ok := result["inner_chart"]; !ok {
		t.Fatal("expected inner_chart in response")
	}
	if _, ok := result["cross_aspects"]; !ok {
		t.Fatal("expected cross_aspects in response")
	}
}

// TestDoubleChart_CSV tests CSV format for double chart.
func TestDoubleChart_CSV(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/chart/double", map[string]interface{}{
		"inner_latitude":  51.5,
		"inner_longitude": -0.1,
		"inner_jd_ut":     2451545.0,
		"outer_latitude":  48.85,
		"outer_longitude": 2.35,
		"outer_jd_ut":     2451600.0,
		"format":          "csv",
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if result["format"] != "csv" {
		t.Fatal("expected csv format")
	}
}

// TestCompositeChart tests the /api/v1/chart/composite endpoint.
func TestCompositeChart(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/chart/composite", map[string]interface{}{
		"person1_latitude":  51.5,
		"person1_longitude": -0.1,
		"person1_jd_ut":     2451545.0,
		"person2_latitude":  48.85,
		"person2_longitude": 2.35,
		"person2_jd_ut":     2451600.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestCompositeChart_CSV tests CSV format for composite chart.
func TestCompositeChart_CSV(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/chart/composite", map[string]interface{}{
		"person1_latitude":  51.5,
		"person1_longitude": -0.1,
		"person1_jd_ut":     2451545.0,
		"person2_latitude":  48.85,
		"person2_longitude": 2.35,
		"person2_jd_ut":     2451600.0,
		"format":            "csv",
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestDavisonChart tests the /api/v1/chart/davison endpoint.
func TestDavisonChart(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/chart/davison", map[string]interface{}{
		"person1_latitude":  51.5,
		"person1_longitude": -0.1,
		"person1_jd_ut":     2451545.0,
		"person2_latitude":  48.85,
		"person2_longitude": 2.35,
		"person2_jd_ut":     2451600.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestHarmonicChart tests the /api/v1/chart/harmonic endpoint.
func TestHarmonicChart(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/chart/harmonic", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
		"harmonic":  7,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestHarmonicChart_CSV tests CSV format for harmonic chart.
func TestHarmonicChart_CSV(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/chart/harmonic", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
		"harmonic":  7,
		"format":    "csv",
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestChartWheel tests the /api/v1/chart/wheel endpoint.
func TestChartWheel(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/chart/wheel", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
		"radius":    300.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if _, ok := result["wheel"]; !ok {
		t.Fatal("expected wheel in response")
	}
	if _, ok := result["signs"]; !ok {
		t.Fatal("expected signs in response")
	}
}

// TestNatalChart_CSV tests CSV format for natal chart.
func TestNatalChart_CSV(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/chart/natal", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
		"format":    "csv",
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if result["format"] != "csv" {
		t.Fatal("expected csv format")
	}
}

// TestTransit tests the /api/v1/transit endpoint.
func TestTransit(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/transit", map[string]interface{}{
		"natal_latitude":  51.5,
		"natal_longitude": -0.1,
		"natal_jd_ut":     2451545.0,
		"start_jd_ut":     2451545.0,
		"end_jd_ut":       2451575.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if _, ok := result["events"]; !ok {
		t.Fatal("expected events in response")
	}
}

// TestTransit_CSV tests CSV format for transit.
func TestTransit_CSV(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/transit", map[string]interface{}{
		"natal_latitude":  51.5,
		"natal_longitude": -0.1,
		"natal_jd_ut":     2451545.0,
		"start_jd_ut":     2451545.0,
		"end_jd_ut":       2451575.0,
		"format":          "csv",
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if result["format"] != "csv" {
		t.Fatal("expected csv format")
	}
}

// TestPrimaryDirections tests the /api/v1/primary-directions endpoint.
func TestPrimaryDirections(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/primary-directions", map[string]interface{}{
		"latitude":    51.5,
		"longitude":   -0.1,
		"natal_jd_ut": 2451545.0,
		"max_age":     50.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestSymbolicDirections tests the /api/v1/symbolic-directions endpoint.
func TestSymbolicDirections(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/symbolic-directions", map[string]interface{}{
		"latitude":    51.5,
		"longitude":   -0.1,
		"natal_jd_ut": 2451545.0,
		"age":         30.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestSolarReturn tests the /api/v1/solar-return endpoint.
func TestSolarReturn(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/solar-return", map[string]interface{}{
		"natal_jd_ut":    2451545.0,
		"natal_latitude": 51.5,
		"natal_longitude": -0.1,
		"search_jd_ut":   2451900.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestSolarReturn_CSV tests CSV format for solar return.
func TestSolarReturn_CSV(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/solar-return", map[string]interface{}{
		"natal_jd_ut":     2451545.0,
		"natal_latitude":  51.5,
		"natal_longitude": -0.1,
		"search_jd_ut":    2451900.0,
		"format":          "csv",
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestSolarReturn_Series tests multiple solar returns.
func TestSolarReturn_Series(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/solar-return", map[string]interface{}{
		"natal_jd_ut":     2451545.0,
		"natal_latitude":  51.5,
		"natal_longitude": -0.1,
		"search_jd_ut":    2451900.0,
		"count":           3,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestLunarReturn tests the /api/v1/lunar-return endpoint.
func TestLunarReturn(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/lunar-return", map[string]interface{}{
		"natal_jd_ut":     2451545.0,
		"natal_latitude":  51.5,
		"natal_longitude": -0.1,
		"search_jd_ut":    2451575.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestLunarReturn_CSV tests CSV format for lunar return.
func TestLunarReturn_CSV(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/lunar-return", map[string]interface{}{
		"natal_jd_ut":     2451545.0,
		"natal_latitude":  51.5,
		"natal_longitude": -0.1,
		"search_jd_ut":    2451575.0,
		"format":          "csv",
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestLunarReturn_Series tests multiple lunar returns.
func TestLunarReturn_Series(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/lunar-return", map[string]interface{}{
		"natal_jd_ut":     2451545.0,
		"natal_latitude":  51.5,
		"natal_longitude": -0.1,
		"search_jd_ut":    2451575.0,
		"count":           3,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestDignity_CSV tests CSV format for dignity.
func TestDignity_CSV(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/dignity", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
		"format":    "csv",
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestBonification tests the /api/v1/bonification endpoint.
func TestBonification(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/bonification", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestDispositors tests the /api/v1/dispositors endpoint.
func TestDispositors(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/dispositors", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestDispositors_Traditional tests traditional dispositors.
func TestDispositors_Traditional(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/dispositors", map[string]interface{}{
		"latitude":    51.5,
		"longitude":   -0.1,
		"jd_ut":       2451545.0,
		"traditional": true,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestProfection tests the /api/v1/profection endpoint.
func TestProfection(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/profection", map[string]interface{}{
		"latitude":    51.5,
		"longitude":   -0.1,
		"natal_jd_ut": 2451545.0,
		"age":         30,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if _, ok := result["annual"]; !ok {
		t.Fatal("expected annual in response")
	}
}

// TestProfection_WithMonthly tests profection with monthly included.
func TestProfection_WithMonthly(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/profection", map[string]interface{}{
		"latitude":        51.5,
		"longitude":       -0.1,
		"natal_jd_ut":     2451545.0,
		"age":             30,
		"include_monthly": true,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if _, ok := result["monthly"]; !ok {
		t.Fatal("expected monthly in response")
	}
}

// TestProfection_WithTimeline tests profection with timeline.
func TestProfection_WithTimeline(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/profection", map[string]interface{}{
		"latitude":    51.5,
		"longitude":   -0.1,
		"natal_jd_ut": 2451545.0,
		"age":         30,
		"start_age":   25,
		"end_age":     35,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if _, ok := result["timeline"]; !ok {
		t.Fatal("expected timeline in response")
	}
}

// TestFirdaria_WithTimeline tests firdaria with timeline.
func TestFirdaria_WithTimeline(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/firdaria", map[string]interface{}{
		"is_day_birth": true,
		"age":          30.0,
		"start_age":    25.0,
		"end_age":      40.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestLots tests the /api/v1/lots endpoint.
func TestLots(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/lots", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if _, ok := result["lots"]; !ok {
		t.Fatal("expected lots in response")
	}
}

// TestLots_CSV tests CSV format for lots.
func TestLots_CSV(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/lots", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
		"format":    "csv",
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestBounds tests the /api/v1/bounds endpoint.
func TestBounds(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/bounds", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if _, ok := result["faces"]; !ok {
		t.Fatal("expected faces in response")
	}
}

// TestAntiscia tests the /api/v1/antiscia endpoint.
func TestAntiscia(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/antiscia", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
		"orb":       2.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if _, ok := result["antiscia_points"]; !ok {
		t.Fatal("expected antiscia_points in response")
	}
}

// TestPlanetaryHours tests the /api/v1/planetary-hours endpoint.
func TestPlanetaryHours(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/planetary-hours", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestHeliacal tests the /api/v1/heliacal endpoint.
func TestHeliacal(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/heliacal", map[string]interface{}{
		"latitude":    51.5,
		"longitude":   -0.1,
		"altitude":    0.0,
		"start_jd_ut": 2451545.0,
		"end_jd_ut":   2451600.0,
		"planets":     []string{"VENUS"},
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestAspectPatterns tests the /api/v1/aspects/patterns endpoint.
func TestAspectPatterns(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/aspects/patterns", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if _, ok := result["patterns"]; !ok {
		t.Fatal("expected patterns in response")
	}
	if _, ok := result["aspects"]; !ok {
		t.Fatal("expected aspects in response")
	}
}

// TestFixedStars tests the /api/v1/fixed-stars endpoint.
func TestFixedStars(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/fixed-stars", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
		"orb":       1.5,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if _, ok := result["conjunctions"]; !ok {
		t.Fatal("expected conjunctions in response")
	}
}

// TestMidpoints tests the /api/v1/midpoints endpoint.
func TestMidpoints(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/midpoints", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestSynastry tests the /api/v1/synastry endpoint.
func TestSynastry(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/synastry", map[string]interface{}{
		"person1_latitude":  51.5,
		"person1_longitude": -0.1,
		"person1_jd_ut":     2451545.0,
		"person2_latitude":  48.85,
		"person2_longitude": 2.35,
		"person2_jd_ut":     2451600.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestLunarPhases tests the /api/v1/lunar/phases endpoint.
func TestLunarPhases(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/lunar/phases", map[string]interface{}{
		"start_jd_ut": 2451545.0,
		"end_jd_ut":   2451575.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if _, ok := result["phases"]; !ok {
		t.Fatal("expected phases in response")
	}
}

// TestLunarPhases_CSV tests CSV format for lunar phases.
func TestLunarPhases_CSV(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/lunar/phases", map[string]interface{}{
		"start_jd_ut": 2451545.0,
		"end_jd_ut":   2451575.0,
		"format":      "csv",
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestEclipses tests the /api/v1/lunar/eclipses endpoint.
func TestEclipses(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/lunar/eclipses", map[string]interface{}{
		"start_jd_ut": 2451545.0,
		"end_jd_ut":   2451910.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if _, ok := result["eclipses"]; !ok {
		t.Fatal("expected eclipses in response")
	}
}

// TestEclipses_CSV tests CSV format for eclipses.
func TestEclipses_CSV(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/lunar/eclipses", map[string]interface{}{
		"start_jd_ut": 2451545.0,
		"end_jd_ut":   2451910.0,
		"format":      "csv",
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestNatalReport tests the /api/v1/report/natal endpoint.
func TestNatalReport(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/report/natal", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestProgressions_CSV tests CSV format for progressions.
func TestProgressions_CSV(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/progressions", map[string]interface{}{
		"natal_jd_ut":   2451545.0,
		"transit_jd_ut": 2460000.0,
		"format":        "csv",
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestSolarArc_CSV tests CSV format for solar arc.
func TestSolarArc_CSV(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/solar-arc", map[string]interface{}{
		"natal_jd_ut":   2451545.0,
		"transit_jd_ut": 2460000.0,
		"format":        "csv",
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestNotFound tests an unknown route.
func TestNotFound(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/nonexistent", map[string]interface{}{}, nil)

	if rec.Code == http.StatusOK {
		t.Fatal("expected non-200 for unknown route")
	}
}

// TestAPIKeyAuth_HealthNoAuth tests that health endpoint doesn't require auth
// (it's a GET and goes through the mux, but auth is checked before mux).
func TestAPIKeyAuth_HealthBypass(t *testing.T) {
	srv := newTestServer("secret123")
	// Health is GET but API key is checked for all requests
	rec := doPost(t, srv, "/api/v1/health", nil, map[string]string{"X-API-Key": "secret123"})

	// With correct key, any path should work through the auth layer
	if rec.Code == http.StatusUnauthorized {
		t.Fatal("should not be unauthorized with correct key")
	}
}
