# Phase D Analysis: Missing Events Discovered

**Status:** ⚠️ CRITICAL SCOPE EXPANSION REQUIRED  
**Date:** 2026-03-31  
**Finding:** Phase D snapshot validation only covered **2% of actual test data**

---

## The Discovery

Initial Phase D validation only checked:
- **52 "Begin" events** at snapshot date (2026-02-01)
- Full test data contains: **4,414 total events**

This is a **99% scope gap**.

---

## Test Data Breakdown

### testcase-1 (JN) — 1,156 events total

**Tr-Na (Transit vs Natal): 200 events**
| Event Type | Count | Status |
|---|---|---|
| Begin | 4 | ✅ Validated |
| Enter | 64 | ❌ Missing |
| Exact | 59 | ❌ Missing |
| Leave | 61 | ❌ Missing |
| **Total** | **188** | **2% coverage** |

**Sp-Na (Secondary Progressions): 52 events**
| Event Type | Count | Status |
|---|---|---|
| Begin | 17 | ✅ Validated (58.8% match) |
| Enter | 11 | ❌ Missing |
| Exact | 9 | ❌ Missing |
| Leave | 12 | ❌ Missing |
| HouseChange | 1 | ❌ Missing |
| SignIngress | 2 | ❌ Missing |
| **Total** | **52** | **33% coverage** |

**Sa-Na (Solar Arc): 31 events**
| Event Type | Count | Status |
|---|---|---|
| Begin | 11 | ✅ Validated (100% match) |
| Enter | 9 | ❌ Missing |
| Exact | 4 | ❌ Missing |
| Leave | 7 | ❌ Missing |
| **Total** | **31** | **35% coverage** |

**Other Chart Type Pairings (missing entirely):**
- Sp-Sp: 25 events
- Tr-Sp: 231 events
- Tr-Sa: 211 events
- Tr-Tr: 394 events
- Tr (standalone): 12 events

---

### testcase-2-transit-1996-2001 (XB) — 1,746 events total

**Tr-Na: 1,326 events**
- **NOT VALIDATED AT ALL**

**Sp-Na: 217 events**
- **NOT VALIDATED AT ALL**

**Other pairings: 203 events**
- **NOT VALIDATED AT ALL**

---

### testcase-2-transit-2001-2006 (XB) — 1,512 events total

**Tr-Na: 1,145 events**
- **NOT VALIDATED AT ALL**

**Sp-Na: 171 events**
- **NOT VALIDATED AT ALL**

**Other pairings: 196 events**
- **NOT VALIDATED AT ALL**

---

## Why Events Are Missing

### Event Type Lifecycle

Each transit aspect goes through phases:

```
Timeline:
  [DATE 1]           [DATE 2]           [DATE 3]           [DATE 4]
  Enter Orb          Exact Peak         Leave Orb          Next Aspect
  ↓                  ↓                  ↓                  ↓
  Enter event        Exact event        Leave event        ---
  (planet getting    (perfect angle)    (planet leaving)
   closer)           (most potent)
```

**Solar Fire records all 4 phases.** SolarSage snapshot testing only checks "Begin" events at a single date.

### Event Type Distribution (JN testcase-1)

- **Enter: 255 events** (planets entering orb)
- **Leave: 253 events** (planets leaving orb)
- **Exact: 247 events** (peak aspects) ⭐ HIGHEST POTENCY
- **Void: 161 events** (void of course Moon)
- **SignIngress: 169 events** (sign changes)
- **Begin: 52 events** (snapshot analysis start)
- **HouseChange: 7 events** (house cusps)
- **Retrograde: 6 events** (planetary stations)
- **Direct: 6 events** (planetary stations)

---

## Why This Matters

### 1. Exact Events Are Most Important (247 events)
These represent **peak transits** where the aspect is most potent. We're not validating ANY of these.

**Example Exact events missing:**
```
Jupiter Quincunx Sun: 5 occurrences in JN timeline
Jupiter Trine Chiron: 5 occurrences
Saturn Conjunction Saturn: 4 occurrences
```

### 2. Enter/Leave Events Complete the Picture (508 events)
Together with Exact, these form the **complete transit cycle**:
- Enter = beginning of influence
- Exact = peak influence
- Leave = fading of influence

SolarSage currently cannot validate the **timing and sequencing** of these phases.

### 3. Broader Chart Pairings Ignored (869+ events)
We only tested Tr-Na/Sp-Na/Sa-Na but SF data includes:
- **Tr-Sp**: 231 events (transit to secondary progressions)
- **Tr-Sa**: 211 events (transit to solar arc)
- **Tr-Tr**: 394 events (transit to transit — rare but important)
- **Sp-Sp**: 25 events (progressed to progressed)

These test different computational paths that we haven't validated.

---

## What Needs to Be Done

### Phase D Expansion (3 stages)

#### Stage 1: Full JN Timeline (testcase-1) — 1,156 events
1. **Tr-Na**: Validate all 188 events (Begin/Enter/Exact/Leave)
2. **Sp-Na**: Validate all 52 events
3. **Sa-Na**: Validate all 31 events
4. **Tr-Sp, Tr-Sa, Tr-Tr, Sp-Sp**: Validate 869 events across mixed pairings

#### Stage 2: XB Timeline 1996-2001 (testcase-2) — 1,746 events
1. **Tr-Na**: 1,326 events across 5-year span
2. **Sp-Na**: 217 events
3. Other pairings: 203 events

#### Stage 3: XB Timeline 2001-2006 (testcase-3) — 1,512 events
1. **Tr-Na**: 1,145 events across 5-year span
2. **Sp-Na**: 171 events
3. Other pairings: 196 events

**Total scope: 4,414 events** requiring validation

---

## Technical Challenges for Full Timeline Validation

### Challenge 1: Event Timing
- **Begin events**: Easy — single snapshot
- **Enter events**: Need to find when aspect enters orb (date/time computation)
- **Exact events**: Need to find exact aspect moment (minor root-finding needed)
- **Leave events**: When aspect exits orb

SolarSage currently computes aspects at a given JD, not the timeline of event phases.

### Challenge 2: Multiple Occurrences
Some aspects repeat multiple times in timeline:
```
Jupiter Quincunx Sun: 5 times in JN span
→ Need to match each SF occurrence to nearest SolarSage computed occurrence
```

### Challenge 3: Event Type Semantics
- **SignIngress** (169): When Moon/planets enter new sign
- **Void** (161): Void of course conditions
- **HouseChange** (7): When planets cross house cusps
- **Retrograde/Direct** (12): Planetary station events

These are **not standard aspects** and require different validation logic.

---

## Immediate Action Items

1. **Identify why SolarSage isn't computing these events**
   - Are Enter/Exact/Leave events being filtered out somewhere?
   - Are multi-date aspect cycles being missed?
   - Is there a date range issue?

2. **Create event-level test validators**
   - For each event type (Enter, Exact, Leave)
   - Validate planet pairs and aspect types match
   - Check date/time accuracy

3. **Expand to multi-period timeline**
   - testcase-2 spans 1996-2001 (5 years)
   - testcase-3 spans 2001-2006 (5 years)
   - Need date/time matching across ranges

4. **Handle non-aspect events**
   - SignIngress, Void, HouseChange, Retrograde, Direct
   - Decide if these need separate validation paths

---

## Questions Requiring Investigation

1. **Why are we missing 1,104 Tr-Na events?**
   - Only 52 Begin events at snapshot
   - Where are the 188 Tr-Na events across full timeline?

2. **Why only 17/52 Sp-Na events validated?**
   - Are secondary progression events different?
   - Different computation path?

3. **What about the 869 events in Tr-Sp, Tr-Sa, Tr-Tr, Sp-Sp?**
   - Are these different computation paths?
   - Computational complexity higher?

4. **Why don't testcase-2 and testcase-3 data exist?**
   - XB has 1,746 + 1,512 = 3,258 events
   - Zero validation coverage

---

## Recommended Implementation Path

### Phase D v2: Progressive Expansion

**Week 1: Stage 1 (JN timeline)**
1. Add `TimelineValidator` to compute aspects across date ranges
2. Implement Enter/Exact/Leave event matchers
3. Validate Tr-Na full 188 events
4. Report divergences and patterns

**Week 2: Expand to other pairings**
1. Add Sp-Na full validation (52 events)
2. Add Sa-Na full validation (31 events)
3. Add Sp-Sp, Tr-Sp, Tr-Sa validation (869 events)

**Week 3: XB timeline**
1. Load testcase-2 (1,746 events, 1996-2001)
2. Implement date-range matching
3. Validate Tr-Na (1,326 events)

**Week 4: Extended XB timeline**
1. Load testcase-3 (1,512 events, 2001-2006)
2. Validate additional 1,145 Tr-Na events

---

## Current Phase D Status vs Target

| Metric | Current | Target | Gap |
|---|---|---|---|
| **testcase-1 Tr-Na** | 4/188 | 188/188 | 184 events |
| **testcase-1 Sp-Na** | 10/52 | 52/52 | 42 events |
| **testcase-1 Sa-Na** | 11/31 | 31/31 | 20 events |
| **Other pairings** | 0/869 | 869/869 | 869 events |
| **testcase-2** | 0/1746 | 1746/1746 | 1,746 events |
| **testcase-3** | 0/1512 | 1512/1512 | 1,512 events |
| **Total Coverage** | 25/4414 | 4414/4414 | **4,389 events (99.4%)** |

---

## Conclusion

**Phase D snapshot validation is only 0.6% complete.**

The discovery of 4,414 test records (vs. 52 validated) reveals a massive gap in test coverage. This is not a failure — it's an opportunity to comprehensively validate SolarSage against Solar Fire across:

- ✅ All event types (Enter, Exact, Leave, Void, etc.)
- ✅ Full timelines (years, not snapshots)
- ✅ All chart pairings (not just Tr-Na)
- ✅ Both reference persons (JN and XB)

**Next phase will expand Phase D from 0.6% to 100% coverage.**

---

**Last Updated:** 2026-03-31  
**Data Source:** Solar Fire v9.0.29 test exports  
**Analysis:** Automated event discovery and classification
