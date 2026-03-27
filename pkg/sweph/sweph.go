package sweph

/*
#cgo CFLAGS: -I${SRCDIR}/../../third_party/swisseph -DTLSOFF

#include "swephexp.h"
#include "swehouse.h"
#include <stdlib.h>
#include <string.h>

// Wrapper to call swe_calc_ut with proper array handling
static int calc_ut(double jd_ut, int ipl, int iflag, double *xx, char *serr) {
    return swe_calc_ut(jd_ut, ipl, iflag, xx, serr);
}

// Wrapper for house calculation
static int houses(double jd_ut, double lat, double lon, int hsys, double *cusps, double *ascmc) {
    return swe_houses(jd_ut, lat, lon, hsys, cusps, ascmc);
}

// Wrapper for house position
static double house_pos(double armc, double geolat, double eps, int hsys, double *xpin, char *serr) {
    return swe_house_pos(armc, geolat, eps, hsys, xpin, serr);
}

// Wrapper for heliacal rising/setting
static int heliacal_ut(double jd_start, double *dgeo, double *datm, double *dobs,
                        char *object_name, int type_event, int iflag, double *dret, char *serr) {
    return swe_heliacal_ut(jd_start, dgeo, datm, dobs, object_name, type_event, iflag, dret, serr);
}

// Wrapper for house calculation with cusp speeds (swe_houses_ex2)
static int houses_ex2(double jd_ut, int iflag, double lat, double lon, int hsys,
                       double *cusps, double *ascmc, double *cusp_speed, double *ascmc_speed) {
    return swe_houses_ex2(jd_ut, iflag, lat, lon, hsys, cusps, ascmc, cusp_speed, ascmc_speed, NULL);
}

// Wrapper for sidereal mode + ayanamsa
static void set_sid_mode(int sid_mode, double t0, double ayan_t0) {
    swe_set_sid_mode(sid_mode, t0, ayan_t0);
}
static double get_ayanamsa_ut(double jd_ut) {
    return swe_get_ayanamsa_ut(jd_ut);
}
*/
import "C"
import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"
)

// Planet IDs matching Swiss Ephemeris constants
const (
	SE_SUN          = C.SE_SUN
	SE_MOON         = C.SE_MOON
	SE_MERCURY      = C.SE_MERCURY
	SE_VENUS        = C.SE_VENUS
	SE_MARS         = C.SE_MARS
	SE_JUPITER      = C.SE_JUPITER
	SE_SATURN       = C.SE_SATURN
	SE_URANUS       = C.SE_URANUS
	SE_NEPTUNE      = C.SE_NEPTUNE
	SE_PLUTO        = C.SE_PLUTO
	SE_CHIRON       = C.SE_CHIRON
	SE_TRUE_NODE    = C.SE_TRUE_NODE
	SE_MEAN_NODE    = C.SE_MEAN_NODE
	SE_MEAN_APOG    = C.SE_MEAN_APOG  // Mean Lilith
	SE_OSCU_APOG    = C.SE_OSCU_APOG  // True Lilith
)

// Calculation flags
const (
	SEFLG_JPLEPH   = C.SEFLG_JPLEPH // Use JPL ephemeris (DE200, DE406, DE431, etc.)
	SEFLG_SWIEPH   = C.SEFLG_SWIEPH // Use Swiss Ephemeris (default DE431)
	SEFLG_MOSEPH   = C.SEFLG_MOSEPH // Use Moshier ephemeris (low precision)
	SEFLG_SPEED    = C.SEFLG_SPEED
)

// House system chars
const (
	HousePlacidus       = 'P'
	HouseKoch           = 'K'
	HouseEqual          = 'E'
	HouseWholeSign      = 'W'
	HouseCampanus       = 'C'
	HouseRegiomontanus  = 'R'
	HousePorphyry       = 'O'
	HouseMorinus        = 'M'
	HouseTopocentric    = 'T' // Polich-Page
	HouseAlcabitius     = 'B'
	HouseMeridian       = 'X' // Axial rotation / Meridian
	HouseSripati        = 'S' // Sripati (Sri Pati) — traditional Indian system
)

// CalcResult holds the result of a planet calculation
type CalcResult struct {
	Longitude     float64
	Latitude      float64
	Distance      float64
	SpeedLong     float64
	SpeedLat      float64
	SpeedDist     float64
	IsRetrograde  bool
}

// HouseResult holds house cusp and angle data
type HouseResult struct {
	Cusps  [13]float64 // index 1-12 = house cusps
	ASC    float64
	MC     float64
	ARMC   float64
	Vertex float64
	EqASC  float64 // East Point (equatorial ascendant)
}

var mu sync.Mutex

// storedPath holds an absolute ephemeris path to prevent dangling pointer issues
var storedPath *C.char

// storedJPLFile holds the JPL ephemeris filename for SEFLG_JPLEPH mode
var storedJPLFile *C.char

// EphemerisType defines which ephemeris to use for calculations
type EphemerisType int

const (
	// EphemerisSwiss uses Swiss Ephemeris (default, DE431-based, highly accurate)
	EphemerisSwiss EphemerisType = iota
	// EphemerisJPL uses JPL ephemeris (DE200, DE406, DE431, DE440, etc.)
	// Requires JPL ephemeris file (.eph or unmerged file)
	EphemerisJPL
	// EphemerisMoshier uses Moshier ephemeris (built-in, lower precision, no files needed)
	EphemerisMoshier
)

// String returns the name of the ephemeris type
func (e EphemerisType) String() string {
	switch e {
	case EphemerisSwiss:
		return "Swiss Ephemeris (DE431)"
	case EphemerisJPL:
		return "JPL Ephemeris"
	case EphemerisMoshier:
		return "Moshier Ephemeris"
	default:
		return "Unknown"
	}
}

// currentEphemeris is the global ephemeris type setting
var currentEphemeris EphemerisType = EphemerisSwiss

// SetEphemerisType sets the global ephemeris type for all calculations.
// This affects all subsequent CalcUT calls.
func SetEphemerisType(ephType EphemerisType) {
	mu.Lock()
	defer mu.Unlock()
	currentEphemeris = ephType
}

// GetEphemerisType returns the current global ephemeris type.
func GetEphemerisType() EphemerisType {
	mu.Lock()
	defer mu.Unlock()
	return currentEphemeris
}

// Init sets the ephemeris path. Converts to absolute path for reliability.
// This path should contain Swiss Ephemeris files (.se1) or JPL ephemeris files.
func Init(ephePath string) {
	mu.Lock()
	defer mu.Unlock()
	// Convert to absolute path to survive working directory changes
	absPath, err := filepath.Abs(ephePath)
	if err != nil {
		absPath = ephePath
	}
	// Free previous stored path if any
	if storedPath != nil {
		C.free(unsafe.Pointer(storedPath))
	}
	// Keep the C string alive for the lifetime of the program
	storedPath = C.CString(absPath)
	C.swe_set_ephe_path(storedPath)
}

// SetJPLFile sets the JPL ephemeris filename for use with EphemerisJPL mode.
// Common filenames: "de406.eph", "de440.eph", "linux_p1550p2650.440" (DE440)
// The file must be in the ephemeris path set by Init().
func SetJPLFile(filename string) {
	mu.Lock()
	defer mu.Unlock()
	if storedJPLFile != nil {
		C.free(unsafe.Pointer(storedJPLFile))
	}
	storedJPLFile = C.CString(filename)
	C.swe_set_jpl_file(storedJPLFile)
}

// Close releases Swiss Ephemeris resources
func Close() {
	mu.Lock()
	defer mu.Unlock()
	C.swe_close()
}

// ConfigureFromEnv configures ephemeris settings from environment variables.
// Environment variables:
//   - SWISSEPH_TYPE: "swiss" (default), "jpl", or "moshier"
//   - SWISSEPH_JPL_FILE: JPL ephemeris filename (e.g., "de406.eph", "de440.eph")
//
// This function should be called after Init().
func ConfigureFromEnv() {
	ephType := os.Getenv("SWISSEPH_TYPE")
	switch strings.ToLower(ephType) {
	case "jpl":
		SetEphemerisType(EphemerisJPL)
		// Check for JPL file
		jplFile := os.Getenv("SWISSEPH_JPL_FILE")
		if jplFile != "" {
			SetJPLFile(jplFile)
		}
	case "moshier":
		SetEphemerisType(EphemerisMoshier)
	default:
		// "swiss" or empty = default Swiss Ephemeris
		SetEphemerisType(EphemerisSwiss)
	}
}

// getFlag returns the appropriate calculation flag based on current ephemeris type
func getFlag() int {
	switch currentEphemeris {
	case EphemerisJPL:
		return SEFLG_JPLEPH | SEFLG_SPEED
	case EphemerisMoshier:
		return SEFLG_MOSEPH | SEFLG_SPEED
	default:
		return SEFLG_SWIEPH | SEFLG_SPEED
	}
}

// CalcUT calculates planet position at given Julian Day UT
func CalcUT(jdUT float64, planet int) (*CalcResult, error) {
	mu.Lock()
	defer mu.Unlock()

	var xx [6]C.double
	var serr [256]C.char
	iflag := C.int(getFlag())

	ret := C.calc_ut(C.double(jdUT), C.int(planet), iflag, &xx[0], &serr[0])
	if ret < 0 {
		return nil, fmt.Errorf("swe_calc_ut error: %s", C.GoString(&serr[0]))
	}

	return &CalcResult{
		Longitude:    float64(xx[0]),
		Latitude:     float64(xx[1]),
		Distance:     float64(xx[2]),
		SpeedLong:    float64(xx[3]),
		SpeedLat:     float64(xx[4]),
		SpeedDist:    float64(xx[5]),
		IsRetrograde: float64(xx[3]) < 0,
	}, nil
}

// Houses calculates house cusps and angles
func Houses(jdUT float64, lat, lon float64, hsys int) (*HouseResult, error) {
	mu.Lock()
	defer mu.Unlock()

	var cusps [13]C.double
	var ascmc [10]C.double

	C.houses(C.double(jdUT), C.double(lat), C.double(lon), C.int(hsys), &cusps[0], &ascmc[0])

	result := &HouseResult{
		ASC:    float64(ascmc[0]),
		MC:     float64(ascmc[1]),
		ARMC:   float64(ascmc[2]),
		Vertex: float64(ascmc[3]),
		EqASC:  float64(ascmc[4]),
	}
	for i := 0; i < 13; i++ {
		result.Cusps[i] = float64(cusps[i])
	}
	return result, nil
}

// HousePos returns the house position (1.0-12.999) of a given ecliptic point
func HousePos(armc, geoLat, eps float64, hsys int, lon, lat float64) (float64, error) {
	mu.Lock()
	defer mu.Unlock()

	var xpin [2]C.double
	xpin[0] = C.double(lon)
	xpin[1] = C.double(lat)
	var serr [256]C.char

	pos := C.house_pos(C.double(armc), C.double(geoLat), C.double(eps), C.int(hsys), &xpin[0], &serr[0])
	if float64(pos) == 0 {
		return 0, fmt.Errorf("swe_house_pos error: %s", C.GoString(&serr[0]))
	}
	return float64(pos), nil
}

// HousesEx2Result extends HouseResult with cusp and angle speeds (degrees/day).
type HousesEx2Result struct {
	HouseResult
	CuspSpeeds [13]float64 // index 1-12 = cusp speeds in °/day
	ASCSpeed   float64
	MCSpeed    float64
}

// HousesEx2 calculates house cusps, angles, and their speeds using swe_houses_ex2.
// The speed data is useful for progressed cusps and dynamic chart animation.
func HousesEx2(jdUT float64, lat, lon float64, hsys int) (*HousesEx2Result, error) {
	mu.Lock()
	defer mu.Unlock()

	var cusps [13]C.double
	var ascmc [10]C.double
	var cuspSpeed [13]C.double
	var ascmcSpeed [10]C.double

	ret := C.houses_ex2(C.double(jdUT), C.int(0), C.double(lat), C.double(lon), C.int(hsys),
		&cusps[0], &ascmc[0], &cuspSpeed[0], &ascmcSpeed[0])
	if ret < 0 {
		return nil, fmt.Errorf("swe_houses_ex2 error for house system %c", hsys)
	}

	result := &HousesEx2Result{
		HouseResult: HouseResult{
			ASC:    float64(ascmc[0]),
			MC:     float64(ascmc[1]),
			ARMC:   float64(ascmc[2]),
			Vertex: float64(ascmc[3]),
			EqASC:  float64(ascmc[4]),
		},
		ASCSpeed: float64(ascmcSpeed[0]),
		MCSpeed:  float64(ascmcSpeed[1]),
	}
	for i := 0; i < 13; i++ {
		result.Cusps[i] = float64(cusps[i])
		result.CuspSpeeds[i] = float64(cuspSpeed[i])
	}
	return result, nil
}

// JulDay converts calendar date to Julian Day
func JulDay(year, month, day int, hour float64, gregorian bool) float64 {
	cal := C.int(C.SE_GREG_CAL)
	if !gregorian {
		cal = C.int(C.SE_JUL_CAL)
	}
	return float64(C.swe_julday(C.int(year), C.int(month), C.int(day), C.double(hour), cal))
}

// RevJul converts Julian Day back to calendar date
func RevJul(jd float64, gregorian bool) (year, month, day int, hour float64) {
	cal := C.int(C.SE_GREG_CAL)
	if !gregorian {
		cal = C.int(C.SE_JUL_CAL)
	}
	var y, m, d C.int
	var h C.double
	C.swe_revjul(C.double(jd), cal, &y, &m, &d, &h)
	return int(y), int(m), int(d), float64(h)
}

// DeltaT returns the difference TT - UT in days for a given JD UT
func DeltaT(jdUT float64) float64 {
	return float64(C.swe_deltat(C.double(jdUT)))
}

// Obliquity returns the obliquity of the ecliptic
func Obliquity(jdUT float64) (float64, error) {
	mu.Lock()
	defer mu.Unlock()

	var xx [6]C.double
	var serr [256]C.char
	ret := C.calc_ut(C.double(jdUT), C.int(C.SE_ECL_NUT), C.int(SEFLG_SWIEPH), &xx[0], &serr[0])
	if ret < 0 {
		return 0, fmt.Errorf("obliquity error: %s", C.GoString(&serr[0]))
	}
	return float64(xx[0]), nil
}

// NormalizeDegrees normalizes an angle to [0, 360)
func NormalizeDegrees(deg float64) float64 {
	deg = math.Mod(deg, 360.0)
	if deg < 0 {
		deg += 360.0
	}
	return deg
}

// Heliacal event type constants
const (
	SE_HELIACAL_RISING  = 1
	SE_HELIACAL_SETTING = 2
	SE_EVENING_FIRST    = 3
	SE_MORNING_LAST     = 4
)

// HeliacalResult holds the result of a heliacal event calculation
type HeliacalResult struct {
	JDStart   float64 // Start of visibility event
	JDOptimum float64 // Optimal observation time
	JDEnd     float64 // End of visibility event
}

// HeliacalUT finds the next heliacal event for an object after jdStart.
// geoLon, geoLat, geoAlt are the observer's geographic coordinates (degrees, meters).
// objectName is the planet name (e.g., "venus", "mercury").
// eventType is one of SE_HELIACAL_RISING, SE_HELIACAL_SETTING, SE_EVENING_FIRST, SE_MORNING_LAST.
func HeliacalUT(jdStart float64, geoLon, geoLat, geoAlt float64, objectName string, eventType int) (*HeliacalResult, error) {
	mu.Lock()
	defer mu.Unlock()

	var dgeo [3]C.double
	dgeo[0] = C.double(geoLon)
	dgeo[1] = C.double(geoLat)
	dgeo[2] = C.double(geoAlt)

	// Default atmospheric conditions
	var datm [4]C.double
	datm[0] = C.double(1013.25) // pressure (mbar)
	datm[1] = C.double(15)      // temperature (C)
	datm[2] = C.double(40)      // humidity (%)
	datm[3] = C.double(0.25)    // extinction coefficient (mag/airdeg)

	// Default observer parameters
	var dobs [6]C.double
	dobs[0] = C.double(36) // age
	dobs[1] = C.double(1)  // Snellen ratio
	dobs[2] = C.double(0)  // telescope aperture (mm), 0 = naked eye
	dobs[3] = C.double(0)  // magnification
	dobs[4] = C.double(0)  // optical type (0 = naked eye)
	dobs[5] = C.double(0)  // field of view (deg)

	cName := C.CString(objectName)
	defer C.free(unsafe.Pointer(cName))

	var dret [50]C.double
	var serr [256]C.char

	ret := C.heliacal_ut(C.double(jdStart), &dgeo[0], &datm[0], &dobs[0],
		cName, C.int(eventType), C.int(0), &dret[0], &serr[0])
	if ret < 0 {
		return nil, fmt.Errorf("swe_heliacal_ut error: %s", C.GoString(&serr[0]))
	}

	return &HeliacalResult{
		JDStart:   float64(dret[0]),
		JDOptimum: float64(dret[1]),
		JDEnd:     float64(dret[2]),
	}, nil
}

// Sidereal mode constants (Swiss Ephemeris SIDM_* values)
const (
	SidmLahiri      = 1  // Lahiri / Chitrapaksha
	SidmRaman       = 3  // B.V. Raman
	SidmKrishnamurti = 5 // Krishnamurti
	SidmFaganBradley = 0 // Fagan-Bradley (Western sidereal)
	SidmYukteshwar  = 7  // Sri Yukteshwar
	SidmTrueCitra   = 27 // True Citra (True Chitrapaksha)
	SidmTrueRevati  = 28 // True Revati
	SidmTruePushya  = 29 // True Pushya
	SidmTrueMula    = 35 // True Mula
)

// GetAyanamsaUT returns the precise ayanamsa value at a given Julian Day UT
// using the Swiss Ephemeris native computation for the given sidereal mode.
func GetAyanamsaUT(jdUT float64, sidMode int) float64 {
	mu.Lock()
	defer mu.Unlock()
	C.set_sid_mode(C.int(sidMode), 0, 0)
	return float64(C.get_ayanamsa_ut(C.double(jdUT)))
}
