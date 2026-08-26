package peerprotocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Message interface {
	Marshal() []byte
	ID() MessageID
}

type KeepAlive struct{}
type Choke struct{}
type Unchoke struct{}
type Interested struct{}
type Uninterested struct{}
type Have struct{ Idx uint32 }
type Bitfield struct{ Pieces []byte }
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

type Extended struct {
	MessageID uint8
	ExtendedMessage
}

func (KeepAlive) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, 0)
	return marshaled
}

func (Choke) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, 1)
	marshaled = append(marshaled, byte(ChokeID))
	return marshaled
}

func (Unchoke) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, 1)
	marshaled = append(marshaled, byte(UnchokeID))
	return marshaled
}

func (Interested) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, 1)
	marshaled = append(marshaled, byte(InterestedID))
	return marshaled

}
func (Uninterested) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, 1)
	marshaled = append(marshaled, byte(UninterestedID))
	return marshaled
}
func (m Have) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, 5)
	marshaled = append(marshaled, byte(HaveID))
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Idx)
	return marshaled
}

func (m Bitfield) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, uint32(1+len(m.Pieces)))
	marshaled = append(marshaled, byte(BitfieldID))
	marshaled = append(marshaled, m.Pieces...)
	return marshaled
}

func (m Request) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, 13)
	marshaled = append(marshaled, byte(RequestID))
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Idx)
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Begin)
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Length)
	return marshaled
}
func (m Piece) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, uint32(9+len(m.Data)))
	marshaled = append(marshaled, byte(PieceID))
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Idx)
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Begin)
	marshaled = append(marshaled, m.Data...)
	return marshaled
}

func (m Cancel) Marshal() []byte {
	marshaled := []byte{}
	marshaled = binary.BigEndian.AppendUint32(marshaled, 13)
	marshaled = append(marshaled, byte(CancelID))
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Idx)
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Begin)
	marshaled = binary.BigEndian.AppendUint32(marshaled, m.Length)
	return marshaled
}

func (m Extended) Marshal() []byte {
	marshaled := []byte{}
	extendedMarshaled := m.ExtendedMessage.Marshal()
	marshaled = binary.BigEndian.AppendUint32(marshaled, uint32(2+len(extendedMarshaled)))
	marshaled = append(marshaled, byte(ExtendedID))
	marshaled = append(marshaled, byte(m.MessageID))
	marshaled = append(marshaled, extendedMarshaled...)
	// fmt.Printf("marshaled -> %v\n", marshaled)
	// fmt.Printf("marshaled -> %v\n", string(marshaled))

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
	data = data[5:]

	switch MessageID(id[0]) {
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
		idx := binary.BigEndian.Uint32(data)
		return Have{idx}, nil
	case BitfieldID:
		return Bitfield{data}, nil
	case RequestID:
		if length != 13 {
			return nil, fmt.Errorf("invalid request message")
		}
		idx := binary.BigEndian.Uint32(data[:4])
		begin := binary.BigEndian.Uint32(data[4:8])
		length := binary.BigEndian.Uint32(data[8:12])
		return Request{idx, begin, length}, nil
	case PieceID:
		if length < 9 {
			return nil, fmt.Errorf("invalid piece message")
		}
		idx := binary.BigEndian.Uint32(data[:4])
		begin := binary.BigEndian.Uint32(data[4:8])
		block := data[8:]
		return Piece{idx, begin, block}, nil

	case CancelID:
		if length != 13 {
			return nil, fmt.Errorf("invalid cancel message")
		}
		idx := binary.BigEndian.Uint32(data[:4])
		begin := binary.BigEndian.Uint32(data[4:8])
		length := binary.BigEndian.Uint32(data[8:12])
		return Cancel{idx, begin, length}, nil
	case ExtendedID:
		if length < 2 {
			return nil, fmt.Errorf("invalid extended message")
		}
		messID := data[0]
		decoded, err := DecodeExtended(ExtendedMessageID(messID), data[1:])
		if err != nil {
			return nil, err
		}
		return Extended{MessageID: messID, ExtendedMessage: decoded}, nil
	default:
		return nil, fmt.Errorf("invalid or non supported message id")
	}
}
