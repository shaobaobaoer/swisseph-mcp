package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/anthropic/swisseph-mcp/pkg/chart"
	"github.com/anthropic/swisseph-mcp/pkg/export"
	"github.com/anthropic/swisseph-mcp/pkg/geo"
	"github.com/anthropic/swisseph-mcp/pkg/julian"
	"github.com/anthropic/swisseph-mcp/pkg/models"
	"github.com/anthropic/swisseph-mcp/pkg/progressions"
	"github.com/anthropic/swisseph-mcp/pkg/sweph"
	"github.com/anthropic/swisseph-mcp/pkg/transit"
)

// Server implements the MCP protocol via JSON-RPC over stdio
type Server struct {
	ephePath string
}

// NewServer creates a new MCP server
func NewServer(ephePath string) *Server {
	return &Server{ephePath: ephePath}
}

// JSON-RPC structures
type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP protocol structures
type initializeResult struct {
	ProtocolVersion string     `json:"protocolVersion"`
	Capabilities    capability `json:"capabilities"`
	ServerInfo      serverInfo `json:"serverInfo"`
}

type capability struct {
	Tools *toolsCap `json:"tools,omitempty"`
}

type toolsCap struct {
	ListChanged bool `json:"listChanged"`
}

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type toolsListResult struct {
	Tools []toolDef `json:"tools"`
}

type toolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

type callToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type callToolResult struct {
	Content []contentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type contentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Run starts the MCP server, reading from stdin and writing to stdout
func (s *Server) Run() error {
	// Initialize Swiss Ephemeris
	absPath, _ := filepath.Abs(s.ephePath)
	sweph.Init(absPath)
	defer sweph.Close()

	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var req jsonRPCRequest
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				return nil
			}
			continue
		}

		resp := s.handleRequest(&req)
		if resp != nil {
			encoder.Encode(resp)
		}
	}
}

func (s *Server) handleRequest(req *jsonRPCRequest) *jsonRPCResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "notifications/initialized":
		return nil // notification, no response
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	default:
		return &jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: -32601, Message: "method not found: " + req.Method},
		}
	}
}

func (s *Server) handleInitialize(req *jsonRPCRequest) *jsonRPCResponse {
	return &jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: initializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: capability{
				Tools: &toolsCap{ListChanged: false},
			},
			ServerInfo: serverInfo{
				Name:    "swisseph-mcp",
				Version: "1.1.0",
			},
		},
	}
}

func (s *Server) handleToolsList(req *jsonRPCRequest) *jsonRPCResponse {
	tools := []toolDef{
		{
			Name:        "geocode",
			Description: "Returns geographic coordinates (lat/lon) and timezone for a location name",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"location_name": {"type": "string", "description": "Location name"}
				},
				"required": ["location_name"]
			}`),
		},
		{
			Name:        "datetime_to_jd",
			Description: "Converts ISO 8601 datetime to Julian Day (UT and TT)",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"datetime": {"type": "string", "description": "ISO 8601 datetime string"},
					"calendar": {"type": "string", "enum": ["GREGORIAN", "JULIAN"], "default": "GREGORIAN"}
				},
				"required": ["datetime"]
			}`),
		},
		{
			Name:        "jd_to_datetime",
			Description: "Converts Julian Day to ISO 8601 datetime string",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"jd": {"type": "number", "description": "Julian Day number"},
					"timezone": {"type": "string", "default": "UTC", "description": "Target timezone"}
				},
				"required": ["jd"]
			}`),
		},
		{
			Name:        "calc_planet_position",
			Description: "Calculate a single planet's ecliptic longitude, latitude, speed, retrograde status and sign",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"planet": {"type": "string", "description": "Planet ID: SUN, MOON, MERCURY, VENUS, MARS, JUPITER, SATURN, URANUS, NEPTUNE, PLUTO, CHIRON, NORTH_NODE_TRUE, etc."},
					"jd_ut": {"type": "number", "description": "Julian Day (UT)"}
				},
				"required": ["planet", "jd_ut"]
			}`),
		},
		{
			Name:        "calc_single_chart",
			Description: "Single chart calculation: compute planet positions, houses, and aspects at a fixed time",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"},
					"orb_config": {"type": "object"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		{
			Name:        "calc_double_chart",
			Description: "Double chart calculation: compute inner/outer chart positions and cross-aspects",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"inner_latitude": {"type": "number"},
					"inner_longitude": {"type": "number"},
					"inner_jd_ut": {"type": "number"},
					"inner_planets": {"type": "array", "items": {"type": "string"}},
					"outer_latitude": {"type": "number"},
					"outer_longitude": {"type": "number"},
					"outer_jd_ut": {"type": "number"},
					"outer_planets": {"type": "array", "items": {"type": "string"}},
					"special_points": {"type": "object"},
					"house_system": {"type": "string", "default": "PLACIDUS"},
					"orb_config": {"type": "object"}
				},
				"required": ["inner_latitude", "inner_longitude", "inner_jd_ut",
					"outer_latitude", "outer_longitude", "outer_jd_ut"]
			}`),
		},
		{
			Name:        "calc_progressions",
			Description: "Secondary progressions: compute progressed planet positions (1 day = 1 year)",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"natal_jd_ut": {"type": "number", "description": "Natal Julian Day (UT)"},
					"transit_jd_ut": {"type": "number", "description": "Transit Julian Day (UT)"},
					"planets": {"type": "array", "items": {"type": "string"}, "description": "Planet list"}
				},
				"required": ["natal_jd_ut", "transit_jd_ut"]
			}`),
		},
		{
			Name:        "calc_solar_arc",
			Description: "Solar arc directions: compute solar arc directed planet positions",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"natal_jd_ut": {"type": "number", "description": "Natal Julian Day (UT)"},
					"transit_jd_ut": {"type": "number", "description": "Transit Julian Day (UT)"},
					"planets": {"type": "array", "items": {"type": "string"}, "description": "Planet list"}
				},
				"required": ["natal_jd_ut", "transit_jd_ut"]
			}`),
		},
		{
			Name:        "calc_transit",
			Description: "Transit calculation: search for all astrological events between transit/progressed/solar-arc and natal bodies over a time range. Supports Tr-Na/Tr-Tr/Tr-Sp/Tr-Sa/Sp-Na/Sp-Sp/Sa-Na, sign/house ingress, stations, void of course.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"natal_latitude": {"type": "number"},
					"natal_longitude": {"type": "number"},
					"natal_jd_ut": {"type": "number"},
					"natal_planets": {"type": "array", "items": {"type": "string"}},
					"transit_latitude": {"type": "number"},
					"transit_longitude": {"type": "number"},
					"start_jd_ut": {"type": "number"},
					"end_jd_ut": {"type": "number"},
					"transit_planets": {"type": "array", "items": {"type": "string"}},
					"progressions_config": {"type": "object", "properties": {
						"enabled": {"type": "boolean"},
						"planets": {"type": "array", "items": {"type": "string"}}
					}},
					"solar_arc_config": {"type": "object", "properties": {
						"enabled": {"type": "boolean"},
						"planets": {"type": "array", "items": {"type": "string"}}
					}},
					"special_points": {"type": "object"},
					"event_config": {"type": "object"},
					"house_system": {"type": "string", "default": "PLACIDUS"},
					"orb_config_transit": {"type": "object"},
					"orb_config_progressions": {"type": "object"},
					"orb_config_solar_arc": {"type": "object"},
					"format": {"type": "string", "enum": ["json", "csv"], "description": "Output format: json (default) or csv (Solar Fire compatible)"},
					"timezone": {"type": "string", "default": "UTC", "description": "Timezone for CSV date/time output"}
				},
				"required": ["natal_latitude", "natal_longitude", "natal_jd_ut",
					"transit_latitude", "transit_longitude", "start_jd_ut", "end_jd_ut"]
			}`),
		},
	}

	return &jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  toolsListResult{Tools: tools},
	}
}

func (s *Server) handleToolsCall(req *jsonRPCRequest) *jsonRPCResponse {
	var params callToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errorResponse(req.ID, -32602, "invalid params")
	}

	var result interface{}
	var err error

	switch params.Name {
	case "geocode":
		result, err = s.handleGeocode(params.Arguments)
	case "datetime_to_jd":
		result, err = s.handleDatetimeToJD(params.Arguments)
	case "jd_to_datetime":
		result, err = s.handleJDToDatetime(params.Arguments)
	case "calc_planet_position":
		result, err = s.handleCalcPlanetPosition(params.Arguments)
	case "calc_single_chart":
		result, err = s.handleCalcSingleChart(params.Arguments)
	case "calc_double_chart":
		result, err = s.handleCalcDoubleChart(params.Arguments)
	case "calc_progressions":
		result, err = s.handleCalcProgressions(params.Arguments)
	case "calc_solar_arc":
		result, err = s.handleCalcSolarArc(params.Arguments)
	case "calc_transit":
		result, err = s.handleCalcTransit(params.Arguments)
	default:
		return errorResponse(req.ID, -32601, "unknown tool: "+params.Name)
	}

	if err != nil {
		return &jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: callToolResult{
				Content: []contentItem{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
				IsError: true,
			},
		}
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return &jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: callToolResult{
			Content: []contentItem{{Type: "text", Text: string(jsonBytes)}},
		},
	}
}

func errorResponse(id interface{}, code int, msg string) *jsonRPCResponse {
	return &jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: msg},
	}
}

// defaultPlanets is the default planet list used across all handlers
var defaultPlanets = []models.PlanetID{
	models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
	models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
	models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune,
	models.PlanetPluto,
}

// === Tool handlers ===

func (s *Server) handleCalcPlanetPosition(args json.RawMessage) (interface{}, error) {
	var input struct {
		Planet models.PlanetID `json:"planet"`
		JDUT   float64         `json:"jd_ut"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	lon, speed, err := chart.CalcPlanetLongitude(input.Planet, input.JDUT)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"planet":       input.Planet,
		"longitude":    lon,
		"speed":        speed,
		"is_retrograde": speed < 0,
		"sign":         models.SignFromLongitude(lon),
		"sign_degree":  models.SignDegreeFromLongitude(lon),
	}, nil
}

func (s *Server) handleGeocode(args json.RawMessage) (interface{}, error) {
	var input struct {
		LocationName string `json:"location_name"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	return geo.Geocode(input.LocationName)
}

func (s *Server) handleDatetimeToJD(args json.RawMessage) (interface{}, error) {
	var input struct {
		Datetime string             `json:"datetime"`
		Calendar models.CalendarType `json:"calendar"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	if input.Calendar == "" {
		input.Calendar = models.CalendarGregorian
	}
	return julian.DateTimeToJD(input.Datetime, input.Calendar)
}

func (s *Server) handleJDToDatetime(args json.RawMessage) (interface{}, error) {
	var input struct {
		JD       float64 `json:"jd"`
		Timezone string  `json:"timezone"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	if input.Timezone == "" {
		input.Timezone = "UTC"
	}
	dt, err := julian.JDToDateTime(input.JD, input.Timezone)
	if err != nil {
		return nil, err
	}
	return map[string]string{"datetime": dt}, nil
}

func (s *Server) handleCalcSingleChart(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude    float64          `json:"latitude"`
		Longitude   float64          `json:"longitude"`
		JDUT        float64          `json:"jd_ut"`
		Planets     []models.PlanetID `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		OrbConfig   *models.OrbConfig  `json:"orb_config"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	if len(input.Planets) == 0 {
		input.Planets = defaultPlanets
	}
	if input.HouseSystem == "" {
		input.HouseSystem = models.HousePlacidus
	}
	orbs := models.DefaultOrbConfig()
	if input.OrbConfig != nil {
		orbs = *input.OrbConfig
	}

	return chart.CalcSingleChart(input.Latitude, input.Longitude, input.JDUT, input.Planets, orbs, input.HouseSystem)
}

func (s *Server) handleCalcDoubleChart(args json.RawMessage) (interface{}, error) {
	var input struct {
		InnerLatitude  float64              `json:"inner_latitude"`
		InnerLongitude float64              `json:"inner_longitude"`
		InnerJDUT      float64              `json:"inner_jd_ut"`
		InnerPlanets   []models.PlanetID    `json:"inner_planets"`
		OuterLatitude  float64              `json:"outer_latitude"`
		OuterLongitude float64              `json:"outer_longitude"`
		OuterJDUT      float64              `json:"outer_jd_ut"`
		OuterPlanets   []models.PlanetID    `json:"outer_planets"`
		SpecialPoints  *models.SpecialPointsConfig `json:"special_points"`
		HouseSystem    models.HouseSystem    `json:"house_system"`
		OrbConfig      *models.OrbConfig     `json:"orb_config"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	if len(input.InnerPlanets) == 0 {
		input.InnerPlanets = defaultPlanets
	}
	if len(input.OuterPlanets) == 0 {
		input.OuterPlanets = defaultPlanets
	}
	if input.HouseSystem == "" {
		input.HouseSystem = models.HousePlacidus
	}
	orbs := models.DefaultOrbConfig()
	if input.OrbConfig != nil {
		orbs = *input.OrbConfig
	}

	innerChart, outerChart, crossAspects, err := chart.CalcDoubleChart(
		input.InnerLatitude, input.InnerLongitude, input.InnerJDUT, input.InnerPlanets,
		input.OuterLatitude, input.OuterLongitude, input.OuterJDUT, input.OuterPlanets,
		input.SpecialPoints, orbs, input.HouseSystem,
	)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"inner_chart":   innerChart,
		"outer_chart":   outerChart,
		"cross_aspects": crossAspects,
	}, nil
}

func (s *Server) handleCalcTransit(args json.RawMessage) (interface{}, error) {
	calcInput, tz, err := s.buildTransitInput(args)
	if err != nil {
		return nil, err
	}

	// Check format parameter
	var formatInput struct {
		Format string `json:"format"`
	}
	json.Unmarshal(args, &formatInput)

	events, err := transit.CalcTransitEvents(calcInput)
	if err != nil {
		return nil, err
	}

	switch formatInput.Format {
	case "csv":
		csv := export.EventsToCSV(events, tz, "")
		return map[string]interface{}{
			"format":      "csv",
			"event_count": len(events),
			"csv":         csv,
		}, nil
	default:
		return map[string]interface{}{
			"events": events,
		}, nil
	}
}

func (s *Server) handleCalcProgressions(args json.RawMessage) (interface{}, error) {
	var input struct {
		NatalJDUT   float64           `json:"natal_jd_ut"`
		TransitJDUT float64           `json:"transit_jd_ut"`
		Planets     []models.PlanetID `json:"planets"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	if len(input.Planets) == 0 {
		input.Planets = defaultPlanets
	}

	age := progressions.Age(input.NatalJDUT, input.TransitJDUT)
	progressedJD := progressions.SecondaryProgressionJD(input.NatalJDUT, input.TransitJDUT)

	type progressedPlanet struct {
		PlanetID     models.PlanetID `json:"planet_id"`
		Longitude    float64         `json:"longitude"`
		Speed        float64         `json:"speed"`
		IsRetrograde bool            `json:"is_retrograde"`
		Sign         string          `json:"sign"`
		SignDegree   float64         `json:"sign_degree"`
	}

	var planets []progressedPlanet
	for _, pid := range input.Planets {
		lon, speed, err := progressions.CalcProgressedLongitude(pid, input.NatalJDUT, input.TransitJDUT)
		if err != nil {
			continue
		}
		planets = append(planets, progressedPlanet{
			PlanetID:     pid,
			Longitude:    lon,
			Speed:        speed,
			IsRetrograde: speed < 0,
			Sign:         models.SignFromLongitude(lon),
			SignDegree:   models.SignDegreeFromLongitude(lon),
		})
	}

	return map[string]interface{}{
		"age":           age,
		"progressed_jd": progressedJD,
		"planets":       planets,
	}, nil
}

func (s *Server) handleCalcSolarArc(args json.RawMessage) (interface{}, error) {
	var input struct {
		NatalJDUT   float64           `json:"natal_jd_ut"`
		TransitJDUT float64           `json:"transit_jd_ut"`
		Planets     []models.PlanetID `json:"planets"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	if len(input.Planets) == 0 {
		input.Planets = defaultPlanets
	}

	age := progressions.Age(input.NatalJDUT, input.TransitJDUT)
	offset, err := progressions.SolarArcOffset(input.NatalJDUT, input.TransitJDUT)
	if err != nil {
		return nil, err
	}

	type solarArcPlanet struct {
		PlanetID   models.PlanetID `json:"planet_id"`
		Longitude  float64         `json:"longitude"`
		Speed      float64         `json:"speed"`
		Sign       string          `json:"sign"`
		SignDegree float64         `json:"sign_degree"`
	}

	var planets []solarArcPlanet
	for _, pid := range input.Planets {
		lon, speed, err := progressions.CalcSolarArcLongitude(pid, input.NatalJDUT, input.TransitJDUT)
		if err != nil {
			continue
		}
		planets = append(planets, solarArcPlanet{
			PlanetID:   pid,
			Longitude:  lon,
			Speed:      speed,
			Sign:       models.SignFromLongitude(lon),
			SignDegree: models.SignDegreeFromLongitude(lon),
		})
	}

	return map[string]interface{}{
		"age":              age,
		"solar_arc_offset": offset,
		"planets":          planets,
	}, nil
}

// buildTransitInput extracts common transit input from JSON arguments
func (s *Server) buildTransitInput(args json.RawMessage) (transit.TransitCalcInput, string, error) {
	var input struct {
		NatalLatitude      float64                      `json:"natal_latitude"`
		NatalLongitude     float64                      `json:"natal_longitude"`
		NatalJDUT          float64                      `json:"natal_jd_ut"`
		NatalPlanets       []models.PlanetID            `json:"natal_planets"`
		TransitLatitude    float64                      `json:"transit_latitude"`
		TransitLongitude   float64                      `json:"transit_longitude"`
		StartJDUT          float64                      `json:"start_jd_ut"`
		EndJDUT            float64                      `json:"end_jd_ut"`
		TransitPlanets     []models.PlanetID            `json:"transit_planets"`
		ProgressionsConfig *models.ProgressionsConfig   `json:"progressions_config"`
		SolarArcConfig     *models.SolarArcConfig       `json:"solar_arc_config"`
		SpecialPoints      *models.SpecialPointsConfig  `json:"special_points"`
		EventConfig        *models.EventConfig          `json:"event_config"`
		HouseSystem        models.HouseSystem           `json:"house_system"`
		OrbConfig             *models.OrbConfig         `json:"orb_config"`
		OrbConfigTransit      *models.OrbConfig         `json:"orb_config_transit"`
		OrbConfigProgressions *models.OrbConfig         `json:"orb_config_progressions"`
		OrbConfigSolarArc     *models.OrbConfig         `json:"orb_config_solar_arc"`
		Timezone              string                    `json:"timezone"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return transit.TransitCalcInput{}, "", err
	}

	if len(input.NatalPlanets) == 0 {
		input.NatalPlanets = defaultPlanets
	}
	if len(input.TransitPlanets) == 0 {
		input.TransitPlanets = defaultPlanets
	}
	if input.HouseSystem == "" {
		input.HouseSystem = models.HousePlacidus
	}
	if input.Timezone == "" {
		input.Timezone = "UTC"
	}
	baseOrbs := models.DefaultOrbConfig()
	if input.OrbConfig != nil {
		baseOrbs = *input.OrbConfig
	}
	orbsTransit := baseOrbs
	if input.OrbConfigTransit != nil {
		orbsTransit = *input.OrbConfigTransit
	}
	orbsProgressions := baseOrbs
	if input.OrbConfigProgressions != nil {
		orbsProgressions = *input.OrbConfigProgressions
	}
	orbsSolarArc := baseOrbs
	if input.OrbConfigSolarArc != nil {
		orbsSolarArc = *input.OrbConfigSolarArc
	}
	eventCfg := models.DefaultEventConfig()
	if input.EventConfig != nil {
		eventCfg = *input.EventConfig
	}

	return transit.TransitCalcInput{
		NatalLat:              input.NatalLatitude,
		NatalLon:              input.NatalLongitude,
		NatalJD:               input.NatalJDUT,
		NatalPlanets:          input.NatalPlanets,
		TransitLat:            input.TransitLatitude,
		TransitLon:            input.TransitLongitude,
		StartJD:               input.StartJDUT,
		EndJD:                 input.EndJDUT,
		TransitPlanets:        input.TransitPlanets,
		ProgressionsConfig:    input.ProgressionsConfig,
		SolarArcConfig:        input.SolarArcConfig,
		SpecialPoints:         input.SpecialPoints,
		EventConfig:           eventCfg,
		OrbConfigTransit:      orbsTransit,
		OrbConfigProgressions: orbsProgressions,
		OrbConfigSolarArc:     orbsSolarArc,
		HouseSystem:           input.HouseSystem,
	}, input.Timezone, nil
}

