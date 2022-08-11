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

// graphResponse represents the graph plots and metadata returned from the data API.
type graphResponse struct {
	Plots    []Plot            `json:"plots"`
	Metadata map[string]string `json:"metadata"`
}

type queryPayload struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// Plot is a point on a graph.
type Plot struct {
	X int64   `json:"t"`
	Y float64 `json:"y"`
}

func (a *API) Query(w http.ResponseWriter, r *http.Request) {
	/*ctx := r.Context()

	// TODO: read query details from body, i.e. start, end time
	payload := queryPayload{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("unable to decode request body: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// extract URL/query string params
	view := mux.Vars(r)["view"]
	if view == "" {
		log.Println("no view string provided in URL route")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	scale, err := calendar.ValidateScale(view)
	if err != nil {
		log.Printf("failed to parse scale: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// fetch calendar events
	events, err := a.calFetcher.Fetch(ctx, time.Time{}) // TODO: plug start time in here
	if err != nil {
		log.Printf("failed to fetch units for \"%s\": %s", scale, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	guidelineUnits := a.weeklyTarget
	if scale == calendar.Month {
		guidelineUnits *= 4
	}

	// construct response
	resp := graphResponse{
		Metadata: map[string]string{
			"guideline": strconv.Itoa(guidelineUnits),
		},
		//Plots: events.GeneratePlots(scale, float64(a.weeklyTarget)),
	}

	// JSON encode response
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(resp); err != nil {
		log.Printf("failed to JSON encode response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}*/
}

type collectPayload struct {
	StartTime time.Time `json:"start_time_override"`
}

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
