package calendar

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	gcal "google.golang.org/api/calendar/v3"
)

type Results struct {
	Events []*gcal.Event
}

type scale int

const (
	Month scale = iota
	Week
	Day
)

func (r *Results) GetPlots(scale scale) {
	// group Events by week and sum each set of units
	var (
		firstEventDate, lastEventDate time.Time
		eventsByWeek                  = make(map[int64]float64)
	)

	// determine number of plots points required based on scale and first/last dates
	for i, event := range r.Events {
		date, units, err := processEvent(event)
		switch err {
		case ErrUnspecifiedUnitCount:
			units = 14 // TODO: fix this
		default:
			log.Printf("failed to parse event date for index %d", i)
			continue
		}

		// truncate date to scale
		var truncDate time.Time
		switch scale {
		case Month:
			truncDate = time.Date(date.Year(), date.Month(), 0, 0, 0, 0, 0, date.Location())
		case Week:
			truncDate = date.Truncate(time.Hour * 24 * 7)
		case Day:
			truncDate = date.Truncate(time.Hour * 24)
		}

		// store first and last events to calculate total required plot count
		if i == 0 {
			firstEventDate = truncDate
		} else if i == len(r.Events)-1 {
			lastEventDate = truncDate
		}

		epoch := truncDate.Unix() * 1000
		eventsByWeek[epoch] = eventsByWeek[epoch] + units
	}

	// determine number of required plot points to generate
	var numPlots int
	switch scale {
	case Month:
		numPlots = int(lastEventDate.Sub(firstEventDate).Hours()/scale.Hours()) + 1
	case Week:
		numPlots = int(lastEventDate.Sub(firstEventDate).Hours()/24/7) + 1
	case Day:
		numPlots = int(lastEventDate.Sub(firstEventDate).Hours()/24) + 1
	}

	// for each required plot point, determine the corresponding date and units from previously processed events. Plot
	// points without a corresponding event will be defaulted to 0 units
	plots := make([]plot, 0, numPlots)
	for i := 0; i < numPlots; i++ {
		nextWeek := firstEventDate.Add(scale * time.Duration(i))
		epoch := nextWeek.Unix() * 1000

		plots = append(plots, plot{
			X: epoch,
			Y: eventsByWeek[epoch],
		})
	}
}

func diff(a, b time.Time) (year, month, day, hour, min, sec int) {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	year = int(y2 - y1)
	month = int(M2 - M1)
	day = int(d2 - d1)
	hour = int(h2 - h1)
	min = int(m2 - m1)
	sec = int(s2 - s1)

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}
	return
}

var (
	re                      = regexp.MustCompile(`(?m)[\d?]*\.?\d*`)
	ErrUnspecifiedUnitCount = errors.New("unknown units documented")
)

func processEvent(event *gcal.Event) (time.Time, float64, error) {
	date := event.Start.DateTime
	if date == "" {
		date = event.Start.Date
	}

	// parse date from string
	parsedTime, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, 0, err
	}

	// parse summary into units
	var units float64
	switch match := re.FindString(event.Summary); match {
	case "":
		return time.Time{}, 0, fmt.Errorf("failed to find a units number for event: %s", event.Summary)

	case "?":
		return time.Time{}, 0, ErrUnspecifiedUnitCount

	default:
		units, err = strconv.ParseFloat(match, 64)
		if err != nil {
			return time.Time{}, 0, fmt.Errorf("failed to parse units number from event %s", event.Summary)
		}
	}

	return parsedTime.UTC(), units, nil
}
