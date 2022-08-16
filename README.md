# Canlendar Graph

A web app for graphing alcohol unit intake documented via Google calendar events. Year, month, week and day views are available.

## Implementation

* Deployed to Google Cloud Run and scales to zero.
* File server for serving static web app files.
* `/api/v1/collect`: Endpoint for scraping alcohol unit data from calendar events via the Google Calendar API; this unit data is then stored in InfluxDB. This endpoint is executed on an interval via Cloud Scheduler. 
* `/api/v1/query`: Endpoint for querying alcohol unit consumption data stored in InfluxDB.

## Setup

1) Create a Service Account (SA) for your project 
1) Invite the SA email address to your calendar
1) Create a key for the Service Account associated with your calendar; download the credentials key file
1) Accept the invite for your SA with the `cmd/insert/main.go` script, passing in your calendar's ID:
```bash
go run main.go --calendar-id "abcdefghijklmop123456789@group.calendar.google.com" --creds-file ./credentials.json
2022/06/22 00:53:52 successfully inserted
```

### Running Locally

1) Ensure the credentials key file downloaded in the setup stage resides as `./config/credentials.json`
1) Run with `go run main.go --local`

### Deploy to GCP

```bash
gcloud run deploy
```