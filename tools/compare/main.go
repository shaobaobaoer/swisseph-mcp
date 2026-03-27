package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/shaobaobaoer/solarsage-mcp/pkg/chart"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/export"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/julian"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/sweph"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/transit"
)

// TestCase holds test case parameters
type TestCase struct {
	Name        string
	NatalJD     float64
	NatalLat    float64
	NatalLon    float64
	Timezone    string
	TzAbbr      string
	SFCSVPaths  []string // Support multiple CSV files
	UseMeanNode bool
	// Expected ASC/MC for verification
	ExpectedASC float64
	ExpectedMC  float64
	// NatalASC/NatalMC override values from SF meta (precise to arcsec)
	// If non-zero, these override calculated natal angles for Tr-Na references
	NatalASC float64
	NatalMC  float64
	// NatalMCForASC: separate MC override for progressed ASC calculation.
	// SF uses different MC base for ASC derivation. Set to -1 to use sweph computed MC.
	NatalMCForASC float64
	// NatalASCForProgressions: controls ASC progression method.
	// Set to -1 to use Solar Arc in Right Ascension method (SF style).
	NatalASCForProgressions float64
	// NatalPlanetOverrides allows specifying exact natal planet positions from SF.
	// Use this to match reference data that may use different ephemeris (DE200 vs DE432).
	// Map key is planet ID (e.g., "MOON", "MERCURY"), value is longitude in degrees.
	NatalPlanetOverrides map[string]float64
	// Natal planets for Tr-Na reference points
	NatalPlanets []models.PlanetID
	// Progressed planets for Tr-Sp, Sp-Na, Sp-Sp (defaults to all planets including Chiron)
	ProgressionsPlanets []models.PlanetID
	// Solar arc planets for Tr-Sa, Sa-Na (defaults to all planets including Chiron)
	SolarArcPlanets []models.PlanetID
	// Event configuration
	EventConfig models.EventConfig
	// Whether to include Moon VOC events
	IncludeVOC bool
}

func main() {
	// Command line flags
	testcase := flag.Int("case", 1, "Test case number (1 or 2)")
	flag.Parse()

	// Initialize ephemeris
	exe, _ := os.Executable()
	ephePath := filepath.Join(filepath.Dir(exe), "..", "..", "third_party", "swisseph", "ephe")
	if _, err := os.Stat(ephePath); err != nil {
		ephePath = filepath.Join(".", "third_party", "swisseph", "ephe")
	}
	sweph.Init(ephePath)
	defer sweph.Close()

	// Select test case
	var tc TestCase
	switch *testcase {
	case 1:
		tc = getTestCase1()
	case 2:
		tc = getTestCase2()
	default:
		fmt.Printf("Unknown test case: %d\n", *testcase)
		return
	}

	fmt.Printf("=== Running Test Case %d: %s ===\n", *testcase, tc.Name)
	runTestCase(tc)
}

// getTestCase1 returns JN test case (already validated)
// Birth: 1997-12-18 17:36:00 AWST (UTC+8), i.e. UTC 09:36:00
// Birth place: Jinshan, China, 30°54'N，121°09'E
// Solar Fire output includes: Tr-Na, Tr-Tr, Tr-Sp, Tr-Sa, Sp-Na, Sp-Sp, Sa-Na, VOC
func getTestCase1() TestCase {
	// JN - Male Chart
	// From Solar Fire meta: JDE = 2450800.900729, DeltaT = +62s
	// JD_UT = 2450800.900729 - 62/86400 ≈ 2450800.900009
	// Note: SF testcase 1 does NOT use Chiron as a natal reference point in Tr-Na
	// SF meta precise values: ASC = 06°Cancer31'45'' = 96.529167°, MC = 21°Pisces29'51'' = 351.497500°
	return TestCase{
		Name:        "JN - Male Chart",
		NatalJD:     2450800.900009,
		NatalLat:    30.9,   // 30°54'N
		NatalLon:    121.15, // 121°09'E
		Timezone:    "Australia/Perth",
		TzAbbr:      "AWST",
		SFCSVPaths:  []string{"testdata/solarfire/testcase-1-transit.csv"},
		UseMeanNode: true,
		ExpectedASC: 96.530,   // 06°Cancer31'45''
		ExpectedMC:  351.500,  // 21°Pisces29'51''
		// SF meta precise values (for accurate natal reference positions)
		NatalASC:    96.529167, // SF meta: 06°Cancer31'45''
		NatalMC:     351.4975,  // SF meta: 21°Pisces29'51''
		// SF precise natal planet positions from meta (to match DE200 ephemeris)
		// These override sweph-computed positions to match SF's ephemeris
		// Keys must match models.PlanetID string values exactly
		NatalPlanetOverrides: map[string]float64{
			"SUN":             266.4833,  // 26°29' Sagittarius
			"MOON":            138.1164,  // 18°06'59" Leo
			"MERCURY":         263.9333,  // 23°56' Sagittarius
			"VENUS":           302.5528,  // 2°33'09" Aquarius
			"MARS":            300.0972,  // 0°05'50" Aquarius
			"JUPITER":         319.5479,  // 19°32'52" Aquarius
			"SATURN":          13.5378,   // 13°32'15" Aries
			"URANUS":          306.4248,  // 6°25'29" Aquarius
			"NEPTUNE":         298.4675,  // 28°28'03" Capricorn
			"PLUTO":           246.2742,  // 6°16'27" Sagittarius
			"CHIRON":          224.0227,  // 14°01'21" Scorpio
			"NORTH_NODE_MEAN": 164.4461,  // 14°26'45" Virgo
		},
		// SF testcase 1: Chiron IS used as natal reference point for Tr-Na
		NatalPlanets: []models.PlanetID{
			models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
			models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
			models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune,
			models.PlanetPluto, models.PlanetChiron, models.PlanetNorthNodeMean,
		},
		EventConfig: models.EventConfig{
			IncludeTrNa:         true,
			IncludeTrTr:         true,
			IncludeTrSp:         true,
			IncludeTrSa:         true,
			IncludeSpNa:         true,
			IncludeSpSp:         true,
			IncludeSaNa:         true,
			IncludeSignIngress:  true,
			IncludeHouseIngress: true,
			IncludeStation:      true,
		},
		IncludeVOC: true,
	}
}

// getTestCase2 returns XB test case
// Birth: Aug 3 1996, 0:30 am, AWST -8:00 (UTC Aug 2 16:30)
// Birth place: Huzhou China, 30°N52', 120°E06'
// Solar Fire output includes: Tr-Na, Sp-Na, Sp-Sp only (no Tr-Tr, Tr-Sp, Tr-Sa, Sa-Na, VOC)
func getTestCase2() TestCase {
	// XB - Female Chart
	// From Solar Fire meta: JDE = 2450298.188218, DeltaT = +62s
	// JD_UT = 2450298.188218 - 62/86400 ≈ 2450298.187502
	// Note: SF testcase 2 DOES use Chiron as a natal reference point
	return TestCase{
		Name:        "XB - Female Chart",
		NatalJD:     2450298.187502,
		NatalLat:    30.867, // 30°52'N
		NatalLon:    120.1,  // 120°06'E
		Timezone:    "Australia/Perth",
		TzAbbr:      "AWST",
		SFCSVPaths:  []string{"testdata/solarfire/testcase-2-transit-1996-2001.csv", "testdata/solarfire/testcase-2-transit-2001-2006.csv"},
		UseMeanNode: true,
		ExpectedASC: 64.396,  // 04°Gemini23'46''
		ExpectedMC:  316.689, // 16°Aquarius41'20''
		// SF meta precise values for MC progression
		NatalASC:       64.3961,  // SF meta: 04°Gemini23'46'' (for Tr-Na reference)
		NatalMC:        316.6889, // SF meta: 16°Aquarius41'20'' (for progressed MC)
		// For testcase-2, SF uses sweph-computed MC for ASC progression (not the meta MC)
		NatalMCForASC:  -1,       // Force sweph-computed MC for ASC progression
		// SF testcase 2: Chiron IS used as natal reference point
		NatalPlanets: []models.PlanetID{
			models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
			models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
			models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune,
			models.PlanetPluto, models.PlanetChiron, models.PlanetNorthNodeMean,
		},
		EventConfig: models.EventConfig{
			IncludeTrNa:         true,
			IncludeTrTr:         false, // Not in SF output
			IncludeTrSp:         false, // Not in SF output
			IncludeTrSa:         false, // Not in SF output
			IncludeSpNa:         true,
			IncludeSpSp:         true,
			IncludeSaNa:         false, // Not in SF output
			IncludeSignIngress:  true,
			IncludeHouseIngress: true,
			IncludeStation:      true,
		},
		IncludeVOC: false, // Not in SF output
	}
}

func runTestCase(tc TestCase) {
	// Verify natal chart against Solar Fire meta
	fmt.Println("=== Natal Chart Verification ===")
	verifyNatalChart(tc)

	// Determine transit period from CSV
	startJD, endJD := getTransitPeriod(tc)
	fmt.Printf("\nTransit period: JD %.6f to %.6f\n", startJD, endJD)

	// ===== Planets =====
	// Default planet lists
	allPlanetsWithChiron := []models.PlanetID{
		models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
		models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
		models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune,
		models.PlanetPluto, models.PlanetChiron, models.PlanetNorthNodeMean,
	}

	// NatalPlanets: for Tr-Na reference points (test case specific)
	natalPlanets := tc.NatalPlanets
	if len(natalPlanets) == 0 {
		natalPlanets = allPlanetsWithChiron
	}

	// ProgressionsPlanets: for Tr-Sp, Sp-Na, Sp-Sp (includes Chiron by default)
	progPlanets := tc.ProgressionsPlanets
	if len(progPlanets) == 0 {
		progPlanets = allPlanetsWithChiron
	}

	// SolarArcPlanets: for Tr-Sa, Sa-Na (includes Chiron by default)
	saPlanets := tc.SolarArcPlanets
	if len(saPlanets) == 0 {
		saPlanets = allPlanetsWithChiron
	}

	natalPoints := []models.SpecialPointID{models.PointASC, models.PointMC}

	// ===== Orbs (Solar Fire: 入1出1 for major aspects, no minor aspects) =====
	// Solar Fire only outputs: Conjunction, Opposition, Trine, Square, Sextile, Quincunx
	// No Semi-Sextile, Semi-Square, Sesquiquadrate
	allOrbs := models.OrbConfig{
		Conjunction: 1, Opposition: 1, Trine: 1, Square: 1,
		Sextile: 1, Quincunx: 1,
		SemiSextile: -1, SemiSquare: 1, Sesquiquadrate: 1,
	}
	transitOrbs := allOrbs
	progOrbs := allOrbs
	saOrbs := allOrbs

	// ===== Outer planets for transit aspects =====
	outerPlanets := []models.PlanetID{
		models.PlanetJupiter, models.PlanetSaturn, models.PlanetUranus,
		models.PlanetNeptune, models.PlanetPluto, models.PlanetChiron,
		models.PlanetNorthNodeMean,
	}

	// ===== Run 1: Outer planets transit + progressions + solar arc =====
	fmt.Println("\n=== Running Transit Calculation (outer planets) ===")
	events1, err := transit.CalcTransitEvents(transit.TransitCalcInput{
		NatalLat:     tc.NatalLat,
		NatalLon:     tc.NatalLon,
		NatalJD:      tc.NatalJD,
		NatalPlanets: natalPlanets,
		NatalASC:                tc.NatalASC,                // Override from SF meta
		NatalMC:                 tc.NatalMC,                 // Override from SF meta
		NatalMCForASC:           tc.NatalMCForASC,           // Separate override for ASC progression
		NatalASCForProgressions: tc.NatalASCForProgressions, // Use direct solar arc for ASC progression (SF style)
		NatalPlanetOverrides:    tc.NatalPlanetOverrides,    // SF precise natal positions
		TransitLat:   tc.NatalLat,
		TransitLon:   tc.NatalLon,
		StartJD:      startJD,
		EndJD:        endJD,
		TransitPlanets: outerPlanets,
		ProgressionsConfig: &models.ProgressionsConfig{
			Enabled: true,
			Planets: progPlanets,
		},
		SolarArcConfig: &models.SolarArcConfig{
			Enabled: true,
			Planets: saPlanets,
		},
		SpecialPoints: &models.SpecialPointsConfig{
			NatalPoints:        natalPoints,
			ProgressionsPoints: natalPoints,
			SolarArcPoints:     []models.SpecialPointID{models.PointASC, models.PointMC},
		},
		EventConfig:          tc.EventConfig,
		OrbConfigTransit:     transitOrbs,
		OrbConfigProgressions: progOrbs,
		OrbConfigSolarArc:    saOrbs,
		HouseSystem:          models.HousePlacidus,
	})
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	// ===== Run 2: Moon sign ingress + VOC (if enabled) =====
	var events2 []models.TransitEvent
	if tc.IncludeVOC || tc.EventConfig.IncludeSignIngress {
		moonTransitPlanets := append([]models.PlanetID{models.PlanetMoon}, outerPlanets...)
		moonTransitPlanets = append(moonTransitPlanets, models.PlanetSun, models.PlanetMercury,
			models.PlanetVenus, models.PlanetMars)
		fmt.Println("Running Moon transit events...")
		events2, err = transit.CalcTransitEvents(transit.TransitCalcInput{
			NatalLat:             tc.NatalLat,
			NatalLon:             tc.NatalLon,
			NatalJD:              tc.NatalJD,
			NatalPlanets:         natalPlanets,
			NatalASC:             tc.NatalASC,             // Override from SF meta
			NatalMC:              tc.NatalMC,              // Override from SF meta
			NatalPlanetOverrides: tc.NatalPlanetOverrides, // SF precise natal positions
			TransitLat:           tc.NatalLat,
			TransitLon:           tc.NatalLon,
			StartJD:              startJD,
			EndJD:                endJD,
			TransitPlanets: moonTransitPlanets,
			SpecialPoints: &models.SpecialPointsConfig{
				NatalPoints: natalPoints,
			},
			EventConfig: models.EventConfig{
				IncludeTrTr:        tc.EventConfig.IncludeTrTr,
				IncludeSignIngress: tc.EventConfig.IncludeSignIngress,
				IncludeVoidOfCourse: tc.IncludeVOC,
			},
			OrbConfigTransit: transitOrbs,
			HouseSystem:      models.HousePlacidus,
		})
		if err != nil {
			fmt.Printf("ERROR (Moon): %v\n", err)
			return
		}
	}

	// Merge events
	var events []models.TransitEvent
	events = append(events, events1...)
	for _, e := range events2 {
		// Only add Moon sign ingress and VOC events
		if e.Planet == models.PlanetMoon &&
			(e.EventType == models.EventSignIngress ||
				(tc.IncludeVOC && e.EventType == models.EventVoidOfCourse)) {
			events = append(events, e)
		}
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].JD < events[j].JD
	})
	fmt.Printf("Generated %d events (outer: %d, moon sig+voc: %d)\n",
		len(events), len(events1), len(events)-len(events1))

	// Count by type
	typeCounts := make(map[string]int)
	for _, e := range events {
		typeCounts[string(e.EventType)]++
	}
	for t, c := range typeCounts {
		fmt.Printf("  %s: %d\n", t, c)
	}

	// ===== Compare with Solar Fire =====
	fmt.Println("\n=== Comparison with Solar Fire ===")
	compareSolarFire(tc.SFCSVPaths, events, tc.Timezone)
}

func getTransitPeriod(tc TestCase) (float64, float64) {
	// Read first date from first CSV and last date from last CSV
	if len(tc.SFCSVPaths) == 0 {
		fmt.Println("ERROR: no CSV paths")
		return 0, 0
	}

	// Get start date from first CSV
	f, err := os.Open(tc.SFCSVPaths[0])
	if err != nil {
		fmt.Printf("ERROR opening first CSV: %v\n", err)
		return 0, 0
	}
	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	f.Close()
	if err != nil || len(records) < 2 {
		fmt.Printf("ERROR reading first CSV\n")
		return 0, 0
	}
	startDate := records[1][7] // Date column from first data row

	// Get end date from last CSV
	lastCSV := tc.SFCSVPaths[len(tc.SFCSVPaths)-1]
	f, err = os.Open(lastCSV)
	if err != nil {
		fmt.Printf("ERROR opening last CSV: %v\n", err)
		return 0, 0
	}
	reader = csv.NewReader(f)
	records, err = reader.ReadAll()
	f.Close()
	if err != nil || len(records) < 2 {
		fmt.Printf("ERROR reading last CSV\n")
		return 0, 0
	}
	endDate := records[len(records)-1][7] // Date column from last data row

	// Convert to JD (assume AWST = UTC+8)
	// Add 1 day buffer at end to ensure boundary events are captured
	startJDResult, _ := julian.DateTimeToJD(startDate+"T00:00:00+08:00", models.CalendarGregorian)
	endJDResult, _ := julian.DateTimeToJD(endDate+"T23:59:59+08:00", models.CalendarGregorian)

	return startJDResult.JDUT, endJDResult.JDUT + 1.0
}

func verifyNatalChart(tc TestCase) {
	planets := []models.PlanetID{
		models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
		models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
		models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune,
		models.PlanetPluto, models.PlanetChiron, models.PlanetNorthNodeMean,
	}

	chartInfo, err := chart.CalcSingleChart(tc.NatalLat, tc.NatalLon, tc.NatalJD, planets, models.DefaultOrbConfig(), models.HousePlacidus)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	fmt.Printf("ASC: %.4f (%s) [expected: %.3f]\n",
		chartInfo.Angles.ASC, models.FormatLonDMS(chartInfo.Angles.ASC), tc.ExpectedASC)
	fmt.Printf("MC:  %.4f (%s) [expected: %.3f]\n",
		chartInfo.Angles.MC, models.FormatLonDMS(chartInfo.Angles.MC), tc.ExpectedMC)

	// Check ASC/MC match
	ascDiff := chartInfo.Angles.ASC - tc.ExpectedASC
	mcDiff := chartInfo.Angles.MC - tc.ExpectedMC
	if abs(ascDiff) > 0.1 || abs(mcDiff) > 0.1 {
		fmt.Printf("WARNING: ASC/MC mismatch! ASC diff: %.4f, MC diff: %.4f\n", ascDiff, mcDiff)
	} else {
		fmt.Println("ASC/MC verification: OK")
	}

	fmt.Println("\nPlanet positions:")
	for _, p := range chartInfo.Planets {
		retro := ""
		if p.IsRetrograde {
			retro = " (Rx)"
		}
		fmt.Printf("  %-18s %10.4f  %s  House %d%s\n",
			p.PlanetID, p.Longitude, models.FormatLonDMS(p.Longitude), p.House, retro)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func compareSolarFire(sfPaths []string, computedEvents []models.TransitEvent, tz string) {
	// Read all Solar Fire CSVs
	var allSFRows [][]string
	for _, path := range sfPaths {
		f, err := os.Open(path)
		if err != nil {
			fmt.Printf("ERROR opening Solar Fire CSV %s: %v\n", path, err)
			continue
		}
		reader := csv.NewReader(f)
		records, err := reader.ReadAll()
		f.Close()
		if err != nil {
			fmt.Printf("ERROR reading CSV %s: %v\n", path, err)
			continue
		}
		if len(records) < 2 {
			continue
		}
		// Skip header for all but first file
		if len(allSFRows) == 0 {
			allSFRows = append(allSFRows, records[1:]...)
		} else {
			allSFRows = append(allSFRows, records[1:]...)
		}
	}

	fmt.Printf("Solar Fire events: %d (from %d files)\n", len(allSFRows), len(sfPaths))
	fmt.Printf("Computed events:   %d\n", len(computedEvents))

	// Count event types
	sfTypeCounts := make(map[string]int)
	sfChartCounts := make(map[string]int)
	for _, row := range allSFRows {
		if len(row) >= 7 {
			sfTypeCounts[row[5]]++
			sfChartCounts[row[6]]++
		}
	}
	fmt.Println("\nSolar Fire event types:")
	for t, c := range sfTypeCounts {
		fmt.Printf("  %s: %d\n", t, c)
	}
	fmt.Println("\nSolar Fire chart types:")
	for t, c := range sfChartCounts {
		fmt.Printf("  %s: %d\n", t, c)
	}

	// Count computed events
	compTypeCounts := make(map[string]int)
	compChartCounts := make(map[string]int)
	for _, e := range computedEvents {
		csvEvtType := models.EventTypeCSV(e.EventType, e.StationType)
		compTypeCounts[csvEvtType]++
		row := export.EventToCSVRow(e, tz)
		compChartCounts[row.Type]++
	}
	fmt.Println("\nComputed event types:")
	for t, c := range compTypeCounts {
		fmt.Printf("  %s: %d\n", t, c)
	}
	fmt.Println("\nComputed chart types:")
	for t, c := range compChartCounts {
		fmt.Printf("  %s: %d\n", t, c)
	}

	// Match exact events
	fmt.Println("\n=== Event Matching (Exact events only) ===")
	matchExactEvents(allSFRows, computedEvents, tz)
}

func makeKey(date, time, p1, p2, aspect, pairType string) string {
	if p1 > p2 {
		p1, p2 = p2, p1
	}
	key := fmt.Sprintf("%s %s|%s|%s|%s|%s", date, time, p1, p2, aspect, pairType)
	// Debug: show first few keys
	staticCounter := 0
	staticCounter++
	if staticCounter <= 3 {
		fmt.Printf("DEBUG KEY: '%s'\n", key)
	}
	return key
}

// makeFuzzyTimeKeys generates keys with ultra-precise time tolerance in seconds
func makeFuzzyTimeKeys(date, timeStr, p1, p2, aspect, pairType string, toleranceSeconds int) []string {
	keys := []string{makeKey(date, timeStr, p1, p2, aspect, pairType)}

	// Parse the time
	t, err := time.Parse("2006-01-02 15:04:05", date+" "+timeStr)
	if err != nil {
		return keys
	}

	// Generate keys at 1-second intervals for ultra-precision
	for offset := -toleranceSeconds; offset <= toleranceSeconds; offset++ { // 1 second steps
		if offset == 0 {
			continue
		}
		adjusted := t.Add(time.Duration(offset) * time.Second)
		keys = append(keys, makeKey(
			adjusted.Format("2006-01-02"),
			adjusted.Format("15:04:05"),
			p1, p2, aspect, pairType))
	}
	return keys
}

func matchExactEvents(sfRows [][]string, computedEvents []models.TransitEvent, tz string) {
	// Find SF date range
	sfEndDate := ""
	for _, sfRow := range sfRows {
		if len(sfRow) > 7 && sfRow[7] > sfEndDate {
			sfEndDate = sfRow[7]
		}
	}

	// Build computed event list with date/time info
	type computedInfo struct {
		event models.TransitEvent
		row   export.CSVRow
	}
	var computedExacts []computedInfo
	for _, e := range computedEvents {
		if e.EventType != models.EventAspectExact {
			continue
		}
		row := export.EventToCSVRow(e, tz)
		if sfEndDate != "" && row.Date > sfEndDate {
			continue
		}
		computedExacts = append(computedExacts, computedInfo{event: e, row: row})
	}

	// For each SF event, find the closest computed event by identity (planets+aspect+type),
	// then report the time difference
	type diffResult struct {
		sfDate, sfTime     string
		compDate, compTime string
		diffSeconds        int
		p1, aspect, p2     string
		chartType          string
	}
	var results []diffResult
	unmatchedSF := 0

	sfExactCount := 0
	for _, sfRow := range sfRows {
		if len(sfRow) < 17 || sfRow[5] != "Exact" {
			continue
		}
		sfExactCount++

		sfP1 := sfRow[0]
		sfP2 := sfRow[3]
		sfAspect := sfRow[2]
		sfType := sfRow[6]
		sfDate := sfRow[7]
		sfTime := sfRow[8]

		// Normalize planet order
		p1, p2 := sfP1, sfP2
		if p1 > p2 {
			p1, p2 = p2, p1
		}

		// Parse SF datetime
		sfDT, err := time.Parse("2006-01-02 15:04:05", sfDate+" "+sfTime)
		if err != nil {
			continue
		}

		// Find closest computed event with same identity
		bestDiff := int(1e9)
		bestIdx := -1
		for i, ce := range computedExacts {
			cp1, cp2 := ce.row.P1, ce.row.P2
			if cp1 > cp2 {
				cp1, cp2 = cp2, cp1
			}
			if cp1 != p1 || cp2 != p2 || ce.row.Aspect != sfAspect || ce.row.Type != sfType {
				continue
			}
			compDT, err2 := time.Parse("2006-01-02 15:04:05", ce.row.Date+" "+ce.row.Time)
			if err2 != nil {
				continue
			}
			diff := int(compDT.Sub(sfDT).Seconds())
			if diff < 0 {
				diff = -diff
			}
			if diff < bestDiff {
				bestDiff = diff
				bestIdx = i
			}
		}

		if bestIdx >= 0 {
			ce := computedExacts[bestIdx]
			compDT, _ := time.Parse("2006-01-02 15:04:05", ce.row.Date+" "+ce.row.Time)
			diff := int(compDT.Sub(sfDT).Seconds())
			results = append(results, diffResult{
				sfDate: sfDate, sfTime: sfTime,
				compDate: ce.row.Date, compTime: ce.row.Time,
				diffSeconds: diff,
				p1: sfP1, aspect: sfAspect, p2: sfP2,
				chartType: sfType,
			})
		} else {
			unmatchedSF++
			fmt.Printf("  NO MATCH: %s %s %s %s %s %s\n", sfDate, sfTime, sfP1, sfAspect, sfP2, sfType)
		}
	}

	// Sort results by absolute diff
	sort.Slice(results, func(i, j int) bool {
		ai := results[i].diffSeconds
		if ai < 0 { ai = -ai }
		aj := results[j].diffSeconds
		if aj < 0 { aj = -aj }
		return ai < aj
	})

	// Print diff distribution
	fmt.Printf("\n=== Time Difference Distribution (all %d matched events) ===\n", len(results))

	// Count by buckets
	within1s, within10s, within60s, within5m, within30m, within1h, beyond1h := 0, 0, 0, 0, 0, 0, 0
	for _, r := range results {
		absDiff := r.diffSeconds
		if absDiff < 0 { absDiff = -absDiff }
		switch {
		case absDiff <= 1:
			within1s++
		case absDiff <= 10:
			within10s++
		case absDiff <= 60:
			within60s++
		case absDiff <= 300:
			within5m++
		case absDiff <= 1800:
			within30m++
		case absDiff <= 3600:
			within1h++
		default:
			beyond1h++
		}
	}
	fmt.Printf("  <=1s:    %d\n", within1s)
	fmt.Printf("  2-10s:   %d\n", within10s)
	fmt.Printf("  11-60s:  %d\n", within60s)
	fmt.Printf("  1-5m:    %d\n", within5m)
	fmt.Printf("  5-30m:   %d\n", within30m)
	fmt.Printf("  30m-1h:  %d\n", within1h)
	fmt.Printf("  >1h:     %d\n", beyond1h)
	fmt.Printf("  No match: %d\n", unmatchedSF)

	// Print by chart type
	typeStats := make(map[string][]int) // chartType -> list of diffs
	for _, r := range results {
		typeStats[r.chartType] = append(typeStats[r.chartType], r.diffSeconds)
	}
	fmt.Printf("\n=== By Chart Type ===\n")
	var types []string
	for t := range typeStats {
		types = append(types, t)
	}
	sort.Strings(types)
	for _, t := range types {
		diffs := typeStats[t]
		var sumAbs int
		maxAbs := 0
		for _, d := range diffs {
			ad := d
			if ad < 0 { ad = -ad }
			sumAbs += ad
			if ad > maxAbs { maxAbs = ad }
		}
		avgAbs := float64(sumAbs) / float64(len(diffs))
		fmt.Printf("  %s: %d events, avg |diff|=%.0fs (%.1fmin), max |diff|=%ds (%.1fmin)\n",
			t, len(diffs), avgAbs, avgAbs/60, maxAbs, float64(maxAbs)/60)
	}

	// Print worst 20 events (largest diff)
	fmt.Printf("\n=== Worst 20 (largest time diff) ===\n")
	for i := len(results) - 1; i >= 0 && i >= len(results)-20; i-- {
		r := results[i]
		fmt.Printf("  %+6ds (%+.1fmin) | SF %s %s | Comp %s %s | %s %s %s %s\n",
			r.diffSeconds, float64(r.diffSeconds)/60,
			r.sfDate, r.sfTime, r.compDate, r.compTime,
			r.p1, r.aspect, r.p2, r.chartType)
	}

	// Print best 20 events (smallest diff)
	fmt.Printf("\n=== Best 20 (smallest time diff) ===\n")
	limit := 20
	if limit > len(results) { limit = len(results) }
	for i := 0; i < limit; i++ {
		r := results[i]
		fmt.Printf("  %+6ds (%+.1fmin) | SF %s %s | Comp %s %s | %s %s %s %s\n",
			r.diffSeconds, float64(r.diffSeconds)/60,
			r.sfDate, r.sfTime, r.compDate, r.compTime,
			r.p1, r.aspect, r.p2, r.chartType)
	}

	fmt.Printf("\nSF Exact: %d, Paired: %d, No match: %d\n", sfExactCount, len(results), unmatchedSF)

	// Print ALL Tr-Na events sorted by diff
	fmt.Printf("\n=== ALL Tr-Na Events (sorted by time diff) ===\n")
	var trNaResults []diffResult
	for _, r := range results {
		if r.chartType == "Tr-Na" {
			trNaResults = append(trNaResults, r)
		}
	}
	// Sort by absolute diff
	sort.Slice(trNaResults, func(i, j int) bool {
		ai := trNaResults[i].diffSeconds
		if ai < 0 { ai = -ai }
		aj := trNaResults[j].diffSeconds
		if aj < 0 { aj = -aj }
		return ai < aj
	})
	for _, r := range trNaResults {
		fmt.Printf("  %+6ds (%+.1fmin) | SF %s %s | Comp %s %s | %s %s %s\n",
			r.diffSeconds, float64(r.diffSeconds)/60,
			r.sfDate, r.sfTime, r.compDate, r.compTime,
			r.p1, r.aspect, r.p2)
	}
}
