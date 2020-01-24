package pstoremem

import (
	"context"

	pi "github.com/RTradeLtd/libp2px-core/peerstore"
	pstore "github.com/RTradeLtd/libp2px/pkg/peerstore"
)

// NewPeerstore creates an in-memory threadsafe collection of peers.
func NewPeerstore(ctx context.Context) pi.Peerstore {
	cctx, cancel := context.WithCancel(ctx)
	return pstore.NewPeerstore(
		cctx, cancel,
		NewKeyBook(),
		NewAddrBook(cctx),
		NewProtoBook(),
		NewPeerMetadata())
}
