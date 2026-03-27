package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/shaobaobaoer/solarsage-mcp/pkg/sweph"
)

func TestMain(m *testing.M) {
	ephePath, _ := filepath.Abs("../../third_party/swisseph/ephe")
	sweph.Init(ephePath)
	defer sweph.Close()
	os.Exit(m.Run())
}

func newTestServer(apiKey string) *Server {
	return NewServer(Config{APIKey: apiKey})
}

func doPost(t *testing.T, srv *Server, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestHealthEndpoint(t *testing.T) {
	srv := newTestServer("")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result map[string]string
	json.NewDecoder(rec.Body).Decode(&result)
	if result["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", result)
	}
}

func TestCORSHeaders(t *testing.T) {
	srv := newTestServer("")
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/chart/natal", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for OPTIONS, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected CORS origin *, got %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("expected CORS methods header")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Fatal("expected CORS headers header")
	}
}

func TestAPIKeyAuth_Rejected(t *testing.T) {
	srv := newTestServer("secret123")
	rec := doPost(t, srv, "/api/v1/chart/natal", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
	}, nil)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without API key, got %d", rec.Code)
	}
}

func TestAPIKeyAuth_WrongKey(t *testing.T) {
	srv := newTestServer("secret123")
	rec := doPost(t, srv, "/api/v1/chart/natal", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
	}, map[string]string{"X-API-Key": "wrongkey"})

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong API key, got %d", rec.Code)
	}
}

func TestAPIKeyAuth_Accepted(t *testing.T) {
	srv := newTestServer("secret123")
	rec := doPost(t, srv, "/api/v1/chart/natal", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
	}, map[string]string{"X-API-Key": "secret123"})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with correct API key, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestNatalChart(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/chart/natal", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %q", ct)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := result["planets"]; !ok {
		t.Fatal("expected planets in response")
	}
}

func TestNatalChart_InvalidJSON(t *testing.T) {
	srv := newTestServer("")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chart/natal", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	srv := newTestServer("")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chart/natal", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET on POST-only endpoint, got %d", rec.Code)
	}
}

func TestPlanetPosition(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/planet/position", map[string]interface{}{
		"planet": "SUN",
		"jd_ut":  2451545.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if _, ok := result["longitude"]; !ok {
		t.Fatal("expected longitude in response")
	}
}

func TestDatetimeToJD(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/datetime/to-jd", map[string]interface{}{
		"datetime": "2000-01-01T12:00:00Z",
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestJDToDatetime(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/datetime/from-jd", map[string]interface{}{
		"jd": 2451545.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLunarPhase(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/lunar/phase", map[string]interface{}{
		"jd_ut": 2451545.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDignity(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/dignity", map[string]interface{}{
		"latitude":  51.5,
		"longitude": -0.1,
		"jd_ut":     2451545.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&result)
	if _, ok := result["dignities"]; !ok {
		t.Fatal("expected dignities in response")
	}
}

func TestFirdaria(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/firdaria", map[string]interface{}{
		"is_day_birth": true,
		"age":          30.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProgressions(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/progressions", map[string]interface{}{
		"natal_jd_ut":   2451545.0,
		"transit_jd_ut": 2460000.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSolarArc(t *testing.T) {
	srv := newTestServer("")
	rec := doPost(t, srv, "/api/v1/solar-arc", map[string]interface{}{
		"natal_jd_ut":   2451545.0,
		"transit_jd_ut": 2460000.0,
	}, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMissingRequiredField(t *testing.T) {
	srv := newTestServer("")
	// Geocode requires location_name
	rec := doPost(t, srv, "/api/v1/geocode", map[string]interface{}{}, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing required field, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestEmptyBody(t *testing.T) {
	srv := newTestServer("")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chart/natal", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty body, got %d", rec.Code)
	}
}
