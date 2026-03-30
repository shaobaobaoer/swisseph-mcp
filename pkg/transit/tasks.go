package transit

import (
	"github.com/shaobaobaoer/solarsage-mcp/pkg/models"
)

// Task is the interface for all transit calculation tasks.
type Task interface {
	Run(ctx *CalcContext) []models.TransitEvent
}

// StationTask calculates station events for a moving body.
type StationTask struct {
	Body MovingBody
}

// Run executes the station task.
func (t *StationTask) Run(ctx *CalcContext) []models.TransitEvent {
	var events []models.TransitEvent

	// Only check stations if body can retrograde
	if !t.Body.CanRetrograde {
		return events
	}

	stations := findStations(t.Body.CalcFn, ctx.StartJD, ctx.EndJD, models.PlanetID(t.Body.ID))
	for _, st := range stations {
		events = append(events, makeStationEvent(t.Body.CalcFn, models.PlanetID(t.Body.ID), t.Body.ChartType, st, ctx.NatalHouses, ctx.NatalJD))
	}

	return events
}

// SignIngressTask calculates sign ingress events for a moving body.
type SignIngressTask struct {
	Body MovingBody
}

// Run executes the sign ingress task.
func (t *SignIngressTask) Run(ctx *CalcContext) []models.TransitEvent {
	stations := findStations(t.Body.CalcFn, ctx.StartJD, ctx.EndJD, models.PlanetID(t.Body.ID))
	intervals := buildMonoIntervals(ctx.StartJD, ctx.EndJD, stations)

	return findSignIngressEvents(t.Body.CalcFn, models.PlanetID(t.Body.ID), t.Body.ChartType, intervals, ctx.NatalHouses, ctx.NatalJD)
}

// HouseIngressTask calculates house ingress events for a moving body.
type HouseIngressTask struct {
	Body MovingBody
}

// Run executes the house ingress task.
func (t *HouseIngressTask) Run(ctx *CalcContext) []models.TransitEvent {
	stations := findStations(t.Body.CalcFn, ctx.StartJD, ctx.EndJD, models.PlanetID(t.Body.ID))
	intervals := buildMonoIntervals(ctx.StartJD, ctx.EndJD, stations)

	return findHouseIngressEvents(t.Body.CalcFn, models.PlanetID(t.Body.ID), t.Body.ChartType, intervals, ctx.NatalHouses, ctx.NatalJD)
}

// AspectRQ1Task calculates aspect events between a moving body and fixed natal references.
type AspectRQ1Task struct {
	Body      MovingBody
	NatalRefs []NatalRef
}

// Run executes the RQ1 aspect task.
func (t *AspectRQ1Task) Run(ctx *CalcContext) []models.TransitEvent {
	var allEvents []models.TransitEvent

	stations := findStations(t.Body.CalcFn, ctx.StartJD, ctx.EndJD, models.PlanetID(t.Body.ID))
	intervals := buildMonoIntervals(ctx.StartJD, ctx.EndJD, stations)

	exactCounters := make(map[string]int)

	for _, ref := range t.NatalRefs {
		events := findAspectEventsRQ1(
			t.Body.CalcFn, models.PlanetID(t.Body.ID), t.Body.ChartType,
			ref.ID, ref.Longitude, ref.ChartType,
			intervals, t.Body.Orbs, ctx.NatalHouses, exactCounters,
			ctx.NatalJD,
		)
		allEvents = append(allEvents, events...)
	}

	return allEvents
}

// AspectRQ2Task calculates aspect events between two moving bodies.
type AspectRQ2Task struct {
	Body1 MovingBody
	Body2 MovingBody
}

// Run executes the RQ2 aspect task.
func (t *AspectRQ2Task) Run(ctx *CalcContext) []models.TransitEvent {
	return findAspectEventsRQ2(
		t.Body1.CalcFn, models.PlanetID(t.Body1.ID), t.Body1.ChartType,
		t.Body2.CalcFn, models.PlanetID(t.Body2.ID), t.Body2.ChartType,
		ctx.StartJD, ctx.EndJD,
		t.Body1.Orbs, ctx.NatalHouses, ctx.NatalJD,
	)
}

// buildTasks generates all calculation tasks based on the input configuration.
func buildTasks(ctx *CalcContext) []Task {
	var tasks []Task
	input := ctx.Input

	// Build moving bodies for each chart type
	transitBodies := buildTransitBodies(input.Charts.Transit)
	progressBodies := buildProgressionBodies(input.Charts.Progressions, input.NatalChart.JD, input.NatalChart.MCOverride, input.NatalChart.MCOverrideForASC, input.NatalChart.ASCOverrideForProgressions)
	solarArcBodies := buildSolarArcBodies(input.Charts.SolarArc, input.NatalChart)

	// Process Transit bodies
	for _, trBody := range transitBodies {
		// Station events
		if input.EventFilter.Station && trBody.CanRetrograde {
			tasks = append(tasks, &StationTask{Body: trBody})
		}

		// Sign ingress
		if input.EventFilter.SignIngress {
			tasks = append(tasks, &SignIngressTask{Body: trBody})
		}

		// House ingress
		if input.EventFilter.HouseIngress {
			tasks = append(tasks, &HouseIngressTask{Body: trBody})
		}

		// TR-NA: Transit → Natal aspects (RQ1)
		if input.EventFilter.TrNa {
			tasks = append(tasks, &AspectRQ1Task{
				Body:      trBody,
				NatalRefs: ctx.NatalRefs,
			})
		}

		// TR-TR: Transit → Transit aspects (RQ2)
		if input.EventFilter.TrTr {
			for _, trBody2 := range transitBodies {
				if shouldPairRQ2(trBody.ID, trBody2.ID) {
					tasks = append(tasks, &AspectRQ2Task{
						Body1: trBody,
						Body2: trBody2,
					})
				}
			}
		}

		// TR-SP: Transit → Progressions aspects (RQ2)
		if input.EventFilter.TrSp {
			for _, spBody := range progressBodies {
				tasks = append(tasks, &AspectRQ2Task{
					Body1: trBody,
					Body2: spBody,
				})
			}
		}

		// TR-SA: Transit → SolarArc aspects (RQ2)
		if input.EventFilter.TrSa {
			for _, saBody := range solarArcBodies {
				tasks = append(tasks, &AspectRQ2Task{
					Body1: trBody,
					Body2: saBody,
				})
			}
		}

		// TR-SR: Transit → SolarReturn aspects (RQ1)
		if input.EventFilter.TrSr && len(ctx.SRRefs) > 0 {
			tasks = append(tasks, &AspectRQ1Task{
				Body:      trBody,
				NatalRefs: ctx.SRRefs,
			})
		}
	}

	// Process Progressions bodies
	for _, spBody := range progressBodies {
		// Station events
		if input.EventFilter.Station {
			tasks = append(tasks, &StationTask{Body: spBody})
		}

		// Sign ingress
		if input.EventFilter.SignIngress {
			tasks = append(tasks, &SignIngressTask{Body: spBody})
		}

		// House ingress
		if input.EventFilter.HouseIngress {
			tasks = append(tasks, &HouseIngressTask{Body: spBody})
		}

		// SP-NA: Progressions → Natal aspects (RQ1)
		if input.EventFilter.SpNa {
			tasks = append(tasks, &AspectRQ1Task{
				Body:      spBody,
				NatalRefs: ctx.NatalRefs,
			})
		}

		// SP-SP: Progressions → Progressions aspects (RQ2)
		if input.EventFilter.SpSp {
			for _, spBody2 := range progressBodies {
				if shouldPairRQ2(spBody.ID, spBody2.ID) {
					tasks = append(tasks, &AspectRQ2Task{
						Body1: spBody,
						Body2: spBody2,
					})
				}
			}
		}

		// SP-SR: Progressions → SolarReturn aspects (RQ1)
		if input.EventFilter.SpSr && len(ctx.SRRefs) > 0 {
			tasks = append(tasks, &AspectRQ1Task{
				Body:      spBody,
				NatalRefs: ctx.SRRefs,
			})
		}
	}

	// Process SolarArc bodies
	for _, saBody := range solarArcBodies {
		// Solar arc bodies never retrograde, so no station events

		// Sign ingress
		if input.EventFilter.SignIngress {
			tasks = append(tasks, &SignIngressTask{Body: saBody})
		}

		// House ingress
		if input.EventFilter.HouseIngress {
			tasks = append(tasks, &HouseIngressTask{Body: saBody})
		}

		// SA-NA: SolarArc → Natal aspects (RQ1)
		if input.EventFilter.SaNa {
			tasks = append(tasks, &AspectRQ1Task{
				Body:      saBody,
				NatalRefs: ctx.NatalRefs,
			})
		}

		// SA-SR: SolarArc → SolarReturn aspects (RQ1)
		if input.EventFilter.SaSr && len(ctx.SRRefs) > 0 {
			tasks = append(tasks, &AspectRQ1Task{
				Body:      saBody,
				NatalRefs: ctx.SRRefs,
			})
		}
	}

	return tasks
}

// planetSFOrder defines the canonical SF ordering for Tr-Tr P1/P2 assignment.
// Lower index = slower planet = P1 in Solar Fire output.
var planetSFOrder = map[string]int{
	string(models.PlanetJupiter):       1,
	string(models.PlanetSaturn):        2,
	string(models.PlanetUranus):        3,
	string(models.PlanetNeptune):       4,
	string(models.PlanetPluto):         5,
	string(models.PlanetChiron):        6,
	string(models.PlanetNorthNodeMean): 7,
	string(models.PlanetMoon):          8,
	string(models.PlanetSun):           9,
	string(models.PlanetMercury):       10,
	string(models.PlanetVenus):         11,
	string(models.PlanetMars):          12,
}

// sfOrder returns the SF order index for a body ID (lower = P1).
func sfOrder(id string) int {
	if o, ok := planetSFOrder[id]; ok {
		return o
	}
	return 99
}

// shouldPairRQ2 returns true if body1 should be P1 (comes first in SF order),
// avoiding duplicate pairs.
func shouldPairRQ2(id1, id2 string) bool {
	o1, o2 := sfOrder(id1), sfOrder(id2)
	if o1 != o2 {
		return o1 < o2
	}
	return id1 < id2
}

// runAll executes all tasks and returns combined events.
func runAll(tasks []Task, ctx *CalcContext) []models.TransitEvent {
	var allEvents []models.TransitEvent

	for _, task := range tasks {
		events := task.Run(ctx)
		allEvents = append(allEvents, events...)
	}

	return allEvents
}
