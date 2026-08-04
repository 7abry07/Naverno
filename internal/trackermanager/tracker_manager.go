package trackermanager

import (
	"Naverno/internal/tracker"
	"Naverno/internal/tracker/httptracker"
	"Naverno/internal/tracker/udptracker"
	"fmt"
	"net/http"
	"net/url"
)

type TrackerManager struct {
	httpTransport *http.Transport
	udpTransport  *udptracker.UDPTransport
}

func New() *TrackerManager {

	udpTransport := udptracker.NewUDPTransport()
	go udpTransport.Run()
	return &TrackerManager{
		httpTransport: &http.Transport{},
		udpTransport:  udpTransport,
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
	case "udp":
		udpTracker := udptracker.New(*parsedAnnounce, m.udpTransport)
		return udpTracker, nil
	}

	return nil, fmt.Errorf("the announce URL scheme is neither http or https")
}
