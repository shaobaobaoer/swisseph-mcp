package transit

import (
	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
)

// TransitChartConfig holds configuration for transit chart calculations.
type TransitChartConfig struct {
	Lat, Lon    float64               // Transit location (affects ASC/MC special points)
	Planets     []models.PlanetID     // Transit planets to calculate
	Points      []models.SpecialPointID // Transit special points to calculate
	Orbs        models.OrbConfig      // Orb configuration for this chart
	HouseSystem models.HouseSystem    // House system (shared across all charts)
}

// ProgressionsChartConfig holds configuration for secondary progressions chart calculations.
type ProgressionsChartConfig struct {
	Planets     []models.PlanetID     // Progressed planets to calculate
	Points      []models.SpecialPointID // Progressed special points to calculate
	Orbs        models.OrbConfig      // Orb configuration for this chart
	Lat, Lon    float64               // Location for progressed special points
	HouseSystem models.HouseSystem    // House system (shared across all charts)
}

// SolarArcChartConfig holds configuration for solar arc chart calculations.
type SolarArcChartConfig struct {
	Planets     []models.PlanetID     // Solar arc planets to calculate
	Points      []models.SpecialPointID // Solar arc special points to calculate
	Orbs        models.OrbConfig      // Orb configuration for this chart
	Lat, Lon    float64               // Location for solar arc special points
	HouseSystem models.HouseSystem    // House system (shared across all charts)
}

// SolarReturnChartConfig holds configuration for solar return chart calculations.
type SolarReturnChartConfig struct {
	Lat, Lon      float64            // Solar return location
	SRChartJD     float64            // Exact SR moment JD (if known)
	NatalJD       float64            // Natal JD (needed to find SR if SRChartJD == 0)
	SearchAfterJD float64            // Start searching for SR after this JD
	Planets       []models.PlanetID
	Points        []models.SpecialPointID
	Orbs          models.OrbConfig
	HouseSystem   models.HouseSystem
}

// ChartSetConfig holds all chart configurations for transit calculation.
type ChartSetConfig struct {
	Transit      *TransitChartConfig      // nil means disabled
	Progressions *ProgressionsChartConfig // nil means disabled
	SolarArc     *SolarArcChartConfig     // nil means disabled
	SolarReturn  *SolarReturnChartConfig  // nil means disabled
}

// NatalChartConfig holds configuration for the natal chart (fixed reference).
type NatalChartConfig struct {
	Lat, Lon float64               // Birth location
	JD       float64               // Birth moment
	Planets  []models.PlanetID     // Natal planets to include
	Points   []models.SpecialPointID // Natal special points to include
	// ASCOverride and MCOverride allow specifying exact natal angle positions.
	// When non-zero, these override the calculated ASC/MC values.
	// Use this to match reference data (e.g., Solar Fire) that may use
	// different obliquity or house calculation parameters.
	ASCOverride float64
	MCOverride  float64
	// MCOverrideForASC is a separate override for ASC progression calculation.
	// Some systems (Solar Fire) use different MC base for ASC derivation than for MC progression.
	// If zero, falls back to MCOverride, then to computed value.
	// NOTE: This is only used if ASCOverrideForProgressions == 0.
	MCOverrideForASC float64
	// ASCOverrideForProgressions: if > 0, use direct solar arc to ASC method.
	// progASC = ASCOverrideForProgressions + solarArc
	// This matches Solar Fire's behavior and takes precedence over MCOverrideForASC.
	ASCOverrideForProgressions float64
	// PlanetOverrides allows specifying exact natal planet positions.
	// Use this to match reference data that may use different ephemeris (DE200 vs DE432).
	// Map key is planet ID (e.g., "MOON", "MERCURY"), value is longitude in degrees.
	PlanetOverrides map[string]float64
}

// TimeRangeConfig holds the time range for transit calculation.
type TimeRangeConfig struct {
	StartJD float64
	EndJD   float64
}

// EventFilterConfig holds flags for which event types to include.
type EventFilterConfig struct {
	// Event types
	Station      bool
	SignIngress  bool
	HouseIngress bool
	VoidOfCourse bool

	// Aspect combinations
	TrNa bool // Transit → Natal
	TrTr bool // Transit → Transit
	TrSp bool // Transit → Progressions
	TrSa bool // Transit → SolarArc
	TrSr bool // Transit → SolarReturn
	SpNa bool // Progressions → Natal
	SpSp bool // Progressions → Progressions
	SpSr bool // Progressions → SolarReturn
	SaNa bool // SolarArc → Natal
	SaSr bool // SolarArc → SolarReturn
}

// DefaultEventFilterConfig returns a config with all events enabled.
func DefaultEventFilterConfig() EventFilterConfig {
	return EventFilterConfig{
		Station:      true,
		SignIngress:  true,
		HouseIngress: true,
		VoidOfCourse: true,
		TrNa:         true,
		TrTr:         true,
		TrSp:         true,
		TrSa:         true,
		TrSr:         true,
		SpNa:         true,
		SpSp:         true,
		SpSr:         true,
		SaNa:         true,
		SaSr:         true,
	}
}
