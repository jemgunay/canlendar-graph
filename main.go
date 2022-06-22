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
)

const recommendedWeeklyUnits = 14

func main() {
	// process flags
	port := flag.Int("port", 8080, "the HTTP server port")
	calendarName := flag.String("calendar-name", "Units Consumed", "the name of the calendar documenting units")
	local := flag.Bool("local", false, "use local credentials.json file rather than default env creds")
	flag.Parse()

	// create calendar requester & configure API
	calendarRequester, err := calendar.New(*calendarName, *local)
	if err != nil {
		log.Printf("failed to create calendar requester: %s", err)
		os.Exit(1)
	}

	apiHandlers := api.New(calendarRequester, recommendedWeeklyUnits)

	router := mux.NewRouter()
	// API handlers
	router.HandleFunc("/api/v1/data/{view}", apiHandlers.Data)
	// HTTP file server
	staticFileHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("static/")))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/static/"
		staticFileHandler.ServeHTTP(w, r)
	})
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// start HTTP server
	log.Printf("starting HTTP server on port %d", *port)
	err = http.ListenAndServe(":"+strconv.Itoa(*port), router)
	log.Printf("HTTP server shut down: %s", err)
}
