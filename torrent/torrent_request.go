package torrent

import (
	"Naverno/internal/requesthandler"
	"fmt"
)

func (t *Torrent) handleRequestResult(res *requesthandler.RequestHandler) {
	if res.Err != nil {
		t.logger.Error("torrent -> error in request handling", "Error", res.Err)
		t.session.RemoveTorrent(t)
		return
	}
	_, ok := t.requestHandlers[res.Request]
	if !ok {
		return
	}
	delete(t.requestHandlers, res.Request)
	p, ok := t.peers[res.Requester]
	if !ok {
		return
	}
	p.Piece(res.Request.Idx, res.Request.Begin, res.Data)
	p.UpdateStats(0, uint64(len(res.Data)))
	t.logger.Error("torrent -> request data sent", "Request", fmt.Sprintf("(%v, %v, %v)", res.Request.Idx, res.Request.Begin, len(res.Data)))
}
