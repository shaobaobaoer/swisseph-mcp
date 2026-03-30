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

// TestTransit tests the /api/v1/events endpoint.
func TestTransit(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/events", map[string]interface{}{
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
	rec := doPost(t, srv, "/api/v1/events", map[string]interface{}{
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

// TestSolarReturn tests the /api/v1/chart/solar-return endpoint.
func TestSolarReturn(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/chart/solar-return", map[string]interface{}{
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
	rec := doPost(t, srv, "/api/v1/chart/solar-return", map[string]interface{}{
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
	rec := doPost(t, srv, "/api/v1/chart/solar-return", map[string]interface{}{
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

// TestLunarReturn tests the /api/v1/chart/lunar-return endpoint.
func TestLunarReturn(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/chart/lunar-return", map[string]interface{}{
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
	rec := doPost(t, srv, "/api/v1/chart/lunar-return", map[string]interface{}{
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
	rec := doPost(t, srv, "/api/v1/chart/lunar-return", map[string]interface{}{
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
	rec := doPost(t, srv, "/api/v1/analysis/dignity", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
		"format":    "csv",
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestBonification tests the /api/v1/analysis/bonification endpoint.
func TestBonification(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/analysis/bonification", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestAspectPatterns tests the /api/v1/analysis/aspects endpoint.
func TestAspectPatterns(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/analysis/aspects", map[string]interface{}{
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

// TestSynastry tests the /api/v1/analysis/synastry endpoint.
func TestSynastry(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/analysis/synastry", map[string]interface{}{
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

// TestProgressions_CSV tests CSV format for progressions.
func TestProgressions_CSV(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/chart/progression", map[string]interface{}{
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
	rec := doPost(t, srv, "/api/v1/chart/solar-arc", map[string]interface{}{
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
