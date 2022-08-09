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

type Storer interface {
	Query(ctx context.Context, options ...QueryOption) ([]Record, error)
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

func NewQuery(opts ...QueryOption) QuerySet {
	q := &QuerySet{
		EndTime:     time.Now().UTC(),
		Aggregation: Week,
	}
	for _, opt := range opts {
		opt(q)
	}

	return *q
}

type QueryOption func(set *QuerySet)

func WithStartTime(startTime time.Time) QueryOption {
	return func(set *QuerySet) {
		set.StartTime = startTime
	}
}

func WithEndTime(endTime time.Time) QueryOption {
	return func(set *QuerySet) {
		set.EndTime = endTime
	}
}

type Aggregation string

const (
	Day   Aggregation = "day"
	Week  Aggregation = "week"
	Month Aggregation = "month"
	Year  Aggregation = "year"
)

func (a Aggregation) Unit() string {
	switch a {
	case Day:
		return "1d"
	case Week:
		return "1w"
	case Month:
		return "1mo"
	case Year:
		return "1y"
	}
	return ""
}

func WithAggregation(aggregation Aggregation) QueryOption {
	return func(set *QuerySet) {
		set.Aggregation = aggregation
	}
}
