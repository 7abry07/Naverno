package tracker

import (
	"Naverno/internal/bencode"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
)

// --------------- Structs -------------------

type httpTracker struct {
	announce *url.URL
}

// --------------- Functions -------------------

func parseV4BencodedPeers(peers bencode.BList) ([]peer, bool) {
	peerList := []peer{}

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

		peerVal := peer{}

		peerVal.Ip = parsedIp
		peerVal.Port = uint16(port)
		peerList = append(peerList, peerVal)
	}

	return peerList, true
}

func parseV4CompactPeers(peers string) ([]peer, bool) {
	peerList := []peer{}

	for {
		if len(peers) == 0 {
			break
		}

		ip := peers[0:4]
		port := peers[4:6]

		parsedIp, err := netip.ParseAddr(fmt.Sprintf("%v.%v.%v.%v", ip[0], ip[1], ip[2], ip[3]))
		if err != nil {
			return []peer{}, false
		}

		peerVal := peer{}
		peerVal.Ip = parsedIp
		peerVal.Port = uint16(port[1]) | uint16(port[0])<<8
		peerList = append(peerList, peerVal)

		peers = peers[6:]
	}
	return peerList, true
}

func parseV6CompactPeers(peers string) ([]peer, bool) {
	peerList := []peer{}

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
			return []peer{}, false
		}

		peerVal := peer{}
		peerVal.Ip = parsedIp
		peerVal.Port = uint16(port[1]) | uint16(port[0])<<8
		peerList = append(peerList, peerVal)

		peers = peers[18:]
	}
	return peerList, true
}

// --------------- Methods -------------------

func (t httpTracker) Announce(ctx context.Context, r AnnounceRequest) (AnnounceResponse, error) {
	urlReq := t.serialize(r)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", urlReq.String(), nil)
	if err != nil {
		return AnnounceResponse{}, errors.Join(InvalidRespErr, err)
	}

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return AnnounceResponse{}, errors.Join(InvalidRespErr, err)
	}

	defer httpResp.Body.Close()
	httpBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return AnnounceResponse{}, errors.Join(InvalidRespErr, err)
	}

	resp, ok := t.deserialize(httpBody)
	if !ok {
		return AnnounceResponse{}, InvalidRespErr
	}

	return resp, nil
}

func (t httpTracker) serialize(r AnnounceRequest) *url.URL {
	fullUrl := t.announce
	query := fmt.Sprintf(
		"info_hash=%v"+
			"&peer_id=%v"+
			"&port=%v"+
			"&uploaded=%v"+
			"&downloaded=%v"+
			"&left=%v"+
			"&compact=1"+
			"&no_peer_id=1"+
			"&event=%v"+
			"&numwant=200"+
			"&ip=%v"+
			"&key=%v",
		url.QueryEscape(string(r.Infohash[:])),
		url.QueryEscape(string(r.PeerID[:])),
		url.QueryEscape(strconv.Itoa(int(r.Port))),
		url.QueryEscape(strconv.Itoa(int(r.Uploaded))),
		url.QueryEscape(strconv.Itoa(int(r.Downloaded))),
		url.QueryEscape(strconv.Itoa(int(r.Left))),
		url.QueryEscape(trackerEventName[r.Event]),
		url.QueryEscape(r.Ip.String()),
		url.QueryEscape(strconv.Itoa(int(r.Key))),
	)
	fullUrl.RawQuery = query

	return fullUrl
}

func (t httpTracker) deserialize(httpResp []byte) (AnnounceResponse, bool) {
	r := AnnounceResponse{}

	decoded, err := bencode.Decode(string(httpResp))
	if err != nil {
		return r, false
	}
	root, ok := decoded.Dict()
	if !ok {
		return r, false
	}

	interval, _ := root.FindIntOrDef("interval", 1800)
	minInterval, _ := root.FindIntOrDef("min interval", 30)
	r.Interval = uint32(interval)
	r.MinInterval = uint32(minInterval)

	warning, warningOk := root.FindStr("warning reason")
	if warningOk {
		str := string(warning)
		r.Warning = &str
	}

	failure, failureOk := root.FindStr("failure reason")
	if failureOk {
		str := string(failure)
		r.Failure = &str
		return r, true
	}

	complete, _ := root.FindIntOrDef("complete", -1)
	incomplete, _ := root.FindIntOrDef("incomplete", -1)
	downloaded, _ := root.FindIntOrDef("downloaded", -1)
	r.Complete = int64(complete)
	r.Incomplete = int64(incomplete)
	r.Downloaded = int64(downloaded)

	peersNode, ok := root.Find("peers")
	if !ok {
		return r, false
	}

	if peersNode.Type() == bencode.List_t {
		peersList, _ := peersNode.List()
		peers, ok := parseV4BencodedPeers(peersList)
		if !ok {
			return r, false
		}
		r.PeerList = append(r.PeerList, peers...)
	} else if peersNode.Type() == bencode.Str_t {
		peersStr, _ := peersNode.Str()
		peers, ok := parseV4CompactPeers(string(peersStr))
		if !ok {
			return r, false
		}
		r.PeerList = append(r.PeerList, peers...)
	}

	peers6Node, ok := root.Find("peers6")
	if ok {
		if peers6Node.Type() == bencode.Str_t {
			peersStr, _ := peers6Node.Str()
			peers, ok := parseV6CompactPeers(string(peersStr))
			if ok {
				r.PeerList = append(r.PeerList, peers...)
			}
		}
	}

	return r, true
}
