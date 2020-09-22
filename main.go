package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	gcal "google.golang.org/api/calendar/v3"

	"github.com/jemgunay/canlendar-graph/calendar"
)

func main() {
	// process flags
	port := flag.Int("port", 8080, "the HTTP server port")
	flag.StringVar(&calendar.Name, "calendar_name", calendar.Name, "the units consumed calendar name")
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
	view := r.URL.Query().Get("view")
	if view == "" {
		log.Println("no view query string provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// fetch calendar events
	events, err := calendar.Fetch()
	if err != nil {
		log.Printf("failed to fetch units for \"%s\": %s", view, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// group events by week and sum each week's units
	var (
		firstEventDate, lastEventDate time.Time
		eventsByWeek                  = make(map[int64]float64)
		scale                         = time.Hour * 24 * 7
	)
	for i, event := range events {
		date, units, err := processEvent(event)
		if err != nil {
			log.Printf("failed to parse event date for index %d", i)
			continue
		}

		// truncate date to week
		weekDate := date.UTC().Truncate(scale)

		// store first and last events to calculate number of weeks
		if i == 0 {
			firstEventDate = weekDate
		} else if i == len(events)-1 {
			lastEventDate = weekDate
		}

		epoch := weekDate.Unix() * 1000
		eventsByWeek[epoch] = eventsByWeek[epoch] + units
	}

	// determine number of weeks (i.e. plot points) to generate
	numPlots := int(lastEventDate.Sub(firstEventDate).Hours()/scale.Hours()) + 1

	// for each required plot point, determine the corresponding date and units from previously processed events. Plot
	// points without a corresponding event will be defaulted to 0 units
	graphUnits := make([]graphUnit, 0, numPlots)
	for i := 0; i < numPlots; i++ {
		nextWeek := firstEventDate.Add(scale * time.Duration(i))
		epoch := nextWeek.Unix() * 1000

		graphUnits = append(graphUnits, graphUnit{
			X: epoch,
			Y: eventsByWeek[epoch],
		})
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

func processEvent(event *gcal.Event) (time.Time, float64, error) {
	date := event.Start.DateTime
	if date == "" {
		date = event.Start.Date
	}

	// parse date from string
	parsedTime, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, 0, err
	}

	// parse summary into units
	var units float64
	switch match := re.FindString(event.Summary); match {
	case "":
		return time.Time{}, 0, fmt.Errorf("failed to find a units number for event: %s", event.Summary)

	case "?":
		// default unknowns to recommended amount
		units = recommendedWeeklyUnits

	default:
		units, err = strconv.ParseFloat(match, 64)
		if err != nil {
			return time.Time{}, 0, fmt.Errorf("failed to parse units number from event %s", event.Summary)
		}
	}

	return parsedTime, units, nil
}
