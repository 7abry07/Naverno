package httptracker

import (
	"Naverno/internal/bencode"
	"net/netip"
)

func ParseBencodedPeers(peers bencode.BList) ([]netip.AddrPort, bool) {
	peerList := []netip.AddrPort{}

	for _, peerNode := range peers {
		p, ok := peerNode.Dict()
		if !ok {
			return peerList, false
		}

		ip, _ := p.FindStrOrDef("ip", "")
		port, _ := p.FindIntOrDef("port", 0)

		if ip == "" || port == 0 {
			continue
		}
		parsedIp, err := netip.ParseAddr(string(ip))
		if err != nil {
			return peerList, false
		}

		peerVal := netip.AddrPortFrom(parsedIp, uint16(port))
		peerList = append(peerList, peerVal)
	}

	return peerList, true
}
