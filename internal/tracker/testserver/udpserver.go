package testserver

import (
	"Naverno/internal/tracker"
	"Naverno/internal/tracker/udptracker"
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"net"
	"net/netip"
	"sync"
)

type action uint32

const (
	action_connect action = iota
	action_announce
	action_error = 3
)

type UDPServer struct {
	listener *net.UDPConn
	connIDs  map[string]uint64
	store    map[[20]byte][]tracker.CompactPeer

	connsIDMut sync.Mutex
	storeMut   sync.Mutex
}

func StartUDPServer() func() {
	s := UDPServer{
		connIDs: make(map[string]uint64),
		store:   make(map[[20]byte][]tracker.CompactPeer),
	}

	addr, _ := net.ResolveUDPAddr("udp", ":8001")
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil
	}
	s.listener = conn
	go s.listen()
	return s.close
}

func (s *UDPServer) close() {
	s.listener.Close()
}

func (s *UDPServer) listen() {
	for {
		buf := make([]byte, 65535)
		read, remote, err := s.listener.ReadFromUDP(buf)
		if err != nil {
			return
		}

		data := bytes.NewBuffer(buf[:read])
		go s.handleRequest(remote, data)
	}
}

func (s *UDPServer) handleRequest(remote *net.UDPAddr, data *bytes.Buffer) {
	if data.Len() < 16 {
		return
	}

	firstWordSlice := make([]byte, 8)
	aSlice := make([]byte, 4)
	tidSlice := make([]byte, 4)
	data.Read(firstWordSlice)
	data.Read(aSlice)
	data.Read(tidSlice)
	firstWord := binary.BigEndian.Uint64(firstWordSlice)
	a := binary.BigEndian.Uint32(aSlice)
	tid := binary.BigEndian.Uint32(tidSlice)

	switch a {
	case uint32(action_connect):
		if firstWord != udptracker.PROTOCOL_ID {
			s.sendFailure(
				remote,
				tid,
				fmt.Errorf("first 8 bytes arent equal to protocol_id (%v) in connect request", udptracker.PROTOCOL_ID))
			return
		}
		connectionID := rand.Uint64()
		res := []byte{}
		res = append(res, aSlice...)
		res = append(res, tidSlice...)
		res = binary.BigEndian.AppendUint64(res, connectionID)
		s.listener.WriteToUDP(res, remote)
		s.connsIDMut.Lock()
		defer s.connsIDMut.Unlock()
		s.connIDs[remote.String()] = connectionID
	case uint32(action_announce):
		connId, ok := s.connIDs[remote.String()]
		if !ok {
			s.sendFailure(remote, tid, fmt.Errorf("need to request a connection id before announcing"))
			return
		}
		if connId != firstWord {
			s.sendFailure(remote, tid, fmt.Errorf("connection id is not what was expected"))
			return
		}
		req, err := parseAnnounceRequest(data)
		if err != nil {
			s.sendFailure(remote, tid, err)
			return
		}
		port := req.Port
		ip := remote.String()
		if (req.Ip != netip.Addr{}) {
			ip = req.Ip.String()
		}
		thisPeer, _ := tracker.NewCompactPeer(ip, port)

		peers, ok := s.store[req.InfoHash]
		if !ok {
			s.store[req.InfoHash] = []tracker.CompactPeer{}
		}

		if len(peers) > int(min(maxNumwant, req.Numwant)) {
			if req.Numwant == -1 {
				peers = peers[0:defNumwant]
			} else {
				peers = peers[0:min(maxNumwant, req.Numwant)]
			}
		}

		switch req.Event {
		case "":
		case "started":
			s.store[req.InfoHash] = append(s.store[req.InfoHash], thisPeer)
		case "stopped":
			out := []tracker.CompactPeer{}

			for _, p := range peers {
				if p != thisPeer {
					out = append(out, p)
				}
			}

			s.store[req.InfoHash] = out
			s.sendAnnounceResponse(remote, tid,
				announceResponse{
					Interval: announceInterval,
					Peers:    []byte{},
					Peers6:   []byte{},
				})

			return
		}

		compactPeers := []byte{}
		compactPeers6 := []byte{}

		for _, peer := range peers {
			if peer == thisPeer {
				continue
			}

			marshaled, err := peer.MarshalBinary()
			if err != nil {
				panic(err)
			}

			if peer.Ip.Is4() {
				compactPeers = append(compactPeers, marshaled...)
			} else {
				compactPeers6 = append(compactPeers6, marshaled...)
			}
		}

		s.sendAnnounceResponse(remote, tid,
			announceResponse{
				Interval: announceInterval,
				Peers:    compactPeers,
				Peers6:   compactPeers6,
			})
	}
}

func (s *UDPServer) sendFailure(remote *net.UDPAddr, tid uint32, err error) {
	data := []byte{}
	data = binary.BigEndian.AppendUint32(data, uint32(action_error))
	data = binary.BigEndian.AppendUint32(data, tid)
	data = append(data, []byte(err.Error())...)
	s.listener.WriteToUDP(data, remote)
}

func (s *UDPServer) sendAnnounceResponse(remote *net.UDPAddr, tid uint32, res announceResponse) {
	data := []byte{}
	data = binary.BigEndian.AppendUint32(data, uint32(action_announce))
	data = binary.BigEndian.AppendUint32(data, tid)
	data = binary.BigEndian.AppendUint32(data, uint32(res.Interval))
	data = binary.BigEndian.AppendUint32(data, 0)
	data = binary.BigEndian.AppendUint32(data, 0)
	data = append(data, res.Peers...)
	s.listener.WriteToUDP(data, remote)
}

func parseAnnounceRequest(buf *bytes.Buffer) (announceRequest, error) {
	if buf.Len() < 80 {
		return announceRequest{}, fmt.Errorf("invalid length in announce request")
	}
	infohash := make([]byte, 20)
	peerid := make([]byte, 20)
	event := make([]byte, 4)
	ip := make([]byte, 4)
	numwant := make([]byte, 4)
	port := make([]byte, 4)

	buf.Read(infohash)
	buf.Read(peerid)
	buf.Read(make([]byte, 24))
	buf.Read(event)
	buf.Read(ip)
	buf.Read(make([]byte, 4))
	buf.Read(numwant)
	buf.Read(port)

	ih := [20]byte{}
	pid := [20]byte{}
	ipArr := [4]byte{}
	copy(ih[:], infohash)
	copy(pid[:], peerid)
	copy(ipArr[:], ip)
	return announceRequest{
		InfoHash: ih,
		PeerID:   pid,
		Port:     binary.BigEndian.Uint16(port),
		Ip:       netip.AddrFrom4(ipArr),
		Event:    tracker.TrackerEvent.String(tracker.TrackerEvent(binary.BigEndian.Uint32(event))),
		Numwant:  int64(binary.BigEndian.Uint32(numwant)),
	}, nil
}
