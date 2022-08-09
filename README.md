# Canlendar Graph

A web app for graphing alcohol unit intake documented via Google calendar events. Month, week and day views are available. 

## Setup

1) Create a Service Account (SA) for your project
1) Invite the SA email address to your calendar
1) Accept the invite for your SA with the `cmd/insert/main.go` script, passing in your calendar's ID:
```bash
go run main.go --calendar-id "abcdefghijklmop123456789@group.calendar.google.com"
2022/06/22 00:53:52 successfully inserted
```

### Running Locally

1) Create an API key for the Service Account associated with your calendar
1) Download the JSON creds file to the repo root as `credentials.json`
1) Run with `go run main.go --local`

### Deploy to GCP

```bash
gcloud app deploy
```