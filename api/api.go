package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/jemgunay/canlendar-graph/calendar"
	"github.com/jemgunay/canlendar-graph/storage"
)

// API defines the HTTP handlers.
type API struct {
	storer     storage.Storer
	calFetcher calendar.Fetcher
}

// New initialises an API.
func New(storer storage.Storer, calFetcher calendar.Fetcher) *API {
	return &API{
		storer:     storer,
		calFetcher: calFetcher,
	}
}

type queryResponse struct {
	Plots    []storage.Plot    `json:"plots"`
	Metadata queryResponseMeta `json:"metadata"`
}

type queryResponseMeta struct {
	Guideline float64 `json:"guideline,omitempty"`
}

// Query collects calendar data from storage and returns it as plottable data points.
func (a *API) Query(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := r.URL.Query()
	aggregation := storage.Aggregation(query.Get("aggregation"))
	startTime, _ := time.Parse(time.RFC3339, query.Get("start_time"))
	endTime, _ := time.Parse(time.RFC3339, query.Get("end_time"))

	opts := []storage.QueryOption{
		storage.WithAggregation(aggregation),
		storage.WithStartTime(startTime),
		storage.WithEndTime(endTime),
	}

	records, err := a.storer.Query(ctx, opts...)
	if err != nil {
		log.Printf("failed to query storage: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// default 0 units for Day aggregation - only show a guideline for Week aggregation and higher
	var guidelineUnits float64
	switch aggregation {
	case storage.Week:
		guidelineUnits = calendar.MaxRecommendedWeeklyUnits
	case storage.Month:
		guidelineUnits = calendar.MaxRecommendedWeeklyUnits * 4
	case storage.Year:
		guidelineUnits = calendar.MaxRecommendedWeeklyUnits * 4 * 12
	}

	resp := queryResponse{
		Metadata: queryResponseMeta{
			Guideline: guidelineUnits,
		},
		Plots: records,
	}

	// JSON encode response
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(resp); err != nil {
		log.Printf("failed to JSON encode response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type collectPayload struct {
	StartTime time.Time `json:"start_time_override"`
}

// Collect scrapes the Google calendar API for new events (i.e. those created since the last scraped event) and writes
// them to storage.
func (a *API) Collect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// get start time override from body
	payload := collectPayload{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("unable to decode request body: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// if no override, get the timestamp for the last written storage record. If there are no records in storage then
	// the start of time will be used
	if payload.StartTime.IsZero() {
		var err error
		if payload.StartTime, err = a.getLastTimestamp(ctx); err != nil {
			log.Printf("failed to read last written timestamp from storage: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// add 24h to ensure we don't recollect the last event
		payload.StartTime = payload.StartTime.Add(time.Hour * 24)
	}

	// fetch calendar events for time range
	eventIter, err := a.calFetcher.Fetch(ctx, payload.StartTime)
	if err != nil {
		if err == calendar.ErrNoEventsFound {
			log.Printf("no new events found since %s", payload.StartTime.Format(time.RFC3339))
			w.WriteHeader(http.StatusNoContent)
			return
		}
		log.Printf("failed to fetch calendar events: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// process calendar events into
	records := make([]storage.Record, 0, eventIter.Count())
	for {
		ev, err := eventIter.Next()
		if err != nil {
			if errors.Is(err, calendar.ErrNoMoreEvents) {
				break
			}
			log.Printf("failed to read event: %s", err)
			continue
		}

		records = append(records, storage.Record{
			Time: ev.Date,
			Fields: map[string]interface{}{
				"units": ev.Units,
			},
		})
	}

	// persist new events to storage
	if err := a.storer.Store(ctx, records...); err != nil {
		log.Printf("failed to persist events to storage: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// getLastTimestamp reads the last written unit timestamp from storage. If there were no results in storage, then the
// value representing the start of time is returned.
func (a *API) getLastTimestamp(ctx context.Context) (time.Time, error) {
	startTime, err := a.storer.ReadLastTimestamp(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrNoResults) {
			// default to the start of time
			return time.Time{}, nil
		}

		return time.Time{}, err
	}

	return startTime, nil
}
