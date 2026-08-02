package udptracker

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/netip"
)

type response interface {
	Action() action
}

type connectResponse struct {
	transactionID uint32
	connectionID  uint64
}

func (connectResponse) Action() action {
	return action_connect
}

type announceResponse struct {
	transactionID uint32
	interval      uint32
	leechers      uint32
	seeders       uint32
	peers         []netip.AddrPort
}

func (announceResponse) Action() action {
	return action_announce
}

type errorResponse struct {
	transactionID uint32
	message       string
}

func (errorResponse) Action() action {
	return action_error
}

func decodeResponse(res *bytes.Buffer) (response, error) {
	a := make([]byte, 4)
	_, err := res.Read(a)
	if err != nil {
		return nil, fmt.Errorf("error while reading response")
	}

	switch binary.BigEndian.Uint32(a) {
	case uint32(action_connect):
		connect := connectResponse{}
		if res.Len() != 12 {
			return nil, fmt.Errorf("invalid length for connect response")
		}

		transactionId := make([]byte, 4)
		connectionId := make([]byte, 8)

		_, err := res.Read(transactionId)
		if err != nil {
			return nil, fmt.Errorf("error while reading connect response")
		}
		_, err = res.Read(connectionId)
		if err != nil {
			return nil, fmt.Errorf("error while reading connect response")
		}

		connect.transactionID = binary.BigEndian.Uint32(transactionId)
		connect.connectionID = binary.BigEndian.Uint64(connectionId)
		return connect, nil
	case uint32(action_announce):
		announce := announceResponse{}
		if res.Len() < 20 {
			return nil, fmt.Errorf("invalid length for announce response")
		}

		transactionId := make([]byte, 4)
		interval := make([]byte, 4)
		leechers := make([]byte, 4)
		seeders := make([]byte, 4)

		_, err := res.Read(transactionId)
		if err != nil {
			return nil, fmt.Errorf("error while reading announce response")
		}
		_, err = res.Read(interval)
		if err != nil {
			return nil, fmt.Errorf("error while reading announce response")
		}
		_, err = res.Read(leechers)
		if err != nil {
			return nil, fmt.Errorf("error while reading announce response")
		}
		_, err = res.Read(seeders)
		if err != nil {
			return nil, fmt.Errorf("error while reading announce response")
		}
		peers, ok := decodePeers(res)
		if !ok {
			return nil, fmt.Errorf("error while decoding peers announce response")
		}

		announce.transactionID = binary.BigEndian.Uint32(transactionId)
		announce.interval = binary.BigEndian.Uint32(interval)
		announce.leechers = binary.BigEndian.Uint32(leechers)
		announce.seeders = binary.BigEndian.Uint32(seeders)
		announce.peers = peers

		return announce, nil
	case uint32(action_error):
		e := errorResponse{}
		if res.Len() < 4 {
			return nil, fmt.Errorf("invalid length for error response")
		}

		transactionId := make([]byte, 4)
		message := []byte{}

		_, err := res.Read(transactionId)
		if err != nil {
			return nil, fmt.Errorf("error while reading error response")
		}

		message, err = io.ReadAll(res)
		if err != nil {
			return nil, fmt.Errorf("error while reading error response")
		}
		e.transactionID = binary.BigEndian.Uint32(transactionId)
		e.message = string(message)
		return e, nil
	}

	return nil, fmt.Errorf("unrecognized action in response")
}

func decodePeers(peers *bytes.Buffer) ([]netip.AddrPort, bool) {
	result := []netip.AddrPort{}
	for peers.Len() != 0 {
		addr := make([]byte, 6)
		_, err := peers.Read(addr)
		if err != nil {
			return nil, false
		}
		ipArr := [4]byte{}
		copy(ipArr[:], addr[:4])

		ip := netip.AddrFrom4(ipArr)
		parsed := netip.AddrPortFrom(ip, binary.BigEndian.Uint16(addr[4:6]))
		result = append(result, parsed)
	}
	return result, true
}
