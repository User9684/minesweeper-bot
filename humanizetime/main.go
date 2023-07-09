package humanizetime

import (
	"fmt"
	"math"
	"time"
)

type Unit struct {
	UnitString string
	Seconds    float64
}

var daysPerMonth = 30.4167

var timeOrder = []Unit{
	{
		UnitString: "decade",
		Seconds:    60 * 60 * 24 * daysPerMonth * 12 * 10,
	},
	{
		UnitString: "year",
		Seconds:    60 * 60 * 24 * daysPerMonth * 12,
	},
	{
		UnitString: "month",
		Seconds:    60 * 60 * 24 * daysPerMonth,
	},
	{
		UnitString: "day",
		Seconds:    60 * 60 * 24,
	},
	{
		UnitString: "hour",
		Seconds:    60 * 60,
	},
	{
		UnitString: "minute",
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

		unitString := unit.UnitString
		if value >= 2 {
			unitString = fmt.Sprintf("%ss", unitString)
		}

		times[unitString] = value
		seconds -= times[unitString] * unit.Seconds

		timestring = fmt.Sprintf("%s, %.0f %s", timestring, value, unitString)
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
