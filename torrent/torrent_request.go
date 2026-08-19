package torrent

import (
	"Naverno/internal/peerprotocol"
	"Naverno/internal/storage"
	"fmt"
)

func (t *Torrent) handleReadResult(res storage.ReadResult) {
	if res.Err != nil {
		t.err = fmt.Errorf("error while reading -> %v", res.Err)
		t.logger.Error("torrent -> error while reading", "Error", res.Err)
		return
	}
	request := peerprotocol.Request{Idx: res.Piece.Idx, Begin: res.Begin, Length: uint32(len(res.Data))}
	requester, ok := t.pendingRequests[request]
	if !ok {
		return
	}
	delete(t.pendingRequests, request)
	p, ok := t.peers[requester]
	if !ok {
		return
	}
	p.Piece(request.Idx, request.Begin, res.Data)
	p.UpdateStats(0, uint64(len(res.Data)))
	t.logger.Error("torrent -> request data sent", "Request", fmt.Sprintf("(%v, %v, %v)", request.Idx, request.Begin, len(res.Data)))
}
