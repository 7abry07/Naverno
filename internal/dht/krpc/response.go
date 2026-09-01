package krpc

// type PingResponse struct {
// 	ID [20]byte
// }
//
// type FindNodeResponse struct {
// 	Nodes []NodeInfo
// }
//
// type GetPeersResponse struct {
// 	Peers []netip.AddrPort
// 	Nodes []NodeInfo
// 	Token [20]byte
// }
//
// type AnnouncePeerResponse struct {
// 	ID [20]byte
// }

// func (r FindNodeResponse) Marshal(transactionID string) []byte {
// 	nodes := []byte{}
// 	for _, node := range r.Nodes {
// 		nodes = append(nodes, node.ID[:]...)
// 		nodes, _ = node.Addrport.AppendBinary(nodes)
// 	}
//
// 	res := response{}
// 	res.T = transactionID
// 	res.Y = "r"
// 	res.R = map[string]any{
// 		"nodes": nodes,
// 	}
//
// 	encoded, err := bencode.EncodeBytes(&res)
// 	if err != nil {
// 		return []byte{}
// 	}
// 	return encoded
// }
// func (r GetPeersResponse) Marshal(transactionID string) []byte {
// 	peers := [][]byte{}
// 	nodes := []byte{}
//
// 	for _, peer := range r.Peers {
// 		marshaled, _ := peer.MarshalBinary()
// 		peers = append(peers, marshaled)
// 	}
//
// 	for _, node := range r.Nodes {
// 		nodes = append(nodes, node.ID[:]...)
// 		nodes, _ = node.Addrport.AppendBinary(nodes)
// 	}
//
// 	res := response{}
// 	res.T = transactionID
// 	res.Y = "r"
//
// 	if len(peers) > 0 {
// 		res.R = map[string]any{"values": peers}
// 	} else if len(nodes) > 0 {
// 		res.R = map[string]any{"nodes": nodes}
// 	} else {
// 		res.R = map[string]any{}
// 	}
// 	res.R["token"] = r.Token
// 	encoded, err := bencode.EncodeBytes(&res)
// 	if err != nil {
// 		return []byte{}
// 	}
// 	return encoded
// }
//
// func (r AnnouncePeerResponse) Marshal(transactionID string) []byte {
// 	res := response{}
// 	res.T = transactionID
// 	res.Y = "r"
// 	res.R = map[string]any{"id": r.ID}
// 	encoded, err := bencode.EncodeBytes(&res)
// 	if err != nil {
// 		return []byte{}
// 	}
// 	return encoded
// }
