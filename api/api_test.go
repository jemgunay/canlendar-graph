package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/jemgunay/canlendar-graph/calendar"
	"github.com/jemgunay/canlendar-graph/storage"

	mock_calendar "github.com/jemgunay/canlendar-graph/calendar/mocks"
	mock_storage "github.com/jemgunay/canlendar-graph/storage/mocks"
)

func TestAPI_Collect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Now()

	// generate events for the calendar event iterator
	mockIter := mock_calendar.NewMockEventIterator(ctrl)
	eventCount := 100
	mockIter.EXPECT().Count().Return(eventCount)
	var current int
	mockIter.EXPECT().Next().DoAndReturn(
		func() (calendar.Event, error) {
			if current == eventCount {
				return calendar.Event{}, calendar.ErrNoMoreEvents
			}

			ev := calendar.Event{
				Date:  now.Add(-time.Hour * 24 * time.Duration(current)),
				Units: float64(current % 10),
			}
			current++
			return ev, nil
		},
	).AnyTimes()

	mockCalendar := mock_calendar.NewMockFetcher(ctrl)
	mockCalendar.EXPECT().Fetch(gomock.Any(), gomock.Any()).Return(mockIter, nil)

	mockStorer := mock_storage.NewMockStorer(ctrl)
	mockStorer.EXPECT().ReadLastTimestamp(gomock.Any()).Return(time.Time{}, storage.ErrNoResults)
	var storedCount int
	mockStorer.EXPECT().Store(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, records ...storage.Record) error {
		storedCount++
		return nil
	})

	api := New(mockStorer, mockCalendar)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
	api.Collect(w, r)

	// validate status
	status := w.Result().StatusCode
	if status != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, status)
	}

	expectedStoredCount := 1
	if storedCount != expectedStoredCount {
		t.Fatalf("expected %d, got %d", expectedStoredCount, storedCount)
	}
}

func TestAPI_Query(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cases := []struct {
		name        string
		aggregation string
		plots       []storage.Plot
		status      int
		respBody    string
	}{
		{
			name:        "week_empty",
			aggregation: "week",
			plots:       []storage.Plot{},
			status:      http.StatusOK,
			respBody:    `{"plots":[],"metadata":{"guideline":14}}`,
		},
		{
			name:        "week_non_empty",
			aggregation: "week",
			plots: []storage.Plot{
				{1, 1},
				{2, 2},
				{3, 3},
			},
			status:   http.StatusOK,
			respBody: `{"plots":[{"t":1,"y":1},{"t":2,"y":2},{"t":3,"y":3}],"metadata":{"guideline":14}}`,
		},
		{
			name:        "day_non_empty",
			aggregation: "day",
			plots: []storage.Plot{
				{1, 1},
				{2, 2},
				{3, 3},
			},
			status:   http.StatusOK,
			respBody: `{"plots":[{"t":1,"y":1},{"t":2,"y":2},{"t":3,"y":3}],"metadata":{}}`,
		},
		{
			name:        "month_empty",
			aggregation: "month",
			plots:       []storage.Plot{},
			status:      http.StatusOK,
			respBody:    `{"plots":[],"metadata":{"guideline":56}}`,
		},
		{
			name:        "year_empty",
			aggregation: "year",
			plots:       []storage.Plot{},
			status:      http.StatusOK,
			respBody:    `{"plots":[],"metadata":{"guideline":672}}`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			mockStorer := mock_storage.NewMockStorer(ctrl)
			mockStorer.EXPECT().Query(gomock.Any(), gomock.Any()).Return(tt.plots, nil).AnyTimes()

			w := httptest.NewRecorder()
			query := fmt.Sprintf("/?aggregation=%s&end_time=%s", tt.aggregation, time.Now().Format(time.RFC3339))
			r := httptest.NewRequest(http.MethodGet, query, nil)

			api := New(mockStorer, nil)
			api.Query(w, r)

			// validate status
			status := w.Result().StatusCode
			if status != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, status)
			}

			respBody, err := io.ReadAll(w.Result().Body)
			if err != nil {
				t.Fatalf("failed to read resp body: %s", err)
			}

			respBody = bytes.TrimSpace(respBody)
			if string(respBody) != tt.respBody {
				t.Fatalf("expected '%s', got '%s'", tt.respBody, respBody)
			}
		})
	}
}
