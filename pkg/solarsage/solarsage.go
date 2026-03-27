// Package solarsage provides a high-level convenience API for common astrology
// calculations. It wraps the lower-level packages with sensible defaults,
// making typical operations simple one-liners.
//
// Initialize once with Init, then use the package-level functions:
//
//	solarsage.Init("/path/to/ephe")
//	defer solarsage.Close()
//
//	chart, _ := solarsage.NatalChart(51.5, -0.1, "2000-01-01T12:00:00Z")
//	phase, _ := solarsage.MoonPhase("2000-01-01T12:00:00Z")
//	sr, _ := solarsage.SolarReturn(51.5, -0.1, "2000-01-01T12:00:00Z", 2025)
package solarsage

import (
	"fmt"
	"strings"
	"time"

	"github.com/shaobaobaoer/solarsage-mcp/internal/aspect"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/chart"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/composite"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/dignity"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/firdaria"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/lunar"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/primary"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/render"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/report"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/returns"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/sweph"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/symbolic"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/synastry"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/transit"
)

// DefaultPlanets is the standard set of 10 planets used when none specified.
var DefaultPlanets = []models.PlanetID{
	models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
	models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
	models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune,
	models.PlanetPluto,
}

// Init initializes the Swiss Ephemeris with the given ephemeris data path.
func Init(ephePath string) {
	sweph.Init(ephePath)
}

// Close releases Swiss Ephemeris resources.
func Close() {
	sweph.Close()
}

// ValidateCoords checks that latitude and longitude are within valid ranges.
func ValidateCoords(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("latitude %.4f out of range [-90, 90]", lat)
	}
	if lon < -180 || lon > 180 {
		return fmt.Errorf("longitude %.4f out of range [-180, 180]", lon)
	}
	return nil
}

// NatalChart calculates a natal chart from a datetime string and coordinates.
// Accepts ISO 8601 datetime (e.g. "2000-01-01T12:00:00Z").
func NatalChart(lat, lon float64, datetime string) (*models.ChartInfo, error) {
	if err := ValidateCoords(lat, lon); err != nil {
		return nil, err
	}
	jd, err := ParseDatetime(datetime)
	if err != nil {
		return nil, fmt.Errorf("natal chart: %w", err)
	}
	return chart.CalcSingleChart(lat, lon, jd, DefaultPlanets,
		models.DefaultOrbConfig(), models.HousePlacidus)
}

// NatalChartFull calculates a natal chart with all available bodies.
func NatalChartFull(lat, lon float64, datetime string) (*models.ChartInfo, error) {
	jd, err := ParseDatetime(datetime)
	if err != nil {
		return nil, fmt.Errorf("natal chart: %w", err)
	}
	allPlanets := []models.PlanetID{
		models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
		models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
		models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune,
		models.PlanetPluto, models.PlanetChiron, models.PlanetNorthNodeTrue,
		models.PlanetSouthNode, models.PlanetLilithMean,
	}
	return chart.CalcSingleChart(lat, lon, jd, allPlanets,
		models.DefaultOrbConfig(), models.HousePlacidus)
}

// Transits searches for transit events over a date range.
func Transits(natalLat, natalLon float64, natalDatetime string, startDatetime, endDatetime string) ([]models.TransitEvent, error) {
	natalJD, err := ParseDatetime(natalDatetime)
	if err != nil {
		return nil, fmt.Errorf("natal datetime: %w", err)
	}
	startJD, err := ParseDatetime(startDatetime)
	if err != nil {
		return nil, fmt.Errorf("start datetime: %w", err)
	}
	endJD, err := ParseDatetime(endDatetime)
	if err != nil {
		return nil, fmt.Errorf("end datetime: %w", err)
	}

	return transit.CalcTransitEvents(transit.TransitCalcInput{
		NatalLat:          natalLat,
		NatalLon:          natalLon,
		NatalJD:           natalJD,
		NatalPlanets:      DefaultPlanets,
		TransitLat:        natalLat,
		TransitLon:        natalLon,
		StartJD:           startJD,
		EndJD:             endJD,
		TransitPlanets:    DefaultPlanets,
		EventConfig:       models.DefaultEventConfig(),
		OrbConfigTransit:  models.DefaultOrbConfig(),
		HouseSystem:       models.HousePlacidus,
	})
}

// SolarReturn calculates the solar return chart for a given year.
func SolarReturn(natalLat, natalLon float64, natalDatetime string, year int) (*returns.ReturnChart, error) {
	natalJD, err := ParseDatetime(natalDatetime)
	if err != nil {
		return nil, fmt.Errorf("natal datetime: %w", err)
	}

	// Calculate approximate search start: Jan 1 of the target year
	searchJD := float64(sweph.JulDay(year, 1, 1, 0, true))

	return returns.CalcSolarReturn(returns.ReturnInput{
		NatalJD:     natalJD,
		NatalLat:    natalLat,
		NatalLon:    natalLon,
		SearchJD:    searchJD,
		Planets:     DefaultPlanets,
		OrbConfig:   models.DefaultOrbConfig(),
		HouseSystem: models.HousePlacidus,
	})
}

// LunarReturn calculates the next lunar return after the given datetime.
func LunarReturn(natalLat, natalLon float64, natalDatetime, searchDatetime string) (*returns.ReturnChart, error) {
	natalJD, err := ParseDatetime(natalDatetime)
	if err != nil {
		return nil, fmt.Errorf("natal datetime: %w", err)
	}
	searchJD, err := ParseDatetime(searchDatetime)
	if err != nil {
		return nil, fmt.Errorf("search datetime: %w", err)
	}

	return returns.CalcLunarReturn(returns.ReturnInput{
		NatalJD:     natalJD,
		NatalLat:    natalLat,
		NatalLon:    natalLon,
		SearchJD:    searchJD,
		Planets:     DefaultPlanets,
		OrbConfig:   models.DefaultOrbConfig(),
		HouseSystem: models.HousePlacidus,
	})
}

// MoonPhase returns the lunar phase at a given datetime.
func MoonPhase(datetime string) (*lunar.PhaseInfo, error) {
	jd, err := ParseDatetime(datetime)
	if err != nil {
		return nil, fmt.Errorf("moon phase: %w", err)
	}
	return lunar.CalcLunarPhase(jd)
}

// Eclipses finds eclipses in a date range.
func Eclipses(startDatetime, endDatetime string) ([]lunar.EclipseInfo, error) {
	startJD, err := ParseDatetime(startDatetime)
	if err != nil {
		return nil, fmt.Errorf("start datetime: %w", err)
	}
	endJD, err := ParseDatetime(endDatetime)
	if err != nil {
		return nil, fmt.Errorf("end datetime: %w", err)
	}
	return lunar.FindEclipses(startJD, endJD)
}

// Compatibility calculates a synastry compatibility score between two people.
func Compatibility(lat1, lon1 float64, datetime1 string, lat2, lon2 float64, datetime2 string) (*synastry.SynastryScore, error) {
	jd1, err := ParseDatetime(datetime1)
	if err != nil {
		return nil, fmt.Errorf("person 1 datetime: %w", err)
	}
	jd2, err := ParseDatetime(datetime2)
	if err != nil {
		return nil, fmt.Errorf("person 2 datetime: %w", err)
	}

	orbs := models.DefaultOrbConfig()
	chart1, err := chart.CalcSingleChart(lat1, lon1, jd1, DefaultPlanets, orbs, models.HousePlacidus)
	if err != nil {
		return nil, fmt.Errorf("person 1 chart: %w", err)
	}
	chart2, err := chart.CalcSingleChart(lat2, lon2, jd2, DefaultPlanets, orbs, models.HousePlacidus)
	if err != nil {
		return nil, fmt.Errorf("person 2 chart: %w", err)
	}

	return synastry.CalcSynastryFromCharts(chart1.Planets, chart2.Planets, orbs), nil
}

// CompositeChart calculates a composite (midpoint) chart for two people.
func CompositeChart(lat1, lon1 float64, datetime1 string, lat2, lon2 float64, datetime2 string) (*composite.CompositeChart, error) {
	jd1, err := ParseDatetime(datetime1)
	if err != nil {
		return nil, fmt.Errorf("person 1 datetime: %w", err)
	}
	jd2, err := ParseDatetime(datetime2)
	if err != nil {
		return nil, fmt.Errorf("person 2 datetime: %w", err)
	}

	return composite.CalcCompositeChart(composite.CompositeInput{
		Person1Lat: lat1, Person1Lon: lon1, Person1JD: jd1,
		Person2Lat: lat2, Person2Lon: lon2, Person2JD: jd2,
		Planets:     DefaultPlanets,
		OrbConfig:   models.DefaultOrbConfig(),
		HouseSystem: models.HousePlacidus,
	})
}

// Dignities returns essential dignities for all planets in a chart.
func Dignities(lat, lon float64, datetime string) ([]dignity.DignityInfo, error) {
	jd, err := ParseDatetime(datetime)
	if err != nil {
		return nil, fmt.Errorf("datetime: %w", err)
	}
	chartInfo, err := chart.CalcSingleChart(lat, lon, jd, DefaultPlanets,
		models.DefaultOrbConfig(), models.HousePlacidus)
	if err != nil {
		return nil, err
	}
	return dignity.CalcChartDignities(chartInfo.Planets), nil
}

// AspectPatterns detects aspect patterns in a natal chart.
func AspectPatterns(lat, lon float64, datetime string) ([]aspect.AspectPattern, error) {
	jd, err := ParseDatetime(datetime)
	if err != nil {
		return nil, fmt.Errorf("datetime: %w", err)
	}
	orbs := models.DefaultOrbConfig()
	chartInfo, err := chart.CalcSingleChart(lat, lon, jd, DefaultPlanets, orbs, models.HousePlacidus)
	if err != nil {
		return nil, err
	}

	var bodies []aspect.Body
	for _, p := range chartInfo.Planets {
		bodies = append(bodies, aspect.Body{
			ID: string(p.PlanetID), Longitude: p.Longitude, Speed: p.Speed,
		})
	}
	return aspect.FindPatterns(chartInfo.Aspects, bodies, orbs), nil
}

// FullReport generates a comprehensive natal chart analysis combining all techniques:
// chart positions, dignities, dispositors, patterns, lots, faces, antiscia,
// fixed stars, midpoints, element/modality balance, and more.
func FullReport(lat, lon float64, datetime string) (*report.ChartReport, error) {
	if err := ValidateCoords(lat, lon); err != nil {
		return nil, err
	}
	jd, err := ParseDatetime(datetime)
	if err != nil {
		return nil, fmt.Errorf("datetime: %w", err)
	}
	return report.GenerateNatalReport(lat, lon, jd)
}



// ChartWheel generates chart wheel rendering coordinates for visualization.
func ChartWheel(lat, lon float64, datetime string) (*render.ChartWheel, error) {
	if err := ValidateCoords(lat, lon); err != nil {
		return nil, err
	}
	jd, err := ParseDatetime(datetime)
	if err != nil {
		return nil, fmt.Errorf("datetime: %w", err)
	}
	chartInfo, err := chart.CalcSingleChart(lat, lon, jd, DefaultPlanets,
		models.DefaultOrbConfig(), models.HousePlacidus)
	if err != nil {
		return nil, err
	}
	return render.CalcChartWheel(chartInfo, 0.4), nil
}

// PlanetPosition returns a single planet's position at a datetime.
func PlanetPosition(planet string, datetime string) (*models.PlanetPosition, error) {
	jd, err := ParseDatetime(datetime)
	if err != nil {
		return nil, fmt.Errorf("datetime: %w", err)
	}
	pid, err := ParsePlanet(planet)
	if err != nil {
		return nil, err
	}
	lon, speed, err := chart.CalcPlanetLongitude(pid, jd)
	if err != nil {
		return nil, err
	}
	return &models.PlanetPosition{
		PlanetID:     pid,
		Longitude:    lon,
		Speed:        speed,
		IsRetrograde: speed < 0,
		Sign:         models.SignFromLongitude(lon),
		SignDegree:   models.SignDegreeFromLongitude(lon),
	}, nil
}

// Firdaria calculates Firdaria planetary periods for a given age.
// isDayBirth should be true for daytime births, false for nighttime births.
func Firdaria(isDayBirth bool, age float64) *firdaria.FirdariaResult {
	return firdaria.CalcFirdaria(isDayBirth, age)
}

// DavisonChart calculates a Davison relationship chart for two people.
// The Davison chart uses the midpoint in time and space between two birth charts.
func DavisonChart(lat1, lon1 float64, datetime1 string, lat2, lon2 float64, datetime2 string) (*composite.DavisonChart, error) {
	if err := ValidateCoords(lat1, lon1); err != nil {
		return nil, fmt.Errorf("person 1: %w", err)
	}
	if err := ValidateCoords(lat2, lon2); err != nil {
		return nil, fmt.Errorf("person 2: %w", err)
	}
	jd1, err := ParseDatetime(datetime1)
	if err != nil {
		return nil, fmt.Errorf("person 1 datetime: %w", err)
	}
	jd2, err := ParseDatetime(datetime2)
	if err != nil {
		return nil, fmt.Errorf("person 2 datetime: %w", err)
	}

	return composite.CalcDavisonChart(composite.CompositeInput{
		Person1Lat: lat1, Person1Lon: lon1, Person1JD: jd1,
		Person2Lat: lat2, Person2Lon: lon2, Person2JD: jd2,
		Planets:     DefaultPlanets,
		OrbConfig:   models.DefaultOrbConfig(),
		HouseSystem: models.HousePlacidus,
	})
}

// PrimaryDirections calculates primary directions for a natal chart.
// Uses the Naibod key and Placidus houses by default.
func PrimaryDirections(lat, lon float64, datetime string, maxAge float64) (*primary.PrimaryDirectionResult, error) {
	if err := ValidateCoords(lat, lon); err != nil {
		return nil, err
	}
	jd, err := ParseDatetime(datetime)
	if err != nil {
		return nil, fmt.Errorf("datetime: %w", err)
	}
	return primary.CalcPrimaryDirections(primary.PrimaryDirectionInput{
		NatalJD:     jd,
		GeoLat:      lat,
		GeoLon:      lon,
		Planets:     DefaultPlanets,
		Aspects:     []models.AspectType{models.AspectConjunction, models.AspectOpposition, models.AspectSquare, models.AspectTrine, models.AspectSextile},
		Key:         primary.KeyNaibod,
		MaxAge:      maxAge,
		HouseSystem: models.HousePlacidus,
	})
}



// Bonification analyzes bonification and maltreatment for all planets in a chart.
func Bonification(lat, lon float64, datetime string) ([]dignity.BonMalInfo, error) {
	if err := ValidateCoords(lat, lon); err != nil {
		return nil, err
	}
	jd, err := ParseDatetime(datetime)
	if err != nil {
		return nil, fmt.Errorf("datetime: %w", err)
	}
	chartInfo, err := chart.CalcSingleChart(lat, lon, jd, DefaultPlanets,
		models.DefaultOrbConfig(), models.HousePlacidus)
	if err != nil {
		return nil, err
	}
	return dignity.CalcChartBonMal(chartInfo.Planets), nil
}

// SymbolicDirections calculates symbolic directions at a given age.
// Uses the one-degree-per-year method by default.
func SymbolicDirections(lat, lon float64, datetime string, age float64) (*symbolic.SymbolicDirectionResult, error) {
	if err := ValidateCoords(lat, lon); err != nil {
		return nil, err
	}
	jd, err := ParseDatetime(datetime)
	if err != nil {
		return nil, fmt.Errorf("datetime: %w", err)
	}
	return symbolic.CalcSymbolicDirections(symbolic.SymbolicInput{
		NatalJD:     jd,
		GeoLat:      lat,
		GeoLon:      lon,
		Age:         age,
		Method:      symbolic.MethodOneDegree,
		Planets:     DefaultPlanets,
		OrbConfig:   models.DefaultOrbConfig(),
		HouseSystem: models.HousePlacidus,
	})
}

// Options configures calculation parameters for the convenience API.
type Options struct {
	Planets     []models.PlanetID
	OrbConfig   models.OrbConfig
	HouseSystem models.HouseSystem
}

// DefaultOptions returns the standard calculation options.
func DefaultOptions() Options {
	return Options{
		Planets:     DefaultPlanets,
		OrbConfig:   models.DefaultOrbConfig(),
		HouseSystem: models.HousePlacidus,
	}
}

// NatalChartWithOptions calculates a natal chart with custom configuration.
func NatalChartWithOptions(lat, lon float64, datetime string, opts Options) (*models.ChartInfo, error) {
	if err := ValidateCoords(lat, lon); err != nil {
		return nil, err
	}
	jd, err := ParseDatetime(datetime)
	if err != nil {
		return nil, fmt.Errorf("natal chart: %w", err)
	}
	planets := opts.Planets
	if len(planets) == 0 {
		planets = DefaultPlanets
	}
	return chart.CalcSingleChart(lat, lon, jd, planets, opts.OrbConfig, opts.HouseSystem)
}

// Person represents a birth chart subject with coordinates and datetime.
type Person struct {
	Lat      float64
	Lon      float64
	Datetime string
}

// BatchNatalCharts calculates natal charts for multiple people concurrently.
func BatchNatalCharts(people []Person) ([]*models.ChartInfo, []error) {
	charts := make([]*models.ChartInfo, len(people))
	errs := make([]error, len(people))
	for i, p := range people {
		charts[i], errs[i] = NatalChart(p.Lat, p.Lon, p.Datetime)
	}
	return charts, errs
}

// BatchCompatibility calculates compatibility scores for all pairs in a group.
type CompatibilityPair struct {
	IndexA int     `json:"index_a"`
	IndexB int     `json:"index_b"`
	Score  float64 `json:"score"`
}

// BatchGroupCompatibility calculates all pairwise compatibility scores.
func BatchGroupCompatibility(people []Person) ([]CompatibilityPair, error) {
	// Pre-calculate all charts
	charts := make([]*models.ChartInfo, len(people))
	for i, p := range people {
		jd, err := ParseDatetime(p.Datetime)
		if err != nil {
			return nil, fmt.Errorf("person %d: %w", i, err)
		}
		c, err := chart.CalcSingleChart(p.Lat, p.Lon, jd, DefaultPlanets,
			models.DefaultOrbConfig(), models.HousePlacidus)
		if err != nil {
			return nil, fmt.Errorf("person %d chart: %w", i, err)
		}
		charts[i] = c
	}

	// Calculate all pairs
	var pairs []CompatibilityPair
	orbs := models.DefaultOrbConfig()
	for i := 0; i < len(charts); i++ {
		for j := i + 1; j < len(charts); j++ {
			score := synastry.CalcSynastryFromCharts(charts[i].Planets, charts[j].Planets, orbs)
			pairs = append(pairs, CompatibilityPair{
				IndexA: i, IndexB: j, Score: score.Compatibility,
			})
		}
	}
	return pairs, nil
}

// ParseDatetime converts an ISO 8601 datetime string to Julian Day (UT).
// Timezone offsets are handled correctly (converted to UT before JD conversion).
//
// Supported formats:
//   - "2000-01-01T12:00:00Z"       (UTC)
//   - "2000-01-01T12:00:00+08:00"  (with timezone offset)
//   - "2000-01-01T12:00:00"        (assumed UTC)
//   - "2000-01-01 12:00:00"
//   - "2000-01-01 12:00"
//   - "2000-01-01"                  (midnight UTC)
func ParseDatetime(s string) (float64, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, f := range formats {
		t, err := time.Parse(f, s)
		if err == nil {
			// Convert to UTC to get proper UT
			utc := t.UTC()
			hour := float64(utc.Hour()) + float64(utc.Minute())/60.0 + float64(utc.Second())/3600.0
			return sweph.JulDay(utc.Year(), int(utc.Month()), utc.Day(), hour, true), nil
		}
	}
	return 0, fmt.Errorf("cannot parse datetime %q (supported: ISO 8601, e.g. \"2000-01-01T12:00:00Z\" or \"2000-01-01\")", s)
}

// ParseHouseSystem converts a case-insensitive house system name to HouseSystem.
// Accepts: "Placidus", "Koch", "Whole Sign", "WHOLE_SIGN", "topocentric", etc.
func ParseHouseSystem(name string) (models.HouseSystem, error) {
	normalized := strings.ToUpper(strings.TrimSpace(name))
	normalized = strings.ReplaceAll(normalized, " ", "_")

	hsMap := map[string]models.HouseSystem{
		"PLACIDUS":      models.HousePlacidus,
		"KOCH":          models.HouseKoch,
		"EQUAL":         models.HouseEqual,
		"WHOLE_SIGN":    models.HouseWholeSign,
		"WHOLESIGN":     models.HouseWholeSign,
		"CAMPANUS":      models.HouseCampanus,
		"REGIOMONTANUS": models.HouseRegiomontanus,
		"PORPHYRY":      models.HousePorphyry,
		"MORINUS":       models.HouseMorinus,
		"TOPOCENTRIC":   models.HouseTopocentric,
		"POLICH-PAGE":   models.HouseTopocentric,
		"POLICH_PAGE":   models.HouseTopocentric,
		"ALCABITIUS":    models.HouseAlcabitius,
		"MERIDIAN":      models.HouseMeridian,
	}

	if hs, ok := hsMap[normalized]; ok {
		return hs, nil
	}
	return "", fmt.Errorf("unknown house system %q (supported: Placidus, Koch, Equal, Whole Sign, Campanus, Regiomontanus, Porphyry, Morinus, Topocentric, Alcabitius, Meridian)", name)
}

// ParsePlanet converts a case-insensitive planet name to PlanetID.
// Accepts: "Sun", "sun", "SUN", "Moon", "moon", etc.
func ParsePlanet(name string) (models.PlanetID, error) {
	normalized := strings.ToUpper(strings.TrimSpace(name))

	// Direct match
	planetMap := map[string]models.PlanetID{
		"SUN":             models.PlanetSun,
		"MOON":            models.PlanetMoon,
		"MERCURY":         models.PlanetMercury,
		"VENUS":           models.PlanetVenus,
		"MARS":            models.PlanetMars,
		"JUPITER":         models.PlanetJupiter,
		"SATURN":          models.PlanetSaturn,
		"URANUS":          models.PlanetUranus,
		"NEPTUNE":         models.PlanetNeptune,
		"PLUTO":           models.PlanetPluto,
		"CHIRON":          models.PlanetChiron,
		"NORTH_NODE":      models.PlanetNorthNodeTrue,
		"NORTH_NODE_TRUE": models.PlanetNorthNodeTrue,
		"NORTH_NODE_MEAN": models.PlanetNorthNodeMean,
		"SOUTH_NODE":      models.PlanetSouthNode,
		"LILITH":          models.PlanetLilithMean,
		"LILITH_MEAN":     models.PlanetLilithMean,
		"LILITH_TRUE":     models.PlanetLilithTrue,
		"TRUE_NODE":       models.PlanetNorthNodeTrue,
		"MEAN_NODE":       models.PlanetNorthNodeMean,
		"NORTHNODE":       models.PlanetNorthNodeTrue,
		"SOUTHNODE":       models.PlanetSouthNode,
	}

	if pid, ok := planetMap[normalized]; ok {
		return pid, nil
	}
	return "", fmt.Errorf("unknown planet %q", name)
}
