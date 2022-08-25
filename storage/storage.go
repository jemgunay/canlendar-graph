package storage

import (
	"context"
	"errors"
	"time"
)

//go:generate mockgen -source=storage.go -destination=mocks/storage.go

// Record is a generic collection dataset retrieved and stored by a Storer.
type Record struct {
	Time   time.Time
	Tags   map[string]string
	Fields map[string]interface{}
}

// Plot is a point on a graph.
type Plot struct {
	X int64   `json:"t"`
	Y float64 `json:"y"`
}

// Storer stored records and queries for records from a data store. It also provides the means to fetch the timestamps
// for the first and last records.
type Storer interface {
	Store(ctx context.Context, records ...Record) error
	Query(ctx context.Context, options ...QueryOption) ([]Plot, error)
	ReadLastTimestamp(ctx context.Context) (time.Time, error)
	ReadFirstTimestamp(ctx context.Context) (time.Time, error)
}

// ErrNoResults indicates that there are no results for the executed query.
var ErrNoResults = errors.New("no results found for query")

// QuerySet defines the parameters for a query.
type QuerySet struct {
	StartTime   time.Time
	EndTime     time.Time
	Aggregation Aggregation
}

// FormatStartTime formats the start time as RFC3339.
func (q QuerySet) FormatStartTime() string {
	return q.StartTime.Format(time.RFC3339)
}

// FormatEndTime formats the end time as RFC3339.
func (q QuerySet) FormatEndTime() string {
	return q.EndTime.Format(time.RFC3339)
}

// NewQuery validates a set of query options and configures a QuerySet given the provided options.
func NewQuery(opts ...QueryOption) (QuerySet, error) {
	q := &QuerySet{}
	for _, opt := range opts {
		opt(q)
	}

	// validate provided option values
	switch {
	case !q.Aggregation.IsValid():
		return *q, errors.New("unsupported aggregation")
	case q.EndTime.IsZero():
		return *q, errors.New("end time must not be zero")
	case q.StartTime.After(q.EndTime):
		return *q, errors.New("start time must not exceed end time")
	}
	return *q, nil
}

// QueryOption is used to provide options to storage queries.
type QueryOption func(set *QuerySet)

// WithStartTime defines the query start time to use.
func WithStartTime(startTime time.Time) QueryOption {
	return func(set *QuerySet) {
		set.StartTime = startTime.UTC()
	}
}

// WithEndTime defines the query end time to use.
func WithEndTime(endTime time.Time) QueryOption {
	return func(set *QuerySet) {
		set.EndTime = endTime.UTC()
	}
}

// Aggregation describes how storage data should be aggregated.
type Aggregation string

const (
	Day   Aggregation = "day"
	Week  Aggregation = "week"
	Month Aggregation = "month"
	Year  Aggregation = "year"
)

// String gets the aggregation name.
func (a Aggregation) String() string {
	return string(a)
}

// IsValid determines if the query aggregation is supported.
func (a Aggregation) IsValid() bool {
	switch a {
	case Day, Week, Month, Year:
		return true
	default:
		return false
	}
}

// WithAggregation defines the query aggregation to use.
func WithAggregation(aggregation Aggregation) QueryOption {
	return func(set *QuerySet) {
		set.Aggregation = aggregation
	}
}
