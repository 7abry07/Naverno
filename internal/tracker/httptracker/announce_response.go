package httptracker

type announceResponse struct {
	failure     string
	warning     string
	retryIn     string
	minInterval int64
	interval    int64
	complete    int64
	incomplete  int64

	peers  any
	peers6 any
}
