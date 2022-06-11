package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jemgunay/canlendar-graph/calendar"
)

var calendarName = "Units Consumed"

func main() {
	// process flags
	port := flag.Int("port", 8080, "the HTTP server port")
	flag.StringVar(&calendarName, "calendar_name", calendarName, "the name of the calendar documenting units")
	flag.Parse()

	// define handlers
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/data/{view}", dataHandler)
	staticFileHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("static/")))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/static/"
		staticFileHandler.ServeHTTP(w, r)
	})
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// start HTTP server
	log.Printf("starting HTTP server on port %d", *port)
	err := http.ListenAndServe(":"+strconv.Itoa(*port), router)
	log.Printf("HTTP server shut down: %s", err)
}

type graphResponse struct {
	Plots  []calendar.Plot   `json:"plots"`
	Config map[string]string `json:"config"`
}

// TODO: get this from elsewhere, i.e. DB driven by config page, persisted in memory
const recommendedWeeklyUnits = 14

// serves up JSON data to be consumed by the root page
func dataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// extract URL/query string params
	pretty := r.URL.Query().Get("pretty") == "true"
	vars := mux.Vars(r)
	view := vars["view"]
	if view == "" {
		log.Println("no view string provided in URL route")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	scale, err := calendar.ValidateScale(view)
	if err != nil {
		log.Printf("failed to parse ")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// fetch calendar events
	events, err := calendar.Fetch(calendarName)
	if err != nil {
		log.Printf("failed to fetch units for \"%s\": %s", scale, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	guidelineUnits := recommendedWeeklyUnits
	if scale == calendar.Month {
		guidelineUnits *= 4
	}

	// construct response
	resp := graphResponse{
		Config: map[string]string{
			"guideline": strconv.Itoa(guidelineUnits),
		},
		Plots: events.GeneratePlots(scale),
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
