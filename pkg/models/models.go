package models

import (
	"fmt"
	"math"
	"strings"
)

// PlanetID represents a celestial body identifier
type PlanetID string

const (
	PlanetSun           PlanetID = "SUN"
	PlanetMoon          PlanetID = "MOON"
	PlanetMercury       PlanetID = "MERCURY"
	PlanetVenus         PlanetID = "VENUS"
	PlanetMars          PlanetID = "MARS"
	PlanetJupiter       PlanetID = "JUPITER"
	PlanetSaturn        PlanetID = "SATURN"
	PlanetUranus        PlanetID = "URANUS"
	PlanetNeptune       PlanetID = "NEPTUNE"
	PlanetPluto         PlanetID = "PLUTO"
	PlanetChiron        PlanetID = "CHIRON"
	PlanetNorthNodeTrue PlanetID = "NORTH_NODE_TRUE"
	PlanetNorthNodeMean PlanetID = "NORTH_NODE_MEAN"
	PlanetSouthNode     PlanetID = "SOUTH_NODE"
	PlanetLilithMean    PlanetID = "LILITH_MEAN"
	PlanetLilithTrue    PlanetID = "LILITH_TRUE"
)

// SpecialPointID represents a derived astrological point
type SpecialPointID string

const (
	PointASC        SpecialPointID = "ASC"
	PointMC         SpecialPointID = "MC"
	PointDSC        SpecialPointID = "DSC"
	PointIC         SpecialPointID = "IC"
	PointVertex     SpecialPointID = "VERTEX"
	PointAntiVertex SpecialPointID = "ANTI_VERTEX"
	PointEastPoint  SpecialPointID = "EAST_POINT"
	PointLotFortune SpecialPointID = "LOT_FORTUNE"
	PointLotSpirit  SpecialPointID = "LOT_SPIRIT"
)

// HouseSystem represents a house system type
type HouseSystem string

const (
	HousePlacidus      HouseSystem = "PLACIDUS"
	HouseKoch          HouseSystem = "KOCH"
	HouseEqual         HouseSystem = "EQUAL"
	HouseWholeSign     HouseSystem = "WHOLE_SIGN"
	HouseCampanus      HouseSystem = "CAMPANUS"
	HouseRegiomontanus HouseSystem = "REGIOMONTANUS"
	HousePorphyry      HouseSystem = "PORPHYRY"
	HouseMorinus       HouseSystem = "MORINUS"
	HouseTopocentric   HouseSystem = "TOPOCENTRIC"
	HouseAlcabitius    HouseSystem = "ALCABITIUS"
	HouseMeridian      HouseSystem = "MERIDIAN"
	HouseSripati       HouseSystem = "SRIPATI"
)

// CalendarType represents calendar type
type CalendarType string

const (
	CalendarGregorian CalendarType = "GREGORIAN"
	CalendarJulian    CalendarType = "JULIAN"
)

// AspectType represents a type of aspect
type AspectType string

const (
	AspectConjunction    AspectType = "Conjunction"
	AspectOpposition     AspectType = "Opposition"
	AspectTrine          AspectType = "Trine"
	AspectSquare         AspectType = "Square"
	AspectSextile        AspectType = "Sextile"
	AspectQuincunx       AspectType = "Quincunx"
	AspectSemiSextile    AspectType = "Semi-Sextile"
	AspectSemiSquare     AspectType = "Semi-Square"
	AspectSesquiquadrate AspectType = "Sesquiquadrate"
)

// aspectCSVNames maps aspect types to CSV column names.
// Only needed for types whose constant value differs from the desired CSV name.
// Currently all AspectType constants already match Solar Fire CSV names.
var aspectCSVNames = map[AspectType]string{}

// AspectCSVName maps aspect types to CSV column names.
func AspectCSVName(at AspectType) string {
	if name, ok := aspectCSVNames[at]; ok {
		return name
	}
	return string(at)
}

// AspectDef defines a standard aspect with its angle
type AspectDef struct {
	Type  AspectType
	Angle float64
}

// StandardAspects lists all standard aspects
var StandardAspects = []AspectDef{
	{AspectConjunction, 0},
	{AspectOpposition, 180},
	{AspectTrine, 120},
	{AspectSquare, 90},
	{AspectSextile, 60},
	{AspectQuincunx, 150},
	{AspectSemiSextile, 30},
	{AspectSemiSquare, 45},
	{AspectSesquiquadrate, 135},
}

// AspectOrbDef defines the orb configuration for a single aspect type.
// This is the flexible, user-friendly format for API configuration.
type AspectOrbDef struct {
	Name        string  `json:"name"`         // Aspect name (e.g., "conjunction", "my_custom_aspect")
	Angle       float64 `json:"angle"`        // Aspect angle in degrees (e.g., 0, 180, 120)
	EnteringOrb float64 `json:"entering_orb"` // Orb for entering/applying aspects
	ExitingOrb  float64 `json:"exiting_orb"`  // Orb for exiting/separating aspects
	Enabled     bool    `json:"enabled"`      // Whether this aspect is enabled (default: true)
}

// OrbConfig holds the orb configuration.
// Can be specified either as:
// 1. A list of AspectOrbDef (flexible format, recommended for APIs)
// 2. Legacy flat fields (backward compatible)
type OrbConfig struct {
	// Definitions is a list of aspect definitions with entering/exiting orbs.
	// When provided, this takes precedence over the legacy fields.
	Definitions []AspectOrbDef `json:"definitions,omitempty"`

	// Legacy fields for backward compatibility
	Conjunction    float64 `json:"conjunction,omitempty"`
	Opposition     float64 `json:"opposition,omitempty"`
	Trine          float64 `json:"trine,omitempty"`
	Square         float64 `json:"square,omitempty"`
	Sextile        float64 `json:"sextile,omitempty"`
	Quincunx       float64 `json:"quincunx,omitempty"`
	SemiSextile    float64 `json:"semi_sextile,omitempty"`
	SemiSquare     float64 `json:"semi_square,omitempty"`
	Sesquiquadrate float64 `json:"sesquiquadrate,omitempty"`
}

// defaultAspectDefs defines the built-in default aspects
// Names use hyphens to match AspectType constants (e.g., "Semi-Sextile")
var defaultAspectDefs = []AspectOrbDef{
	{Name: "conjunction", Angle: 0, EnteringOrb: 8, ExitingOrb: 8, Enabled: true},
	{Name: "opposition", Angle: 180, EnteringOrb: 8, ExitingOrb: 8, Enabled: true},
	{Name: "trine", Angle: 120, EnteringOrb: 7, ExitingOrb: 7, Enabled: true},
	{Name: "square", Angle: 90, EnteringOrb: 7, ExitingOrb: 7, Enabled: true},
	{Name: "sextile", Angle: 60, EnteringOrb: 5, ExitingOrb: 5, Enabled: true},
	{Name: "quincunx", Angle: 150, EnteringOrb: 3, ExitingOrb: 3, Enabled: true},
	{Name: "semi-sextile", Angle: 30, EnteringOrb: 2, ExitingOrb: 2, Enabled: true},
	{Name: "semi-square", Angle: 45, EnteringOrb: 2, ExitingOrb: 2, Enabled: true},
	{Name: "sesquiquadrate", Angle: 135, EnteringOrb: 2, ExitingOrb: 2, Enabled: true},
}

// DefaultOrbConfig returns default orb configuration
func DefaultOrbConfig() OrbConfig {
	return OrbConfig{
		Definitions: defaultAspectDefs,
		// Also populate legacy fields for backward compatibility
		Conjunction:    8,
		Opposition:     8,
		Trine:          7,
		Square:         7,
		Sextile:        5,
		Quincunx:       3,
		SemiSextile:    2,
		SemiSquare:     2,
		Sesquiquadrate: 2,
	}
}

// GetAspectDefs returns the effective aspect definitions.
// If Definitions is provided, returns that; otherwise builds from legacy fields.
// A negative orb value disables the aspect. Zero uses the default.
func (o OrbConfig) GetAspectDefs() []AspectOrbDef {
	if len(o.Definitions) > 0 {
		return o.Definitions
	}
	// Build from legacy fields, using defaults if zero
	// Names must match AspectType constants for backward compatibility
	return []AspectOrbDef{
		buildDef("conjunction", 0, o.Conjunction, 8),
		buildDef("opposition", 180, o.Opposition, 8),
		buildDef("trine", 120, o.Trine, 7),
		buildDef("square", 90, o.Square, 7),
		buildDef("sextile", 60, o.Sextile, 5),
		buildDef("quincunx", 150, o.Quincunx, 3),
		buildDef("semi-sextile", 30, o.SemiSextile, 2),
		buildDef("semi-square", 45, o.SemiSquare, 2),
		buildDef("sesquiquadrate", 135, o.Sesquiquadrate, 2),
	}
}

// buildDef creates an AspectOrbDef. A negative orb value disables the aspect.
// Zero uses the default.
func buildDef(name string, angle, orb, defaultOrb float64) AspectOrbDef {
	if orb < 0 {
		return AspectOrbDef{Name: name, Angle: angle, Enabled: false}
	}
	effectiveOrb := defaultIfZero(orb, defaultOrb)
	return AspectOrbDef{Name: name, Angle: angle, EnteringOrb: effectiveOrb, ExitingOrb: effectiveOrb, Enabled: true}
}

// defaultIfZero returns val if non-zero, otherwise returns def
func defaultIfZero(val, def float64) float64 {
	if val == 0 {
		return def
	}
	return val
}

// GetOrbForAspect returns the appropriate orb for an aspect based on entering/exiting state.
// Looks up by aspect angle (with tolerance for matching).
func (o OrbConfig) GetOrbForAspect(angle float64, isEntering bool) float64 {
	defs := o.GetAspectDefs()
	for _, def := range defs {
		if !def.Enabled {
			continue
		}
		// Match by angle (within 1 degree tolerance)
		if math.Abs(def.Angle-angle) < 1.0 {
			if isEntering {
				return def.EnteringOrb
			}
			return def.ExitingOrb
		}
	}
	return 0
}

// GetOrb returns the orb for a given aspect type (backward compatible).
func (o OrbConfig) GetOrb(at AspectType) float64 {
	defs := o.GetAspectDefs()
	atLower := strings.ToLower(string(at))
	for _, def := range defs {
		if strings.ToLower(def.Name) == atLower && def.Enabled {
			return def.EnteringOrb // backward compatible: return entering orb
		}
	}
	return 0
}

// GetEnteringOrb returns the entering orb for an aspect type (backward compatible).
func (o OrbConfig) GetEnteringOrb(at AspectType) float64 {
	return o.GetOrb(at)
}

// GetExitingOrb returns the exiting orb for an aspect type (backward compatible).
func (o OrbConfig) GetExitingOrb(at AspectType) float64 {
	defs := o.GetAspectDefs()
	for _, def := range defs {
		if def.Name == string(at) && def.Enabled {
			return def.ExitingOrb
		}
	}
	return 0
}

// ChartType represents which chart a body belongs to
type ChartType string

const (
	ChartTransit      ChartType = "TRANSIT"
	ChartNatal        ChartType = "NATAL"
	ChartProgressions ChartType = "PROGRESSIONS"
	ChartSolarArc     ChartType = "SOLAR_ARC"
	ChartSolarReturn  ChartType = "SOLAR_RETURN"
	ChartLunarReturn  ChartType = "LUNAR_RETURN"
)

// EventType represents a transit event type
type EventType string

const (
	EventAspectBegin  EventType = "ASPECT_BEGIN"
	EventAspectEnter  EventType = "ASPECT_ENTER"
	EventAspectExact  EventType = "ASPECT_EXACT"
	EventAspectLeave  EventType = "ASPECT_LEAVE"
	EventSignIngress  EventType = "SIGN_INGRESS"
	EventHouseIngress EventType = "HOUSE_INGRESS"
	EventStation      EventType = "STATION"
	EventVoidOfCourse EventType = "VOID_OF_COURSE"
)

// StationType represents retrograde/direct station
type StationType string

const (
	StationRetrograde StationType = "RETROGRADE"
	StationDirect     StationType = "DIRECT"
)

// PlanetPosition holds calculated position data for a planet
type PlanetPosition struct {
	PlanetID    PlanetID `json:"planet_id"`
	Longitude   float64  `json:"longitude"`
	Latitude    float64  `json:"latitude"`
	Speed       float64  `json:"speed"`
	IsRetrograde bool    `json:"is_retrograde"`
	Sign        string   `json:"sign"`
	SignDegree  float64  `json:"sign_degree"`
	House       int      `json:"house"`
}

// AnglesInfo holds the four angles
type AnglesInfo struct {
	ASC float64 `json:"asc"`
	MC  float64 `json:"mc"`
	DSC float64 `json:"dsc"`
	IC  float64 `json:"ic"`
}

// AspectInfo holds aspect data between two bodies
type AspectInfo struct {
	PlanetA     string     `json:"planet_a"`
	PlanetB     string     `json:"planet_b"`
	AspectType  AspectType `json:"aspect_type"`
	AspectAngle float64    `json:"aspect_angle"`
	ActualAngle float64    `json:"actual_angle"`
	Orb         float64    `json:"orb"`
	IsApplying  bool       `json:"is_applying"`
}

// ChartInfo holds complete chart data
type ChartInfo struct {
	Planets []PlanetPosition `json:"planets"`
	Houses  []float64        `json:"houses"`
	Angles  AnglesInfo       `json:"angles"`
	Aspects []AspectInfo     `json:"aspects"`
}

// CrossAspectInfo holds aspect data between two charts
type CrossAspectInfo struct {
	InnerBody   string     `json:"inner_body"`
	OuterBody   string     `json:"outer_body"`
	AspectType  AspectType `json:"aspect_type"`
	AspectAngle float64    `json:"aspect_angle"`
	ActualAngle float64    `json:"actual_angle"`
	Orb         float64    `json:"orb"`
	IsApplying  bool       `json:"is_applying"`
}

// GeoLocation holds geographic coordinates
type GeoLocation struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timezone    string  `json:"timezone"`
	DisplayName string  `json:"display_name"`
}

// JulianDayResult holds Julian Day conversion result
type JulianDayResult struct {
	JDUT float64 `json:"jd_ut"`
	JDTT float64 `json:"jd_tt"`
}

// SpecialPointsConfig configures which special points to include
type SpecialPointsConfig struct {
	InnerPoints        []SpecialPointID `json:"inner_points,omitempty"`
	OuterPoints        []SpecialPointID `json:"outer_points,omitempty"`
	NatalPoints        []SpecialPointID `json:"natal_points,omitempty"`
	TransitPoints      []SpecialPointID `json:"transit_points,omitempty"`
	ProgressionsPoints []SpecialPointID `json:"progressions_points,omitempty"`
	SolarArcPoints     []SpecialPointID `json:"solar_arc_points,omitempty"`
}

// EventConfig configures which event types to include
type EventConfig struct {
	IncludeTrNa         bool `json:"include_tr_na"`
	IncludeTrTr         bool `json:"include_tr_tr"`
	IncludeTrSp         bool `json:"include_tr_sp"`
	IncludeTrSa         bool `json:"include_tr_sa"`
	IncludeSpNa         bool `json:"include_sp_na"`
	IncludeSpSp         bool `json:"include_sp_sp"`
	IncludeSaNa         bool `json:"include_sa_na"`
	IncludeSignIngress  bool `json:"include_sign_ingress"`
	IncludeHouseIngress bool `json:"include_house_ingress"`
	IncludeStation      bool `json:"include_station"`
	IncludeVoidOfCourse bool `json:"include_void_of_course"`
}

// DefaultEventConfig returns config with all events enabled
func DefaultEventConfig() EventConfig {
	return EventConfig{
		IncludeTrNa:         true,
		IncludeTrTr:         true,
		IncludeTrSp:         true,
		IncludeTrSa:         true,
		IncludeSpNa:         true,
		IncludeSpSp:         true,
		IncludeSaNa:         true,
		IncludeSignIngress:  true,
		IncludeHouseIngress: true,
		IncludeStation:      true,
		IncludeVoidOfCourse: true,
	}
}

// ProgressionsConfig configures secondary progressions
type ProgressionsConfig struct {
	Enabled bool       `json:"enabled"`
	Planets []PlanetID `json:"planets"`
}

// SolarArcConfig configures solar arc directions
type SolarArcConfig struct {
	Enabled bool       `json:"enabled"`
	Planets []PlanetID `json:"planets"`
}

// TransitEvent represents an astrological transit event
type TransitEvent struct {
	EventType       EventType  `json:"event_type"`
	ChartType       ChartType  `json:"chart_type"`
	Planet          PlanetID   `json:"planet"`
	JD              float64    `json:"jd"`
	Age             float64    `json:"age,omitempty"`
	PlanetLongitude float64    `json:"planet_longitude"`
	PlanetSign      string     `json:"planet_sign"`
	PlanetHouse     int        `json:"planet_house"`
	IsRetrograde    bool       `json:"is_retrograde"`

	// Aspect events
	TargetChartType    ChartType  `json:"target_chart_type,omitempty"`
	Target             string     `json:"target,omitempty"`
	TargetLongitude    float64    `json:"target_longitude,omitempty"`
	TargetSign         string     `json:"target_sign,omitempty"`
	TargetHouse        int        `json:"target_house,omitempty"`
	TargetIsRetrograde bool       `json:"target_is_retrograde,omitempty"`
	AspectType         AspectType `json:"aspect_type,omitempty"`
	AspectAngle        float64    `json:"aspect_angle,omitempty"`
	OrbAtEnter         float64    `json:"orb_at_enter,omitempty"`
	OrbAtLeave         float64    `json:"orb_at_leave,omitempty"`
	ExactCount         int        `json:"exact_count,omitempty"`

	// Sign ingress
	FromSign string `json:"from_sign,omitempty"`
	ToSign   string `json:"to_sign,omitempty"`

	// House ingress
	FromHouse int `json:"from_house,omitempty"`
	ToHouse   int `json:"to_house,omitempty"`

	// Station
	StationType StationType `json:"station_type,omitempty"`

	// Void of course
	VoidStartJD      float64  `json:"void_start_jd,omitempty"`
	VoidEndJD        float64  `json:"void_end_jd,omitempty"`
	LastAspectType   string   `json:"last_aspect_type,omitempty"`
	LastAspectTarget string   `json:"last_aspect_target,omitempty"`
	NextSign         string   `json:"next_sign,omitempty"`
}

// ZodiacSigns maps sign index (0-11) to sign name
var ZodiacSigns = []string{
	"Aries", "Taurus", "Gemini", "Cancer",
	"Leo", "Virgo", "Libra", "Scorpio",
	"Sagittarius", "Capricorn", "Aquarius", "Pisces",
}

// ZodiacAbbr maps sign index (0-11) to 3-letter abbreviation
var ZodiacAbbr = []string{
	"Ari", "Tau", "Gem", "Can",
	"Leo", "Vir", "Lib", "Sco",
	"Sag", "Cap", "Aqu", "Pis",
}

// signIndex returns the zodiac sign index (0-11) for a given ecliptic longitude
func signIndex(lon float64) int {
	idx := int(lon / 30.0)
	if idx < 0 {
		return 0
	}
	if idx > 11 {
		return 11
	}
	return idx
}

// SignFromLongitude returns the zodiac sign name for a given ecliptic longitude
func SignFromLongitude(lon float64) string {
	return ZodiacSigns[signIndex(lon)]
}

// SignAbbrFromLongitude returns the 3-letter zodiac abbreviation
func SignAbbrFromLongitude(lon float64) string {
	return ZodiacAbbr[signIndex(lon)]
}

// SignDegreeFromLongitude returns the degree within the sign (0-30)
func SignDegreeFromLongitude(lon float64) float64 {
	return lon - float64(int(lon/30.0))*30.0
}

// DMS holds degrees, minutes, seconds
type DMS struct {
	Degrees int
	Minutes int
	Seconds int
	Neg     bool
}

// ToDMS converts a decimal degree value to degrees, minutes, seconds
func ToDMS(deg float64) DMS {
	neg := deg < 0
	if neg {
		deg = -deg
	}
	d := int(deg)
	mf := (deg - float64(d)) * 60.0
	m := int(mf)
	s := int((mf - float64(m)) * 60.0)
	return DMS{Degrees: d, Minutes: m, Seconds: s, Neg: neg}
}

// String returns DMS in "DD°MM'SS\"" format
func (d DMS) String() string {
	sign := ""
	if d.Neg {
		sign = "-"
	}
	return fmt.Sprintf("%s%d°%02d'%02d\"", sign, d.Degrees, d.Minutes, d.Seconds)
}

// FormatLonDMS formats an ecliptic longitude as "DD°MM'SS\" Sign"
// e.g. 266.4997 -> "26°29'59\" Sag"
func FormatLonDMS(lon float64) string {
	signDeg := SignDegreeFromLongitude(lon)
	abbr := SignAbbrFromLongitude(lon)
	dms := ToDMS(signDeg)
	return fmt.Sprintf("%d°%02d'%02d\"%s", dms.Degrees, dms.Minutes, dms.Seconds, abbr)
}

// FormatDMS formats a decimal degree as "DD°MM'SS\""
func FormatDMS(deg float64) string {
	return ToDMS(deg).String()
}

// bodyDisplayNames maps internal IDs to display names
var bodyDisplayNames = map[string]string{
	string(PlanetSun):           "Sun",
	string(PlanetMoon):          "Moon",
	string(PlanetMercury):       "Mercury",
	string(PlanetVenus):         "Venus",
	string(PlanetMars):          "Mars",
	string(PlanetJupiter):       "Jupiter",
	string(PlanetSaturn):        "Saturn",
	string(PlanetUranus):        "Uranus",
	string(PlanetNeptune):       "Neptune",
	string(PlanetPluto):         "Pluto",
	string(PlanetChiron):        "Chiron",
	string(PlanetNorthNodeTrue): "NorthNode",
	string(PlanetNorthNodeMean): "NorthNode",
	string(PlanetSouthNode):     "SouthNode",
	string(PlanetLilithMean):    "Lilith",
	string(PlanetLilithTrue):    "Lilith",
	string(PointASC):            "ASC",
	string(PointMC):             "MC",
	string(PointDSC):            "DSC",
	string(PointIC):             "IC",
	string(PointVertex):         "Vertex",
	string(PointAntiVertex):     "AntiVertex",
	string(PointEastPoint):      "EastPoint",
	string(PointLotFortune):     "LotFortune",
	string(PointLotSpirit):      "LotSpirit",
}

// BodyDisplayName returns the display name for a planet or special point ID string.
func BodyDisplayName(id string) string {
	if name, ok := bodyDisplayNames[id]; ok {
		return name
	}
	return id
}

// chartTypeShortMap maps ChartType to its short code
var chartTypeShortMap = map[ChartType]string{
	ChartTransit:      "Tr",
	ChartNatal:        "Na",
	ChartProgressions: "Sp",
	ChartSolarArc:     "Sa",
	ChartSolarReturn:  "Sr",
	ChartLunarReturn:  "Lr",
}

// ChartTypeShort returns the short code for a ChartType (e.g. "Tr", "Na", "Sp", "Sa")
func ChartTypeShort(ct ChartType) string {
	if s, ok := chartTypeShortMap[ct]; ok {
		return s
	}
	return string(ct)
}

// EventTypeCSV returns the CSV event type string
func EventTypeCSV(et EventType, stationType StationType) string {
	switch et {
	case EventAspectBegin:
		return "Begin"
	case EventAspectEnter:
		return "Enter"
	case EventAspectExact:
		return "Exact"
	case EventAspectLeave:
		return "Leave"
	case EventSignIngress:
		return "SignIngress"
	case EventStation:
		if stationType == StationDirect {
			return "Direct"
		}
		return "Retrograde"
	case EventVoidOfCourse:
		return "Void"
	case EventHouseIngress:
		return "HouseIngress"
	default:
		return string(et)
	}
}

// FormatSignDegreeCSV formats a sign degree for CSV output.
// Truncates to arcminute precision, then rounds to 3 decimal places.
func FormatSignDegreeCSV(signDeg float64) float64 {
	d := int(signDeg)
	m := int((signDeg - float64(d)) * 60.0)
	return math.Round((float64(d)+float64(m)/60.0)*1000) / 1000
}
