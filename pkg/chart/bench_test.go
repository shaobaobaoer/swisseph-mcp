package chart

import (
	"testing"

	"github.com/anthropic/swisseph-mcp/pkg/models"
)

var defaultPlanets = []models.PlanetID{
	models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
	models.PlanetVenus, models.PlanetMars, models.PlanetJupiter,
	models.PlanetSaturn, models.PlanetUranus, models.PlanetNeptune,
	models.PlanetPluto,
}

func BenchmarkCalcSingleChart(b *testing.B) {
	orbs := models.DefaultOrbConfig()
	for i := 0; i < b.N; i++ {
		CalcSingleChart(51.5074, -0.1278, 2451545.0, defaultPlanets, orbs, models.HousePlacidus)
	}
}

func BenchmarkCalcPlanetLongitude(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CalcPlanetLongitude(models.PlanetSun, 2451545.0)
	}
}

func BenchmarkCalcDoubleChart(b *testing.B) {
	orbs := models.DefaultOrbConfig()
	for i := 0; i < b.N; i++ {
		CalcDoubleChart(
			51.5074, -0.1278, 2451545.0, defaultPlanets,
			51.5074, -0.1278, 2451545.0+365.25, defaultPlanets,
			nil, orbs, models.HousePlacidus,
		)
	}
}
