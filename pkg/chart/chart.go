package chart

import (
	"fmt"
	"math"

	"github.com/anthropic/swisseph-mcp/internal/aspect"
	"github.com/anthropic/swisseph-mcp/pkg/models"
	"github.com/anthropic/swisseph-mcp/pkg/sweph"
)

// CalcSingleChart computes a single chart (natal or event chart)
func CalcSingleChart(lat, lon, jdUT float64, planets []models.PlanetID, orbs models.OrbConfig, hsys models.HouseSystem) (*models.ChartInfo, error) {
	hsysChar := models.HouseSystemToChar(hsys)

	// Calculate houses
	houseResult, err := sweph.Houses(jdUT, lat, lon, hsysChar)
	if err != nil {
		return nil, fmt.Errorf("house calculation failed: %w", err)
	}

	// Get obliquity for house position calculation
	eps, err := sweph.Obliquity(jdUT)
	if err != nil {
		return nil, fmt.Errorf("obliquity calculation failed: %w", err)
	}

	houses := make([]float64, 12)
	for i := 0; i < 12; i++ {
		houses[i] = houseResult.Cusps[i+1]
	}

	angles := models.AnglesInfo{
		ASC: houseResult.ASC,
		MC:  houseResult.MC,
		DSC: sweph.NormalizeDegrees(houseResult.ASC + 180),
		IC:  sweph.NormalizeDegrees(houseResult.MC + 180),
	}

	// Calculate planet positions
	var positions []models.PlanetPosition
	var bodies []aspect.Body

	for _, pid := range planets {
		pos, err := calcPlanetPosition(pid, jdUT, houseResult.ARMC, lat, eps, hsysChar, houses)
		if err != nil {
			return nil, err
		}
		positions = append(positions, *pos)
		bodies = append(bodies, aspect.Body{
			ID:        string(pid),
			Longitude: pos.Longitude,
			Speed:     pos.Speed,
		})
	}

	// Calculate aspects between all planets (same set)
	aspects := aspect.FindAspects(bodies, bodies, orbs, true)

	return &models.ChartInfo{
		Planets: positions,
		Houses:  houses,
		Angles:  angles,
		Aspects: aspects,
	}, nil
}

// CalcDoubleChart computes a double chart (transit or synastry)
func CalcDoubleChart(
	innerLat, innerLon, innerJD float64, innerPlanets []models.PlanetID,
	outerLat, outerLon, outerJD float64, outerPlanets []models.PlanetID,
	specialPoints *models.SpecialPointsConfig,
	orbs models.OrbConfig, hsys models.HouseSystem,
) (*models.ChartInfo, *models.ChartInfo, []models.CrossAspectInfo, error) {

	// Calculate inner chart
	innerChart, err := CalcSingleChart(innerLat, innerLon, innerJD, innerPlanets, orbs, hsys)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("inner chart: %w", err)
	}

	// Calculate outer chart
	outerChart, err := CalcSingleChart(outerLat, outerLon, outerJD, outerPlanets, orbs, hsys)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("outer chart: %w", err)
	}

	// Build body lists for cross-aspect calculation
	var innerBodies, outerBodies []aspect.Body

	for _, p := range innerChart.Planets {
		innerBodies = append(innerBodies, aspect.Body{
			ID:        string(p.PlanetID),
			Longitude: p.Longitude,
			Speed:     p.Speed,
		})
	}
	for _, p := range outerChart.Planets {
		outerBodies = append(outerBodies, aspect.Body{
			ID:        string(p.PlanetID),
			Longitude: p.Longitude,
			Speed:     p.Speed,
		})
	}

	// Add special points if configured
	if specialPoints != nil {
		innerSP := calcSpecialPoints(specialPoints.InnerPoints, innerChart.Angles,
			innerLat, innerLon, innerJD, orbs, hsys)
		innerBodies = append(innerBodies, innerSP...)

		outerSP := calcSpecialPoints(specialPoints.OuterPoints, outerChart.Angles,
			outerLat, outerLon, outerJD, orbs, hsys)
		outerBodies = append(outerBodies, outerSP...)
	}

	crossAspects := aspect.FindCrossAspects(innerBodies, outerBodies, orbs)

	return innerChart, outerChart, crossAspects, nil
}

// calcPlanetPosition calculates position for a single planet
func calcPlanetPosition(pid models.PlanetID, jdUT, armc, lat, eps float64, hsysChar int, houses []float64) (*models.PlanetPosition, error) {
	// Handle South Node as opposite of North Node
	if pid == models.PlanetSouthNode {
		northPos, err := calcPlanetPosition(models.PlanetNorthNodeTrue, jdUT, armc, lat, eps, hsysChar, houses)
		if err != nil {
			return nil, err
		}
		lon := sweph.NormalizeDegrees(northPos.Longitude + 180)
		return &models.PlanetPosition{
			PlanetID:     pid,
			Longitude:    lon,
			Latitude:     -northPos.Latitude,
			Speed:        northPos.Speed,
			IsRetrograde: northPos.IsRetrograde,
			Sign:         models.SignFromLongitude(lon),
			SignDegree:   models.SignDegreeFromLongitude(lon),
			House:        findHouse(lon, houses),
		}, nil
	}

	sweID, ok := models.PlanetToSweID(pid)
	if !ok {
		return nil, fmt.Errorf("unknown planet: %s", pid)
	}

	result, err := sweph.CalcUT(jdUT, sweID)
	if err != nil {
		return nil, fmt.Errorf("calc %s: %w", pid, err)
	}

	house := findHouseFromPos(armc, lat, eps, hsysChar, result.Longitude, result.Latitude)

	return &models.PlanetPosition{
		PlanetID:     pid,
		Longitude:    result.Longitude,
		Latitude:     result.Latitude,
		Speed:        result.SpeedLong,
		IsRetrograde: result.IsRetrograde,
		Sign:         models.SignFromLongitude(result.Longitude),
		SignDegree:   models.SignDegreeFromLongitude(result.Longitude),
		House:        house,
	}, nil
}

// findHouseFromPos uses swe_house_pos to determine house number
func findHouseFromPos(armc, lat, eps float64, hsysChar int, lon, eclLat float64) int {
	pos, err := sweph.HousePos(armc, lat, eps, hsysChar, lon, eclLat)
	if err != nil {
		return findHouse(lon, nil)
	}
	return int(pos)
}

// findHouse determines which house a longitude falls in, using house cusps
func findHouse(lon float64, cusps []float64) int {
	if len(cusps) < 12 {
		return int(lon/30.0) + 1
	}
	for i := 0; i < 12; i++ {
		next := (i + 1) % 12
		c1 := cusps[i]
		c2 := cusps[next]
		if c2 < c1 { // wraps around 0°
			if lon >= c1 || lon < c2 {
				return i + 1
			}
		} else {
			if lon >= c1 && lon < c2 {
				return i + 1
			}
		}
	}
	return 1
}

// calcSpecialPoints calculates special point bodies for aspect computation
func calcSpecialPoints(points []models.SpecialPointID, angles models.AnglesInfo,
	lat, lon, jdUT float64, orbs models.OrbConfig, hsys models.HouseSystem) []aspect.Body {

	var bodies []aspect.Body

	hsysChar := models.HouseSystemToChar(hsys)
	houseResult, _ := sweph.Houses(jdUT, lat, lon, hsysChar)

	for _, p := range points {
		var longitude float64
		switch p {
		case models.PointASC:
			longitude = angles.ASC
		case models.PointMC:
			longitude = angles.MC
		case models.PointDSC:
			longitude = angles.DSC
		case models.PointIC:
			longitude = angles.IC
		case models.PointVertex:
			if houseResult != nil {
				longitude = houseResult.Vertex
			}
		case models.PointAntiVertex:
			if houseResult != nil {
				longitude = sweph.NormalizeDegrees(houseResult.Vertex + 180)
			}
		case models.PointEastPoint:
			if houseResult != nil {
				longitude = houseResult.EqASC
			}
		case models.PointLotFortune:
			longitude = calcLotOfFortune(jdUT, angles.ASC)
		case models.PointLotSpirit:
			longitude = calcLotOfSpirit(jdUT, angles.ASC)
		default:
			continue
		}

		bodies = append(bodies, aspect.Body{
			ID:        string(p),
			Longitude: longitude,
			Speed:     0, // special points treated as fixed for aspect applying/separating
		})
	}

	return bodies
}

// calcLot computes an Arabic lot: ASC + bodyA - bodyB (day) or ASC + bodyB - bodyA (night)
func calcLot(jdUT, asc float64, dayA, dayB int) float64 {
	a, _ := sweph.CalcUT(jdUT, dayA)
	b, _ := sweph.CalcUT(jdUT, dayB)
	if a == nil || b == nil {
		return 0
	}
	if isDayChart(a.Longitude, asc) {
		return sweph.NormalizeDegrees(asc + b.Longitude - a.Longitude)
	}
	return sweph.NormalizeDegrees(asc + a.Longitude - b.Longitude)
}

// calcLotOfFortune: Day = ASC + Moon - Sun; Night = ASC + Sun - Moon
func calcLotOfFortune(jdUT, asc float64) float64 {
	return calcLot(jdUT, asc, sweph.SE_SUN, sweph.SE_MOON)
}

// calcLotOfSpirit: Day = ASC + Sun - Moon; Night = ASC + Moon - Sun
func calcLotOfSpirit(jdUT, asc float64) float64 {
	return calcLot(jdUT, asc, sweph.SE_MOON, sweph.SE_SUN)
}

// isDayChart checks if Sun is above the horizon (simplified check)
func isDayChart(sunLon, asc float64) bool {
	dsc := sweph.NormalizeDegrees(asc + 180)
	// Sun is above horizon if it's in the upper half (ASC counter-clockwise to DSC)
	diff := sweph.NormalizeDegrees(sunLon - dsc)
	return diff >= 0 && diff < 180
}

// CalcPlanetLongitude is a helper to get a planet's longitude at a given time
func CalcPlanetLongitude(pid models.PlanetID, jdUT float64) (longitude float64, speed float64, err error) {
	if pid == models.PlanetSouthNode {
		sweID, _ := models.PlanetToSweID(models.PlanetNorthNodeTrue)
		r, err := sweph.CalcUT(jdUT, sweID)
		if err != nil {
			return 0, 0, err
		}
		return sweph.NormalizeDegrees(r.Longitude + 180), r.SpeedLong, nil
	}

	sweID, ok := models.PlanetToSweID(pid)
	if !ok {
		return 0, 0, fmt.Errorf("unknown planet: %s", pid)
	}
	r, err := sweph.CalcUT(jdUT, sweID)
	if err != nil {
		return 0, 0, err
	}
	return r.Longitude, r.SpeedLong, nil
}

// CalcSpecialPointLongitude calculates the longitude of a special point at a given time
func CalcSpecialPointLongitude(sp models.SpecialPointID, lat, lon, jdUT float64, hsys models.HouseSystem) (float64, error) {
	hsysChar := models.HouseSystemToChar(hsys)
	hr, err := sweph.Houses(jdUT, lat, lon, hsysChar)
	if err != nil {
		return 0, err
	}

	switch sp {
	case models.PointASC:
		return hr.ASC, nil
	case models.PointMC:
		return hr.MC, nil
	case models.PointDSC:
		return sweph.NormalizeDegrees(hr.ASC + 180), nil
	case models.PointIC:
		return sweph.NormalizeDegrees(hr.MC + 180), nil
	case models.PointVertex:
		return hr.Vertex, nil
	case models.PointAntiVertex:
		return sweph.NormalizeDegrees(hr.Vertex + 180), nil
	case models.PointEastPoint:
		return hr.EqASC, nil
	case models.PointLotFortune:
		return calcLotOfFortune(jdUT, hr.ASC), nil
	case models.PointLotSpirit:
		return calcLotOfSpirit(jdUT, hr.ASC), nil
	default:
		return 0, fmt.Errorf("unknown special point: %s", sp)
	}
}

// CalcNatalFixedHouses returns house cusps for natal chart (fixed for transit calculations)
func CalcNatalFixedHouses(lat, lon, jdUT float64, hsys models.HouseSystem) ([]float64, error) {
	hsysChar := models.HouseSystemToChar(hsys)
	hr, err := sweph.Houses(jdUT, lat, lon, hsysChar)
	if err != nil {
		return nil, err
	}
	cusps := make([]float64, 12)
	for i := 0; i < 12; i++ {
		cusps[i] = hr.Cusps[i+1]
	}
	return cusps, nil
}

// FindHouseForLongitude determines which house a planet is in given fixed natal cusps
func FindHouseForLongitude(lon float64, cusps []float64) int {
	return findHouse(lon, cusps)
}

// WrapAngle normalizes an angle to [-180, 180)
func WrapAngle(a float64) float64 {
	a = math.Mod(a+180, 360)
	if a < 0 {
		a += 360
	}
	return a - 180
}
