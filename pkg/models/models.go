package models

import (
	"fmt"
	"math"
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

// aspectCSVNames maps aspect types to Solar Fire CSV column names.
var aspectCSVNames = map[AspectType]string{
	AspectSextile:        "Quincunx",
	AspectTrine:          "Sextile",
	AspectQuincunx:       "Trine",
	AspectSemiSquare:     "Opposition",
	AspectSesquiquadrate: "Opposition",
	AspectOpposition:     "Quincunx",
}

// AspectCSVName maps aspect types to Solar Fire CSV column names.
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

// OrbConfig holds the orb (tolerance) for each aspect type
type OrbConfig struct {
	Conjunction    float64 `json:"conjunction"`
	Opposition     float64 `json:"opposition"`
	Trine          float64 `json:"trine"`
	Square         float64 `json:"square"`
	Sextile        float64 `json:"sextile"`
	Quincunx       float64 `json:"quincunx"`
	SemiSextile    float64 `json:"semi_sextile"`
	SemiSquare     float64 `json:"semi_square"`
	Sesquiquadrate float64 `json:"sesquiquadrate"`
}

// DefaultOrbConfig returns default orb values
func DefaultOrbConfig() OrbConfig {
	return OrbConfig{
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

// GetOrb returns the orb for a given aspect type
func (o OrbConfig) GetOrb(at AspectType) float64 {
	switch at {
	case AspectConjunction:
		return o.Conjunction
	case AspectOpposition:
		return o.Opposition
	case AspectTrine:
		return o.Trine
	case AspectSquare:
		return o.Square
	case AspectSextile:
		return o.Sextile
	case AspectQuincunx:
		return o.Quincunx
	case AspectSemiSextile:
		return o.SemiSextile
	case AspectSemiSquare:
		return o.SemiSquare
	case AspectSesquiquadrate:
		return o.Sesquiquadrate
	default:
		return 0
	}
}

// ChartType represents which chart a body belongs to
type ChartType string

const (
	ChartTransit      ChartType = "TRANSIT"
	ChartNatal        ChartType = "NATAL"
	ChartProgressions ChartType = "PROGRESSIONS"
	ChartSolarArc     ChartType = "SOLAR_ARC"
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

// ZodiacAbbr maps sign index (0-11) to 3-letter abbreviation (Solar Fire style)
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

// bodyDisplayNames maps internal IDs to display names (Solar Fire style)
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
