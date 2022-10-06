package main

import (
	"context"
	"flag"
	"log"
	"os"

	gcal "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func main() {
	calendarID := flag.String("calendar-id", "", "the ID of the Units Consumed calendar, e.g. abcdefghijklmop123456789@group.calendar.google.com")
	credentialsFile := flag.String("creds-file", "./credentials.json", "the credentials file")
	flag.Parse()

	if *calendarID == "" {
		log.Println("no calendar ID provided")
		os.Exit(1)
	}

	if *credentialsFile == "" {
		log.Println("no credentials file provided")
		os.Exit(1)
	}

	srv, err := gcal.NewService(context.Background(), option.WithCredentialsFile(*credentialsFile))
	if err != nil {
		log.Printf("failed to get client: %v", err)
		os.Exit(1)
	}

	entry := &gcal.CalendarListEntry{
		Id: *calendarID,
	}

	// get a list of calendars
	// https://developers.google.com/calendar/api/v3/reference/calendarList/insert
	if _, err = srv.CalendarList.Insert(entry).Do(); err != nil {
		log.Printf("unable to insert calendar: %v", err)
		os.Exit(1)
	}

	log.Println("successfully inserted calendar")
}
