package krpc

// type PingQuery struct {
// 	ID [20]byte
// }
//
// type FindNodeQuery struct {
// 	ID     [20]byte
// 	Target [20]byte
// }
//
// type GetPeersQuery struct {
// 	ID       [20]byte
// 	Infohash [20]byte
// }
//
// type AnnouncePeerQuery struct {
// 	ID       [20]byte
// 	Infohash [20]byte
// 	Token    [20]byte
// 	Port     uint16
// }
//
// func (q PingQuery) Marshal(transactionID string) []byte {
// 	qu := query{}
// 	qu.T = transactionID
// 	qu.Y = "q"
// 	qu.Q = "ping"
// 	qu.A = map[string]any{
// 		"id": q.ID,
// 	}
// 	encoded, err := bencode.EncodeBytes(&qu)
// 	if err != nil {
// 		return []byte{}
// 	}
// 	return encoded
// }
//
// func (q FindNodeQuery) Marshal(transactionID string) []byte {
// 	qu := query{}
// 	qu.T = transactionID
// 	qu.Y = "q"
// 	qu.Q = "find_node"
// 	qu.A = map[string]any{
// 		"id":     q.ID,
// 		"target": q.Target,
// 	}
// 	encoded, err := bencode.EncodeBytes(&qu)
// 	if err != nil {
// 		return []byte{}
// 	}
// 	return encoded
// }
//
// func (q GetPeersQuery) Marshal(transactionID string) []byte {
// 	qu := query{}
// 	qu.T = transactionID
// 	qu.Y = "q"
// 	qu.Q = "get_peers"
// 	qu.A = map[string]any{
// 		"id":        q.ID,
// 		"info_hash": q.Infohash,
// 	}
// 	encoded, err := bencode.EncodeBytes(&qu)
// 	if err != nil {
// 		return []byte{}
// 	}
// 	return encoded
// }
//
// func (q AnnouncePeerQuery) Marshal(transactionID string) []byte {
// 	qu := query{}
// 	qu.T = transactionID
// 	qu.Y = "q"
// 	qu.Q = "get_peers"
// 	qu.A = map[string]any{
// 		"id":        q.ID,
// 		"info_hash": q.Infohash,
// 		"port":      q.Port,
// 		"token":     q.Token,
// 	}
// 	encoded, err := bencode.EncodeBytes(&qu)
// 	if err != nil {
// 		return []byte{}
// 	}
// 	return encoded
// }
