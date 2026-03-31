package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/solarsage"
)

// Phase D Timeline Validation API Endpoints
// Validates transit events against Solar Fire reference data

// TimelineValidationRequest represents input for timeline validation
type TimelineValidationRequest struct {
	CSVPath          string              `json:"csv_path"`          // Path to Solar Fire CSV
	NatalJD          float64             `json:"natal_jd"`          // Natal JD (UT)
	NatalLat         float64             `json:"natal_lat"`         // Natal latitude
	NatalLon         float64             `json:"natal_lon"`         // Natal longitude
	Planets          []string            `json:"planets,omitempty"` // Planet list (optional)
	ValidatorType    string              `json:"validator_type"`    // Validator to run
	Format           string              `json:"format,omitempty"`  // Output format (json/csv)
}

// TimelineValidationResponse represents validation results
type TimelineValidationResponse struct {
	ValidatorType    string                 `json:"validator_type"`
	TotalRecords     int                    `json:"total_records"`
	Matches          int                    `json:"matches"`
	Divergences      int                    `json:"divergences"`
	MatchRate        float64                `json:"match_rate"`
	ExecutionTimeMs  float64                `json:"execution_time_ms"`
	ByEventType      map[string]EventStats  `json:"by_event_type,omitempty"`
	ByChartType      map[string]ChartStats  `json:"by_chart_type,omitempty"`
	Summary          string                 `json:"summary"`
}

// EventStats represents statistics for an event type
type EventStats struct {
	EventType   string  `json:"event_type"`
	Matches     int     `json:"matches"`
	Divergences int     `json:"divergences"`
	Total       int     `json:"total"`
	MatchRate   float64 `json:"match_rate"`
}

// ChartStats represents statistics for a chart type
type ChartStats struct {
	ChartType   string  `json:"chart_type"`
	Matches     int     `json:"matches"`
	Divergences int     `json:"divergences"`
	Total       int     `json:"total"`
	MatchRate   float64 `json:"match_rate"`
}

// handleTimelineValidation validates all Phase D validators
// POST /api/v1/validation/timeline
func (s *Server) handleTimelineValidation(w http.ResponseWriter, r *http.Request) {
	var req TimelineValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	// Parse planets
	planets := defaultPlanets
	if len(req.Planets) > 0 {
		planets = make([]models.PlanetID, len(req.Planets))
		for i, p := range req.Planets {
			planets[i] = models.PlanetID(p)
		}
	}

	// Load SF CSV
	sfRecords, err := solarsage.ParseSFCSV(req.CSVPath, "", "", "")
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to load CSV: "+err.Error())
		return
	}

	// Route to appropriate validator
	var report *solarsage.TimelineValidationReport

	switch req.ValidatorType {
	case "tr-na":
		report = solarsage.ValidateTimelineTrNa(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets)
	case "sp-na":
		report = solarsage.ValidateTimelineSpNa(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets)
	case "sa-na":
		report = solarsage.ValidateTimelineSaNa(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets)
	case "advanced":
		report = solarsage.ValidateTimelineAdvancedPairings(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets)
	case "sp-sp":
		report = solarsage.ValidateTimelineSpSp(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets)
	case "tr-tr":
		report = solarsage.ValidateTimelineTrTr(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets)
	case "void":
		report = solarsage.ValidateTimelineVoidOfCourse(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets)
	case "signingress":
		report = solarsage.ValidateTimelineSignIngress(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets)
	case "housechange":
		report = solarsage.ValidateTimelineHouseChange(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets)
	case "stations":
		report = solarsage.ValidateTimelineStations(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets)
	default:
		writeError(w, http.StatusBadRequest, "unknown validator type: "+req.ValidatorType)
		return
	}

	// Convert report to response
	resp := TimelineValidationResponse{
		ValidatorType:   req.ValidatorType,
		TotalRecords:    report.TotalSFRecords,
		Matches:         report.TotalMatches,
		Divergences:     report.TotalDivergences,
		MatchRate:       report.MatchRate,
		ExecutionTimeMs: report.ExecutionTimeMs,
		Summary:         buildSummary(req.ValidatorType, report),
	}

	// Add breakdown if requested
	if req.Format == "detailed" {
		resp.ByEventType = make(map[string]EventStats)
		for et, stats := range report.ByEventType {
			resp.ByEventType[et] = EventStats{
				EventType:   et,
				Matches:     stats.Matches,
				Divergences: stats.Divergences,
				Total:       stats.Count,
				MatchRate:   stats.MatchRate,
			}
		}

		resp.ByChartType = make(map[string]ChartStats)
		for ct, stats := range report.ByChartType {
			resp.ByChartType[ct] = ChartStats{
				ChartType:   ct,
				Matches:     stats.Matches,
				Divergences: stats.Divergences,
				Total:       stats.Count,
				MatchRate:   stats.MatchRate,
			}
		}
	}

	// Export as CSV if requested
	if req.Format == "csv" {
		csv, err := ExportDetailedValidationCSV(req.ValidatorType, report)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to export CSV: "+err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename="+req.ValidatorType+"-validation.csv")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(csv))
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// handlePhaseDAggregated validates all Phase D validators in one call
// POST /api/v1/validation/phase-d
func (s *Server) handlePhaseDAggregated(w http.ResponseWriter, r *http.Request) {
	var req TimelineValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	// Parse planets
	planets := defaultPlanets
	if len(req.Planets) > 0 {
		planets = make([]models.PlanetID, len(req.Planets))
		for i, p := range req.Planets {
			planets[i] = models.PlanetID(p)
		}
	}

	// Load SF CSV
	sfRecords, err := solarsage.ParseSFCSV(req.CSVPath, "", "", "")
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to load CSV: "+err.Error())
		return
	}

	// Run all validators
	validators := map[string]func() *solarsage.TimelineValidationReport{
		"tr-na":      func() *solarsage.TimelineValidationReport { return solarsage.ValidateTimelineTrNa(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets) },
		"sp-na":      func() *solarsage.TimelineValidationReport { return solarsage.ValidateTimelineSpNa(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets) },
		"sa-na":      func() *solarsage.TimelineValidationReport { return solarsage.ValidateTimelineSaNa(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets) },
		"advanced":   func() *solarsage.TimelineValidationReport { return solarsage.ValidateTimelineAdvancedPairings(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets) },
		"sp-sp":      func() *solarsage.TimelineValidationReport { return solarsage.ValidateTimelineSpSp(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets) },
		"tr-tr":      func() *solarsage.TimelineValidationReport { return solarsage.ValidateTimelineTrTr(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets) },
		"void":       func() *solarsage.TimelineValidationReport { return solarsage.ValidateTimelineVoidOfCourse(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets) },
		"signingress": func() *solarsage.TimelineValidationReport { return solarsage.ValidateTimelineSignIngress(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets) },
		"housechange": func() *solarsage.TimelineValidationReport { return solarsage.ValidateTimelineHouseChange(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets) },
		"stations":   func() *solarsage.TimelineValidationReport { return solarsage.ValidateTimelineStations(sfRecords, req.NatalJD, req.NatalLat, req.NatalLon, planets) },
	}

	results := make(map[string]TimelineValidationResponse)
	reports := make(map[string]*solarsage.TimelineValidationReport) // For CSV export
	totalEvents := 0
	totalMatches := 0

	for name, fn := range validators {
		report := fn()
		if report == nil || report.TotalSFRecords == 0 {
			continue
		}

		resp := TimelineValidationResponse{
			ValidatorType:   name,
			TotalRecords:    report.TotalSFRecords,
			Matches:         report.TotalMatches,
			Divergences:     report.TotalDivergences,
			MatchRate:       report.MatchRate,
			ExecutionTimeMs: report.ExecutionTimeMs,
			Summary:         buildSummary(name, report),
		}

		results[name] = resp
		reports[name] = report // Store for CSV export
		totalEvents += report.TotalSFRecords
		totalMatches += report.TotalMatches
	}

	// Calculate overall coverage
	overallRate := 0.0
	if totalEvents > 0 {
		overallRate = float64(totalMatches) * 100.0 / float64(totalEvents)
	}

	// Export as CSV if requested
	if req.Format == "csv" {
		csv, err := ExportAggregatedValidationCSV(reports)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to export CSV: "+err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=phase-d-validation.csv")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(csv))
		return
	}

	response := map[string]interface{}{
		"validators":     results,
		"total_events":   totalEvents,
		"total_matches":  totalMatches,
		"overall_coverage": overallRate,
		"status":         "complete",
	}

	writeJSON(w, http.StatusOK, response)
}

// buildSummary creates a human-readable summary for a validator
func buildSummary(validatorType string, report *solarsage.TimelineValidationReport) string {
	if report.TotalSFRecords == 0 {
		return "No records to validate"
	}

	rate := strconv.FormatFloat(report.MatchRate, 'f', 1, 64)
	return validatorType + ": " + strconv.Itoa(report.TotalMatches) + "/" + strconv.Itoa(report.TotalSFRecords) + " (" + rate + "%)"
}
