package udptracker

import "encoding/binary"

const (
	PROTOCOL_ID = uint64(0x41727101980)
)

type connectRequest struct {
	transactionID uint32
}

type announceRequest struct {
	connectionID  uint64
	transactionID uint32
	infohash      [20]byte
	peerID        [20]byte
	downloaded    uint64
	left          uint64
	uploaded      uint64
	event         uint32
	ip            uint32
	key           uint32
	numwant       uint32
	port          uint16
}

func (req connectRequest) encode() []byte {
	encoded := []byte{}
	encoded = binary.BigEndian.AppendUint64(encoded, PROTOCOL_ID)
	encoded = binary.BigEndian.AppendUint32(encoded, uint32(action_connect))
	encoded = binary.BigEndian.AppendUint32(encoded, req.transactionID)
	return encoded
}

func (req announceRequest) encode() []byte {
	encoded := []byte{}
	encoded = binary.BigEndian.AppendUint64(encoded, req.connectionID)
	encoded = binary.BigEndian.AppendUint32(encoded, uint32(action_announce))
	encoded = binary.BigEndian.AppendUint32(encoded, req.transactionID)
	encoded = append(encoded, req.infohash[:]...)
	encoded = append(encoded, req.peerID[:]...)
	encoded = binary.BigEndian.AppendUint64(encoded, req.downloaded)
	encoded = binary.BigEndian.AppendUint64(encoded, req.left)
	encoded = binary.BigEndian.AppendUint64(encoded, req.uploaded)
	encoded = binary.BigEndian.AppendUint32(encoded, req.event)
	encoded = binary.BigEndian.AppendUint32(encoded, req.ip)
	encoded = binary.BigEndian.AppendUint32(encoded, req.key)
	encoded = binary.BigEndian.AppendUint32(encoded, req.numwant)
	encoded = binary.BigEndian.AppendUint16(encoded, req.port)
	return encoded
}
