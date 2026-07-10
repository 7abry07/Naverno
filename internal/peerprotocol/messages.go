package peerprotocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Message interface {
	Marshal() []byte
}

type KeepAlive struct{}

type Choke struct{}
type Unchoke struct{}
type Interested struct{}
type Uninterested struct{}

type Have struct {
	Idx uint32
}

type Bitfield struct {
	Pieces []byte
}

type Request struct {
	Idx    uint32
	Begin  uint32
	Length uint32
}

type Piece struct {
	Idx   uint32
	Begin uint32
	Data  []byte
}

type Cancel struct {
	Idx    uint32
	Begin  uint32
	Length uint32
}

func (KeepAlive) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, 0)
	return marshaled
}

func (Choke) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, 1)
	marshaled = append(marshaled, 0)
	return marshaled
}

func (Unchoke) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, 1)
	marshaled = append(marshaled, 1)
	return marshaled
}

func (Interested) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, 1)
	marshaled = append(marshaled, 2)
	return marshaled

}
func (Uninterested) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, 1)
	marshaled = append(marshaled, 3)
	return marshaled
}
func (m Have) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, 5)
	marshaled = append(marshaled, 4)
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Idx)
	return marshaled
}

func (m Bitfield) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, uint32(1+len(m.Pieces)))
	marshaled = append(marshaled, 5)
	marshaled = append(marshaled, m.Pieces...)
	return marshaled
}

func (m Request) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, 13)
	marshaled = append(marshaled, 6)
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Idx)
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Begin)
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Length)
	return marshaled
}
func (m Piece) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, uint32(9+len(m.Data)))
	marshaled = append(marshaled, 7)
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Idx)
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Begin)
	marshaled = append(marshaled, m.Data...)
	return marshaled
}

func (m Cancel) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, 13)
	marshaled = append(marshaled, 8)
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Idx)
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Begin)
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Length)
	return marshaled
}

func Decode(data []byte) (Message, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid message")
	}
	if len(data) == 4 {
		if !bytes.Equal(data, []byte{0, 0, 0, 0}) {
			return nil, fmt.Errorf("invalid keepalive message")
		}
		return KeepAlive{}, nil
	}

	length := binary.BigEndian.Uint32(data[0:4])
	id := data[4:5]

	switch id[0] {
	case ChokeID:
		if length != 1 {
			return nil, fmt.Errorf("invalid choke message")
		}
		return Choke{}, nil
	case UnchokeID:
		if length != 1 {
			return nil, fmt.Errorf("invalid unchoke message")
		}
		return Unchoke{}, nil
	case InterestedID:
		if length != 1 {
			return nil, fmt.Errorf("invalid interested message")
		}
		return Interested{}, nil
	case UninterestedID:
		if length != 1 {
			return nil, fmt.Errorf("invalid uninterested message")
		}
		return Uninterested{}, nil
	case HaveID:
		if length != 5 {
			return nil, fmt.Errorf("invalid have message")
		}
		payload := data[5:]
		idx := binary.BigEndian.Uint32(payload)
		return Have{idx}, nil
	case BitfieldID:
		return Bitfield{data[5:]}, nil
	case RequestID:
		if length != 13 {
			return nil, fmt.Errorf("invalid request message")
		}
		payload := data[5:]
		idx := binary.BigEndian.Uint32(payload[:4])
		begin := binary.BigEndian.Uint32(payload[4:8])
		length := binary.BigEndian.Uint32(payload[8:12])
		return Request{idx, begin, length}, nil
	case PieceID:
		if length < 9 {
			return nil, fmt.Errorf("invalid piece message")
		}
		payload := data[5:]
		idx := binary.BigEndian.Uint32(payload[:4])
		begin := binary.BigEndian.Uint32(payload[4:8])
		block := payload[8:]
		return Piece{idx, begin, block}, nil

	case CancelID:
		if length != 13 {
			return nil, fmt.Errorf("invalid cancel message")
		}
		payload := data[5:]
		idx := binary.BigEndian.Uint32(payload[:4])
		begin := binary.BigEndian.Uint32(payload[4:8])
		length := binary.BigEndian.Uint32(payload[8:12])
		return Cancel{idx, begin, length}, nil
	default:
		return nil, fmt.Errorf("invalid or non supported message id")
	}
}
