package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	swisseph "github.com/shaobaobaoer/solarsage-mcp"
	"github.com/shaobaobaoer/solarsage-mcp/internal/aspect"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/antiscia"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/ashtakavarga"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/bounds"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/chart"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/composite"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/dignity"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/divisional"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/export"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/firdaria"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/fixedstars"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/heliacal"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/geo"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/harmonic"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/julian"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/dispositor"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/lots"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/lunar"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/midpoint"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/planetary"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/primary"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/profection"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/progressions"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/render"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/report"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/returns"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/synastry"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/symbolic"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/sweph"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/transit"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/vedic"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/yoga"
)

// Server implements the MCP protocol via JSON-RPC over stdio
type Server struct {
	ephePath string
}

// orbConfigSchema defines the JSON schema for orb configuration with entering/exiting support
var orbConfigSchema = json.RawMessage(`{
	"type": "object",
	"description": "Aspect orb configuration. Supports unified orbs or separate entering/exiting orbs.",
	"properties": {
		"conjunction": {"type": "number", "default": 8, "description": "Orb for conjunction (0°)"},
		"opposition": {"type": "number", "default": 8, "description": "Orb for opposition (180°)"},
		"trine": {"type": "number", "default": 7, "description": "Orb for trine (120°)"},
		"square": {"type": "number", "default": 7, "description": "Orb for square (90°)"},
		"sextile": {"type": "number", "default": 5, "description": "Orb for sextile (60°)"},
		"quincunx": {"type": "number", "default": 3, "description": "Orb for quincunx (150°)"},
		"semi_sextile": {"type": "number", "default": 2, "description": "Orb for semi-sextile (30°)"},
		"semi_square": {"type": "number", "default": 2, "description": "Orb for semi-square (45°)"},
		"sesquiquadrate": {"type": "number", "default": 2, "description": "Orb for sesquiquadrate (135°)"},
		"entering_orbs": {
			"type": "object",
			"description": "Separate orbs for entering/applying aspects (overrides main orbs when specified)",
			"properties": {
				"conjunction": {"type": "number"},
				"opposition": {"type": "number"},
				"trine": {"type": "number"},
				"square": {"type": "number"},
				"sextile": {"type": "number"},
				"quincunx": {"type": "number"},
				"semi_sextile": {"type": "number"},
				"semi_square": {"type": "number"},
				"sesquiquadrate": {"type": "number"}
			}
		},
		"exiting_orbs": {
			"type": "object",
			"description": "Separate orbs for exiting/separating aspects (overrides main orbs when specified)",
			"properties": {
				"conjunction": {"type": "number"},
				"opposition": {"type": "number"},
				"trine": {"type": "number"},
				"square": {"type": "number"},
				"sextile": {"type": "number"},
				"quincunx": {"type": "number"},
				"semi_sextile": {"type": "number"},
				"semi_square": {"type": "number"},
				"sesquiquadrate": {"type": "number"}
			}
		}
	}
}`)

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
	sweph.ConfigureFromEnv() // Configure ephemeris type from SWISSEPH_TYPE / SWISSEPH_JPL_FILE
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
				Name:    "solarsage-mcp",
				Version: swisseph.Version,
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
			Description: "Single chart calculation: compute planet positions, houses, and aspects at a fixed time. Supports JSON (default) and CSV output.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"},
					"orb_config": {"type": "object"},
					"format": {"type": "string", "enum": ["json", "csv"], "description": "Output format: json (default) or csv"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		{
			Name:        "calc_double_chart",
			Description: "Double chart calculation: compute inner/outer chart positions and cross-aspects. Supports JSON (default) and CSV output.",
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
					"orb_config": {"type": "object"},
					"format": {"type": "string", "enum": ["json", "csv"], "description": "Output format: json (default) or csv"}
				},
				"required": ["inner_latitude", "inner_longitude", "inner_jd_ut",
					"outer_latitude", "outer_longitude", "outer_jd_ut"]
			}`),
		},
		{
			Name:        "calc_progressions",
			Description: "Secondary progressions: compute progressed planet positions (1 day = 1 year). Supports JSON and CSV.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"natal_jd_ut": {"type": "number", "description": "Natal Julian Day (UT)"},
					"transit_jd_ut": {"type": "number", "description": "Transit Julian Day (UT)"},
					"planets": {"type": "array", "items": {"type": "string"}, "description": "Planet list"},
					"format": {"type": "string", "enum": ["json", "csv"]}
				},
				"required": ["natal_jd_ut", "transit_jd_ut"]
			}`),
		},
		{
			Name:        "calc_solar_arc",
			Description: "Solar arc directions: compute solar arc directed planet positions. Supports JSON and CSV.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"natal_jd_ut": {"type": "number", "description": "Natal Julian Day (UT)"},
					"transit_jd_ut": {"type": "number", "description": "Transit Julian Day (UT)"},
					"planets": {"type": "array", "items": {"type": "string"}, "description": "Planet list"},
					"format": {"type": "string", "enum": ["json", "csv"]}
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
					"format": {"type": "string", "enum": ["json", "csv"], "description": "Output format: json (default) or csv"},
					"timezone": {"type": "string", "default": "UTC", "description": "Timezone for CSV date/time output"}
				},
				"required": ["natal_latitude", "natal_longitude", "natal_jd_ut",
					"transit_latitude", "transit_longitude", "start_jd_ut", "end_jd_ut"]
			}`),
		},
	}

	// New feature tools
	tools = append(tools,
		toolDef{
			Name:        "calc_solar_return",
			Description: "Calculate a solar return chart: the exact moment the Sun returns to its natal longitude, with full chart data",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"natal_jd_ut": {"type": "number", "description": "Natal Julian Day (UT)"},
					"natal_latitude": {"type": "number"},
					"natal_longitude": {"type": "number"},
					"search_jd_ut": {"type": "number", "description": "Start searching from this JD (typically the birthday anniversary)"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"},
					"count": {"type": "integer", "description": "Number of consecutive returns (default 1)", "default": 1}
				},
				"required": ["natal_jd_ut", "natal_latitude", "natal_longitude", "search_jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_lunar_return",
			Description: "Calculate a lunar return chart: the exact moment the Moon returns to its natal longitude, with full chart data",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"natal_jd_ut": {"type": "number", "description": "Natal Julian Day (UT)"},
					"natal_latitude": {"type": "number"},
					"natal_longitude": {"type": "number"},
					"search_jd_ut": {"type": "number", "description": "Start searching from this JD"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"},
					"count": {"type": "integer", "description": "Number of consecutive returns (default 1)", "default": 1}
				},
				"required": ["natal_jd_ut", "natal_latitude", "natal_longitude", "search_jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_dignity",
			Description: "Calculate essential dignities (rulership, exaltation, detriment, fall), mutual receptions, and sect for a chart",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_composite_chart",
			Description: "Calculate a composite (midpoint) chart for two people, used in relationship astrology",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"person1_latitude": {"type": "number"},
					"person1_longitude": {"type": "number"},
					"person1_jd_ut": {"type": "number"},
					"person2_latitude": {"type": "number"},
					"person2_longitude": {"type": "number"},
					"person2_jd_ut": {"type": "number"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"},
					"orb_config": {"type": "object"},
					"format": {"type": "string", "enum": ["json", "csv"]}
				},
				"required": ["person1_latitude", "person1_longitude", "person1_jd_ut",
					"person2_latitude", "person2_longitude", "person2_jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_aspect_patterns",
			Description: "Detect aspect patterns (Grand Trine, T-Square, Grand Cross, Yod, Kite, Stellium) in a natal chart",
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
		toolDef{
			Name:        "calc_fixed_stars",
			Description: "Find fixed star conjunctions with chart planets from a catalog of 50+ major stars",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"},
					"orb": {"type": "number", "default": 1.5, "description": "Conjunction orb in degrees (default 1.5)"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_midpoints",
			Description: "Compute midpoint tree with 90-degree Cosmobiology dial sort and planet-on-midpoint activations",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"},
					"orb": {"type": "number", "default": 1.5, "description": "Midpoint activation orb (default 1.5)"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_harmonic_chart",
			Description: "Calculate Nth harmonic (divisional) chart. Common harmonics: 5th (quintile), 7th (septile), 9th (novile)",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"},
					"harmonic": {"type": "integer", "description": "Harmonic number (1-180)"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"},
					"orb_config": {"type": "object"}
				},
				"required": ["latitude", "longitude", "jd_ut", "harmonic"]
			}`),
		},
		toolDef{
			Name:        "calc_planetary_hours",
			Description: "Calculate the 24 Chaldean planetary hours for a given date and location, with sunrise/sunset times",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_profection",
			Description: "Calculate annual and monthly profections (time-lord technique) for a given age",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"natal_jd_ut": {"type": "number"},
					"age": {"type": "integer", "description": "Age in years for annual profection"},
					"house_system": {"type": "string", "default": "PLACIDUS"},
					"include_monthly": {"type": "boolean", "default": false},
					"start_age": {"type": "integer", "description": "Start age for timeline (optional)"},
					"end_age": {"type": "integer", "description": "End age for timeline (optional)"}
				},
				"required": ["latitude", "longitude", "natal_jd_ut", "age"]
			}`),
		},
		toolDef{
			Name:        "calc_antiscia",
			Description: "Calculate antiscia (solstice mirror points) and contra-antiscia (equinox mirror points) for all chart planets",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"},
					"orb": {"type": "number", "default": 2.0, "description": "Orb for finding antiscia pairs"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_lots",
			Description: "Calculate Arabic lots/parts (Fortune, Spirit, Eros, Necessity, Victory, Nemesis, and more) with day/night reversal",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_bounds",
			Description: "Calculate Chaldean decans and Egyptian/Ptolemaic terms (bounds) for all chart planets",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_lunar_phase",
			Description: "Get the current lunar phase, illumination, and phase angle at a given time",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"jd_ut": {"type": "number"}
				},
				"required": ["jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_lunar_phases",
			Description: "Find all new moons, full moons, and quarter moons in a date range",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"start_jd_ut": {"type": "number"},
					"end_jd_ut": {"type": "number"}
				},
				"required": ["start_jd_ut", "end_jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_eclipses",
			Description: "Find solar and lunar eclipses in a date range with type classification",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"start_jd_ut": {"type": "number"},
					"end_jd_ut": {"type": "number"}
				},
				"required": ["start_jd_ut", "end_jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_synastry",
			Description: "Calculate synastry compatibility score between two charts with category breakdown (love, passion, communication, etc.)",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"person1_latitude": {"type": "number"},
					"person1_longitude": {"type": "number"},
					"person1_jd_ut": {"type": "number"},
					"person2_latitude": {"type": "number"},
					"person2_longitude": {"type": "number"},
					"person2_jd_ut": {"type": "number"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"},
					"orb_config": {"type": "object"}
				},
				"required": ["person1_latitude", "person1_longitude", "person1_jd_ut",
					"person2_latitude", "person2_longitude", "person2_jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_dispositors",
			Description: "Calculate dispositorship chains and final dispositor for a natal chart",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"},
					"traditional": {"type": "boolean", "default": false, "description": "Use traditional rulers (Mars for Scorpio, etc.)"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_natal_report",
			Description: "Generate a comprehensive natal chart analysis combining all techniques: dignities, dispositors, patterns, lots, faces, antiscia, fixed stars, midpoints, element/modality balance, and more",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_sidereal_chart",
			Description: "Calculate a sidereal (Vedic/Jyotish) natal chart with Nakshatras, padas, and Vimshottari lords",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"},
					"ayanamsa": {"type": "string", "enum": ["LAHIRI", "RAMAN", "KRISHNAMURTI", "FAGAN_BRADLEY", "YUKTESHWAR"], "default": "LAHIRI"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_vimshottari_dasha",
			Description: "Calculate Vimshottari Maha Dasha periods from the Moon's sidereal position",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"},
					"ayanamsa": {"type": "string", "enum": ["LAHIRI", "RAMAN", "KRISHNAMURTI", "FAGAN_BRADLEY", "YUKTESHWAR"], "default": "LAHIRI"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_chart_wheel",
			Description: "Generate chart wheel rendering coordinates (x/y positions for planets, houses, aspects, signs) for SVG/Canvas visualization",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"},
					"radius": {"type": "number", "default": 0.4, "description": "Chart radius (0-0.5, default 0.4)"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_firdaria",
			Description: "Calculate Firdaria (Persian planetary period system) for traditional astrology. Returns major periods, sub-periods, and current active period.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"is_day_birth": {"type": "boolean", "description": "True if Sun was above horizon at birth"},
					"age": {"type": "number", "description": "Current age in years"},
					"start_age": {"type": "number", "description": "Start age for timeline (optional)"},
					"end_age": {"type": "number", "description": "End age for timeline (optional)"}
				},
				"required": ["is_day_birth", "age"]
			}`),
		},
		toolDef{
			Name:        "calc_davison_chart",
			Description: "Calculate Davison relationship chart (time-space midpoint method). Casts a real chart for the midpoint time and location of two birth charts.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"person1_latitude": {"type": "number"},
					"person1_longitude": {"type": "number"},
					"person1_jd_ut": {"type": "number"},
					"person2_latitude": {"type": "number"},
					"person2_longitude": {"type": "number"},
					"person2_jd_ut": {"type": "number"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"}
				},
				"required": ["person1_latitude", "person1_longitude", "person1_jd_ut", "person2_latitude", "person2_longitude", "person2_jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_primary_directions",
			Description: "Calculate Primary Directions (oldest Western predictive technique). Uses diurnal rotation to advance chart points. Supports Naibod and Ptolemy keys.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"natal_jd_ut": {"type": "number"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"aspects": {"type": "array", "items": {"type": "string"}, "description": "Aspect types to check (default: major aspects)"},
					"direction_key": {"type": "string", "enum": ["NAIBOD", "PTOLEMY", "SOLAR_ARC"], "default": "NAIBOD"},
					"max_age": {"type": "number", "default": 100},
					"house_system": {"type": "string", "default": "PLACIDUS"}
				},
				"required": ["latitude", "longitude", "natal_jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_divisional_chart",
			Description: "Calculate Vedic divisional chart (Varga). Supports D1-D60 including Navamsa (D9), Dasamsa (D10), etc.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"},
					"varga": {"type": "string", "enum": ["D1","D2","D3","D4","D7","D9","D10","D12","D16","D20","D24","D27","D30","D40","D45","D60"], "default": "D9"},
					"ayanamsa": {"type": "string", "enum": ["LAHIRI", "RAMAN", "KRISHNAMURTI", "FAGAN_BRADLEY", "YUKTESHWAR"], "default": "LAHIRI"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_ashtakavarga",
			Description: "Calculate Ashtakavarga (Vedic point-based planetary strength system). Returns bindu tables for each planet and Sarvashtakavarga.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"},
					"ayanamsa": {"type": "string", "enum": ["LAHIRI", "RAMAN", "KRISHNAMURTI", "FAGAN_BRADLEY", "YUKTESHWAR"], "default": "LAHIRI"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_yogas",
			Description: "Analyze Vedic Yogas (planetary combinations). Detects Mahapurusha, Raja, Dhana, Gajakesari, Budhaditya, and Chandra-Mangala yogas.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"},
					"ayanamsa": {"type": "string", "enum": ["LAHIRI", "RAMAN", "KRISHNAMURTI", "FAGAN_BRADLEY", "YUKTESHWAR"], "default": "LAHIRI"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_bonification",
			Description: "Analyze bonification and maltreatment (classical astrology). Evaluates benefic/malefic influences including combustion, besiegement, and planetary aspects.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"jd_ut": {"type": "number"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"}
				},
				"required": ["latitude", "longitude", "jd_ut"]
			}`),
		},
		toolDef{
			Name:        "calc_symbolic_directions",
			Description: "Calculate Symbolic Directions (1-degree-per-year, Naibod, Profection, or custom rate). Advances all natal positions by a fixed arc per year of life.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"natal_jd_ut": {"type": "number"},
					"age": {"type": "number", "description": "Age in years"},
					"method": {"type": "string", "enum": ["ONE_DEGREE", "NAIBOD", "PROFECTION", "CUSTOM"], "default": "ONE_DEGREE"},
					"custom_rate": {"type": "number", "description": "Custom rate in degrees/year (only for CUSTOM method)"},
					"planets": {"type": "array", "items": {"type": "string"}},
					"house_system": {"type": "string", "default": "PLACIDUS"}
				},
				"required": ["latitude", "longitude", "natal_jd_ut", "age"]
			}`),
		},
		toolDef{
			Name:        "calc_heliacal_events",
			Description: "Calculate heliacal risings and settings of visible planets (Mercury-Saturn). Uses Swiss Ephemeris visibility algorithms for classical astrology.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"latitude": {"type": "number"},
					"longitude": {"type": "number"},
					"altitude": {"type": "number", "default": 0, "description": "Observer altitude in meters"},
					"start_jd_ut": {"type": "number"},
					"end_jd_ut": {"type": "number"},
					"planets": {"type": "array", "items": {"type": "string"}, "description": "Planets to check (default: Mercury, Venus, Mars, Jupiter, Saturn)"}
				},
				"required": ["latitude", "longitude", "start_jd_ut", "end_jd_ut"]
			}`),
		},
	)



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
	case "calc_solar_return":
		result, err = s.handleCalcSolarReturn(params.Arguments)
	case "calc_lunar_return":
		result, err = s.handleCalcLunarReturn(params.Arguments)
	case "calc_dignity":
		result, err = s.handleCalcDignity(params.Arguments)
	case "calc_composite_chart":
		result, err = s.handleCalcCompositeChart(params.Arguments)
	case "calc_aspect_patterns":
		result, err = s.handleCalcAspectPatterns(params.Arguments)
	case "calc_fixed_stars":
		result, err = s.handleCalcFixedStars(params.Arguments)
	case "calc_midpoints":
		result, err = s.handleCalcMidpoints(params.Arguments)
	case "calc_harmonic_chart":
		result, err = s.handleCalcHarmonicChart(params.Arguments)
	case "calc_planetary_hours":
		result, err = s.handleCalcPlanetaryHours(params.Arguments)
	case "calc_profection":
		result, err = s.handleCalcProfection(params.Arguments)
	case "calc_antiscia":
		result, err = s.handleCalcAntiscia(params.Arguments)
	case "calc_lots":
		result, err = s.handleCalcLots(params.Arguments)
	case "calc_bounds":
		result, err = s.handleCalcBounds(params.Arguments)
	case "calc_lunar_phase":
		result, err = s.handleCalcLunarPhase(params.Arguments)
	case "calc_lunar_phases":
		result, err = s.handleCalcLunarPhases(params.Arguments)
	case "calc_eclipses":
		result, err = s.handleCalcEclipses(params.Arguments)
	case "calc_synastry":
		result, err = s.handleCalcSynastry(params.Arguments)
	case "calc_dispositors":
		result, err = s.handleCalcDispositors(params.Arguments)
	case "calc_natal_report":
		result, err = s.handleCalcNatalReport(params.Arguments)
	case "calc_sidereal_chart":
		result, err = s.handleCalcSiderealChart(params.Arguments)
	case "calc_vimshottari_dasha":
		result, err = s.handleCalcVimshottariDasha(params.Arguments)
	case "calc_chart_wheel":
		result, err = s.handleCalcChartWheel(params.Arguments)
	case "calc_firdaria":
		result, err = s.handleCalcFirdaria(params.Arguments)
	case "calc_davison_chart":
		result, err = s.handleCalcDavisonChart(params.Arguments)
	case "calc_primary_directions":
		result, err = s.handleCalcPrimaryDirections(params.Arguments)
	case "calc_divisional_chart":
		result, err = s.handleCalcDivisionalChart(params.Arguments)
	case "calc_ashtakavarga":
		result, err = s.handleCalcAshtakavarga(params.Arguments)
	case "calc_yogas":
		result, err = s.handleCalcYogas(params.Arguments)
	case "calc_bonification":
		result, err = s.handleCalcBonification(params.Arguments)
	case "calc_symbolic_directions":
		result, err = s.handleCalcSymbolicDirections(params.Arguments)
	case "calc_heliacal_events":
		result, err = s.handleCalcHeliacalEvents(params.Arguments)
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

// orbOrDefault returns the custom orb config if provided, otherwise the fallback
func orbOrDefault(custom *models.OrbConfig, fallback models.OrbConfig) models.OrbConfig {
	if custom != nil {
		return *custom
	}
	return fallback
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
		Latitude    float64            `json:"latitude"`
		Longitude   float64            `json:"longitude"`
		JDUT        float64            `json:"jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		OrbConfig   *models.OrbConfig  `json:"orb_config"`
		Format      string             `json:"format"`
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

	chartInfo, err := chart.CalcSingleChart(input.Latitude, input.Longitude, input.JDUT, input.Planets,
		orbOrDefault(input.OrbConfig, models.DefaultOrbConfig()), input.HouseSystem)
	if err != nil {
		return nil, err
	}

	if input.Format == "csv" {
		return map[string]interface{}{
			"format":   "csv",
			"planets":  export.ChartToCSV(chartInfo),
			"aspects":  export.AspectsToCSV(chartInfo.Aspects),
			"houses":   export.HousesToCSV(chartInfo.Houses, chartInfo.Angles),
		}, nil
	}
	return chartInfo, nil
}

func (s *Server) handleCalcDoubleChart(args json.RawMessage) (interface{}, error) {
	var input struct {
		InnerLatitude  float64                    `json:"inner_latitude"`
		InnerLongitude float64                    `json:"inner_longitude"`
		InnerJDUT      float64                    `json:"inner_jd_ut"`
		InnerPlanets   []models.PlanetID          `json:"inner_planets"`
		OuterLatitude  float64                    `json:"outer_latitude"`
		OuterLongitude float64                    `json:"outer_longitude"`
		OuterJDUT      float64                    `json:"outer_jd_ut"`
		OuterPlanets   []models.PlanetID          `json:"outer_planets"`
		SpecialPoints  *models.SpecialPointsConfig `json:"special_points"`
		HouseSystem    models.HouseSystem          `json:"house_system"`
		OrbConfig      *models.OrbConfig           `json:"orb_config"`
		Format         string                      `json:"format"`
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
	orbs := orbOrDefault(input.OrbConfig, models.DefaultOrbConfig())
	innerChart, outerChart, crossAspects, err := chart.CalcDoubleChart(
		input.InnerLatitude, input.InnerLongitude, input.InnerJDUT, input.InnerPlanets,
		input.OuterLatitude, input.OuterLongitude, input.OuterJDUT, input.OuterPlanets,
		input.SpecialPoints, orbs, input.HouseSystem,
	)
	if err != nil {
		return nil, err
	}

	if input.Format == "csv" {
		return map[string]interface{}{
			"format":         "csv",
			"inner_planets":  export.ChartToCSV(innerChart),
			"outer_planets":  export.ChartToCSV(outerChart),
			"cross_aspects":  export.CrossAspectsToCSV(crossAspects),
		}, nil
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

// directedToPositions converts directed planets to PlanetPosition for CSV export
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

// directedPlanet holds a planet position from progressions or solar arc
type directedPlanet struct {
	PlanetID     models.PlanetID `json:"planet_id"`
	Longitude    float64         `json:"longitude"`
	Speed        float64         `json:"speed"`
	IsRetrograde bool            `json:"is_retrograde,omitempty"`
	Sign         string          `json:"sign"`
	SignDegree   float64         `json:"sign_degree"`
}

// calcDirectedPlanets computes positions for a list of planets using the given calc function
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

func (s *Server) handleCalcProgressions(args json.RawMessage) (interface{}, error) {
	var input struct {
		NatalJDUT   float64           `json:"natal_jd_ut"`
		TransitJDUT float64           `json:"transit_jd_ut"`
		Planets     []models.PlanetID `json:"planets"`
		Format      string            `json:"format"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	if len(input.Planets) == 0 {
		input.Planets = defaultPlanets
	}

	dp := calcDirectedPlanets(input.Planets, input.NatalJDUT, input.TransitJDUT, progressions.CalcProgressedLongitude)

	if input.Format == "csv" {
		positions := directedToPositions(dp)
		return map[string]interface{}{
			"format":  "csv",
			"age":     progressions.Age(input.NatalJDUT, input.TransitJDUT),
			"planets": export.PositionsToCSV(positions),
		}, nil
	}

	return map[string]interface{}{
		"age":           progressions.Age(input.NatalJDUT, input.TransitJDUT),
		"progressed_jd": progressions.SecondaryProgressionJD(input.NatalJDUT, input.TransitJDUT),
		"planets":       dp,
	}, nil
}

func (s *Server) handleCalcSolarArc(args json.RawMessage) (interface{}, error) {
	var input struct {
		NatalJDUT   float64           `json:"natal_jd_ut"`
		TransitJDUT float64           `json:"transit_jd_ut"`
		Planets     []models.PlanetID `json:"planets"`
		Format      string            `json:"format"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	if len(input.Planets) == 0 {
		input.Planets = defaultPlanets
	}

	offset, err := progressions.SolarArcOffset(input.NatalJDUT, input.TransitJDUT)
	if err != nil {
		return nil, err
	}

	dp := calcDirectedPlanets(input.Planets, input.NatalJDUT, input.TransitJDUT, progressions.CalcSolarArcLongitude)

	if input.Format == "csv" {
		return map[string]interface{}{
			"format":           "csv",
			"age":              progressions.Age(input.NatalJDUT, input.TransitJDUT),
			"solar_arc_offset": offset,
			"planets":          export.PositionsToCSV(directedToPositions(dp)),
		}, nil
	}

	return map[string]interface{}{
		"age":              progressions.Age(input.NatalJDUT, input.TransitJDUT),
		"solar_arc_offset": offset,
		"planets":          dp,
	}, nil
}

func (s *Server) handleCalcSolarReturn(args json.RawMessage) (interface{}, error) {
	var input struct {
		NatalJDUT    float64            `json:"natal_jd_ut"`
		NatalLat     float64            `json:"natal_latitude"`
		NatalLon     float64            `json:"natal_longitude"`
		SearchJDUT   float64            `json:"search_jd_ut"`
		Planets      []models.PlanetID  `json:"planets"`
		HouseSystem  models.HouseSystem `json:"house_system"`
		Count        int                `json:"count"`
		Format       string             `json:"format"`
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
	if input.Count <= 0 {
		input.Count = 1
	}

	ri := returns.ReturnInput{
		NatalJD: input.NatalJDUT, NatalLat: input.NatalLat, NatalLon: input.NatalLon,
		SearchJD: input.SearchJDUT, Planets: input.Planets,
		OrbConfig: models.DefaultOrbConfig(), HouseSystem: input.HouseSystem,
	}

	if input.Count == 1 {
		rc, err := returns.CalcSolarReturn(ri)
		if err != nil {
			return nil, err
		}
		if input.Format == "csv" && rc.Chart != nil {
			return map[string]interface{}{
				"format":    "csv",
				"return_jd": rc.ReturnJD,
				"age":       rc.Age,
				"planets":   export.ChartToCSV(rc.Chart),
				"aspects":   export.AspectsToCSV(rc.Chart.Aspects),
				"houses":    export.HousesToCSV(rc.Chart.Houses, rc.Chart.Angles),
			}, nil
		}
		return rc, nil
	}
	return returns.CalcSolarReturnSeries(ri, input.Count)
}

func (s *Server) handleCalcLunarReturn(args json.RawMessage) (interface{}, error) {
	var input struct {
		NatalJDUT    float64            `json:"natal_jd_ut"`
		NatalLat     float64            `json:"natal_latitude"`
		NatalLon     float64            `json:"natal_longitude"`
		SearchJDUT   float64            `json:"search_jd_ut"`
		Planets      []models.PlanetID  `json:"planets"`
		HouseSystem  models.HouseSystem `json:"house_system"`
		Count        int                `json:"count"`
		Format       string             `json:"format"`
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
	if input.Count <= 0 {
		input.Count = 1
	}

	ri := returns.ReturnInput{
		NatalJD: input.NatalJDUT, NatalLat: input.NatalLat, NatalLon: input.NatalLon,
		SearchJD: input.SearchJDUT, Planets: input.Planets,
		OrbConfig: models.DefaultOrbConfig(), HouseSystem: input.HouseSystem,
	}

	if input.Count == 1 {
		rc, err := returns.CalcLunarReturn(ri)
		if err != nil {
			return nil, err
		}
		if input.Format == "csv" && rc.Chart != nil {
			return map[string]interface{}{
				"format":    "csv",
				"return_jd": rc.ReturnJD,
				"age":       rc.Age,
				"planets":   export.ChartToCSV(rc.Chart),
				"aspects":   export.AspectsToCSV(rc.Chart.Aspects),
				"houses":    export.HousesToCSV(rc.Chart.Houses, rc.Chart.Angles),
			}, nil
		}
		return rc, nil
	}
	return returns.CalcLunarReturnSeries(ri, input.Count)
}

func (s *Server) handleCalcDignity(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude    float64            `json:"latitude"`
		Longitude   float64            `json:"longitude"`
		JDUT        float64            `json:"jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		Format      string             `json:"format"`
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

	chartInfo, err := chart.CalcSingleChart(input.Latitude, input.Longitude, input.JDUT,
		input.Planets, models.DefaultOrbConfig(), input.HouseSystem)
	if err != nil {
		return nil, err
	}

	dignities := dignity.CalcChartDignities(chartInfo.Planets)
	receptions := dignity.FindMutualReceptions(chartInfo.Planets)

	// Determine day/night
	isDayChart := chart.IsDayChart(input.JDUT, chartInfo.Angles.ASC)
	var sects []dignity.SectInfo
	for _, p := range chartInfo.Planets {
		sects = append(sects, dignity.CalcSect(p.PlanetID, isDayChart))
	}

	if input.Format == "csv" {
		return map[string]interface{}{
			"format":    "csv",
			"dignities": export.DignityToCSV(dignities),
		}, nil
	}

	return map[string]interface{}{
		"dignities":         dignities,
		"mutual_receptions": receptions,
		"sect":              sects,
		"is_day_chart":      isDayChart,
	}, nil
}

func (s *Server) handleCalcCompositeChart(args json.RawMessage) (interface{}, error) {
	var input struct {
		Person1Lat float64            `json:"person1_latitude"`
		Person1Lon float64            `json:"person1_longitude"`
		Person1JD  float64            `json:"person1_jd_ut"`
		Person2Lat float64            `json:"person2_latitude"`
		Person2Lon float64            `json:"person2_longitude"`
		Person2JD  float64            `json:"person2_jd_ut"`
		Planets    []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		OrbConfig  *models.OrbConfig  `json:"orb_config"`
		Format     string             `json:"format"`
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

	cc, err := composite.CalcCompositeChart(composite.CompositeInput{
		Person1Lat: input.Person1Lat, Person1Lon: input.Person1Lon, Person1JD: input.Person1JD,
		Person2Lat: input.Person2Lat, Person2Lon: input.Person2Lon, Person2JD: input.Person2JD,
		Planets: input.Planets, OrbConfig: orbOrDefault(input.OrbConfig, models.DefaultOrbConfig()),
		HouseSystem: input.HouseSystem,
	})
	if err != nil {
		return nil, err
	}

	if input.Format == "csv" {
		return map[string]interface{}{
			"format":  "csv",
			"planets": export.ChartToCSV(&models.ChartInfo{Planets: cc.Planets}),
			"aspects": export.AspectsToCSV(cc.Aspects),
			"houses":  export.HousesToCSV(cc.Houses, cc.Angles),
		}, nil
	}

	return cc, nil
}

func (s *Server) handleCalcAspectPatterns(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude    float64            `json:"latitude"`
		Longitude   float64            `json:"longitude"`
		JDUT        float64            `json:"jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
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
	orbs := orbOrDefault(input.OrbConfig, models.DefaultOrbConfig())

	chartInfo, err := chart.CalcSingleChart(input.Latitude, input.Longitude, input.JDUT,
		input.Planets, orbs, input.HouseSystem)
	if err != nil {
		return nil, err
	}

	// Build body list for pattern detection
	var bodies []aspect.Body
	for _, p := range chartInfo.Planets {
		bodies = append(bodies, aspect.Body{
			ID: string(p.PlanetID), Longitude: p.Longitude, Speed: p.Speed,
		})
	}

	patterns := aspect.FindPatterns(chartInfo.Aspects, bodies, orbs)

	return map[string]interface{}{
		"patterns": patterns,
		"aspects":  chartInfo.Aspects,
	}, nil
}

func (s *Server) handleCalcFixedStars(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude    float64            `json:"latitude"`
		Longitude   float64            `json:"longitude"`
		JDUT        float64            `json:"jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		Orb         float64            `json:"orb"`
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
	if input.Orb <= 0 {
		input.Orb = 1.5
	}

	chartInfo, err := chart.CalcSingleChart(input.Latitude, input.Longitude, input.JDUT,
		input.Planets, models.DefaultOrbConfig(), input.HouseSystem)
	if err != nil {
		return nil, err
	}

	conjunctions := fixedstars.FindConjunctions(chartInfo.Planets, input.Orb, input.JDUT)

	return map[string]interface{}{
		"conjunctions":  conjunctions,
		"catalog_count": len(fixedstars.Catalog),
	}, nil
}

func (s *Server) handleCalcMidpoints(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude    float64            `json:"latitude"`
		Longitude   float64            `json:"longitude"`
		JDUT        float64            `json:"jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		Orb         float64            `json:"orb"`
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

	chartInfo, err := chart.CalcSingleChart(input.Latitude, input.Longitude, input.JDUT,
		input.Planets, models.DefaultOrbConfig(), input.HouseSystem)
	if err != nil {
		return nil, err
	}

	tree := midpoint.CalcMidpoints(chartInfo.Planets, input.Orb)
	return tree, nil
}

func (s *Server) handleCalcHarmonicChart(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude    float64            `json:"latitude"`
		Longitude   float64            `json:"longitude"`
		JDUT        float64            `json:"jd_ut"`
		Harmonic    int                `json:"harmonic"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		OrbConfig   *models.OrbConfig  `json:"orb_config"`
		Format      string             `json:"format"`
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

	hc, err := harmonic.CalcHarmonicChart(input.Latitude, input.Longitude, input.JDUT,
		input.Harmonic, input.Planets, orbOrDefault(input.OrbConfig, models.DefaultOrbConfig()), input.HouseSystem)
	if err != nil {
		return nil, err
	}

	if input.Format == "csv" {
		return map[string]interface{}{
			"format":   "csv",
			"harmonic": hc.Harmonic,
			"planets":  export.PositionsToCSV(hc.Planets),
			"aspects":  export.AspectsToCSV(hc.Aspects),
		}, nil
	}
	return hc, nil
}

func (s *Server) handleCalcPlanetaryHours(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		JDUT      float64 `json:"jd_ut"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	return planetary.CalcPlanetaryHours(input.JDUT, input.Latitude, input.Longitude)
}

func (s *Server) handleCalcProfection(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude       float64            `json:"latitude"`
		Longitude      float64            `json:"longitude"`
		NatalJDUT      float64            `json:"natal_jd_ut"`
		Age            int                `json:"age"`
		HouseSystem    models.HouseSystem `json:"house_system"`
		IncludeMonthly bool               `json:"include_monthly"`
		StartAge       *int               `json:"start_age"`
		EndAge         *int               `json:"end_age"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	if input.HouseSystem == "" {
		input.HouseSystem = models.HousePlacidus
	}

	chartInfo, err := chart.CalcSingleChart(input.Latitude, input.Longitude, input.NatalJDUT,
		defaultPlanets, models.DefaultOrbConfig(), input.HouseSystem)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	annual := profection.CalcAnnualProfection(chartInfo.Angles.ASC, chartInfo.Houses, input.Age)
	result["annual"] = annual

	if input.IncludeMonthly {
		result["monthly"] = profection.CalcMonthlyProfections(chartInfo.Angles.ASC, input.Age)
	}

	if input.StartAge != nil && input.EndAge != nil {
		result["timeline"] = profection.ProfectionTimeline(chartInfo.Angles.ASC, chartInfo.Houses, *input.StartAge, *input.EndAge)
	}

	return result, nil
}

func (s *Server) handleCalcAntiscia(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude    float64            `json:"latitude"`
		Longitude   float64            `json:"longitude"`
		JDUT        float64            `json:"jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		Orb         float64            `json:"orb"`
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

	chartInfo, err := chart.CalcSingleChart(input.Latitude, input.Longitude, input.JDUT,
		input.Planets, models.DefaultOrbConfig(), input.HouseSystem)
	if err != nil {
		return nil, err
	}

	points := antiscia.CalcChartAntiscia(chartInfo.Planets)
	pairs := antiscia.FindAntisciaPairs(chartInfo.Planets, input.Orb)

	return map[string]interface{}{
		"antiscia_points": points,
		"antiscia_pairs":  pairs,
	}, nil
}

func (s *Server) handleCalcLots(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude    float64            `json:"latitude"`
		Longitude   float64            `json:"longitude"`
		JDUT        float64            `json:"jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		Format      string             `json:"format"`
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

	chartInfo, err := chart.CalcSingleChart(input.Latitude, input.Longitude, input.JDUT,
		input.Planets, models.DefaultOrbConfig(), input.HouseSystem)
	if err != nil {
		return nil, err
	}

	isDayChart := chart.IsDayChart(input.JDUT, chartInfo.Angles.ASC)
	lotResults := lots.CalcStandardLots(chartInfo.Planets, chartInfo.Angles.ASC, isDayChart)

	if input.Format == "csv" {
		return map[string]interface{}{
			"format": "csv",
			"lots":   export.LotsToCSV(lotResults),
		}, nil
	}

	return map[string]interface{}{
		"lots":         lotResults,
		"is_day_chart": isDayChart,
	}, nil
}

func (s *Server) handleCalcBounds(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude    float64            `json:"latitude"`
		Longitude   float64            `json:"longitude"`
		JDUT        float64            `json:"jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
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

	chartInfo, err := chart.CalcSingleChart(input.Latitude, input.Longitude, input.JDUT,
		input.Planets, models.DefaultOrbConfig(), input.HouseSystem)
	if err != nil {
		return nil, err
	}

	faces := bounds.CalcChartFaces(chartInfo.Planets)
	return map[string]interface{}{
		"faces": faces,
	}, nil
}

func (s *Server) handleCalcLunarPhase(args json.RawMessage) (interface{}, error) {
	var input struct {
		JDUT float64 `json:"jd_ut"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	return lunar.CalcLunarPhase(input.JDUT)
}

func (s *Server) handleCalcLunarPhases(args json.RawMessage) (interface{}, error) {
	var input struct {
		StartJDUT float64 `json:"start_jd_ut"`
		EndJDUT   float64 `json:"end_jd_ut"`
		Format    string  `json:"format"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	phases, err := lunar.FindLunarPhases(input.StartJDUT, input.EndJDUT)
	if err != nil {
		return nil, err
	}
	if input.Format == "csv" {
		return map[string]interface{}{
			"format": "csv",
			"phases": export.LunarPhasesToCSV(phases),
			"count":  len(phases),
		}, nil
	}
	return map[string]interface{}{
		"phases": phases,
		"count":  len(phases),
	}, nil
}

func (s *Server) handleCalcEclipses(args json.RawMessage) (interface{}, error) {
	var input struct {
		StartJDUT float64 `json:"start_jd_ut"`
		EndJDUT   float64 `json:"end_jd_ut"`
		Format    string  `json:"format"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	eclipses, err := lunar.FindEclipses(input.StartJDUT, input.EndJDUT)
	if err != nil {
		return nil, err
	}
	if input.Format == "csv" {
		return map[string]interface{}{
			"format":   "csv",
			"eclipses": export.EclipsesToCSV(eclipses),
			"count":    len(eclipses),
		}, nil
	}
	return map[string]interface{}{
		"eclipses": eclipses,
		"count":    len(eclipses),
	}, nil
}

func (s *Server) handleCalcSynastry(args json.RawMessage) (interface{}, error) {
	var input struct {
		Person1Lat float64            `json:"person1_latitude"`
		Person1Lon float64            `json:"person1_longitude"`
		Person1JD  float64            `json:"person1_jd_ut"`
		Person2Lat float64            `json:"person2_latitude"`
		Person2Lon float64            `json:"person2_longitude"`
		Person2JD  float64            `json:"person2_jd_ut"`
		Planets    []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		OrbConfig  *models.OrbConfig  `json:"orb_config"`
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
	orbs := orbOrDefault(input.OrbConfig, models.DefaultOrbConfig())

	chart1, err := chart.CalcSingleChart(input.Person1Lat, input.Person1Lon, input.Person1JD,
		input.Planets, orbs, input.HouseSystem)
	if err != nil {
		return nil, fmt.Errorf("person 1 chart: %w", err)
	}
	chart2, err := chart.CalcSingleChart(input.Person2Lat, input.Person2Lon, input.Person2JD,
		input.Planets, orbs, input.HouseSystem)
	if err != nil {
		return nil, fmt.Errorf("person 2 chart: %w", err)
	}

	score := synastry.CalcSynastryFromCharts(chart1.Planets, chart2.Planets, orbs)
	return score, nil
}

func (s *Server) handleCalcDispositors(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude    float64            `json:"latitude"`
		Longitude   float64            `json:"longitude"`
		JDUT        float64            `json:"jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		Traditional bool               `json:"traditional"`
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

	chartInfo, err := chart.CalcSingleChart(input.Latitude, input.Longitude, input.JDUT,
		input.Planets, models.DefaultOrbConfig(), input.HouseSystem)
	if err != nil {
		return nil, err
	}

	return dispositor.CalcDispositors(chartInfo.Planets, input.Traditional), nil
}

func (s *Server) handleCalcSiderealChart(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude  float64        `json:"latitude"`
		Longitude float64        `json:"longitude"`
		JDUT      float64        `json:"jd_ut"`
		Ayanamsa  vedic.Ayanamsa `json:"ayanamsa"`
		Format    string         `json:"format"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	if input.Ayanamsa == "" {
		input.Ayanamsa = vedic.AyanamsaLahiri
	}

	sc, err := vedic.CalcSiderealChart(input.Latitude, input.Longitude, input.JDUT, input.Ayanamsa)
	if err != nil {
		return nil, err
	}

	if input.Format == "csv" {
		var sb strings.Builder
		sb.WriteString("Planet,TropicalLon,SiderealLon,SiderealSign,SiderealDeg,Nakshatra,Pada,NakshatraLord,Glyph\n")
		for _, p := range sc.Planets {
			sb.WriteString(fmt.Sprintf("%s,%.4f,%.4f,%s,%.4f,%s,%d,%s,%s\n",
				models.BodyDisplayName(string(p.PlanetID)),
				p.Longitude, p.SiderealLon,
				p.SiderealSign, p.SiderealDeg,
				p.Nakshatra, p.NakshatraPada,
				models.BodyDisplayName(string(p.NakshatraLord)),
				models.PlanetGlyph(p.PlanetID),
			))
		}
		return map[string]interface{}{
			"format":   "csv",
			"ayanamsa": string(sc.Ayanamsa),
			"planets":  sb.String(),
		}, nil
	}
	return sc, nil
}

func (s *Server) handleCalcVimshottariDasha(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude  float64        `json:"latitude"`
		Longitude float64        `json:"longitude"`
		JDUT      float64        `json:"jd_ut"`
		Ayanamsa  vedic.Ayanamsa `json:"ayanamsa"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	if input.Ayanamsa == "" {
		input.Ayanamsa = vedic.AyanamsaLahiri
	}

	sc, err := vedic.CalcSiderealChart(input.Latitude, input.Longitude, input.JDUT, input.Ayanamsa)
	if err != nil {
		return nil, err
	}

	// Find Moon's sidereal longitude
	var moonSidLon float64
	for _, p := range sc.Planets {
		if p.PlanetID == models.PlanetMoon {
			moonSidLon = p.SiderealLon
			break
		}
	}

	periods := vedic.CalcVimshottariDasha(moonSidLon)
	return map[string]interface{}{
		"moon_nakshatra": sc.Planets[1].Nakshatra, // Moon is always index 1
		"moon_sidereal":  moonSidLon,
		"dasha_periods":  periods,
	}, nil
}

func (s *Server) handleCalcChartWheel(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude    float64            `json:"latitude"`
		Longitude   float64            `json:"longitude"`
		JDUT        float64            `json:"jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
		Radius      float64            `json:"radius"`
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

	chartInfo, err := chart.CalcSingleChart(input.Latitude, input.Longitude, input.JDUT,
		input.Planets, models.DefaultOrbConfig(), input.HouseSystem)
	if err != nil {
		return nil, err
	}

	wheel := render.CalcChartWheel(chartInfo, input.Radius)
	signs := render.CalcSignSegments(chartInfo.Angles.ASC, input.Radius)

	return map[string]interface{}{
		"wheel": wheel,
		"signs": signs,
	}, nil
}

func (s *Server) handleCalcNatalReport(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		JDUT      float64 `json:"jd_ut"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	return report.GenerateNatalReport(input.Latitude, input.Longitude, input.JDUT)
}

func (s *Server) handleCalcFirdaria(args json.RawMessage) (interface{}, error) {
	var input struct {
		IsDayBirth bool     `json:"is_day_birth"`
		Age        float64  `json:"age"`
		StartAge   *float64 `json:"start_age"`
		EndAge     *float64 `json:"end_age"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	result := firdaria.CalcFirdaria(input.IsDayBirth, input.Age)

	if input.StartAge != nil && input.EndAge != nil {
		result.Periods = firdaria.CalcFirdariaTimeline(input.IsDayBirth, *input.StartAge, *input.EndAge)
	}

	return result, nil
}

func (s *Server) handleCalcDavisonChart(args json.RawMessage) (interface{}, error) {
	var input struct {
		Person1Lat  float64            `json:"person1_latitude"`
		Person1Lon  float64            `json:"person1_longitude"`
		Person1JD   float64            `json:"person1_jd_ut"`
		Person2Lat  float64            `json:"person2_latitude"`
		Person2Lon  float64            `json:"person2_longitude"`
		Person2JD   float64            `json:"person2_jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	if input.HouseSystem == "" {
		input.HouseSystem = models.HousePlacidus
	}

	return composite.CalcDavisonChart(composite.CompositeInput{
		Person1Lat:  input.Person1Lat,
		Person1Lon:  input.Person1Lon,
		Person1JD:   input.Person1JD,
		Person2Lat:  input.Person2Lat,
		Person2Lon:  input.Person2Lon,
		Person2JD:   input.Person2JD,
		Planets:     input.Planets,
		OrbConfig:   models.DefaultOrbConfig(),
		HouseSystem: input.HouseSystem,
	})
}

func (s *Server) handleCalcPrimaryDirections(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude     float64            `json:"latitude"`
		Longitude    float64            `json:"longitude"`
		NatalJDUT    float64            `json:"natal_jd_ut"`
		Planets      []models.PlanetID  `json:"planets"`
		Aspects      []models.AspectType `json:"aspects"`
		DirectionKey string             `json:"direction_key"`
		MaxAge       float64            `json:"max_age"`
		HouseSystem  models.HouseSystem `json:"house_system"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	if input.HouseSystem == "" {
		input.HouseSystem = models.HousePlacidus
	}
	if input.MaxAge == 0 {
		input.MaxAge = 100
	}
	key := primary.KeyNaibod
	if input.DirectionKey != "" {
		key = primary.DirectionKey(input.DirectionKey)
	}
	if len(input.Aspects) == 0 {
		input.Aspects = []models.AspectType{
			models.AspectConjunction, models.AspectOpposition,
			models.AspectTrine, models.AspectSquare, models.AspectSextile,
		}
	}

	return primary.CalcPrimaryDirections(primary.PrimaryDirectionInput{
		NatalJD:     input.NatalJDUT,
		GeoLat:      input.Latitude,
		GeoLon:      input.Longitude,
		Planets:     input.Planets,
		Aspects:     input.Aspects,
		Key:         key,
		MaxAge:      input.MaxAge,
		HouseSystem: input.HouseSystem,
	})
}

func (s *Server) handleCalcDivisionalChart(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		JDUT      float64 `json:"jd_ut"`
		Varga     string  `json:"varga"`
		Ayanamsa  string  `json:"ayanamsa"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	if input.Varga == "" {
		input.Varga = "D9"
	}
	if input.Ayanamsa == "" {
		input.Ayanamsa = "LAHIRI"
	}

	return divisional.CalcDivisionalChart(input.Latitude, input.Longitude, input.JDUT,
		divisional.VargaType(input.Varga), vedic.Ayanamsa(input.Ayanamsa))
}

func (s *Server) handleCalcAshtakavarga(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		JDUT      float64 `json:"jd_ut"`
		Ayanamsa  string  `json:"ayanamsa"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	if input.Ayanamsa == "" {
		input.Ayanamsa = "LAHIRI"
	}

	siderealChart, err := vedic.CalcSiderealChart(input.Latitude, input.Longitude, input.JDUT,
		vedic.Ayanamsa(input.Ayanamsa))
	if err != nil {
		return nil, err
	}

	return ashtakavarga.CalcAshtakavarga(siderealChart.Planets, siderealChart.SiderealAngles.ASC), nil
}

func (s *Server) handleCalcYogas(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		JDUT      float64 `json:"jd_ut"`
		Ayanamsa  string  `json:"ayanamsa"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	if input.Ayanamsa == "" {
		input.Ayanamsa = "LAHIRI"
	}

	siderealChart, err := vedic.CalcSiderealChart(input.Latitude, input.Longitude, input.JDUT,
		vedic.Ayanamsa(input.Ayanamsa))
	if err != nil {
		return nil, err
	}

	return yoga.AnalyzeYogas(siderealChart.Planets, siderealChart.Houses, siderealChart.SiderealAngles.ASC), nil
}

func (s *Server) handleCalcBonification(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude    float64            `json:"latitude"`
		Longitude   float64            `json:"longitude"`
		JDUT        float64            `json:"jd_ut"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
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

	chartInfo, err := chart.CalcSingleChart(input.Latitude, input.Longitude, input.JDUT,
		input.Planets, models.DefaultOrbConfig(), input.HouseSystem)
	if err != nil {
		return nil, err
	}

	return dignity.CalcChartBonMal(chartInfo.Planets), nil
}

func (s *Server) handleCalcSymbolicDirections(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude    float64            `json:"latitude"`
		Longitude   float64            `json:"longitude"`
		NatalJDUT   float64            `json:"natal_jd_ut"`
		Age         float64            `json:"age"`
		Method      string             `json:"method"`
		CustomRate  float64            `json:"custom_rate"`
		Planets     []models.PlanetID  `json:"planets"`
		HouseSystem models.HouseSystem `json:"house_system"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}
	if input.HouseSystem == "" {
		input.HouseSystem = models.HousePlacidus
	}
	method := symbolic.MethodOneDegree
	if input.Method != "" {
		method = symbolic.DirectionMethod(input.Method)
	}

	return symbolic.CalcSymbolicDirections(symbolic.SymbolicInput{
		NatalJD:     input.NatalJDUT,
		GeoLat:      input.Latitude,
		GeoLon:      input.Longitude,
		Age:         input.Age,
		Method:      method,
		CustomRate:  input.CustomRate,
		Planets:     input.Planets,
		OrbConfig:   models.DefaultOrbConfig(),
		HouseSystem: input.HouseSystem,
	})
}

func (s *Server) handleCalcHeliacalEvents(args json.RawMessage) (interface{}, error) {
	var input struct {
		Latitude  float64           `json:"latitude"`
		Longitude float64           `json:"longitude"`
		Altitude  float64           `json:"altitude"`
		StartJDUT float64           `json:"start_jd_ut"`
		EndJDUT   float64           `json:"end_jd_ut"`
		Planets   []models.PlanetID `json:"planets"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, err
	}

	return heliacal.CalcHeliacalEvents(input.Latitude, input.Longitude, input.Altitude,
		input.StartJDUT, input.EndJDUT, input.Planets)
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
	baseOrbs := orbOrDefault(input.OrbConfig, models.DefaultOrbConfig())
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
		OrbConfigTransit:      orbOrDefault(input.OrbConfigTransit, baseOrbs),
		OrbConfigProgressions: orbOrDefault(input.OrbConfigProgressions, baseOrbs),
		OrbConfigSolarArc:     orbOrDefault(input.OrbConfigSolarArc, baseOrbs),
		HouseSystem:           input.HouseSystem,
	}, input.Timezone, nil
}

