package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/anthropic/swisseph-mcp/pkg/sweph"
)

func TestMain(m *testing.M) {
	ephePath := filepath.Join("..", "..", "third_party", "swisseph", "ephe")
	sweph.Init(ephePath)
	code := m.Run()
	sweph.Close()
	os.Exit(code)
}

func TestHandleInitialize(t *testing.T) {
	s := NewServer(".")
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}
	resp := s.handleRequest(req)
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.Error != nil {
		t.Fatalf("Initialize error: %v", resp.Error)
	}
	result, ok := resp.Result.(initializeResult)
	if !ok {
		t.Fatal("Result is not initializeResult")
	}
	if result.ServerInfo.Name != "swisseph-mcp" {
		t.Errorf("Server name = %q, want swisseph-mcp", result.ServerInfo.Name)
	}
}

func TestHandleToolsList(t *testing.T) {
	s := NewServer(".")
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}
	resp := s.handleRequest(req)
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.Error != nil {
		t.Fatalf("tools/list error: %v", resp.Error)
	}
	result, ok := resp.Result.(toolsListResult)
	if !ok {
		t.Fatal("Result is not toolsListResult")
	}
	expectedTools := []string{"geocode", "datetime_to_jd", "jd_to_datetime", "calc_planet_position", "calc_single_chart", "calc_double_chart", "calc_progressions", "calc_solar_arc", "calc_transit"}
	if len(result.Tools) != len(expectedTools) {
		t.Errorf("Expected %d tools, got %d", len(expectedTools), len(result.Tools))
	}
	toolNames := make(map[string]bool)
	for _, tool := range result.Tools {
		toolNames[tool.Name] = true
	}
	for _, name := range expectedTools {
		if !toolNames[name] {
			t.Errorf("Missing tool: %s", name)
		}
	}
}

func TestHandleGeocode(t *testing.T) {
	s := NewServer(".")
	args, _ := json.Marshal(map[string]string{"location_name": "Beijing"})
	params, _ := json.Marshal(callToolParams{Name: "geocode", Arguments: args})
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params:  params,
	}
	resp := s.handleRequest(req)
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.Error != nil {
		t.Fatalf("geocode error: %v", resp.Error)
	}

	result, ok := resp.Result.(callToolResult)
	if !ok {
		t.Fatal("Result is not callToolResult")
	}
	if result.IsError {
		t.Fatalf("Tool returned error: %s", result.Content[0].Text)
	}
	if len(result.Content) == 0 {
		t.Fatal("Empty content")
	}
}

func TestHandleDatetimeToJD(t *testing.T) {
	s := NewServer(".")
	args, _ := json.Marshal(map[string]string{"datetime": "2000-01-01T12:00:00+00:00"})
	params, _ := json.Marshal(callToolParams{Name: "datetime_to_jd", Arguments: args})
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "tools/call",
		Params:  params,
	}
	resp := s.handleRequest(req)
	if resp == nil || resp.Error != nil {
		t.Fatal("datetime_to_jd failed")
	}
	result := resp.Result.(callToolResult)
	if result.IsError {
		t.Fatalf("Tool error: %s", result.Content[0].Text)
	}
}

func TestHandleCalcSingleChart(t *testing.T) {
	s := NewServer(".")
	args, _ := json.Marshal(map[string]interface{}{
		"latitude":  39.9042,
		"longitude": 116.4074,
		"jd_ut":     2451545.0,
	})
	params, _ := json.Marshal(callToolParams{Name: "calc_single_chart", Arguments: args})
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "tools/call",
		Params:  params,
	}
	resp := s.handleRequest(req)
	if resp == nil || resp.Error != nil {
		t.Fatal("calc_single_chart failed")
	}
	result := resp.Result.(callToolResult)
	if result.IsError {
		t.Fatalf("Tool error: %s", result.Content[0].Text)
	}
}

func TestHandleCalcProgressions(t *testing.T) {
	s := NewServer(".")
	args, _ := json.Marshal(map[string]interface{}{
		"natal_jd_ut":   2448057.5208,
		"transit_jd_ut": 2460310.667,
	})
	params, _ := json.Marshal(callToolParams{Name: "calc_progressions", Arguments: args})
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      9,
		Method:  "tools/call",
		Params:  params,
	}
	resp := s.handleRequest(req)
	if resp == nil || resp.Error != nil {
		t.Fatal("calc_progressions failed")
	}
	result := resp.Result.(callToolResult)
	if result.IsError {
		t.Fatalf("Tool error: %s", result.Content[0].Text)
	}
}

func TestHandleCalcSolarArc(t *testing.T) {
	s := NewServer(".")
	args, _ := json.Marshal(map[string]interface{}{
		"natal_jd_ut":   2448057.5208,
		"transit_jd_ut": 2460310.667,
	})
	params, _ := json.Marshal(callToolParams{Name: "calc_solar_arc", Arguments: args})
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      10,
		Method:  "tools/call",
		Params:  params,
	}
	resp := s.handleRequest(req)
	if resp == nil || resp.Error != nil {
		t.Fatal("calc_solar_arc failed")
	}
	result := resp.Result.(callToolResult)
	if result.IsError {
		t.Fatalf("Tool error: %s", result.Content[0].Text)
	}
}

func TestHandleCalcTransit(t *testing.T) {
	s := NewServer(".")
	args, _ := json.Marshal(map[string]interface{}{
		"natal_latitude":   39.9042,
		"natal_longitude":  116.4074,
		"natal_jd_ut":      2448057.5208,
		"transit_latitude":  39.9042,
		"transit_longitude": 116.4074,
		"start_jd_ut":      2460310.667,
		"end_jd_ut":        2460320.667,
		"transit_planets":  []string{"SUN"},
		"event_config": map[string]bool{
			"include_tr_na":        true,
			"include_sign_ingress": true,
		},
	})
	params, _ := json.Marshal(callToolParams{Name: "calc_transit", Arguments: args})
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      6,
		Method:  "tools/call",
		Params:  params,
	}
	resp := s.handleRequest(req)
	if resp == nil || resp.Error != nil {
		t.Fatal("calc_transit failed")
	}
	result := resp.Result.(callToolResult)
	if result.IsError {
		t.Fatalf("Tool error: %s", result.Content[0].Text)
	}
}

func TestHandleCalcDoubleChart(t *testing.T) {
	s := NewServer(".")
	args, _ := json.Marshal(map[string]interface{}{
		"inner_latitude":  39.9042,
		"inner_longitude": 116.4074,
		"inner_jd_ut":     2448057.5208,
		"outer_latitude":  39.9042,
		"outer_longitude": 116.4074,
		"outer_jd_ut":     2460310.667,
		"special_points": map[string]interface{}{
			"inner_points": []string{"ASC", "MC"},
		},
	})
	params, _ := json.Marshal(callToolParams{Name: "calc_double_chart", Arguments: args})
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      11,
		Method:  "tools/call",
		Params:  params,
	}
	resp := s.handleRequest(req)
	if resp == nil || resp.Error != nil {
		t.Fatal("calc_double_chart failed")
	}
	result := resp.Result.(callToolResult)
	if result.IsError {
		t.Fatalf("Tool error: %s", result.Content[0].Text)
	}
}

func TestHandleCalcPlanetPosition(t *testing.T) {
	s := NewServer(".")
	args, _ := json.Marshal(map[string]interface{}{
		"planet": "SUN",
		"jd_ut":  2451545.0,
	})
	params, _ := json.Marshal(callToolParams{Name: "calc_planet_position", Arguments: args})
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      12,
		Method:  "tools/call",
		Params:  params,
	}
	resp := s.handleRequest(req)
	if resp == nil || resp.Error != nil {
		t.Fatal("calc_planet_position failed")
	}
	result := resp.Result.(callToolResult)
	if result.IsError {
		t.Fatalf("Tool error: %s", result.Content[0].Text)
	}
}

func TestHandleJDToDatetime(t *testing.T) {
	s := NewServer(".")
	args, _ := json.Marshal(map[string]interface{}{
		"jd":       2451545.0,
		"timezone": "Asia/Shanghai",
	})
	params, _ := json.Marshal(callToolParams{Name: "jd_to_datetime", Arguments: args})
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      13,
		Method:  "tools/call",
		Params:  params,
	}
	resp := s.handleRequest(req)
	if resp == nil || resp.Error != nil {
		t.Fatal("jd_to_datetime failed")
	}
	result := resp.Result.(callToolResult)
	if result.IsError {
		t.Fatalf("Tool error: %s", result.Content[0].Text)
	}
}

func TestHandleUnknownTool(t *testing.T) {
	s := NewServer(".")
	args, _ := json.Marshal(map[string]string{})
	params, _ := json.Marshal(callToolParams{Name: "nonexistent", Arguments: args})
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      7,
		Method:  "tools/call",
		Params:  params,
	}
	resp := s.handleRequest(req)
	if resp == nil {
		t.Fatal("Expected response")
	}
	if resp.Error == nil {
		t.Error("Expected error for unknown tool")
	}
}

func TestHandleUnknownMethod(t *testing.T) {
	s := NewServer(".")
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      8,
		Method:  "unknown/method",
	}
	resp := s.handleRequest(req)
	if resp == nil {
		t.Fatal("Expected response")
	}
	if resp.Error == nil {
		t.Error("Expected error for unknown method")
	}
}

func TestNotificationsInitialized(t *testing.T) {
	s := NewServer(".")
	req := &jsonRPCRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	resp := s.handleRequest(req)
	if resp != nil {
		t.Error("notifications/initialized should return nil (no response)")
	}
}
