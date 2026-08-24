package peerprotocol

type ExtendedMessageID byte
type UTMetadataMessageID uint8

const (
	ExtendedHandshakeID ExtendedMessageID = 0
	UTMetadataID        ExtendedMessageID = 3

	UTMetadataRequestID  UTMetadataMessageID = 0
	UTMetadataResponseID UTMetadataMessageID = 1
	UTMetadataRejectID   UTMetadataMessageID = 2
)

var extendedStr = map[ExtendedMessageID]string{
	ExtendedHandshakeID: "handshake",
	UTMetadataID:        "ut_metadata",
}

func (m ExtendedMessageID) String() string { return extendedStr[m] }

func (m ExtendedHandshake) LocalID() ExtendedMessageID  { return ExtendedHandshakeID }
func (m UTMetadataRequest) LocalID() ExtendedMessageID  { return UTMetadataID }
func (m UTMetadataResponse) LocalID() ExtendedMessageID { return UTMetadataID }
func (m UTMetadataReject) LocalID() ExtendedMessageID   { return UTMetadataID }
