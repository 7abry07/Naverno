package trackermanager

import (
	"Naverno/internal/tracker"
	"Naverno/internal/tracker/httptracker"
	"fmt"
	"net/http"
	"net/url"
)

type TrackerManager struct {
	httpTransport *http.Transport
}

func New() *TrackerManager {
	return &TrackerManager{
		httpTransport: &http.Transport{},
	}
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
