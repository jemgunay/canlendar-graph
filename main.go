package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/jemgunay/canlendar-graph/calendar"
)

var calendarName = "Units Consumed"

func main() {
	// process flags
	port := flag.Int("port", 8080, "the HTTP server port")
	flag.StringVar(&calendarName, "calendar_name", calendarName, "the name of the calendar documenting units")
	flag.Parse()

	// define handlers
	staticFileHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("static/")))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// reroute calls to the root to index.html via the file server
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		r.URL.Path = "/static/"
		staticFileHandler.ServeHTTP(w, r)
	})
	http.Handle("/static/", staticFileHandler)
	http.HandleFunc("/data", dataHandler)

	// start HTTP server
	log.Printf("starting up HTTP server on port %d", *port)
	err := http.ListenAndServe(":"+strconv.Itoa(*port), nil)
	log.Printf("server shut down: %s", err)
}

type graphResponse struct {
	Plots  []calendar.Plot   `json:"plots"`
	Config map[string]string `json:"config"`
}

const recommendedWeeklyUnits = 14

// serves up JSON data to be consumed by the root page
func dataHandler(w http.ResponseWriter, r *http.Request) {
	// extract query string params
	pretty := r.URL.Query().Get("pretty") == "true"
	view := r.URL.Query().Get("view")
	if view == "" {
		log.Println("no view query string provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// fetch calendar events
	events, err := calendar.Fetch(calendarName)
	if err != nil {
		log.Printf("failed to fetch units for \"%s\": %s", view, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	guidelineUnits := recommendedWeeklyUnits
	if view == "month" {
		guidelineUnits *= 4
	}

	// construct response
	resp := graphResponse{
		Config: map[string]string{
			"guideline": strconv.Itoa(guidelineUnits),
		},
	}

	switch view {
	case "month":
		resp.Plots = events.GeneratePlots(calendar.Month)
	case "week":
		resp.Plots = events.GeneratePlots(calendar.Week)
	case "day":
		resp.Plots = events.GeneratePlots(calendar.Day)
	}

	// JSON encode response
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if pretty {
		encoder.SetIndent("", "\t")
	}
	if err := encoder.Encode(resp); err != nil {
		log.Printf("failed to JSON encode response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
