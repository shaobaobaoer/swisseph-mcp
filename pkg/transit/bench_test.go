package transit

import (
	"testing"

	"github.com/anthropic/swisseph-mcp/pkg/models"
)

func BenchmarkCalcTransitEvents_30Days(b *testing.B) {
	planets := []models.PlanetID{
		models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
		models.PlanetVenus, models.PlanetMars,
	}
	input := TransitCalcInput{
		NatalLat: 51.5074, NatalLon: -0.1278,
		NatalJD: 2451545.0, NatalPlanets: planets,
		TransitLat: 51.5074, TransitLon: -0.1278,
		StartJD: 2460676.5, EndJD: 2460706.5, // 30 days
		TransitPlanets:   planets,
		EventConfig:      models.DefaultEventConfig(),
		OrbConfigTransit: models.DefaultOrbConfig(),
		HouseSystem:      models.HousePlacidus,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalcTransitEvents(input)
	}
}

func BenchmarkCalcTransitEvents_1Year_OuterPlanets(b *testing.B) {
	natalPlanets := []models.PlanetID{
		models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
		models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
		models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune,
		models.PlanetPluto,
	}
	transitPlanets := []models.PlanetID{
		models.PlanetJupiter, models.PlanetSaturn, models.PlanetUranus,
		models.PlanetNeptune, models.PlanetPluto,
	}
	input := TransitCalcInput{
		NatalLat: 51.5074, NatalLon: -0.1278,
		NatalJD: 2451545.0, NatalPlanets: natalPlanets,
		TransitLat: 51.5074, TransitLon: -0.1278,
		StartJD: 2460676.5, EndJD: 2460676.5 + 365.25, // 1 year
		TransitPlanets: transitPlanets,
		EventConfig: models.EventConfig{
			IncludeTrNa:    true,
			IncludeStation: true,
		},
		OrbConfigTransit: models.OrbConfig{
			Conjunction: 1, Opposition: 1, Trine: 1, Square: 1,
			Sextile: 1, Quincunx: 1, SemiSquare: 1, Sesquiquadrate: 1,
		},
		HouseSystem: models.HousePlacidus,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalcTransitEvents(input)
	}
}
