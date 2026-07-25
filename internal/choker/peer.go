package choker

import "time"

type Peer interface {
	UploadRate() uint64
	DownloadRate() uint64
	ConnectedAt() time.Time
	IsInterested() bool
}
