package handshaker

import "net"

type HandshakedConn struct {
	net.Conn
	PeerID     [20]byte
	Extensions [8]byte
	Error      error
}
