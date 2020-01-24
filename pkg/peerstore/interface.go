package peerstore

import core "github.com/RTradeLtd/libp2px-core/peerstore"

// Deprecated: use github.com/RTradeLtd/libp2px-core/peerstore.ErrNotFound instead.
var ErrNotFound = core.ErrNotFound

var (
	// Deprecated: use github.com/RTradeLtd/libp2px-core/peerstore.AddressTTL instead.
	AddressTTL = core.AddressTTL

	// Deprecated: use github.com/RTradeLtd/libp2px-core/peerstore.TempAddrTTL instead.
	TempAddrTTL = core.TempAddrTTL

	// Deprecated: use github.com/RTradeLtd/libp2px-core/peerstore.ProviderAddrTTL instead.
	ProviderAddrTTL = core.ProviderAddrTTL

	// Deprecated: use github.com/RTradeLtd/libp2px-core/peerstore.RecentlyConnectedAddrTTL instead.
	RecentlyConnectedAddrTTL = core.RecentlyConnectedAddrTTL

	// Deprecated: use github.com/RTradeLtd/libp2px-core/peerstore.OwnObservedAddrTTL instead.
	OwnObservedAddrTTL = core.OwnObservedAddrTTL
)

const (
	// Deprecated: use github.com/RTradeLtd/libp2px-core/peerstore.PermanentAddrTTL instead.
	PermanentAddrTTL = core.PermanentAddrTTL

	// Deprecated: use github.com/RTradeLtd/libp2px-core/peerstore.ConnectedAddrTTL instead.
	ConnectedAddrTTL = core.ConnectedAddrTTL
)

// Deprecated: use github.com/RTradeLtd/libp2px-core/peerstore.Peerstore instead.
type Peerstore = core.Peerstore

// Deprecated: use github.com/RTradeLtd/libp2px-core/peerstore.PeerMetadata instead.
type PeerMetadata = core.PeerMetadata

// Deprecated: use github.com/RTradeLtd/libp2px-core/peerstore.AddrBook instead.
type AddrBook = core.AddrBook

// Deprecated: use github.com/RTradeLtd/libp2px-core/peerstore.KeyBook instead.
type KeyBook = core.KeyBook

// Deprecated: use github.com/RTradeLtd/libp2px-core/peerstore.ProtoBook instead.
type ProtoBook = core.ProtoBook
