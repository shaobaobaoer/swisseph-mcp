package solarsage

import (
	"testing"
)


func TestDavisonChart_OK(t *testing.T) {
	dc, err := DavisonChart(51.5, -0.1, "2000-01-01T12:00:00Z", 48.85, 2.35, "1995-06-15T08:00:00Z")
	if err != nil {
		t.Fatalf("DavisonChart: %v", err)
	}
	if dc == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestDavisonChart_InvalidCoords1(t *testing.T) {
	_, err := DavisonChart(999, -0.1, "2000-01-01", 48.85, 2.35, "1995-06-15")
	if err == nil {
		t.Error("expected error for invalid person 1 coords")
	}
}

func TestDavisonChart_InvalidCoords2(t *testing.T) {
	_, err := DavisonChart(51.5, -0.1, "2000-01-01", 999, 2.35, "1995-06-15")
	if err == nil {
		t.Error("expected error for invalid person 2 coords")
	}
}

func TestDavisonChart_InvalidDatetime1(t *testing.T) {
	_, err := DavisonChart(51.5, -0.1, "bad", 48.85, 2.35, "1995-06-15")
	if err == nil {
		t.Error("expected error for bad person 1 datetime")
	}
}

func TestDavisonChart_InvalidDatetime2(t *testing.T) {
	_, err := DavisonChart(51.5, -0.1, "2000-01-01", 48.85, 2.35, "bad")
	if err == nil {
		t.Error("expected error for bad person 2 datetime")
	}
}


func TestBonification_OK(t *testing.T) {
	result, err := Bonification(51.5, -0.1, "2000-01-01T12:00:00Z")
	if err != nil {
		t.Fatalf("Bonification: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty bonification result")
	}
}

func TestBonification_InvalidCoords(t *testing.T) {
	_, err := Bonification(999, -0.1, "2000-01-01")
	if err == nil {
		t.Error("expected error for invalid coords")
	}
}

func TestBonification_InvalidDatetime(t *testing.T) {
	_, err := Bonification(51.5, -0.1, "bad")
	if err == nil {
		t.Error("expected error for bad datetime")
	}
}

