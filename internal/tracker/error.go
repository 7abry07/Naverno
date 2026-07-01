package tracker

import "errors"

var (
	InvalidUrlErr    = errors.New("the tracker url is invalid")
	InvalidSchemeErr = errors.New("the tracker url scheme is invalid")
	InvalidRespErr   = errors.New("the tracker response is invalid")
)
