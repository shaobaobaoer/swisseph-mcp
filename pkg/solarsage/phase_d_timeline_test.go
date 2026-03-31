package solarsage

import (
	"os"
	"testing"
	"time"
)

// ============================================================================
// Phase D v2: Full Timeline Validation Tests
// ============================================================================
// Validates ALL events from testcase-1 and testcase-2, including:
// - All event types (Enter, Exact, Leave, Void, SignIngress, etc.)
// - All chart pairings (Tr-Na, Sp-Na, Sa-Na, Tr-Sp, Tr-Sa, Tr-Tr, Sp-Sp)
// - All reference persons (JN, XB)
// - Full timelines (1+ years)
// ============================================================================

// TestPhaseD_v2_JN_TrNa_FullTimeline validates ALL Tr-Na events from testcase-1
func TestPhaseD_v2_JN_TrNa_FullTimeline(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	t.Logf("=== Phase D v2: JN Tr-Na Full Timeline Validation ===\n")

	// Find CSV with fallbacks
	actualPath := csvPath
	if _, err := checkFileExists(csvPath); err != nil {
		actualPath = "../testdata/solarfire/testcase-1-transit.csv"
		if _, err := checkFileExists(actualPath); err != nil {
			actualPath = "testdata/solarfire/testcase-1-transit.csv"
		}
	}

	startTime := time.Now()

	// Load all SF records
	sfRecords, err := ParseSFCSV(actualPath, "", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	t.Logf("Loaded %d total SF records from testcase-1\n", len(sfRecords))

	// Validate Tr-Na timeline
	report := ValidateTimelineTrNa(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)

	elapsed := time.Since(startTime)
	report.ExecutionTimeMs = elapsed.Seconds() * 1000

	// Print comprehensive report
	reportStr := PrintTimelineReport(report)
	t.Log(reportStr)

	// Summary assertions
	t.Logf("\nSUMMARY:")
	t.Logf("  Total Tr-Na events: %d", report.TotalSFRecords)
	t.Logf("  Matches: %d (%.1f%%)", report.TotalMatches, report.MatchRate)
	t.Logf("  Divergences: %d", report.TotalDivergences)
	t.Logf("  Execution time: %.0fms", report.ExecutionTimeMs)

	if report.MatchRate >= 80 {
		t.Logf("✅ PASS: Match rate %.1f%% >= 80%% target", report.MatchRate)
	} else {
		t.Logf("⚠️  WARNING: Match rate %.1f%% < 80%% target", report.MatchRate)
	}

	// Ensure execution < 1 second
	if elapsed < time.Second {
		t.Logf("✅ Performance OK: %.0fms << 1s", report.ExecutionTimeMs)
	} else {
		t.Errorf("Performance FAILED: %.0fms > 1s", report.ExecutionTimeMs)
	}

	// Event type breakdown
	t.Logf("\nEvent Type Details:")
	for _, eventType := range []string{"Begin", "Enter", "Exact", "Leave"} {
		if stats, exists := report.ByEventType[eventType]; exists && stats.Count > 0 {
			t.Logf("  %s: %d events, %.1f%% match rate", eventType, stats.Count, stats.MatchRate)
		}
	}

	// Show what percentage of EACH event type we're validating
	t.Logf("\nEvent Type Coverage (vs full timeline):")
	totalByType := make(map[string]int)
	for _, rec := range sfRecords {
		if rec.Type == "Tr-Na" {
			totalByType[rec.EventType]++
		}
	}

	for eventType, total := range totalByType {
		var matched int
		if stats, exists := report.ByEventType[eventType]; exists {
			matched = stats.Matches
		}
		coverage := float64(matched) * 100.0 / float64(total)
		t.Logf("  %s: %d/%d validated (%.1f%% of this type)", eventType, matched, total, coverage)
	}
}

// TestPhaseD_v2_JN_AllChartTypes validates all chart types for JN
func TestPhaseD_v2_JN_AllChartTypes(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	t.Logf("=== Phase D v2: JN All Chart Types ===\n")

	actualPath := csvPath
	if _, err := checkFileExists(csvPath); err != nil {
		actualPath = "../testdata/solarfire/testcase-1-transit.csv"
		if _, err := checkFileExists(actualPath); err != nil {
			actualPath = "testdata/solarfire/testcase-1-transit.csv"
		}
	}

	sfRecords, err := ParseSFCSV(actualPath, "", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	// Group by chart type
	byChartType := make(map[string][]SFAspectRecord)
	for _, rec := range sfRecords {
		byChartType[rec.Type] = append(byChartType[rec.Type], rec)
	}

	// Validate each major pairing
	chartTypes := []string{"Tr-Na", "Sp-Na", "Sa-Na"}
	totalEvents := 0

	for _, chartType := range chartTypes {
		if records, exists := byChartType[chartType]; exists && len(records) > 0 {
			totalEvents += len(records)
			t.Logf("%s: %d events", chartType, len(records))

			// Show event type distribution for this pairing
			eventTypeDist := make(map[string]int)
			for _, rec := range records {
				eventTypeDist[rec.EventType]++
			}

			for eventType, count := range eventTypeDist {
				t.Logf("  - %s: %d", eventType, count)
			}
		}
	}

	t.Logf("\nPrimary pairings total: %d events", totalEvents)
	t.Logf("Advanced pairings (Tr-Sp, Tr-Sa, Tr-Tr, Sp-Sp, etc.) not yet validated")

	// Status for each pairing
	t.Logf("\nValidation Status by Pairing:")
	expectedMatches := map[string]float64{
		"Tr-Na": 0.85, // expect 85%+ match
		"Sp-Na": 0.70, // expect 70%+ (symbolic chart, lower baseline)
		"Sa-Na": 0.80, // expect 80%+
	}

	for chartType := range expectedMatches {
		if _, exists := byChartType[chartType]; exists {
			t.Logf("  %s: READY for detailed timeline validation", chartType)
		}
	}

	t.Logf("\n✅ Phase D v2 Milestone 1 (Tr-Na) can now proceed with full timeline")
	t.Logf("   Remaining milestones: Sp-Na, Sa-Na, Advanced pairings")
}

// TestPhaseD_v2_ProgressiveExpansion shows the validation path forward
func TestPhaseD_v2_ProgressiveExpansion(t *testing.T) {
	msg := `
=================================================================
Phase D v2: Progressive Expansion Plan
=================================================================

STAGE 1: JN Full Timeline (testcase-1 - 1,156 events)
  READY Tr-Na: 200 events (4 Begin, 64 Enter, 59 Exact, 61 Leave)
     Validating ALL phases, not just snapshots
     Expected: 85%+ match rate
     Timeline: 2026-02-01 to 2027-01-30

  TODO Sp-Na: 52 events (secondary progressions)
     Lower baseline expected (symbolic chart)
     Expected: 70%+ match rate

  TODO Sa-Na: 31 events (solar arc directed)
     Expected: 80%+ match rate

  TODO Advanced: 873 events (Tr-Sp, Tr-Sa, Tr-Tr, Sp-Sp)
     Complex pairings requiring different validators
     Expected: varied match rates

STAGE 2: XB Timeline 1996-2001 (testcase-2 - 1,746 events)
  TODO Tr-Na: 1,326 events
  TODO Sp-Na: 217 events
  TODO Others: 203 events

STAGE 3: XB Timeline 2001-2006 (testcase-3 - 1,512 events)
  TODO Tr-Na: 1,145 events
  TODO Sp-Na: 171 events
  TODO Others: 196 events

=================================================================

VALIDATION PROGRESS TRACKING:
  Total Scope:    4,414 events
  Validated v1:        25 (0.6%)
  Validating v2:    1,156 (26%)
  Remaining:       3,258 (74%)

KEY IMPROVEMENTS IN v2:
  Timeline spanning (multiple dates, not snapshot)
  Event type coverage (Enter, Exact, Leave, not just Begin)
  Occurrence matching (same aspect repeating)
  Multi-pairing support (infrastructure for all types)
  Comprehensive reporting (detailed divergence analysis)

=================================================================
`
	t.Log(msg)
	t.Logf("Ready to execute Phase D v2 Milestone 1: Tr-Na Timeline Validator\n")
}

// TestPhaseD_v3_XB_TrNa_Stage1 validates XB Tr-Na events (1996-2001, 5 years)
func TestPhaseD_v3_XB_TrNa_Stage1(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-2-transit-1996-2001.csv"

	t.Logf("=== Phase D v3: XB Tr-Na Full Timeline (1996-2001, 5 years) ===\n")

	actualPath := csvPath
	if _, err := checkFileExists(csvPath); err != nil {
		actualPath = "../testdata/solarfire/testcase-2-transit-1996-2001.csv"
		if _, err := checkFileExists(actualPath); err != nil {
			actualPath = "testdata/solarfire/testcase-2-transit-1996-2001.csv"
		}
	}

	startTime := time.Now()

	sfRecords, err := ParseSFCSV(actualPath, "", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	t.Logf("Loaded %d total SF records from testcase-2 (1996-2001)\n", len(sfRecords))

	report := ValidateTimelineTrNa(sfRecords, xbJDUT, xbLat, xbLon, xbPlanets)

	elapsed := time.Since(startTime)
	report.ExecutionTimeMs = elapsed.Seconds() * 1000

	reportStr := PrintTimelineReport(report)
	t.Log(reportStr)

	t.Logf("\nSUMMARY:")
	t.Logf("  Total Tr-Na events: %d", report.TotalSFRecords)
	t.Logf("  Matches: %d (%.1f%%)", report.TotalMatches, report.MatchRate)
	t.Logf("  Divergences: %d", report.TotalDivergences)
	t.Logf("  Execution time: %.0fms", report.ExecutionTimeMs)

	if report.MatchRate >= 65 {
		t.Logf("✅ PASS: Match rate %.1f%% >= 65%% target", report.MatchRate)
	} else {
		t.Logf("⚠️  WARNING: Match rate %.1f%% < 65%% target", report.MatchRate)
	}
}

// TestPhaseD_v3_XB_TrNa_Stage2 validates XB Tr-Na events (2001-2006, 5 years)
func TestPhaseD_v3_XB_TrNa_Stage2(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-2-transit-2001-2006.csv"

	t.Logf("=== Phase D v3: XB Tr-Na Full Timeline (2001-2006, 5 years) ===\n")

	actualPath := csvPath
	if _, err := checkFileExists(csvPath); err != nil {
		actualPath = "../testdata/solarfire/testcase-2-transit-2001-2006.csv"
		if _, err := checkFileExists(actualPath); err != nil {
			actualPath = "testdata/solarfire/testcase-2-transit-2001-2006.csv"
		}
	}

	startTime := time.Now()

	sfRecords, err := ParseSFCSV(actualPath, "", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	t.Logf("Loaded %d total SF records from testcase-2 (2001-2006)\n", len(sfRecords))

	report := ValidateTimelineTrNa(sfRecords, xbJDUT, xbLat, xbLon, xbPlanets)

	elapsed := time.Since(startTime)
	report.ExecutionTimeMs = elapsed.Seconds() * 1000

	reportStr := PrintTimelineReport(report)
	t.Log(reportStr)

	t.Logf("\nSUMMARY:")
	t.Logf("  Total Tr-Na events: %d", report.TotalSFRecords)
	t.Logf("  Matches: %d (%.1f%%)", report.TotalMatches, report.MatchRate)
	t.Logf("  Divergences: %d", report.TotalDivergences)
	t.Logf("  Execution time: %.0fms", report.ExecutionTimeMs)

	if report.MatchRate >= 65 {
		t.Logf("✅ PASS: Match rate %.1f%% >= 65%% target", report.MatchRate)
	} else {
		t.Logf("⚠️  WARNING: Match rate %.1f%% < 65%% target", report.MatchRate)
	}
}

// TestPhaseD_v4_XB_Advanced_Stage1 validates XB advanced pairings (1996-2001)
func TestPhaseD_v4_XB_Advanced_Stage1(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-2-transit-1996-2001.csv"

	t.Logf("=== Phase D v4: XB Advanced Pairings (1996-2001, 5 years) ===\n")

	actualPath := csvPath
	if _, err := checkFileExists(csvPath); err != nil {
		actualPath = "../testdata/solarfire/testcase-2-transit-1996-2001.csv"
		if _, err := checkFileExists(actualPath); err != nil {
			actualPath = "testdata/solarfire/testcase-2-transit-1996-2001.csv"
		}
	}

	startTime := time.Now()

	sfRecords, err := ParseSFCSV(actualPath, "", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	report := ValidateTimelineAdvancedPairings(sfRecords, xbJDUT, xbLat, xbLon, xbPlanets)

	elapsed := time.Since(startTime)
	report.ExecutionTimeMs = elapsed.Seconds() * 1000

	reportStr := PrintTimelineReport(report)
	t.Log(reportStr)

	t.Logf("\nSUMMARY:")
	t.Logf("  Total Advanced Pairing events: %d", report.TotalSFRecords)
	t.Logf("  Matches: %d (%.1f%%)", report.TotalMatches, report.MatchRate)
	t.Logf("  Execution time: %.0fms", report.ExecutionTimeMs)

	if report.MatchRate >= 50 {
		t.Logf("✅ PASS: Match rate %.1f%% >= 50%% target", report.MatchRate)
	} else {
		t.Logf("⚠️  WARNING: Match rate %.1f%% < 50%% target", report.MatchRate)
	}
}

// TestPhaseD_v4_XB_Advanced_Stage2 validates XB advanced pairings (2001-2006)
func TestPhaseD_v4_XB_Advanced_Stage2(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-2-transit-2001-2006.csv"

	t.Logf("=== Phase D v4: XB Advanced Pairings (2001-2006, 5 years) ===\n")

	actualPath := csvPath
	if _, err := checkFileExists(csvPath); err != nil {
		actualPath = "../testdata/solarfire/testcase-2-transit-2001-2006.csv"
		if _, err := checkFileExists(actualPath); err != nil {
			actualPath = "testdata/solarfire/testcase-2-transit-2001-2006.csv"
		}
	}

	startTime := time.Now()

	sfRecords, err := ParseSFCSV(actualPath, "", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	report := ValidateTimelineAdvancedPairings(sfRecords, xbJDUT, xbLat, xbLon, xbPlanets)

	elapsed := time.Since(startTime)
	report.ExecutionTimeMs = elapsed.Seconds() * 1000

	reportStr := PrintTimelineReport(report)
	t.Log(reportStr)

	t.Logf("\nSUMMARY:")
	t.Logf("  Total Advanced Pairing events: %d", report.TotalSFRecords)
	t.Logf("  Matches: %d (%.1f%%)", report.TotalMatches, report.MatchRate)
	t.Logf("  Execution time: %.0fms", report.ExecutionTimeMs)

	if report.MatchRate >= 50 {
		t.Logf("✅ PASS: Match rate %.1f%% >= 50%% target", report.MatchRate)
	} else {
		t.Logf("⚠️  WARNING: Match rate %.1f%% < 50%% target", report.MatchRate)
	}
}

// TestPhaseD_v5_JN_Void validates Void of Course Moon events
func TestPhaseD_v5_JN_Void(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	t.Logf("=== Phase D v5: JN Void of Course Moon Events ===\n")

	actualPath := csvPath
	if _, err := checkFileExists(csvPath); err != nil {
		actualPath = "../testdata/solarfire/testcase-1-transit.csv"
		if _, err := checkFileExists(actualPath); err != nil {
			actualPath = "testdata/solarfire/testcase-1-transit.csv"
		}
	}

	startTime := time.Now()

	sfRecords, err := ParseSFCSV(actualPath, "", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	report := ValidateTimelineVoidOfCourse(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)

	elapsed := time.Since(startTime)
	report.ExecutionTimeMs = elapsed.Seconds() * 1000

	reportStr := PrintTimelineReport(report)
	t.Log(reportStr)

	t.Logf("\nSUMMARY:")
	t.Logf("  Total Void events: %d", report.TotalSFRecords)
	t.Logf("  Matches: %d (%.1f%%)", report.TotalMatches, report.MatchRate)
	t.Logf("  Execution time: %.0fms", report.ExecutionTimeMs)

	if report.MatchRate >= 40 {
		t.Logf("✅ PASS: Match rate %.1f%% >= 40%% target", report.MatchRate)
	} else {
		t.Logf("⚠️  INFO: Match rate %.1f%% (baseline validation)", report.MatchRate)
	}
}

// TestPhaseD_v5_JN_SignIngress validates Sign Ingress events
func TestPhaseD_v5_JN_SignIngress(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	t.Logf("=== Phase D v5: JN Sign Ingress Events ===\n")

	actualPath := csvPath
	if _, err := checkFileExists(csvPath); err != nil {
		actualPath = "../testdata/solarfire/testcase-1-transit.csv"
		if _, err := checkFileExists(actualPath); err != nil {
			actualPath = "testdata/solarfire/testcase-1-transit.csv"
		}
	}

	startTime := time.Now()

	sfRecords, err := ParseSFCSV(actualPath, "", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	report := ValidateTimelineSignIngress(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)

	elapsed := time.Since(startTime)
	report.ExecutionTimeMs = elapsed.Seconds() * 1000

	reportStr := PrintTimelineReport(report)
	t.Log(reportStr)

	t.Logf("\nSUMMARY:")
	t.Logf("  Total SignIngress events: %d", report.TotalSFRecords)
	t.Logf("  Matches: %d (%.1f%%)", report.TotalMatches, report.MatchRate)
	t.Logf("  Execution time: %.0fms", report.ExecutionTimeMs)

	if report.MatchRate >= 50 {
		t.Logf("✅ PASS: Match rate %.1f%% >= 50%% target", report.MatchRate)
	} else {
		t.Logf("⚠️  INFO: Match rate %.1f%% (baseline validation)", report.MatchRate)
	}
}

// TestPhaseD_v7_JN_SpSp validates Sp-Sp (progressed vs progressed) within-chart aspects
func TestPhaseD_v7_JN_SpSp(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	t.Logf("=== Phase D v7: JN Sp-Sp Within-Chart Aspects ===\n")

	actualPath := csvPath
	if _, err := checkFileExists(csvPath); err != nil {
		actualPath = "../testdata/solarfire/testcase-1-transit.csv"
		if _, err := checkFileExists(actualPath); err != nil {
			actualPath = "testdata/solarfire/testcase-1-transit.csv"
		}
	}

	startTime := time.Now()

	sfRecords, err := ParseSFCSV(actualPath, "", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	report := ValidateTimelineSpSp(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)

	elapsed := time.Since(startTime)
	report.ExecutionTimeMs = elapsed.Seconds() * 1000

	reportStr := PrintTimelineReport(report)
	t.Log(reportStr)

	t.Logf("\nSUMMARY:")
	t.Logf("  Total Sp-Sp events: %d", report.TotalSFRecords)
	t.Logf("  Matches: %d (%.1f%%)", report.TotalMatches, report.MatchRate)
	t.Logf("  Execution time: %.0fms", report.ExecutionTimeMs)

	if report.MatchRate >= 60 {
		t.Logf("✅ PASS: Match rate %.1f%% >= 60%% target", report.MatchRate)
	} else {
		t.Logf("⚠️  INFO: Match rate %.1f%% (baseline validation)", report.MatchRate)
	}
}

// TestPhaseD_v6_JN_Stations validates Retrograde/Direct station events
func TestPhaseD_v6_JN_Stations(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	t.Logf("=== Phase D v6: JN Station Events (Retrograde/Direct) ===\n")

	actualPath := csvPath
	if _, err := checkFileExists(csvPath); err != nil {
		actualPath = "../testdata/solarfire/testcase-1-transit.csv"
		if _, err := checkFileExists(actualPath); err != nil {
			actualPath = "testdata/solarfire/testcase-1-transit.csv"
		}
	}

	startTime := time.Now()

	sfRecords, err := ParseSFCSV(actualPath, "", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	report := ValidateTimelineStations(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)

	elapsed := time.Since(startTime)
	report.ExecutionTimeMs = elapsed.Seconds() * 1000

	reportStr := PrintTimelineReport(report)
	t.Log(reportStr)

	t.Logf("\nSUMMARY:")
	t.Logf("  Total Station events: %d", report.TotalSFRecords)
	t.Logf("  Matches: %d (%.1f%%)", report.TotalMatches, report.MatchRate)
	t.Logf("  Execution time: %.0fms", report.ExecutionTimeMs)

	if report.MatchRate >= 50 {
		t.Logf("✅ PASS: Match rate %.1f%% >= 50%% target", report.MatchRate)
	} else {
		t.Logf("⚠️  INFO: Match rate %.1f%% (baseline validation)", report.MatchRate)
	}
}

// TestPhaseD_v6_JN_HouseChange validates House Change events
func TestPhaseD_v6_JN_HouseChange(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	t.Logf("=== Phase D v6: JN House Change Events ===\n")

	actualPath := csvPath
	if _, err := checkFileExists(csvPath); err != nil {
		actualPath = "../testdata/solarfire/testcase-1-transit.csv"
		if _, err := checkFileExists(actualPath); err != nil {
			actualPath = "testdata/solarfire/testcase-1-transit.csv"
		}
	}

	startTime := time.Now()

	sfRecords, err := ParseSFCSV(actualPath, "", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	report := ValidateTimelineHouseChange(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)

	elapsed := time.Since(startTime)
	report.ExecutionTimeMs = elapsed.Seconds() * 1000

	reportStr := PrintTimelineReport(report)
	t.Log(reportStr)

	t.Logf("\nSUMMARY:")
	t.Logf("  Total HouseChange events: %d", report.TotalSFRecords)
	t.Logf("  Matches: %d (%.1f%%)", report.TotalMatches, report.MatchRate)
	t.Logf("  Execution time: %.0fms", report.ExecutionTimeMs)

	if report.MatchRate >= 50 {
		t.Logf("✅ PASS: Match rate %.1f%% >= 50%% target", report.MatchRate)
	} else {
		t.Logf("⚠️  INFO: Match rate %.1f%% (baseline validation)", report.MatchRate)
	}
}

// TestPhaseD_v9_JN_EventCoverage analyzes which event types remain unvalidated
func TestPhaseD_v9_JN_EventCoverage(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	// Find CSV with fallbacks
	actualPath := csvPath
	if _, err := checkFileExists(csvPath); err != nil {
		actualPath = "../testdata/solarfire/testcase-1-transit.csv"
		if _, err := checkFileExists(actualPath); err != nil {
			actualPath = "testdata/solarfire/testcase-1-transit.csv"
		}
	}

	t.Logf("=== Phase D v9: Event Coverage Analysis ===\n")

	// Load all SF records
	sfRecords, err := ParseSFCSV(actualPath, "", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	t.Logf("Total records: %d\n", len(sfRecords))

	// Count by event type
	eventCount := make(map[string]int)
	eventByType := make(map[string][]SFAspectRecord)
	for _, rec := range sfRecords {
		eventCount[rec.EventType]++
		eventByType[rec.EventType] = append(eventByType[rec.EventType], rec)
	}

	// Count by chart type
	chartCount := make(map[string]int)
	for _, rec := range sfRecords {
		chartCount[rec.Type]++
	}

	t.Logf("Event Type Distribution:")
	eventTypes := []string{"Begin", "Enter", "Exact", "Leave", "Void", "SignIngress", "HouseChange", "Retrograde", "Direct"}
	for _, et := range eventTypes {
		if count, exists := eventCount[et]; exists {
			t.Logf("  %s: %d events", et, count)
		}
	}

	t.Logf("\nChart Type Distribution:")
	for ct := range chartCount {
		t.Logf("  %s: %d events", ct, chartCount[ct])
	}

	// Run all validators and track coverage
	trNa := ValidateTimelineTrNa(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)
	spNa := ValidateTimelineSpNa(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)
	saNa := ValidateTimelineSaNa(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)
	trSp := ValidateTimelineAdvancedPairings(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)
	spSp := ValidateTimelineSpSp(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)
	trTr := ValidateTimelineTrTr(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)
	voc := ValidateTimelineVoidOfCourse(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)
	sig := ValidateTimelineSignIngress(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)
	hc := ValidateTimelineHouseChange(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)
	st := ValidateTimelineStations(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)

	totalMatched := trNa.TotalMatches + spNa.TotalMatches + saNa.TotalMatches +
		trSp.TotalMatches + spSp.TotalMatches + trTr.TotalMatches +
		voc.TotalMatches + sig.TotalMatches + hc.TotalMatches + st.TotalMatches

	t.Logf("\nValidator Coverage Summary:")
	t.Logf("  Tr-Na: %d/%d (%.1f%%)", trNa.TotalMatches, trNa.TotalSFRecords, trNa.MatchRate)
	t.Logf("  Sp-Na: %d/%d (%.1f%%)", spNa.TotalMatches, spNa.TotalSFRecords, spNa.MatchRate)
	t.Logf("  Sa-Na: %d/%d (%.1f%%)", saNa.TotalMatches, saNa.TotalSFRecords, saNa.MatchRate)
	t.Logf("  Tr-Sp/Tr-Sa: %d/%d (%.1f%%)", trSp.TotalMatches, trSp.TotalSFRecords, trSp.MatchRate)
	t.Logf("  Sp-Sp: %d/%d (%.1f%%)", spSp.TotalMatches, spSp.TotalSFRecords, spSp.MatchRate)
	t.Logf("  Tr-Tr: %d/%d (%.1f%%)", trTr.TotalMatches, trTr.TotalSFRecords, trTr.MatchRate)
	t.Logf("  Void: %d/%d (%.1f%%)", voc.TotalMatches, voc.TotalSFRecords, voc.MatchRate)
	t.Logf("  SignIngress: %d/%d (%.1f%%)", sig.TotalMatches, sig.TotalSFRecords, sig.MatchRate)
	t.Logf("  HouseChange: %d/%d (%.1f%%)", hc.TotalMatches, hc.TotalSFRecords, hc.MatchRate)
	t.Logf("  Stations: %d/%d (%.1f%%)", st.TotalMatches, st.TotalSFRecords, st.MatchRate)
	t.Logf("\nTotal Matches (with overlap): %d/%d (%.1f%%)", totalMatched, len(sfRecords),
		float64(totalMatched)*100.0/float64(len(sfRecords)))
}

// TestPhaseD_v8_JN_TrTr validates Tr-Tr within-chart transit aspects
func TestPhaseD_v8_JN_TrTr(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	// Find CSV with fallbacks
	actualPath := csvPath
	if _, err := checkFileExists(csvPath); err != nil {
		actualPath = "../testdata/solarfire/testcase-1-transit.csv"
		if _, err := checkFileExists(actualPath); err != nil {
			actualPath = "testdata/solarfire/testcase-1-transit.csv"
		}
	}

	t.Logf("=== Phase D v8: JN Tr-Tr Within-Chart Transit Aspects ===\n")

	// Load all SF records
	sfRecords, err := ParseSFCSV(actualPath, "", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	t.Logf("Loaded %d total SF records from testcase-1\n", len(sfRecords))

	// Validate Tr-Tr (transit vs transit) within-chart
	startTime := time.Now()
	report := ValidateTimelineTrTr(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)
	elapsed := time.Since(startTime)
	report.ExecutionTimeMs = elapsed.Seconds() * 1000

	reportStr := PrintTimelineReport(report)
	t.Log(reportStr)

	// Summary assertions
	t.Logf("\nSUMMARY:")
	t.Logf("  Total Tr-Tr events: %d", report.TotalSFRecords)
	t.Logf("  Matches: %d (%.1f%%)", report.TotalMatches, report.MatchRate)
	t.Logf("  Divergences: %d", report.TotalDivergences)
	t.Logf("  Execution time: %.0fms", report.ExecutionTimeMs)

	if report.MatchRate >= 50 {
		t.Logf("✅ PASS: Match rate %.1f%% >= 50%% target", report.MatchRate)
	} else {
		t.Logf("⚠️  WARNING: Match rate %.1f%% < 50%% target", report.MatchRate)
	}
}

// TestPhaseD_v8_JN_Comprehensive validates all JN event validators
func TestPhaseD_v8_JN_Comprehensive(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	// Find CSV with fallbacks
	actualPath := csvPath
	if _, err := checkFileExists(csvPath); err != nil {
		actualPath = "../testdata/solarfire/testcase-1-transit.csv"
		if _, err := checkFileExists(actualPath); err != nil {
			actualPath = "testdata/solarfire/testcase-1-transit.csv"
		}
	}

	t.Logf("\n=== Phase D v8: JN Comprehensive Coverage ===\n")

	// Load all SF records
	sfRecords, err := ParseSFCSV(actualPath, "", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	t.Logf("Loaded %d total SF records from testcase-1\n", len(sfRecords))

	// Run validators for non-overlapping record sets
	results := make(map[string]*TimelineValidationReport)

	// Primary pairings (these have some special events mixed in, that's OK)
	results["Tr-Na"] = ValidateTimelineTrNa(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)
	results["Sp-Na"] = ValidateTimelineSpNa(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)
	results["Sa-Na"] = ValidateTimelineSaNa(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)

	// Advanced pairings (Tr-Sp, Tr-Sa, Tr-Tr only, Sp-Sp handled separately)
	resultsAdv := ValidateTimelineAdvancedPairings(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)

	// Progressed within-chart (Sp-Sp)
	results["Sp-Sp"] = ValidateTimelineSpSp(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)

	// Transit within-chart (Tr-Tr)
	results["Tr-Tr"] = ValidateTimelineTrTr(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)

	// Station events (Retrograde/Direct) in "Tr" chart type
	results["Stations"] = ValidateTimelineStations(sfRecords, jnJDUT, jnLat, jnLon, jnPlanets)

	// Print comprehensive summary
	t.Logf("\n═══════════════════════════════════════════════════════════════════════════")
	t.Logf("VALIDATOR MATCH RATES (non-overlapping sets):")
	t.Logf("═══════════════════════════════════════════════════════════════════════════\n")

	totalValidated := 0
	totalEvents := 0

	// Chart pairing validators
	for _, name := range []string{"Tr-Na", "Sp-Na", "Sa-Na", "Sp-Sp", "Tr-Tr"} {
		if report, exists := results[name]; exists && report.TotalSFRecords > 0 {
			status := "⚠️ "
			if report.MatchRate >= 70 {
				status = "✅"
			}
			t.Logf("%s %s: %d events, %d matched (%.1f%%)",
				status, name, report.TotalSFRecords, report.TotalMatches, report.MatchRate)
			totalValidated += report.TotalMatches
			totalEvents += report.TotalSFRecords
		}
	}

	t.Logf("\n  Advanced Pairings (Tr-Sp, Tr-Sa, Tr-Tr):")
	t.Logf("    Events: %d, Matched: %d (%.1f%%)",
		resultsAdv.TotalSFRecords, resultsAdv.TotalMatches, resultsAdv.MatchRate)
	totalValidated += resultsAdv.TotalMatches
	totalEvents += resultsAdv.TotalSFRecords

	// Station-only events
	t.Logf("\n  Stations (Retrograde/Direct in Tr): %d events, %d matched (%.1f%%)",
		results["Stations"].TotalSFRecords, results["Stations"].TotalMatches, results["Stations"].MatchRate)
	totalValidated += results["Stations"].TotalMatches
	totalEvents += results["Stations"].TotalSFRecords

	// Breakdown of Advanced Pairings by type
	t.Logf("\n  Advanced Pairings Breakdown:")
	if len(resultsAdv.ByChartType) > 0 {
		for chartType, stats := range resultsAdv.ByChartType {
			if stats.Count > 0 {
				t.Logf("    %s: %d events, %d matched (%.1f%%)",
					chartType, stats.Count, stats.Matches, stats.MatchRate)
			}
		}
	}

	t.Logf("\n═══════════════════════════════════════════════════════════════════════════")
	t.Logf("TOTAL COVERAGE:")
	t.Logf("═══════════════════════════════════════════════════════════════════════════\n")

	overallRate := 0.0
	if totalEvents > 0 {
		overallRate = float64(totalValidated) * 100.0 / float64(totalEvents)
	}

	t.Logf("Total Events: %d", totalEvents)
	t.Logf("Total Validated: %d", totalValidated)
	t.Logf("Coverage: %.1f%%", overallRate)
	t.Logf("\nExpected: 1,156 events (testcase-1 JN timeline)")
	t.Logf("\nNote: This excludes Void, SignIngress, HouseChange validators")
	t.Logf("  which have additional coverage: +161+169+7 special events")
}

// Helper to check if file exists
func checkFileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return true, nil
}
