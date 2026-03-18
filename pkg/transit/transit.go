package transit

import (
	"math"
	"sort"

	"github.com/anthropic/swisseph-mcp/pkg/chart"
	"github.com/anthropic/swisseph-mcp/pkg/models"
	"github.com/anthropic/swisseph-mcp/pkg/progressions"
	"github.com/anthropic/swisseph-mcp/pkg/sweph"
)

const (
	dayStep   = 1.0         // 1 day coarse scan step
	bisectEps = 1.0 / 86400 // ~1 second precision
	fineStep  = 0.5         // fine scan step for fast movers (transit planets)
)

// adaptiveStep returns an appropriate scan step based on body speeds.
// For fast movers (transit planets): 0.5 day
// For slow movers (progressions, solar arc): up to 7 days
func adaptiveStep(speed1, speed2 float64) float64 {
	maxSpeed := math.Abs(speed1)
	if math.Abs(speed2) > maxSpeed {
		maxSpeed = math.Abs(speed2)
	}
	if maxSpeed < 0.01 {
		// Very slow (solar arc): 7 days still gives <0.07° resolution
		return 7.0
	}
	if maxSpeed < 0.1 {
		// Slow (progressions): 2 days
		return 2.0
	}
	return fineStep
}

// StationInfo represents a retrograde/direct station
type StationInfo struct {
	JD          float64
	IsDirecting bool // true = station direct (retro -> direct)
}

// MonoInterval represents a monotonic longitude interval (no station inside)
type MonoInterval struct {
	Start float64
	End   float64
}

// TransitCalcInput holds all inputs for transit calculation
type TransitCalcInput struct {
	NatalLat     float64
	NatalLon     float64
	NatalJD      float64
	NatalPlanets []models.PlanetID

	TransitLat float64
	TransitLon float64

	StartJD float64
	EndJD   float64

	TransitPlanets []models.PlanetID

	ProgressionsConfig *models.ProgressionsConfig
	SolarArcConfig     *models.SolarArcConfig

	SpecialPoints *models.SpecialPointsConfig
	EventConfig   models.EventConfig

	OrbConfigTransit      models.OrbConfig
	OrbConfigProgressions models.OrbConfig
	OrbConfigSolarArc     models.OrbConfig

	HouseSystem models.HouseSystem
}

// bodyCalcFunc is a function that returns longitude and speed at a given JD
type bodyCalcFunc func(jd float64) (lon, speed float64, err error)

// CalcTransitEvents computes all transit events in the given time range
func CalcTransitEvents(input TransitCalcInput) ([]models.TransitEvent, error) {
	var allEvents []models.TransitEvent

	// Pre-calculate natal chart data (fixed)
	natalHouses, err := chart.CalcNatalFixedHouses(input.NatalLat, input.NatalLon, input.NatalJD, input.HouseSystem)
	if err != nil {
		return nil, err
	}

	// Collect natal reference points (planets + special points) - fixed positions
	type refPoint struct {
		ID        string
		Longitude float64
	}
	var natalRefs []refPoint

	for _, pid := range input.NatalPlanets {
		lon, _, err := chart.CalcPlanetLongitude(pid, input.NatalJD)
		if err != nil {
			continue
		}
		natalRefs = append(natalRefs, refPoint{ID: string(pid), Longitude: lon})
	}
	if input.SpecialPoints != nil {
		for _, sp := range input.SpecialPoints.NatalPoints {
			lon, err := chart.CalcSpecialPointLongitude(sp, input.NatalLat, input.NatalLon, input.NatalJD, input.HouseSystem)
			if err != nil {
				continue
			}
			natalRefs = append(natalRefs, refPoint{ID: string(sp), Longitude: lon})
		}
	}

	// =====================================================================
	// Transit planets: stations, sign/house ingress, Tr-Na, Tr-Tr, Tr-Sp, Tr-Sa
	// =====================================================================
	for _, tPlanet := range input.TransitPlanets {
		calcFn := makeTransitCalcFn(tPlanet)

		// Find stations
		stations := findStations(calcFn, input.StartJD, input.EndJD, tPlanet)
		intervals := buildMonoIntervals(input.StartJD, input.EndJD, stations)

		// Station events
		if input.EventConfig.IncludeStation {
			for _, st := range stations {
				lon, _, _ := calcFn(st.JD)
				stType := models.StationRetrograde
				if st.IsDirecting {
					stType = models.StationDirect
				}
				allEvents = append(allEvents, models.TransitEvent{
					EventType:       models.EventStation,
					ChartType:       models.ChartTransit,
					Planet:          tPlanet,
					JD:              st.JD,
					Age:             progressions.Age(input.NatalJD, st.JD),
					PlanetLongitude: lon,
					PlanetSign:      models.SignFromLongitude(lon),
					PlanetHouse:     chart.FindHouseForLongitude(lon, natalHouses),
					IsRetrograde:    stType == models.StationRetrograde,
					StationType:     stType,
				})
			}
		}

		// Sign ingress (transit planet)
		if input.EventConfig.IncludeSignIngress {
			events := findSignIngressEvents(calcFn, tPlanet, models.ChartTransit, intervals, natalHouses, input.NatalJD)
			allEvents = append(allEvents, events...)
		}

		// House ingress (transit planet)
		if input.EventConfig.IncludeHouseIngress {
			events := findHouseIngressEvents(calcFn, tPlanet, models.ChartTransit, intervals, natalHouses, input.NatalJD)
			allEvents = append(allEvents, events...)
		}

		// Tr-Na: transit planet vs fixed natal point (RQ1)
		if input.EventConfig.IncludeTrNa {
			exactCounters := make(map[string]int)
			for _, ref := range natalRefs {
				events := findAspectEventsRQ1(
					calcFn, tPlanet, models.ChartTransit,
					ref.ID, ref.Longitude, models.ChartNatal,
					intervals, input.OrbConfigTransit, natalHouses, exactCounters,
					input.NatalJD,
				)
				allEvents = append(allEvents, events...)
			}
		}

		// Tr-Tr: transit planet vs other transit planets (RQ2)
		if input.EventConfig.IncludeTrTr {
			for _, tPlanet2 := range input.TransitPlanets {
				if string(tPlanet) >= string(tPlanet2) {
					continue // avoid duplicates
				}
				calcFn2 := makeTransitCalcFn(tPlanet2)
				events := findAspectEventsRQ2(
					calcFn, tPlanet, models.ChartTransit,
					calcFn2, tPlanet2, models.ChartTransit,
					input.StartJD, input.EndJD,
					input.OrbConfigTransit, natalHouses, input.NatalJD,
				)
				allEvents = append(allEvents, events...)
			}
		}

		// Tr vs transit special points (dynamic, RQ2)
		if input.EventConfig.IncludeTrTr && input.SpecialPoints != nil {
			for _, sp := range input.SpecialPoints.TransitPoints {
				calcFnSP := makeTransitSpecialPointCalcFn(sp, input.TransitLat, input.TransitLon, input.HouseSystem)
				events := findAspectEventsRQ2(
					calcFn, tPlanet, models.ChartTransit,
					calcFnSP, models.PlanetID(sp), models.ChartTransit,
					input.StartJD, input.EndJD,
					input.OrbConfigTransit, natalHouses, input.NatalJD,
				)
				allEvents = append(allEvents, events...)
			}
		}

		// Tr-Sp: transit planet vs progressed planets (RQ2)
		if input.EventConfig.IncludeTrSp && input.ProgressionsConfig != nil && input.ProgressionsConfig.Enabled {
			for _, pPlanet := range input.ProgressionsConfig.Planets {
				calcFnP := makeProgressionsCalcFn(pPlanet, input.NatalJD)
				events := findAspectEventsRQ2(
					calcFn, tPlanet, models.ChartTransit,
					calcFnP, pPlanet, models.ChartProgressions,
					input.StartJD, input.EndJD,
					input.OrbConfigTransit, natalHouses, input.NatalJD,
				)
				allEvents = append(allEvents, events...)
			}
			// Tr vs progressed special points
			if input.SpecialPoints != nil {
				for _, sp := range input.SpecialPoints.ProgressionsPoints {
					calcFnPSP := makeProgressionsSpecialPointCalcFn(sp, input.TransitLat, input.TransitLon, input.NatalJD, input.HouseSystem)
					events := findAspectEventsRQ2(
						calcFn, tPlanet, models.ChartTransit,
						calcFnPSP, models.PlanetID(sp), models.ChartProgressions,
						input.StartJD, input.EndJD,
						input.OrbConfigTransit, natalHouses, input.NatalJD,
					)
					allEvents = append(allEvents, events...)
				}
			}
		}

		// Tr-Sa: transit planet vs solar arc planets (RQ2)
		if input.EventConfig.IncludeTrSa && input.SolarArcConfig != nil && input.SolarArcConfig.Enabled {
			for _, saPlanet := range input.SolarArcConfig.Planets {
				calcFnSA := makeSolarArcCalcFn(saPlanet, input.NatalJD)
				events := findAspectEventsRQ2(
					calcFn, tPlanet, models.ChartTransit,
					calcFnSA, saPlanet, models.ChartSolarArc,
					input.StartJD, input.EndJD,
					input.OrbConfigTransit, natalHouses, input.NatalJD,
				)
				allEvents = append(allEvents, events...)
			}
			// Tr vs solar arc special points (ASC, MC)
			if input.SpecialPoints != nil {
				for _, sp := range input.SpecialPoints.SolarArcPoints {
					calcFnSASP := makeSolarArcSpecialPointCalcFn(sp, input.NatalLat, input.NatalLon, input.NatalJD, input.HouseSystem)
					events := findAspectEventsRQ2(
						calcFn, tPlanet, models.ChartTransit,
						calcFnSASP, models.PlanetID(sp), models.ChartSolarArc,
						input.StartJD, input.EndJD,
						input.OrbConfigTransit, natalHouses, input.NatalJD,
					)
					allEvents = append(allEvents, events...)
				}
			}
		}
	}

	// =====================================================================
	// Secondary Progressions: Sp-Na, Sp-Sp
	// =====================================================================
	if input.ProgressionsConfig != nil && input.ProgressionsConfig.Enabled {
		for _, pPlanet := range input.ProgressionsConfig.Planets {
			calcFnP := makeProgressionsCalcFn(pPlanet, input.NatalJD)

			// Progressed planets move very slowly, stations are rare but possible
			pStations := findStations(calcFnP, input.StartJD, input.EndJD, pPlanet)
			pIntervals := buildMonoIntervals(input.StartJD, input.EndJD, pStations)

			// Station events for progressed planets
			if input.EventConfig.IncludeStation {
				for _, st := range pStations {
					lon, _, _ := calcFnP(st.JD)
					stType := models.StationRetrograde
					if st.IsDirecting {
						stType = models.StationDirect
					}
					allEvents = append(allEvents, models.TransitEvent{
						EventType:       models.EventStation,
						ChartType:       models.ChartProgressions,
						Planet:          pPlanet,
						JD:              st.JD,
						Age:             progressions.Age(input.NatalJD, st.JD),
						PlanetLongitude: lon,
						PlanetSign:      models.SignFromLongitude(lon),
						PlanetHouse:     chart.FindHouseForLongitude(lon, natalHouses),
						IsRetrograde:    stType == models.StationRetrograde,
						StationType:     stType,
					})
				}
			}

			// Sign ingress for progressed planets
			if input.EventConfig.IncludeSignIngress {
				events := findSignIngressEvents(calcFnP, pPlanet, models.ChartProgressions, pIntervals, natalHouses, input.NatalJD)
				allEvents = append(allEvents, events...)
			}

			// House ingress for progressed planets
			if input.EventConfig.IncludeHouseIngress {
				events := findHouseIngressEvents(calcFnP, pPlanet, models.ChartProgressions, pIntervals, natalHouses, input.NatalJD)
				allEvents = append(allEvents, events...)
			}

			// Sp-Na: progressed planet vs fixed natal (RQ1)
			if input.EventConfig.IncludeSpNa {
				exactCounters := make(map[string]int)
				for _, ref := range natalRefs {
					events := findAspectEventsRQ1(
						calcFnP, pPlanet, models.ChartProgressions,
						ref.ID, ref.Longitude, models.ChartNatal,
						pIntervals, input.OrbConfigProgressions, natalHouses, exactCounters,
						input.NatalJD,
					)
					allEvents = append(allEvents, events...)
				}
			}

			// Sp-Sp: progressed planet vs other progressed planets (RQ2)
			if input.EventConfig.IncludeSpSp {
				for _, pPlanet2 := range input.ProgressionsConfig.Planets {
					if string(pPlanet) >= string(pPlanet2) {
						continue
					}
					calcFnP2 := makeProgressionsCalcFn(pPlanet2, input.NatalJD)
					events := findAspectEventsRQ2(
						calcFnP, pPlanet, models.ChartProgressions,
						calcFnP2, pPlanet2, models.ChartProgressions,
						input.StartJD, input.EndJD,
						input.OrbConfigProgressions, natalHouses, input.NatalJD,
					)
					allEvents = append(allEvents, events...)
				}
			}
		}
	}

	// =====================================================================
	// Solar Arc: Sa-Na
	// =====================================================================
	if input.SolarArcConfig != nil && input.SolarArcConfig.Enabled {
		for _, saPlanet := range input.SolarArcConfig.Planets {
			calcFnSA := makeSolarArcCalcFn(saPlanet, input.NatalJD)

			// Solar arc planets move at ~1°/year, no retrograde
			saIntervals := []MonoInterval{{Start: input.StartJD, End: input.EndJD}}

			// Sa-Na: solar arc planet vs fixed natal (RQ1)
			if input.EventConfig.IncludeSaNa {
				exactCounters := make(map[string]int)
				for _, ref := range natalRefs {
					events := findAspectEventsRQ1(
						calcFnSA, saPlanet, models.ChartSolarArc,
						ref.ID, ref.Longitude, models.ChartNatal,
						saIntervals, input.OrbConfigSolarArc, natalHouses, exactCounters,
						input.NatalJD,
					)
					allEvents = append(allEvents, events...)
				}
			}

			// Sign ingress for solar arc planets
			if input.EventConfig.IncludeSignIngress {
				events := findSignIngressEvents(calcFnSA, saPlanet, models.ChartSolarArc, saIntervals, natalHouses, input.NatalJD)
				allEvents = append(allEvents, events...)
			}

			// House ingress for solar arc planets
			if input.EventConfig.IncludeHouseIngress {
				events := findHouseIngressEvents(calcFnSA, saPlanet, models.ChartSolarArc, saIntervals, natalHouses, input.NatalJD)
				allEvents = append(allEvents, events...)
			}
		}

		// Sa vs solar arc special points (RQ2-like, but solar arc points move at same rate)
		if input.EventConfig.IncludeSaNa && input.SpecialPoints != nil {
			for _, sp := range input.SpecialPoints.SolarArcPoints {
				// Solar arc special point = natal special point + solar arc offset
				// This is a dynamic point, so use RQ2 scan against natal refs
				calcFnSASP := makeSolarArcSpecialPointCalcFn(sp, input.NatalLat, input.NatalLon, input.NatalJD, input.HouseSystem)
				saIntervals := []MonoInterval{{Start: input.StartJD, End: input.EndJD}}
				exactCounters := make(map[string]int)
				for _, ref := range natalRefs {
					events := findAspectEventsRQ1(
						calcFnSASP, models.PlanetID(sp), models.ChartSolarArc,
						ref.ID, ref.Longitude, models.ChartNatal,
						saIntervals, input.OrbConfigSolarArc, natalHouses, exactCounters,
						input.NatalJD,
					)
					allEvents = append(allEvents, events...)
				}
			}
		}
	}

	// =====================================================================
	// Void of Course Moon
	// =====================================================================
	if input.EventConfig.IncludeVoidOfCourse {
		vocEvents := findVoidOfCourse(allEvents, input.StartJD, input.EndJD)
		allEvents = append(allEvents, vocEvents...)
	}

	// Sort all events by JD
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].JD < allEvents[j].JD
	})

	return allEvents, nil
}

// =====================================================================
// Calc function factories
// =====================================================================

func makeTransitCalcFn(planet models.PlanetID) bodyCalcFunc {
	return func(jd float64) (float64, float64, error) {
		return chart.CalcPlanetLongitude(planet, jd)
	}
}

func makeProgressionsCalcFn(planet models.PlanetID, natalJD float64) bodyCalcFunc {
	return func(jd float64) (float64, float64, error) {
		return progressions.CalcProgressedLongitude(planet, natalJD, jd)
	}
}

func makeSolarArcCalcFn(planet models.PlanetID, natalJD float64) bodyCalcFunc {
	return func(jd float64) (float64, float64, error) {
		return progressions.CalcSolarArcLongitude(planet, natalJD, jd)
	}
}

// makeTransitSpecialPointCalcFn creates a calc function for a dynamic transit special point
func makeTransitSpecialPointCalcFn(sp models.SpecialPointID, lat, lon float64, hsys models.HouseSystem) bodyCalcFunc {
	return func(jd float64) (float64, float64, error) {
		spLon, err := chart.CalcSpecialPointLongitude(sp, lat, lon, jd, hsys)
		if err != nil {
			return 0, 0, err
		}
		// Estimate speed numerically (forward finite difference)
		dt := 0.01
		spLon2, err := chart.CalcSpecialPointLongitude(sp, lat, lon, jd+dt, hsys)
		if err != nil {
			return spLon, 0, nil
		}
		speed := (spLon2 - spLon) / dt
		// Handle wrap-around
		if speed > 180 {
			speed -= 360
		} else if speed < -180 {
			speed += 360
		}
		return spLon, speed, nil
	}
}

// makeProgressionsSpecialPointCalcFn creates a calc function for a progressed special point
// Uses the Q1/Solar Arc in RA method for ASC/MC progression (standard Solar Fire method).
func makeProgressionsSpecialPointCalcFn(sp models.SpecialPointID, lat, lon float64, natalJD float64, hsys models.HouseSystem) bodyCalcFunc {
	return func(jd float64) (float64, float64, error) {
		spLon, err := progressions.CalcProgressedSpecialPoint(sp, natalJD, jd, lat, lon, hsys)
		if err != nil {
			return 0, 0, err
		}
		// Estimate speed numerically
		dt := 1.0 // 1 day step for numerical derivative (angles move slowly in progressions)
		spLon2, err := progressions.CalcProgressedSpecialPoint(sp, natalJD, jd+dt, lat, lon, hsys)
		if err != nil {
			return spLon, 0, nil
		}
		speed := spLon2 - spLon
		if speed > 180 {
			speed -= 360
		} else if speed < -180 {
			speed += 360
		}
		return spLon, speed, nil
	}
}

// makeSolarArcSpecialPointCalcFn creates a calc function for a solar arc directed special point
// Solar arc special point = natal special point longitude + solar arc offset
func makeSolarArcSpecialPointCalcFn(sp models.SpecialPointID, lat, lon float64, natalJD float64, hsys models.HouseSystem) bodyCalcFunc {
	// Pre-compute natal special point longitude (fixed)
	natalSpLon, _ := chart.CalcSpecialPointLongitude(sp, lat, lon, natalJD, hsys)
	return func(jd float64) (float64, float64, error) {
		offset, err := progressions.SolarArcOffset(natalJD, jd)
		if err != nil {
			return 0, 0, err
		}
		directed := sweph.NormalizeDegrees(natalSpLon + offset)
		// Speed ~ sun's progressed speed / JulianYear
		pJD := progressions.SecondaryProgressionJD(natalJD, jd)
		_, sunSpeed, _ := chart.CalcPlanetLongitude(models.PlanetSun, pJD)
		speed := sunSpeed / progressions.JulianYear
		return directed, speed, nil
	}
}

// =====================================================================
// Station detection
// =====================================================================

func findStations(calcFn bodyCalcFunc, startJD, endJD float64, planet models.PlanetID) []StationInfo {
	var stations []StationInfo

	// Sun and Moon never retrograde
	if planet == models.PlanetSun || planet == models.PlanetMoon {
		return stations
	}

	prevSpeed := getSpeed(calcFn, startJD)

	for jd := startJD + dayStep; jd <= endJD; jd += dayStep {
		curSpeed := getSpeed(calcFn, jd)

		if prevSpeed*curSpeed < 0 {
			stationJD := bisectStation(calcFn, jd-dayStep, jd)
			isDirecting := prevSpeed < 0 && curSpeed > 0
			stations = append(stations, StationInfo{
				JD:          stationJD,
				IsDirecting: isDirecting,
			})
		}
		prevSpeed = curSpeed
	}

	return stations
}

func getSpeed(calcFn bodyCalcFunc, jd float64) float64 {
	_, speed, err := calcFn(jd)
	if err != nil {
		return 0
	}
	return speed
}

func bisectStation(calcFn bodyCalcFunc, lo, hi float64) float64 {
	speedLo := getSpeed(calcFn, lo)

	for hi-lo > bisectEps {
		mid := (lo + hi) / 2
		speedMid := getSpeed(calcFn, mid)
		if speedLo*speedMid <= 0 {
			hi = mid
		} else {
			lo = mid
			speedLo = speedMid
		}
	}
	return (lo + hi) / 2
}

// =====================================================================
// Monotonic intervals
// =====================================================================

func buildMonoIntervals(startJD, endJD float64, stations []StationInfo) []MonoInterval {
	var intervals []MonoInterval
	prev := startJD
	for _, st := range stations {
		if st.JD > prev && st.JD <= endJD {
			intervals = append(intervals, MonoInterval{Start: prev, End: st.JD})
			prev = st.JD
		}
	}
	if prev < endJD {
		intervals = append(intervals, MonoInterval{Start: prev, End: endJD})
	}
	return intervals
}

// =====================================================================
// RQ1: moving body vs fixed reference point
// =====================================================================

func findAspectEventsRQ1(
	calcFn bodyCalcFunc, planet models.PlanetID, chartType models.ChartType,
	targetID string, targetLon float64, targetChartType models.ChartType,
	intervals []MonoInterval, orbs models.OrbConfig,
	natalHouses []float64, exactCounters map[string]int,
	natalJD float64,
) []models.TransitEvent {
	var events []models.TransitEvent

	for _, asp := range models.StandardAspects {
		orb := orbs.GetOrb(asp.Type)
		if orb == 0 {
			continue
		}

		counterKey := targetID + ":" + string(asp.Type)
		inAspect := false

		// Check initial state and emit Begin event if already in aspect
		if len(intervals) > 0 {
			initLon, _, _ := calcFn(intervals[0].Start)
			initDiff := angleDiffToAspect(initLon, targetLon, asp.Angle)
			if math.Abs(initDiff) <= orb {
				inAspect = true
				initRetro := getSpeed(calcFn, intervals[0].Start) < 0
				events = append(events, models.TransitEvent{
					EventType:          models.EventAspectBegin,
					ChartType:          chartType,
					Planet:             planet,
					JD:                 intervals[0].Start,
					Age:                progressions.Age(natalJD, intervals[0].Start),
					PlanetLongitude:    initLon,
					PlanetSign:         models.SignFromLongitude(initLon),
					PlanetHouse:        chart.FindHouseForLongitude(initLon, natalHouses),
					IsRetrograde:       initRetro,
					TargetChartType:    targetChartType,
					Target:             targetID,
					TargetLongitude:    targetLon,
					TargetSign:         models.SignFromLongitude(targetLon),
					TargetHouse:        chart.FindHouseForLongitude(targetLon, natalHouses),
					TargetIsRetrograde: false,
					AspectType:         asp.Type,
					AspectAngle:        asp.Angle,
				})
			}
		}

		for _, interval := range intervals {
			prevJD := interval.Start
			prevLon, prevSpeed, _ := calcFn(prevJD)
			prevDiff := angleDiffToAspect(prevLon, targetLon, asp.Angle)
			step := adaptiveStep(prevSpeed, 0) // target is fixed (speed=0)

			for jd := interval.Start + step; jd <= interval.End+step*0.5; jd += step {
				if jd > interval.End {
					jd = interval.End
				}
				curLon, _, _ := calcFn(jd)
				curDiff := angleDiffToAspect(curLon, targetLon, asp.Angle)

				// ENTER
				if !inAspect && math.Abs(curDiff) <= orb && math.Abs(prevDiff) > orb {
					enterJD := bisectThreshold(calcFn, targetLon, asp.Angle, orb, prevJD, jd, true)
					eLon, _, _ := calcFn(enterJD)
					eRetro := getSpeed(calcFn, enterJD) < 0
					events = append(events, models.TransitEvent{
						EventType:          models.EventAspectEnter,
						ChartType:          chartType,
						Planet:             planet,
						JD:                 enterJD,
						Age:                progressions.Age(natalJD, enterJD),
						PlanetLongitude:    eLon,
						PlanetSign:         models.SignFromLongitude(eLon),
						PlanetHouse:        chart.FindHouseForLongitude(eLon, natalHouses),
						IsRetrograde:       eRetro,
						TargetChartType:    targetChartType,
						Target:             targetID,
						TargetLongitude:    targetLon,
						TargetSign:         models.SignFromLongitude(targetLon),
						TargetHouse:        chart.FindHouseForLongitude(targetLon, natalHouses),
						TargetIsRetrograde: false,
						AspectType:         asp.Type,
						AspectAngle:        asp.Angle,
						OrbAtEnter:         math.Abs(angleDiffToAspect(eLon, targetLon, asp.Angle)),
					})
					inAspect = true
				}

				// EXACT: sign change AND both diffs are within a reasonable range
				// (guards against wrap-around false positives at ±180 boundary)
				if prevDiff*curDiff < 0 && math.Abs(prevDiff) < 90 && math.Abs(curDiff) < 90 {
					exactJD := bisectExact(calcFn, targetLon, asp.Angle, prevJD, jd)
					exactCounters[counterKey]++
					eLon, _, _ := calcFn(exactJD)
					eRetro := getSpeed(calcFn, exactJD) < 0
					events = append(events, models.TransitEvent{
						EventType:          models.EventAspectExact,
						ChartType:          chartType,
						Planet:             planet,
						JD:                 exactJD,
						Age:                progressions.Age(natalJD, exactJD),
						PlanetLongitude:    eLon,
						PlanetSign:         models.SignFromLongitude(eLon),
						PlanetHouse:        chart.FindHouseForLongitude(eLon, natalHouses),
						IsRetrograde:       eRetro,
						TargetChartType:    targetChartType,
						Target:             targetID,
						TargetLongitude:    targetLon,
						TargetSign:         models.SignFromLongitude(targetLon),
						TargetHouse:        chart.FindHouseForLongitude(targetLon, natalHouses),
						TargetIsRetrograde: false,
						AspectType:         asp.Type,
						AspectAngle:        asp.Angle,
						ExactCount:         exactCounters[counterKey],
					})
				}

				// LEAVE
				if inAspect && math.Abs(curDiff) > orb && math.Abs(prevDiff) <= orb {
					leaveJD := bisectThreshold(calcFn, targetLon, asp.Angle, orb, prevJD, jd, false)
					eLon, _, _ := calcFn(leaveJD)
					eRetro := getSpeed(calcFn, leaveJD) < 0
					events = append(events, models.TransitEvent{
						EventType:          models.EventAspectLeave,
						ChartType:          chartType,
						Planet:             planet,
						JD:                 leaveJD,
						Age:                progressions.Age(natalJD, leaveJD),
						PlanetLongitude:    eLon,
						PlanetSign:         models.SignFromLongitude(eLon),
						PlanetHouse:        chart.FindHouseForLongitude(eLon, natalHouses),
						IsRetrograde:       eRetro,
						TargetChartType:    targetChartType,
						Target:             targetID,
						TargetLongitude:    targetLon,
						TargetSign:         models.SignFromLongitude(targetLon),
						TargetHouse:        chart.FindHouseForLongitude(targetLon, natalHouses),
						TargetIsRetrograde: false,
						AspectType:         asp.Type,
						AspectAngle:        asp.Angle,
						OrbAtLeave:         math.Abs(angleDiffToAspect(eLon, targetLon, asp.Angle)),
					})
					inAspect = false
				}

				prevJD = jd
				prevDiff = curDiff

				if jd >= interval.End {
					break
				}
			}
		}
	}

	return events
}

// intersectIntervals returns the intersection of two sorted interval sets
func intersectIntervals(a, b []MonoInterval) []MonoInterval {
	var result []MonoInterval
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		lo := math.Max(a[i].Start, b[j].Start)
		hi := math.Min(a[i].End, b[j].End)
		if lo < hi {
			result = append(result, MonoInterval{Start: lo, End: hi})
		}
		if a[i].End < b[j].End {
			i++
		} else {
			j++
		}
	}
	return result
}

// =====================================================================
// RQ2: moving body vs moving body
// Per plan: take intersection of both bodies' monotonic intervals,
// then scan within each sub-interval where relative motion is monotonic.
// =====================================================================

func findAspectEventsRQ2(
	calcFn1 bodyCalcFunc, planet1 models.PlanetID, chartType1 models.ChartType,
	calcFn2 bodyCalcFunc, planet2 models.PlanetID, chartType2 models.ChartType,
	startJD, endJD float64,
	orbs models.OrbConfig,
	natalHouses []float64, natalJD float64,
) []models.TransitEvent {
	var events []models.TransitEvent

	// Build monotonic intervals for each body, then intersect
	stations1 := findStations(calcFn1, startJD, endJD, planet1)
	stations2 := findStations(calcFn2, startJD, endJD, planet2)
	intervals1 := buildMonoIntervals(startJD, endJD, stations1)
	intervals2 := buildMonoIntervals(startJD, endJD, stations2)
	subIntervals := intersectIntervals(intervals1, intervals2)
	if len(subIntervals) == 0 && len(intervals1) > 0 && len(intervals2) > 0 {
		// Fallback should not normally happen when both interval sets are non-empty
		subIntervals = []MonoInterval{{Start: startJD, End: endJD}}
	}
	if len(subIntervals) == 0 {
		return events // no intervals to scan
	}

	for _, asp := range models.StandardAspects {
		orb := orbs.GetOrb(asp.Type)
		if orb == 0 {
			continue
		}

		inAspect := false
		exactCount := 0

		// Check initial state and emit Begin event if already in aspect
		lon1Init, _, _ := calcFn1(startJD)
		lon2Init, speed2Init, _ := calcFn2(startJD)
		initDiff := angleDiffToAspect(lon1Init, lon2Init, asp.Angle)
		if math.Abs(initDiff) <= orb {
			inAspect = true
			initRetro1 := getSpeed(calcFn1, startJD) < 0
			initRetro2 := speed2Init < 0
			events = append(events, models.TransitEvent{
				EventType:          models.EventAspectBegin,
				ChartType:          chartType1,
				Planet:             planet1,
				JD:                 startJD,
				Age:                progressions.Age(natalJD, startJD),
				PlanetLongitude:    lon1Init,
				PlanetSign:         models.SignFromLongitude(lon1Init),
				PlanetHouse:        chart.FindHouseForLongitude(lon1Init, natalHouses),
				IsRetrograde:       initRetro1,
				TargetChartType:    chartType2,
				Target:             string(planet2),
				TargetLongitude:    lon2Init,
				TargetSign:         models.SignFromLongitude(lon2Init),
				TargetHouse:        chart.FindHouseForLongitude(lon2Init, natalHouses),
				TargetIsRetrograde: initRetro2,
				AspectType:         asp.Type,
				AspectAngle:        asp.Angle,
			})
		}

		for _, interval := range subIntervals {
			prevJD := interval.Start
			lon1Start, speed1Start, _ := calcFn1(prevJD)
			lon2Start, speed2Start, _ := calcFn2(prevJD)
			prevDiff := angleDiffToAspect(lon1Start, lon2Start, asp.Angle)

			// Compute step per-interval based on actual speeds
			step := adaptiveStep(speed1Start, speed2Start)

		for jd := interval.Start + step; jd <= interval.End+step*0.5; jd += step {
			if jd > interval.End {
				jd = interval.End
			}

			lon1, _, _ := calcFn1(jd)
			lon2, _, _ := calcFn2(jd)
			curDiff := angleDiffToAspect(lon1, lon2, asp.Angle)

			// ENTER
			if !inAspect && math.Abs(curDiff) <= orb && math.Abs(prevDiff) > orb {
				enterJD := bisectThresholdRQ2(calcFn1, calcFn2, asp.Angle, orb, prevJD, jd, true)
				eLon1, _, _ := calcFn1(enterJD)
				eLon2, eSpeed2, _ := calcFn2(enterJD)
				eRetro1 := getSpeed(calcFn1, enterJD) < 0
				eRetro2 := eSpeed2 < 0
				events = append(events, models.TransitEvent{
					EventType:          models.EventAspectEnter,
					ChartType:          chartType1,
					Planet:             planet1,
					JD:                 enterJD,
					Age:                progressions.Age(natalJD, enterJD),
					PlanetLongitude:    eLon1,
					PlanetSign:         models.SignFromLongitude(eLon1),
					PlanetHouse:        chart.FindHouseForLongitude(eLon1, natalHouses),
					IsRetrograde:       eRetro1,
					TargetChartType:    chartType2,
					Target:             string(planet2),
					TargetLongitude:    eLon2,
					TargetSign:         models.SignFromLongitude(eLon2),
					TargetHouse:        chart.FindHouseForLongitude(eLon2, natalHouses),
					TargetIsRetrograde: eRetro2,
					AspectType:         asp.Type,
					AspectAngle:        asp.Angle,
					OrbAtEnter:         math.Abs(angleDiffToAspect(eLon1, eLon2, asp.Angle)),
				})
				inAspect = true
			}

			// EXACT: sign change AND both diffs within reasonable range
			if prevDiff*curDiff < 0 && math.Abs(prevDiff) < 90 && math.Abs(curDiff) < 90 {
				exactJD := bisectExactRQ2(calcFn1, calcFn2, asp.Angle, prevJD, jd)
				exactCount++
				eLon1, _, _ := calcFn1(exactJD)
				eLon2, eSpeed2, _ := calcFn2(exactJD)
				eRetro1 := getSpeed(calcFn1, exactJD) < 0
				eRetro2 := eSpeed2 < 0
				events = append(events, models.TransitEvent{
					EventType:          models.EventAspectExact,
					ChartType:          chartType1,
					Planet:             planet1,
					JD:                 exactJD,
					Age:                progressions.Age(natalJD, exactJD),
					PlanetLongitude:    eLon1,
					PlanetSign:         models.SignFromLongitude(eLon1),
					PlanetHouse:        chart.FindHouseForLongitude(eLon1, natalHouses),
					IsRetrograde:       eRetro1,
					TargetChartType:    chartType2,
					Target:             string(planet2),
					TargetLongitude:    eLon2,
					TargetSign:         models.SignFromLongitude(eLon2),
					TargetHouse:        chart.FindHouseForLongitude(eLon2, natalHouses),
					TargetIsRetrograde: eRetro2,
					AspectType:         asp.Type,
					AspectAngle:        asp.Angle,
					ExactCount:         exactCount,
				})
			}

			// LEAVE
			if inAspect && math.Abs(curDiff) > orb && math.Abs(prevDiff) <= orb {
				leaveJD := bisectThresholdRQ2(calcFn1, calcFn2, asp.Angle, orb, prevJD, jd, false)
				eLon1, _, _ := calcFn1(leaveJD)
				eLon2, eSpeed2, _ := calcFn2(leaveJD)
				eRetro1 := getSpeed(calcFn1, leaveJD) < 0
				eRetro2 := eSpeed2 < 0
				events = append(events, models.TransitEvent{
					EventType:          models.EventAspectLeave,
					ChartType:          chartType1,
					Planet:             planet1,
					JD:                 leaveJD,
					Age:                progressions.Age(natalJD, leaveJD),
					PlanetLongitude:    eLon1,
					PlanetSign:         models.SignFromLongitude(eLon1),
					PlanetHouse:        chart.FindHouseForLongitude(eLon1, natalHouses),
					IsRetrograde:       eRetro1,
					TargetChartType:    chartType2,
					Target:             string(planet2),
					TargetLongitude:    eLon2,
					TargetSign:         models.SignFromLongitude(eLon2),
					TargetHouse:        chart.FindHouseForLongitude(eLon2, natalHouses),
					TargetIsRetrograde: eRetro2,
					AspectType:         asp.Type,
					AspectAngle:        asp.Angle,
					OrbAtLeave:         math.Abs(angleDiffToAspect(eLon1, eLon2, asp.Angle)),
				})
				inAspect = false
			}

			prevJD = jd
			prevDiff = curDiff

			if jd >= interval.End {
				break
			}
		}
		} // end subIntervals loop
	}

	return events
}

// =====================================================================
// Sign and house ingress
// =====================================================================

func findSignIngressEvents(calcFn bodyCalcFunc, planet models.PlanetID, chartType models.ChartType,
	intervals []MonoInterval, natalHouses []float64, natalJD float64) []models.TransitEvent {
	var events []models.TransitEvent

	for _, interval := range intervals {
		prevJD := interval.Start
		prevLon, prevSpeed, _ := calcFn(prevJD)
		prevSign := int(prevLon / 30.0)

		step := adaptiveStep(prevSpeed, 0)
		for jd := interval.Start + step; jd <= interval.End+step*0.5; jd += step {
			if jd > interval.End {
				jd = interval.End
			}
			curLon, _, _ := calcFn(jd)
			curSign := int(curLon / 30.0)

			if curSign != prevSign {
				crossJD := bisectSignBoundary(calcFn, prevJD, jd, prevSign)
				cLon, _, _ := calcFn(crossJD)
				cRetro := getSpeed(calcFn, crossJD) < 0

				fromSign := models.ZodiacSigns[prevSign%12]
				toSign := models.ZodiacSigns[curSign%12]

				events = append(events, models.TransitEvent{
					EventType:       models.EventSignIngress,
					ChartType:       chartType,
					Planet:          planet,
					JD:              crossJD,
					Age:             progressions.Age(natalJD, crossJD),
					PlanetLongitude: cLon,
					PlanetSign:      toSign,
					PlanetHouse:     chart.FindHouseForLongitude(cLon, natalHouses),
					IsRetrograde:    cRetro,
					FromSign:        fromSign,
					ToSign:          toSign,
				})
			}

			prevJD = jd
			prevSign = curSign

			if jd >= interval.End {
				break
			}
		}
	}
	return events
}

func findHouseIngressEvents(calcFn bodyCalcFunc, planet models.PlanetID, chartType models.ChartType,
	intervals []MonoInterval, natalHouses []float64, natalJD float64) []models.TransitEvent {
	var events []models.TransitEvent
	if len(natalHouses) < 12 {
		return events
	}

	for _, interval := range intervals {
		prevJD := interval.Start
		prevLon, prevSpeed, _ := calcFn(prevJD)
		prevHouse := chart.FindHouseForLongitude(prevLon, natalHouses)

		step := adaptiveStep(prevSpeed, 0)
		for jd := interval.Start + step; jd <= interval.End+step*0.5; jd += step {
			if jd > interval.End {
				jd = interval.End
			}
			curLon, _, _ := calcFn(jd)
			curHouse := chart.FindHouseForLongitude(curLon, natalHouses)

			if curHouse != prevHouse {
				crossJD := bisectHouseBoundary(calcFn, prevJD, jd, natalHouses, prevHouse)
				cLon, _, _ := calcFn(crossJD)
				cRetro := getSpeed(calcFn, crossJD) < 0

				events = append(events, models.TransitEvent{
					EventType:       models.EventHouseIngress,
					ChartType:       chartType,
					Planet:          planet,
					JD:              crossJD,
					Age:             progressions.Age(natalJD, crossJD),
					PlanetLongitude: cLon,
					PlanetSign:      models.SignFromLongitude(cLon),
					PlanetHouse:     curHouse,
					IsRetrograde:    cRetro,
					FromHouse:       prevHouse,
					ToHouse:         curHouse,
				})
			}

			prevJD = jd
			prevHouse = curHouse

			if jd >= interval.End {
				break
			}
		}
	}
	return events
}

// =====================================================================
// Void of Course Moon
// =====================================================================

// findVoidOfCourse derives void-of-course periods from already computed events.
// VOC = period between Moon's last aspect LEAVE and next SIGN_INGRESS.
// We consider all Moon aspect events: Tr-Na and Tr-Tr (also Tr-Sp, Tr-Sa if present).
// Additionally, ASPECT_EXACT events reset the VOC (Moon is still engaged).
func findVoidOfCourse(events []models.TransitEvent, startJD, endJD float64) []models.TransitEvent {
	var vocEvents []models.TransitEvent

	// Collect Moon aspect leaves, exacts, enters, and sign ingresses, sorted by JD
	type moonEvent struct {
		JD        float64
		IsLeave   bool
		IsExact   bool
		IsEnter   bool
		IsIngress bool
		Event     models.TransitEvent
	}

	var moonEvts []moonEvent
	for _, e := range events {
		// Moon as the moving body (planet field)
		isMoonMoving := e.Planet == models.PlanetMoon && e.ChartType == models.ChartTransit
		// Moon as the target body (in Tr-Tr events where another planet aspects Moon)
		isMoonTarget := e.Target == string(models.PlanetMoon) && e.TargetChartType == models.ChartTransit

		if !isMoonMoving && !isMoonTarget {
			continue
		}

		switch e.EventType {
		case models.EventAspectLeave:
			moonEvts = append(moonEvts, moonEvent{JD: e.JD, IsLeave: true, Event: e})
		case models.EventAspectExact:
			moonEvts = append(moonEvts, moonEvent{JD: e.JD, IsExact: true, Event: e})
		case models.EventAspectEnter:
			moonEvts = append(moonEvts, moonEvent{JD: e.JD, IsEnter: true, Event: e})
		case models.EventSignIngress:
			if isMoonMoving { // Only Moon's own sign ingress matters
				moonEvts = append(moonEvts, moonEvent{JD: e.JD, IsIngress: true, Event: e})
			}
		}
	}

	sort.Slice(moonEvts, func(i, j int) bool {
		return moonEvts[i].JD < moonEvts[j].JD
	})

	// Walk through events chronologically.
	// VOC starts at the last aspect LEAVE (of any Tr-Tr aspect) before each sign ingress.
	// This matches Solar Fire's VOC definition: the last major aspect the Moon makes
	// before entering the next sign.
	for i := 0; i < len(moonEvts); i++ {
		if !moonEvts[i].IsIngress {
			continue
		}
		ingressEvt := moonEvts[i]

		// Find the last LEAVE before this ingress (scanning backward)
		var lastLeave *moonEvent
		for j := i - 1; j >= 0; j-- {
			if moonEvts[j].IsIngress {
				break // Previous sign change — stop looking
			}
			if moonEvts[j].IsLeave {
				evt := moonEvts[j]
				lastLeave = &evt
				break
			}
		}

		if lastLeave != nil {
			// Determine the aspect target — for Tr-Tr where Moon is the target,
			// use Planet as the target display
			lastAspectType := string(lastLeave.Event.AspectType)
			lastAspectTarget := lastLeave.Event.Target
			planetLon := lastLeave.Event.PlanetLongitude
			planetSign := lastLeave.Event.PlanetSign
			planetHouse := lastLeave.Event.PlanetHouse
			if lastLeave.Event.Planet != models.PlanetMoon {
				// Moon was the target, not the moving planet
				lastAspectTarget = string(lastLeave.Event.Planet)
				planetLon = lastLeave.Event.TargetLongitude
				planetSign = lastLeave.Event.TargetSign
				planetHouse = lastLeave.Event.TargetHouse
			}

			vocEvents = append(vocEvents, models.TransitEvent{
				EventType:        models.EventVoidOfCourse,
				ChartType:        models.ChartTransit,
				Planet:           models.PlanetMoon,
				JD:               lastLeave.JD,
				PlanetLongitude:  planetLon,
				PlanetSign:       planetSign,
				PlanetHouse:      planetHouse,
				IsRetrograde:     false,
				VoidStartJD:      lastLeave.JD,
				VoidEndJD:        ingressEvt.JD,
				LastAspectType:   lastAspectType,
				LastAspectTarget: lastAspectTarget,
				NextSign:         ingressEvt.Event.ToSign,
			})
		}
	}

	return vocEvents
}

// =====================================================================
// Math helpers
// =====================================================================

// angleDiffToAspect returns the signed angular difference between actual separation and target aspect.
// For conjunction (0) and opposition (180), uses a signed difference that properly crosses zero.
// For other aspects, uses shortestAngle which avoids false zero-crossings.
// Result is in [-180, 180), where 0 means exact aspect.
func angleDiffToAspect(lon1, lon2, aspectAngle float64) float64 {
	if aspectAngle == 0 {
		// Conjunction: signed diff crosses zero when bodies are conjunct
		return wrapAngle(lon1 - lon2)
	}
	if aspectAngle == 180 {
		// Opposition: signed diff crosses zero when bodies are in opposition
		return wrapAngle(lon1 - lon2 - 180)
	}
	// Other aspects: use unsigned shortestAngle (works correctly for 0 < angle < 180)
	actualAngle := shortestAngle(lon1, lon2)
	return wrapAngle(actualAngle - aspectAngle)
}

// directedAngleDiff implements the plan's direction-aware normalization:
// Δθ = wrap((θ₁ - θ₂) × sgn(d), -180°, 180°)
// This properly distinguishes applying vs separating during retrograde.
// sgn(d) = +1 for direct, -1 for retrograde
func directedAngleDiff(lon1, lon2, aspectAngle float64, speed float64) float64 {
	sgn := 1.0
	if speed < 0 {
		sgn = -1.0
	}
	rawDiff := shortestAngle(lon1, lon2) - aspectAngle
	return wrapAngle(rawDiff * sgn)
}

func shortestAngle(lon1, lon2 float64) float64 {
	diff := math.Abs(lon1 - lon2)
	if diff > 180 {
		diff = 360 - diff
	}
	return diff
}

func wrapAngle(a float64) float64 {
	a = math.Mod(a+180, 360)
	if a < 0 {
		a += 360
	}
	return a - 180
}

// =====================================================================
// Bisection helpers
// =====================================================================

func bisectThreshold(calcFn bodyCalcFunc, targetLon, aspectAngle, orb, lo, hi float64, entering bool) float64 {
	for hi-lo > bisectEps {
		mid := (lo + hi) / 2
		midLon, _, _ := calcFn(mid)
		midDiff := math.Abs(angleDiffToAspect(midLon, targetLon, aspectAngle))

		if entering {
			if midDiff > orb {
				lo = mid
			} else {
				hi = mid
			}
		} else {
			if midDiff <= orb {
				lo = mid
			} else {
				hi = mid
			}
		}
	}
	return (lo + hi) / 2
}

func bisectExact(calcFn bodyCalcFunc, targetLon, aspectAngle, lo, hi float64) float64 {
	loLon, _, _ := calcFn(lo)
	loDiff := angleDiffToAspect(loLon, targetLon, aspectAngle)

	for hi-lo > bisectEps {
		mid := (lo + hi) / 2
		midLon, _, _ := calcFn(mid)
		midDiff := angleDiffToAspect(midLon, targetLon, aspectAngle)

		if loDiff*midDiff <= 0 {
			hi = mid
		} else {
			lo = mid
			loDiff = midDiff
		}
	}
	return (lo + hi) / 2
}

func bisectThresholdRQ2(calcFn1, calcFn2 bodyCalcFunc, aspectAngle, orb, lo, hi float64, entering bool) float64 {
	for hi-lo > bisectEps {
		mid := (lo + hi) / 2
		lon1, _, _ := calcFn1(mid)
		lon2, _, _ := calcFn2(mid)
		midDiff := math.Abs(angleDiffToAspect(lon1, lon2, aspectAngle))

		if entering {
			if midDiff > orb {
				lo = mid
			} else {
				hi = mid
			}
		} else {
			if midDiff <= orb {
				lo = mid
			} else {
				hi = mid
			}
		}
	}
	return (lo + hi) / 2
}

func bisectExactRQ2(calcFn1, calcFn2 bodyCalcFunc, aspectAngle, lo, hi float64) float64 {
	lon1Lo, _, _ := calcFn1(lo)
	lon2Lo, _, _ := calcFn2(lo)
	loDiff := angleDiffToAspect(lon1Lo, lon2Lo, aspectAngle)

	for hi-lo > bisectEps {
		mid := (lo + hi) / 2
		lon1Mid, _, _ := calcFn1(mid)
		lon2Mid, _, _ := calcFn2(mid)
		midDiff := angleDiffToAspect(lon1Mid, lon2Mid, aspectAngle)

		if loDiff*midDiff <= 0 {
			hi = mid
		} else {
			lo = mid
			loDiff = midDiff
		}
	}
	return (lo + hi) / 2
}

func bisectSignBoundary(calcFn bodyCalcFunc, lo, hi float64, prevSign int) float64 {
	for hi-lo > bisectEps {
		mid := (lo + hi) / 2
		midLon, _, _ := calcFn(mid)
		midSign := int(midLon / 30.0)
		if midSign == prevSign {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

func bisectHouseBoundary(calcFn bodyCalcFunc, lo, hi float64, cusps []float64, prevHouse int) float64 {
	for hi-lo > bisectEps {
		mid := (lo + hi) / 2
		midLon, _, _ := calcFn(mid)
		midHouse := chart.FindHouseForLongitude(midLon, cusps)
		if midHouse == prevHouse {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}
