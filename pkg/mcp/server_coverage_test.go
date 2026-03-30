package mcp

import (
	"encoding/json"
	"testing"
)

func TestHandleCalcTransit_CSVFormat(t *testing.T) {
	s := NewServer(".")
	args, _ := json.Marshal(map[string]interface{}{
		"natal_latitude":   39.9042,
		"natal_longitude":  116.4074,
		"natal_jd_ut":      2448057.5208,
		"transit_latitude":  39.9042,
		"transit_longitude": 116.4074,
		"start_jd_ut":      2460310.667,
		"end_jd_ut":        2460315.667,
		"transit_planets":  []string{"SUN"},
		"format":           "csv",
		"timezone":         "Asia/Shanghai",
		"event_config": map[string]bool{
			"include_tr_na":        true,
			"include_sign_ingress": true,
		},
	})
	params, _ := json.Marshal(callToolParams{Name: "calc_transit", Arguments: args})
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      20,
		Method:  "tools/call",
		Params:  params,
	}
	resp := s.handleRequest(req)
	if resp == nil || resp.Error != nil {
		t.Fatal("calc_transit CSV format failed")
	}
	result := resp.Result.(callToolResult)
	if result.IsError {
		t.Fatalf("Tool error: %s", result.Content[0].Text)
	}
}

func TestBuildTransitInput_Defaults(t *testing.T) {
	s := NewServer(".")
	args, _ := json.Marshal(map[string]interface{}{
		"natal_latitude":   39.9,
		"natal_longitude":  116.4,
		"natal_jd_ut":      2448057.5,
		"transit_latitude":  39.9,
		"transit_longitude": 116.4,
		"start_jd_ut":      2460310.0,
		"end_jd_ut":        2460315.0,
	})
	input, tz, err := s.buildTransitInput(args)
	if err != nil {
		t.Fatalf("buildTransitInput error: %v", err)
	}
	if tz != "UTC" {
		t.Errorf("Default timezone = %q, want UTC", tz)
	}
	if len(input.NatalChart.Planets) != 10 {
		t.Errorf("Default natal planets = %d, want 10", len(input.NatalChart.Planets))
	}
	if len(input.Charts.Transit.Planets) != 10 {
		t.Errorf("Default transit planets = %d, want 10", len(input.Charts.Transit.Planets))
	}
}

func TestBuildTransitInput_CustomOrbs(t *testing.T) {
	s := NewServer(".")
	args, _ := json.Marshal(map[string]interface{}{
		"natal_latitude":   39.9,
		"natal_longitude":  116.4,
		"natal_jd_ut":      2448057.5,
		"transit_latitude":  39.9,
		"transit_longitude": 116.4,
		"start_jd_ut":      2460310.0,
		"end_jd_ut":        2460315.0,
		"orb_config":       map[string]float64{"conjunction": 10},
		"orb_config_transit": map[string]float64{"conjunction": 5},
		"orb_config_progressions": map[string]float64{"conjunction": 3},
		"orb_config_solar_arc": map[string]float64{"conjunction": 2},
		"timezone":         "Asia/Shanghai",
		"house_system":     "KOCH",
		"event_config":     map[string]bool{"include_tr_na": true},
	})
	input, tz, err := s.buildTransitInput(args)
	if err != nil {
		t.Fatalf("buildTransitInput error: %v", err)
	}
	if tz != "Asia/Shanghai" {
		t.Errorf("Timezone = %q, want Asia/Shanghai", tz)
	}
	if input.Charts.Transit.Orbs.Conjunction != 5 {
		t.Errorf("Transit orb conjunction = %f, want 5", input.Charts.Transit.Orbs.Conjunction)
	}
}

func TestBuildTransitInput_InvalidJSON(t *testing.T) {
	s := NewServer(".")
	_, _, err := s.buildTransitInput(json.RawMessage(`{invalid}`))
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestHandleToolsCall_InvalidParams(t *testing.T) {
	s := NewServer(".")
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      30,
		Method:  "tools/call",
		Params:  json.RawMessage(`{invalid}`),
	}
	resp := s.handleRequest(req)
	if resp == nil || resp.Error == nil {
		t.Error("Expected error for invalid params")
	}
}

func TestHandleCalcSingleChart_WithCustomPlanets(t *testing.T) {
	s := NewServer(".")
	args, _ := json.Marshal(map[string]interface{}{
		"latitude":     39.9,
		"longitude":    116.4,
		"jd_ut":        2451545.0,
		"planets":      []string{"SUN", "MOON", "MERCURY"},
		"house_system": "KOCH",
		"orb_config":   map[string]float64{"conjunction": 10, "opposition": 10},
	})
	params, _ := json.Marshal(callToolParams{Name: "calc_single_chart", Arguments: args})
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      31,
		Method:  "tools/call",
		Params:  params,
	}
	resp := s.handleRequest(req)
	if resp == nil || resp.Error != nil {
		t.Fatal("calc_single_chart with custom planets failed")
	}
}

func TestHandleCalcDoubleChart_WithCustomParams(t *testing.T) {
	s := NewServer(".")
	args, _ := json.Marshal(map[string]interface{}{
		"inner_latitude":  39.9,
		"inner_longitude": 116.4,
		"inner_jd_ut":     2448057.5,
		"inner_planets":   []string{"SUN"},
		"outer_latitude":  39.9,
		"outer_longitude": 116.4,
		"outer_jd_ut":     2451545.0,
		"outer_planets":   []string{"MOON"},
		"house_system":    "WHOLE_SIGN",
		"orb_config":      map[string]float64{"conjunction": 10},
	})
	params, _ := json.Marshal(callToolParams{Name: "calc_double_chart", Arguments: args})
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      32,
		Method:  "tools/call",
		Params:  params,
	}
	resp := s.handleRequest(req)
	if resp == nil || resp.Error != nil {
		t.Fatal("calc_double_chart with custom params failed")
	}
}

func TestHandleCalcProgressions_CustomPlanets(t *testing.T) {
	s := NewServer(".")
	args, _ := json.Marshal(map[string]interface{}{
		"natal_jd_ut":   2448057.5208,
		"transit_jd_ut": 2460310.667,
		"planets":       []string{"SUN", "MOON"},
	})
	params, _ := json.Marshal(callToolParams{Name: "calc_progressions", Arguments: args})
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      33,
		Method:  "tools/call",
		Params:  params,
	}
	resp := s.handleRequest(req)
	if resp == nil || resp.Error != nil {
		t.Fatal("calc_progressions with custom planets failed")
	}
}

func TestHandleCalcSolarArc_CustomPlanets(t *testing.T) {
	s := NewServer(".")
	args, _ := json.Marshal(map[string]interface{}{
		"natal_jd_ut":   2448057.5208,
		"transit_jd_ut": 2460310.667,
		"planets":       []string{"SUN", "MOON"},
	})
	params, _ := json.Marshal(callToolParams{Name: "calc_solar_arc", Arguments: args})
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      34,
		Method:  "tools/call",
		Params:  params,
	}
	resp := s.handleRequest(req)
	if resp == nil || resp.Error != nil {
		t.Fatal("calc_solar_arc with custom planets failed")
	}
}
