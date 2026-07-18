package trackertest

import (
	"Naverno/internal/tracker"
	"context"
	"fmt"
)

type MockFailingTracker struct {
	announce string
}

func NewFailingMock() *MockFailingTracker {
	return &MockFailingTracker{"failing"}
}

func (t *MockFailingTracker) Announce(ctx context.Context, req tracker.AnnounceRequest) (*tracker.AnnounceResponse, error) {
	return nil, fmt.Errorf("supposed to fail")
}

func (t *MockFailingTracker) URL() string {
	return t.announce
}
