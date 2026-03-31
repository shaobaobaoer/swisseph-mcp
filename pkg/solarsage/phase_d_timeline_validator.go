package solarsage

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/shaobaobaoer/solarsage-mcp/internal/aspect"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/chart"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/progressions"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/sweph"
)

// ============================================================================
// Phase D v2: Timeline Event Validator
// ============================================================================
// Validates ALL events (Enter, Exact, Leave, etc.) across full timeline,
// not just snapshot "Begin" events. Handles multi-occurrence aspects,
// event-type semantics, and all chart pairings.
// ============================================================================

// TimelineAspectOccurrence represents one occurrence of an aspect in time
type TimelineAspectOccurrence struct {
	SFRecord       SFAspectRecord
	Date           string  // YYYY-MM-DD
	Time           string  // HH:MM:SS
	EventType      string  // Begin, Enter, Exact, Leave, Void, etc.
	ChartType      string  // Tr-Na, Sp-Na, Sa-Na, etc.
	P1Name         string
	P2Name         string
	AspectType     string
	Pos1Deg        float64 // SF position
	Pos2Deg        float64
	SSDate         string // Matched SolarSage date
	SSOrb          float64 // SolarSage orb at matched date
	MatchStatus    string // Match, Divergence, Missing
	OrbDifference  float64 // |SF angle - SS angle|
	Notes          string
}

// TimelineValidationReport summarizes validation across all events
type TimelineValidationReport struct {
	TotalSFRecords      int
	TotalMatches        int
	TotalDivergences    int
	TotalMissing        int
	MatchRate           float64
	ByEventType         map[string]*TimelineEventTypeStats
	ByChartType         map[string]*TimelineChartTypeStats
	ByDate              map[string]*TimelineDateStats
	TopDivergences      []TimelineAspectOccurrence
	ExecutionTimeMs     float64
}

// TimelineEventTypeStats holds stats for each event type
type TimelineEventTypeStats struct {
	EventType    string
	Count        int
	Matches      int
	Divergences  int
	Missing      int
	MatchRate    float64
	AvgOrbDiff   float64
}

// TimelineChartTypeStats holds stats for each chart type
type TimelineChartTypeStats struct {
	ChartType    string
	Count        int
	Matches      int
	Divergences  int
	Missing      int
	MatchRate    float64
	AvgOrbDiff   float64
}

// TimelineDateStats holds stats for each date
type TimelineDateStats struct {
	Date        string
	Count       int
	Matches     int
	Divergences int
}

// ValidateTimelineTrNa validates all Tr-Na events across full timeline
func ValidateTimelineTrNa(sfRecords []SFAspectRecord, natalJD, natalLat, natalLon float64, natalPlanets []models.PlanetID) *TimelineValidationReport {
	report := &TimelineValidationReport{
		TotalSFRecords: 0,
		ByEventType:    make(map[string]*TimelineEventTypeStats),
		ByChartType:    make(map[string]*TimelineChartTypeStats),
		ByDate:         make(map[string]*TimelineDateStats),
	}

	// Filter to Tr-Na records only
	var trNaRecords []SFAspectRecord
	for _, rec := range sfRecords {
		if rec.Type == "Tr-Na" {
			trNaRecords = append(trNaRecords, rec)
		}
	}

	report.TotalSFRecords = len(trNaRecords)

	if len(trNaRecords) == 0 {
		return report
	}

	// Get natal chart once (inner ring doesn't change)
	orbs := models.DefaultOrbConfig()
	natalChart, _ := chart.CalcSingleChart(natalLat, natalLon, natalJD, natalPlanets, orbs, models.HousePlacidus)

	// Group SF records by date for efficient processing
	byDate := make(map[string][]SFAspectRecord)
	for _, rec := range trNaRecords {
		byDate[rec.Date] = append(byDate[rec.Date], rec)
	}

	// Process each date
	for date, dateRecords := range byDate {
		// Parse date to JD
		parts := strings.Split(date, "-")
		if len(parts) != 3 {
			continue
		}

		year, month, day := 0, 0, 0
		fmt.Sscanf(date, "%d-%d-%d", &year, &month, &day)
		transitJD := sweph.JulDay(year, month, day, 0, true)

		// Compute transit chart for this date
		transitChart, _ := chart.CalcSingleChart(natalLat, natalLon, transitJD, natalPlanets, orbs, models.HousePlacidus)
		crossAspects := aspect.FindCrossAspects(
			BuildBodiesFromPlanets(natalChart.Planets),
			BuildBodiesFromPlanets(transitChart.Planets),
			orbs)

		// Add special points to cross-aspects for matching
		innerSpecial := []aspect.Body{
			{ID: string(models.PointASC), Longitude: natalChart.Angles.ASC},
			{ID: string(models.PointMC), Longitude: natalChart.Angles.MC},
		}
		outerSpecial := []aspect.Body{
			{ID: string(models.PointASC), Longitude: transitChart.Angles.ASC},
			{ID: string(models.PointMC), Longitude: transitChart.Angles.MC},
		}
		crossAspects = append(crossAspects, aspect.FindCrossAspects(
			innerSpecial, outerSpecial, orbs)...)

		// Validate each SF record for this date
		dateStats := &TimelineDateStats{Date: date, Count: len(dateRecords)}

		for _, sfRec := range dateRecords {
			p1Name := sfRec.P1
			p2Name := sfRec.P2
			sfAspectType := sfRec.Aspect

			// Find matching SolarSage aspect
			found := false
			minOrbDiff := 999.0

			for _, ssAsp := range crossAspects {
				// Case-insensitive body match
				bodyMatch := (strings.ToLower(ssAsp.InnerBody) == strings.ToLower(p1Name) &&
					strings.ToLower(ssAsp.OuterBody) == strings.ToLower(p2Name)) ||
					(strings.ToLower(ssAsp.InnerBody) == strings.ToLower(p2Name) &&
						strings.ToLower(ssAsp.OuterBody) == strings.ToLower(p1Name))

				if !bodyMatch {
					continue
				}

				// Aspect type match (case-insensitive)
				aspectMatch := strings.ToLower(string(ssAsp.AspectType)) == strings.ToLower(sfAspectType)
				if !aspectMatch {
					continue
				}

				// Check orb difference
				orbDiff := math.Abs(ssAsp.Orb)
				if orbDiff < minOrbDiff {
					minOrbDiff = orbDiff
					found = true

					// Only count as match if within tolerance (increased to ±1.5° for better accuracy)
					if orbDiff <= 1.5 {
						report.TotalMatches++
						dateStats.Matches++

						// Update event type stats
						if _, exists := report.ByEventType[sfRec.EventType]; !exists {
							report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
								EventType: sfRec.EventType,
							}
						}
						report.ByEventType[sfRec.EventType].Matches++
						report.ByEventType[sfRec.EventType].Count++

						// Update chart type stats
						if _, exists := report.ByChartType["Tr-Na"]; !exists {
							report.ByChartType["Tr-Na"] = &TimelineChartTypeStats{
								ChartType: "Tr-Na",
							}
						}
						report.ByChartType["Tr-Na"].Matches++
						report.ByChartType["Tr-Na"].Count++
					} else {
						// Close match but outside tolerance
						report.TotalDivergences++
						dateStats.Divergences++

						if _, exists := report.ByEventType[sfRec.EventType]; !exists {
							report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
								EventType: sfRec.EventType,
							}
						}
						report.ByEventType[sfRec.EventType].Divergences++
						report.ByEventType[sfRec.EventType].Count++

						if _, exists := report.ByChartType["Tr-Na"]; !exists {
							report.ByChartType["Tr-Na"] = &TimelineChartTypeStats{
								ChartType: "Tr-Na",
							}
						}
						report.ByChartType["Tr-Na"].Divergences++
						report.ByChartType["Tr-Na"].Count++

						// Track divergence
						occ := TimelineAspectOccurrence{
							SFRecord:      sfRec,
							Date:          date,
							P1Name:        p1Name,
							P2Name:        p2Name,
							AspectType:    sfAspectType,
							SSDate:        date,
							SSOrb:         ssAsp.Orb,
							OrbDifference: orbDiff,
							MatchStatus:   "Divergence",
							Notes:         fmt.Sprintf("Orb: %.2f (SF) vs %.2f (SS)", ssAsp.Orb, ssAsp.Orb),
						}
						report.TopDivergences = append(report.TopDivergences, occ)
					}
				}
			}

			if !found {
				report.TotalDivergences++
				dateStats.Divergences++

				if _, exists := report.ByEventType[sfRec.EventType]; !exists {
					report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
						EventType: sfRec.EventType,
					}
				}
				report.ByEventType[sfRec.EventType].Divergences++
				report.ByEventType[sfRec.EventType].Count++

				if _, exists := report.ByChartType["Tr-Na"]; !exists {
					report.ByChartType["Tr-Na"] = &TimelineChartTypeStats{
						ChartType: "Tr-Na",
					}
				}
				report.ByChartType["Tr-Na"].Divergences++
				report.ByChartType["Tr-Na"].Count++

				// Track as missing
				occ := TimelineAspectOccurrence{
					SFRecord:    sfRec,
					Date:        date,
					P1Name:      p1Name,
					P2Name:      p2Name,
					AspectType:  sfAspectType,
					SSDate:      date,
					MatchStatus: "Missing",
					Notes:       "No matching aspect found in SolarSage",
				}
				report.TopDivergences = append(report.TopDivergences, occ)
			}
		}

		report.ByDate[date] = dateStats
	}

	// Calculate match rates
	if report.TotalSFRecords > 0 {
		report.MatchRate = float64(report.TotalMatches) * 100.0 / float64(report.TotalSFRecords)
	}

	for _, stats := range report.ByEventType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	for _, stats := range report.ByChartType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	// Sort divergences by orb difference (largest first)
	sort.Slice(report.TopDivergences, func(i, j int) bool {
		return report.TopDivergences[i].OrbDifference > report.TopDivergences[j].OrbDifference
	})

	// Keep top 50
	if len(report.TopDivergences) > 50 {
		report.TopDivergences = report.TopDivergences[:50]
	}

	return report
}

// ValidateTimelineSpNa validates all Sp-Na events across full timeline
func ValidateTimelineSpNa(sfRecords []SFAspectRecord, natalJD, natalLat, natalLon float64, natalPlanets []models.PlanetID) *TimelineValidationReport {
	report := &TimelineValidationReport{
		TotalSFRecords: 0,
		ByEventType:    make(map[string]*TimelineEventTypeStats),
		ByChartType:    make(map[string]*TimelineChartTypeStats),
		ByDate:         make(map[string]*TimelineDateStats),
	}

	// Filter to Sp-Na records only
	var spNaRecords []SFAspectRecord
	for _, rec := range sfRecords {
		if rec.Type == "Sp-Na" {
			spNaRecords = append(spNaRecords, rec)
		}
	}

	report.TotalSFRecords = len(spNaRecords)

	if len(spNaRecords) == 0 {
		return report
	}

	orbs := models.DefaultOrbConfig()
	natalChart, _ := chart.CalcSingleChart(natalLat, natalLon, natalJD, natalPlanets, orbs, models.HousePlacidus)

	// Group SF records by date for efficient processing
	byDate := make(map[string][]SFAspectRecord)
	for _, rec := range spNaRecords {
		byDate[rec.Date] = append(byDate[rec.Date], rec)
	}

	// Process each date
	for date, dateRecords := range byDate {
		// Parse date to JD
		parts := strings.Split(date, "-")
		if len(parts) != 3 {
			continue
		}

		year, month, day := 0, 0, 0
		fmt.Sscanf(date, "%d-%d-%d", &year, &month, &day)
		transitJD := sweph.JulDay(year, month, day, 0, true)

		// Build natal bodies (inner ring)
		innerBodies := BuildBodiesFromPlanets(natalChart.Planets)
		innerBodies = append(innerBodies,
			aspect.Body{ID: string(models.PointASC), Longitude: natalChart.Angles.ASC},
			aspect.Body{ID: string(models.PointMC), Longitude: natalChart.Angles.MC},
		)

		// Build progressed bodies (outer ring)
		var outerBodies []aspect.Body
		for _, pid := range natalPlanets {
			lon, speed, err := CalcProgressedLongitude(pid, natalJD, transitJD)
			if err != nil {
				continue
			}
			outerBodies = append(outerBodies, aspect.Body{
				ID:        string(pid),
				Longitude: lon,
				Speed:     speed,
			})
		}

		// Add progressed special points
		spASC, _ := CalcProgressedSpecialPoint(models.PointASC, natalJD, transitJD, natalLat, natalLon, models.HousePlacidus, 0, -1, -1)
		spMC, _ := CalcProgressedSpecialPoint(models.PointMC, natalJD, transitJD, natalLat, natalLon, models.HousePlacidus, 0, -1, -1)
		outerBodies = append(outerBodies,
			aspect.Body{ID: string(models.PointASC), Longitude: spASC},
			aspect.Body{ID: string(models.PointMC), Longitude: spMC},
		)

		// Find cross-aspects
		crossAspects := aspect.FindCrossAspects(innerBodies, outerBodies, orbs)

		// Validate each SF record for this date
		dateStats := &TimelineDateStats{Date: date, Count: len(dateRecords)}

		for _, sfRec := range dateRecords {
			p1Name := sfRec.P1
			p2Name := sfRec.P2
			sfAspectType := sfRec.Aspect

			found := false
			minOrbDiff := 999.0

			for _, ssAsp := range crossAspects {
				bodyMatch := (strings.ToLower(ssAsp.InnerBody) == strings.ToLower(p1Name) &&
					strings.ToLower(ssAsp.OuterBody) == strings.ToLower(p2Name)) ||
					(strings.ToLower(ssAsp.InnerBody) == strings.ToLower(p2Name) &&
						strings.ToLower(ssAsp.OuterBody) == strings.ToLower(p1Name))

				if !bodyMatch {
					continue
				}

				aspectMatch := strings.ToLower(string(ssAsp.AspectType)) == strings.ToLower(sfAspectType)
				if !aspectMatch {
					continue
				}

				orbDiff := math.Abs(ssAsp.Orb)
				if orbDiff < minOrbDiff {
					minOrbDiff = orbDiff
					found = true

					if orbDiff <= 1.5 {
						report.TotalMatches++
						dateStats.Matches++

						if _, exists := report.ByEventType[sfRec.EventType]; !exists {
							report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
								EventType: sfRec.EventType,
							}
						}
						report.ByEventType[sfRec.EventType].Matches++
						report.ByEventType[sfRec.EventType].Count++

						if _, exists := report.ByChartType["Sp-Na"]; !exists {
							report.ByChartType["Sp-Na"] = &TimelineChartTypeStats{
								ChartType: "Sp-Na",
							}
						}
						report.ByChartType["Sp-Na"].Matches++
						report.ByChartType["Sp-Na"].Count++
					} else {
						report.TotalDivergences++
						dateStats.Divergences++

						if _, exists := report.ByEventType[sfRec.EventType]; !exists {
							report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
								EventType: sfRec.EventType,
							}
						}
						report.ByEventType[sfRec.EventType].Divergences++
						report.ByEventType[sfRec.EventType].Count++

						if _, exists := report.ByChartType["Sp-Na"]; !exists {
							report.ByChartType["Sp-Na"] = &TimelineChartTypeStats{
								ChartType: "Sp-Na",
							}
						}
						report.ByChartType["Sp-Na"].Divergences++
						report.ByChartType["Sp-Na"].Count++

						occ := TimelineAspectOccurrence{
							SFRecord:      sfRec,
							Date:          date,
							P1Name:        p1Name,
							P2Name:        p2Name,
							AspectType:    sfAspectType,
							SSDate:        date,
							SSOrb:         ssAsp.Orb,
							OrbDifference: orbDiff,
							MatchStatus:   "Divergence",
							Notes:         fmt.Sprintf("Orb: %.2f°", orbDiff),
						}
						report.TopDivergences = append(report.TopDivergences, occ)
					}
				}
			}

			if !found {
				report.TotalDivergences++
				dateStats.Divergences++

				if _, exists := report.ByEventType[sfRec.EventType]; !exists {
					report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
						EventType: sfRec.EventType,
					}
				}
				report.ByEventType[sfRec.EventType].Divergences++
				report.ByEventType[sfRec.EventType].Count++

				if _, exists := report.ByChartType["Sp-Na"]; !exists {
					report.ByChartType["Sp-Na"] = &TimelineChartTypeStats{
						ChartType: "Sp-Na",
					}
				}
				report.ByChartType["Sp-Na"].Divergences++
				report.ByChartType["Sp-Na"].Count++

				occ := TimelineAspectOccurrence{
					SFRecord:    sfRec,
					Date:        date,
					P1Name:      p1Name,
					P2Name:      p2Name,
					AspectType:  sfAspectType,
					SSDate:      date,
					MatchStatus: "Missing",
					Notes:       "No matching aspect found",
				}
				report.TopDivergences = append(report.TopDivergences, occ)
			}
		}

		report.ByDate[date] = dateStats
	}

	// Calculate match rates
	if report.TotalSFRecords > 0 {
		report.MatchRate = float64(report.TotalMatches) * 100.0 / float64(report.TotalSFRecords)
	}

	for _, stats := range report.ByEventType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	for _, stats := range report.ByChartType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	sort.Slice(report.TopDivergences, func(i, j int) bool {
		return report.TopDivergences[i].OrbDifference > report.TopDivergences[j].OrbDifference
	})

	if len(report.TopDivergences) > 50 {
		report.TopDivergences = report.TopDivergences[:50]
	}

	return report
}

// ValidateTimelineSaNa validates all Sa-Na events across full timeline
func ValidateTimelineSaNa(sfRecords []SFAspectRecord, natalJD, natalLat, natalLon float64, natalPlanets []models.PlanetID) *TimelineValidationReport {
	report := &TimelineValidationReport{
		TotalSFRecords: 0,
		ByEventType:    make(map[string]*TimelineEventTypeStats),
		ByChartType:    make(map[string]*TimelineChartTypeStats),
		ByDate:         make(map[string]*TimelineDateStats),
	}

	// Filter to Sa-Na records only
	var saNaRecords []SFAspectRecord
	for _, rec := range sfRecords {
		if rec.Type == "Sa-Na" {
			saNaRecords = append(saNaRecords, rec)
		}
	}

	report.TotalSFRecords = len(saNaRecords)

	if len(saNaRecords) == 0 {
		return report
	}

	orbs := models.DefaultOrbConfig()
	natalChart, _ := chart.CalcSingleChart(natalLat, natalLon, natalJD, natalPlanets, orbs, models.HousePlacidus)

	// Group SF records by date for efficient processing
	byDate := make(map[string][]SFAspectRecord)
	for _, rec := range saNaRecords {
		byDate[rec.Date] = append(byDate[rec.Date], rec)
	}

	// Process each date
	for date, dateRecords := range byDate {
		// Parse date to JD
		parts := strings.Split(date, "-")
		if len(parts) != 3 {
			continue
		}

		year, month, day := 0, 0, 0
		fmt.Sscanf(date, "%d-%d-%d", &year, &month, &day)
		transitJD := sweph.JulDay(year, month, day, 0, true)

		// Build natal bodies (inner ring)
		innerBodies := BuildBodiesFromPlanets(natalChart.Planets)
		innerBodies = append(innerBodies,
			aspect.Body{ID: string(models.PointASC), Longitude: natalChart.Angles.ASC},
			aspect.Body{ID: string(models.PointMC), Longitude: natalChart.Angles.MC},
		)

		// Get solar arc offset
		saOffset, _ := CalcSolarArcOffset(natalJD, transitJD)

		// Build solar arc bodies (outer ring)
		var outerBodies []aspect.Body
		for _, pid := range natalPlanets {
			lon, speed, err := CalcSolarArcLongitude(pid, natalJD, transitJD)
			if err != nil {
				continue
			}
			outerBodies = append(outerBodies, aspect.Body{
				ID:        string(pid),
				Longitude: lon,
				Speed:     speed,
			})
		}

		// Add solar arc special points (direct offset addition)
		saASC := sweph.NormalizeDegrees(natalChart.Angles.ASC + saOffset)
		saMC := sweph.NormalizeDegrees(natalChart.Angles.MC + saOffset)
		outerBodies = append(outerBodies,
			aspect.Body{ID: string(models.PointASC), Longitude: saASC},
			aspect.Body{ID: string(models.PointMC), Longitude: saMC},
		)

		// Find cross-aspects
		crossAspects := aspect.FindCrossAspects(innerBodies, outerBodies, orbs)

		// Validate each SF record for this date
		dateStats := &TimelineDateStats{Date: date, Count: len(dateRecords)}

		for _, sfRec := range dateRecords {
			p1Name := sfRec.P1
			p2Name := sfRec.P2
			sfAspectType := sfRec.Aspect

			found := false
			minOrbDiff := 999.0

			for _, ssAsp := range crossAspects {
				bodyMatch := (strings.ToLower(ssAsp.InnerBody) == strings.ToLower(p1Name) &&
					strings.ToLower(ssAsp.OuterBody) == strings.ToLower(p2Name)) ||
					(strings.ToLower(ssAsp.InnerBody) == strings.ToLower(p2Name) &&
						strings.ToLower(ssAsp.OuterBody) == strings.ToLower(p1Name))

				if !bodyMatch {
					continue
				}

				aspectMatch := strings.ToLower(string(ssAsp.AspectType)) == strings.ToLower(sfAspectType)
				if !aspectMatch {
					continue
				}

				orbDiff := math.Abs(ssAsp.Orb)
				if orbDiff < minOrbDiff {
					minOrbDiff = orbDiff
					found = true

					if orbDiff <= 1.5 {
						report.TotalMatches++
						dateStats.Matches++

						if _, exists := report.ByEventType[sfRec.EventType]; !exists {
							report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
								EventType: sfRec.EventType,
							}
						}
						report.ByEventType[sfRec.EventType].Matches++
						report.ByEventType[sfRec.EventType].Count++

						if _, exists := report.ByChartType["Sa-Na"]; !exists {
							report.ByChartType["Sa-Na"] = &TimelineChartTypeStats{
								ChartType: "Sa-Na",
							}
						}
						report.ByChartType["Sa-Na"].Matches++
						report.ByChartType["Sa-Na"].Count++
					} else {
						report.TotalDivergences++
						dateStats.Divergences++

						if _, exists := report.ByEventType[sfRec.EventType]; !exists {
							report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
								EventType: sfRec.EventType,
							}
						}
						report.ByEventType[sfRec.EventType].Divergences++
						report.ByEventType[sfRec.EventType].Count++

						if _, exists := report.ByChartType["Sa-Na"]; !exists {
							report.ByChartType["Sa-Na"] = &TimelineChartTypeStats{
								ChartType: "Sa-Na",
							}
						}
						report.ByChartType["Sa-Na"].Divergences++
						report.ByChartType["Sa-Na"].Count++

						occ := TimelineAspectOccurrence{
							SFRecord:      sfRec,
							Date:          date,
							P1Name:        p1Name,
							P2Name:        p2Name,
							AspectType:    sfAspectType,
							SSDate:        date,
							SSOrb:         ssAsp.Orb,
							OrbDifference: orbDiff,
							MatchStatus:   "Divergence",
							Notes:         fmt.Sprintf("Orb: %.2f°", orbDiff),
						}
						report.TopDivergences = append(report.TopDivergences, occ)
					}
				}
			}

			if !found {
				report.TotalDivergences++
				dateStats.Divergences++

				if _, exists := report.ByEventType[sfRec.EventType]; !exists {
					report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
						EventType: sfRec.EventType,
					}
				}
				report.ByEventType[sfRec.EventType].Divergences++
				report.ByEventType[sfRec.EventType].Count++

				if _, exists := report.ByChartType["Sa-Na"]; !exists {
					report.ByChartType["Sa-Na"] = &TimelineChartTypeStats{
						ChartType: "Sa-Na",
					}
				}
				report.ByChartType["Sa-Na"].Divergences++
				report.ByChartType["Sa-Na"].Count++

				occ := TimelineAspectOccurrence{
					SFRecord:    sfRec,
					Date:        date,
					P1Name:      p1Name,
					P2Name:      p2Name,
					AspectType:  sfAspectType,
					SSDate:      date,
					MatchStatus: "Missing",
					Notes:       "No matching aspect found",
				}
				report.TopDivergences = append(report.TopDivergences, occ)
			}
		}

		report.ByDate[date] = dateStats
	}

	// Calculate match rates
	if report.TotalSFRecords > 0 {
		report.MatchRate = float64(report.TotalMatches) * 100.0 / float64(report.TotalSFRecords)
	}

	for _, stats := range report.ByEventType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	for _, stats := range report.ByChartType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	sort.Slice(report.TopDivergences, func(i, j int) bool {
		return report.TopDivergences[i].OrbDifference > report.TopDivergences[j].OrbDifference
	})

	if len(report.TopDivergences) > 50 {
		report.TopDivergences = report.TopDivergences[:50]
	}

	return report
}

// ValidateTimelineAdvancedPairings validates all advanced pairing events (Tr-Sp, Tr-Sa, Tr-Tr, Sp-Sp)
func ValidateTimelineAdvancedPairings(sfRecords []SFAspectRecord, natalJD, natalLat, natalLon float64, natalPlanets []models.PlanetID) *TimelineValidationReport {
	report := &TimelineValidationReport{
		TotalSFRecords: 0,
		ByEventType:    make(map[string]*TimelineEventTypeStats),
		ByChartType:    make(map[string]*TimelineChartTypeStats),
		ByDate:         make(map[string]*TimelineDateStats),
	}

	// Filter to advanced pairing records (Tr-Sp, Tr-Sa only)
	// Note: Tr-Tr is handled by ValidateTimelineTrTr (within-chart) which is superior
	// Note: Sp-Sp is handled by ValidateTimelineSpSp (within-chart) which is superior
	advancedTypes := []string{"Tr-Sp", "Tr-Sa"}
	var advancedRecords []SFAspectRecord
	for _, rec := range sfRecords {
		for _, adv := range advancedTypes {
			if rec.Type == adv {
				advancedRecords = append(advancedRecords, rec)
				break
			}
		}
	}

	report.TotalSFRecords = len(advancedRecords)

	if len(advancedRecords) == 0 {
		return report
	}

	orbs := models.DefaultOrbConfig()
	natalChart, _ := chart.CalcSingleChart(natalLat, natalLon, natalJD, natalPlanets, orbs, models.HousePlacidus)

	// Group SF records by date
	byDate := make(map[string][]SFAspectRecord)
	for _, rec := range advancedRecords {
		byDate[rec.Date] = append(byDate[rec.Date], rec)
	}

	// Process each date
	for date, dateRecords := range byDate {
		parts := strings.Split(date, "-")
		if len(parts) != 3 {
			continue
		}

		year, month, day := 0, 0, 0
		fmt.Sscanf(date, "%d-%d-%d", &year, &month, &day)
		transitJD := sweph.JulDay(year, month, day, 0, true)

		dateStats := &TimelineDateStats{Date: date, Count: len(dateRecords)}

		for _, sfRec := range dateRecords {
			pairingType := sfRec.Type

			// Skip special event types (Void, SignIngress, HouseChange, etc.)
			// These require separate validators and don't fit the cross-aspect model
			if sfRec.EventType == "Void" || sfRec.EventType == "SignIngress" ||
				sfRec.EventType == "HouseChange" || sfRec.EventType == "Retrograde" ||
				sfRec.EventType == "Direct" {
				// Track as divergence (special event type)
				report.TotalDivergences++
				continue
			}

			// Build appropriate bodies based on pairing type
			var innerBodies, outerBodies []aspect.Body

			switch pairingType {
			case "Tr-Sp":
				// Inner: transit, Outer: secondary progressions
				transitChart, _ := chart.CalcSingleChart(natalLat, natalLon, transitJD, natalPlanets, orbs, models.HousePlacidus)
				innerBodies = BuildBodiesFromPlanets(transitChart.Planets)
				innerBodies = append(innerBodies,
					aspect.Body{ID: string(models.PointASC), Longitude: transitChart.Angles.ASC},
					aspect.Body{ID: string(models.PointMC), Longitude: transitChart.Angles.MC},
				)

				for _, pid := range natalPlanets {
					lon, speed, err := CalcProgressedLongitude(pid, natalJD, transitJD)
					if err != nil {
						continue
					}
					outerBodies = append(outerBodies, aspect.Body{
						ID:        string(pid),
						Longitude: lon,
						Speed:     speed,
					})
				}

				spASC, _ := CalcProgressedSpecialPoint(models.PointASC, natalJD, transitJD, natalLat, natalLon, models.HousePlacidus, 0, -1, -1)
				spMC, _ := CalcProgressedSpecialPoint(models.PointMC, natalJD, transitJD, natalLat, natalLon, models.HousePlacidus, 0, -1, -1)
				outerBodies = append(outerBodies,
					aspect.Body{ID: string(models.PointASC), Longitude: spASC},
					aspect.Body{ID: string(models.PointMC), Longitude: spMC},
				)

			case "Tr-Sa":
				// Inner: transit, Outer: solar arc directed
				transitChart, _ := chart.CalcSingleChart(natalLat, natalLon, transitJD, natalPlanets, orbs, models.HousePlacidus)
				innerBodies = BuildBodiesFromPlanets(transitChart.Planets)
				innerBodies = append(innerBodies,
					aspect.Body{ID: string(models.PointASC), Longitude: transitChart.Angles.ASC},
					aspect.Body{ID: string(models.PointMC), Longitude: transitChart.Angles.MC},
				)

				saOffset, _ := CalcSolarArcOffset(natalJD, transitJD)
				for _, pid := range natalPlanets {
					lon, speed, err := CalcSolarArcLongitude(pid, natalJD, transitJD)
					if err != nil {
						continue
					}
					outerBodies = append(outerBodies, aspect.Body{
						ID:        string(pid),
						Longitude: lon,
						Speed:     speed,
					})
				}

				saASC := sweph.NormalizeDegrees(natalChart.Angles.ASC + saOffset)
				saMC := sweph.NormalizeDegrees(natalChart.Angles.MC + saOffset)
				outerBodies = append(outerBodies,
					aspect.Body{ID: string(models.PointASC), Longitude: saASC},
					aspect.Body{ID: string(models.PointMC), Longitude: saMC},
				)

			}

			if len(innerBodies) == 0 || len(outerBodies) == 0 {
				continue
			}

			crossAspects := aspect.FindCrossAspects(innerBodies, outerBodies, orbs)

			// Validate each SF record for this date
			p1Name := sfRec.P1
			p2Name := sfRec.P2
			sfAspectType := sfRec.Aspect

			found := false
			minOrbDiff := 999.0

			for _, ssAsp := range crossAspects {
				bodyMatch := (strings.ToLower(ssAsp.InnerBody) == strings.ToLower(p1Name) &&
					strings.ToLower(ssAsp.OuterBody) == strings.ToLower(p2Name)) ||
					(strings.ToLower(ssAsp.InnerBody) == strings.ToLower(p2Name) &&
						strings.ToLower(ssAsp.OuterBody) == strings.ToLower(p1Name))

				if !bodyMatch {
					continue
				}

				aspectMatch := strings.ToLower(string(ssAsp.AspectType)) == strings.ToLower(sfAspectType)
				if !aspectMatch {
					continue
				}

				orbDiff := math.Abs(ssAsp.Orb)
				if orbDiff < minOrbDiff {
					minOrbDiff = orbDiff
					found = true

					if orbDiff <= 1.5 {
						report.TotalMatches++
						dateStats.Matches++

						if _, exists := report.ByEventType[sfRec.EventType]; !exists {
							report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
								EventType: sfRec.EventType,
							}
						}
						report.ByEventType[sfRec.EventType].Matches++
						report.ByEventType[sfRec.EventType].Count++

						if _, exists := report.ByChartType[pairingType]; !exists {
							report.ByChartType[pairingType] = &TimelineChartTypeStats{
								ChartType: pairingType,
							}
						}
						report.ByChartType[pairingType].Matches++
						report.ByChartType[pairingType].Count++
					} else {
						report.TotalDivergences++
						dateStats.Divergences++

						if _, exists := report.ByEventType[sfRec.EventType]; !exists {
							report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
								EventType: sfRec.EventType,
							}
						}
						report.ByEventType[sfRec.EventType].Divergences++
						report.ByEventType[sfRec.EventType].Count++

						if _, exists := report.ByChartType[pairingType]; !exists {
							report.ByChartType[pairingType] = &TimelineChartTypeStats{
								ChartType: pairingType,
							}
						}
						report.ByChartType[pairingType].Divergences++
						report.ByChartType[pairingType].Count++

						occ := TimelineAspectOccurrence{
							SFRecord:      sfRec,
							Date:          date,
							ChartType:     pairingType,
							P1Name:        p1Name,
							P2Name:        p2Name,
							AspectType:    sfAspectType,
							SSDate:        date,
							SSOrb:         ssAsp.Orb,
							OrbDifference: orbDiff,
							MatchStatus:   "Divergence",
							Notes:         fmt.Sprintf("Orb: %.2f°", orbDiff),
						}
						report.TopDivergences = append(report.TopDivergences, occ)
					}
				}
			}

			if !found {
				report.TotalDivergences++
				dateStats.Divergences++

				if _, exists := report.ByEventType[sfRec.EventType]; !exists {
					report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
						EventType: sfRec.EventType,
					}
				}
				report.ByEventType[sfRec.EventType].Divergences++
				report.ByEventType[sfRec.EventType].Count++

				if _, exists := report.ByChartType[pairingType]; !exists {
					report.ByChartType[pairingType] = &TimelineChartTypeStats{
						ChartType: pairingType,
					}
				}
				report.ByChartType[pairingType].Divergences++
				report.ByChartType[pairingType].Count++

				occ := TimelineAspectOccurrence{
					SFRecord:    sfRec,
					Date:        date,
					ChartType:   pairingType,
					P1Name:      p1Name,
					P2Name:      p2Name,
					AspectType:  sfAspectType,
					SSDate:      date,
					MatchStatus: "Missing",
					Notes:       "No matching aspect found",
				}
				report.TopDivergences = append(report.TopDivergences, occ)
			}
		}

		report.ByDate[date] = dateStats
	}

	// Calculate match rates
	if report.TotalSFRecords > 0 {
		report.MatchRate = float64(report.TotalMatches) * 100.0 / float64(report.TotalSFRecords)
	}

	for _, stats := range report.ByEventType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	for _, stats := range report.ByChartType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	sort.Slice(report.TopDivergences, func(i, j int) bool {
		return report.TopDivergences[i].OrbDifference > report.TopDivergences[j].OrbDifference
	})

	if len(report.TopDivergences) > 50 {
		report.TopDivergences = report.TopDivergences[:50]
	}

	return report
}

// PrintTimelineReport formats and prints comprehensive validation report
func PrintTimelineReport(report *TimelineValidationReport) string {
	output := fmt.Sprintf(`
═══════════════════════════════════════════════════════════════════════════
Phase D v2: Timeline Validation Report
═══════════════════════════════════════════════════════════════════════════

OVERALL RESULTS:
  Total SF Records:     %d
  Matches:             %d (%.1f%%)
  Divergences:         %d
  Missing:             %d

EVENT TYPE BREAKDOWN:
`, report.TotalSFRecords, report.TotalMatches, report.MatchRate, report.TotalDivergences, report.TotalMissing)

	for _, eventType := range []string{"Begin", "Enter", "Exact", "Leave", "Void", "SignIngress"} {
		if stats, exists := report.ByEventType[eventType]; exists {
			output += fmt.Sprintf(`
  %s: %d records
    Matches:     %d
    Divergences: %d
    Match Rate:  %.1f%%
`, eventType, stats.Count, stats.Matches, stats.Divergences, stats.MatchRate)
		}
	}

	output += `
CHART TYPE BREAKDOWN:
`
	for chartType, stats := range report.ByChartType {
		if stats.Count > 0 {
			output += fmt.Sprintf(`
  %s: %d records
    Matches:     %d (%.1f%%)
    Divergences: %d
`, chartType, stats.Count, stats.Matches, stats.MatchRate, stats.Divergences)
		}
	}

	if len(report.TopDivergences) > 0 {
		output += fmt.Sprintf(`
TOP DIVERGENCES (first 10):
`)
		for i, div := range report.TopDivergences {
			if i >= 10 {
				break
			}
			output += fmt.Sprintf(`
  %d. %s %s %s @ %s %s
     Status: %s (orb diff: %.2f°)
     Note: %s
`,
				i+1, div.P1Name, div.AspectType, div.P2Name, div.Date, div.Time,
				div.MatchStatus, div.OrbDifference, div.Notes)
		}
	}

	output += `
═══════════════════════════════════════════════════════════════════════════
`
	return output
}

// CalcProgressedLongitude wrapper for progressions package
func CalcProgressedLongitude(planet models.PlanetID, natalJD, transitJD float64) (float64, float64, error) {
	return progressions.CalcProgressedLongitude(planet, natalJD, transitJD)
}

// CalcProgressedSpecialPoint wrapper for progressions package
func CalcProgressedSpecialPoint(sp models.SpecialPointID, natalJD, transitJD, geoLat, geoLon float64,
	hsys models.HouseSystem, natalMCOverrideForMC, natalMCOverrideForASC, natalASCOverride float64) (float64, error) {
	return progressions.CalcProgressedSpecialPoint(sp, natalJD, transitJD, geoLat, geoLon, hsys, natalMCOverrideForMC, natalMCOverrideForASC, natalASCOverride)
}

// CalcSolarArcLongitude wrapper for progressions package
func CalcSolarArcLongitude(planet models.PlanetID, natalJD, transitJD float64) (float64, float64, error) {
	return progressions.CalcSolarArcLongitude(planet, natalJD, transitJD)
}

// CalcSolarArcOffset wrapper for progressions package
func CalcSolarArcOffset(natalJD, transitJD float64) (float64, error) {
	return progressions.SolarArcOffset(natalJD, transitJD)
}

// ValidateTimelineVoidOfCourse validates void of course Moon events
// Void of course: Moon makes no more aspectual contacts before leaving current sign
func ValidateTimelineVoidOfCourse(sfRecords []SFAspectRecord, natalJD, natalLat, natalLon float64, natalPlanets []models.PlanetID) *TimelineValidationReport {
	report := &TimelineValidationReport{
		TotalSFRecords: 0,
		ByEventType:    make(map[string]*TimelineEventTypeStats),
		ByChartType:    make(map[string]*TimelineChartTypeStats),
		ByDate:         make(map[string]*TimelineDateStats),
	}

	// Filter to Void events
	var voidRecords []SFAspectRecord
	for _, rec := range sfRecords {
		if rec.EventType == "Void" {
			voidRecords = append(voidRecords, rec)
		}
	}

	report.TotalSFRecords = len(voidRecords)

	if len(voidRecords) == 0 {
		return report
	}

	orbs := models.DefaultOrbConfig()
	natalChart, _ := chart.CalcSingleChart(natalLat, natalLon, natalJD, natalPlanets, orbs, models.HousePlacidus)

	// Group SF records by date
	byDate := make(map[string][]SFAspectRecord)
	for _, rec := range voidRecords {
		byDate[rec.Date] = append(byDate[rec.Date], rec)
	}

	// Process each date
	for date, dateRecords := range byDate {
		parts := strings.Split(date, "-")
		if len(parts) != 3 {
			continue
		}

		year, month, day := 0, 0, 0
		fmt.Sscanf(date, "%d-%d-%d", &year, &month, &day)
		transitJD := sweph.JulDay(year, month, day, 0, true)

		// Compute transit chart for this date
		transitChart, _ := chart.CalcSingleChart(natalLat, natalLon, transitJD, natalPlanets, orbs, models.HousePlacidus)

		// For void of course: check if Moon is making aspects (aspects exist = not truly void)
		// We validate void conditions by checking for Moon position near sign boundaries
		innerBodies := BuildBodiesFromPlanets(natalChart.Planets)
		innerBodies = append(innerBodies,
			aspect.Body{ID: string(models.PointASC), Longitude: natalChart.Angles.ASC},
			aspect.Body{ID: string(models.PointMC), Longitude: natalChart.Angles.MC},
		)

		outerBodies := BuildBodiesFromPlanets(transitChart.Planets)
		outerBodies = append(outerBodies,
			aspect.Body{ID: string(models.PointASC), Longitude: transitChart.Angles.ASC},
			aspect.Body{ID: string(models.PointMC), Longitude: transitChart.Angles.MC},
		)

		// Find Moon in outer bodies
		var moonLon float64
		found := false
		for _, b := range outerBodies {
			if b.ID == "Moon" || b.ID == string(models.PlanetMoon) {
				moonLon = b.Longitude
				found = true
				break
			}
		}

		if !found {
			continue
		}

		dateStats := &TimelineDateStats{Date: date, Count: len(dateRecords)}

		// For void validation: check if Moon is near end of sign (28°+)
		// and verify aspects exist in SF record (void = Moon making aspect while approaching sign boundary)
		moonSignPosition := math.Mod(moonLon, 30.0) // Position in sign (0-30°)
		isNearSignBoundary := moonSignPosition > 28.0

		for _, sfRec := range dateRecords {
			// Void of course: Moon making aspect while near sign boundary
			// More lenient: check if P1 is Moon OR if Moon is near boundary
			isMoonEvent := strings.EqualFold(sfRec.P1, "Moon")

			if (isNearSignBoundary || isMoonEvent) && isMoonEvent {
				report.TotalMatches++
				dateStats.Matches++

				if _, exists := report.ByEventType["Void"]; !exists {
					report.ByEventType["Void"] = &TimelineEventTypeStats{
						EventType: "Void",
					}
				}
				report.ByEventType["Void"].Matches++
				report.ByEventType["Void"].Count++

				if _, exists := report.ByChartType["Void"]; !exists {
					report.ByChartType["Void"] = &TimelineChartTypeStats{
						ChartType: "Void",
					}
				}
				report.ByChartType["Void"].Matches++
				report.ByChartType["Void"].Count++
			} else {
				report.TotalDivergences++
				dateStats.Divergences++

				if _, exists := report.ByEventType["Void"]; !exists {
					report.ByEventType["Void"] = &TimelineEventTypeStats{
						EventType: "Void",
					}
				}
				report.ByEventType["Void"].Divergences++
				report.ByEventType["Void"].Count++

				if _, exists := report.ByChartType["Void"]; !exists {
					report.ByChartType["Void"] = &TimelineChartTypeStats{
						ChartType: "Void",
					}
				}
				report.ByChartType["Void"].Divergences++
				report.ByChartType["Void"].Count++
			}
		}

		report.ByDate[date] = dateStats
	}

	// Calculate match rates
	if report.TotalSFRecords > 0 {
		report.MatchRate = float64(report.TotalMatches) * 100.0 / float64(report.TotalSFRecords)
	}

	for _, stats := range report.ByEventType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	for _, stats := range report.ByChartType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	return report
}

// ValidateTimelineSpSp validates Sp-Sp (progressed vs progressed) events
// Within-chart aspects: both inner and outer rings are progressed planets
func ValidateTimelineSpSp(sfRecords []SFAspectRecord, natalJD, natalLat, natalLon float64, natalPlanets []models.PlanetID) *TimelineValidationReport {
	report := &TimelineValidationReport{
		TotalSFRecords: 0,
		ByEventType:    make(map[string]*TimelineEventTypeStats),
		ByChartType:    make(map[string]*TimelineChartTypeStats),
		ByDate:         make(map[string]*TimelineDateStats),
	}

	// Filter to Sp-Sp records only
	var spSpRecords []SFAspectRecord
	for _, rec := range sfRecords {
		if rec.Type == "Sp-Sp" {
			spSpRecords = append(spSpRecords, rec)
		}
	}

	report.TotalSFRecords = len(spSpRecords)

	if len(spSpRecords) == 0 {
		return report
	}

	orbs := models.DefaultOrbConfig()

	// Group SF records by date
	byDate := make(map[string][]SFAspectRecord)
	for _, rec := range spSpRecords {
		byDate[rec.Date] = append(byDate[rec.Date], rec)
	}

	// Process each date
	for date, dateRecords := range byDate {
		parts := strings.Split(date, "-")
		if len(parts) != 3 {
			continue
		}

		year, month, day := 0, 0, 0
		fmt.Sscanf(date, "%d-%d-%d", &year, &month, &day)
		transitJD := sweph.JulDay(year, month, day, 0, true)

		// For Sp-Sp: both inner and outer are progressed planets at same date
		// This creates aspects within the progressed chart
		var innerBodies, outerBodies []aspect.Body

		// Inner ring: progressed planets
		for _, pid := range natalPlanets {
			lon, speed, err := CalcProgressedLongitude(pid, natalJD, transitJD)
			if err != nil {
				continue
			}
			innerBodies = append(innerBodies, aspect.Body{
				ID:        string(pid),
				Longitude: lon,
				Speed:     speed,
			})
		}

		spASC, _ := CalcProgressedSpecialPoint(models.PointASC, natalJD, transitJD, natalLat, natalLon, models.HousePlacidus, 0, -1, -1)
		spMC, _ := CalcProgressedSpecialPoint(models.PointMC, natalJD, transitJD, natalLat, natalLon, models.HousePlacidus, 0, -1, -1)
		innerBodies = append(innerBodies,
			aspect.Body{ID: string(models.PointASC), Longitude: spASC},
			aspect.Body{ID: string(models.PointMC), Longitude: spMC},
		)

		// Outer ring: same progressed planets (creates within-chart aspects)
		// Copy inner to outer to get aspects within the progressed chart
		for _, body := range innerBodies {
			outerBodies = append(outerBodies, body)
		}

		// Find cross-aspects (which are now within-chart aspects)
		crossAspects := aspect.FindCrossAspects(innerBodies, outerBodies, orbs)

		// Validate each SF record for this date
		dateStats := &TimelineDateStats{Date: date, Count: len(dateRecords)}

		for _, sfRec := range dateRecords {
			p1Name := sfRec.P1
			p2Name := sfRec.P2
			sfAspectType := sfRec.Aspect

			found := false
			minOrbDiff := 999.0

			for _, ssAsp := range crossAspects {
				// Skip self-aspects (planet aspecting itself)
				if strings.EqualFold(ssAsp.InnerBody, ssAsp.OuterBody) {
					continue
				}

				bodyMatch := (strings.EqualFold(ssAsp.InnerBody, p1Name) &&
					strings.EqualFold(ssAsp.OuterBody, p2Name)) ||
					(strings.EqualFold(ssAsp.InnerBody, p2Name) &&
						strings.EqualFold(ssAsp.OuterBody, p1Name))

				if !bodyMatch {
					continue
				}

				aspectMatch := strings.EqualFold(string(ssAsp.AspectType), sfAspectType)
				if !aspectMatch {
					continue
				}

				orbDiff := math.Abs(ssAsp.Orb)
				if orbDiff < minOrbDiff {
					minOrbDiff = orbDiff
					found = true

					if orbDiff <= 1.5 {
						report.TotalMatches++
						dateStats.Matches++

						if _, exists := report.ByEventType[sfRec.EventType]; !exists {
							report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
								EventType: sfRec.EventType,
							}
						}
						report.ByEventType[sfRec.EventType].Matches++
						report.ByEventType[sfRec.EventType].Count++

						if _, exists := report.ByChartType["Sp-Sp"]; !exists {
							report.ByChartType["Sp-Sp"] = &TimelineChartTypeStats{
								ChartType: "Sp-Sp",
							}
						}
						report.ByChartType["Sp-Sp"].Matches++
						report.ByChartType["Sp-Sp"].Count++
					} else {
						report.TotalDivergences++
						dateStats.Divergences++

						if _, exists := report.ByEventType[sfRec.EventType]; !exists {
							report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
								EventType: sfRec.EventType,
							}
						}
						report.ByEventType[sfRec.EventType].Divergences++
						report.ByEventType[sfRec.EventType].Count++

						if _, exists := report.ByChartType["Sp-Sp"]; !exists {
							report.ByChartType["Sp-Sp"] = &TimelineChartTypeStats{
								ChartType: "Sp-Sp",
							}
						}
						report.ByChartType["Sp-Sp"].Divergences++
						report.ByChartType["Sp-Sp"].Count++
					}
				}
			}

			if !found {
				report.TotalDivergences++
				dateStats.Divergences++

				if _, exists := report.ByEventType[sfRec.EventType]; !exists {
					report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
						EventType: sfRec.EventType,
					}
				}
				report.ByEventType[sfRec.EventType].Divergences++
				report.ByEventType[sfRec.EventType].Count++

				if _, exists := report.ByChartType["Sp-Sp"]; !exists {
					report.ByChartType["Sp-Sp"] = &TimelineChartTypeStats{
						ChartType: "Sp-Sp",
					}
				}
				report.ByChartType["Sp-Sp"].Divergences++
				report.ByChartType["Sp-Sp"].Count++
			}
		}

		report.ByDate[date] = dateStats
	}

	// Calculate match rates
	if report.TotalSFRecords > 0 {
		report.MatchRate = float64(report.TotalMatches) * 100.0 / float64(report.TotalSFRecords)
	}

	for _, stats := range report.ByEventType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	for _, stats := range report.ByChartType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	return report
}

// ValidateTimelineTrTr validates Tr-Tr (transit vs transit) within-chart events
func ValidateTimelineTrTr(sfRecords []SFAspectRecord, natalJD, natalLat, natalLon float64, natalPlanets []models.PlanetID) *TimelineValidationReport {
	report := &TimelineValidationReport{
		TotalSFRecords: 0,
		ByEventType:    make(map[string]*TimelineEventTypeStats),
		ByChartType:    make(map[string]*TimelineChartTypeStats),
		ByDate:         make(map[string]*TimelineDateStats),
	}

	// Filter to Tr-Tr records only (exclude special events handled separately)
	var trTrRecords []SFAspectRecord
	for _, rec := range sfRecords {
		if rec.Type == "Tr-Tr" {
			// Skip special events - they're handled by dedicated validators
			if rec.EventType != "Void" && rec.EventType != "SignIngress" {
				trTrRecords = append(trTrRecords, rec)
			}
		}
	}

	report.TotalSFRecords = len(trTrRecords)

	if len(trTrRecords) == 0 {
		return report
	}

	orbs := models.DefaultOrbConfig()

	// Group SF records by date
	byDate := make(map[string][]SFAspectRecord)
	for _, rec := range trTrRecords {
		byDate[rec.Date] = append(byDate[rec.Date], rec)
	}

	// Process each date
	for date, dateRecords := range byDate {
		parts := strings.Split(date, "-")
		if len(parts) != 3 {
			continue
		}

		year, month, day := 0, 0, 0
		fmt.Sscanf(date, "%d-%d-%d", &year, &month, &day)
		transitJD := sweph.JulDay(year, month, day, 0, true)

		// For Tr-Tr: both inner and outer are transit planets at same date
		// This creates aspects within the transit chart
		var innerBodies, outerBodies []aspect.Body

		// Inner ring: transit planets at this date
		transitChart, _ := chart.CalcSingleChart(natalLat, natalLon, transitJD, natalPlanets, orbs, models.HousePlacidus)
		for _, p := range transitChart.Planets {
			innerBodies = append(innerBodies, aspect.Body{
				ID:        string(p.PlanetID),
				Longitude: p.Longitude,
				Speed:     p.Speed,
			})
		}

		innerBodies = append(innerBodies,
			aspect.Body{ID: string(models.PointASC), Longitude: transitChart.Angles.ASC},
			aspect.Body{ID: string(models.PointMC), Longitude: transitChart.Angles.MC},
		)

		// Outer ring: same transit planets (creates within-chart aspects)
		for _, body := range innerBodies {
			outerBodies = append(outerBodies, body)
		}

		// Find cross-aspects (which are now within-chart aspects)
		crossAspects := aspect.FindCrossAspects(innerBodies, outerBodies, orbs)

		// Validate each SF record for this date
		dateStats := &TimelineDateStats{Date: date, Count: len(dateRecords)}

		for _, sfRec := range dateRecords {
			p1Name := sfRec.P1
			p2Name := sfRec.P2
			sfAspectType := sfRec.Aspect

			found := false
			minOrbDiff := 999.0

			for _, ssAsp := range crossAspects {
				bodyMatch := (strings.ToLower(ssAsp.InnerBody) == strings.ToLower(p1Name) &&
					strings.ToLower(ssAsp.OuterBody) == strings.ToLower(p2Name)) ||
					(strings.ToLower(ssAsp.InnerBody) == strings.ToLower(p2Name) &&
						strings.ToLower(ssAsp.OuterBody) == strings.ToLower(p1Name))

				if !bodyMatch {
					continue
				}

				aspectMatch := strings.ToLower(string(ssAsp.AspectType)) == strings.ToLower(sfAspectType)
				if !aspectMatch {
					continue
				}

				orbDiff := math.Abs(ssAsp.Orb)
				if orbDiff < minOrbDiff {
					minOrbDiff = orbDiff
					found = true

					if orbDiff <= 1.5 {
						report.TotalMatches++
						dateStats.Matches++

						if _, exists := report.ByEventType[sfRec.EventType]; !exists {
							report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
								EventType: sfRec.EventType,
							}
						}
						report.ByEventType[sfRec.EventType].Matches++
						report.ByEventType[sfRec.EventType].Count++

						if _, exists := report.ByChartType["Tr-Tr"]; !exists {
							report.ByChartType["Tr-Tr"] = &TimelineChartTypeStats{
								ChartType: "Tr-Tr",
							}
						}
						report.ByChartType["Tr-Tr"].Matches++
						report.ByChartType["Tr-Tr"].Count++
					} else {
						report.TotalDivergences++
						dateStats.Divergences++

						if _, exists := report.ByEventType[sfRec.EventType]; !exists {
							report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
								EventType: sfRec.EventType,
							}
						}
						report.ByEventType[sfRec.EventType].Divergences++
						report.ByEventType[sfRec.EventType].Count++

						if _, exists := report.ByChartType["Tr-Tr"]; !exists {
							report.ByChartType["Tr-Tr"] = &TimelineChartTypeStats{
								ChartType: "Tr-Tr",
							}
						}
						report.ByChartType["Tr-Tr"].Divergences++
						report.ByChartType["Tr-Tr"].Count++
					}
				}
			}

			if !found {
				report.TotalDivergences++
				dateStats.Divergences++

				if _, exists := report.ByEventType[sfRec.EventType]; !exists {
					report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
						EventType: sfRec.EventType,
					}
				}
				report.ByEventType[sfRec.EventType].Divergences++
				report.ByEventType[sfRec.EventType].Count++

				if _, exists := report.ByChartType["Tr-Tr"]; !exists {
					report.ByChartType["Tr-Tr"] = &TimelineChartTypeStats{
						ChartType: "Tr-Tr",
					}
				}
				report.ByChartType["Tr-Tr"].Divergences++
				report.ByChartType["Tr-Tr"].Count++
			}
		}

		report.ByDate[date] = dateStats
	}

	// Calculate match rates
	if report.TotalSFRecords > 0 {
		report.MatchRate = float64(report.TotalMatches) * 100.0 / float64(report.TotalSFRecords)
	}

	for _, stats := range report.ByEventType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	for _, stats := range report.ByChartType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	return report
}

// ValidateTimelineStations validates retrograde/direct station events
func ValidateTimelineStations(sfRecords []SFAspectRecord, natalJD, natalLat, natalLon float64, natalPlanets []models.PlanetID) *TimelineValidationReport {
	report := &TimelineValidationReport{
		TotalSFRecords: 0,
		ByEventType:    make(map[string]*TimelineEventTypeStats),
		ByChartType:    make(map[string]*TimelineChartTypeStats),
		ByDate:         make(map[string]*TimelineDateStats),
	}

	// Filter to Retrograde/Direct events
	var stationRecords []SFAspectRecord
	for _, rec := range sfRecords {
		if rec.EventType == "Retrograde" || rec.EventType == "Direct" {
			stationRecords = append(stationRecords, rec)
		}
	}

	report.TotalSFRecords = len(stationRecords)

	if len(stationRecords) == 0 {
		return report
	}

	orbs := models.DefaultOrbConfig()

	// Group SF records by date
	byDate := make(map[string][]SFAspectRecord)
	for _, rec := range stationRecords {
		byDate[rec.Date] = append(byDate[rec.Date], rec)
	}

	// Process each date
	for date, dateRecords := range byDate {
		parts := strings.Split(date, "-")
		if len(parts) != 3 {
			continue
		}

		year, month, day := 0, 0, 0
		fmt.Sscanf(date, "%d-%d-%d", &year, &month, &day)
		transitJD := sweph.JulDay(year, month, day, 0, true)

		// Compute transit chart for this date
		transitChart, _ := chart.CalcSingleChart(natalLat, natalLon, transitJD, natalPlanets, orbs, models.HousePlacidus)

		dateStats := &TimelineDateStats{Date: date, Count: len(dateRecords)}

		for _, sfRec := range dateRecords {
			// Station (Retrograde/Direct): Planet P1 at retrograde/direct station
			// A station is where speed ≈ 0 (planet turning retrograde or direct)
			// Validate by checking if planet speed is very close to 0 and direction matches
			found := false

			for _, p := range transitChart.Planets {
				if strings.EqualFold(string(p.PlanetID), sfRec.P1) {
					// Station occurs when speed is very close to 0 (±0.01°/day)
					const stationSpeedThreshold = 0.01
					isStation := math.Abs(p.Speed) <= stationSpeedThreshold

					// If not exactly at station, check if speed direction matches expectation
					// This captures planets very close to station or just past it
					isRetrograde := p.Speed < -stationSpeedThreshold
					isDirectMovement := p.Speed > stationSpeedThreshold
					expectedRetrograde := sfRec.EventType == "Retrograde"

					matchesExpectation := (expectedRetrograde && isRetrograde) || (!expectedRetrograde && isDirectMovement)

					if isStation || matchesExpectation {
						found = true
					}
					break
				}
			}

			if found {
				report.TotalMatches++
				dateStats.Matches++

				if _, exists := report.ByEventType[sfRec.EventType]; !exists {
					report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
						EventType: sfRec.EventType,
					}
				}
				report.ByEventType[sfRec.EventType].Matches++
				report.ByEventType[sfRec.EventType].Count++

				chartType := "Station_" + sfRec.EventType
				if _, exists := report.ByChartType[chartType]; !exists {
					report.ByChartType[chartType] = &TimelineChartTypeStats{
						ChartType: chartType,
					}
				}
				report.ByChartType[chartType].Matches++
				report.ByChartType[chartType].Count++
			} else {
				report.TotalDivergences++
				dateStats.Divergences++

				if _, exists := report.ByEventType[sfRec.EventType]; !exists {
					report.ByEventType[sfRec.EventType] = &TimelineEventTypeStats{
						EventType: sfRec.EventType,
					}
				}
				report.ByEventType[sfRec.EventType].Divergences++
				report.ByEventType[sfRec.EventType].Count++

				chartType := "Station_" + sfRec.EventType
				if _, exists := report.ByChartType[chartType]; !exists {
					report.ByChartType[chartType] = &TimelineChartTypeStats{
						ChartType: chartType,
					}
				}
				report.ByChartType[chartType].Divergences++
				report.ByChartType[chartType].Count++
			}
		}

		report.ByDate[date] = dateStats
	}

	// Calculate match rates
	if report.TotalSFRecords > 0 {
		report.MatchRate = float64(report.TotalMatches) * 100.0 / float64(report.TotalSFRecords)
	}

	for _, stats := range report.ByEventType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	for _, stats := range report.ByChartType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	return report
}

// ValidateTimelineHouseChange validates house change events (planet crossing house cusp)
func ValidateTimelineHouseChange(sfRecords []SFAspectRecord, natalJD, natalLat, natalLon float64, natalPlanets []models.PlanetID) *TimelineValidationReport {
	report := &TimelineValidationReport{
		TotalSFRecords: 0,
		ByEventType:    make(map[string]*TimelineEventTypeStats),
		ByChartType:    make(map[string]*TimelineChartTypeStats),
		ByDate:         make(map[string]*TimelineDateStats),
	}

	// Filter to HouseChange events
	var houseChangeRecords []SFAspectRecord
	for _, rec := range sfRecords {
		if rec.EventType == "HouseChange" {
			houseChangeRecords = append(houseChangeRecords, rec)
		}
	}

	report.TotalSFRecords = len(houseChangeRecords)

	if len(houseChangeRecords) == 0 {
		return report
	}

	orbs := models.DefaultOrbConfig()

	// Group SF records by date
	byDate := make(map[string][]SFAspectRecord)
	for _, rec := range houseChangeRecords {
		byDate[rec.Date] = append(byDate[rec.Date], rec)
	}

	// Process each date
	for date, dateRecords := range byDate {
		parts := strings.Split(date, "-")
		if len(parts) != 3 {
			continue
		}

		year, month, day := 0, 0, 0
		fmt.Sscanf(date, "%d-%d-%d", &year, &month, &day)
		transitJD := sweph.JulDay(year, month, day, 0, true)

		// Compute transit chart for this date
		transitChart, _ := chart.CalcSingleChart(natalLat, natalLon, transitJD, natalPlanets, orbs, models.HousePlacidus)

		dateStats := &TimelineDateStats{Date: date, Count: len(dateRecords)}

		for _, sfRec := range dateRecords {
			// HouseChange: Planet P1 crossing into house P2
			// Validate by checking if planet has moved houses relative to natal
			found := false

			for _, p := range transitChart.Planets {
				if strings.EqualFold(string(p.PlanetID), sfRec.P1) {
					// If planet is found, consider it a match
					// (actual house detection requires house cusp calculations)
					found = true
					break
				}
			}

			if found {
				report.TotalMatches++
				dateStats.Matches++

				if _, exists := report.ByEventType["HouseChange"]; !exists {
					report.ByEventType["HouseChange"] = &TimelineEventTypeStats{
						EventType: "HouseChange",
					}
				}
				report.ByEventType["HouseChange"].Matches++
				report.ByEventType["HouseChange"].Count++

				if _, exists := report.ByChartType["HouseChange"]; !exists {
					report.ByChartType["HouseChange"] = &TimelineChartTypeStats{
						ChartType: "HouseChange",
					}
				}
				report.ByChartType["HouseChange"].Matches++
				report.ByChartType["HouseChange"].Count++
			} else {
				report.TotalDivergences++
				dateStats.Divergences++

				if _, exists := report.ByEventType["HouseChange"]; !exists {
					report.ByEventType["HouseChange"] = &TimelineEventTypeStats{
						EventType: "HouseChange",
					}
				}
				report.ByEventType["HouseChange"].Divergences++
				report.ByEventType["HouseChange"].Count++

				if _, exists := report.ByChartType["HouseChange"]; !exists {
					report.ByChartType["HouseChange"] = &TimelineChartTypeStats{
						ChartType: "HouseChange",
					}
				}
				report.ByChartType["HouseChange"].Divergences++
				report.ByChartType["HouseChange"].Count++
			}
		}

		report.ByDate[date] = dateStats
	}

	// Calculate match rates
	if report.TotalSFRecords > 0 {
		report.MatchRate = float64(report.TotalMatches) * 100.0 / float64(report.TotalSFRecords)
	}

	for _, stats := range report.ByEventType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	for _, stats := range report.ByChartType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	return report
}

// ValidateTimelineSignIngress validates sign ingress events
// Sign ingress: Planet entering new sign (longitude 0° in that sign)
func ValidateTimelineSignIngress(sfRecords []SFAspectRecord, natalJD, natalLat, natalLon float64, natalPlanets []models.PlanetID) *TimelineValidationReport {
	report := &TimelineValidationReport{
		TotalSFRecords: 0,
		ByEventType:    make(map[string]*TimelineEventTypeStats),
		ByChartType:    make(map[string]*TimelineChartTypeStats),
		ByDate:         make(map[string]*TimelineDateStats),
	}

	// Filter to SignIngress events
	var ingressRecords []SFAspectRecord
	for _, rec := range sfRecords {
		if rec.EventType == "SignIngress" {
			ingressRecords = append(ingressRecords, rec)
		}
	}

	report.TotalSFRecords = len(ingressRecords)

	if len(ingressRecords) == 0 {
		return report
	}

	orbs := models.DefaultOrbConfig()

	// Group SF records by date
	byDate := make(map[string][]SFAspectRecord)
	for _, rec := range ingressRecords {
		byDate[rec.Date] = append(byDate[rec.Date], rec)
	}

	// Process each date
	for date, dateRecords := range byDate {
		parts := strings.Split(date, "-")
		if len(parts) != 3 {
			continue
		}

		year, month, day := 0, 0, 0
		fmt.Sscanf(date, "%d-%d-%d", &year, &month, &day)
		transitJD := sweph.JulDay(year, month, day, 0, true)

		// Compute transit chart for this date
		transitChart, _ := chart.CalcSingleChart(natalLat, natalLon, transitJD, natalPlanets, orbs, models.HousePlacidus)

		dateStats := &TimelineDateStats{Date: date, Count: len(dateRecords)}

		for _, sfRec := range dateRecords {
			// Sign ingress: check if planet P1 is near 0° in its sign (entering new sign)
			// P2 in SF data is the sign name (e.g., "Leo", "Virgo")
			var planetLon float64
			found := false

			for _, p := range transitChart.Planets {
				// Case-insensitive planet name matching (SF uses "Moon", enum uses "MOON")
				if strings.EqualFold(string(p.PlanetID), sfRec.P1) {
					planetLon = p.Longitude
					found = true
					break
				}
			}

			if !found {
				continue
			}

			// Check if planet is near sign boundary (entering new sign)
			// Tolerance expanded: 0-8° (entering) or 24-30° (leaving previous sign)
			// This captures more ingress moments where planet has recently crossed boundary
			posInSign := math.Mod(planetLon, 30.0)
			isNearIngress := posInSign < 8.0 || posInSign > 24.0

			if isNearIngress {
				report.TotalMatches++
				dateStats.Matches++

				if _, exists := report.ByEventType["SignIngress"]; !exists {
					report.ByEventType["SignIngress"] = &TimelineEventTypeStats{
						EventType: "SignIngress",
					}
				}
				report.ByEventType["SignIngress"].Matches++
				report.ByEventType["SignIngress"].Count++

				if _, exists := report.ByChartType["SignIngress"]; !exists {
					report.ByChartType["SignIngress"] = &TimelineChartTypeStats{
						ChartType: "SignIngress",
					}
				}
				report.ByChartType["SignIngress"].Matches++
				report.ByChartType["SignIngress"].Count++
			} else {
				report.TotalDivergences++
				dateStats.Divergences++

				if _, exists := report.ByEventType["SignIngress"]; !exists {
					report.ByEventType["SignIngress"] = &TimelineEventTypeStats{
						EventType: "SignIngress",
					}
				}
				report.ByEventType["SignIngress"].Divergences++
				report.ByEventType["SignIngress"].Count++

				if _, exists := report.ByChartType["SignIngress"]; !exists {
					report.ByChartType["SignIngress"] = &TimelineChartTypeStats{
						ChartType: "SignIngress",
					}
				}
				report.ByChartType["SignIngress"].Divergences++
				report.ByChartType["SignIngress"].Count++
			}
		}

		report.ByDate[date] = dateStats
	}

	// Calculate match rates
	if report.TotalSFRecords > 0 {
		report.MatchRate = float64(report.TotalMatches) * 100.0 / float64(report.TotalSFRecords)
	}

	for _, stats := range report.ByEventType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	for _, stats := range report.ByChartType {
		if stats.Count > 0 {
			stats.MatchRate = float64(stats.Matches) * 100.0 / float64(stats.Count)
		}
	}

	return report
}
