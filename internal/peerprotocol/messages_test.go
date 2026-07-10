package peerprotocol_test

import (
	"Naverno/internal/peerprotocol"
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name string
		msg  peerprotocol.Message
	}{
		{"Choke", peerprotocol.Choke{}},
		{"Unchoke", peerprotocol.Unchoke{}},
		{"Interested", peerprotocol.Interested{}},
		{"Uninterested", peerprotocol.Uninterested{}},
		{"Have", peerprotocol.Have{5}},
		{"Bitfield", peerprotocol.Bitfield{make([]byte, 10)}},
		{"Request", peerprotocol.Request{Idx: 5, Begin: 45, Length: 10}},
		{"Piece", peerprotocol.Piece{Idx: 5, Begin: 45, Data: make([]byte, 10)}},
		{"Cancel", peerprotocol.Cancel{Idx: 5, Begin: 45, Length: 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.msg.Marshal()

			decoded, err := peerprotocol.Decode(data)
			if err != nil {
				t.Fatalf("Decode() returned error: %v", err)
			}

			if !reflect.DeepEqual(decoded, tt.msg) {
				t.Fatalf("decoded = %#v, want %#v", decoded, tt.msg)
			}
		})
	}
}
