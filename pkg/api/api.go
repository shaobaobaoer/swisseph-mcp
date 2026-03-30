package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/shaobaobaoer/solarsage-mcp/internal/aspect"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/chart"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/composite"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/dignity"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/export"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/geo"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/julian"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/lunar"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/progressions"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/returns"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/synastry"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/transit"
)

// defaultPlanets is the standard set of 10 planets used when none specified.
var defaultPlanets = []models.PlanetID{
	models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
	models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
	models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune,
	models.PlanetPluto,
}

// Config holds configuration for the API server.
type Config struct {
	EphePath string
	APIKey   string
	Port     int
}

// Server is the HTTP API server.
type Server struct {
	mux    *http.ServeMux
	apiKey string
}

// NewServer creates a new API server with all routes registered.
func NewServer(cfg Config) *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		apiKey: cfg.APIKey,
	}
	s.registerRoutes()
	return s
}

// ServeHTTP implements http.Handler with CORS and optional API key auth.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if s.apiKey != "" && r.Header.Get("X-API-Key") != s.apiKey {
		writeError(w, http.StatusUnauthorized, "invalid or missing API key")
		return
	}

	s.mux.ServeHTTP(w, r)
}

// Run starts the HTTP server on the given address.
func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s)
}

func (s *Server) registerRoutes() {
	// Health check (GET)
	s.mux.HandleFunc("/api/v1/health", s.handleHealth)

	// Basic
	s.mux.HandleFunc("/api/v1/geocode", s.requirePOST(s.handleGeocode))
	s.mux.HandleFunc("/api/v1/datetime/to-jd", s.requirePOST(s.handleDatetimeToJD))
	s.mux.HandleFunc("/api/v1/datetime/from-jd", s.requirePOST(s.handleJDToDatetime))
	s.mux.HandleFunc("/api/v1/planet/position", s.requirePOST(s.handlePlanetPosition))

	// Charts
	s.mux.HandleFunc("/api/v1/chart/natal", s.requirePOST(s.handleNatalChart))
	s.mux.HandleFunc("/api/v1/chart/double", s.requirePOST(s.handleDoubleChart))
	s.mux.HandleFunc("/api/v1/chart/composite", s.requirePOST(s.handleCompositeChart))
	s.mux.HandleFunc("/api/v1/chart/davison", s.requirePOST(s.handleDavisonChart))

	// Predictive
	s.mux.HandleFunc("/api/v1/transit", s.requirePOST(s.handleTransit))
	s.mux.HandleFunc("/api/v1/progressions", s.requirePOST(s.handleProgressions))
	s.mux.HandleFunc("/api/v1/solar-arc", s.requirePOST(s.handleSolarArc))
	s.mux.HandleFunc("/api/v1/solar-return", s.requirePOST(s.handleSolarReturn))
	s.mux.HandleFunc("/api/v1/lunar-return", s.requirePOST(s.handleLunarReturn))

	// Traditional
	s.mux.HandleFunc("/api/v1/dignity", s.requirePOST(s.handleDignity))
	s.mux.HandleFunc("/api/v1/bonification", s.requirePOST(s.handleBonification))

	// Analysis
	s.mux.HandleFunc("/api/v1/aspects/patterns", s.requirePOST(s.handleAspectPatterns))
	s.mux.HandleFunc("/api/v1/synastry", s.requirePOST(s.handleSynastry))

	// Lunar
	s.mux.HandleFunc("/api/v1/lunar/phase", s.requirePOST(s.handleLunarPhase))
	s.mux.HandleFunc("/api/v1/lunar/phases", s.requirePOST(s.handleLunarPhases))
	s.mux.HandleFunc("/api/v1/lunar/eclipses", s.requirePOST(s.handleEclipses))
}

// requirePOST wraps a handler to reject non-POST methods.
func (s *Server) requirePOST(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed, use POST")
			return
		}
		h(w, r)
	}
}

// --- Helpers ---

func decodeJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  msg,
		"status": status,
	})
}

// orbOrDefault returns the custom orb config if provided, otherwise the fallback.
func orbOrDefault(custom *models.OrbConfig, fallback models.OrbConfig) models.OrbConfig {
	if custom != nil {
		return *custom
	}
	return fallback
}

// --- Health ---

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- Basic Endpoints ---

func (s *Server) handleGeocode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LocationName string `json:"location_name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.LocationName == "" {
		writeError(w, http.StatusBadRequest, "location_name is required")
		return
	}
	result, err := geo.Geocode(req.LocationName)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleDatetimeToJD(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Datetime string              `json:"datetime"`
		Calendar models.CalendarType `json:"calendar"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Datetime == "" {
		writeError(w, http.StatusBadRequest, "datetime is required")
		return
	}
	if req.Calendar == "" {
		req.Calendar = models.CalendarGregorian
	}
	result, err := julian.DateTimeToJD(req.Datetime, req.Calendar)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleJDToDatetime(w http.ResponseWriter, r *http.Request) {
	var req struct {
		JD       float64 `json:"jd"`
		Timezone string  `json:"timezone"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Timezone == "" {
		req.Timezone = "UTC"
	}
	dt, err := julian.JDToDateTime(req.JD, req.Timezone)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"datetime": dt})
}

func (s *Server) handlePlanetPosition(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Planet models.PlanetID `json:"planet"`
		JDUT   float64         `json:"jd_ut"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	lon, speed, err := chart.CalcPlanetLongitude(req.Planet, req.JDUT)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"planet":        req.Planet,
		"longitude":     lon,
		"speed":         speed,
		"is_retrograde": speed < 0,
		"sign":          models.SignFromLongitude(lon),
		"sign_degree":   models.SignDegreeFromLongitude(lon),
	})
}

// --- Chart Endpoints ---

func (s *Server) handleNatalChart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Latitude    float64            `json:"latitude"`
		Longitude   float64            `json:"longitude"`
		JDUT        float64            `json:"jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		OrbConfig   *models.OrbConfig  `json:"orb_config"`
		Format      string             `json:"format"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Planets) == 0 {
		req.Planets = defaultPlanets
	}
	if req.HouseSystem == "" {
		req.HouseSystem = models.HousePlacidus
	}

	chartInfo, err := chart.CalcSingleChart(req.Latitude, req.Longitude, req.JDUT,
		req.Planets, orbOrDefault(req.OrbConfig, models.DefaultOrbConfig()), req.HouseSystem)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if req.Format == "csv" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"format":  "csv",
			"planets": export.ChartToCSV(chartInfo),
			"aspects": export.AspectsToCSV(chartInfo.Aspects),
			"houses":  export.HousesToCSV(chartInfo.Houses, chartInfo.Angles),
		})
		return
	}
	writeJSON(w, http.StatusOK, chartInfo)
}

func (s *Server) handleDoubleChart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		InnerLatitude  float64                     `json:"inner_latitude"`
		InnerLongitude float64                     `json:"inner_longitude"`
		InnerJDUT      float64                     `json:"inner_jd_ut"`
		InnerPlanets   []models.PlanetID           `json:"inner_planets"`
		OuterLatitude  float64                     `json:"outer_latitude"`
		OuterLongitude float64                     `json:"outer_longitude"`
		OuterJDUT      float64                     `json:"outer_jd_ut"`
		OuterPlanets   []models.PlanetID           `json:"outer_planets"`
		SpecialPoints  *models.SpecialPointsConfig `json:"special_points"`
		HouseSystem    models.HouseSystem          `json:"house_system"`
		OrbConfig      *models.OrbConfig           `json:"orb_config"`
		Format         string                      `json:"format"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.InnerPlanets) == 0 {
		req.InnerPlanets = defaultPlanets
	}
	if len(req.OuterPlanets) == 0 {
		req.OuterPlanets = defaultPlanets
	}
	if req.HouseSystem == "" {
		req.HouseSystem = models.HousePlacidus
	}
	orbs := orbOrDefault(req.OrbConfig, models.DefaultOrbConfig())

	innerChart, outerChart, crossAspects, err := chart.CalcDoubleChart(
		req.InnerLatitude, req.InnerLongitude, req.InnerJDUT, req.InnerPlanets,
		req.OuterLatitude, req.OuterLongitude, req.OuterJDUT, req.OuterPlanets,
		req.SpecialPoints, orbs, req.HouseSystem,
	)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if req.Format == "csv" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"format":        "csv",
			"inner_planets": export.ChartToCSV(innerChart),
			"outer_planets": export.ChartToCSV(outerChart),
			"cross_aspects": export.CrossAspectsToCSV(crossAspects),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"inner_chart":   innerChart,
		"outer_chart":   outerChart,
		"cross_aspects": crossAspects,
	})
}

func (s *Server) handleCompositeChart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Person1Lat  float64            `json:"person1_latitude"`
		Person1Lon  float64            `json:"person1_longitude"`
		Person1JD   float64            `json:"person1_jd_ut"`
		Person2Lat  float64            `json:"person2_latitude"`
		Person2Lon  float64            `json:"person2_longitude"`
		Person2JD   float64            `json:"person2_jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		OrbConfig   *models.OrbConfig  `json:"orb_config"`
		Format      string             `json:"format"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Planets) == 0 {
		req.Planets = defaultPlanets
	}
	if req.HouseSystem == "" {
		req.HouseSystem = models.HousePlacidus
	}

	cc, err := composite.CalcCompositeChart(composite.CompositeInput{
		Person1Lat: req.Person1Lat, Person1Lon: req.Person1Lon, Person1JD: req.Person1JD,
		Person2Lat: req.Person2Lat, Person2Lon: req.Person2Lon, Person2JD: req.Person2JD,
		Planets:     req.Planets,
		OrbConfig:   orbOrDefault(req.OrbConfig, models.DefaultOrbConfig()),
		HouseSystem: req.HouseSystem,
	})
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if req.Format == "csv" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"format":  "csv",
			"planets": export.ChartToCSV(&models.ChartInfo{Planets: cc.Planets}),
			"aspects": export.AspectsToCSV(cc.Aspects),
			"houses":  export.HousesToCSV(cc.Houses, cc.Angles),
		})
		return
	}
	writeJSON(w, http.StatusOK, cc)
}

func (s *Server) handleDavisonChart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Person1Lat  float64            `json:"person1_latitude"`
		Person1Lon  float64            `json:"person1_longitude"`
		Person1JD   float64            `json:"person1_jd_ut"`
		Person2Lat  float64            `json:"person2_latitude"`
		Person2Lon  float64            `json:"person2_longitude"`
		Person2JD   float64            `json:"person2_jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.HouseSystem == "" {
		req.HouseSystem = models.HousePlacidus
	}

	result, err := composite.CalcDavisonChart(composite.CompositeInput{
		Person1Lat: req.Person1Lat, Person1Lon: req.Person1Lon, Person1JD: req.Person1JD,
		Person2Lat: req.Person2Lat, Person2Lon: req.Person2Lon, Person2JD: req.Person2JD,
		Planets:     req.Planets,
		OrbConfig:   models.DefaultOrbConfig(),
		HouseSystem: req.HouseSystem,
	})
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// --- Predictive Endpoints ---

func (s *Server) handleTransit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NatalLatitude         float64                     `json:"natal_latitude"`
		NatalLongitude        float64                     `json:"natal_longitude"`
		NatalJDUT             float64                     `json:"natal_jd_ut"`
		NatalPlanets          []models.PlanetID           `json:"natal_planets"`
		TransitLatitude       float64                     `json:"transit_latitude"`
		TransitLongitude      float64                     `json:"transit_longitude"`
		StartJDUT             float64                     `json:"start_jd_ut"`
		EndJDUT               float64                     `json:"end_jd_ut"`
		TransitPlanets        []models.PlanetID           `json:"transit_planets"`
		ProgressionsConfig    *models.ProgressionsConfig  `json:"progressions_config"`
		SolarArcConfig        *models.SolarArcConfig      `json:"solar_arc_config"`
		SpecialPoints         *models.SpecialPointsConfig `json:"special_points"`
		EventConfig           *models.EventConfig         `json:"event_config"`
		HouseSystem           models.HouseSystem          `json:"house_system"`
		OrbConfig             *models.OrbConfig           `json:"orb_config"`
		OrbConfigTransit      *models.OrbConfig           `json:"orb_config_transit"`
		OrbConfigProgressions *models.OrbConfig           `json:"orb_config_progressions"`
		OrbConfigSolarArc     *models.OrbConfig           `json:"orb_config_solar_arc"`
		Timezone              string                      `json:"timezone"`
		Format                string                      `json:"format"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.NatalPlanets) == 0 {
		req.NatalPlanets = defaultPlanets
	}
	if len(req.TransitPlanets) == 0 {
		req.TransitPlanets = defaultPlanets
	}
	if req.HouseSystem == "" {
		req.HouseSystem = models.HousePlacidus
	}
	if req.Timezone == "" {
		req.Timezone = "UTC"
	}
	baseOrbs := orbOrDefault(req.OrbConfig, models.DefaultOrbConfig())
	eventCfg := models.DefaultEventConfig()
	if req.EventConfig != nil {
		eventCfg = *req.EventConfig
	}

	// Build EventFilterConfig from request EventConfig
	eventFilter := transit.EventFilterConfig{
		TrNa:        eventCfg.IncludeTrNa,
		TrTr:        eventCfg.IncludeTrTr,
		TrSp:        eventCfg.IncludeTrSp,
		TrSa:        eventCfg.IncludeTrSa,
		SpNa:        eventCfg.IncludeSpNa,
		SpSp:        eventCfg.IncludeSpSp,
		SaNa:        eventCfg.IncludeSaNa,
		Station:     eventCfg.IncludeStation,
		SignIngress: eventCfg.IncludeSignIngress,
		HouseIngress: eventCfg.IncludeHouseIngress,
		VoidOfCourse: eventCfg.IncludeVoidOfCourse,
	}

	// Build structured calcInput
	calcInput := transit.TransitCalcInput{
		NatalChart: transit.NatalChartConfig{
			Lat:     req.NatalLatitude,
			Lon:     req.NatalLongitude,
			JD:      req.NatalJDUT,
			Planets: req.NatalPlanets,
		},
		TimeRange: transit.TimeRangeConfig{
			StartJD: req.StartJDUT,
			EndJD:   req.EndJDUT,
		},
		Charts: transit.ChartSetConfig{
			Transit: &transit.TransitChartConfig{
				Lat:         req.TransitLatitude,
				Lon:         req.TransitLongitude,
				Planets:     req.TransitPlanets,
				Orbs:        orbOrDefault(req.OrbConfigTransit, baseOrbs),
				HouseSystem: req.HouseSystem,
				Points:      nil, // populated from SpecialPoints if needed
			},
			Progressions: func() *transit.ProgressionsChartConfig {
				if req.ProgressionsConfig != nil {
					return &transit.ProgressionsChartConfig{
						Planets: req.ProgressionsConfig.Planets,
						Points:  nil, // populated from SpecialPoints if needed
						Orbs:    orbOrDefault(req.OrbConfigProgressions, baseOrbs),
					}
				}
				return nil
			}(),
			SolarArc: func() *transit.SolarArcChartConfig {
				if req.SolarArcConfig != nil {
					return &transit.SolarArcChartConfig{
						Planets:     req.SolarArcConfig.Planets,
						Points:      nil, // populated from SpecialPoints if needed
						Orbs:        orbOrDefault(req.OrbConfigSolarArc, baseOrbs),
						Lat:         req.TransitLatitude,
						Lon:         req.TransitLongitude,
						HouseSystem: req.HouseSystem,
					}
				}
				return nil
			}(),
		},
		EventFilter: eventFilter,
		// Keep old flat fields for backward compatibility with transit package
		NatalLat:      req.NatalLatitude,
		NatalLon:      req.NatalLongitude,
		NatalJD:       req.NatalJDUT,
		NatalPlanets:  req.NatalPlanets,
	}

	events, err := transit.CalcTransitEvents(calcInput)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if req.Format == "csv" {
		csv := export.EventsToCSV(events, req.Timezone, "")
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"format":      "csv",
			"event_count": len(events),
			"csv":         csv,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
	})
}

// directedPlanet holds a planet position from progressions or solar arc.
type directedPlanet struct {
	PlanetID     models.PlanetID `json:"planet_id"`
	Longitude    float64         `json:"longitude"`
	Speed        float64         `json:"speed"`
	IsRetrograde bool            `json:"is_retrograde,omitempty"`
	Sign         string          `json:"sign"`
	SignDegree   float64         `json:"sign_degree"`
}

type planetCalcFunc func(pid models.PlanetID, natalJD, transitJD float64) (float64, float64, error)

func calcDirectedPlanets(planets []models.PlanetID, natalJD, transitJD float64, calc planetCalcFunc) []directedPlanet {
	var result []directedPlanet
	for _, pid := range planets {
		lon, speed, err := calc(pid, natalJD, transitJD)
		if err != nil {
			continue
		}
		result = append(result, directedPlanet{
			PlanetID:     pid,
			Longitude:    lon,
			Speed:        speed,
			IsRetrograde: speed < 0,
			Sign:         models.SignFromLongitude(lon),
			SignDegree:   models.SignDegreeFromLongitude(lon),
		})
	}
	return result
}

func directedToPositions(dp []directedPlanet) []models.PlanetPosition {
	positions := make([]models.PlanetPosition, len(dp))
	for i, d := range dp {
		positions[i] = models.PlanetPosition{
			PlanetID: d.PlanetID, Longitude: d.Longitude, Speed: d.Speed,
			IsRetrograde: d.IsRetrograde, Sign: d.Sign, SignDegree: d.SignDegree,
		}
	}
	return positions
}

func (s *Server) handleProgressions(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NatalJDUT   float64           `json:"natal_jd_ut"`
		TransitJDUT float64           `json:"transit_jd_ut"`
		Planets     []models.PlanetID `json:"planets"`
		Format      string            `json:"format"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Planets) == 0 {
		req.Planets = defaultPlanets
	}

	dp := calcDirectedPlanets(req.Planets, req.NatalJDUT, req.TransitJDUT, progressions.CalcProgressedLongitude)

	if req.Format == "csv" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"format":  "csv",
			"age":     progressions.Age(req.NatalJDUT, req.TransitJDUT),
			"planets": export.PositionsToCSV(directedToPositions(dp)),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"age":           progressions.Age(req.NatalJDUT, req.TransitJDUT),
		"progressed_jd": progressions.SecondaryProgressionJD(req.NatalJDUT, req.TransitJDUT),
		"planets":       dp,
	})
}

func (s *Server) handleSolarArc(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NatalJDUT   float64           `json:"natal_jd_ut"`
		TransitJDUT float64           `json:"transit_jd_ut"`
		Planets     []models.PlanetID `json:"planets"`
		Format      string            `json:"format"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Planets) == 0 {
		req.Planets = defaultPlanets
	}

	offset, err := progressions.SolarArcOffset(req.NatalJDUT, req.TransitJDUT)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	dp := calcDirectedPlanets(req.Planets, req.NatalJDUT, req.TransitJDUT, progressions.CalcSolarArcLongitude)

	if req.Format == "csv" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"format":           "csv",
			"age":              progressions.Age(req.NatalJDUT, req.TransitJDUT),
			"solar_arc_offset": offset,
			"planets":          export.PositionsToCSV(directedToPositions(dp)),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"age":              progressions.Age(req.NatalJDUT, req.TransitJDUT),
		"solar_arc_offset": offset,
		"planets":          dp,
	})
}

func (s *Server) handleSolarReturn(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NatalJDUT   float64            `json:"natal_jd_ut"`
		NatalLat    float64            `json:"natal_latitude"`
		NatalLon    float64            `json:"natal_longitude"`
		SearchJDUT  float64            `json:"search_jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		Count       int                `json:"count"`
		Format      string             `json:"format"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Planets) == 0 {
		req.Planets = defaultPlanets
	}
	if req.HouseSystem == "" {
		req.HouseSystem = models.HousePlacidus
	}
	if req.Count <= 0 {
		req.Count = 1
	}

	ri := returns.ReturnInput{
		NatalJD: req.NatalJDUT, NatalLat: req.NatalLat, NatalLon: req.NatalLon,
		SearchJD: req.SearchJDUT, Planets: req.Planets,
		OrbConfig: models.DefaultOrbConfig(), HouseSystem: req.HouseSystem,
	}

	if req.Count == 1 {
		rc, err := returns.CalcSolarReturn(ri)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		if req.Format == "csv" && rc.Chart != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"format":    "csv",
				"return_jd": rc.ReturnJD,
				"age":       rc.Age,
				"planets":   export.ChartToCSV(rc.Chart),
				"aspects":   export.AspectsToCSV(rc.Chart.Aspects),
				"houses":    export.HousesToCSV(rc.Chart.Houses, rc.Chart.Angles),
			})
			return
		}
		writeJSON(w, http.StatusOK, rc)
		return
	}

	result, err := returns.CalcSolarReturnSeries(ri, req.Count)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleLunarReturn(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NatalJDUT   float64            `json:"natal_jd_ut"`
		NatalLat    float64            `json:"natal_latitude"`
		NatalLon    float64            `json:"natal_longitude"`
		SearchJDUT  float64            `json:"search_jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		Count       int                `json:"count"`
		Format      string             `json:"format"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Planets) == 0 {
		req.Planets = defaultPlanets
	}
	if req.HouseSystem == "" {
		req.HouseSystem = models.HousePlacidus
	}
	if req.Count <= 0 {
		req.Count = 1
	}

	ri := returns.ReturnInput{
		NatalJD: req.NatalJDUT, NatalLat: req.NatalLat, NatalLon: req.NatalLon,
		SearchJD: req.SearchJDUT, Planets: req.Planets,
		OrbConfig: models.DefaultOrbConfig(), HouseSystem: req.HouseSystem,
	}

	if req.Count == 1 {
		rc, err := returns.CalcLunarReturn(ri)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		if req.Format == "csv" && rc.Chart != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"format":    "csv",
				"return_jd": rc.ReturnJD,
				"age":       rc.Age,
				"planets":   export.ChartToCSV(rc.Chart),
				"aspects":   export.AspectsToCSV(rc.Chart.Aspects),
				"houses":    export.HousesToCSV(rc.Chart.Houses, rc.Chart.Angles),
			})
			return
		}
		writeJSON(w, http.StatusOK, rc)
		return
	}

	result, err := returns.CalcLunarReturnSeries(ri, req.Count)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// --- Traditional Endpoints ---

func (s *Server) handleDignity(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Latitude    float64            `json:"latitude"`
		Longitude   float64            `json:"longitude"`
		JDUT        float64            `json:"jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		Format      string             `json:"format"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Planets) == 0 {
		req.Planets = defaultPlanets
	}
	if req.HouseSystem == "" {
		req.HouseSystem = models.HousePlacidus
	}

	chartInfo, err := chart.CalcSingleChart(req.Latitude, req.Longitude, req.JDUT,
		req.Planets, models.DefaultOrbConfig(), req.HouseSystem)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	dignities := dignity.CalcChartDignities(chartInfo.Planets)
	receptions := dignity.FindMutualReceptions(chartInfo.Planets)

	isDayChart := chart.IsDayChart(req.JDUT, chartInfo.Angles.ASC)
	var sects []dignity.SectInfo
	for _, p := range chartInfo.Planets {
		sects = append(sects, dignity.CalcSect(p.PlanetID, isDayChart))
	}

	if req.Format == "csv" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"format":    "csv",
			"dignities": export.DignityToCSV(dignities),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"dignities":         dignities,
		"mutual_receptions": receptions,
		"sect":              sects,
		"is_day_chart":      isDayChart,
	})
}

func (s *Server) handleBonification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Latitude    float64            `json:"latitude"`
		Longitude   float64            `json:"longitude"`
		JDUT        float64            `json:"jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Planets) == 0 {
		req.Planets = defaultPlanets
	}
	if req.HouseSystem == "" {
		req.HouseSystem = models.HousePlacidus
	}

	chartInfo, err := chart.CalcSingleChart(req.Latitude, req.Longitude, req.JDUT,
		req.Planets, models.DefaultOrbConfig(), req.HouseSystem)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, dignity.CalcChartBonMal(chartInfo.Planets))
}

// --- Analysis Endpoints ---

func (s *Server) handleAspectPatterns(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Latitude    float64            `json:"latitude"`
		Longitude   float64            `json:"longitude"`
		JDUT        float64            `json:"jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		OrbConfig   *models.OrbConfig  `json:"orb_config"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Planets) == 0 {
		req.Planets = defaultPlanets
	}
	if req.HouseSystem == "" {
		req.HouseSystem = models.HousePlacidus
	}
	orbs := orbOrDefault(req.OrbConfig, models.DefaultOrbConfig())

	chartInfo, err := chart.CalcSingleChart(req.Latitude, req.Longitude, req.JDUT,
		req.Planets, orbs, req.HouseSystem)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	var bodies []aspect.Body
	for _, p := range chartInfo.Planets {
		bodies = append(bodies, aspect.Body{
			ID: string(p.PlanetID), Longitude: p.Longitude, Speed: p.Speed,
		})
	}

	patterns := aspect.FindPatterns(chartInfo.Aspects, bodies, orbs)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"patterns": patterns,
		"aspects":  chartInfo.Aspects,
	})
}

func (s *Server) handleSynastry(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Person1Lat  float64            `json:"person1_latitude"`
		Person1Lon  float64            `json:"person1_longitude"`
		Person1JD   float64            `json:"person1_jd_ut"`
		Person2Lat  float64            `json:"person2_latitude"`
		Person2Lon  float64            `json:"person2_longitude"`
		Person2JD   float64            `json:"person2_jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		OrbConfig   *models.OrbConfig  `json:"orb_config"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Planets) == 0 {
		req.Planets = defaultPlanets
	}
	if req.HouseSystem == "" {
		req.HouseSystem = models.HousePlacidus
	}
	orbs := orbOrDefault(req.OrbConfig, models.DefaultOrbConfig())

	chart1, err := chart.CalcSingleChart(req.Person1Lat, req.Person1Lon, req.Person1JD,
		req.Planets, orbs, req.HouseSystem)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, fmt.Sprintf("person 1 chart: %v", err))
		return
	}
	chart2, err := chart.CalcSingleChart(req.Person2Lat, req.Person2Lon, req.Person2JD,
		req.Planets, orbs, req.HouseSystem)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, fmt.Sprintf("person 2 chart: %v", err))
		return
	}

	score := synastry.CalcSynastryFromCharts(chart1.Planets, chart2.Planets, orbs)
	writeJSON(w, http.StatusOK, score)
}

// --- Lunar Endpoints ---

func (s *Server) handleLunarPhase(w http.ResponseWriter, r *http.Request) {
	var req struct {
		JDUT float64 `json:"jd_ut"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := lunar.CalcLunarPhase(req.JDUT)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleLunarPhases(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StartJDUT float64 `json:"start_jd_ut"`
		EndJDUT   float64 `json:"end_jd_ut"`
		Format    string  `json:"format"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	phases, err := lunar.FindLunarPhases(req.StartJDUT, req.EndJDUT)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if req.Format == "csv" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"format": "csv",
			"phases": export.LunarPhasesToCSV(phases),
			"count":  len(phases),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"phases": phases,
		"count":  len(phases),
	})
}

func (s *Server) handleEclipses(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StartJDUT float64 `json:"start_jd_ut"`
		EndJDUT   float64 `json:"end_jd_ut"`
		Format    string  `json:"format"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	eclipses, err := lunar.FindEclipses(req.StartJDUT, req.EndJDUT)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	if req.Format == "csv" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"format":   "csv",
			"eclipses": export.EclipsesToCSV(eclipses),
			"count":    len(eclipses),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"eclipses": eclipses,
		"count":    len(eclipses),
	})
}
