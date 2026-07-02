package tracker

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net/netip"
)

type CompactPeer struct {
	Ip   netip.Addr
	Port uint16
}

func NewCompactPeer(ip string, port uint16) (CompactPeer, error) {
	if ip == "" {
		return CompactPeer{netip.Addr{}, port}, nil
	}

	parsed, err := netip.ParseAddr(ip)
	if err != nil {
		return CompactPeer{}, err
	}
	return CompactPeer{parsed, port}, nil
}

func (p *CompactPeer) MarshalBinary() ([]byte, error) {
	marshaled, err := p.Ip.MarshalBinary()
	if err != nil {
		return []byte{}, err
	}
	port := make([]byte, 2)
	binary.BigEndian.PutUint16(port, p.Port)
	marshaled = append(marshaled, port...)

	return marshaled, nil
}

func (p *CompactPeer) UnmarshalBinary(input []byte) error {
	if len(input) == 6 {
		ip := input[0:4]
		port := input[4:6]

		parsedIp, err := netip.ParseAddr(fmt.Sprintf("%v.%v.%v.%v", ip[0], ip[1], ip[2], ip[3]))
		if err != nil {
			return err
		}
		p.Ip = parsedIp
		p.Port = uint16(port[1]) | uint16(port[0])<<8
	} else if len(input) == 18 {
		ip := input[0:16]
		port := input[16:18]

		parsedIp, err := netip.ParseAddr(fmt.Sprintf("%x:%x:%x:%x:%x:%x:%x:%x",
			uint16(ip[1])|uint16(ip[0])<<8,
			uint16(ip[3])|uint16(ip[2])<<8,
			uint16(ip[5])|uint16(ip[4])<<8,
			uint16(ip[7])|uint16(ip[6])<<8,
			uint16(ip[9])|uint16(ip[8])<<8, uint16(ip[11])|uint16(ip[10])<<8,
			uint16(ip[13])|uint16(ip[12])<<8,
			uint16(ip[15])|uint16(ip[14])<<8))

		if err != nil {
			return err
		}

		p.Ip = parsedIp
		p.Port = uint16(port[1]) | uint16(port[0])<<8
	} else {
		return errors.New("invalid address")
	}
	return nil
}

func ParseV4CompactPeers(peers string) ([]netip.AddrPort, bool) {
	peerList := []netip.AddrPort{}

	for {
		if len(peers) == 0 {
			break
		}

		peer, err := NewCompactPeer("", 0)
		if err != nil {
			return []netip.AddrPort{}, false
		}

		err = peer.UnmarshalBinary([]byte(peers[0:6]))
		if err != nil {
			return []netip.AddrPort{}, false
		}

		peerList = append(peerList, netip.AddrPortFrom(peer.Ip, peer.Port))

		peers = peers[6:]
	}
	return peerList, true
}

func ParseV6CompactPeers(peers string) ([]netip.AddrPort, bool) {
	peerList := []netip.AddrPort{}

	for {
		if len(peers) == 0 {
			break
		}

		peer, err := NewCompactPeer("", 0)
		if err != nil {
			return []netip.AddrPort{}, false
		}

		err = peer.UnmarshalBinary([]byte(peers[0:18]))
		if err != nil {
			return []netip.AddrPort{}, false
		}

		peerList = append(peerList, netip.AddrPortFrom(peer.Ip, peer.Port))

		peers = peers[18:]
	}
	return peerList, true
}
