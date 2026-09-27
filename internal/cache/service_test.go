package cache

import (
	"context"
	"errors"
	"log"
	"testing"
	"time"

	mwslog "github.com/cself-sdccd-edu/mws-api/internal/log"
)

type testStore struct {
	entry          *Entry
	getErr         error
	claimResult    bool
	claimErr       error
	saveErr        error
	failRefreshErr error
	tryStartCalled int
	saveCalled     int
	failCalled     int
}

func (s *testStore) Get(ctx context.Context, key string) (*Entry, error) {
	return s.entry, s.getErr
}

func (s *testStore) TryStartRefresh(ctx context.Context, key string, lease time.Duration) (bool, error) {
	s.tryStartCalled++
	return s.claimResult, s.claimErr
}

func (s *testStore) Save(ctx context.Context, key string, data []byte) error {
	s.saveCalled++
	return s.saveErr
}

func (s *testStore) FailRefresh(ctx context.Context, key string, err error) error {
	s.failCalled++
	return s.failRefreshErr
}

type testQASClient struct {
	queryCalled int
	data        []byte
	err         error
}

func (c *testQASClient) Query(ctx context.Context, queryName string, term string) ([]byte, error) {
	c.queryCalled++
	return c.data, c.err
}

type testEventLogger struct {
	events []mwslog.LogEvent
	err    error
}

func (l *testEventLogger) Log(ctx context.Context, event mwslog.LogEvent) error {
	l.events = append(l.events, event)
	return l.err
}

func TestGetReturnsAcceptableStaleEntryWhenRefreshClaimFails(t *testing.T) {
	staleData := []byte(`{"result":"stale but acceptable"}`)

	store := &testStore{
		entry: &Entry{
			Data:      staleData,
			HasData:   true,
			UpdatedAt: time.Now().Add(-15 * time.Minute),
		},
		claimErr: errors.New("temporary SQL error"),
	}

	qasClient := &testQASClient{}
	eventLogger := &testEventLogger{}

	service := NewService(
		store,
		qasClient,
		log.Default(),
		eventLogger,
		10*time.Minute,
		12*time.Hour,
		3*time.Minute,
		2*time.Minute,
	)

	entry, err := service.Get(
		context.Background(),
		"schedule:2267:ugrd",
		RefreshRequest{
			QueryName: "X_SR_CLSSCHED_DATA",
			Term:      "2267",
		},
	)

	if err != nil {
		t.Fatalf("expected stale entry without error, got: %v", err)
	}

	if entry == nil {
		t.Fatal("expected stale cache entry, got nil")
	}

	if string(entry.Data) != string(staleData) {
		t.Fatalf("returned data = %q, expected %q", entry.Data, staleData)
	}

	if store.tryStartCalled != 1 {
		t.Fatalf("TryStartRefresh called %d times, expected 1", store.tryStartCalled)
	}

	if qasClient.queryCalled != 0 {
		t.Fatalf("QAS Query called %d times, expected 0", qasClient.queryCalled)
	}

	if store.saveCalled != 0 {
		t.Fatalf("Save called %d times, expected 0", store.saveCalled)
	}

	if !hasEvent(eventLogger.events, "cache_refresh_claim_error") {
		t.Fatal("expected cache_refresh_claim_error event")
	}
}

func TestGetReturnsErrorWhenExpiredEntryCannotClaimRefresh(t *testing.T) {
	store := &testStore{
		entry: &Entry{
			Data:      []byte(`{"result":"too old"}`),
			HasData:   true,
			UpdatedAt: time.Now().Add(-13 * time.Hour),
		},
		claimErr: errors.New("temporary SQL error"),
	}

	qasClient := &testQASClient{}
	eventLogger := &testEventLogger{}

	service := NewService(
		store,
		qasClient,
		log.Default(),
		eventLogger,
		10*time.Minute,
		12*time.Hour,
		3*time.Minute,
		2*time.Minute,
	)

	entry, err := service.Get(
		context.Background(),
		"schedule:2267:ugrd",
		RefreshRequest{
			QueryName: "X_SR_CLSSCHED_DATA",
			Term:      "2267",
		},
	)

	if err == nil {
		t.Fatal("expected refresh-claim error for excessively old entry")
	}

	if entry != nil {
		t.Fatal("expected no returned entry when cached data is too old")
	}

	if store.tryStartCalled != 1 {
		t.Fatalf("TryStartRefresh called %d times, expected 1", store.tryStartCalled)
	}

	if qasClient.queryCalled != 0 {
		t.Fatalf("QAS Query called %d times, expected 0", qasClient.queryCalled)
	}
}

func hasEvent(events []mwslog.LogEvent, eventName string) bool {
	for _, event := range events {
		if event.Event == eventName {
			return true
		}
	}

	return false
}
