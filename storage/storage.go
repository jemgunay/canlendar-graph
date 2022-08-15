package storage

import (
	"context"
	"errors"
	"time"
)

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

type Storer interface {
	Query(ctx context.Context, options ...QueryOption) ([]Plot, error)
	ReadLastTimestamp(ctx context.Context) (time.Time, error)
	Store(ctx context.Context, records ...Record) error
}

// ErrNoResults indicates that there are no results for the executed query.
var ErrNoResults = errors.New("no results found for query")

type QuerySet struct {
	StartTime   time.Time
	EndTime     time.Time
	Aggregation Aggregation
}

func (q QuerySet) FormatStartTime() string {
	return q.StartTime.Format(time.RFC3339)
}

func (q QuerySet) FormatEndTime() string {
	return q.EndTime.Format(time.RFC3339)
}

func NewQuery(opts ...QueryOption) (QuerySet, error) {
	q := &QuerySet{}
	for _, opt := range opts {
		opt(q)
	}

	// check for required fields
	if !q.Aggregation.IsValid() {
		return *q, errors.New("unsupported aggregation")
	}

	// set defaults
	if q.EndTime.IsZero() {
		return *q, errors.New("end time must not be zero")
	}
	if q.StartTime.After(q.EndTime) {
		return *q, errors.New("start time must not exceed end time")
	}
	return *q, nil
}

type QueryOption func(set *QuerySet)

func WithStartTime(startTime time.Time) QueryOption {
	return func(set *QuerySet) {
		set.StartTime = startTime.UTC()
	}
}

func WithEndTime(endTime time.Time) QueryOption {
	return func(set *QuerySet) {
		set.EndTime = endTime.UTC()
	}
}

type Aggregation string

const (
	Day   Aggregation = "day"
	Week  Aggregation = "week"
	Month Aggregation = "month"
	Year  Aggregation = "year"
)

func (a Aggregation) String() string {
	return string(a)
}

func (a Aggregation) IsValid() bool {
	switch a {
	case Day, Week, Month, Year:
		return true
	default:
		return false
	}
}

func WithAggregation(aggregation Aggregation) QueryOption {
	return func(set *QuerySet) {
		set.Aggregation = aggregation
	}
}
