package influx

import (
	"context"
	"fmt"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdbapi "github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/jemgunay/canlendar-graph/storage"

	"github.com/jemgunay/canlendar-graph/config"
)

var _ storage.Storer = &Requester{}

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

func (r Requester) Query(ctx context.Context, options ...storage.QueryOption) ([]storage.Record, error) {
	queryOpts := storage.NewQuery(options...)
	// TODO: aggregate sum by window

	query := `from(bucket: "` + bucket + `")
	  	|> range(start: ` + queryOpts.StartTime.Format(time.RFC3339) + `, stop: ` + queryOpts.EndTime.Format(time.RFC3339) + `)
	  	|> filter(fn:(r) =>
	    	r._measurement == "` + measurement + `"
	  	)
		|> aggregateWindow(every: ` + queryOpts.Aggregation.Unit() + `, fn: sum)`

	result, err := r.readClient.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query influx: %s", err)
	}

	data := make(map[string]interface{})
	for result.Next() {
		data[result.Record().Field()] = result.Record().Value()
	}

	// return data, nil
	return nil, nil
}

func (r Requester) ReadLastTimestamp(ctx context.Context) (time.Time, error) {
	query := `from(bucket: "` + bucket + `")
  	|> range(start: 0, stop: now())
  	|> filter(fn:(r) =>
    	r._measurement == "` + measurement + `"
  	)
  	|> last()`

	result, err := r.readClient.Query(ctx, query)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to query influx: %s", err)
	}

	var t time.Time
	for result.Next() {
		t2 := result.Record().Time()
		if t2.After(t) {
			t = t2
		}
	}

	if t.IsZero() {
		return time.Time{}, storage.ErrNoResults
	}

	return t, nil
}

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
		return fmt.Errorf("writing points to influx failed: %s", err)
	}

	return nil
}
