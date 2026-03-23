package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	return TestCase{
		Name:        "JN - Male Chart",
		NatalJD:     2450800.900009,
		NatalLat:    30.9,   // 30°54'N
		NatalLon:    121.15, // 121°09'E
		Timezone:    "Australia/Perth",
		TzAbbr:      "AWST",
		SFCSVPaths:  []string{"testdata/solarfire/testcase-1-transit.csv"},
		UseMeanNode: true,
		ExpectedASC: 96.530,  // 06°Cancer31'45''
		ExpectedMC:  351.500, // 21°Pisces29'51''
		// SF testcase 1: Chiron is NOT used as natal reference point
		NatalPlanets: []models.PlanetID{
			models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
			models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
			models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune,
			models.PlanetPluto, models.PlanetNorthNodeMean,
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
			NatalLat:     tc.NatalLat,
			NatalLon:     tc.NatalLon,
			NatalJD:      tc.NatalJD,
			NatalPlanets: natalPlanets,
			TransitLat:   tc.NatalLat,
			TransitLon:   tc.NatalLon,
			StartJD:      startJD,
			EndJD:        endJD,
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

func makeKey(date, p1, p2, aspect, pairType string) string {
	if p1 > p2 {
		p1, p2 = p2, p1
	}
	return fmt.Sprintf("%s|%s|%s|%s|%s", date, p1, p2, aspect, pairType)
}

func makeFuzzyKeys(date, p1, p2, aspect, pairType string) []string {
	keys := []string{makeKey(date, p1, p2, aspect, pairType)}
	t, err := time.Parse("2006-01-02", date)
	if err == nil {
		// ±1 day for transit events
		days := 1
		// ±5 days for Sp-Sp/Sp-Na (progressed planets move very slowly)
		if strings.HasPrefix(pairType, "Sp-") || strings.HasPrefix(pairType, "Sa-") {
			days = 5
		}
		for d := -days; d <= days; d++ {
			if d == 0 {
				continue
			}
			keys = append(keys, makeKey(t.AddDate(0, 0, d).Format("2006-01-02"), p1, p2, aspect, pairType))
		}
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

	computedExacts := make(map[string]models.TransitEvent)
	for _, e := range computedEvents {
		if e.EventType != models.EventAspectExact {
			continue
		}
		row := export.EventToCSVRow(e, tz)
		// Skip events beyond SF date range (from endJD buffer)
		if sfEndDate != "" && row.Date > sfEndDate {
			continue
		}
		key := makeKey(row.Date, row.P1, row.P2, row.Aspect, row.Type)
		computedExacts[key] = e
	}

	matched := 0
	unmatched := 0
	sfExactCount := 0
	for _, sfRow := range sfRows {
		if len(sfRow) < 17 || sfRow[5] != "Exact" {
			continue
		}
		sfExactCount++
		keys := makeFuzzyKeys(sfRow[7], sfRow[0], sfRow[3], sfRow[2], sfRow[6])
		found := false
		for _, key := range keys {
			if ce, ok := computedExacts[key]; ok {
				matched++
				if matched <= 5 {
					row := export.EventToCSVRow(ce, tz)
					fmt.Printf("  MATCH: SF %s %s %s %s %s %s | Computed %s %s %s\n",
						sfRow[7], sfRow[8], sfRow[0], sfRow[2], sfRow[3], sfRow[6],
						row.Date, row.Time, row.Aspect)
				}
				found = true
				break
			}
		}
		if !found {
			if unmatched < 30 {
				fmt.Printf("  SF unmatched: %s %s %s %s %s %s\n",
					sfRow[7], sfRow[8], sfRow[0], sfRow[2], sfRow[3], sfRow[6])
			}
			unmatched++
		}
	}
	fmt.Printf("\nSF Exact events: %d, Matched: %d, Unmatched: %d\n",
		sfExactCount, matched, unmatched)

	// Check for extra computed events (use fuzzy matching like forward match)
	sfExactKeys := make(map[string]bool)
	for _, sfRow := range sfRows {
		if len(sfRow) < 17 || sfRow[5] != "Exact" {
			continue
		}
		// Add all fuzzy keys (date ±1 day, swapped planets)
		for _, key := range makeFuzzyKeys(sfRow[7], sfRow[0], sfRow[3], sfRow[2], sfRow[6]) {
			sfExactKeys[key] = true
		}
	}
	extra := 0
	for key := range computedExacts {
		if !sfExactKeys[key] {
			extra++
			if extra <= 20 {
				parts := strings.Split(key, "|")
				fmt.Printf("  Extra computed: %s\n", strings.Join(parts, " "))
			}
		}
	}
	fmt.Printf("Extra computed Exact events: %d\n", extra)
}
