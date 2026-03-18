package geo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ringsaturn/tzf"

	"github.com/anthropic/swisseph-mcp/pkg/models"
)

var finder tzf.F

func init() {
	var err error
	finder, err = tzf.NewDefaultFinder()
	if err != nil {
		panic(fmt.Sprintf("failed to init timezone finder: %v", err))
	}
}

// TimezoneFromCoords returns the IANA timezone for a given latitude/longitude.
func TimezoneFromCoords(lat, lon float64) string {
	tz := finder.GetTimezoneName(lon, lat)
	if tz == "" {
		// Fallback: rough estimate from longitude
		offset := int(lon / 15.0)
		if offset == 0 {
			return "UTC"
		}
		if offset > 0 {
			return fmt.Sprintf("Etc/GMT-%d", offset)
		}
		return fmt.Sprintf("Etc/GMT+%d", -offset)
	}
	return tz
}

// nominatimResult represents the Nominatim API response
type nominatimResult struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

// Geocode converts a location name to geographic coordinates using Nominatim.
func Geocode(locationName string) (*models.GeoLocation, error) {
	name := strings.TrimSpace(locationName)
	if name == "" {
		return nil, fmt.Errorf("empty location name")
	}
	return geocodeNominatim(name)
}

// httpClient is the HTTP client used for geocoding requests; injectable for testing
var httpClient = &http.Client{Timeout: 10 * time.Second}

func geocodeNominatim(locationName string) (*models.GeoLocation, error) {
	u := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1",
		url.QueryEscape(locationName))

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "SwissephMCP/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geocode request failed: %w", err)
	}
	defer resp.Body.Close()

	var results []nominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("geocode decode failed: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("location not found: %s", locationName)
	}

	lat, err := strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude %q: %w", results[0].Lat, err)
	}
	lon, err := strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude %q: %w", results[0].Lon, err)
	}

	tz := TimezoneFromCoords(lat, lon)

	return &models.GeoLocation{
		Latitude:    lat,
		Longitude:   lon,
		Timezone:    tz,
		DisplayName: results[0].DisplayName,
	}, nil
}
