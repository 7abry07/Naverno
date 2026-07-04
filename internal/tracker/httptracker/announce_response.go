package httptracker

import "github.com/zeebo/bencode"

type announceResponse struct {
	Failure     string             `bencode:"failure reason"`
	RetryIn     string             `bencode:"retry in"`
	Warning     string             `bencode:"warning message"`
	TrackerID   string             `bencode:"tracker id"`
	MinInterval int64              `bencode:"min interval"`
	Interval    int64              `bencode:"interval"`
	Complete    int64              `bencode:"complete"`
	Incomplete  int64              `bencode:"incomplete"`
	Peers       bencode.RawMessage `bencode:"peers"`
	Peers6      bencode.RawMessage `bencode:"peers6"`
}
