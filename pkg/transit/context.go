package transit

import (
	"github.com/shaobaobaoer/solarsage-mcp/pkg/chart"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/returns"
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

	// Solar Return chart data (fixed, if enabled)
	SRRefs []NatalRef
	SRJD   float64

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

	// Calculate natal houses
	natalHouses, err := chart.CalcNatalFixedHouses(input.NatalChart.Lat, input.NatalChart.Lon, input.NatalChart.JD, input.HouseSystem)
	if err != nil {
		return nil, err
	}

	// Build natal reference points
	natalRefs := buildNatalRefs(input, natalHouses)

	ctx := &CalcContext{
		NatalHouses:  natalHouses,
		NatalRefs:    natalRefs,
		NatalJD:      input.NatalChart.JD,
		StartJD:      input.TimeRange.StartJD,
		EndJD:        input.TimeRange.EndJD,
		StationCache: make(map[string][]StationInfo),
		Input:        input,
	}

	// Build Solar Return reference points if configured
	if input.Charts.SolarReturn != nil {
		srRefs, srJD, err := buildSRRefs(input)
		if err != nil {
			return nil, err
		}
		ctx.SRRefs = srRefs
		ctx.SRJD = srJD
	}

	return ctx, nil
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

// buildSRRefs calculates Solar Return reference points (planets + special points).
func buildSRRefs(input TransitCalcInput) ([]NatalRef, float64, error) {
	srCfg := input.Charts.SolarReturn

	// Resolve SR JD
	var srJD float64
	if srCfg.SRChartJD != 0 {
		srJD = srCfg.SRChartJD
	} else {
		// Calculate SR by finding when Sun returns to natal longitude
		searchJD := srCfg.SearchAfterJD
		if searchJD == 0 {
			searchJD = input.TimeRange.StartJD
		}
		ret, err := returns.CalcSolarReturn(returns.ReturnInput{
			NatalJD:     srCfg.NatalJD,
			NatalLat:    srCfg.Lat,
			NatalLon:    srCfg.Lon,
			SearchJD:    searchJD,
			Planets:     srCfg.Planets,
			OrbConfig:   srCfg.Orbs,
			HouseSystem: srCfg.HouseSystem,
		})
		if err != nil {
			return nil, 0, err
		}
		srJD = ret.ReturnJD
	}

	var refs []NatalRef

	// SR planets
	for _, pid := range srCfg.Planets {
		lon, _, err := chart.CalcPlanetLongitude(pid, srJD)
		if err != nil {
			continue
		}
		refs = append(refs, NatalRef{
			ID:        string(pid),
			Longitude: lon,
			ChartType: models.ChartSolarReturn,
		})
	}

	// SR special points
	for _, sp := range srCfg.Points {
		lon, err := chart.CalcSpecialPointLongitude(sp, srCfg.Lat, srCfg.Lon, srJD, srCfg.HouseSystem)
		if err != nil {
			continue
		}
		refs = append(refs, NatalRef{
			ID:        string(sp),
			Longitude: lon,
			ChartType: models.ChartSolarReturn,
		})
	}

	return refs, srJD, nil
}
