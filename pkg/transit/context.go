package transit

import (
	"github.com/shaobaobaoer/solarsage-mcp/pkg/chart"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
)

// NatalRef represents a fixed natal reference point (planet or special point).
type NatalRef struct {
	ID        string
	Longitude float64
	ChartType models.ChartType
}

// CalcContext holds all pre-calculated read-only data for transit calculation,
// avoiding redundant computations across different task types.
type CalcContext struct {
	// Natal chart data (fixed)
	NatalHouses []float64
	NatalRefs   []NatalRef
	NatalJD     float64

	// Time range
	StartJD float64
	EndJD   float64

	// Station cache: key = planet/special point ID
	StationCache map[string][]StationInfo

	// Original input (for factory functions)
	Input TransitCalcInput
}

// buildCalcContext pre-calculates all fixed data needed for transit calculation.
func buildCalcContext(input TransitCalcInput) (*CalcContext, error) {
	// Normalize input: map old fields to new structure if needed
	input = normalizeInput(input)

	// Calculate natal houses
	natalHouses, err := chart.CalcNatalFixedHouses(input.NatalChart.Lat, input.NatalChart.Lon, input.NatalChart.JD, input.HouseSystem)
	if err != nil {
		return nil, err
	}

	// Build natal reference points
	natalRefs := buildNatalRefs(input, natalHouses)

	return &CalcContext{
		NatalHouses:  natalHouses,
		NatalRefs:    natalRefs,
		NatalJD:      input.NatalChart.JD,
		StartJD:      input.TimeRange.StartJD,
		EndJD:        input.TimeRange.EndJD,
		StationCache: make(map[string][]StationInfo),
		Input:        input,
	}, nil
}

// normalizeInput maps old flat fields to new structured fields if the new fields are not set.
func normalizeInput(input TransitCalcInput) TransitCalcInput {
	// If new structure is already populated, use it as-is
	if input.NatalChart.JD != 0 {
		return input
	}

	// Map old fields to new structure
	input.NatalChart = NatalChartConfig{
		Lat:                       input.NatalLat,
		Lon:                       input.NatalLon,
		JD:                        input.NatalJD,
		Planets:                   input.NatalPlanets,
		ASCOverride:               input.NatalASC,
		MCOverride:                input.NatalMC,
		MCOverrideForASC:          input.NatalMCForASC,
		ASCOverrideForProgressions: input.NatalASCForProgressions,
		PlanetOverrides:           input.NatalPlanetOverrides,
	}

	if input.SpecialPoints != nil {
		input.NatalChart.Points = input.SpecialPoints.NatalPoints
	}

	input.TimeRange = TimeRangeConfig{
		StartJD: input.StartJD,
		EndJD:   input.EndJD,
	}

	input.EventFilter = EventFilterConfig{
		Station:      input.EventConfig.IncludeStation,
		SignIngress:  input.EventConfig.IncludeSignIngress,
		HouseIngress: input.EventConfig.IncludeHouseIngress,
		VoidOfCourse: input.EventConfig.IncludeVoidOfCourse,
		TrNa:         input.EventConfig.IncludeTrNa,
		TrTr:         input.EventConfig.IncludeTrTr,
		TrSp:         input.EventConfig.IncludeTrSp,
		TrSa:         input.EventConfig.IncludeTrSa,
		SpNa:         input.EventConfig.IncludeSpNa,
		SpSp:         input.EventConfig.IncludeSpSp,
		SaNa:         input.EventConfig.IncludeSaNa,
	}

	// Build Transit chart config if enabled
	if len(input.TransitPlanets) > 0 {
		input.Charts.Transit = &TransitChartConfig{
			Lat:         input.TransitLat,
			Lon:         input.TransitLon,
			Planets:     input.TransitPlanets,
			Orbs:        input.OrbConfigTransit,
			HouseSystem: input.HouseSystem,
		}
		if input.SpecialPoints != nil {
			input.Charts.Transit.Points = input.SpecialPoints.TransitPoints
		}
	}

	// Build Progressions chart config if enabled
	if input.ProgressionsConfig != nil && input.ProgressionsConfig.Enabled {
		input.Charts.Progressions = &ProgressionsChartConfig{
			Planets:     input.ProgressionsConfig.Planets,
			Orbs:        input.OrbConfigProgressions,
			Lat:         input.TransitLat,
			Lon:         input.TransitLon,
			HouseSystem: input.HouseSystem,
		}
		if input.SpecialPoints != nil {
			input.Charts.Progressions.Points = input.SpecialPoints.ProgressionsPoints
		}
	}

	// Build SolarArc chart config if enabled
	if input.SolarArcConfig != nil && input.SolarArcConfig.Enabled {
		input.Charts.SolarArc = &SolarArcChartConfig{
			Planets:     input.SolarArcConfig.Planets,
			Orbs:        input.OrbConfigSolarArc,
			Lat:         input.TransitLat,
			Lon:         input.TransitLon,
			HouseSystem: input.HouseSystem,
		}
		if input.SpecialPoints != nil {
			input.Charts.SolarArc.Points = input.SpecialPoints.SolarArcPoints
		}
	}

	return input
}

// buildNatalRefs collects all natal reference points (planets + special points).
func buildNatalRefs(input TransitCalcInput, natalHouses []float64) []NatalRef {
	var refs []NatalRef

	// Natal planets
	for _, pid := range input.NatalChart.Planets {
		var lon float64
		var err error

		// Check for planet override
		if input.NatalChart.PlanetOverrides != nil {
			if override, ok := input.NatalChart.PlanetOverrides[string(pid)]; ok {
				lon = override
			} else {
				lon, _, err = chart.CalcPlanetLongitude(pid, input.NatalChart.JD)
			}
		} else {
			lon, _, err = chart.CalcPlanetLongitude(pid, input.NatalChart.JD)
		}

		if err != nil {
			continue
		}
		refs = append(refs, NatalRef{
			ID:        string(pid),
			Longitude: lon,
			ChartType: models.ChartNatal,
		})
	}

	// Natal special points
	for _, sp := range input.NatalChart.Points {
		var lon float64
		var err error

		// Use override values for ASC/MC if provided
		switch sp {
		case models.PointASC:
			if input.NatalChart.ASCOverride != 0 {
				lon = input.NatalChart.ASCOverride
			} else {
				lon, err = chart.CalcSpecialPointLongitude(sp, input.NatalChart.Lat, input.NatalChart.Lon, input.NatalChart.JD, input.HouseSystem)
			}
		case models.PointMC:
			if input.NatalChart.MCOverride != 0 {
				lon = input.NatalChart.MCOverride
			} else {
				lon, err = chart.CalcSpecialPointLongitude(sp, input.NatalChart.Lat, input.NatalChart.Lon, input.NatalChart.JD, input.HouseSystem)
			}
		default:
			lon, err = chart.CalcSpecialPointLongitude(sp, input.NatalChart.Lat, input.NatalChart.Lon, input.NatalChart.JD, input.HouseSystem)
		}

		if err != nil {
			continue
		}
		refs = append(refs, NatalRef{
			ID:        string(sp),
			Longitude: lon,
			ChartType: models.ChartNatal,
		})
	}

	return refs
}

// GetStations returns cached stations for a planet, computing them if needed.
func (ctx *CalcContext) GetStations(key string, calcFn bodyCalcFunc, planet models.PlanetID) []StationInfo {
	if stations, ok := ctx.StationCache[key]; ok {
		return stations
	}

	stations := findStations(calcFn, ctx.StartJD, ctx.EndJD, planet)
	ctx.StationCache[key] = stations
	return stations
}
