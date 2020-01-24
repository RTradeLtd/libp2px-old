package relay

import (
	"context"
	"time"

	inet "github.com/RTradeLtd/libp2px-core/network"
	peer "github.com/RTradeLtd/libp2px-core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

var _ inet.Notifiee = (*Notifiee)(nil)

// Notifiee does ??
type Notifiee Relay

func (r *Relay) notifiee() inet.Notifiee {
	return (*Notifiee)(r)
}

// Relay ??
func (n *Notifiee) Relay() *Relay {
	return (*Relay)(n)
}

// Listen satisfies the network.Notifiee interface
func (n *Notifiee) Listen(net inet.Network, a ma.Multiaddr) {}

// ListenClose satisfies the network.Notifiee interface
func (n *Notifiee) ListenClose(net inet.Network, a ma.Multiaddr) {}

// OpenedStream satisfies the network.Notifiee interface
func (n *Notifiee) OpenedStream(net inet.Network, s inet.Stream) {}

// ClosedStream satisfies the network.Notifiee interface
func (n *Notifiee) ClosedStream(net inet.Network, s inet.Stream) {}

// Connected satisfies the network.Notifiee interface
func (n *Notifiee) Connected(s inet.Network, c inet.Conn) {
	if n.Relay().Matches(c.RemoteMultiaddr()) {
		return
	}

	go func(id peer.ID) {
		ctx, cancel := context.WithTimeout(n.ctx, time.Second)
		defer cancel()

		canhop, err := n.Relay().CanHop(ctx, id)
		if err != nil {
			return
		}

		if canhop {
			n.mx.Lock()
			n.relays[id] = struct{}{}
			n.mx.Unlock()
			n.host.ConnManager().TagPeer(id, "relay-hop", 2)
		}
	}(c.RemotePeer())
}

// Disconnected satisfies the network.Notifiee interface
func (n *Notifiee) Disconnected(s inet.Network, c inet.Conn) {}
