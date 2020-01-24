package pstoremem


import (
	pi "github.com/RTradeLtd/libp2px-core/peerstore"
	pstore "github.com/RTradeLtd/libp2px/pkg/peerstore"
)

// NewPeerstore creates an in-memory threadsafe collection of peers.
func NewPeerstore() pi.Peerstore {
	return pstore.NewPeerstore(
		NewKeyBook(),
		NewAddrBook(),
		NewProtoBook(),
		NewPeerMetadata())
}
