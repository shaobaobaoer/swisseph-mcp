# Phase D Expansion Roadmap — Full Timeline Validation

**Status:** 🔍 DISCOVERY PHASE COMPLETE — EXPANSION ROADMAP READY  
**Date:** 2026-03-31  
**Scope Gap:** 0.6% → Target 100% coverage (4,389 additional events)

---

## Executive Summary

Initial Phase D snapshot validation revealed a **99.4% scope gap**:
- ✅ **Validated:** 25 events (0.6%)
- ❌ **Missing:** 4,389 events (99.4%)

This is not a failure—it's an opportunity for **comprehensive timeline validation** across all event types, chart pairings, and reference persons.

---

## Gap Analysis by Event Type

### Major Aspect Events (Enter/Exact/Leave)

**These form the transit lifecycle and are critical for accuracy:**

| Event Type | Count | Coverage | Status | Potency |
|---|---|---|---|---|
| **Enter** | 255 | 0% | ❌ Missing | Beginning of influence |
| **Exact** | 247 | 0% | ❌ Missing | ⭐ PEAK — Most Important |
| **Leave** | 253 | 0% | ❌ Missing | Fading influence |
| **Subtotal** | **755** | **0%** | ❌ Critical Gap | **Complete lifecycle** |

### Other Event Types

| Event Type | Count | Coverage | Status |
|---|---|---|---|
| Begin | 52 | 100% | ✅ Complete |
| Void | 161 | 0% | ❌ Missing |
| SignIngress | 169 | 0% | ❌ Missing |
| HouseChange | 7 | 0% | ❌ Missing |
| Retrograde | 6 | 0% | ❌ Missing |
| Direct | 6 | 0% | ❌ Missing |
| **Subtotal** | **407** | **12%** | Partially validated |

---

## Gap Analysis by Chart Type

### Primary Pairings (should validate first)

| Pairing | Total | Validated | Missing | Coverage |
|---|---|---|---|---|
| **Tr-Na** (Transit vs Natal) | 200 | 4 | 196 | 2% ❌ |
| **Sp-Na** (Secondary Progressions) | 52 | 17 | 35 | 33% ⚠️ |
| **Sa-Na** (Solar Arc) | 31 | 11 | 20 | 35% ⚠️ |
| **Subtotal** | **283** | **32** | **251** | **11%** |

### Advanced Pairings (currently unvalidated)

| Pairing | Total | Validated | Missing | Status |
|---|---|---|---|---|
| Sp-Sp | 25 | 0 | 25 | ❌ Not tested |
| Tr-Sp | 231 | 0 | 231 | ❌ Not tested |
| Tr-Sa | 211 | 0 | 211 | ❌ Not tested |
| Tr-Tr | 394 | 0 | 394 | ❌ Not tested |
| Other | 12 | 0 | 12 | ❌ Not tested |
| **Subtotal** | **873** | **0** | **873** | **0%** |

---

## Gap Analysis by Reference Person & Timeline

### JN (Male) — testcase-1

- **Total events:** 1,156
- **Validated:** 25 (2%)
- **Missing:** 1,131 (98%)
- **Timeline:** 2026-02-01 → 2027-01-30 (1 year)
- **Status:** Partially explored, mostly unvalidated

### XB (Female) — testcase-2-transit-1996-2001

- **Total events:** 1,746
- **Validated:** 0 (0%)
- **Missing:** 1,746 (100%)
- **Timeline:** 1996-08-03 → 2001-08-03 (5 years)
- **Status:** ❌ Completely unvalidated

### XB (Female) — testcase-2-transit-2001-2006

- **Total events:** 1,512
- **Validated:** 0 (0%)
- **Missing:** 1,512 (100%)
- **Timeline:** 2001-08-03 → 2006-08-03 (5 years)
- **Status:** ❌ Completely unvalidated

---

## Technical Requirements for Expansion

### 1. Timeline Event Matching

**Challenge:** SF records event phases (Enter → Exact → Leave) across dates, not just snapshots.

**Solution:** Implement timeline matcher that:
- Finds each SF event's occurrence date/time
- Computes equivalent SolarSage positions on that date
- Matches planet pairs within tolerance
- Validates event type (Enter/Exact/Leave semantics)

**Implementation effort:** Medium (requires date-based iteration)

### 2. Multi-Occurrence Handling

**Challenge:** Same aspect occurs multiple times:
```
Jupiter Quincunx Sun: 5 separate cycles
→ Need to match each to closest SolarSage occurrence
```

**Solution:** Implement occurrence matching:
- For each unique aspect pair in SF
- Find all SolarSage occurrences
- Match to nearest SF occurrence (by date/time/angle)

**Implementation effort:** Medium (requires aspect cycle tracking)

### 3. Event Type Semantics

**Challenge:** Event types have different meanings:
- **Enter:** Orb aspect begins forming (distance decreasing)
- **Exact:** Perfect aspect angle (distance = 0)
- **Leave:** Orb aspect separating (distance increasing)
- **Void:** Moon void of course (special condition)
- **SignIngress:** Crossing sign boundary

**Solution:** Event-specific validators:
- For Enter/Exact/Leave: angle difference check
- For Void: special Moon aspects check
- For SignIngress: sign boundary logic

**Implementation effort:** High (domain-specific rules)

### 4. Extended Timeline Support

**Challenge:** testcase-2 spans 5-10 years, not 1 day.

**Solution:** Efficient timeline iteration:
- Load all records from period
- Group by date
- Process each date's aspects
- Track multi-occurrence sequences

**Implementation effort:** Medium (data structure optimization)

### 5. Chart Pairing Diversity

**Challenge:** Tr-Sp, Tr-Sa, Tr-Tr require different computation:
- Tr-Sp: Transit to progressed (different outer ring computation)
- Tr-Sa: Transit to solar arc (offset-based positions)
- Tr-Tr: Transit to transit (both moving planets)

**Solution:** Pairing-specific validators:
- Tr-Na: Already working (baseline for others)
- Tr-Sp: Compute progressed positions for outer ring
- Tr-Sa: Compute solar arc positions for outer ring
- Tr-Tr: Both natal and transit moving

**Implementation effort:** High (different computation paths)

---

## Validation Roadmap

### Phase D v1 (Complete) — Snapshot Validation
- ✅ Single date analysis (2026-02-01)
- ✅ Begin events only (52 records)
- ✅ Primary pairings (Tr-Na, Sp-Na, Sa-Na)
- ✅ Aspect-level matching
- ✅ 75% match rate achieved

### Phase D v2 (NEXT) — Timeline Expansion - Stage 1

**Scope:** Full JN timeline (testcase-1)  
**Target:** 1,156 events across all event types

#### Milestone 1: Tr-Na Full Timeline
- [ ] Implement timeline matcher
- [ ] Validate all 200 Tr-Na events (not just 4 Begin)
- [ ] Include Enter (64), Exact (59), Leave (61) phases
- [ ] Report match rates by event type
- **Effort:** 3-4 days

#### Milestone 2: Sp-Na Full Timeline
- [ ] Apply timeline matcher to Sp-Na
- [ ] Validate all 52 events
- [ ] Handle progressed planet computation
- **Effort:** 2 days

#### Milestone 3: Sa-Na Full Timeline
- [ ] Apply timeline matcher to Sa-Na
- [ ] Validate all 31 events
- [ ] Handle solar arc computation
- **Effort:** 2 days

#### Milestone 4: Advanced Pairings
- [ ] Implement Tr-Sp matcher (231 events)
- [ ] Implement Tr-Sa matcher (211 events)
- [ ] Implement Tr-Tr matcher (394 events)
- [ ] Implement Sp-Sp matcher (25 events)
- **Effort:** 4-5 days

**Stage 1 Total:** ~10 days, 873 additional events validated

---

### Phase D v3 — Timeline Expansion - Stage 2

**Scope:** XB timeline 1996-2001 (testcase-2)  
**Target:** 1,746 events

- [ ] Reuse Stage 1 matchers
- [ ] Extend timeline support to 5-year spans
- [ ] Validate all 1,326 Tr-Na events (XB)
- [ ] Validate 217 Sp-Na events (XB)
- [ ] Validate 203 other pairing events (XB)
- **Effort:** 3-4 days (matchers already built)

---

### Phase D v4 — Timeline Expansion - Stage 3

**Scope:** XB timeline 2001-2006 (testcase-3)  
**Target:** 1,512 events

- [ ] Apply existing matchers
- [ ] Validate all 1,145 Tr-Na events (XB second period)
- [ ] Validate 171 Sp-Na events (XB second period)
- [ ] Validate 196 other pairing events (XB second period)
- **Effort:** 1-2 days (direct application)

---

## Implementation Strategy

### Phase D v2 Pseudo-Code

```
function ValidateTimelineAspects(sfCsvPath, natalJD, natalLat, natalLon):
    sfRecords = ParseSFCSV(sfCsvPath)
    
    // Group by aspect pair
    byAspectPair = GroupBy(sfRecords, (P1, Aspect, P2))
    
    for each uniqueAspect in byAspectPair:
        sfOccurrences = GetOccurrences(uniqueAspect)  // all instances
        
        // Find SolarSage equivalents
        ssOccurrences = []
        for each sfOccurrence in sfOccurrences:
            ssDate = sfOccurrence.Date
            
            // Compute bodies at that date
            natalChart = CalcSingleChart(natalJD)
            transitChart = CalcSingleChart(ssDate)
            crossAspects = FindCrossAspects(natalChart, transitChart)
            
            // Find matching aspect in SolarSage
            ssAspect = FindMatchingAspect(crossAspects, P1, P2, 
                                         sfOccurrence.AspectType)
            if ssAspect found:
                ssOccurrences.append((ssDate, ssAspect))
        
        // Validate each SF occurrence matched
        for i, sfOcc in enumerate(sfOccurrences):
            if i < len(ssOccurrences):
                ssOcc = ssOccurrences[i]
                ValidateMatch(sfOcc, ssOcc)  // angle, date, aspect type
            else:
                RecordDivergence(sfOcc)  // missing in SolarSage
    
    return ValidationReport(matches, divergences)
```

### Key Implementation Considerations

1. **Date iteration:** Loop over all unique dates in SF data
2. **Occurrence matching:** Handle multiple instances of same aspect
3. **Tolerance:** Allow ±1.0° angle tolerance, ±1 day time tolerance
4. **Event type specifics:** 
   - Enter: aspect just entering orb (angle decreasing)
   - Exact: minimum angle difference
   - Leave: aspect just leaving orb (angle increasing)

---

## Expected Outcomes

### Phase D v2 (1,156 events)
- **Coverage increase:** 0.6% → 27%
- **Events validated:** 1,156
- **Estimated completion:** 10 days
- **Quality gate:** 80%+ match rate (target 90%+)

### Phase D v3 (1,746 events)
- **Coverage increase:** 27% → 66%
- **Events validated:** 1,156 + 1,746 = 2,902
- **Estimated completion:** +3-4 days
- **Quality gate:** Consistent match rates with v2

### Phase D v4 (1,512 events)
- **Coverage increase:** 66% → 100%
- **Events validated:** 4,414 (all data)
- **Estimated completion:** +1-2 days
- **Quality gate:** Consistent match rates

---

## Success Criteria

### Minimum Viable Completion (Phase D v2)

- [ ] All 200 Tr-Na events validated (not just 4)
- [ ] 80%+ match rate for Tr-Na across all event types
- [ ] Include Exact events (247 most important)
- [ ] Document divergence patterns
- [ ] < 1s execution time for full timeline

### Full Completion (Phase D v1-4)

- [ ] All 4,414 events validated
- [ ] 80%+ match rate across all chart types
- [ ] Both reference persons (JN, XB) validated
- [ ] All event types understood
- [ ] Full timeline support (years, not days)
- [ ] Comprehensive divergence analysis

---

## Files Created/Modified

**Analysis files:**
- `docs/PHASE_D_MISSING_EVENTS_ANALYSIS.md` — Detailed scope gap
- `docs/PHASE_D_EXPANSION_ROADMAP.md` — This file
- `pkg/solarsage/phase_d_comprehensive_test.go` — Discovery test suite

**Upcoming (Phase D v2):**
- `pkg/solarsage/phase_d_timeline_validator.go` — Timeline matching logic
- `pkg/solarsage/phase_d_timeline_test.go` — Full timeline validation tests

---

## Next Steps

1. **Approve expansion scope** — Confirm targeting full 4,414 events
2. **Start Phase D v2, Milestone 1** — Tr-Na full timeline (200 events)
3. **Build timeline matcher** — Core logic for date-based validation
4. **Implement occurrence tracking** — Handle multi-instance aspects
5. **Generate comprehensive report** — All 4,414 events analyzed

---

**Status:** Ready to proceed with Phase D v2  
**Estimated Total Duration:** 14-18 days (full 4,414 event validation)  
**Quality Target:** 85%+ match rate across all data

---

**Last Updated:** 2026-03-31  
**Next Review:** After Phase D v2 Milestone 1 completion  
**Owner:** SolarSage Phase D Expansion Team
