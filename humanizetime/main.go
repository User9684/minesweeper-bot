package humanizetime

import (
	"fmt"
	"math"
	"time"
)

var DaysPerMonth = 30.4167

type Unit struct {
	UnitString string
	Seconds    float64
}

var daysPerMonth = 30.4167

var timeOrder = []Unit{
	{
		UnitString: "decades",
		Seconds:    60 * 60 * 24 * daysPerMonth * 12 * 10,
	},
	{
		UnitString: "years",
		Seconds:    60 * 60 * 24 * daysPerMonth * 12,
	},
	{
		UnitString: "months",
		Seconds:    60 * 60 * 24 * daysPerMonth,
	},
	{
		UnitString: "days",
		Seconds:    60 * 60 * 24,
	},
	{
		UnitString: "hours",
		Seconds:    60 * 60,
	},
	{
		UnitString: "minutes",
		Seconds:    60,
	},
}

func HumanizeDuration(t time.Duration, secondPrecision int) string {
	times := make(map[string]float64)

	seconds := t.Seconds()

	var timestring string

	for _, unit := range timeOrder {
		value := math.Floor(seconds / unit.Seconds)
		if value <= 0 {
			continue
		}

		times[unit.UnitString] = value
		seconds -= times[unit.UnitString] * unit.Seconds

		timestring = fmt.Sprintf("%s, %v %s", timestring, value, unit.UnitString)
	}

	if len(times) >= 1 {
		timestring = fmt.Sprintf("%s and ", timestring)[2:]
	}

	timestring = fmt.Sprintf("%s%v seconds", timestring, round(seconds, secondPrecision))

	return timestring
}

func round(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
