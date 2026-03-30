package mcp

import (
	"encoding/json"
	"testing"
)

// CSV format tests for handlers that support it

func TestHandleCalcSolarArc_CSV(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"natal_jd_ut": 2451545.0,
		"transit_jd_ut": 2460000.0,
		"format": "csv"
	}`)
	result, err := s.handleCalcSolarArc(args)
	if err != nil {
		t.Fatalf("handleCalcSolarArc CSV: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcSolarReturn_CSV(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"natal_jd_ut": 2451545.0,
		"natal_latitude": 51.5074,
		"natal_longitude": -0.1278,
		"search_jd_ut": 2451905.0,
		"format": "csv"
	}`)
	result, err := s.handleCalcSolarReturn(args)
	if err != nil {
		t.Fatalf("handleCalcSolarReturn CSV: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcLunarReturn_CSV(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"natal_jd_ut": 2451545.0,
		"natal_latitude": 51.5074,
		"natal_longitude": -0.1278,
		"search_jd_ut": 2451570.0,
		"format": "csv"
	}`)
	result, err := s.handleCalcLunarReturn(args)
	if err != nil {
		t.Fatalf("handleCalcLunarReturn CSV: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcSingleChart_CSV(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"latitude": 51.5074,
		"longitude": -0.1278,
		"jd_ut": 2451545.0,
		"format": "csv"
	}`)
	result, err := s.handleCalcSingleChart(args)
	if err != nil {
		t.Fatalf("handleCalcSingleChart CSV: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcDoubleChart_CSV(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"inner_latitude": 51.5074,
		"inner_longitude": -0.1278,
		"inner_jd_ut": 2451545.0,
		"outer_latitude": 48.8566,
		"outer_longitude": 2.3522,
		"outer_jd_ut": 2451600.0,
		"format": "csv"
	}`)
	result, err := s.handleCalcDoubleChart(args)
	if err != nil {
		t.Fatalf("handleCalcDoubleChart CSV: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcTransit_CSVFormat2(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"natal_latitude": 51.5074,
		"natal_longitude": -0.1278,
		"natal_jd_ut": 2451545.0,
		"start_jd_ut": 2451545.0,
		"end_jd_ut": 2451575.0,
		"format": "csv"
	}`)
	result, err := s.handleCalcTransit(args)
	if err != nil {
		t.Fatalf("handleCalcTransit CSV: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcDignity_CSV(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"latitude": 51.5074,
		"longitude": -0.1278,
		"jd_ut": 2451545.0,
		"format": "csv"
	}`)
	result, err := s.handleCalcDignity(args)
	if err != nil {
		t.Fatalf("handleCalcDignity CSV: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcCompositeChart_CSV(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"person1_latitude": 51.5074,
		"person1_longitude": -0.1278,
		"person1_jd_ut": 2451545.0,
		"person2_latitude": 48.8566,
		"person2_longitude": 2.3522,
		"person2_jd_ut": 2451600.0,
		"format": "csv"
	}`)
	result, err := s.handleCalcCompositeChart(args)
	if err != nil {
		t.Fatalf("handleCalcCompositeChart CSV: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcLunarPhases_CSV(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"start_jd_ut": 2451545.0,
		"end_jd_ut": 2451575.0,
		"format": "csv"
	}`)
	result, err := s.handleCalcLunarPhases(args)
	if err != nil {
		t.Fatalf("handleCalcLunarPhases CSV: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcEclipses_CSV(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"start_jd_ut": 2451545.0,
		"end_jd_ut": 2451910.0,
		"format": "csv"
	}`)
	result, err := s.handleCalcEclipses(args)
	if err != nil {
		t.Fatalf("handleCalcEclipses CSV: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

// Test with custom parameters


