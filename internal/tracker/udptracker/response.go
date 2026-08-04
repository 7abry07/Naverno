package udptracker

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func getResponseInfo(res *bytes.Buffer) (action, uint32, error) {
	if res.Len() < 8 {
		return 0, 0, fmt.Errorf("invalid length")
	}
	a := make([]byte, 4)
	tid := make([]byte, 4)

	_, err := res.Read(a)
	if err != nil {
		return 0, 0, fmt.Errorf("error while reading response")
	}
	_, err = res.Read(tid)
	if err != nil {
		return 0, 0, fmt.Errorf("error while reading response")
	}

	switch binary.BigEndian.Uint32(a) {
	case uint32(action_connect):
		return action_connect, binary.BigEndian.Uint32(tid), nil
	case uint32(action_announce):
		return action_announce, binary.BigEndian.Uint32(tid), nil
	case uint32(action_error):
		return action_error, binary.BigEndian.Uint32(tid), nil
	default:
		return 0, 0, fmt.Errorf("unrecognized action")
	}
}
