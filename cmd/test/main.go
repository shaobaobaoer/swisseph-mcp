package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anthropic/swisseph-mcp/pkg/chart"
	"github.com/anthropic/swisseph-mcp/pkg/geo"
	"github.com/anthropic/swisseph-mcp/pkg/julian"
	"github.com/anthropic/swisseph-mcp/pkg/models"
	"github.com/anthropic/swisseph-mcp/pkg/progressions"
	"github.com/anthropic/swisseph-mcp/pkg/sweph"
	"github.com/anthropic/swisseph-mcp/pkg/transit"
)

func main() {
	// Initialize ephemeris
	exe, _ := os.Executable()
	ephePath := filepath.Join(filepath.Dir(exe), "..", "..", "third_party", "swisseph", "ephe")
	if _, err := os.Stat(ephePath); err != nil {
		ephePath = filepath.Join(".", "third_party", "swisseph", "ephe")
	}
	sweph.Init(ephePath)
	defer sweph.Close()

	fmt.Println("========================================")
	fmt.Println("  Swisseph MCP Functional Test")
	fmt.Println("========================================")

	// Test 1: Geocode
	fmt.Println("\n--- Test 2.1: Geocode location to coordinates ---")
	loc, err := geo.Geocode("Beijing")
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		fmt.Printf("  Beijing: lat=%.4f, lon=%.4f, tz=%s, name=%s\n",
			loc.Latitude, loc.Longitude, loc.Timezone, loc.DisplayName)
	}

	loc2, err := geo.Geocode("london")
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		fmt.Printf("  London: lat=%.4f, lon=%.4f, tz=%s\n",
			loc2.Latitude, loc2.Longitude, loc2.Timezone)
	}

	// Test 2: DateTime to JD
	fmt.Println("\n--- Test 2.2: Datetime to Julian Day ---")
	jdResult, err := julian.DateTimeToJD("1990-06-15T08:30:00+08:00", models.CalendarGregorian)
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		fmt.Printf("  1990-06-15T08:30:00+08:00:\n")
		fmt.Printf("    JD(UT) = %.6f\n", jdResult.JDUT)
		fmt.Printf("    JD(TT) = %.6f\n", jdResult.JDTT)
	}

	// Test 2b: JD to DateTime
	fmt.Println("\n--- Test 2.2b: Julian Day to Datetime ---")
	dt, err := julian.JDToDateTime(jdResult.JDUT, "Asia/Shanghai")
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		fmt.Printf("  JD %.6f → %s\n", jdResult.JDUT, dt)
	}

	// Test 3: Single Chart
	fmt.Println("\n--- Test 3.1.1: Single Chart Calculation ---")
	planets := []models.PlanetID{
		models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
		models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
		models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune,
		models.PlanetPluto,
	}
	orbs := models.DefaultOrbConfig()

	chartInfo, err := chart.CalcSingleChart(
		39.9042, 116.4074, jdResult.JDUT,
		planets, orbs, models.HousePlacidus,
	)
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		fmt.Printf("  Natal chart (1990-06-15 08:30 Beijing)\n")
		fmt.Printf("  ASC: %.2f° (%s)\n", chartInfo.Angles.ASC, models.SignFromLongitude(chartInfo.Angles.ASC))
		fmt.Printf("  MC:  %.2f° (%s)\n", chartInfo.Angles.MC, models.SignFromLongitude(chartInfo.Angles.MC))
		fmt.Println("  Planet positions:")
		for _, p := range chartInfo.Planets {
			retro := ""
			if p.IsRetrograde {
				retro = " (R)"
			}
			fmt.Printf("    %-10s %6.2f° %s %5.2f°  House%d%s\n",
				p.PlanetID, p.Longitude, p.Sign, p.SignDegree, p.House, retro)
		}
		fmt.Printf("  Aspect count: %d\n", len(chartInfo.Aspects))
		for i, a := range chartInfo.Aspects {
			if i >= 5 {
				fmt.Printf("    ... (total %d aspects)\n", len(chartInfo.Aspects))
				break
			}
			applying := "separating"
			if a.IsApplying {
				applying = "applying"
			}
			fmt.Printf("    %s %s %s (orb %.2f°, %s)\n",
				a.PlanetA, a.AspectType, a.PlanetB, a.Orb, applying)
		}
	}

	// Test 4: Double Chart
	fmt.Println("\n--- Test 3.1.2: Double Chart Calculation ---")
	transitJD, _ := julian.DateTimeToJD("2024-01-01T12:00:00+08:00", models.CalendarGregorian)
	innerChart, outerChart, crossAspects, err := chart.CalcDoubleChart(
		39.9042, 116.4074, jdResult.JDUT, planets,
		39.9042, 116.4074, transitJD.JDUT, planets,
		&models.SpecialPointsConfig{
			InnerPoints: []models.SpecialPointID{models.PointASC, models.PointMC},
		},
		orbs, models.HousePlacidus,
	)
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		fmt.Printf("  Inner planets: %d, Outer planets: %d\n", len(innerChart.Planets), len(outerChart.Planets))
		fmt.Printf("  Cross Aspect count: %d\n", len(crossAspects))
		for i, ca := range crossAspects {
			if i >= 5 {
				fmt.Printf("    ... (total %d cross aspects)\n", len(crossAspects))
				break
			}
			fmt.Printf("    Inner %s %s Outer %s (orb %.2f°)\n",
				ca.InnerBody, ca.AspectType, ca.OuterBody, ca.Orb)
		}
	}

	// Test 5: Progressions engine
	fmt.Println("\n--- Test: Secondary Progressions Engine ---")
	// At age ~34 (2024-01-01), check progressed Sun position
	pLon, pSpeed, err := progressions.CalcProgressedLongitude(models.PlanetSun, jdResult.JDUT, transitJD.JDUT)
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		age := progressions.Age(jdResult.JDUT, transitJD.JDUT)
		fmt.Printf("  Age: %.3f\n", age)
		fmt.Printf("  Progressed Sun: %.4f° %s (speed %.6f°/day)\n",
			pLon, models.SignFromLongitude(pLon), pSpeed)
	}

	// Solar Arc
	fmt.Println("\n--- Test: Solar Arc Direction Engine ---")
	saLon, saSpeed, err := progressions.CalcSolarArcLongitude(models.PlanetMars, jdResult.JDUT, transitJD.JDUT)
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		fmt.Printf("  Solar Arc Mars: %.4f° %s (speed %.6f°/day)\n",
			saLon, models.SignFromLongitude(saLon), saSpeed)
	}

	// Test 6: Transit Tr-Na only (quick, 30 days)
	fmt.Println("\n--- Test 3.2.1: Transit Calculation (Tr-Na, 30day) ---")
	startJD := transitJD.JDUT
	endJD := startJD + 30.0

	transitEvents, err := transit.CalcTransitEvents(transit.TransitCalcInput{
		NatalLat:     39.9042,
		NatalLon:     116.4074,
		NatalJD:      jdResult.JDUT,
		NatalPlanets: planets,
		TransitLat:   39.9042,
		TransitLon:   116.4074,
		StartJD:      startJD,
		EndJD:        endJD,
		TransitPlanets: []models.PlanetID{
			models.PlanetSun, models.PlanetMercury, models.PlanetVenus, models.PlanetMars,
		},
		EventConfig: models.EventConfig{
			IncludeTrNa:         true,
			IncludeSignIngress:  true,
			IncludeHouseIngress: true,
			IncludeStation:      true,
		},
		OrbConfigTransit: orbs,
		HouseSystem:      models.HousePlacidus,
	})
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		fmt.Printf("  Search range: 30 days\n")
		fmt.Printf("  Found %d events\n", len(transitEvents))
		printEvents(transitEvents, 15, jdResult.JDUT)
	}

	// Test 7: Full transit with Tr-Tr (30 days, fast planets)
	fmt.Println("\n--- Test: Tr-Tr Transit Aspects (30day) ---")
	trtrEvents, err := transit.CalcTransitEvents(transit.TransitCalcInput{
		NatalLat:     39.9042,
		NatalLon:     116.4074,
		NatalJD:      jdResult.JDUT,
		NatalPlanets: planets,
		TransitLat:   39.9042,
		TransitLon:   116.4074,
		StartJD:      startJD,
		EndJD:        endJD,
		TransitPlanets: []models.PlanetID{
			models.PlanetSun, models.PlanetMercury, models.PlanetVenus,
		},
		EventConfig: models.EventConfig{
			IncludeTrTr: true,
		},
		OrbConfigTransit: orbs,
		HouseSystem:      models.HousePlacidus,
	})
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		fmt.Printf("  Tr-Tr events: %d\n", len(trtrEvents))
		printEvents(trtrEvents, 10, jdResult.JDUT)
	}

	// Test 8: Sp-Na (progressions vs natal, 1 year)
	fmt.Println("\n--- Test: Sp-Na Secondary Progressions (1 year) ---")
	spnaEvents, err := transit.CalcTransitEvents(transit.TransitCalcInput{
		NatalLat:     39.9042,
		NatalLon:     116.4074,
		NatalJD:      jdResult.JDUT,
		NatalPlanets: planets,
		TransitLat:   39.9042,
		TransitLon:   116.4074,
		StartJD:      startJD,
		EndJD:        startJD + 365.25,
		TransitPlanets: []models.PlanetID{},
		ProgressionsConfig: &models.ProgressionsConfig{
			Enabled: true,
			Planets: []models.PlanetID{
				models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
				models.PlanetVenus, models.PlanetMars,
			},
		},
		EventConfig: models.EventConfig{
			IncludeSpNa:        true,
			IncludeSignIngress: true,
			IncludeStation:     true,
		},
		OrbConfigProgressions: models.OrbConfig{
			Conjunction: 1, Opposition: 1, Trine: 1, Square: 1,
			Sextile: 1, Quincunx: 0.5, SemiSextile: 0.5,
			SemiSquare: 0.5, Sesquiquadrate: 0.5,
		},
		HouseSystem: models.HousePlacidus,
	})
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		fmt.Printf("  Sp-Na events: %d\n", len(spnaEvents))
		printEvents(spnaEvents, 10, jdResult.JDUT)
	}

	// Test 9: Sa-Na (solar arc vs natal, 1 year)
	fmt.Println("\n--- Test: Sa-Na Solar Arc Direction (1 year) ---")
	sanaEvents, err := transit.CalcTransitEvents(transit.TransitCalcInput{
		NatalLat:     39.9042,
		NatalLon:     116.4074,
		NatalJD:      jdResult.JDUT,
		NatalPlanets: planets,
		TransitLat:   39.9042,
		TransitLon:   116.4074,
		StartJD:      startJD,
		EndJD:        startJD + 365.25,
		TransitPlanets: []models.PlanetID{},
		SolarArcConfig: &models.SolarArcConfig{
			Enabled: true,
			Planets: []models.PlanetID{
				models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
				models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
				models.PlanetSaturn,
			},
		},
		EventConfig: models.EventConfig{
			IncludeSaNa:        true,
			IncludeSignIngress: true,
		},
		OrbConfigSolarArc: models.OrbConfig{
			Conjunction: 1, Opposition: 1, Trine: 1, Square: 1,
			Sextile: 1, Quincunx: 0.5, SemiSextile: 0.5,
			SemiSquare: 0.5, Sesquiquadrate: 0.5,
		},
		HouseSystem: models.HousePlacidus,
	})
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		fmt.Printf("  Sa-Na events: %d\n", len(sanaEvents))
		printEvents(sanaEvents, 10, jdResult.JDUT)
	}

	// Test 10: Void of Course Moon (7 days)
	fmt.Println("\n--- Test: Moon Void of Course (7day) ---")
	vocEvents, err := transit.CalcTransitEvents(transit.TransitCalcInput{
		NatalLat:     39.9042,
		NatalLon:     116.4074,
		NatalJD:      jdResult.JDUT,
		NatalPlanets: planets,
		TransitLat:   39.9042,
		TransitLon:   116.4074,
		StartJD:      startJD,
		EndJD:        startJD + 7.0,
		TransitPlanets: []models.PlanetID{
			models.PlanetMoon,
		},
		EventConfig: models.EventConfig{
			IncludeTrNa:         true,
			IncludeSignIngress:  true,
			IncludeVoidOfCourse: true,
		},
		OrbConfigTransit: orbs,
		HouseSystem:      models.HousePlacidus,
	})
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		vocCount := 0
		for _, e := range vocEvents {
			if e.EventType == models.EventVoidOfCourse {
				vocCount++
				startDT, _ := julian.JDToDateTime(e.VoidStartJD, "Asia/Shanghai")
				endDT, _ := julian.JDToDateTime(e.VoidEndJD, "Asia/Shanghai")
				duration := (e.VoidEndJD - e.VoidStartJD) * 24.0
				fmt.Printf("  Void of Course: %s → %s (%.1f hours)\n", startDT, endDT, duration)
				fmt.Printf("    Last aspect: %s → %s, next sign: %s\n",
					e.LastAspectType, e.LastAspectTarget, e.NextSign)
			}
		}
		fmt.Printf("  Moon VOC count in 7 days: %d\n", vocCount)
	}

	// Test 11: Tr-Sp (transit vs progressed, 90 days)
	fmt.Println("\n--- Test: Tr-Sp Transit x Progressions (90day) ---")
	trspEvents, err := transit.CalcTransitEvents(transit.TransitCalcInput{
		NatalLat:     39.9042,
		NatalLon:     116.4074,
		NatalJD:      jdResult.JDUT,
		NatalPlanets: planets,
		TransitLat:   39.9042,
		TransitLon:   116.4074,
		StartJD:      startJD,
		EndJD:        startJD + 90,
		TransitPlanets: []models.PlanetID{
			models.PlanetSun, models.PlanetMars,
		},
		ProgressionsConfig: &models.ProgressionsConfig{
			Enabled: true,
			Planets: []models.PlanetID{
				models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
			},
		},
		EventConfig: models.EventConfig{
			IncludeTrSp: true,
		},
		OrbConfigTransit: models.OrbConfig{
			Conjunction: 2, Opposition: 2, Trine: 2, Square: 2,
			Sextile: 2, Quincunx: 1,
		},
		HouseSystem: models.HousePlacidus,
	})
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		fmt.Printf("  Tr-Sp events: %d\n", len(trspEvents))
		printEvents(trspEvents, 10, jdResult.JDUT)
	}

	// Test 12: Full pipeline (all event types, 30 days)
	fmt.Println("\n--- Test: All Event Types (30day) ---")
	allPlanets := []models.PlanetID{
		models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
		models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
		models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune,
		models.PlanetPluto, models.PlanetChiron,
	}
	fullEvents, err := transit.CalcTransitEvents(transit.TransitCalcInput{
		NatalLat:     39.9042,
		NatalLon:     116.4074,
		NatalJD:      jdResult.JDUT,
		NatalPlanets: allPlanets,
		TransitLat:   39.9042,
		TransitLon:   116.4074,
		StartJD:      startJD,
		EndJD:        startJD + 30,
		TransitPlanets: allPlanets,
		ProgressionsConfig: &models.ProgressionsConfig{
			Enabled: true,
			Planets: []models.PlanetID{models.PlanetSun, models.PlanetMoon},
		},
		SolarArcConfig: &models.SolarArcConfig{
			Enabled: true,
			Planets: []models.PlanetID{models.PlanetSun, models.PlanetMars},
		},
		SpecialPoints: &models.SpecialPointsConfig{
			NatalPoints: []models.SpecialPointID{models.PointASC, models.PointMC},
		},
		EventConfig:           models.DefaultEventConfig(),
		OrbConfigTransit:      orbs,
		OrbConfigProgressions: models.OrbConfig{Conjunction: 1, Opposition: 1, Trine: 1, Square: 1, Sextile: 1},
		OrbConfigSolarArc:     models.OrbConfig{Conjunction: 1, Opposition: 1, Trine: 1, Square: 1, Sextile: 1},
		HouseSystem:           models.HousePlacidus,
	})
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
	} else {
		// Count by event type and chart type
		eventCounts := make(map[string]int)
		chartCounts := make(map[string]int)
		for _, e := range fullEvents {
			eventCounts[string(e.EventType)]++
			if e.EventType == models.EventAspectEnter || e.EventType == models.EventAspectExact || e.EventType == models.EventAspectLeave {
				chartCounts[string(e.ChartType)+"→"+string(e.TargetChartType)]++
			}
		}
		fmt.Printf("  Total events: %d\n", len(fullEvents))
		fmt.Printf("  Event type stats:\n")
		for t, c := range eventCounts {
			fmt.Printf("    %-20s %d\n", t, c)
		}
		fmt.Printf("  Chart type stats:\n")
		for t, c := range chartCounts {
			fmt.Printf("    %-30s %d\n", t, c)
		}
	}

	// Sample JSON output
	if len(transitEvents) > 0 {
		fmt.Println("\n--- Sample JSON output ---")
		sample := transitEvents[0]
		j, _ := json.MarshalIndent(sample, "  ", "  ")
		fmt.Printf("  %s\n", string(j))
	}

	fmt.Println("\n========================================")
	fmt.Println("  All tests complete!")
	fmt.Println("========================================")
}

func printEvents(events []models.TransitEvent, max int, natalJD float64) {
	shown := 0
	for _, e := range events {
		if shown >= max {
			fmt.Printf("    ... (total %d events)\n", len(events))
			break
		}
		retro := ""
		if e.IsRetrograde {
			retro = "(R)"
		}
		dtStr, _ := julian.JDToDateTime(e.JD, "Asia/Shanghai")
		ageStr := fmt.Sprintf("[%.3f]", e.Age)

		switch e.EventType {
		case models.EventAspectEnter:
			fmt.Printf("    %s %s %s %s enters %s(%s) %s (orb %.2f°) %s %s\n",
				dtStr, ageStr, e.ChartType, e.Planet, e.Target, e.TargetChartType, e.AspectType, e.OrbAtEnter, e.PlanetSign, retro)
		case models.EventAspectExact:
			fmt.Printf("    %s %s %s %s exact %s(%s) %s (#%d hit) %s %s\n",
				dtStr, ageStr, e.ChartType, e.Planet, e.Target, e.TargetChartType, e.AspectType, e.ExactCount, e.PlanetSign, retro)
		case models.EventAspectLeave:
			fmt.Printf("    %s %s %s %s leaves %s(%s) %s (orb %.2f°) %s %s\n",
				dtStr, ageStr, e.ChartType, e.Planet, e.Target, e.TargetChartType, e.AspectType, e.OrbAtLeave, e.PlanetSign, retro)
		case models.EventSignIngress:
			fmt.Printf("    %s %s %s %s sign change %s → %s %s\n",
				dtStr, ageStr, e.ChartType, e.Planet, e.FromSign, e.ToSign, retro)
		case models.EventHouseIngress:
			fmt.Printf("    %s %s %s %s house change house %d → house %d %s\n",
				dtStr, ageStr, e.ChartType, e.Planet, e.FromHouse, e.ToHouse, retro)
		case models.EventStation:
			fmt.Printf("    %s %s %s %s station %s %s %s\n",
				dtStr, ageStr, e.ChartType, e.Planet, e.StationType, e.PlanetSign, retro)
		case models.EventVoidOfCourse:
			startDT, _ := julian.JDToDateTime(e.VoidStartJD, "Asia/Shanghai")
			endDT, _ := julian.JDToDateTime(e.VoidEndJD, "Asia/Shanghai")
			fmt.Printf("    Moon Void of Course: %s → %s\n", startDT, endDT)
		}
		shown++
	}
}
