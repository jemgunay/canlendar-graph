package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jemgunay/canlendar-graph/calendar"
)

type API struct {
	requester    *calendar.Requester
	weeklyTarget int
}

func New(requester *calendar.Requester, weeklyTarget int) *API {
	return &API{
		requester:    requester,
		weeklyTarget: weeklyTarget,
	}
}

// graphResponse represents the config and graph plots returned from the data API.
type graphResponse struct {
	Plots  []calendar.Plot   `json:"plots"`
	Config map[string]string `json:"config"`
}

func (a *API) Data(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
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
	events, err := a.requester.Fetch()
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
		Config: map[string]string{
			"guideline": strconv.Itoa(guidelineUnits),
		},
		Plots: events.GeneratePlots(scale, float64(a.weeklyTarget)),
	}

	// JSON encode response
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(resp); err != nil {
		log.Printf("failed to JSON encode response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
