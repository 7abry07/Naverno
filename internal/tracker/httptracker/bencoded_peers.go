package httptracker

import (
	"Naverno/internal/bencode"
	"Naverno/internal/tracker"
	"net/netip"
)

func ParseBencodedPeers(peers bencode.BList) ([]tracker.Peer, bool) {
	peerList := []tracker.Peer{}

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

		peerVal := tracker.Peer{}

		peerVal.Ip = parsedIp
		peerVal.Port = uint16(port)
		peerList = append(peerList, peerVal)
	}

	return peerList, true
}
