package udptracker

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/netip"
)

type announceRequest struct {
	connectionID uint64
	infohash     [20]byte
	peerID       [20]byte
	downloaded   uint64
	left         uint64
	uploaded     uint64
	event        uint32
	ip           uint32
	key          uint32
	numwant      uint32
	port         uint16
}

type announceResponse struct {
	interval uint32
	leechers uint32
	seeders  uint32
	peers    []netip.AddrPort
}

func (req announceRequest) Encode(transactionID uint32) []byte {
	encoded := []byte{}
	encoded = binary.BigEndian.AppendUint64(encoded, req.connectionID)
	encoded = binary.BigEndian.AppendUint32(encoded, uint32(action_announce))
	encoded = binary.BigEndian.AppendUint32(encoded, transactionID)
	encoded = append(encoded, req.infohash[:]...)
	encoded = append(encoded, req.peerID[:]...)
	encoded = binary.BigEndian.AppendUint64(encoded, req.downloaded)
	encoded = binary.BigEndian.AppendUint64(encoded, req.left)
	encoded = binary.BigEndian.AppendUint64(encoded, req.uploaded)
	encoded = binary.BigEndian.AppendUint32(encoded, req.event)
	encoded = binary.BigEndian.AppendUint32(encoded, req.ip)
	encoded = binary.BigEndian.AppendUint32(encoded, req.key)
	encoded = binary.BigEndian.AppendUint32(encoded, req.numwant)
	encoded = binary.BigEndian.AppendUint16(encoded, req.port)
	return encoded
}

func (res *announceResponse) decode(data *bytes.Buffer) error {
	if data.Len() < 12 {
		return fmt.Errorf("invalid length for announce response")
	}

	interval := make([]byte, 4)
	leechers := make([]byte, 4)
	seeders := make([]byte, 4)

	_, err := data.Read(interval)
	if err != nil {
		return fmt.Errorf("error while reading announce response")
	}
	_, err = data.Read(leechers)
	if err != nil {
		return fmt.Errorf("error while reading announce response")
	}
	_, err = data.Read(seeders)
	if err != nil {
		return fmt.Errorf("error while reading announce response")
	}

	peers := []netip.AddrPort{}
	for data.Len() != 0 {
		addr := make([]byte, 6)
		_, err := data.Read(addr)
		if err != nil {
			return fmt.Errorf("error while reading peers in announce response")
		}
		ipArr := [4]byte{}
		copy(ipArr[:], addr[:4])

		ip := netip.AddrFrom4(ipArr)
		parsed := netip.AddrPortFrom(ip, binary.BigEndian.Uint16(addr[4:6]))
		peers = append(peers, parsed)
	}

	res.interval = binary.BigEndian.Uint32(interval)
	res.leechers = binary.BigEndian.Uint32(leechers)
	res.seeders = binary.BigEndian.Uint32(seeders)
	res.peers = peers
	return nil
}

func (announceResponse) Action() action {
	return action_announce
}
