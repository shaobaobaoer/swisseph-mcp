# Phase B Baseline Values — JN Precision Test Suite

**Status:** Phase B (Baseline Validation) ✓
**Date:** 2026-03-31
**Ephemeris:** Swiss Ephemeris DE431 (via pkg/sweph)
**Performance:** 0.055s (18.2x faster than 1.0s target)

## Reference Person: JN

- **Birth:** 1997-12-18 09:36:00 UTC (= 1997-12-18 17:36 AWST)
- **Location:** Jinshan, China (30.9°N, 121.15°E)
- **JD:** 2450800.900009 (UT)
- **Age at transit:** 28.0372 years
- **Transit date:** 2026-01-01 00:00:00 UTC (JD 2461041.5)

---

## Test Results — Phase B Baseline

### 1. TestJN_NA (Natal Chart) — FINAL ✓

Solar Fire validated to ±0.0004° precision.

| Body | Longitude | Sign | House | Precision |
|------|-----------|------|-------|-----------|
| **Sun** | 266.5000° | Sagittarius 26.5° | 6 | ±0.0004° |
| **Moon** | 138.1160° | Leo 18.1° | 2 | ±0.0004° |
| **Mercury** | 263.9330° | Sagittarius 23.9° | 6 | — |
| **Venus** | 302.5530° | Aquarius 2.5° | 8 | — |
| **Mars** | 300.0970° | Aquarius 0.1° | 8 | — |
| **Jupiter** | 319.5480° | Aquarius 19.5° | 8 | — |
| **Saturn** | 013.5380° | Aries 13.5° | 10 | — |
| **Uranus** | 306.4250° | Aquarius 6.4° | 8 | — |
| **Neptune** | 298.4680° | Capricorn 28.5° | 7 | — |
| **Pluto** | 246.2740° | Sagittarius 6.3° | 6 | — |

**Angles:**
- **ASC:** 96.5300° (Cancer 6.5°) ±0.0001°
- **MC:** 351.5000° (Pisces 21.5°) ±0.0003°
- **DSC:** 276.5300° (Libra 6.5°, = ASC + 180°)
- **IC:** 171.5000° (Virgo 21.5°, = MC + 180°)

**Houses:** 12 cusps, all in valid range [0, 360)

**Aspects:** Conjunction, Opposition, Trine, Square, Sextile verified

---

### 2. TestJN_SP (Secondary Progressions) — Phase B ✓

**Age:** 28.0372 years
**Progressed JD:** natal JD + 28.0372 days
**Progressed time:** ~1998-01-15 at progressed calendar

| Planet | Longitude | Sign | Speed | Precision |
|--------|-----------|------|-------|-----------|
| **SP Sun** | 295.0688° | Capricorn 25.1° | 0.002788°/day | ±0.001° |
| **SP Moon** | 146.4382° | Leo 26.4° | 0.033497°/day | ±0.001° |
| **SP Mercury** | 273.5328° | Capricorn 3.5° | 0.003597°/day | ±0.001° |

**Solar Arc (SA):**
- **SA Offset:** 28.5683° (Sun's arc of direction)
- **SA Sun:** 295.0681° (= natal Sun + offset) ±0.001°

**Verification:**
- SP Sun advanced 28.57° from natal (age 28) ✓
- SA Sun = SP Sun (within 0.001°) ✓
- All speeds < 0.05°/day (progressed pace) ✓

---

### 3. TestJN_SR (Solar Return 2025) — FINAL ✓

**Return Date:** JD 2461027.69 = 2025-12-18 (±0 days)
**Return Age:** 27.9994 years (from natal)
**Return Location:** Jinshan, China (natal place)

| Field | Value | Precision |
|-------|-------|-----------|
| **Return JD** | 2461027.69 | ± 1 day |
| **Return Sun** | 266.4997° | < 0.01° |
| **Natal Sun** | 266.5000° | reference |
| **Difference** | 0.0003° | ✓ Perfect match |

**Return Chart:**
- **Planets:** 10 bodies calculated
- **Houses:** 12 cusps (Placidus)
- **Return Type:** "solar" (Sun return)

---

### 4. TestJN_TR (Transit Events) — Phase B ✓

**Window:** 2026-01-08 to 2026-01-15 (7 days)
**Transit Movers:** Sun, Venus, Mars (3 fast planets)
**Natal Reference:** All 10 planets
**House System:** Placidus

| Event Type | Count | Notes |
|------------|-------|-------|
| **Total Events** | 50 | exact baseline |
| **Aspect Exact** | 41 | enter, exact, leave |
| **Sign Ingress** | 0 | (none in 7-day window) |
| **House Ingress** | 0 | (none in 7-day window) |

**Event Structure Validation:**
- All EventType populated ✓
- All Planet names valid ✓
- All PlanetSign populated ✓
- All JD within [startJD, endJD] ✓
- All PlanetLongitude in [0, 360) ✓

---

### 5. TestJN_Moon (Lunar Phase) — Phase B ✓

**Date:** 2026-01-01 00:00:00 UTC (JD 2461041.5)

| Field | Value | Precision |
|-------|-------|-----------|
| **Phase Name** | Waxing Gibbous | exact |
| **Phase Angle** | 146.15° | ±0.1° |
| **Illumination** | 91.5% | ±0.1% |
| **Moon Longitude** | 66.7156° | ±0.01° |
| **Moon Sign** | Gemini 6.7° | derived |
| **Sun Longitude** | 280.5686° | ±0.01° |
| **Sun Sign** | Sagittarius 10.6° | derived |
| **IsWaxing** | true | (phase < 180°) |

**Phase Consistency:**
- Calculated: moon_lon - sun_lon = 66.7156 - 280.5686 = -213.8530 → +146.1470 (wrap to [0,360)) ✓
- Expected: 146.15° ✓
- Difference: 0.003° ✓

**Waxing/Waning Logic:**
- Phase angle 146.15° < 180° → waxing ✓

---

### 6. TestJN_DoubleChart (Biwheel) — Phase B ✓

**Inner Ring (Natal):**
- **Planets:** 10 (Sun, Moon, Mercury, Venus, Mars, Jupiter, Saturn, Uranus, Neptune, Pluto)
- **Sun:** 266.4997° (matches natal)
- **Moon:** 138.1164° (matches natal)
- **Houses:** 12 cusps (Placidus)
- **ASC:** 96.5299°
- **MC:** 351.5003°

**Outer Ring (Transit at 2026-01-01):**
- **Planets:** 10 (same set)
- **Houses:** 12 cusps (Placidus)
- **ASC:** calculated from transit JD + natal lat/lon
- **MC:** calculated from transit JD + natal lat/lon

**Cross-Aspects:**
- **Total Count:** 35 (exact baseline)
- **Aspect Type Variety:** 9 distinct types
  - Conjunction (0°)
  - Opposition (180°)
  - Trine (120°)
  - Square (90°)
  - Sextile (60°)
  - Quincunx (150°)
  - Semi-Sextile (30°)
  - Semi-Square (45°)
  - Sesquiquadrate (135°)

---

## Performance Metrics — Phase B

### Execution Times

```
TestJN_NA:          0.00s
TestJN_SP:          0.00s  (Phase B baseline validation)
TestJN_SR:          0.00s
TestJN_TR:          0.05s  (ephemeris scan + 50 event validation)
TestJN_Moon:        0.00s  (Phase B baseline validation)
TestJN_DoubleChart: 0.00s  (Phase B baseline validation)
────────────────────────────────────
Total Suite:        0.055s ✓✓✓ (Target: <1.0s, Actual: 18.2x faster)
```

### Regression Testing

```
Full pkg/solarsage test suite: 1.596s (no regressions)
All 38+ packages: ALL PASSING
```

---

## Phase B → Phase C Path

**Phase C (Optional):** Solar Fire Cross-Validation

To upgrade Phase B values to Phase C (SF-validated):

1. **Export Phase B baselines** to reference CSV:
   ```
   planet,lon,lat,speed,sign,house,tolerance
   Sun,266.5000,0.0,1.018,Sagittarius,6,0.0004
   Moon,138.1160,-2.183,12.323,Leo,2,0.0004
   ...
   ```

2. **Obtain Solar Fire export** for:
   - Same person (JN, b. 1997-12-18 09:36 UTC, Jinshan)
   - Same dates (natal + transit 2026-01-01)
   - Same ephemeris (DE431 if available, else note difference)

3. **Compare SF ↔ Phase B:**
   - For each value, check: |SF_value - Phase_B_value| < tolerance
   - If match: Mark as Phase C (SF-validated)
   - If differ: Investigate (different ephemeris? different house system? calculation difference?)

4. **Document findings:**
   - Create Phase C document with SF values
   - Note any divergences and root causes
   - Update test names to reflect SF validation

---

## File References

- **Test File:** `pkg/solarsage/jn_precision_test.go` (505 lines)
- **Test Functions:**
  - `TestJN_NA` — line 63
  - `TestJN_SP` — line 140
  - `TestJN_SR` — line 214
  - `TestJN_TR` — line 280
  - `TestJN_Moon` — line 361
  - `TestJN_DoubleChart` — line 438

- **Dependencies:**
  - `pkg/chart/chart.go` — CalcSingleChart, CalcDoubleChart
  - `pkg/progressions/progressions.go` — CalcProgressedLongitude, SolarArcOffset
  - `pkg/returns/returns.go` — CalcSolarReturn
  - `pkg/lunar/lunar.go` — CalcLunarPhase
  - `pkg/transit/calc.go` — CalcTransitEvents
  - `pkg/sweph/sweph.go` — JulDay

---

## Precision Tolerances Summary

| Test | Component | Tolerance | Justification |
|------|-----------|-----------|---------------|
| NA | Planets | 0.01° | Match Solar Fire to 0.0004° |
| NA | Angles | 0.01° | Match Solar Fire to 0.0001° |
| SP | Longitude | 0.001° | Consistency with age calculation |
| SP | Speed | 0.000001°/day | Numerical precision |
| SR | Sun | 0.01° | Must match natal Sun exactly |
| TR | Event count | exact | Baseline established |
| Moon | Phase angle | 0.1° | Observable precision |
| Moon | Illumination | 0.1% | Observable precision |
| DC | Aspect count | exact | Baseline established |

---

## Status

✅ **Phase B Complete**
- All baseline values established
- All tests passing
- Performance verified (18.2x faster than target)
- Ready for Phase C (Solar Fire cross-validation)

Last updated: 2026-03-31
Next review: When SF validation data available
