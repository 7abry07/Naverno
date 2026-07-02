package http_tracker

import "Naverno/internal/bencode"

type announceResponse struct {
	failure     string
	warning     string
	retryIn     string
	minInterval int32
	interval    int32
	complete    int64
	incomplete  int64

	peers  bencode.BNode
	peers6 bencode.BNode
}
