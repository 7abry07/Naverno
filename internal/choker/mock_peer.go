package choker

import "time"

type MockPeer struct {
	Upload      uint64
	Download    uint64
	Interested  bool
	connectedAt time.Time
}

func NewMockPeer(upload, download uint64, interested bool, connectedAt time.Time) *MockPeer {
	return &MockPeer{
		Upload:      upload,
		Download:    download,
		Interested:  interested,
		connectedAt: connectedAt,
	}
}

func (p *MockPeer) UploadRate() uint64 {
	return p.Upload
}
func (p *MockPeer) DownloadRate() uint64 {
	return p.Download
}
func (p *MockPeer) ConnectedAt() time.Time {
	return p.connectedAt
}
func (p *MockPeer) IsInterested() bool {
	return p.Interested
}
