package chart

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/anthropic/swisseph-mcp/pkg/models"
	"github.com/anthropic/swisseph-mcp/pkg/sweph"
)

func TestMain(m *testing.M) {
	ephePath := filepath.Join("..", "..", "third_party", "swisseph", "ephe")
	sweph.Init(ephePath)
	code := m.Run()
	sweph.Close()
	os.Exit(code)
}

// 1990-06-15 00:30 UTC, Beijing
var testJD = sweph.JulDay(1990, 6, 15, 0.5, true)

func TestCalcSingleChart(t *testing.T) {
	planets := []models.PlanetID{
		models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
		models.PlanetVenus, models.PlanetMars,
	}
	orbs := models.DefaultOrbConfig()

	info, err := CalcSingleChart(39.9042, 116.4074, testJD, planets, orbs, models.HousePlacidus)
	if err != nil {
		t.Fatalf("CalcSingleChart: %v", err)
	}

	// Should have 5 planets
	if len(info.Planets) != 5 {
		t.Errorf("Expected 5 planets, got %d", len(info.Planets))
	}

	// Should have 12 houses
	if len(info.Houses) != 12 {
		t.Errorf("Expected 12 houses, got %d", len(info.Houses))
	}

	// Check Sun is in Gemini (60-90°)
	var sunPos *models.PlanetPosition
	for i, p := range info.Planets {
		if p.PlanetID == models.PlanetSun {
			sunPos = &info.Planets[i]
			break
		}
	}
	if sunPos == nil {
		t.Fatal("Sun not found in chart")
	}
	if sunPos.Longitude < 60 || sunPos.Longitude > 90 {
		t.Errorf("Sun longitude = %f, expected in Gemini (60-90)", sunPos.Longitude)
	}
	if sunPos.Sign != "Gemini" {
		t.Errorf("Sun sign = %q, expected Gemini", sunPos.Sign)
	}

	// ASC should be valid
	if info.Angles.ASC < 0 || info.Angles.ASC >= 360 {
		t.Errorf("ASC = %f, out of range", info.Angles.ASC)
	}
	// DSC = ASC + 180
	expectedDSC := sweph.NormalizeDegrees(info.Angles.ASC + 180)
	if math.Abs(info.Angles.DSC-expectedDSC) > 0.01 {
		t.Errorf("DSC = %f, expected %f (ASC+180)", info.Angles.DSC, expectedDSC)
	}

	// Should have some aspects
	if len(info.Aspects) == 0 {
		t.Error("Expected some aspects, got 0")
	}
}

func TestCalcDoubleChart(t *testing.T) {
	planets := []models.PlanetID{
		models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
	}
	orbs := models.DefaultOrbConfig()
	transitJD := sweph.JulDay(2024, 1, 1, 4.0, true)

	inner, outer, cross, err := CalcDoubleChart(
		39.9042, 116.4074, testJD, planets,
		39.9042, 116.4074, transitJD, planets,
		nil, orbs, models.HousePlacidus,
	)
	if err != nil {
		t.Fatalf("CalcDoubleChart: %v", err)
	}

	if len(inner.Planets) != 3 {
		t.Errorf("Inner planets: %d, want 3", len(inner.Planets))
	}
	if len(outer.Planets) != 3 {
		t.Errorf("Outer planets: %d, want 3", len(outer.Planets))
	}
	// Cross aspects should exist (3x3 = 9 pairs to check)
	if len(cross) == 0 {
		t.Log("Warning: no cross aspects found (possible but unlikely)")
	}
}

func TestCalcDoubleChart_WithSpecialPoints(t *testing.T) {
	planets := []models.PlanetID{models.PlanetSun}
	orbs := models.DefaultOrbConfig()
	transitJD := sweph.JulDay(2024, 1, 1, 4.0, true)

	sp := &models.SpecialPointsConfig{
		InnerPoints: []models.SpecialPointID{models.PointASC, models.PointMC},
	}
	_, _, cross, err := CalcDoubleChart(
		39.9042, 116.4074, testJD, planets,
		39.9042, 116.4074, transitJD, planets,
		sp, orbs, models.HousePlacidus,
	)
	if err != nil {
		t.Fatalf("CalcDoubleChart with special points: %v", err)
	}
	// With special points, more cross-aspects are possible
	_ = cross
}

func TestCalcPlanetLongitude(t *testing.T) {
	lon, speed, err := CalcPlanetLongitude(models.PlanetSun, testJD)
	if err != nil {
		t.Fatalf("CalcPlanetLongitude Sun: %v", err)
	}
	// Sun in Gemini
	if lon < 60 || lon > 90 {
		t.Errorf("Sun lon = %f, expected 60-90", lon)
	}
	if speed < 0.9 || speed > 1.1 {
		t.Errorf("Sun speed = %f, expected ~1.0", speed)
	}
}

func TestCalcPlanetLongitude_SouthNode(t *testing.T) {
	northLon, _, err := CalcPlanetLongitude(models.PlanetNorthNodeTrue, testJD)
	if err != nil {
		t.Fatalf("North node: %v", err)
	}
	southLon, _, err := CalcPlanetLongitude(models.PlanetSouthNode, testJD)
	if err != nil {
		t.Fatalf("South node: %v", err)
	}
	diff := math.Abs(northLon - southLon)
	if diff > 180 {
		diff = 360 - diff
	}
	if math.Abs(diff-180) > 0.01 {
		t.Errorf("North-South node diff = %f, expected 180", diff)
	}
}

func TestCalcSpecialPointLongitude(t *testing.T) {
	asc, err := CalcSpecialPointLongitude(models.PointASC, 39.9042, 116.4074, testJD, models.HousePlacidus)
	if err != nil {
		t.Fatalf("ASC: %v", err)
	}
	if asc < 0 || asc >= 360 {
		t.Errorf("ASC = %f, out of range", asc)
	}

	dsc, err := CalcSpecialPointLongitude(models.PointDSC, 39.9042, 116.4074, testJD, models.HousePlacidus)
	if err != nil {
		t.Fatalf("DSC: %v", err)
	}
	expectedDSC := sweph.NormalizeDegrees(asc + 180)
	if math.Abs(dsc-expectedDSC) > 0.01 {
		t.Errorf("DSC = %f, expected %f", dsc, expectedDSC)
	}
}

func TestCalcNatalFixedHouses(t *testing.T) {
	cusps, err := CalcNatalFixedHouses(39.9042, 116.4074, testJD, models.HousePlacidus)
	if err != nil {
		t.Fatalf("CalcNatalFixedHouses: %v", err)
	}
	if len(cusps) != 12 {
		t.Errorf("Expected 12 cusps, got %d", len(cusps))
	}
	for i, c := range cusps {
		if c < 0 || c >= 360 {
			t.Errorf("Cusp[%d] = %f, out of range", i, c)
		}
	}
}

func TestFindHouseForLongitude(t *testing.T) {
	cusps := []float64{0, 30, 60, 90, 120, 150, 180, 210, 240, 270, 300, 330}
	tests := []struct {
		lon  float64
		want int
	}{
		{15, 1},
		{45, 2},
		{330, 12},
		{359, 12},
	}
	for _, tt := range tests {
		got := FindHouseForLongitude(tt.lon, cusps)
		if got != tt.want {
			t.Errorf("FindHouseForLongitude(%f) = %d, want %d", tt.lon, got, tt.want)
		}
	}
}

func TestCalcSingleChart_AllBodies(t *testing.T) {
	// Test with all supported celestial bodies including Chiron, nodes, Lilith
	allPlanets := []models.PlanetID{
		models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
		models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
		models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune,
		models.PlanetPluto, models.PlanetChiron,
		models.PlanetNorthNodeTrue, models.PlanetNorthNodeMean,
		models.PlanetSouthNode, models.PlanetLilithMean, models.PlanetLilithTrue,
	}
	orbs := models.DefaultOrbConfig()

	info, err := CalcSingleChart(39.9042, 116.4074, testJD, allPlanets, orbs, models.HousePlacidus)
	if err != nil {
		t.Fatalf("CalcSingleChart all bodies: %v", err)
	}

	if len(info.Planets) != len(allPlanets) {
		t.Errorf("Expected %d planets, got %d", len(allPlanets), len(info.Planets))
	}

	// Verify each body has valid data
	for _, p := range info.Planets {
		if p.Longitude < 0 || p.Longitude >= 360 {
			t.Errorf("%s: longitude %f out of range", p.PlanetID, p.Longitude)
		}
		if p.Sign == "" {
			t.Errorf("%s: empty sign", p.PlanetID)
		}
		if p.House < 1 || p.House > 12 {
			t.Errorf("%s: house %d out of range", p.PlanetID, p.House)
		}
		if p.SignDegree < 0 || p.SignDegree >= 30 {
			t.Errorf("%s: sign degree %f out of range", p.PlanetID, p.SignDegree)
		}
	}
}

func TestCalcSpecialPointLongitude_AllPoints(t *testing.T) {
	points := []models.SpecialPointID{
		models.PointASC, models.PointMC, models.PointDSC, models.PointIC,
		models.PointVertex, models.PointAntiVertex, models.PointEastPoint,
		models.PointLotFortune, models.PointLotSpirit,
	}
	for _, sp := range points {
		lon, err := CalcSpecialPointLongitude(sp, 39.9042, 116.4074, testJD, models.HousePlacidus)
		if err != nil {
			t.Errorf("%s: %v", sp, err)
			continue
		}
		if lon < 0 || lon >= 360 {
			t.Errorf("%s: longitude %f out of range", sp, lon)
		}
	}
}

func TestCalcSingleChart_DifferentHouseSystems(t *testing.T) {
	planets := []models.PlanetID{models.PlanetSun, models.PlanetMoon}
	orbs := models.DefaultOrbConfig()

	systems := []models.HouseSystem{
		models.HousePlacidus, models.HouseKoch, models.HouseEqual,
		models.HouseWholeSign, models.HouseCampanus,
		models.HouseRegiomontanus, models.HousePorphyry,
	}
	for _, hsys := range systems {
		info, err := CalcSingleChart(39.9042, 116.4074, testJD, planets, orbs, hsys)
		if err != nil {
			t.Errorf("House system %s: %v", hsys, err)
			continue
		}
		if len(info.Houses) != 12 {
			t.Errorf("House system %s: expected 12 cusps, got %d", hsys, len(info.Houses))
		}
		// Planets should always be the same regardless of house system
		if len(info.Planets) != 2 {
			t.Errorf("House system %s: expected 2 planets, got %d", hsys, len(info.Planets))
		}
	}
}

// =============================================================================
// JN Natal Chart Precision Test - Based on Solar Fire Chart Analysis Report
//
// Subject: JN, Male Chart
// Birth: 1997-12-18 17:36:00 AWST (UTC+8), i.e. UTC 09:36:00
// Birth place: Jinshan, China, 30°54'N，121°09'E
// House system: Placidus
// JDE = 2450800.900729 (TT), JD_UT ≈ 2450800.900009
// =============================================================================

func TestCalcSingleChart_JN_Natal(t *testing.T) {
	const (
		jnJDUT = 2450800.900009
		jnLat  = 30.9    // 30°54'N
		jnLon  = 121.15  // 121°09'E
		tolLon = 0.01    // longitude tolerance 0.01°
		tolLat = 0.05    // latitude tolerance 0.05°
		tolSpd = 0.02    // speed tolerance 0.02 deg/day
		tolAng = 0.01    // angle/cusp tolerance 0.01°
	)

	planets := []models.PlanetID{
		models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
		models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
		models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune,
		models.PlanetPluto, models.PlanetChiron,
		models.PlanetNorthNodeMean, models.PlanetLilithMean,
	}
	orbs := models.OrbConfig{
		Conjunction: 8, Opposition: 8, Trine: 7, Square: 7,
		Sextile: 5, Quincunx: 3, SemiSextile: 2,
		SemiSquare: 2, Sesquiquadrate: 2,
	}

	info, err := CalcSingleChart(jnLat, jnLon, jnJDUT, planets, orbs, models.HousePlacidus)
	if err != nil {
		t.Fatalf("CalcSingleChart JN: %v", err)
	}

	// --- Planet position verification ---
	// Solar Fire reference data (DMS to decimal conversion)
	type expected struct {
		pid      models.PlanetID
		lon      float64
		lat      float64
		speed    float64
		retro    bool
		sign     string
		signDeg  float64
		house    int
	}
	// Reference from Solar Fire, 4 corrections based on Swiss Ephemeris output:
	//   1. CHIRON lon: report 221.022 conflicts with sign_degree 14.022 Scorpio,
	//      correct = 210 + 14.022 = 224.022 (report typo)
	//   2. SATURN speed: report 0.200 deg/day unreasonable for Saturn (near station),
	//      actual ~0.004 deg/day
	//   3. LILITH_MEAN: report 351.497 same as MC, copy error,
	//      actual ~180.395 (Libra 0.395)
	//   4. JUPITER house: report says 9, swe_house_pos returns 8
	//      (Jupiter 319.5 between 8th cusp 298.7 and 9th cusp 322.7)
	wantPlanets := []expected{
		{models.PlanetMoon, 138.116, -2.183, 12.323, false, "Leo", 18.116, 2},
		{models.PlanetSun, 266.500, 0.0, 1.018, false, "Sagittarius", 26.500, 6},
		{models.PlanetMercury, 263.933, 2.216, -1.361, true, "Sagittarius", 23.933, 6},
		{models.PlanetVenus, 302.553, -0.879, 0.315, false, "Aquarius", 2.553, 8},
		{models.PlanetMars, 300.097, -1.221, 0.782, false, "Aquarius", 0.097, 8},
		{models.PlanetJupiter, 319.548, -0.879, 0.188, false, "Aquarius", 19.548, 8},
		{models.PlanetSaturn, 13.538, -2.571, 0.004, false, "Aries", 13.538, 10},
		{models.PlanetUranus, 306.425, -0.604, 0.048, false, "Aquarius", 6.425, 8},
		{models.PlanetNeptune, 298.468, 0.367, 0.033, false, "Capricorn", 28.468, 7},
		{models.PlanetPluto, 246.274, 11.872, 0.038, false, "Sagittarius", 6.274, 6},
		{models.PlanetChiron, 224.023, 1.159, 0.117, false, "Scorpio", 14.023, 5},
		{models.PlanetLilithMean, 180.395, 1.417, 0.111, false, "Libra", 0.395, 4},
	}

	// Build lookup map from computed results
	posMap := make(map[models.PlanetID]models.PlanetPosition)
	for _, p := range info.Planets {
		posMap[p.PlanetID] = p
	}

	for _, w := range wantPlanets {
		got, ok := posMap[w.pid]
		if !ok {
			t.Errorf("planet %s not found", w.pid)
			continue
		}
		t.Run(string(w.pid), func(t *testing.T) {
			if math.Abs(got.Longitude-w.lon) > tolLon {
				t.Errorf("longitude: got %.4f, want %.3f (diff %.4f)", got.Longitude, w.lon, got.Longitude-w.lon)
			}
			if math.Abs(got.Latitude-w.lat) > tolLat {
				t.Errorf("latitude: got %.4f, want %.3f (diff %.4f)", got.Latitude, w.lat, got.Latitude-w.lat)
			}
			if math.Abs(got.Speed-w.speed) > tolSpd {
				t.Errorf("speed: got %.4f, want %.3f (diff %.4f)", got.Speed, w.speed, got.Speed-w.speed)
			}
			if got.IsRetrograde != w.retro {
				t.Errorf("retrograde: got %v, want %v", got.IsRetrograde, w.retro)
			}
			if got.Sign != w.sign {
				t.Errorf("sign: got %q, want %q", got.Sign, w.sign)
			}
			if math.Abs(got.SignDegree-w.signDeg) > tolLon {
				t.Errorf("sign degree: got %.4f, want %.3f", got.SignDegree, w.signDeg)
			}
			if got.House != w.house {
				t.Errorf("house: got %d, want %d", got.House, w.house)
			}
		})
	}

	// --- House cusp verification ---
	wantHouses := []float64{
		96.530, 118.656, 142.694, 171.500, 206.129, 242.977,
		276.530, 298.656, 322.694, 351.500, 26.129, 62.977,
	}
	if len(info.Houses) != 12 {
		t.Fatalf("cusp count: got %d, want 12", len(info.Houses))
	}
	for i, wh := range wantHouses {
		if math.Abs(info.Houses[i]-wh) > tolAng {
			t.Errorf("cusp[%d]: got %.4f, want %.3f (diff %.4f)", i+1, info.Houses[i], wh, info.Houses[i]-wh)
		}
	}

	// --- Angles verification ---
	t.Run("Angles", func(t *testing.T) {
		if math.Abs(info.Angles.ASC-96.530) > tolAng {
			t.Errorf("ASC: got %.4f, want 96.530", info.Angles.ASC)
		}
		if math.Abs(info.Angles.MC-351.500) > tolAng {
			t.Errorf("MC: got %.4f, want 351.500", info.Angles.MC)
		}
		if math.Abs(info.Angles.DSC-276.530) > tolAng {
			t.Errorf("DSC: got %.4f, want 276.530", info.Angles.DSC)
		}
		if math.Abs(info.Angles.IC-171.500) > tolAng {
			t.Errorf("IC: got %.4f, want 171.500", info.Angles.IC)
		}
	})

	// --- Aspect verification (auto-calculated, verify key aspects) ---
	t.Run("Aspects", func(t *testing.T) {
		// build aspect lookup helper
		type aspectKey struct{ a, b string }
		aspectMap := make(map[aspectKey]models.AspectInfo)
		for _, asp := range info.Aspects {
			aspectMap[aspectKey{asp.PlanetA, asp.PlanetB}] = asp
			aspectMap[aspectKey{asp.PlanetB, asp.PlanetA}] = asp
		}

		// Venus-Mars conjunction: 302.553 vs 300.097，diff 2.456°; orb 8°
		if asp, ok := aspectMap[aspectKey{"VENUS", "MARS"}]; ok {
			if asp.AspectType != models.AspectConjunction {
				t.Errorf("Venus-Mars: got %s, want Conjunction", asp.AspectType)
			}
			if math.Abs(asp.Orb-2.456) > 0.05 {
				t.Errorf("Venus-Mars orb: got %.3f, want ~2.456", asp.Orb)
			}
		} else {
			t.Error("Venus-Mars conjunction not found")
		}

		// Mars-Neptune conjunction: 300.097 vs 298.468，diff 1.629°
		if asp, ok := aspectMap[aspectKey{"MARS", "NEPTUNE"}]; ok {
			if asp.AspectType != models.AspectConjunction {
				t.Errorf("Mars-Neptune: got %s, want Conjunction", asp.AspectType)
			}
			if math.Abs(asp.Orb-1.629) > 0.05 {
				t.Errorf("Mars-Neptune orb: got %.3f, want ~1.629", asp.Orb)
			}
		} else {
			t.Error("Mars-Neptune conjunction not found")
		}

		// Moon-Saturn trine: 138.116 vs 13.538，diff 124.578°; from 120 by4.578°
		if asp, ok := aspectMap[aspectKey{"MOON", "SATURN"}]; ok {
			if asp.AspectType != models.AspectTrine {
				t.Errorf("Moon-Saturn: got %s, want Trine", asp.AspectType)
			}
			if math.Abs(asp.Orb-4.578) > 0.05 {
				t.Errorf("Moon-Saturn orb: got %.3f, want ~4.578", asp.Orb)
			}
		} else {
			t.Error("Moon-Saturn trine not found")
		}

		// Mars-Uranus conjunction: 300.097 vs 306.425，diff 6.328°
		if asp, ok := aspectMap[aspectKey{"MARS", "URANUS"}]; ok {
			if asp.AspectType != models.AspectConjunction {
				t.Errorf("Mars-Uranus: got %s, want Conjunction", asp.AspectType)
			}
			if math.Abs(asp.Orb-6.328) > 0.05 {
				t.Errorf("Mars-Uranus orb: got %.3f, want ~6.328", asp.Orb)
			}
		} else {
			t.Error("Mars-Uranus conjunction not found")
		}

		// ensure reasonable aspect count (13 bodies should yield many aspects)
		if len(info.Aspects) < 10 {
			t.Errorf("aspect count too low: %d, expected >= 10", len(info.Aspects))
		}
	})
}

func TestWrapAngle(t *testing.T) {
	tests := []struct {
		in, want float64
	}{
		{0, 0},
		{180, -180},
		{-180, -180},
		{90, 90},
		{-90, -90},
		{270, -90},
		{-270, 90},
		{360, 0},
	}
	for _, tt := range tests {
		got := WrapAngle(tt.in)
		if math.Abs(got-tt.want) > 0.001 {
			t.Errorf("WrapAngle(%f) = %f, want %f", tt.in, got, tt.want)
		}
	}
}
