// Package heliacal provides heliacal rising and setting calculations using
// the Swiss Ephemeris visibility algorithms. It determines when planets
// first become visible or disappear near the horizon relative to the Sun.
//
// Heliacal phenomena are the most ancient form of astronomical observation:
//   - Heliacal Rising  — a planet rises just before the Sun, becoming briefly
//     visible in the pre-dawn sky after a period of invisibility.
//   - Heliacal Setting — a planet sets just after the Sun, making its last
//     appearance in the evening sky before disappearing behind the Sun.
//   - Evening First    — first visibility of a planet in the western evening sky.
//   - Morning Last     — last visibility of a planet in the eastern morning sky.
//
// CalcHeliacalEvents searches for all four event types for the five classical
// visible planets (Mercury through Saturn) over a given date range and location.
// The Swiss Ephemeris swe_heliacal_ut function is used internally, which models
// atmospheric extinction, observer altitude, and horizon dip.
//
// Example:
//
//	result, err := heliacal.CalcHeliacalEvents(heliacal.HeliacalInput{
//	    Lat:     37.97,
//	    Lon:     23.72,
//	    StartJD: 2460676.5,
//	    EndJD:   2461041.5,
//	})
package heliacal
