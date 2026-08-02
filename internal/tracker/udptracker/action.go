package udptracker

type action uint32

const (
	action_connect action = iota
	action_announce
	action_error
)
