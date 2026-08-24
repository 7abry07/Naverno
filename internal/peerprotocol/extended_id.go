package peerprotocol

type ExtendedMessageID uint8
type UTMetadataMessageID uint8

const (
	HandshakeID  ExtendedMessageID = 0
	UTMetadataID ExtendedMessageID = 3

	UTMetadataRequestID  UTMetadataMessageID = 0
	UTMetadataResponseID UTMetadataMessageID = 1
	UTMetadataRejectID   UTMetadataMessageID = 2
)

var extendedStr = map[ExtendedMessageID]string{
	HandshakeID:  "handshake",
	UTMetadataID: "ut_metadata",
}

func (m Handshake) ID() ExtendedMessageID          { return HandshakeID }
func (m UTMetadataRequest) ID() ExtendedMessageID  { return UTMetadataID }
func (m UTMetadataResponse) ID() ExtendedMessageID { return UTMetadataID }
func (m UTMetadataReject) ID() ExtendedMessageID   { return UTMetadataID }
