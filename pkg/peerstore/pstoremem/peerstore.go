package pstoremem

import pstore "github.com/RTradeLtd/libp2px/pkg/peerstore"

// NewPeerstore creates an in-memory threadsafe collection of peers.
func NewPeerstore() pstore.Peerstore {
	return pstore.NewPeerstore(
		NewKeyBook(),
		NewAddrBook(),
		NewProtoBook(),
		NewPeerMetadata())
}
