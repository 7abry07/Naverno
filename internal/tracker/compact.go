package tracker

import (
	"fmt"
	"net/netip"
)

func ParseV4CompactPeers(peers string) ([]netip.AddrPort, bool) {
	peerList := []netip.AddrPort{}

	for {
		if len(peers) == 0 {
			break
		}

		ip := peers[0:4]
		port := peers[4:6]

		parsedIp, err := netip.ParseAddr(fmt.Sprintf("%v.%v.%v.%v", ip[0], ip[1], ip[2], ip[3]))
		if err != nil {
			return []netip.AddrPort{}, false
		}

		peerVal := netip.AddrPortFrom(parsedIp, uint16(port[1])|uint16(port[0])<<8)
		peerList = append(peerList, peerVal)

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

		ip := peers[0:16]
		port := peers[16:18]

		parsedIp, err := netip.ParseAddr(fmt.Sprintf("%x:%x:%x:%x:%x:%x:%x:%x",
			uint16(ip[1])|uint16(ip[0])<<8,
			uint16(ip[3])|uint16(ip[2])<<8,
			uint16(ip[5])|uint16(ip[4])<<8,
			uint16(ip[7])|uint16(ip[6])<<8,
			uint16(ip[9])|uint16(ip[8])<<8, uint16(ip[11])|uint16(ip[10])<<8,
			uint16(ip[13])|uint16(ip[12])<<8,
			uint16(ip[15])|uint16(ip[14])<<8))

		if err != nil {
			return []netip.AddrPort{}, false
		}

		peerVal := netip.AddrPortFrom(parsedIp, uint16(port[1])|uint16(port[0])<<8)
		peerList = append(peerList, peerVal)

		peers = peers[18:]
	}
	return peerList, true
}
