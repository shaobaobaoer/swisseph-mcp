package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/anthropic/swisseph-mcp/pkg/chart"
	"github.com/anthropic/swisseph-mcp/pkg/export"
	"github.com/anthropic/swisseph-mcp/pkg/julian"
	"github.com/anthropic/swisseph-mcp/pkg/models"
	"github.com/anthropic/swisseph-mcp/pkg/sweph"
	"github.com/anthropic/swisseph-mcp/pkg/transit"
)

func main() {
	// Initialize ephemeris
	exe, _ := os.Executable()
	ephePath := filepath.Join(filepath.Dir(exe), "..", "..", "third_party", "swisseph", "ephe")
	if _, err := os.Stat(ephePath); err != nil {
		ephePath = filepath.Join(".", "third_party", "swisseph", "ephe")
	}
	sweph.Init(ephePath)
	defer sweph.Close()

	// ===== JN Natal Chart Parameters =====
	// From Solar Fire meta: JDE = 2450800.900729, DeltaT = +62s
	// JD_UT = 2450800.900729 - 62/86400 = 2450800.900011
	// Location derived by matching ASC/MC from Solar Fire meta exactly:
	natalJD := 2450800.900011
	natalLat := 30.902808
	natalLon := 121.146738

	// Verify natal chart against Solar Fire meta
	fmt.Println("=== Natal Chart Verification ===")
	verifyNatalChart(natalLat, natalLon, natalJD)

	// ===== Transit Period =====
	// 2026-02-01 00:00:00 AWST (UTC+8) to 2027-02-01 00:00:00 AWST
	startJDResult, err := julian.DateTimeToJD("2026-02-01T00:00:00+08:00", models.CalendarGregorian)
	if err != nil {
		fmt.Printf("ERROR computing start JD: %v\n", err)
		return
	}
	endJDResult, err := julian.DateTimeToJD("2027-02-01T00:00:00+08:00", models.CalendarGregorian)
	if err != nil {
		fmt.Printf("ERROR computing end JD: %v\n", err)
		return
	}
	startJD := startJDResult.JDUT
	endJD := endJDResult.JDUT

	fmt.Printf("\nTransit period: JD %.6f to %.6f\n", startJD, endJD)

	// ===== Planets =====
	// Solar Fire uses Mean Node (verified by position match)
	allPlanets := []models.PlanetID{
		models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
		models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
		models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune,
		models.PlanetPluto, models.PlanetChiron, models.PlanetNorthNodeMean,
	}

	natalPoints := []models.SpecialPointID{models.PointASC, models.PointMC}

	// ===== Orbs (Solar Fire: all 9 aspects at 1 degree) =====
	allOrbs := models.OrbConfig{
		Conjunction: 1, Opposition: 1, Trine: 1, Square: 1,
		Sextile: 1, Quincunx: 1,
		SemiSextile: 0, SemiSquare: 1, Sesquiquadrate: 1,
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
		NatalLat:     natalLat,
		NatalLon:     natalLon,
		NatalJD:      natalJD,
		NatalPlanets: allPlanets,
		TransitLat:   natalLat,
		TransitLon:   natalLon,
		StartJD:      startJD,
		EndJD:        endJD,
		TransitPlanets: outerPlanets,
		ProgressionsConfig: &models.ProgressionsConfig{
			Enabled: true,
			Planets: allPlanets,
		},
		SolarArcConfig: &models.SolarArcConfig{
			Enabled: true,
			Planets: allPlanets,
		},
		SpecialPoints: &models.SpecialPointsConfig{
			NatalPoints:        natalPoints,
			ProgressionsPoints: natalPoints,
			SolarArcPoints:     []models.SpecialPointID{models.PointASC, models.PointMC},
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
		OrbConfigTransit:      transitOrbs,
		OrbConfigProgressions: progOrbs,
		OrbConfigSolarArc:     saOrbs,
		HouseSystem:           models.HousePlacidus,
	})
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	// ===== Run 2: Moon sign ingress + VOC =====
	// Moon needs all transit planets to form Tr-Tr aspects (for VOC detection)
	moonTransitPlanets := append([]models.PlanetID{models.PlanetMoon}, outerPlanets...)
	moonTransitPlanets = append(moonTransitPlanets, models.PlanetSun, models.PlanetMercury,
		models.PlanetVenus, models.PlanetMars)
	fmt.Println("Running Moon transit events...")
	events2, err := transit.CalcTransitEvents(transit.TransitCalcInput{
		NatalLat:     natalLat,
		NatalLon:     natalLon,
		NatalJD:      natalJD,
		NatalPlanets: allPlanets,
		TransitLat:   natalLat,
		TransitLon:   natalLon,
		StartJD:      startJD,
		EndJD:        endJD,
		TransitPlanets: moonTransitPlanets,
		SpecialPoints: &models.SpecialPointsConfig{
			NatalPoints: natalPoints,
		},
		EventConfig: models.EventConfig{
			IncludeTrTr:         true,
			IncludeSignIngress:  true,
			IncludeVoidOfCourse: true,
		},
		OrbConfigTransit: transitOrbs,
		HouseSystem:      models.HousePlacidus,
	})
	if err != nil {
		fmt.Printf("ERROR (Moon): %v\n", err)
		return
	}

	// Merge: keep all from run1, keep only Moon SignIngress and VOC from run2
	// (Solar Fire Tr-Tr for Moon only shows SignIngress and VOC, not individual aspects)
	var events []models.TransitEvent
	events = append(events, events1...)
	for _, e := range events2 {
		if e.Planet == models.PlanetMoon &&
			(e.EventType == models.EventSignIngress || e.EventType == models.EventVoidOfCourse) {
			events = append(events, e)
		}
	}
	// Sort by JD
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

	// ===== Export CSV =====
	csvOutput := export.EventsToCSV(events, "Australia/Perth", "AWST")
	outputPath := "/opt/share/JN_transit_computed.csv"
	if err := os.WriteFile(outputPath, []byte(csvOutput), 0644); err != nil {
		fmt.Printf("ERROR writing CSV: %v\n", err)
		return
	}
	fmt.Printf("\nCSV written to %s\n", outputPath)

	// ===== Compare with Solar Fire =====
	fmt.Println("\n=== Comparison with Solar Fire ===")
	sfPath := "/opt/share/JN_transit (1).csv"
	compareSolarFire(sfPath, events)
}

func verifyNatalChart(lat, lon, jdUT float64) {
	planets := []models.PlanetID{
		models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
		models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
		models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune,
		models.PlanetPluto, models.PlanetChiron, models.PlanetNorthNodeTrue,
	}

	chartInfo, err := chart.CalcSingleChart(lat, lon, jdUT, planets, models.DefaultOrbConfig(), models.HousePlacidus)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	// Expected from Solar Fire meta:
	// ASC: 06 Cancer 31'45'' = 96.529
	// MC:  21 Pisces 29'51'' = 351.498
	fmt.Printf("ASC: %.4f (%s) [expected: 96.529 Cancer]\n",
		chartInfo.Angles.ASC, models.FormatLonDMS(chartInfo.Angles.ASC))
	fmt.Printf("MC:  %.4f (%s) [expected: 351.498 Pisces]\n",
		chartInfo.Angles.MC, models.FormatLonDMS(chartInfo.Angles.MC))

	// Expected planet positions from Solar Fire meta:
	expected := map[string]string{
		"SUN":             "26Sag29'59\"",
		"MOON":            "18Leo06'58\"",
		"MERCURY":         "23Sag56'00\"",
		"VENUS":           "02Aqu33'09\"",
		"MARS":            "00Aqu05'50\"",
		"JUPITER":         "19Aqu32'52\"",
		"SATURN":          "13Ari32'15\"",
		"URANUS":          "06Aqu25'29\"",
		"NEPTUNE":         "28Cap28'03\"",
		"PLUTO":           "06Sag16'27\"",
		"CHIRON":          "14Sco01'21\"",
		"NORTH_NODE_TRUE": "14Vir26'45\"",
	}
	fmt.Println("\nPlanet positions:")
	for _, p := range chartInfo.Planets {
		expStr := expected[string(p.PlanetID)]
		computed := models.FormatLonDMS(p.Longitude)
		match := "OK"
		if expStr != "" && computed != expStr {
			match = fmt.Sprintf("MISMATCH (expected %s)", expStr)
		}
		retro := ""
		if p.IsRetrograde {
			retro = " (Rx)"
		}
		fmt.Printf("  %-18s %10.4f  %s  House %d%s  %s\n",
			p.PlanetID, p.Longitude, computed, p.House, retro, match)
	}
}

func compareSolarFire(sfPath string, computedEvents []models.TransitEvent) {
	f, err := os.Open(sfPath)
	if err != nil {
		fmt.Printf("ERROR opening Solar Fire CSV: %v\n", err)
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("ERROR reading CSV: %v\n", err)
		return
	}

	if len(records) < 2 {
		fmt.Println("Solar Fire CSV has no data rows")
		return
	}

	// Skip header
	sfRows := records[1:]
	fmt.Printf("Solar Fire events: %d\n", len(sfRows))
	fmt.Printf("Computed events:   %d\n", len(computedEvents))

	// Count Solar Fire events by type
	sfTypeCounts := make(map[string]int)
	sfChartCounts := make(map[string]int)
	for _, row := range sfRows {
		if len(row) >= 7 {
			sfTypeCounts[row[5]]++ // EventType column
			sfChartCounts[row[6]]++ // Type column
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

	// Count computed events by equivalent types
	compTypeCounts := make(map[string]int)
	compChartCounts := make(map[string]int)
	for _, e := range computedEvents {
		csvEvtType := models.EventTypeCSV(e.EventType, e.StationType)
		compTypeCounts[csvEvtType]++

		row := export.EventToCSVRow(e, "Australia/Perth")
		compChartCounts[row.Type]++
	}
	fmt.Println("\nComputed event types (CSV names):")
	for t, c := range compTypeCounts {
		fmt.Printf("  %s: %d\n", t, c)
	}
	fmt.Println("\nComputed chart types:")
	for t, c := range compChartCounts {
		fmt.Printf("  %s: %d\n", t, c)
	}

	// Show first 20 Solar Fire events and first 20 computed events side by side
	fmt.Println("\n=== First 20 Solar Fire events ===")
	for i := 0; i < 20 && i < len(sfRows); i++ {
		if len(sfRows[i]) >= 17 {
			row := sfRows[i]
			fmt.Printf("  %s %s %s %s | %s | %s | %s %s | %s %s %s | %s %s %s\n",
				row[0], row[2], row[3], row[5], row[6], row[7], row[8],
				row[10], row[11], row[12], row[13], row[14], row[15], row[16])
		}
	}

	fmt.Println("\n=== First 20 computed events ===")
	for i := 0; i < 20 && i < len(computedEvents); i++ {
		row := export.EventToCSVRow(computedEvents[i], "Asia/Shanghai")
		line := export.CSVRowToString(row)
		// Truncate for display
		if len(line) > 120 {
			line = line[:120] + "..."
		}
		fmt.Printf("  %s\n", line)
	}

	// Match events by date and planet pair
	fmt.Println("\n=== Event Matching (Exact events only) ===")
	matchExactEvents(sfRows, computedEvents)
}

// makeKey creates an order-independent key for matching events
func makeKey(date, p1, p2, aspect, pairType string) string {
	// Sort p1, p2 for order-independent matching
	if p1 > p2 {
		p1, p2 = p2, p1
	}
	return fmt.Sprintf("%s|%s|%s|%s|%s", date, p1, p2, aspect, pairType)
}

func matchExactEvents(sfRows [][]string, computedEvents []models.TransitEvent) {
	// Build lookup for computed Exact events
	computedExacts := make(map[string]models.TransitEvent)
	for _, e := range computedEvents {
		if e.EventType != models.EventAspectExact {
			continue
		}
		row := export.EventToCSVRow(e, "Australia/Perth")
		key := makeKey(row.Date, row.P1, row.P2, row.Aspect, row.Type)
		computedExacts[key] = e
	}

	// Check SF Exact events
	matched := 0
	unmatched := 0
	sfExactCount := 0
	for _, sfRow := range sfRows {
		if len(sfRow) < 17 || sfRow[5] != "Exact" {
			continue
		}
		sfExactCount++
		key := makeKey(sfRow[7], sfRow[0], sfRow[3], sfRow[2], sfRow[6])
		if ce, ok := computedExacts[key]; ok {
			matched++
			// Show time comparison for first few matches
			if matched <= 5 {
				row := export.EventToCSVRow(ce, "Australia/Perth")
				fmt.Printf("  MATCH: SF %s %s %s %s %s %s | Computed %s %s %s\n",
					sfRow[7], sfRow[8], sfRow[0], sfRow[2], sfRow[3], sfRow[6],
					row.Date, row.Time, row.Aspect)
			}
		} else {
			if unmatched < 30 {
				fmt.Printf("  SF unmatched: %s %s %s %s %s %s\n",
					sfRow[7], sfRow[8], sfRow[0], sfRow[2], sfRow[3], sfRow[6])
			}
			unmatched++
		}
	}
	fmt.Printf("\nSF Exact events: %d, Matched: %d, Unmatched: %d\n",
		sfExactCount, matched, unmatched)

	// Check for computed Exact events not in SF
	sfExactKeys := make(map[string]bool)
	for _, sfRow := range sfRows {
		if len(sfRow) < 17 || sfRow[5] != "Exact" {
			continue
		}
		key := makeKey(sfRow[7], sfRow[0], sfRow[3], sfRow[2], sfRow[6])
		sfExactKeys[key] = true
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
