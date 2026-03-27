package solarsage

import (
	"testing"
)

func TestFirdaria_OK(t *testing.T) {
	result := Firdaria(true, 30.0)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.CurrentPeriod == nil {
		t.Error("expected current period")
	}
}

func TestFirdaria_NightBirth(t *testing.T) {
	result := Firdaria(false, 45.0)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

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

func TestPrimaryDirections_OK(t *testing.T) {
	result, err := PrimaryDirections(51.5, -0.1, "2000-01-01T12:00:00Z", 50)
	if err != nil {
		t.Fatalf("PrimaryDirections: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestPrimaryDirections_InvalidCoords(t *testing.T) {
	_, err := PrimaryDirections(999, -0.1, "2000-01-01", 50)
	if err == nil {
		t.Error("expected error for invalid coords")
	}
}

func TestPrimaryDirections_InvalidDatetime(t *testing.T) {
	_, err := PrimaryDirections(51.5, -0.1, "bad", 50)
	if err == nil {
		t.Error("expected error for bad datetime")
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

func TestSymbolicDirections_OK(t *testing.T) {
	result, err := SymbolicDirections(51.5, -0.1, "2000-01-01T12:00:00Z", 30.0)
	if err != nil {
		t.Fatalf("SymbolicDirections: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestSymbolicDirections_InvalidCoords(t *testing.T) {
	_, err := SymbolicDirections(999, -0.1, "2000-01-01", 30.0)
	if err == nil {
		t.Error("expected error for invalid coords")
	}
}

func TestSymbolicDirections_InvalidDatetime(t *testing.T) {
	_, err := SymbolicDirections(51.5, -0.1, "bad", 30.0)
	if err == nil {
		t.Error("expected error for bad datetime")
	}
}
