package krpc

import (
	"encoding/binary"
	"fmt"
	"maps"
	// "reflect"

	"github.com/zeebo/bencode"
)

type ErrorCode int

const (
	GenericErrorCode  ErrorCode = 201
	ServerErrorCode   ErrorCode = 202
	ProtocolErrorCode ErrorCode = 203
	UknownMethodCode  ErrorCode = 204
)

type Message interface {
	Marshal(transactionID uint16) []byte
}

type Error struct {
	Message string
	Code    ErrorCode
}

func (e Error) Error() string {
	return e.Message
}

type Query struct {
	Name       string
	Parameters map[string]any
}

type Response struct {
	Values map[string]any
}

func (e Error) Marshal(transactionID uint16) []byte {
	var err_ struct {
		T string `bencode:"t"`
		Y string `bencode:"y"`
		E []any  `bencode:"e"`
	}
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, transactionID)
	err_.T = string(b)
	err_.Y = "e"
	err_.E = []any{e.Code, e.Message}
	encoded, err := bencode.EncodeBytes(&err_)
	if err != nil {
		return []byte{}
	}
	return encoded
}

func (q Query) Marshal(transactionID uint16) []byte {
	var qu struct {
		T string         `bencode:"t"`
		Y string         `bencode:"y"`
		Q string         `bencode:"q"`
		A map[string]any `bencode:"a"`
	}

	qu.A = make(map[string]any)

	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, transactionID)
	qu.T = string(b)
	qu.Y = "q"
	qu.Q = q.Name
	maps.Copy(qu.A, q.Parameters)
	encoded, err := bencode.EncodeBytes(&qu)
	if err != nil {
		return []byte{}
	}
	return encoded
}

func (r Response) Marshal(transactionID uint16) []byte {
	var res struct {
		T string         `bencode:"t"`
		Y string         `bencode:"y"`
		R map[string]any `bencode:"r"`
	}

	res.R = make(map[string]any)

	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, transactionID)
	res.T = string(b)
	res.Y = "r"
	res.R = map[string]any{}
	maps.Copy(res.R, r.Values)
	encoded, err := bencode.EncodeBytes(&res)
	if err != nil {
		return []byte{}
	}
	return encoded
}

func Decode(data []byte) (Message, uint16, error) {
	var msg struct {
		T string `bencode:"t"`
		Y string `bencode:"y"`

		// error
		E []any `bencode:"e"`

		// response
		R map[string]any `bencode:"r"`

		// query
		Q string         `bencode:"q"`
		A map[string]any `bencode:"a"`
	}

	msg.R = make(map[string]any)
	msg.A = make(map[string]any)

	err := bencode.DecodeBytes(data, &msg)
	if err != nil {
		return nil, 0, Error{Code: ProtocolErrorCode, Message: err.Error()}
	}

	tid := uint16(0)
	if len(msg.T) != 2 {
		return nil, 0, Error{Code: ProtocolErrorCode, Message: "invalid or missing transaction ID"}
	}
	tid = binary.BigEndian.Uint16([]byte(msg.T))

	switch msg.Y {
	case "q":
		q := Query{}
		q.Parameters = make(map[string]any)
		q.Name = msg.Q
		if msg.A == nil {
			return nil, tid, Error{Code: ProtocolErrorCode, Message: "missing parameters in query"}
		}
		maps.Copy(q.Parameters, msg.A)
		return q, tid, nil
	case "r":
		r := Response{}
		r.Values = make(map[string]any)
		if msg.R == nil {
			return nil, tid, Error{Code: ProtocolErrorCode, Message: "missing values in response"}
		}
		maps.Copy(r.Values, msg.R)
		return r, tid, nil
	case "e":
		e := Error{}
		if len(msg.E) < 2 {
			return nil, tid, Error{Code: ProtocolErrorCode, Message: "invalid list length in error"}
		}
		code, ok := msg.E[0].(int64)
		if !ok {
			return nil, tid, Error{Code: ProtocolErrorCode, Message: "invalid type for error code (int) in error"}
		}
		mess, ok := msg.E[1].(string)
		if !ok {
			return nil, tid, Error{Code: ProtocolErrorCode, Message: "invalid type for error message (string) in error"}
		}
		e.Code = ErrorCode(code)
		e.Message = mess
		return e, tid, nil
	default:
		return nil, tid, Error{Code: UknownMethodCode, Message: fmt.Sprintf("unknown method: \"%v\"", msg.Y)}
	}
}
