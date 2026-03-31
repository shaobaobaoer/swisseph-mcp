package api

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/shaobaobaoer/solarsage-mcp/pkg/solarsage"
)

// ExportValidationResultsCSV converts validation results to CSV format
func ExportValidationResultsCSV(validatorType string, report *solarsage.TimelineValidationReport) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"Validator",
		"Total_Records",
		"Matches",
		"Divergences",
		"Match_Rate_%",
		"Execution_Time_ms",
		"Summary",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("write header: %w", err)
	}

	// Write summary row
	rate := fmt.Sprintf("%.1f", report.MatchRate)
	execTime := fmt.Sprintf("%.2f", report.ExecutionTimeMs)
	summary := validatorType + ": " + strconv.Itoa(report.TotalMatches) + "/" + strconv.Itoa(report.TotalSFRecords) + " (" + rate + "%)"

	row := []string{
		validatorType,
		strconv.Itoa(report.TotalSFRecords),
		strconv.Itoa(report.TotalMatches),
		strconv.Itoa(report.TotalDivergences),
		rate,
		execTime,
		summary,
	}
	if err := writer.Write(row); err != nil {
		return "", fmt.Errorf("write row: %w", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("write CSV: %w", err)
	}

	return buf.String(), nil
}

// ExportDetailedValidationCSV exports breakdown by event type and chart type
func ExportDetailedValidationCSV(validatorType string, report *solarsage.TimelineValidationReport) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header with two sections: Summary + Event Type Breakdown
	header := []string{
		"Section",
		"Category",
		"Matches",
		"Divergences",
		"Total",
		"Match_Rate_%",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("write header: %w", err)
	}

	// Summary row
	rate := fmt.Sprintf("%.1f", report.MatchRate)
	summaryRow := []string{
		"Summary",
		validatorType,
		strconv.Itoa(report.TotalMatches),
		strconv.Itoa(report.TotalDivergences),
		strconv.Itoa(report.TotalSFRecords),
		rate,
	}
	if err := writer.Write(summaryRow); err != nil {
		return "", fmt.Errorf("write summary: %w", err)
	}

	// Event Type breakdown
	for eventType, stats := range report.ByEventType {
		eventRate := fmt.Sprintf("%.1f", stats.MatchRate)
		eventRow := []string{
			"EventType",
			eventType,
			strconv.Itoa(stats.Matches),
			strconv.Itoa(stats.Divergences),
			strconv.Itoa(stats.Count),
			eventRate,
		}
		if err := writer.Write(eventRow); err != nil {
			return "", fmt.Errorf("write event type: %w", err)
		}
	}

	// Chart Type breakdown
	for chartType, stats := range report.ByChartType {
		chartRate := fmt.Sprintf("%.1f", stats.MatchRate)
		chartRow := []string{
			"ChartType",
			chartType,
			strconv.Itoa(stats.Matches),
			strconv.Itoa(stats.Divergences),
			strconv.Itoa(stats.Count),
			chartRate,
		}
		if err := writer.Write(chartRow); err != nil {
			return "", fmt.Errorf("write chart type: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("finalize CSV: %w", err)
	}

	return buf.String(), nil
}

// ExportAggregatedValidationCSV exports results from all validators
func ExportAggregatedValidationCSV(results map[string]*solarsage.TimelineValidationReport) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"Validator",
		"Total_Records",
		"Matches",
		"Divergences",
		"Match_Rate_%",
		"Execution_Time_ms",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("write header: %w", err)
	}

	// Write each validator's results
	for validatorName, report := range results {
		if report == nil || report.TotalSFRecords == 0 {
			continue
		}

		rate := fmt.Sprintf("%.1f", report.MatchRate)
		execTime := fmt.Sprintf("%.2f", report.ExecutionTimeMs)

		row := []string{
			validatorName,
			strconv.Itoa(report.TotalSFRecords),
			strconv.Itoa(report.TotalMatches),
			strconv.Itoa(report.TotalDivergences),
			rate,
			execTime,
		}
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("write validator row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("finalize CSV: %w", err)
	}

	return buf.String(), nil
}
