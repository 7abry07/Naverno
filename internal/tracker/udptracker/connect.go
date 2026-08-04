package udptracker

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const PROTOCOL_ID = uint64(0x41727101980)

type connectRequest struct{}
type connectResponse struct {
	connectionID uint64
}

func (req connectRequest) Encode(transactionID uint32) []byte {
	encoded := []byte{}
	encoded = binary.BigEndian.AppendUint64(encoded, PROTOCOL_ID)
	encoded = binary.BigEndian.AppendUint32(encoded, uint32(action_connect))
	encoded = binary.BigEndian.AppendUint32(encoded, transactionID)
	return encoded
}

func (res *connectResponse) decode(data *bytes.Buffer) error {
	if data.Len() != 8 {
		return fmt.Errorf("invalid length for connect response")
	}

	connectionId := make([]byte, 8)

	_, err := data.Read(connectionId)
	if err != nil {
		return fmt.Errorf("error while reading connect response")
	}

	res.connectionID = binary.BigEndian.Uint64(connectionId)
	return nil
}

func (connectResponse) Action() action {
	return action_connect
}
