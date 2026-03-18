package models

import (
	"math"
	"testing"
)

func TestSignFromLongitude(t *testing.T) {
	tests := []struct {
		lon  float64
		want string
	}{
		{0, "Aries"},
		{15, "Aries"},
		{30, "Taurus"},
		{59.99, "Taurus"},
		{60, "Gemini"},
		{90, "Cancer"},
		{120, "Leo"},
		{150, "Virgo"},
		{180, "Libra"},
		{210, "Scorpio"},
		{240, "Sagittarius"},
		{270, "Capricorn"},
		{300, "Aquarius"},
		{330, "Pisces"},
		{359.99, "Pisces"},
	}
	for _, tt := range tests {
		got := SignFromLongitude(tt.lon)
		if got != tt.want {
			t.Errorf("SignFromLongitude(%f) = %q, want %q", tt.lon, got, tt.want)
		}
	}
}

func TestSignDegreeFromLongitude(t *testing.T) {
	tests := []struct {
		lon, want float64
	}{
		{0, 0},
		{15, 15},
		{30, 0},
		{45, 15},
		{93.5, 3.5},
		{359, 29},
	}
	for _, tt := range tests {
		got := SignDegreeFromLongitude(tt.lon)
		if math.Abs(got-tt.want) > 0.01 {
			t.Errorf("SignDegreeFromLongitude(%f) = %f, want %f", tt.lon, got, tt.want)
		}
	}
}

func TestDefaultOrbConfig(t *testing.T) {
	orbs := DefaultOrbConfig()
	if orbs.Conjunction != 8 {
		t.Errorf("Default conjunction orb = %f, want 8", orbs.Conjunction)
	}
	if orbs.Trine != 7 {
		t.Errorf("Default trine orb = %f, want 7", orbs.Trine)
	}
}

func TestGetOrb(t *testing.T) {
	orbs := DefaultOrbConfig()
	tests := []struct {
		at   AspectType
		want float64
	}{
		{AspectConjunction, 8},
		{AspectOpposition, 8},
		{AspectTrine, 7},
		{AspectSquare, 7},
		{AspectSextile, 5},
		{AspectQuincunx, 3},
		{AspectSemiSextile, 2},
		{AspectSemiSquare, 2},
		{AspectSesquiquadrate, 2},
		{AspectType("UNKNOWN"), 0},
	}
	for _, tt := range tests {
		got := orbs.GetOrb(tt.at)
		if got != tt.want {
			t.Errorf("GetOrb(%s) = %f, want %f", tt.at, got, tt.want)
		}
	}
}

func TestDefaultEventConfig(t *testing.T) {
	cfg := DefaultEventConfig()
	if !cfg.IncludeTrNa {
		t.Error("Default IncludeTrNa should be true")
	}
	if !cfg.IncludeTrTr {
		t.Error("Default IncludeTrTr should be true")
	}
	if !cfg.IncludeVoidOfCourse {
		t.Error("Default IncludeVoidOfCourse should be true")
	}
}

func TestStandardAspects(t *testing.T) {
	if len(StandardAspects) != 9 {
		t.Errorf("StandardAspects has %d entries, want 9", len(StandardAspects))
	}
	// Verify angles
	expected := map[AspectType]float64{
		AspectConjunction:    0,
		AspectOpposition:     180,
		AspectTrine:          120,
		AspectSquare:         90,
		AspectSextile:        60,
		AspectQuincunx:       150,
		AspectSemiSextile:    30,
		AspectSemiSquare:     45,
		AspectSesquiquadrate: 135,
	}
	for _, a := range StandardAspects {
		if exp, ok := expected[a.Type]; ok {
			if a.Angle != exp {
				t.Errorf("Aspect %s angle = %f, want %f", a.Type, a.Angle, exp)
			}
		}
	}
}
