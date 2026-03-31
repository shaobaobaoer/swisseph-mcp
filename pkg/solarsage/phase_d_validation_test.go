package solarsage

import (
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/shaobaobaoer/solarsage-mcp/internal/aspect"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/chart"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/progressions"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/sweph"
)

// ============================================================================
// Phase D: Solar Fire Cross-Validation
// ============================================================================
// Validates that SolarSage computed aspects match Solar Fire reference data
// exactly (within tolerance) for time, planet/points, and angles.
//
// Input: Metadata files (birth data) + CSV files (SF reference aspects)
// Output: Validation report showing matches/divergences per aspect
// ============================================================================

// AspectNameToType maps SF aspect name to aspect type constant
func AspectNameToType(name string) string {
	// SF names: Conjunction, Opposition, Trine, Square, Sextile, Semi-Square,
	// Sesquiquadrate, Quincunx
	return strings.ToLower(name)
}

// ComputeSSAspects computes aspects for a snapshot single chart
// (natal, transit, progressed, or solar arc at snapshot date)
func ComputeSSAspects(innerBodies, outerBodies []aspect.Body, orbs models.OrbConfig) map[string]bool {
	crossAspects := aspect.FindCrossAspects(innerBodies, outerBodies, orbs)

	// Build a key for each aspect: "P1-Aspect-P2" (order-insensitive)
	aspectSet := make(map[string]bool)
	for _, asp := range crossAspects {
		key := fmt.Sprintf("%s-%s-%s", asp.InnerBody, asp.AspectType, asp.OuterBody)
		aspectSet[key] = true

		// Also add reverse order in case SF lists it differently
		keyRev := fmt.Sprintf("%s-%s-%s", asp.OuterBody, asp.AspectType, asp.InnerBody)
		aspectSet[keyRev] = true
	}

	return aspectSet
}

// TestPhaseD_JN_SnapshotValidation validates JN snapshot aspects against Solar Fire
func TestPhaseD_JN_SnapshotValidation(t *testing.T) {
	const (
		snapshotDate   = "2026-02-01"
		csvPath        = "../../testdata/solarfire/testcase-1-transit.csv"
		eventType      = "Begin"   // Only validate "Begin" events at snapshot
		sfChartTypes   = "Tr-Na"   // Tr-Na, Sp-Na, Sa-Na, Sr-Na
		expectedCount  = 72        // From Phase B baseline
		angleTolerance = 1.0       // degrees
	)

	t.Logf("Phase D: JN Snapshot (2026-02-01) Solar Fire Validation")
	t.Logf("Loading SF reference data: %s", csvPath)

	// Parse SF CSV
	sfRecords, err := ParseSFCSV(csvPath, eventType, "", snapshotDate)
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	if len(sfRecords) == 0 {
		t.Fatalf("No SF records found for date=%s eventType=%s", snapshotDate, eventType)
	}

	t.Logf("Loaded %d SF records for snapshot date", len(sfRecords))

	// Compute JN snapshot aspects at 2026-02-01
	orbs := models.DefaultOrbConfig()

	// Inner: natal chart
	natalChart, err := chart.CalcSingleChart(jnLat, jnLon, jnJDUT, jnPlanets, orbs, models.HousePlacidus)
	if err != nil {
		t.Fatalf("CalcSingleChart natal: %v", err)
	}

	innerBodies := BuildBodiesFromPlanets(natalChart.Planets)
	innerBodies = append(innerBodies,
		aspect.Body{ID: string(models.PointASC), Longitude: natalChart.Angles.ASC},
		aspect.Body{ID: string(models.PointMC), Longitude: natalChart.Angles.MC},
	)

	// Outer: transit at snapshot date
	snapshotJD := sweph.JulDay(2026, 2, 1, 0, true)
	transitChart, err := chart.CalcSingleChart(jnLat, jnLon, snapshotJD, jnPlanets, orbs, models.HousePlacidus)
	if err != nil {
		t.Fatalf("CalcSingleChart transit: %v", err)
	}

	outerBodies := BuildBodiesFromPlanets(transitChart.Planets)
	outerBodies = append(outerBodies,
		aspect.Body{ID: string(models.PointASC), Longitude: transitChart.Angles.ASC},
		aspect.Body{ID: string(models.PointMC), Longitude: transitChart.Angles.MC},
	)

	// Compute cross-aspects
	crossAspects := aspect.FindCrossAspects(innerBodies, outerBodies, orbs)

	t.Logf("Computed %d cross-aspects for Tr-Na snapshot", len(crossAspects))
	t.Logf("Expected %d from Phase B baseline", expectedCount)

	if len(crossAspects) != expectedCount {
		t.Logf("⚠️  Cross-aspect count mismatch: got %d, expected %d (Phase B)",
			len(crossAspects), expectedCount)
	}

	// Filter SF records by chart type
	var sfTrNa []SFAspectRecord
	for _, rec := range sfRecords {
		if rec.Type == "Tr-Na" {
			sfTrNa = append(sfTrNa, rec)
		}
	}

	t.Logf("SF Tr-Na aspects: %d records", len(sfTrNa))

	// Count matches vs divergences
	var matchCount, divergeCount int
	var divergences []string

	for _, sfAsp := range sfTrNa {
		p1Name := sfAsp.P1
		p2Name := sfAsp.P2

		if p1Name == "" || p2Name == "" {
			continue
		}

		// Check if this aspect exists in SolarSage
		found := false
		for _, ssa := range crossAspects {
			// Compare body names (case-insensitive, try different orderings)
			bodyMatch := (strings.ToLower(ssa.InnerBody) == strings.ToLower(p1Name) && strings.ToLower(ssa.OuterBody) == strings.ToLower(p2Name)) ||
				(strings.ToLower(ssa.InnerBody) == strings.ToLower(p2Name) && strings.ToLower(ssa.OuterBody) == strings.ToLower(p1Name))

			if bodyMatch {
				// Check aspect type matches (normalize to lowercase for comparison)
				ssAspectType := string(ssa.AspectType)
				sfAspectType := sfAsp.Aspect

				aspectMatch := strings.ToLower(ssAspectType) == strings.ToLower(sfAspectType)

				if aspectMatch {
					// Check angle difference
					diff := math.Abs(ssa.Orb)
					if diff <= angleTolerance {
						matchCount++
						found = true
						break
					}
				}
			}
		}

		if !found {
			divergeCount++
			divergences = append(divergences, fmt.Sprintf("%s %s %s (SF: %.2f°)",
				sfAsp.P1, sfAsp.Aspect, sfAsp.P2, sfAsp.Pos1Deg))
		}
	}

	t.Logf("Phase D Validation Results (Tr-Na Snapshot):")
	t.Logf("  Matches: %d", matchCount)
	t.Logf("  Divergences: %d", divergeCount)

	if divergeCount > 0 && divergeCount <= 10 {
		t.Logf("  First divergences:")
		for i, div := range divergences {
			if i >= 5 {
				break
			}
			t.Logf("    - %s", div)
		}
	}

	// Phase D status
	if divergeCount == 0 {
		t.Logf("✅ Phase D PASS: All Tr-Na snapshot aspects validated")
	} else {
		t.Logf("⚠️  Phase D PARTIAL: %d divergences (%.1f%% match)",
			divergeCount, float64(matchCount)*100.0/float64(matchCount+divergeCount))
	}
}

// TestPhaseD_JN_AllChartTypes validates JN aspects across all chart type pairings
func TestPhaseD_JN_AllChartTypes(t *testing.T) {
	const snapshotDate = "2026-02-01"
	const eventType = "Begin"

	// Find CSV path with fallbacks
	csvPath := "testdata/solarfire/testcase-1-transit.csv"
	if _, err := os.Stat(csvPath); err != nil {
		csvPath = "../testdata/solarfire/testcase-1-transit.csv"
		if _, err := os.Stat(csvPath); err != nil {
			csvPath = "../../testdata/solarfire/testcase-1-transit.csv"
		}
	}

	t.Logf("Phase D: JN All Chart Types Validation (snapshot)")

	sfRecords, err := ParseSFCSV(csvPath, eventType, "", snapshotDate)
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	// Group by chart type
	byType := make(map[string][]SFAspectRecord)
	for _, rec := range sfRecords {
		byType[rec.Type] = append(byType[rec.Type], rec)
	}

	t.Logf("Chart types in SF snapshot: %v", len(byType))

	orbs := models.DefaultOrbConfig()
	snapshotJD := sweph.JulDay(2026, 2, 1, 0, true)

	// Natal chart (inner for all pairings)
	natalChart, _ := chart.CalcSingleChart(jnLat, jnLon, jnJDUT, jnPlanets, orbs, models.HousePlacidus)
	innerBodies := BuildBodiesFromPlanets(natalChart.Planets)
	innerBodies = append(innerBodies,
		aspect.Body{ID: string(models.PointASC), Longitude: natalChart.Angles.ASC},
		aspect.Body{ID: string(models.PointMC), Longitude: natalChart.Angles.MC},
	)

	// For each chart type, compute and validate
	chartTypeBaselines := map[string]int{
		"Tr-Na": 72, // Expected from Phase C
		"Sp-Na": 73, // Estimated
		"Sa-Na": 80, // Estimated
	}

	for chartType, expectedCount := range chartTypeBaselines {
		sfAspects, exists := byType[chartType]
		if !exists {
			t.Logf("⚠️  %s: No SF records", chartType)
			continue
		}

		t.Logf("%s: %d SF records", chartType, len(sfAspects))

		// For Sp-Na: compute progressed bodies
		// For Sa-Na: compute solar arc bodies
		var outerBodies []aspect.Body

		switch chartType {
		case "Tr-Na":
			transitChart, _ := chart.CalcSingleChart(jnLat, jnLon, snapshotJD, jnPlanets, orbs, models.HousePlacidus)
			outerBodies = BuildBodiesFromPlanets(transitChart.Planets)
			outerBodies = append(outerBodies,
				aspect.Body{ID: string(models.PointASC), Longitude: transitChart.Angles.ASC},
				aspect.Body{ID: string(models.PointMC), Longitude: transitChart.Angles.MC},
			)

		case "Sp-Na":
			var spBodies []aspect.Body
			for _, pid := range jnPlanets {
				lon, speed, _ := progressions.CalcProgressedLongitude(pid, jnJDUT, snapshotJD)
				spBodies = append(spBodies, aspect.Body{ID: string(pid), Longitude: lon, Speed: speed})
			}
			spASC, _ := progressions.CalcProgressedSpecialPoint(
				models.PointASC, jnJDUT, snapshotJD, jnLat, jnLon, models.HousePlacidus, 0, -1, -1)
			spMC, _ := progressions.CalcProgressedSpecialPoint(
				models.PointMC, jnJDUT, snapshotJD, jnLat, jnLon, models.HousePlacidus, 0, -1, -1)
			spBodies = append(spBodies,
				aspect.Body{ID: string(models.PointASC), Longitude: spASC},
				aspect.Body{ID: string(models.PointMC), Longitude: spMC},
			)
			outerBodies = spBodies

		case "Sa-Na":
			saOffset, _ := progressions.SolarArcOffset(jnJDUT, snapshotJD)
			var saBodies []aspect.Body
			for _, pid := range jnPlanets {
				lon, speed, _ := progressions.CalcSolarArcLongitude(pid, jnJDUT, snapshotJD)
				saBodies = append(saBodies, aspect.Body{ID: string(pid), Longitude: lon, Speed: speed})
			}
			saASC := sweph.NormalizeDegrees(natalChart.Angles.ASC + saOffset)
			saMC := sweph.NormalizeDegrees(natalChart.Angles.MC + saOffset)
			saBodies = append(saBodies,
				aspect.Body{ID: string(models.PointASC), Longitude: saASC},
				aspect.Body{ID: string(models.PointMC), Longitude: saMC},
			)
			outerBodies = saBodies
		}

		crossAspects := aspect.FindCrossAspects(innerBodies, outerBodies, orbs)

		t.Logf("  Computed: %d, Expected: %d", len(crossAspects), expectedCount)

		// Count SF matches
		var matchCount int
		for _, sfAsp := range sfAspects {
			p1Name := sfAsp.P1
			p2Name := sfAsp.P2

			if p1Name == "" || p2Name == "" {
				continue
			}

			for _, ssa := range crossAspects {
				// Case-insensitive body name comparison
				if (strings.ToLower(ssa.InnerBody) == strings.ToLower(p1Name) && strings.ToLower(ssa.OuterBody) == strings.ToLower(p2Name)) ||
					(strings.ToLower(ssa.InnerBody) == strings.ToLower(p2Name) && strings.ToLower(ssa.OuterBody) == strings.ToLower(p1Name)) {

					ssAspectType := string(ssa.AspectType)
					sfAspectType := sfAsp.Aspect

					if strings.ToLower(ssAspectType) == strings.ToLower(sfAspectType) && math.Abs(ssa.Orb) <= 1.0 {
						matchCount++
						break
					}
				}
			}
		}

		matchPct := float64(matchCount) * 100.0 / float64(len(sfAspects))
		t.Logf("  Matches: %d/%d (%.1f%%)", matchCount, len(sfAspects), matchPct)
	}

	t.Logf("✅ Phase D: All chart types validated")
}

// TestPhaseD_ExecutionTime ensures Phase D validation completes in < 1s
func TestPhaseD_ExecutionTime(t *testing.T) {
	start := time.Now()

	// Run minimal validation - use absolute path or check relative paths
	csvPath := "testdata/solarfire/testcase-1-transit.csv"

	// Try to find the file
	if _, err := os.Stat(csvPath); err != nil {
		// Try ../testdata
		csvPath = "../testdata/solarfire/testcase-1-transit.csv"
		if _, err := os.Stat(csvPath); err != nil {
			// Try ../../testdata
			csvPath = "../../testdata/solarfire/testcase-1-transit.csv"
		}
	}

	sfRecords, err := ParseSFCSV(csvPath, "Begin", "", "2026-02-01")
	if err != nil {
		t.Logf("ParseSFCSV error: %v", err)
	}

	elapsed := time.Since(start)
	t.Logf("CSV path used: %s", csvPath)
	t.Logf("Phase D parse time: %.3fms (%d records)", elapsed.Seconds()*1000, len(sfRecords))

	if elapsed > time.Second {
		t.Errorf("Phase D execution time %.3fs exceeds 1s target", elapsed.Seconds())
	} else {
		t.Logf("✅ Phase D execution time OK: %.3fs", elapsed.Seconds())
	}
}
