package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jemgunay/canlendar-graph/api"
	"github.com/jemgunay/canlendar-graph/calendar"
	"github.com/jemgunay/canlendar-graph/config"
	"github.com/jemgunay/canlendar-graph/storage/influx"
)

func main() {
	// process flags
	calendarName := flag.String("calendar-name", "Units Consumed", "the name of the calendar documenting units")
	local := flag.Bool("local", false, "use local credentials.json file rather than default env creds")
	flag.Parse()

	conf := config.New()

	// create requesters & configure API
	const recommendedWeeklyUnits = 14
	calendarRequester, err := calendar.New(*calendarName, *local, recommendedWeeklyUnits)
	if err != nil {
		log.Printf("failed to create calendar requester: %s", err)
		os.Exit(1)
	}

	influxRequester := influx.New(conf.Influx)
	apiHandlers := api.New(influxRequester, calendarRequester)

	router := mux.NewRouter()
	// API handlers
	router.HandleFunc("/api/v1/query", apiHandlers.Query).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/collect", apiHandlers.Collect).Methods(http.MethodPost)

	// HTTP file server
	staticFileHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("static/")))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/static/"
		staticFileHandler.ServeHTTP(w, r)
	})
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// start HTTP server
	log.Printf("starting HTTP server on port %d", conf.Port)
	err = http.ListenAndServe(":"+strconv.Itoa(conf.Port), router)
	log.Printf("HTTP server shut down: %s", err)
}
