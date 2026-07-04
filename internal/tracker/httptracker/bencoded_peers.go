package httptracker

import (
	"net/netip"

	"github.com/zeebo/bencode"
)

type peer struct {
	Ip   string `bencode:"ip"`
	Port uint16 `bencode:"port"`
}

func ParseBencodedPeers(in []byte) ([]netip.AddrPort, bool) {
	peers := []netip.AddrPort{}
	peerList := []peer{}
	bencode.DecodeBytes(in, &peerList)

	for _, p := range peerList {
		ip, err := netip.ParseAddr(p.Ip)
		if err != nil {
			return []netip.AddrPort{}, false
		}
		peers = append(peers, netip.AddrPortFrom(ip, p.Port))
	}

	return peers, true
}
