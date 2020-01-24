package relay

import (
	"context"
	"fmt"

	"github.com/RTradeLtd/libp2px-core/host"
	"github.com/RTradeLtd/libp2px-core/transport"

	tptu "github.com/RTradeLtd/libp2px/pkg/transports/upgrader"
	ma "github.com/multiformats/go-multiaddr"
)

// PCIRCUIT is deprecated
const PCIRCUIT = ma.P_CIRCUIT

// Protocol is Deprecated: use ma.ProtocolWithCode(P_CIRCUIT)
var Protocol = ma.ProtocolWithCode(PCIRCUIT)

var circuitAddr = ma.Cast(Protocol.VCode)

var _ transport.Transport = (*Transport)(nil)

// Transport ??
type Transport Relay

// Relay does ?
func (t *Transport) Relay() *Relay {
	return (*Relay)(t)
}

// Transport does ??
func (r *Relay) Transport() *Transport {
	return (*Transport)(r)
}

// Listen does ??
func (t *Transport) Listen(laddr ma.Multiaddr) (transport.Listener, error) {
	// TODO: Ensure we have a connection to the relay, if specified. Also,
	// make sure the multiaddr makes sense.
	if !t.Relay().Matches(laddr) {
		return nil, fmt.Errorf("%s is not a relay address", laddr)
	}
	return t.upgrader.UpgradeListener(t, t.Relay().Listener()), nil
}

// CanDial indicates if we can dial th eper
func (t *Transport) CanDial(raddr ma.Multiaddr) bool {
	return t.Relay().Matches(raddr)
}

// Proxy returns if the transport is a proxy
func (t *Transport) Proxy() bool {
	return true
}

// Protocols returns the protocols of this transport
func (t *Transport) Protocols() []int {
	return []int{PCIRCUIT}
}

// AddRelayTransport constructs a relay and adds it as a transport to the host network.
func AddRelayTransport(ctx context.Context, h host.Host, upgrader *tptu.Upgrader, opts ...Opt) error {
	_, ok := h.Network().(transport.TransportNetwork)
	if !ok {
		return fmt.Errorf("%v is not a transport network", h.Network())
	}

	_, err := NewRelay(ctx, h, upgrader, opts...)
	if err != nil {
		return err
	}

	// There's no nice way to handle these errors as we have no way to tear
	// down the relay.
	// TODO
	return nil
}
