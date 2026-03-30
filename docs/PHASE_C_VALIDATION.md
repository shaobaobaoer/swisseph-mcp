# Phase C Validation — Solar Fire Cross-Validation Protocol

**Status:** Planning Phase (Preparation Complete, Awaiting SF Data)
**Date:** 2026-03-31
**Purpose:** Cross-validate Phase B baselines against Solar Fire reference data

---

## Overview

Phase C validates the Phase B baseline values by comparing them against Solar Fire 9 export data for the same natal charts and transit dates.

**When to proceed:**
- Solar Fire reference exports are obtained for JN and XB
- Ephemeris version used in SF export is documented (DE200/DE406/DE431)
- Import process is executed per Section 2 below

---

## Data Sources

### Phase B Baselines (Complete ✓)

**Source:** `docs/PHASE_B_BASELINES.csv`

Exported Phase B values:
- JN (Male, 1997-12-18 09:36 UTC, Jinshan, China)
- XB (Female, 1996-08-03 00:30 AWST, Huzhou, China)
- Chart types: NA (natal), SP (secondary progressions), SR (solar return), TR (transit), Moon (lunar phase), DC (double chart)
- Tolerances: Established per component

### Solar Fire Reference (To Be Obtained)

Required SF exports:
1. **JN Natal** — Birth chart (1997-12-18 09:36:00 UTC, 30.9°N 121.15°E)
   - All 10 planets
   - Angles (ASC, MC, DSC, IC)
   - House cusps (Placidus)

2. **JN Secondary Progressions** — Age 28.04 years
   - Sun, Moon, Mercury positions at progressed date
   - Solar arc offset

3. **JN Solar Return 2025** — Return date ≈ Dec 18, 2025
   - Return date (JD)
   - Return Sun position
   - All planets + angles

4. **JN Transits** — Window 2026-01-08 to 2026-01-15
   - Transit events (enter/exact/leave aspects)
   - Event count
   - Event types

5. **JN Lunar Phase** — 2026-01-01 00:00 UTC
   - Phase name
   - Phase angle
   - Illumination percentage
   - Moon longitude

6. **XB Natal** — Birth chart (1996-08-03 00:30 AWST = 1996-08-02 16:30 UTC, 30.867°N 120.1°E)
   - All 10 planets
   - Angles + house cusps

7. **XB Secondary Progressions** — Age 29.41 years

8. **XB Solar Return 2026** — Return date ≈ Aug 3, 2026

9. **XB Transits** — Window 2026-01-08 to 2026-01-15

---

## Validation Procedure

### Step 1: Import Solar Fire Data

When SF exports are received, create reference files in `docs/` directory:

```
docs/SF_JN_NATAL.csv         # JN Solar Fire natal export
docs/SF_JN_SP.csv            # JN Solar Fire secondary progressions
docs/SF_JN_SR.csv            # JN Solar Fire solar return 2025
docs/SF_JN_TR.csv            # JN Solar Fire transit events
docs/SF_JN_MOON.csv          # JN Solar Fire lunar phase 2026-01-01
docs/SF_XB_NATAL.csv         # XB Solar Fire natal export
docs/SF_XB_SP.csv            # XB Solar Fire secondary progressions
docs/SF_XB_SR.csv            # XB Solar Fire solar return 2026
docs/SF_XB_TR.csv            # XB Solar Fire transit events
docs/SF_XB_MOON.csv          # XB Solar Fire lunar phase 2026-01-01
```

### Step 2: Comparison Process

For each component (NA, SP, SR, TR, Moon, DC):

#### NA (Natal Chart)

**Compare:** Phase B longitude vs SF longitude

```
For each planet (Sun, Moon, Mercury, Venus, Mars, Jupiter, Saturn, Uranus, Neptune, Pluto):
  diff = |Phase_B_lon - SF_lon|
  If diff > tolerance:
    ✗ DIVERGENCE: document (see Section 3)
  Else:
    ✓ MATCH: within tolerance
```

**Tolerances:**
- Sun, Moon: ±0.0004° (high precision)
- Mercury–Pluto: ±0.01°
- Angles (ASC, MC): ±0.0001° – ±0.0003°

**Example:**
```
Planet: Sun
Phase B: 266.5000°
SF:      266.5001°
Diff:    0.0001° ✓ MATCH (within ±0.0004°)
```

#### SP (Secondary Progressions)

**Compare:** Phase B SP lon vs SF SP lon

```
For each progressed planet (Sun, Moon, Mercury):
  diff = |Phase_B_SP_lon - SF_SP_lon|
  If diff > 0.001°:
    ✗ DIVERGENCE
  Else:
    ✓ MATCH
```

**Also validate:** Solar Arc offset matches (within ±0.001°)

#### SR (Solar Return)

**Compare:**
- Return date (JD): within ±1 day
- Return Sun: within ±0.01° of natal Sun
- Return Sun matches natal exactly (< 0.01° diff)

#### TR (Transit Events)

**Compare:** Event counts

```
Phase B event count: N
SF event count: M

If N == M:
  ✓ COUNT MATCH
Else:
  ? INVESTIGATE: difference may indicate:
    - Different transit movers (Phase B uses fast planets only)
    - Different orb settings
    - Different aspect type filters
```

**Event details (if SF provides structure):**
- Event JD within window
- Event type (enter/exact/leave)
- Planet names and signs
- Aspect types

#### Moon (Lunar Phase)

**Compare:**
- Phase Name (text match or close phase range)
- Phase Angle: within ±0.1°
- Illumination: within ±0.1%
- Moon Longitude: within ±0.01°

#### DC (Double Chart / Biwheel)

**Compare:** Cross-aspect counts

```
Phase B cross-aspects: 35 (JN) or 52 (XB)
SF cross-aspects: M

If counts differ:
  ? INVESTIGATE: difference may indicate:
    - Different aspect type filters
    - Different orb settings
    - Different house system implementation
```

### Step 3: Document Divergences

If any comparison fails tolerance (Step 2), investigate root cause:

**Common causes:**

| Divergence | Likely Cause | Action |
|---|---|---|
| Natal longitude off by 0.01°–0.1° | Different ephemeris (DE200 vs DE431) | Document ephemeris used |
| Progressed position off | Age calculation difference | Check natal JD precision |
| Return date off by days | Different return definition | Check return algorithm |
| Transit count mismatch | Different movers or orbs | Document SF settings used |
| Phase angle off by 0.5°+ | Different house system affecting houses | Verify both use Placidus |

**Document in:** `docs/PHASE_C_DIVERGENCES.md` (create if divergences found)

---

## Success Criteria

Phase C is **VALIDATED** when:

- [  ] JN Natal: All 13 components match within tolerance
- [  ] JN SP: All 4 components match within tolerance
- [  ] JN SR: Return date + Sun position match
- [  ] JN TR: Event count matches or difference explained
- [  ] JN Moon: All 4 components match within tolerance
- [  ] XB Natal: All 13 components match within tolerance
- [  ] XB SP: All 3 components match within tolerance
- [  ] XB SR: Return date + Sun position match
- [  ] XB TR: Event count matches or difference explained
- [  ] XB Moon: All 4 components match within tolerance
- [  ] All divergences documented (if any)
- [  ] Ephemeris version documented
- [  ] Comparison methodology documented

---

## Phase C → Phase D Path (Optional)

If Phase C validation is successful:

1. **Rename Tests to Phase C**
   ```go
   TestJN_NA → TestJN_NA_SolarFireValidated
   // ... all 12 tests
   ```

2. **Update Documentation**
   - Rename `PHASE_B_BASELINE.md` → `PHASE_C_SOLAR_FIRE_VALIDATED.md`
   - Create `PHASE_C_VALIDATION_REPORT.md` with results

3. **Archive Phase B**
   - Move `PHASE_B_*.md` → `archive/`
   - Preserve git history (no deletion, just archival)

4. **Production Readiness: Phase D**
   - SF-validated tests become production reference
   - Ready for feature branch merging to main
   - Suitable for public release documentation

---

## Files for Phase C

**Input (When Obtained):**
- Solar Fire exports (SF_*.csv)

**Output (To Create):**
- `docs/PHASE_C_COMPARISON_RESULTS.md` — Results summary
- `docs/PHASE_C_DIVERGENCES.md` — If divergences found (optional)
- Updated test file with Phase C assertions (if validation succeeds)
- Updated documentation (this file + PHASE_B_BASELINE.md)

**Current Status:**
- ✓ Phase B baselines exported to `docs/PHASE_B_BASELINES.csv`
- ✓ Validation procedure documented
- ⏳ Awaiting Solar Fire data

---

## Ephemeris Reference

Swiss Ephemeris DE431 is used in Phase B. If Solar Fire uses a different ephemeris, document:

| Ephemeris | Period | Precision | Notes |
|---|---|---|---|
| DE200 | 1600–2200 | ~0.1" | Older, may show 0.01°–0.1° divergence |
| DE406 | 1900–2100 | ~0.01" | Intermediate, generally close |
| DE431 | 1900–2500 | ~0.1ms | Modern, should match Phase B closely |

If divergence > 0.05°, re-run Phase B tests with matching ephemeris if available.

---

## Timeline

- **Baseline (Phase B):** ✓ Complete 2026-03-31
- **Data Obtainment:** ⏳ Awaiting SF export
- **Validation (Phase C):** 1–2 hours once data available
- **Documentation:** ½ hour
- **Phase D Upgrade:** ½ hour if validation succeeds

---

**Status:** Ready for Phase C validation when Solar Fire data available
**Last Updated:** 2026-03-31
**Next Step:** Import SF data → Execute comparison → Document results
