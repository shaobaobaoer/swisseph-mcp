package solarsage

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/shaobaobaoer/solarsage-mcp/internal/aspect"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
)

// SFAspectRecord represents a single aspect record from Solar Fire CSV export
type SFAspectRecord struct {
	P1        string  // Planet/point name
	P1House   int
	Aspect    string  // Aspect name (Conjunction, Sextile, etc.)
	P2        string
	P2House   int
	EventType string  // Begin, Exact, Leave, Void, SignIngress, Station
	Type      string  // Tr-Na, Sp-Na, Sa-Na, Sr-Na, Sp-Sp, Tr-Sp, Tr-Sa, Sa-Sp, Sa-Sa
	Date      string  // YYYY-MM-DD
	Time      string  // HH:MM:SS
	Timezone  string  // AWST, etc.
	Age       float64
	Pos1Deg   float64 // Degrees
	Pos1Sign  string  // Sign name
	Pos1Dir   string  // Dir or Rx
	Pos2Deg   float64
	Pos2Sign  string
	Pos2Dir   string
}

// ParseSFCSV reads Solar Fire CSV and filters by event type and date range
func ParseSFCSV(csvPath string, eventType, chartType, snapshotDate string) ([]SFAspectRecord, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("open CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read CSV: %w", err)
	}

	if len(records) < 1 {
		return nil, fmt.Errorf("empty CSV")
	}

	var results []SFAspectRecord

	// Parse header
	headerMap := make(map[string]int)
	for i, col := range records[0] {
		// Remove BOM and trim whitespace from column names
		col = strings.TrimSpace(strings.TrimPrefix(col, "\ufeff"))
		headerMap[col] = i
	}

	// Debug: Log header columns
	if len(headerMap) < 5 {
		headerList := make([]string, 0, len(records[0]))
		for i, col := range records[0] {
			headerList = append(headerList, fmt.Sprintf("%d:%s", i, col))
		}
		return nil, fmt.Errorf("header parse failed (got %d cols, first row: %v)", len(headerMap), headerList)
	}

	// Validate required columns
	requiredCols := []string{"P1", "Aspect", "P2", "EventType", "Type", "Date", "Time", "Pos1_Deg", "Pos2_Deg"}
	for _, col := range requiredCols {
		if _, exists := headerMap[col]; !exists {
			// Debug: list available columns
			availableCols := make([]string, 0, len(headerMap))
			for col := range headerMap {
				availableCols = append(availableCols, col)
			}
			return nil, fmt.Errorf("missing column: %s (available: %v)", col, availableCols)
		}
	}

	// Parse data rows
	for _, record := range records[1:] {
		if len(record) <= headerMap["Pos2_Deg"] {
			continue // Skip malformed rows
		}

		et := strings.TrimSpace(record[headerMap["EventType"]])
		ct := strings.TrimSpace(record[headerMap["Type"]])
		d := strings.TrimSpace(record[headerMap["Date"]])

		// Filter by eventType, chartType, date
		if eventType != "" && et != eventType {
			continue
		}
		if chartType != "" && ct != chartType {
			continue
		}
		if snapshotDate != "" && d != snapshotDate {
			continue
		}

		// Parse numeric fields
		pos1, _ := strconv.ParseFloat(strings.TrimSpace(record[headerMap["Pos1_Deg"]]), 64)
		pos2, _ := strconv.ParseFloat(strings.TrimSpace(record[headerMap["Pos2_Deg"]]), 64)
		age, _ := strconv.ParseFloat(strings.TrimSpace(record[headerMap["Age"]]), 64)

		p1House := 0
		if h, exists := headerMap["P1_House"]; exists && len(record) > h {
			p1House, _ = strconv.Atoi(strings.TrimSpace(record[h]))
		}
		p2House := 0
		if h, exists := headerMap["P2_House"]; exists && len(record) > h {
			p2House, _ = strconv.Atoi(strings.TrimSpace(record[h]))
		}

		results = append(results, SFAspectRecord{
			P1:        strings.TrimSpace(record[headerMap["P1"]]),
			P1House:   p1House,
			Aspect:    strings.TrimSpace(record[headerMap["Aspect"]]),
			P2:        strings.TrimSpace(record[headerMap["P2"]]),
			P2House:   p2House,
			EventType: et,
			Type:      ct,
			Date:      d,
			Time:      strings.TrimSpace(record[headerMap["Time"]]),
			Timezone:  record[headerMap["Timezone"]],
			Age:       age,
			Pos1Deg:   pos1,
			Pos1Sign:  record[headerMap["Pos1_Sign"]],
			Pos1Dir:   record[headerMap["Pos1_Dir"]],
			Pos2Deg:   pos2,
			Pos2Sign:  record[headerMap["Pos2_Sign"]],
			Pos2Dir:   record[headerMap["Pos2_Dir"]],
		})
	}

	return results, nil
}

// MapSFBodyName converts Solar Fire body name to models.PlanetID
func MapSFBodyName(sfName string) models.PlanetID {
	switch sfName {
	case "Sun":
		return models.PlanetSun
	case "Moon":
		return models.PlanetMoon
	case "Mercury":
		return models.PlanetMercury
	case "Venus":
		return models.PlanetVenus
	case "Mars":
		return models.PlanetMars
	case "Jupiter":
		return models.PlanetJupiter
	case "Saturn":
		return models.PlanetSaturn
	case "Uranus":
		return models.PlanetUranus
	case "Neptune":
		return models.PlanetNeptune
	case "Pluto":
		return models.PlanetPluto
	case "Chiron":
		return models.PlanetChiron
	case "NorthNode":
		return models.PlanetNorthNodeMean
	default:
		return ""
	}
}

// MapSFPointName converts Solar Fire point name to models.SpecialPointID
func MapSFPointName(sfName string) models.SpecialPointID {
	switch sfName {
	case "ASC":
		return models.PointASC
	case "MC":
		return models.PointMC
	default:
		return ""
	}
}

// ParseSFMetadata extracts birth data from Solar Fire meta file
// Returns: (birth date string, latitude, longitude, error)
func ParseSFMetadata(metaPath string) (birthDate string, lat, lon float64, err error) {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return "", 0, 0, fmt.Errorf("read metadata: %w", err)
	}

	text := string(data)
	lines := strings.Split(text, "\n")

	// Example format from testcase-1-meta.txt:
	// "Mar 17 2026 Solar Fire v9.0.29 Page 1"
	// "*** CHART ANALYSIS REPORT ***"
	// "JN - Male Chart"
	// ...
	// "DeltaT = +62s; ET = 9:37:02 am Dec 18 1997; JDE = 2450800.900729"
	// ...
	// No explicit lat/lon in meta file, so we use defaults from test data

	// Extract birth date from line with "ET = "
	etPattern := regexp.MustCompile(`ET = [^;]*(\w+ \d+ \d{4})`)
	for _, line := range lines {
		matches := etPattern.FindStringSubmatch(line)
		if len(matches) > 1 {
			birthDate = matches[1] // e.g., "Dec 18 1997"
			break
		}
	}

	if birthDate == "" {
		// Fallback: extract from first line
		// "Mar 17 2026 Solar Fire v9.0.29 Page 1" → but this is report date, not birth date
		// For now, use hard-coded test data
		birthDate = "" // Will be provided by test
	}

	// For Phase D, lat/lon are from test constants (jnLat, jnLon, etc.)
	return birthDate, lat, lon, nil
}

// BuildBodiesFromPlanets converts a PlanetPosition slice to aspect.Body slice.
// Used for SP and SA biwheel tests where CalcDoubleChart cannot be used directly.
func BuildBodiesFromPlanets(planets []models.PlanetPosition) []aspect.Body {
	bodies := make([]aspect.Body, 0, len(planets))
	for _, p := range planets {
		bodies = append(bodies, aspect.Body{
			ID:        string(p.PlanetID),
			Longitude: p.Longitude,
			Speed:     p.Speed,
		})
	}
	return bodies
}
