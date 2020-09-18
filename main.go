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
	X int64   `json:"t"`
	Y float64 `json:"y"`
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
		log.Printf("failed to fetch units for \"%s\": %s", operation, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	graphUnits := make([]graphUnit, 0, len(events))
	var currentDateEpoch int64
	for i, item := range events {
		// process date
		date := item.Start.DateTime
		if date == "" {
			date = item.Start.Date
		}

		// determine epoch (ms) from date string if the date is known for this experience (YYYY-MM-DD)
		parsedDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			log.Printf("failed to parse %s into datetime on event %d", date, i)
			continue
		}
		truncatedDate := parsedDate.UTC().Truncate(time.Hour * 24 * 7)
		truncatedDateEpoch := truncatedDate.Unix() * 1000

		// if week has changed, add to units set
		if currentDateEpoch != truncatedDateEpoch {
			graphUnits = append(graphUnits, graphUnit{
				X: truncatedDateEpoch,
			})
			currentDateEpoch = truncatedDateEpoch
		}

		// parse summary into units
		var units float64
		switch match := re.FindString(item.Summary); match {
		case "":
			log.Printf("failed to parse units number from %s on event %d", item.Summary, i)
			continue
		case "?":
			// default unknowns to recommended amount
			units = recommendedWeeklyUnits
		default:
			var err error
			units, err = strconv.ParseFloat(match, 64)
			if err != nil {
				log.Printf("failed to parse units number from %s on event %d", item.Summary, i)
				continue
			}
		}

		graphUnits[len(graphUnits)-1].Y += units
	}

	// JSON encode response
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	if pretty {
		encoder.SetIndent("", "\t")
	}
	if err := encoder.Encode(graphUnits); err != nil {
		log.Printf("failed to JSON encode response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
