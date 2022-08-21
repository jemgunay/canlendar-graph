package api

import (
	"bytes"
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
	mockStorer.EXPECT().Store(gomock.Any(), gomock.Any()).Return(nil)

	api := New(mockStorer, mockCalendar)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{}`))
	api.Collect(w, r)

	// validate status
	status := w.Result().StatusCode
	if status != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, status)
	}

	// TODO: validate body
}

func TestAPI_Query(t *testing.T) {
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
	opts := []storage.QueryOption{
		storage.WithAggregation(),
	}
	mockStorer.EXPECT().Query(gomock.Any(), gomock.Any()).Return(, nil)

	api := New(mockStorer, mockCalendar)

	w := httptest.NewRecorder()
	query := "/?aggregation=week"
	r := httptest.NewRequest(http.MethodGet, query, nil)
	api.Collect(w, r)

	// validate status
	status := w.Result().StatusCode
	if status != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, status)
	}
}
