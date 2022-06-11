# Canlendar Graph

A web app for graphing alcohol unit intake documented via Google calendar events. Month, week and day views are available. 

## Run Locally

```bash
go build && ./canlendar-graph
```

## Deploy to GCP

```bash
make deploy
make attach_log
```

## TODO

- Move secrets to gcloud 
- Fix month view + clean up dirty diff func 
- Spinner for loading graphs
- Improve auth + hooking up to Google API