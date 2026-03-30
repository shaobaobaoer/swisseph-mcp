package transit

import (
	"testing"

	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
)

func BenchmarkCalcTransitEvents_30Days(b *testing.B) {
	planets := []models.PlanetID{
		models.PlanetSun, models.PlanetMoon, models.PlanetMercury,
		models.PlanetVenus, models.PlanetMars,
	}
	input := TransitCalcInput{
		NatalChart: NatalChartConfig{
			Lat:     51.5074,
			Lon:     -0.1278,
			JD:      2451545.0,
			Planets: planets,
		},
		TimeRange: TimeRangeConfig{
			StartJD: 2460676.5,
			EndJD:   2460706.5, // 30 days
		},
		Charts: ChartSetConfig{
			Transit: &TransitChartConfig{
				Lat:         51.5074,
				Lon:         -0.1278,
				Planets:     planets,
				Orbs:        models.DefaultOrbConfig(),
				HouseSystem: models.HousePlacidus,
			},
		},
		EventFilter: EventFilterConfig{
			TrNa: true,
		},
		HouseSystem: models.HousePlacidus,
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
	customOrbs := models.OrbConfig{
		Conjunction:    1,
		Opposition:     1,
		Trine:          1,
		Square:         1,
		Sextile:        1,
		Quincunx:       1,
		SemiSquare:     1,
		Sesquiquadrate: 1,
	}
	input := TransitCalcInput{
		NatalChart: NatalChartConfig{
			Lat:     51.5074,
			Lon:     -0.1278,
			JD:      2451545.0,
			Planets: natalPlanets,
		},
		TimeRange: TimeRangeConfig{
			StartJD: 2460676.5,
			EndJD:   2460676.5 + 365.25, // 1 year
		},
		Charts: ChartSetConfig{
			Transit: &TransitChartConfig{
				Lat:         51.5074,
				Lon:         -0.1278,
				Planets:     transitPlanets,
				Orbs:        customOrbs,
				HouseSystem: models.HousePlacidus,
			},
		},
		EventFilter: EventFilterConfig{
			TrNa:    true,
			Station: true,
		},
		HouseSystem: models.HousePlacidus,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalcTransitEvents(input)
	}
}
