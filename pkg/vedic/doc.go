// Package vedic provides sidereal chart calculations with support for
// multiple ayanamsa systems (Lahiri, Raman, Krishnamurti, Fagan-Bradley,
// Yukteshwar). Includes Nakshatra computation and Vimshottari Dasha periods.
//
// The sidereal zodiac differs from the tropical zodiac by the ayanamsa —
// the accumulated precession of the equinoxes since the original alignment
// of the two zodiacs (approximately 23–25° currently). Vedic astrology
// (Jyotish) uses the sidereal zodiac almost exclusively.
//
// Key functions:
//   - CalcSiderealChart  — full sidereal natal chart with Nakshatra data for each planet
//   - TropicalToSidereal — subtract the ayanamsa from a tropical longitude
//   - CalcNakshatra      — return the Nakshatra (lunar mansion), pada (quarter), and Vimshottari lord
//   - CalcVimshottariDasha — compute the full Maha Dasha sequence from the Moon's Nakshatra
//
// Supported ayanamsa systems:
//   - Lahiri       (Chitrapaksha) — the official standard of the Indian Government
//   - Raman        — system of B.V. Raman
//   - Krishnamurti — KP (Krishnamurti Paddhati) system
//   - Fagan-Bradley — Western sidereal astrology standard
//   - Yukteshwar   — system of Sri Yukteshwar Giri
package vedic
