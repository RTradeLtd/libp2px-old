package peerstore

import (
	core "github.com/RTradeLtd/libp2px-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

// Deprecated: use github.com/RTradeLtd/libp2px-core/peer.Info instead.
type PeerInfo = core.AddrInfo

// Deprecated: use github.com/RTradeLtd/libp2px-core/peer.ErrInvalidAddr instead.
var ErrInvalidAddr = core.ErrInvalidAddr

// Deprecated: use github.com/RTradeLtd/libp2px-core/peer.AddrInfoFromP2pAddr instead.
func InfoFromP2pAddr(m ma.Multiaddr) (*core.AddrInfo, error) {
	return core.AddrInfoFromP2pAddr(m)
}

// Deprecated: use github.com/RTradeLtd/libp2px-core/peer.AddrInfoToP2pAddrs instead.
func InfoToP2pAddrs(pi *core.AddrInfo) ([]ma.Multiaddr, error) {
	return core.AddrInfoToP2pAddrs(pi)
}
