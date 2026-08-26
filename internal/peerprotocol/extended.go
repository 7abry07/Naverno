package peerprotocol

import (
	"bytes"
	"fmt"

	"github.com/zeebo/bencode"
)

type ExtendedMessage interface {
	Marshal() []byte
	LocalID() ExtendedMessageID
}

type ExtendedHandshake struct {
	IDs          map[string]uint8
	MetadataSize int
}

type UTMetadataRequest struct{ Piece uint32 }
type UTMetadataReject struct{ Piece uint32 }
type UTMetadataResponse struct {
	Piece uint32
	Data  []byte
}

func (m ExtendedHandshake) Marshal() []byte {
	// var hs struct {
	// 	Ids          map[string]int `bencode:"m"`
	// 	MetadataSize int            `bencode:"metadata_size"`
	// }

	hs := map[string]any{}
	hs["m"] = m.IDs
	if m.MetadataSize > 0 {
		hs["metadata_size"] = m.MetadataSize
	}

	// hs.Ids = make(map[string]int)
	// hs.MetadataSize = m.MetadataSize
	// for str, ids := range m.IDs {
	// 	hs.Ids[str] = int(ids)
	// }
	marshaled, _ := bencode.EncodeBytes(hs)
	return marshaled
}

func (m UTMetadataRequest) Marshal() []byte {
	var req struct {
		Id    int `bencode:"msg_type"`
		Piece int `bencode:"piece"`
	}
	req.Id = int(UTMetadataRequestID)
	req.Piece = int(m.Piece)
	marshaled, _ := bencode.EncodeBytes(req)
	return marshaled
}

func (m UTMetadataResponse) Marshal() []byte {
	var res struct {
		Id    int `bencode:"msg_type"`
		Piece int `bencode:"piece"`
		Size  int `bencode:"total_size"`
	}
	res.Id = int(UTMetadataResponseID)
	res.Piece = int(m.Piece)
	res.Size = len(m.Data)
	marshaled, _ := bencode.EncodeBytes(res)
	marshaled = append(marshaled, m.Data...)
	return marshaled
}

func (m UTMetadataReject) Marshal() []byte {
	var rej struct {
		Id    int `bencode:"msg_type"`
		Piece int `bencode:"piece"`
	}
	rej.Id = int(UTMetadataRejectID)
	rej.Piece = int(m.Piece)
	marshaled, _ := bencode.EncodeBytes(rej)
	return marshaled
}

func DecodeExtended(id ExtendedMessageID, data []byte) (ExtendedMessage, error) {
	switch id {
	case ExtendedHandshakeID:
		var msg struct {
			Ids          map[string]int `bencode:"m"`
			MetadataSize int            `bencode:"metadata_size"`
		}

		decoder := bencode.NewDecoder(bytes.NewBuffer(data))
		err := decoder.Decode(&msg)
		if err != nil {
			return nil, fmt.Errorf("invalid metadata message -> %v", err)
		}
		hs := ExtendedHandshake{
			IDs: make(map[string]uint8),
		}
		for str, id := range msg.Ids {
			hs.IDs[str] = uint8(id)
		}
		hs.MetadataSize = msg.MetadataSize
		return hs, nil
	case UTMetadataID:
		var msg struct {
			Id    int `bencode:"msg_type"`
			Piece int `bencode:"piece"`
			Size  int `bencode:"total_size"`
		}
		decoder := bencode.NewDecoder(bytes.NewBuffer(data))
		err := decoder.Decode(&msg)
		if err != nil {
			return nil, fmt.Errorf("invalid metadata message -> %v", err)
		}

		switch msg.Id {
		case int(UTMetadataRequestID):
			return UTMetadataRequest{Piece: uint32(msg.Piece)}, nil
		case int(UTMetadataResponseID):
			if msg.Size == 0 {
				return nil, fmt.Errorf("invalid metadata message")
			}
			data = data[decoder.BytesParsed():]
			size := min(len(data), msg.Size)

			res := UTMetadataResponse{
				Piece: uint32(msg.Piece),
				Data:  data[:size],
			}
			return res, nil
		case int(UTMetadataRejectID):
			return UTMetadataReject{Piece: uint32(msg.Piece)}, nil
		}
	}
	return nil, fmt.Errorf("unrecognized message id")
}
