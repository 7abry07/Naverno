package trackermanager

import (
	"Naverno/internal/tracker"
	"Naverno/internal/tracker/httptracker"
	"log/slog"
	"net/http"
	"net/url"
)

// -------------- Structs -------------------

type TrackerManager struct {
	logger *slog.Logger

	httpTransport *http.Transport
}

// -------------- Functions -------------------

func New(logger *slog.Logger) *TrackerManager {
	m := TrackerManager{}
	m.logger = logger
	m.httpTransport = &http.Transport{}
	return &m
}

// -------------- Methods -------------------

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
		httpTracker := httptracker.New(m.logger, parsedAnnounce, m.httpTransport)
		return httpTracker, nil
	}

	return nil, tracker.InvalidSchemeErr
}
