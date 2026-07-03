package httptracker

import (
	"net/netip"
)

func ParseBencodedPeers(peers []any) ([]netip.AddrPort, bool) {
	peerList := []netip.AddrPort{}

	for _, peerNode := range peers {
		p, ok := peerNode.(map[string]any)
		if !ok {
			return peerList, false
		}

		ip, ok := p["ip"].(string)
		port, ok1 := p["port"].(int64)

		if !ok || !ok1 {
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
