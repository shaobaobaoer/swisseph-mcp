package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestErrorPaths_InvalidJSON verifies that all POST endpoints return 400 for invalid JSON.
func TestErrorPaths_InvalidJSON(t *testing.T) {
	srv := newTestServer("")
	endpoints := []string{
		"/api/v1/geocode",
		"/api/v1/datetime/to-jd",
		"/api/v1/datetime/from-jd",
		"/api/v1/planet/position",
		"/api/v1/chart/natal",
		"/api/v1/chart/double",
		"/api/v1/chart/composite",
		"/api/v1/chart/davison",
		"/api/v1/chart/harmonic",
		"/api/v1/chart/wheel",
		"/api/v1/transit",
		"/api/v1/progressions",
		"/api/v1/solar-arc",
		"/api/v1/primary-directions",
		"/api/v1/symbolic-directions",
		"/api/v1/solar-return",
		"/api/v1/lunar-return",
		"/api/v1/dignity",
		"/api/v1/bonification",
		"/api/v1/dispositors",
		"/api/v1/profection",
		"/api/v1/firdaria",
		"/api/v1/lots",
		"/api/v1/bounds",
		"/api/v1/antiscia",
		"/api/v1/planetary-hours",
		"/api/v1/heliacal",
		"/api/v1/aspects/patterns",
		"/api/v1/fixed-stars",
		"/api/v1/midpoints",
		"/api/v1/synastry",
		"/api/v1/lunar/phase",
		"/api/v1/lunar/phases",
		"/api/v1/lunar/eclipses",
		"/api/v1/report/natal",
	}

	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, ep, bytes.NewBufferString("{invalid"))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected 400 for invalid JSON at %s, got %d", ep, rec.Code)
			}
		})
	}
}

// TestErrorPaths_EmptyBody verifies that all POST endpoints return 400 for nil body.
func TestErrorPaths_EmptyBody(t *testing.T) {
	srv := newTestServer("")
	endpoints := []string{
		"/api/v1/geocode",
		"/api/v1/datetime/to-jd",
		"/api/v1/datetime/from-jd",
		"/api/v1/planet/position",
		"/api/v1/chart/natal",
		"/api/v1/chart/double",
		"/api/v1/chart/composite",
		"/api/v1/chart/davison",
		"/api/v1/chart/harmonic",
		"/api/v1/chart/wheel",
		"/api/v1/transit",
		"/api/v1/progressions",
		"/api/v1/solar-arc",
		"/api/v1/primary-directions",
		"/api/v1/symbolic-directions",
		"/api/v1/solar-return",
		"/api/v1/lunar-return",
		"/api/v1/dignity",
		"/api/v1/bonification",
		"/api/v1/dispositors",
		"/api/v1/profection",
		"/api/v1/firdaria",
		"/api/v1/lots",
		"/api/v1/bounds",
		"/api/v1/antiscia",
		"/api/v1/planetary-hours",
		"/api/v1/heliacal",
		"/api/v1/aspects/patterns",
		"/api/v1/fixed-stars",
		"/api/v1/midpoints",
		"/api/v1/synastry",
		"/api/v1/lunar/phase",
		"/api/v1/lunar/phases",
		"/api/v1/lunar/eclipses",
		"/api/v1/report/natal",
	}

	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, ep, nil)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected 400 for empty body at %s, got %d", ep, rec.Code)
			}
		})
	}
}

// TestErrorPaths_MethodNotAllowed verifies all POST endpoints reject GET.
func TestErrorPaths_MethodNotAllowed(t *testing.T) {
	srv := newTestServer("")
	endpoints := []string{
		"/api/v1/chart/double",
		"/api/v1/chart/composite",
		"/api/v1/chart/davison",
		"/api/v1/chart/harmonic",
		"/api/v1/chart/wheel",
		"/api/v1/transit",
		"/api/v1/primary-directions",
		"/api/v1/symbolic-directions",
		"/api/v1/solar-return",
		"/api/v1/lunar-return",
		"/api/v1/bonification",
		"/api/v1/dispositors",
		"/api/v1/profection",
		"/api/v1/lots",
		"/api/v1/bounds",
		"/api/v1/antiscia",
		"/api/v1/planetary-hours",
		"/api/v1/heliacal",
		"/api/v1/aspects/patterns",
		"/api/v1/fixed-stars",
		"/api/v1/midpoints",
		"/api/v1/synastry",
		"/api/v1/lunar/phases",
		"/api/v1/lunar/eclipses",
		"/api/v1/report/natal",
	}

	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, ep, nil)
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)
			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected 405 at %s, got %d", ep, rec.Code)
			}
		})
	}
}

// TestOrbOrDefault_WithCustomOrb verifies custom orb config is used.
func TestOrbOrDefault_WithCustomOrb(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/chart/natal", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
		"orb_config": map[string]interface{}{
			"conjunction": 10.0,
			"opposition":  8.0,
			"trine":       6.0,
			"square":      6.0,
			"sextile":     4.0,
		},
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestGeocode_EmptyLocationName verifies 400 for empty location_name.
func TestGeocode_EmptyLocationName(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/geocode", map[string]interface{}{
		"location_name": "",
	}, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty location_name, got %d", rec.Code)
	}
}

// TestDatetimeToJD_EmptyDatetime verifies 400 for empty datetime.
func TestDatetimeToJD_EmptyDatetime(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/datetime/to-jd", map[string]interface{}{
		"datetime": "",
	}, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty datetime, got %d", rec.Code)
	}
}

// TestDatetimeToJD_InvalidFormat verifies 422 for bad datetime format.
func TestDatetimeToJD_InvalidFormat(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/datetime/to-jd", map[string]interface{}{
		"datetime": "not-a-date",
	}, nil)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for bad datetime, got %d", rec.Code)
	}
}

// TestJDToDatetime_InvalidTimezone verifies 422 for bad timezone.
func TestJDToDatetime_InvalidTimezone(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/datetime/from-jd", map[string]interface{}{
		"jd":       2451545.0,
		"timezone": "Invalid/Zone",
	}, nil)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for bad timezone, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestPlanetPosition_InvalidPlanet verifies 422 for bad planet.
func TestPlanetPosition_InvalidPlanet(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/planet/position", map[string]interface{}{
		"planet": "INVALID_PLANET",
		"jd_ut":  2451545.0,
	}, nil)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for invalid planet, got %d: %s", rec.Code, rec.Body.String())
	}
}
