package mcp

import (
	"encoding/json"
	"testing"
)

func TestHandleCalcDavisonChart(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"person1_latitude": 51.5074,
		"person1_longitude": -0.1278,
		"person1_jd_ut": 2451545.0,
		"person2_latitude": 48.8566,
		"person2_longitude": 2.3522,
		"person2_jd_ut": 2451600.0
	}`)
	result, err := s.handleCalcDavisonChart(args)
	if err != nil {
		t.Fatalf("handleCalcDavisonChart: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcBonification(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"latitude": 51.5074,
		"longitude": -0.1278,
		"jd_ut": 2451545.0
	}`)
	result, err := s.handleCalcBonification(args)
	if err != nil {
		t.Fatalf("handleCalcBonification: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestHandleCalcProgressions_CSV(t *testing.T) {
	s := newTestServer()
	args := json.RawMessage(`{
		"natal_jd_ut": 2451545.0,
		"transit_jd_ut": 2460000.0,
		"format": "csv"
	}`)
	result, err := s.handleCalcProgressions(args)
	if err != nil {
		t.Fatalf("handleCalcProgressions CSV: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}


