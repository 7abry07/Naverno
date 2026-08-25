package infodownloader

type Peer interface{ RequestMetadata(piece uint32) }
