package transit

import (
	"math"
	"sort"

	"github.com/shaobaobaoer/solarsage-mcp/pkg/chart"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/progressions"
)

const (
	dayStep   = 1.0         // 1 day step for station scanning
	bisectEps = 1.0 / 86400 // ~1 second precision
)

// adaptiveStep returns an appropriate scan step based on body speeds and orb.
// The step must be small enough to detect sign changes in the angular difference.
//
// Planet speed reference (degrees/day):
//   Moon ~13, Sun ~1, Mercury ~0.5-1.5, Venus ~0.8-1.2, Mars ~0.5-0.8
//   Jupiter ~0.08, Saturn ~0.03, Uranus ~0.01, Neptune ~0.006, Pluto ~0.004
//   Progressions ~0.003, Solar arc ~0.003
//
// For RQ1 (moving vs fixed): step based on moving body speed
// For RQ2 (moving vs moving): step based on relative speed
func adaptiveStep(speed1, speed2 float64, orb float64) float64 {
	// Relative speed determines how fast the angular difference changes
	relSpeed := math.Abs(speed1) + math.Abs(speed2)

	// Ensure we get at least 4 samples per orb crossing
	// step = orb / relSpeed / 4 (capped)
	if relSpeed > 0.001 {
		step := orb / relSpeed / 4.0
		// Clamp to reasonable range
		if step < 0.1 {
			step = 0.1 // minimum 2.4 hours
		}
		if step > 5.0 {
			step = 5.0 // maximum 5 days
		}
		return step
	}

	// Ultra-slow: both bodies nearly stationary (near station)
	// Use 1 day to catch any subtle motion
	return 1.0
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
	// New structured fields (for refactored code)
	NatalChart  NatalChartConfig
	TimeRange   TimeRangeConfig
	Charts      ChartSetConfig
	EventFilter EventFilterConfig

	// Old flat fields (for backward compatibility with tests)
	NatalLat     float64
	NatalLon     float64
	NatalJD      float64
	NatalPlanets []models.PlanetID

	// NatalASC and NatalMC allow overriding the calculated natal angles.
	// If non-zero, these values are used instead of computing from NatalJD/Lat/Lon.
	// This is useful when matching reference data (e.g., Solar Fire) that uses
	// slightly different obliquity or house calculation parameters.
	NatalASC float64
	NatalMC  float64
	// NatalMCForASC is a separate override for progressed ASC calculation.
	// Solar Fire uses different MC base for ASC derivation than for MC progression.
	// Set to -1 to force using sweph-computed MC for ASC (even if NatalMC is set).
	NatalMCForASC float64
	// NatalASCForProgressions: if > 0, use direct solar arc method for progressed ASC.
	// progASC = NatalASCForProgressions + solarArc
	// This matches Solar Fire's behavior exactly.
	NatalASCForProgressions float64
	// NatalPlanetOverrides allows specifying exact natal planet positions.
	// Use this to match reference data that may use different ephemeris (DE200 vs DE432).
	// Map key is planet ID (e.g., "MOON", "MERCURY"), value is longitude in degrees.
	NatalPlanetOverrides map[string]float64

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

// posAt returns planet position info at a given JD
func posAt(calcFn bodyCalcFunc, jd float64, natalHouses []float64) (lon float64, sign string, house int, retro bool) {
	lon, _, _ = calcFn(jd)
	retro = getSpeed(calcFn, jd) < 0
	sign = models.SignFromLongitude(lon)
	house = chart.FindHouseForLongitude(lon, natalHouses)
	return
}

// makeStationEvent creates a station event
func makeStationEvent(calcFn bodyCalcFunc, planet models.PlanetID, chartType models.ChartType,
	st StationInfo, natalHouses []float64, natalJD float64) models.TransitEvent {
	lon, sign, house, _ := posAt(calcFn, st.JD, natalHouses)
	stType := models.StationRetrograde
	if st.IsDirecting {
		stType = models.StationDirect
	}
	return models.TransitEvent{
		EventType:       models.EventStation,
		ChartType:       chartType,
		Planet:          planet,
		JD:              st.JD,
		Age:             progressions.Age(natalJD, st.JD),
		PlanetLongitude: lon,
		PlanetSign:      sign,
		PlanetHouse:     house,
		IsRetrograde:    stType == models.StationRetrograde,
		StationType:     stType,
	}
}

// rq1Scan holds fixed parameters for RQ1 (moving body vs fixed point) aspect scanning
type rq1Scan struct {
	calcFn      bodyCalcFunc
	planet      models.PlanetID
	chartType   models.ChartType
	targetID    string
	targetLon   float64
	targetCT    models.ChartType
	natalHouses []float64
	natalJD     float64
}

// event creates a TransitEvent for an RQ1 aspect event
func (s *rq1Scan) event(eventType models.EventType, jd float64, asp models.AspectDef) models.TransitEvent {
	lon, sign, house, retro := posAt(s.calcFn, jd, s.natalHouses)
	tSign := models.SignFromLongitude(s.targetLon)
	tHouse := chart.FindHouseForLongitude(s.targetLon, s.natalHouses)
	return models.TransitEvent{
		EventType:       eventType,
		ChartType:       s.chartType,
		Planet:          s.planet,
		JD:              jd,
		Age:             progressions.Age(s.natalJD, jd),
		PlanetLongitude: lon,
		PlanetSign:      sign,
		PlanetHouse:     house,
		IsRetrograde:    retro,
		TargetChartType: s.targetCT,
		Target:          s.targetID,
		TargetLongitude: s.targetLon,
		TargetSign:      tSign,
		TargetHouse:     tHouse,
		AspectType:      asp.Type,
		AspectAngle:     asp.Angle,
	}
}

// rq2Scan holds fixed parameters for RQ2 (moving body vs moving body) aspect scanning
type rq2Scan struct {
	calcFn1    bodyCalcFunc
	planet1    models.PlanetID
	chartType1 models.ChartType
	calcFn2    bodyCalcFunc
	planet2    models.PlanetID
	chartType2 models.ChartType
	natalHouses []float64
	natalJD     float64
}

// event creates a TransitEvent for an RQ2 aspect event
func (s *rq2Scan) event(eventType models.EventType, jd float64, asp models.AspectDef) models.TransitEvent {
	lon1, sign1, house1, retro1 := posAt(s.calcFn1, jd, s.natalHouses)
	lon2, speed2, _ := s.calcFn2(jd)
	return models.TransitEvent{
		EventType:          eventType,
		ChartType:          s.chartType1,
		Planet:             s.planet1,
		JD:                 jd,
		Age:                progressions.Age(s.natalJD, jd),
		PlanetLongitude:    lon1,
		PlanetSign:         sign1,
		PlanetHouse:        house1,
		IsRetrograde:       retro1,
		TargetChartType:    s.chartType2,
		Target:             string(s.planet2),
		TargetLongitude:    lon2,
		TargetSign:         models.SignFromLongitude(lon2),
		TargetHouse:        chart.FindHouseForLongitude(lon2, s.natalHouses),
		TargetIsRetrograde: speed2 < 0,
		AspectType:         asp.Type,
		AspectAngle:        asp.Angle,
	}
}
func wrapDelta(d float64) float64 {
	if d > 180 {
		return d - 360
	}
	if d < -180 {
		return d + 360
	}
	return d
}

// numericalSpeed estimates speed via forward finite difference, handling wrap-around
func numericalSpeed(lonFn func(float64) (float64, error), jd, dt float64) float64 {
	lon1, err1 := lonFn(jd)
	lon2, err2 := lonFn(jd + dt)
	if err1 != nil || err2 != nil {
		return 0
	}
	return wrapDelta(lon2-lon1) / dt
}



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

func findSignIngressEvents(calcFn bodyCalcFunc, planet models.PlanetID, chartType models.ChartType,
	intervals []MonoInterval, natalHouses []float64, natalJD float64) []models.TransitEvent {
	var events []models.TransitEvent

	for _, interval := range intervals {
		prevJD := interval.Start
		prevLon, prevSpeed, _ := calcFn(prevJD)
		prevSign := int(prevLon / 30.0)

		step := adaptiveStep(prevSpeed, 0, 30.0)
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

		step := adaptiveStep(prevSpeed, 0, 30.0)
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
	// Standard VOC definition: the last major aspect the Moon makes
	// before entering the next sign.
	for i := 0; i < len(moonEvts); i++ {
		if !moonEvts[i].IsIngress {
			continue
		}
		ingressEvt := moonEvts[i]

		// Find the last aspect event before this ingress (scanning backward)
		// Priority: Leave > Exact > Enter (uses the last aspect event)
		var lastLeave *moonEvent
		for j := i - 1; j >= 0; j-- {
			if moonEvts[j].IsIngress {
				break // Previous sign change — stop looking
			}
			if moonEvts[j].IsLeave || moonEvts[j].IsExact || moonEvts[j].IsEnter {
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

func angleDiffToAspect(lon1, lon2, aspectAngle float64) float64 {
	if aspectAngle == 0 {
		return wrapAngle(lon1 - lon2)
	}
	if aspectAngle == 180 {
		return wrapAngle(lon1 - lon2 - 180)
	}
	actualAngle := shortestAngle(lon1, lon2)
	return wrapAngle(actualAngle - aspectAngle)
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

func findAspectEventsRQ1(
	calcFn bodyCalcFunc, planet models.PlanetID, chartType models.ChartType,
	targetID string, targetLon float64, targetChartType models.ChartType,
	intervals []MonoInterval, orbs models.OrbConfig,
	natalHouses []float64, exactCounters map[string]int,
	natalJD float64,
) []models.TransitEvent {
	var events []models.TransitEvent

	scan := &rq1Scan{
		calcFn: calcFn, planet: planet, chartType: chartType,
		targetID: targetID, targetLon: targetLon, targetCT: targetChartType,
		natalHouses: natalHouses, natalJD: natalJD,
	}

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
				events = append(events, scan.event(models.EventAspectBegin, intervals[0].Start, asp))
			}
		}

		for _, interval := range intervals {
			prevJD := interval.Start
			prevLon, prevSpeed, _ := calcFn(prevJD)
			prevDiff := angleDiffToAspect(prevLon, targetLon, asp.Angle)
			step := adaptiveStep(prevSpeed, 0, orb)

			for jd := interval.Start + step; jd <= interval.End+step*0.5; jd += step {
				if jd > interval.End {
					jd = interval.End
				}
				curLon, _, _ := calcFn(jd)
				curDiff := angleDiffToAspect(curLon, targetLon, asp.Angle)

				// ENTER
				if !inAspect && math.Abs(curDiff) <= orb && math.Abs(prevDiff) > orb {
					enterJD := bisectThreshold(calcFn, targetLon, asp.Angle, orb, prevJD, jd, true)
					e := scan.event(models.EventAspectEnter, enterJD, asp)
					e.OrbAtEnter = math.Abs(angleDiffToAspect(e.PlanetLongitude, targetLon, asp.Angle))
					events = append(events, e)
					inAspect = true
				}

				// EXACT: sign change AND both diffs are within a reasonable range
				if prevDiff*curDiff < 0 && math.Abs(prevDiff) < 90 && math.Abs(curDiff) < 90 {
					exactJD := bisectExact(calcFn, targetLon, asp.Angle, prevJD, jd)
					exactCounters[counterKey]++
					e := scan.event(models.EventAspectExact, exactJD, asp)
					e.ExactCount = exactCounters[counterKey]
					events = append(events, e)
				}

				// LEAVE
				if inAspect && math.Abs(curDiff) > orb && math.Abs(prevDiff) <= orb {
					leaveJD := bisectThreshold(calcFn, targetLon, asp.Angle, orb, prevJD, jd, false)
					e := scan.event(models.EventAspectLeave, leaveJD, asp)
					e.OrbAtLeave = math.Abs(angleDiffToAspect(e.PlanetLongitude, targetLon, asp.Angle))
					events = append(events, e)
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

func findAspectEventsRQ2(
	calcFn1 bodyCalcFunc, planet1 models.PlanetID, chartType1 models.ChartType,
	calcFn2 bodyCalcFunc, planet2 models.PlanetID, chartType2 models.ChartType,
	startJD, endJD float64,
	orbs models.OrbConfig,
	natalHouses []float64, natalJD float64,
) []models.TransitEvent {
	var events []models.TransitEvent

	scan := &rq2Scan{
		calcFn1: calcFn1, planet1: planet1, chartType1: chartType1,
		calcFn2: calcFn2, planet2: planet2, chartType2: chartType2,
		natalHouses: natalHouses, natalJD: natalJD,
	}

	// Build monotonic intervals for each body, then intersect
	stations1 := findStations(calcFn1, startJD, endJD, planet1)
	stations2 := findStations(calcFn2, startJD, endJD, planet2)
	intervals1 := buildMonoIntervals(startJD, endJD, stations1)
	intervals2 := buildMonoIntervals(startJD, endJD, stations2)
	subIntervals := intersectIntervals(intervals1, intervals2)
	if len(subIntervals) == 0 && len(intervals1) > 0 && len(intervals2) > 0 {
		subIntervals = []MonoInterval{{Start: startJD, End: endJD}}
	}
	if len(subIntervals) == 0 {
		return events
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
		lon2Init, _, _ := calcFn2(startJD)
		initDiff := angleDiffToAspect(lon1Init, lon2Init, asp.Angle)
		if math.Abs(initDiff) <= orb {
			inAspect = true
			events = append(events, scan.event(models.EventAspectBegin, startJD, asp))
		}

		// prevIntervalEndDiff tracks the diff at the end of the previous sub-interval,
		// used to detect true zero-crossings at station boundaries.
		prevIntervalEndDiff := initDiff

		for _, interval := range subIntervals {
			prevJD := interval.Start
			lon1Start, speed1Start, _ := calcFn1(prevJD)
			lon2Start, speed2Start, _ := calcFn2(prevJD)
			prevDiff := angleDiffToAspect(lon1Start, lon2Start, asp.Angle)

			// Check if aspect is near-exact at sub-interval start (station boundary).
			// Only emit exact if the diff crossed zero at the station (opposite sign from
			// end of previous interval), preventing spurious events when the planet merely
			// grazes the orb boundary at its station without a true zero-crossing.
			if inAspect && math.Abs(prevDiff) < 0.01 && prevJD > startJD+0.5 && prevIntervalEndDiff*prevDiff < 0 {
				exactCount++
				e := scan.event(models.EventAspectExact, prevJD, asp)
				e.ExactCount = exactCount
				events = append(events, e)
			}

			step := adaptiveStep(speed1Start, speed2Start, orb)

		for jd := interval.Start + step; jd <= interval.End+step*0.5; jd += step {
			if jd > interval.End {
				jd = interval.End
			}

			lon1, speed1Cur, _ := calcFn1(jd)
			lon2, speed2Cur, _ := calcFn2(jd)
			curDiff := angleDiffToAspect(lon1, lon2, asp.Angle)

			step = adaptiveStep(speed1Cur, speed2Cur, orb)

			// ENTER
			if !inAspect && math.Abs(curDiff) <= orb && math.Abs(prevDiff) > orb {
				enterJD := bisectThresholdRQ2(calcFn1, calcFn2, asp.Angle, orb, prevJD, jd, true)
				e := scan.event(models.EventAspectEnter, enterJD, asp)
				e.OrbAtEnter = math.Abs(angleDiffToAspect(e.PlanetLongitude, e.TargetLongitude, asp.Angle))
				events = append(events, e)
				inAspect = true
			}

			// EXACT
			isSignChange := prevDiff*curDiff < 0 && math.Abs(prevDiff) < 90 && math.Abs(curDiff) < 90
			isNearZero := math.Abs(curDiff) < 0.001 && math.Abs(prevDiff) > 0.0001 && math.Abs(prevDiff) < 90
			if isSignChange || isNearZero {
				exactJD := bisectExactRQ2(calcFn1, calcFn2, asp.Angle, prevJD, jd)
				exactCount++
				e := scan.event(models.EventAspectExact, exactJD, asp)
				e.ExactCount = exactCount
				events = append(events, e)
			}

			// LEAVE
			if inAspect && math.Abs(curDiff) > orb && math.Abs(prevDiff) <= orb {
				leaveJD := bisectThresholdRQ2(calcFn1, calcFn2, asp.Angle, orb, prevJD, jd, false)
				e := scan.event(models.EventAspectLeave, leaveJD, asp)
				e.OrbAtLeave = math.Abs(angleDiffToAspect(e.PlanetLongitude, e.TargetLongitude, asp.Angle))
				events = append(events, e)
				inAspect = false
			}

			prevJD = jd
			prevDiff = curDiff

			if jd >= interval.End {
				break
			}
		}
		// Update prevIntervalEndDiff to the diff at the end of this interval,
		// so the next interval's station boundary check can detect true zero-crossings.
		prevIntervalEndDiff = prevDiff
		} // end subIntervals loop
	}

	return events
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

// directedAngleDiff implements the plan's direction-aware normalization:
// Δθ = wrap((θ₁ - θ₂) × sgn(d), -180°, 180°)
// This properly distinguishes applying vs separating during retrograde.
