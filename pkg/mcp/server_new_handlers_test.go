package mcp

import (
	"encoding/json"
	"testing"
)

const testEphe = "../../third_party/swisseph/ephe"

func newTestServer() *Server {
	return NewServer(testEphe)
}

func TestHandleCalcSolarReturn(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"natal_jd_ut": 2451545.0,
		"natal_latitude": 51.5074,
		"natal_longitude": -0.1278,
		"search_jd_ut": 2451905.0
	}`)
	result, err := s.handleCalcSolarReturn(args)
	if err != nil {
		t.Fatalf("handleCalcSolarReturn: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcSolarReturn_Series(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"natal_jd_ut": 2451545.0,
		"natal_latitude": 51.5074,
		"natal_longitude": -0.1278,
		"search_jd_ut": 2451905.0,
		"count": 2
	}`)
	result, err := s.handleCalcSolarReturn(args)
	if err != nil {
		t.Fatalf("handleCalcSolarReturn series: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcLunarReturn(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"natal_jd_ut": 2451545.0,
		"natal_latitude": 51.5074,
		"natal_longitude": -0.1278,
		"search_jd_ut": 2451570.0
	}`)
	result, err := s.handleCalcLunarReturn(args)
	if err != nil {
		t.Fatalf("handleCalcLunarReturn: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcLunarReturn_Series(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"natal_jd_ut": 2451545.0,
		"natal_latitude": 51.5074,
		"natal_longitude": -0.1278,
		"search_jd_ut": 2451570.0,
		"count": 2
	}`)
	result, err := s.handleCalcLunarReturn(args)
	if err != nil {
		t.Fatalf("handleCalcLunarReturn series: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcDignity(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"latitude": 51.5074,
		"longitude": -0.1278,
		"jd_ut": 2451545.0
	}`)
	result, err := s.handleCalcDignity(args)
	if err != nil {
		t.Fatalf("handleCalcDignity: %v", err)
	}
	m := result.(map[string]interface{})
	if m["dignities"] == nil {
		t.Error("missing dignities field")
	}
	if m["mutual_receptions"] == nil {
		t.Error("missing mutual_receptions field")
	}
	if m["sect"] == nil {
		t.Error("missing sect field")
	}
}

func TestHandleCalcCompositeChart(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"person1_latitude": 51.5074,
		"person1_longitude": -0.1278,
		"person1_jd_ut": 2451545.0,
		"person2_latitude": 40.7128,
		"person2_longitude": -74.006,
		"person2_jd_ut": 2451910.0
	}`)
	result, err := s.handleCalcCompositeChart(args)
	if err != nil {
		t.Fatalf("handleCalcCompositeChart: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcAspectPatterns(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"latitude": 51.5074,
		"longitude": -0.1278,
		"jd_ut": 2451545.0
	}`)
	result, err := s.handleCalcAspectPatterns(args)
	if err != nil {
		t.Fatalf("handleCalcAspectPatterns: %v", err)
	}
	m := result.(map[string]interface{})
	if m["patterns"] == nil {
		t.Error("missing patterns field")
	}
}

func TestHandleCalcLunarPhase(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{"jd_ut": 2451545.0}`)
	result, err := s.handleCalcLunarPhase(args)
	if err != nil {
		t.Fatalf("handleCalcLunarPhase: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcLunarPhases(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"start_jd_ut": 2451545.0,
		"end_jd_ut": 2451575.0
	}`)
	result, err := s.handleCalcLunarPhases(args)
	if err != nil {
		t.Fatalf("handleCalcLunarPhases: %v", err)
	}
	m := result.(map[string]interface{})
	if m["phases"] == nil {
		t.Error("missing phases field")
	}
}

func TestHandleCalcEclipses(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"start_jd_ut": 2451545.0,
		"end_jd_ut": 2451910.0
	}`)
	result, err := s.handleCalcEclipses(args)
	if err != nil {
		t.Fatalf("handleCalcEclipses: %v", err)
	}
	m := result.(map[string]interface{})
	if m["eclipses"] == nil {
		t.Error("missing eclipses field")
	}
}

func TestHandleCalcSynastry(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"person1_latitude": 51.5074,
		"person1_longitude": -0.1278,
		"person1_jd_ut": 2451545.0,
		"person2_latitude": 40.7128,
		"person2_longitude": -74.006,
		"person2_jd_ut": 2451910.0
	}`)
	result, err := s.handleCalcSynastry(args)
	if err != nil {
		t.Fatalf("handleCalcSynastry: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

// Invalid JSON tests for all new handlers
func TestHandleCalcSolarReturn_InvalidJSON(t *testing.T) {
	s := newTestServer()
	_, err := s.handleCalcSolarReturn(json.RawMessage(`{invalid`))
	if err == nil {
		t.Error("expected error")
	}
}

func TestHandleCalcLunarReturn_InvalidJSON(t *testing.T) {
	s := newTestServer()
	_, err := s.handleCalcLunarReturn(json.RawMessage(`{invalid`))
	if err == nil {
		t.Error("expected error")
	}
}

func TestHandleCalcDignity_InvalidJSON(t *testing.T) {
	s := newTestServer()
	_, err := s.handleCalcDignity(json.RawMessage(`{invalid`))
	if err == nil {
		t.Error("expected error")
	}
}

func TestHandleCalcCompositeChart_InvalidJSON(t *testing.T) {
	s := newTestServer()
	_, err := s.handleCalcCompositeChart(json.RawMessage(`{invalid`))
	if err == nil {
		t.Error("expected error")
	}
}

func TestHandleCalcAspectPatterns_InvalidJSON(t *testing.T) {
	s := newTestServer()
	_, err := s.handleCalcAspectPatterns(json.RawMessage(`{invalid`))
	if err == nil {
		t.Error("expected error")
	}
}

func TestHandleCalcLunarPhase_InvalidJSON(t *testing.T) {
	s := newTestServer()
	_, err := s.handleCalcLunarPhase(json.RawMessage(`{invalid`))
	if err == nil {
		t.Error("expected error")
	}
}

func TestHandleCalcLunarPhases_InvalidJSON(t *testing.T) {
	s := newTestServer()
	_, err := s.handleCalcLunarPhases(json.RawMessage(`{invalid`))
	if err == nil {
		t.Error("expected error")
	}
}

func TestHandleCalcEclipses_InvalidJSON(t *testing.T) {
	s := newTestServer()
	_, err := s.handleCalcEclipses(json.RawMessage(`{invalid`))
	if err == nil {
		t.Error("expected error")
	}
}

func TestHandleCalcSynastry_InvalidJSON(t *testing.T) {
	s := newTestServer()
	_, err := s.handleCalcSynastry(json.RawMessage(`{invalid`))
	if err == nil {
		t.Error("expected error")
	}
}

// Test dispatch of new tools via handleToolsCall
func TestHandleToolsCall_NewTools(t *testing.T) {
	s := newTestServer()

	tools := []struct {
		name string
		args string
	}{
		{"calc_dignity", `{"latitude":51.5,"longitude":-0.1,"jd_ut":2451545.0}`},
		{"calc_lunar_phase", `{"jd_ut":2451545.0}`},
		{"calc_aspect_patterns", `{"latitude":51.5,"longitude":-0.1,"jd_ut":2451545.0}`},
		{"calc_composite_chart", `{"person1_latitude":51.5,"person1_longitude":-0.1,"person1_jd_ut":2451545.0,"person2_latitude":40.7,"person2_longitude":-74.0,"person2_jd_ut":2451910.0}`},
		{"calc_solar_return", `{"natal_jd_ut":2451545.0,"natal_latitude":51.5,"natal_longitude":-0.1,"search_jd_ut":2451905.0}`},
		{"calc_lunar_return", `{"natal_jd_ut":2451545.0,"natal_latitude":51.5,"natal_longitude":-0.1,"search_jd_ut":2451570.0}`},
		{"calc_synastry", `{"person1_latitude":51.5,"person1_longitude":-0.1,"person1_jd_ut":2451545.0,"person2_latitude":40.7,"person2_longitude":-74.0,"person2_jd_ut":2451910.0}`},
		{"calc_lunar_phases", `{"start_jd_ut":2451545.0,"end_jd_ut":2451575.0}`},
		{"calc_eclipses", `{"start_jd_ut":2451545.0,"end_jd_ut":2451910.0}`},
	}

	for _, tt := range tools {
		params := callToolParams{
			Name:      tt.name,
			Arguments: json.RawMessage(tt.args),
		}
		paramsJSON, _ := json.Marshal(params)
		req := &jsonRPCRequest{
			JSONRPC: "2.0",
			ID:      float64(1),
			Method:  "tools/call",
			Params:  paramsJSON,
		}
		resp := s.handleRequest(req)
		if resp == nil {
			t.Errorf("%s: nil response", tt.name)
			continue
		}
		if resp.Error != nil {
			t.Errorf("%s: error %s", tt.name, resp.Error.Message)
		}
	}
}

func TestHandleCalcTransit_CSV(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"natal_latitude": 51.5074,
		"natal_longitude": -0.1278,
		"natal_jd_ut": 2451545.0,
		"transit_latitude": 51.5074,
		"transit_longitude": -0.1278,
		"start_jd_ut": 2451545.0,
		"end_jd_ut": 2451560.0,
		"format": "csv",
		"timezone": "Europe/London"
	}`)
	result, err := s.handleCalcTransit(args)
	if err != nil {
		t.Fatalf("handleCalcTransit CSV: %v", err)
	}
	m := result.(map[string]interface{})
	if m["format"] != "csv" {
		t.Errorf("format = %v, want csv", m["format"])
	}
}

func TestHandleCalcTransit_JSON(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"natal_latitude": 51.5074,
		"natal_longitude": -0.1278,
		"natal_jd_ut": 2451545.0,
		"transit_latitude": 51.5074,
		"transit_longitude": -0.1278,
		"start_jd_ut": 2451545.0,
		"end_jd_ut": 2451550.0
	}`)
	result, err := s.handleCalcTransit(args)
	if err != nil {
		t.Fatalf("handleCalcTransit JSON: %v", err)
	}
	m := result.(map[string]interface{})
	if m["events"] == nil {
		t.Error("missing events field")
	}
}

func TestHandleCalcSingleChart_Default(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"latitude": 51.5074,
		"longitude": -0.1278,
		"jd_ut": 2451545.0
	}`)
	result, err := s.handleCalcSingleChart(args)
	if err != nil {
		t.Fatalf("handleCalcSingleChart: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcDoubleChart_Default(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"inner_latitude": 51.5074,
		"inner_longitude": -0.1278,
		"inner_jd_ut": 2451545.0,
		"outer_latitude": 40.7128,
		"outer_longitude": -74.006,
		"outer_jd_ut": 2451910.0
	}`)
	result, err := s.handleCalcDoubleChart(args)
	if err != nil {
		t.Fatalf("handleCalcDoubleChart: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcProgressions_Default(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"natal_jd_ut": 2448000.5,
		"transit_jd_ut": 2451545.0
	}`)
	result, err := s.handleCalcProgressions(args)
	if err != nil {
		t.Fatalf("handleCalcProgressions: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleDatetimeToJD_Default(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{"datetime": "2000-01-01T12:00:00Z"}`)
	result, err := s.handleDatetimeToJD(args)
	if err != nil {
		t.Fatalf("handleDatetimeToJD: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleToolsCall_UnknownTool(t *testing.T) {
	s := newTestServer()
	params := callToolParams{
		Name:      "nonexistent_tool",
		Arguments: json.RawMessage(`{}`),
	}
	paramsJSON, _ := json.Marshal(params)
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      float64(1),
		Method:  "tools/call",
		Params:  paramsJSON,
	}
	resp := s.handleRequest(req)
	if resp == nil || resp.Error == nil {
		t.Error("expected error for unknown tool")
	}
}
