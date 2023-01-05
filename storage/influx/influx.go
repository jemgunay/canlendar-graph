package influx

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdbapi "github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/jemgunay/canlendar-graph/storage"

	"github.com/jemgunay/canlendar-graph/config"
)

var _ storage.Storer = (*Requester)(nil)

const (
	bucket      = "life-metrics"
	measurement = "alcohol_units"
)

// Requester is used to write to and query influx.
type Requester struct {
	writeClient influxdbapi.WriteAPIBlocking
	readClient  influxdbapi.QueryAPI
}

// New returns an initialised influx requester.
func New(conf config.Influx) Requester {
	client := influxdb2.NewClient(conf.Host, conf.Token)
	return Requester{
		writeClient: client.WriteAPIBlocking(conf.Org, bucket),
		readClient:  client.QueryAPI(conf.Org),
	}
}

type aggregationConfig struct {
	unit   string
	offset string
}

func newAggregationConfig(unit, offset string) aggregationConfig {
	return aggregationConfig{
		unit:   unit,
		offset: offset,
	}
}

// maps aggregation to their Influx-query segments.
var aggregationLookup = map[storage.Aggregation]aggregationConfig{
	storage.Day:   newAggregationConfig("1d", "0s"),
	storage.Week:  newAggregationConfig("1w", "-3d"),
	storage.Month: newAggregationConfig("1mo", "0s"),
	storage.Year:  newAggregationConfig("1y", "0s"),
}

// Query queries Influx given the provided options. Returns storage.ErrNoResults if no records are available for the
// given time range.
func (r Requester) Query(ctx context.Context, options ...storage.QueryOption) ([]storage.Plot, error) {
	log.Printf("executing influx query")

	queryOpts, err := storage.NewQuery(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// if no start time is provided, default it to the oldest record's timestamp
	if queryOpts.StartTime.IsZero() {
		queryOpts.StartTime, err = r.ReadFirstTimestamp(ctx)
		if err != nil {
			if err != storage.ErrNoResults {
				return nil, storage.ErrNoResults
			}
			return nil, fmt.Errorf("failed to determine start time from available records: %w", err)
		}
	}

	aggregate, ok := aggregationLookup[queryOpts.Aggregation]
	if !ok {
		return nil, errors.New("unsupported aggregation provided")
	}

	// build flux query
	query := `from(bucket: "` + bucket + `")
	  	|> range(start: ` + queryOpts.FormatStartTime() + `, stop: ` + queryOpts.FormatEndTime() + `)
	  	|> filter(fn:(r) =>
	    	r._measurement == "` + measurement + `" and
			r._field == "units"
	  	)
		|> aggregateWindow(every: ` + aggregate.unit + `, fn: sum, createEmpty: true, offset: ` + aggregate.offset + `)`

	result, err := r.readClient.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query influx: %w", err)
	}

	var records []storage.Plot
	for result.Next() {
		var units float64

		val := result.Record().Value()
		switch tVal := val.(type) {
		case float64:
			units = tVal
		case nil:
			units = 0
		default:
			log.Printf("unexpected unit value read from record: %s (%T)", val, val)
			continue
		}

		// influx returns the upper of the aggregated date range, so subtract to give us the start date (i.e. first day
		// of the year, month, day - rather than the last)
		recordTime := result.Record().Time()
		switch queryOpts.Aggregation {
		case storage.Year:
			recordTime = recordTime.AddDate(-1, 0, 0)
		case storage.Month:
			recordTime = recordTime.AddDate(0, -1, 0)
		case storage.Week:
			recordTime = recordTime.AddDate(0, 0, -7)
		case storage.Day:
			recordTime = recordTime.AddDate(0, 0, -1)
		}

		records = append(records, storage.Plot{
			// millis required by ChartJS
			X: recordTime.UnixMilli(),
			Y: units,
		})
	}

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse influx query response: %w", err)
	}

	if len(records) == 0 {
		return nil, storage.ErrNoResults
	}

	return records, nil
}

// ReadLastTimestamp returns the timestamp for the record with the newest timestamp. storage.ErrNoResults is returned
// if there are no records.
func (r Requester) ReadLastTimestamp(ctx context.Context) (time.Time, error) {
	log.Printf("reading last record timestamp from influx")

	query := `from(bucket: "` + bucket + `")
  	|> range(start: 0, stop: now())
  	|> filter(fn:(r) =>
    	r._measurement == "` + measurement + `" and
		r._field == "units"
  	)
  	|> last()`

	result, err := r.readClient.Query(ctx, query)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to query influx: %w", err)
	}

	var t time.Time
	if result.Next() {
		t = result.Record().Time()
	}

	if err := result.Err(); err != nil {
		return time.Time{}, fmt.Errorf("failed to parse influx query response: %w", err)
	}

	if t.IsZero() {
		return time.Time{}, storage.ErrNoResults
	}

	return t, nil
}

// ReadFirstTimestamp returns the timestamp for the record with the oldest timestamp. storage.ErrNoResults is returned
// if there are no records.
func (r Requester) ReadFirstTimestamp(ctx context.Context) (time.Time, error) {
	log.Printf("reading first record timestamp from influx")

	query := `from(bucket: "` + bucket + `")
  	|> range(start: 0, stop: now())
  	|> filter(fn:(r) =>
    	r._measurement == "` + measurement + `" and
		r._field == "units"
  	)
  	|> first()`

	result, err := r.readClient.Query(ctx, query)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to query influx: %w", err)
	}

	t := time.Now().UTC()
	if result.Next() {
		t = result.Record().Time()
	}

	if err := result.Err(); err != nil {
		return time.Time{}, fmt.Errorf("failed to parse influx query response: %w", err)
	}

	if t.IsZero() {
		return time.Time{}, storage.ErrNoResults
	}

	return t, nil
}

// Store writes the set of records to Influx as a batch of points.
func (r Requester) Store(ctx context.Context, records ...storage.Record) error {
	log.Printf("storing records to influx: %d", len(records))

	// no new data to store so skip writing to influx
	if len(records) == 0 {
		return nil
	}

	points := make([]*write.Point, 0, len(records))
	for _, result := range records {
		point := influxdb2.NewPoint(
			measurement,
			result.Tags,
			result.Fields,
			result.Time,
		)
		points = append(points, point)
	}

	if err := r.writeClient.WritePoint(ctx, points...); err != nil {
		return fmt.Errorf("writing points to influx failed: %w", err)
	}

	return nil
}
