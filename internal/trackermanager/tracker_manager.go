package trackermanager

import (
	"Naverno/internal/tracker"
	"Naverno/internal/tracker/httptracker"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

type TrackerManager struct {
	logger *slog.Logger

	httpTransport *http.Transport
}

func New(logger *slog.Logger) *TrackerManager {
	m := TrackerManager{}

	if logger == nil {
		panic("passed logger is nil")
	}

	m.logger = logger
	m.httpTransport = &http.Transport{}
	return &m
}

func (m *TrackerManager) Close() {
	m.httpTransport.CloseIdleConnections()
}

func (m *TrackerManager) Get(announce string) (tracker.Tracker, error) {
	parsedAnnounce, err := url.Parse(announce)
	if err != nil {
		return nil, err
	}

	switch parsedAnnounce.Scheme {
	case "http", "https":
		httpTracker := httptracker.New(*parsedAnnounce, m.httpTransport)
		return httpTracker, nil
	}

	return nil, fmt.Errorf("the announce URL scheme is neither http or https")
}
