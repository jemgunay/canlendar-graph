package calendar

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/net/context"
	gcal "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Fetcher interface {
	Fetch(ctx context.Context, startTime time.Time) (EventIterator, error)
}

var _ Fetcher = &Requester{}

type Requester struct {
	calendarID          string
	service             *gcal.Service
	unknownDefaultUnits float64
}

func New(calendarName string, isLocal bool, unknownDefaultUnits float64) (*Requester, error) {
	options := []option.ClientOption{
		option.WithScopes(gcal.CalendarReadonlyScope),
	}

	// if running locally, read credentials from file. Otherwise use env defaults
	if isLocal {
		log.Println("reading auth config from credentials.json file")
		options = append(options, option.WithCredentialsFile("config/credentials.json"))
	}

	service, err := gcal.NewService(context.Background(), options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Calendar API client: %w", err)
	}

	// get a list of calendars
	list, err := service.CalendarList.List().Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve calendar list: %w", err)
	}

	// iterate over all calendars and locate the corresponding ID for the target calendar name
	var calendarID string
	for _, item := range list.Items {
		if item.Summary == calendarName {
			calendarID = item.Id
			break
		}
	}

	// validate that calendar ID was found for target calendar
	if calendarID == "" {
		return nil, fmt.Errorf("failed to find ID for the '%s' calendar", calendarName)
	}

	return &Requester{
		calendarID:          calendarID,
		service:             service,
		unknownDefaultUnits: unknownDefaultUnits,
	}, nil
}

type Event struct {
	Date         time.Time
	Units        float64
	unitsUnknown bool
}

var ErrNoEventsFound = errors.New("no events found")

// Fetch fetches a set of events for a given calendar name.
func (r *Requester) Fetch(ctx context.Context, startTime time.Time) (EventIterator, error) {
	// request all Events for target calendar
	req := r.service.Events.List(r.calendarID).
		TimeMin(startTime.Format(time.RFC3339)).
		ShowDeleted(false).
		SingleEvents(true).
		OrderBy("startTime").
		Context(ctx)

	events, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events: %w", err)
	}

	if len(events.Items) == 0 {
		return nil, ErrNoEventsFound
	}

	return &Iterator{
		events:              events,
		unknownDefaultUnits: r.unknownDefaultUnits,
	}, nil
}

func (r *Requester) GetDefaultUnits() float64 {
	return r.unknownDefaultUnits
}

type EventIterator interface {
	Next() (Event, error)
	Count() int
}

type Iterator struct {
	events              *gcal.Events
	current             int
	unknownDefaultUnits float64
}

var ErrNoMoreEvents = errors.New("no more events to iterate over")

func (i *Iterator) Next() (Event, error) {
	if i.current == len(i.events.Items) {
		return Event{}, ErrNoMoreEvents
	}
	ev, err := processEvent(i.events.Items[i.current])
	if err != nil {
		return ev, err
	}

	if ev.unitsUnknown {
		ev.Units = i.unknownDefaultUnits
	}

	i.current++
	return ev, nil
}

func (i *Iterator) Count() int {
	return len(i.events.Items)
}

var summaryUnitsRegex = regexp.MustCompile(`(?m)[\d?]*\.?\d*`)

// processEvent processes the date and number of units from the calendar event summary. Returns a units count of -1 if
// the unit amount was specified as unknown in the event summary, i.e. "?" instead of a number.
func processEvent(event *gcal.Event) (Event, error) {
	ev := Event{}

	var err error
	// parse date from string
	switch {
	case event.Start.DateTime != "":
		ev.Date, err = time.Parse(time.RFC3339, event.Start.DateTime)
	case event.Start.Date != "":
		ev.Date, err = time.Parse("2006-01-02", event.Start.Date)
	default:
		return ev, errors.New("no valid date found on event")
	}
	if err != nil {
		return ev, err
	}

	// parse summary into units
	switch match := summaryUnitsRegex.FindString(event.Summary); match {
	case "":
		// invalid summary provided
		return ev, fmt.Errorf("failed to find a units number for event: %s", event.Summary)
	case "?":
		// indicates that the consumed number of units was unknown
		ev.unitsUnknown = true
	default:
		// parse number of units into float
		if ev.Units, err = strconv.ParseFloat(match, 64); err != nil {
			return ev, fmt.Errorf("failed to parse units number from event %s", event.Summary)
		}
	}

	return ev, nil
}
