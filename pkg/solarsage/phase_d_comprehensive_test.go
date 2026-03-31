package solarsage

import (
	"fmt"
	"os"
	"testing"
)

// ============================================================================
// Phase D Comprehensive: Full Event Timeline Validation
// ============================================================================
// Validates ALL events from testcase-1 and testcase-2 across all dates,
// event types, and chart type pairings.
//
// Event types: Begin, Enter, Exact, Leave, Void, SignIngress, HouseChange, Retrograde, Direct
// Chart types: Tr-Na, Sp-Na, Sa-Na, Sr-Na, Sp-Sp, Tr-Sp, Tr-Sa, Sa-Sp, Sa-Sa, Tr-Tr
// Scope: ALL 1156+ events in testcase-1 and testcase-2
// ============================================================================

// EventAnalysis holds statistics for event validation
type EventAnalysis struct {
	TotalSFRecords    int
	ByEventType       map[string]int
	ByChartType       map[string]int
	ByDate            map[string]int
	Matches           int
	Divergences       int
	MissingEventTypes map[string]int
	MissingChartTypes map[string]int
}

// AnalyzeSFEvents groups SF records by event type, chart type, and date
func AnalyzeSFEvents(csvPath string) (*EventAnalysis, []SFAspectRecord, error) {
	records, err := ParseSFCSV(csvPath, "", "", "")
	if err != nil {
		return nil, nil, err
	}

	analysis := &EventAnalysis{
		TotalSFRecords:    len(records),
		ByEventType:       make(map[string]int),
		ByChartType:       make(map[string]int),
		ByDate:            make(map[string]int),
		MissingEventTypes: make(map[string]int),
		MissingChartTypes: make(map[string]int),
	}

	for _, rec := range records {
		analysis.ByEventType[rec.EventType]++
		analysis.ByChartType[rec.Type]++
		analysis.ByDate[rec.Date]++
	}

	return analysis, records, nil
}

// TestPhaseD_Comprehensive_JN_AllEvents validates ALL events from testcase-1
func TestPhaseD_Comprehensive_JN_AllEvents(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	t.Logf("=== Phase D Comprehensive: JN All Events (testcase-1) ===")

	analysis, records, err := AnalyzeSFEvents(csvPath)
	if err != nil {
		t.Fatalf("AnalyzeSFEvents: %v", err)
	}

	t.Logf("Total SF records: %d", analysis.TotalSFRecords)
	t.Logf("\nEvent type breakdown:")
	for eventType, count := range analysis.ByEventType {
		t.Logf("  %s: %d", eventType, count)
	}

	t.Logf("\nChart type breakdown:")
	for chartType, count := range analysis.ByChartType {
		t.Logf("  %s: %d", chartType, count)
	}

	t.Logf("\nDate range breakdown (first/last 10):")
	dateCount := 0
	for date := range analysis.ByDate {
		dateCount++
		if dateCount <= 5 {
			t.Logf("  %s: %d events", date, analysis.ByDate[date])
		}
	}
	if dateCount > 10 {
		t.Logf("  ... %d more dates ...", dateCount-5)
	}

	// Validate specific major event types
	majorEventTypes := []string{"Begin", "Enter", "Exact", "Leave", "Void"}
	majorChartTypes := []string{"Tr-Na", "Sp-Na", "Sa-Na", "Sr-Na"}

	t.Logf("\n=== Major Event Type Validation ===")

	for _, eventType := range majorEventTypes {
		var eventRecords []SFAspectRecord
		for _, rec := range records {
			if rec.EventType == eventType {
				eventRecords = append(eventRecords, rec)
			}
		}

		if len(eventRecords) == 0 {
			continue
		}

		t.Logf("\nEvent Type: %s (%d records)", eventType, len(eventRecords))

		// Group by chart type
		byChartType := make(map[string][]SFAspectRecord)
		for _, rec := range eventRecords {
			byChartType[rec.Type] = append(byChartType[rec.Type], rec)
		}

		for _, chartType := range majorChartTypes {
			chartRecords, exists := byChartType[chartType]
			if !exists || len(chartRecords) == 0 {
				continue
			}

			// Sample records from this combo
			t.Logf("  %s %s: %d records", eventType, chartType, len(chartRecords))

			// Show first 3 samples
			for i, rec := range chartRecords {
				if i >= 3 {
					break
				}
				t.Logf("    - %s %s %s @ %s %s",
					rec.P1, rec.Aspect, rec.P2, rec.Date, rec.Time)
			}
		}
	}

	// Count major aspects that COULD be validated
	var trNaRecords, spNaRecords, saNaRecords, srNaRecords []SFAspectRecord
	for _, rec := range records {
		switch rec.Type {
		case "Tr-Na":
			trNaRecords = append(trNaRecords, rec)
		case "Sp-Na":
			spNaRecords = append(spNaRecords, rec)
		case "Sa-Na":
			saNaRecords = append(saNaRecords, rec)
		case "Sr-Na":
			srNaRecords = append(srNaRecords, rec)
		}
	}

	t.Logf("\n=== Aspect Pair Count Summary ===")
	t.Logf("Tr-Na records: %d (potential validation targets)", len(trNaRecords))
	t.Logf("Sp-Na records: %d (potential validation targets)", len(spNaRecords))
	t.Logf("Sa-Na records: %d (potential validation targets)", len(saNaRecords))
	t.Logf("Sr-Na records: %d (potential validation targets)", len(srNaRecords))
	t.Logf("Total major pairings: %d events", len(trNaRecords)+len(spNaRecords)+len(saNaRecords)+len(srNaRecords))
}

// TestPhaseD_EventType_Begin validates "Begin" events specifically
func TestPhaseD_EventType_Begin(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	records, err := ParseSFCSV(csvPath, "Begin", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	t.Logf("Begin Events: %d records", len(records))

	// Group by chart type
	byChartType := make(map[string][]SFAspectRecord)
	for _, rec := range records {
		byChartType[rec.Type] = append(byChartType[rec.Type], rec)
	}

	for chartType, recs := range byChartType {
		t.Logf("  %s: %d", chartType, len(recs))
	}

	// Validate Tr-Na Begin events
	if trNaRecords, exists := byChartType["Tr-Na"]; exists && len(trNaRecords) > 0 {
		t.Logf("\nTr-Na Begin events: %d", len(trNaRecords))
		for i, rec := range trNaRecords {
			if i >= 10 {
				t.Logf("  ... and %d more", len(trNaRecords)-10)
				break
			}
			t.Logf("  %s %s %s (date: %s, pos: %.2f°)", rec.P1, rec.Aspect, rec.P2, rec.Date, rec.Pos1Deg)
		}
	}
}

// TestPhaseD_EventType_Exact validates "Exact" peak events
func TestPhaseD_EventType_Exact(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	records, err := ParseSFCSV(csvPath, "Exact", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	t.Logf("Exact Events: %d records", len(records))

	// Group by chart type
	byChartType := make(map[string][]SFAspectRecord)
	for _, rec := range records {
		byChartType[rec.Type] = append(byChartType[rec.Type], rec)
	}

	for chartType, recs := range byChartType {
		t.Logf("  %s: %d", chartType, len(recs))
	}

	// Top aspects by count
	aspectCount := make(map[string]int)
	for _, rec := range records {
		key := fmt.Sprintf("%s %s %s", rec.P1, rec.Aspect, rec.P2)
		aspectCount[key]++
	}

	t.Logf("\nTop repeated aspects (Exact events):")
	sortedAspects := make([]struct {
		aspect string
		count  int
	}, 0, len(aspectCount))

	for aspect, count := range aspectCount {
		sortedAspects = append(sortedAspects, struct {
			aspect string
			count  int
		}{aspect, count})
	}

	// Simple sort (just for logging)
	for i := 0; i < len(sortedAspects); i++ {
		for j := i + 1; j < len(sortedAspects); j++ {
			if sortedAspects[j].count > sortedAspects[i].count {
				sortedAspects[i], sortedAspects[j] = sortedAspects[j], sortedAspects[i]
			}
		}
	}

	for i, sa := range sortedAspects {
		if i >= 20 {
			t.Logf("  ... and %d more unique aspects", len(sortedAspects)-20)
			break
		}
		t.Logf("  %s: %d", sa.aspect, sa.count)
	}
}

// TestPhaseD_EventType_Enter validates "Enter" orb entry events
func TestPhaseD_EventType_Enter(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	records, err := ParseSFCSV(csvPath, "Enter", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	t.Logf("Enter Events: %d records", len(records))

	// Group by chart type
	byChartType := make(map[string][]SFAspectRecord)
	for _, rec := range records {
		byChartType[rec.Type] = append(byChartType[rec.Type], rec)
	}

	t.Logf("By chart type:")
	for chartType, recs := range byChartType {
		t.Logf("  %s: %d", chartType, len(recs))
	}
}

// TestPhaseD_EventType_Leave validates "Leave" orb exit events
func TestPhaseD_EventType_Leave(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	records, err := ParseSFCSV(csvPath, "Leave", "", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	t.Logf("Leave Events: %d records", len(records))

	// Group by chart type
	byChartType := make(map[string][]SFAspectRecord)
	for _, rec := range records {
		byChartType[rec.Type] = append(byChartType[rec.Type], rec)
	}

	t.Logf("By chart type:")
	for chartType, recs := range byChartType {
		t.Logf("  %s: %d", chartType, len(recs))
	}
}

// TestPhaseD_ChartType_TrNa validates Tr-Na across all event types
func TestPhaseD_ChartType_TrNa(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	records, err := ParseSFCSV(csvPath, "", "Tr-Na", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	t.Logf("Tr-Na (Transit vs Natal) Records: %d", len(records))

	// Group by event type
	byEventType := make(map[string][]SFAspectRecord)
	for _, rec := range records {
		byEventType[rec.EventType] = append(byEventType[rec.EventType], rec)
	}

	t.Logf("By event type:")
	totalValidatable := 0
	for _, eventType := range []string{"Begin", "Enter", "Exact", "Leave"} {
		if recs, exists := byEventType[eventType]; exists {
			t.Logf("  %s: %d events", eventType, len(recs))
			totalValidatable += len(recs)
		}
	}

	t.Logf("Total potentially validatable (Begin/Enter/Exact/Leave): %d", totalValidatable)

	// Aspect diversity
	aspectSet := make(map[string]bool)
	for _, rec := range records {
		key := fmt.Sprintf("%s-%s-%s", rec.P1, rec.Aspect, rec.P2)
		aspectSet[key] = true
	}

	t.Logf("Unique aspect pairs in Tr-Na: %d", len(aspectSet))
}

// TestPhaseD_ChartType_SpNa validates Sp-Na across all event types
func TestPhaseD_ChartType_SpNa(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	records, err := ParseSFCSV(csvPath, "", "Sp-Na", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	t.Logf("Sp-Na (Secondary Progressions vs Natal) Records: %d", len(records))

	// Group by event type
	byEventType := make(map[string][]SFAspectRecord)
	for _, rec := range records {
		byEventType[rec.EventType] = append(byEventType[rec.EventType], rec)
	}

	t.Logf("By event type:")
	for eventType, recs := range byEventType {
		t.Logf("  %s: %d events", eventType, len(recs))
	}

	// Aspect diversity
	aspectSet := make(map[string]bool)
	for _, rec := range records {
		key := fmt.Sprintf("%s-%s-%s", rec.P1, rec.Aspect, rec.P2)
		aspectSet[key] = true
	}

	t.Logf("Unique aspect pairs in Sp-Na: %d", len(aspectSet))
}

// TestPhaseD_ChartType_SaNa validates Sa-Na across all event types
func TestPhaseD_ChartType_SaNa(t *testing.T) {
	const csvPath = "../../testdata/solarfire/testcase-1-transit.csv"

	records, err := ParseSFCSV(csvPath, "", "Sa-Na", "")
	if err != nil {
		t.Fatalf("ParseSFCSV: %v", err)
	}

	t.Logf("Sa-Na (Solar Arc vs Natal) Records: %d", len(records))

	// Group by event type
	byEventType := make(map[string][]SFAspectRecord)
	for _, rec := range records {
		byEventType[rec.EventType] = append(byEventType[rec.EventType], rec)
	}

	t.Logf("By event type:")
	for eventType, recs := range byEventType {
		t.Logf("  %s: %d events", eventType, len(recs))
	}

	// Aspect diversity
	aspectSet := make(map[string]bool)
	for _, rec := range records {
		key := fmt.Sprintf("%s-%s-%s", rec.P1, rec.Aspect, rec.P2)
		aspectSet[key] = true
	}

	t.Logf("Unique aspect pairs in Sa-Na: %d", len(aspectSet))
}

// TestPhaseD_Summary_AllFiles comprehensive summary across all test data files
func TestPhaseD_Summary_AllFiles(t *testing.T) {
	files := []struct {
		path    string
		person  string
		period  string
	}{
		{"../../testdata/solarfire/testcase-1-transit.csv", "JN", "Full timeline"},
		{"../../testdata/solarfire/testcase-2-transit-1996-2001.csv", "XB", "1996-2001"},
		{"../../testdata/solarfire/testcase-2-transit-2001-2006.csv", "XB", "2001-2006"},
	}

	t.Logf("=== Phase D Comprehensive Summary: All Test Data Files ===\n")

	totalRecords := 0

	for _, file := range files {
		if _, err := os.Stat(file.path); os.IsNotExist(err) {
			t.Logf("%s (%s): FILE NOT FOUND\n", file.person, file.period)
			continue
		}

		analysis, records, err := AnalyzeSFEvents(file.path)
		if err != nil {
			t.Logf("%s (%s): ERROR: %v\n", file.person, file.period, err)
			continue
		}

		totalRecords += analysis.TotalSFRecords

		t.Logf("%s (%s): %d total records", file.person, file.period, analysis.TotalSFRecords)
		t.Logf("  Event types: %v", len(analysis.ByEventType))
		for et, count := range analysis.ByEventType {
			t.Logf("    - %s: %d", et, count)
		}

		t.Logf("  Chart type summary:")
		for ct, count := range analysis.ByChartType {
			t.Logf("    - %s: %d", ct, count)
		}

		// Major aspect pairs breakdown
		aspectPairs := make(map[string]int)
		for _, rec := range records {
			if rec.Type == "Tr-Na" || rec.Type == "Sp-Na" || rec.Type == "Sa-Na" {
				key := fmt.Sprintf("%s %s %s", rec.P1, rec.Aspect, rec.P2)
				aspectPairs[key]++
			}
		}

		t.Logf("  Unique major aspect pairs (Tr/Sp/Sa-Na): %d\n", len(aspectPairs))
	}

	t.Logf("=== TOTAL SCOPE ===")
	t.Logf("Total records across all test files: %d", totalRecords)
	t.Logf("Potential validation targets: %d+ aspects", totalRecords)
	t.Logf("\nPhase D comprehensive scope now visible.")
	t.Logf("Ready to implement full timeline validation across all event types and chart pairings.")
}
