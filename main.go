package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/jemgunay/canlendar-graph/calendar"
)

func main() {
	// process flags
	port := flag.Int("port", 8080, "the HTTP server port")
	flag.StringVar(&calendar.CalendarName, "calendar_name", calendar.CalendarName, "the units consumed calendar name")
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

type graphUnit struct {
	X int64  `json:"t"`
	Y string `json:"y"`
}

var re = regexp.MustCompile(`(?m)[\d?]*\.?\d*`)

const (
	recommendedWeeklyUnits  = 14
	recommendedMonthlyUnits = recommendedWeeklyUnits * 4
)

// serves up JSON data to be consumed by the root page
func dataHandler(w http.ResponseWriter, r *http.Request) {
	// extract query string params
	pretty := r.URL.Query().Get("pretty") == "true"
	operation := r.URL.Query().Get("operation")
	if operation == "" {
		log.Println("no operation query string provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// fetch calendar events
	events, err := calendar.Fetch()
	if err != nil {
		log.Printf("failed to fetch data for \"%s\": %s", operation, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data := make([]graphUnit, 0, len(events))
	for i, item := range events {
		// process date
		date := item.Start.DateTime
		if date == "" {
			date = item.Start.Date
		}

		// determine epoch (ms) from date string if the date is known for this experience
		parsedDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			log.Printf("failed to parse %s into datetime on event %d", date, i)
			continue
		}
		dateEpoch := parsedDate.Unix() * 1000

		// parse summary into units
		match := re.FindString(item.Summary)
		if match == "" {
			log.Printf("failed to parse units number from %s on event %d", item.Summary, i)
			continue
		} else if match == "?" {
			// default unknowns to recommended amount
			match = strconv.Itoa(recommendedWeeklyUnits)
		}

		data = append(data, graphUnit{
			X: dateEpoch,
			Y: match,
		})
	}

	// JSON encode response
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if pretty {
		encoder.SetIndent("", "\t")
	}
	if err := encoder.Encode(data); err != nil {
		log.Printf("failed to JSON encode response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
