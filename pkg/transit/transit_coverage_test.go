package transit

import (
	"testing"

	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
)

func TestCalcTransitEvents_WithTransitSpecialPoints(t *testing.T) {
	events, err := CalcTransitEvents(TransitCalcInput{
		NatalChart: NatalChartConfig{
			Lat:     39.9042,
			Lon:     116.4074,
			JD:      natalJD,
			Planets: []models.PlanetID{models.PlanetSun},
			Points:  []models.SpecialPointID{models.PointASC, models.PointMC},
		},
		TimeRange: TimeRangeConfig{
			StartJD: startJD,
			EndJD:   startJD + 10,
		},
		Charts: ChartSetConfig{
			Transit: &TransitChartConfig{
				Lat:         39.9042,
				Lon:         116.4074,
				Planets:     []models.PlanetID{models.PlanetMoon},
				Points:      []models.SpecialPointID{models.PointASC, models.PointMC},
				Orbs:        orbs,
				HouseSystem: models.HousePlacidus,
			},
		},
		EventFilter: EventFilterConfig{
			TrNa: true,
			TrTr: true,
		},
		HouseSystem: models.HousePlacidus,
	})
	if err != nil {
		t.Fatalf("CalcTransitEvents with special points error: %v", err)
	}
	_ = events
}

func TestCalcTransitEvents_WithProgressionsSpecialPoints(t *testing.T) {
	events, err := CalcTransitEvents(TransitCalcInput{
		NatalChart: NatalChartConfig{
			Lat:     39.9042,
			Lon:     116.4074,
			JD:      natalJD,
			Planets: []models.PlanetID{models.PlanetSun},
		},
		TimeRange: TimeRangeConfig{
			StartJD: startJD,
			EndJD:   startJD + 30,
		},
		Charts: ChartSetConfig{
			Transit: &TransitChartConfig{
				Lat:         39.9042,
				Lon:         116.4074,
				Planets:     []models.PlanetID{models.PlanetSun},
				Orbs:        orbs,
				HouseSystem: models.HousePlacidus,
			},
			Progressions: &ProgressionsChartConfig{
				Planets:     []models.PlanetID{models.PlanetSun, models.PlanetMoon},
				Points:      []models.SpecialPointID{models.PointASC, models.PointMC},
				Orbs:        orbs,
				Lat:         39.9042,
				Lon:         116.4074,
				HouseSystem: models.HousePlacidus,
			},
		},
		EventFilter: EventFilterConfig{
			TrSp: true,
			SpNa: true,
		},
		HouseSystem: models.HousePlacidus,
	})
	if err != nil {
		t.Fatalf("CalcTransitEvents with progressions special points error: %v", err)
	}
	_ = events
}

func TestCalcTransitEvents_WithSolarArcSpecialPoints(t *testing.T) {
	events, err := CalcTransitEvents(TransitCalcInput{
		NatalChart: NatalChartConfig{
			Lat:     39.9042,
			Lon:     116.4074,
			JD:      natalJD,
			Planets: []models.PlanetID{models.PlanetSun},
		},
		TimeRange: TimeRangeConfig{
			StartJD: startJD,
			EndJD:   startJD + 30,
		},
		Charts: ChartSetConfig{
			Transit: &TransitChartConfig{
				Lat:         39.9042,
				Lon:         116.4074,
				Planets:     []models.PlanetID{models.PlanetSun},
				Orbs:        orbs,
				HouseSystem: models.HousePlacidus,
			},
			SolarArc: &SolarArcChartConfig{
				Planets:     []models.PlanetID{models.PlanetSun, models.PlanetMoon},
				Points:      []models.SpecialPointID{models.PointASC, models.PointMC},
				Orbs:        orbs,
				Lat:         39.9042,
				Lon:         116.4074,
				HouseSystem: models.HousePlacidus,
			},
		},
		EventFilter: EventFilterConfig{
			SaNa: true,
		},
		HouseSystem: models.HousePlacidus,
	})
	if err != nil {
		t.Fatalf("CalcTransitEvents with solar arc special points error: %v", err)
	}
	_ = events
}

func TestDirectedAngleDiff(t *testing.T) {
	// Test with direct motion (speed > 0)
	got := directedAngleDiff(100, 10, 90, 1.0)
	if got < -180 || got > 180 {
		t.Errorf("directedAngleDiff direct = %f, want in [-180, 180]", got)
	}

	// Test with retrograde motion (speed < 0)
	got2 := directedAngleDiff(100, 10, 90, -0.5)
	if got2 < -180 || got2 > 180 {
		t.Errorf("directedAngleDiff retrograde = %f, want in [-180, 180]", got2)
	}

	// Test conjunction (aspect angle = 0)
	got3 := directedAngleDiff(15, 15, 0, 1.0)
	if got3 != 0 {
		t.Errorf("directedAngleDiff conjunction exact = %f, want 0", got3)
	}
}
