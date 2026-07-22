package tracker

type TrackerEvent uint8

const (
	TRACKER_NONE TrackerEvent = iota
	TRACKER_STARTED
	TRACKER_COMPLETED
	TRACKER_STOPPED
)

var trackerEventName = map[TrackerEvent]string{
	TRACKER_NONE:      "",
	TRACKER_STARTED:   "started",
	TRACKER_COMPLETED: "completed",
	TRACKER_STOPPED:   "stopped",
}

func (e TrackerEvent) String() string {
	return trackerEventName[e]
}
