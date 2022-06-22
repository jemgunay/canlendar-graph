package calendar

import (
	"errors"
	"fmt"
	"log"

	"golang.org/x/net/context"
	gcal "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Requester struct {
	calendarID string
	service    *gcal.Service
}

func New(calendarName string, isLocal bool) (*Requester, error) {
	options := []option.ClientOption{
		option.WithScopes(gcal.CalendarReadonlyScope),
	}

	// if running locally, read credentials from file. Otherwise use env defaults
	if isLocal {
		log.Println("reading auth config from credentials.json file")
		options = append(options, option.WithCredentialsFile("credentials.json"))
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
		calendarID: calendarID,
		service:    service,
	}, nil
}

// Fetch fetches a set of events for a given calendar name.
func (r *Requester) Fetch() (*Results, error) {
	// request all Events for target calendar
	req := r.service.Events.List(r.calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		OrderBy("startTime")

	events, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events: %w", err)
	}

	if len(events.Items) == 0 {
		return nil, errors.New("no events found")
	}

	return &Results{
		Events: events.Items,
	}, nil
}
