package routedhost

import (
	"context"
	"fmt"
	"time"

	"github.com/RTradeLtd/libp2px-core/connmgr"
	"github.com/RTradeLtd/libp2px-core/event"
	"github.com/RTradeLtd/libp2px-core/host"
	"github.com/RTradeLtd/libp2px-core/network"
	"github.com/RTradeLtd/libp2px-core/peer"
	"github.com/RTradeLtd/libp2px-core/peerstore"
	"github.com/RTradeLtd/libp2px-core/protocol"
	"go.uber.org/zap"

	ma "github.com/multiformats/go-multiaddr"
)

// AddressTTL is the expiry time for our addresses.
// We expire them quickly.
const AddressTTL = time.Second * 10

// RoutedHost is a p2p Host that includes a routing system.
// This allows the Host to find the addresses for peers when
// it does not have them.
type RoutedHost struct {
	host   host.Host // embedded other host.
	route  Routing
	logger *zap.Logger
}

// Routing is an interface that indicates we can find peers
type Routing interface {
	FindPeer(context.Context, peer.ID) (peer.AddrInfo, error)
}

// Wrap is used to wrap a non-routed libp2px host, and leverage
// the content routing system to find peers if we do not
// alreayd know them
func Wrap(h host.Host, r Routing, logger *zap.Logger) *RoutedHost {
	return &RoutedHost{h, r, logger.Named("routedhost")}
}

// Connect ensures there is a connection between this host and the peer with
// given peer.ID. See (host.Host).Connect for more information.
//
// RoutedHost's Connect differs in that if the host has no addresses for a
// given peer, it will use its routing system to try to find some.
func (rh *RoutedHost) Connect(ctx context.Context, pi peer.AddrInfo) error {
	// first, check if we're already connected.
	if rh.Network().Connectedness(pi.ID) == network.Connected {
		return nil
	}

	// if we were given some addresses, keep + use them.
	if len(pi.Addrs) > 0 {
		rh.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.TempAddrTTL)
	}

	// Check if we have some addresses in our recent memory.
	addrs := rh.Peerstore().Addrs(pi.ID)
	if len(addrs) < 1 {
		// no addrs? find some with the routing system.
		var err error
		addrs, err = rh.findPeerAddrs(ctx, pi.ID)
		if err != nil {
			return err
		}
	}

	// Issue 448: if our address set includes routed specific relay addrs,
	// we need to make sure the relay's addr itself is in the peerstore or else
	// we wont be able to dial it.
	for _, addr := range addrs {
		_, err := addr.ValueForProtocol(ma.P_CIRCUIT)
		if err != nil {
			// not a relay address
			continue
		}

		if addr.Protocols()[0].Code != ma.P_P2P {
			// not a routed relay specific address
			continue
		}

		relay, _ := addr.ValueForProtocol(ma.P_P2P)

		relayID, err := peer.IDFromString(relay)
		if err != nil {
			rh.logger.Debug("failed to parse relay id", zap.String("relay", relay), zap.Error(err))
			continue
		}

		if len(rh.Peerstore().Addrs(relayID)) > 0 {
			// we already have addrs for this relay
			continue
		}

		relayAddrs, err := rh.findPeerAddrs(ctx, relayID)
		if err != nil {
			rh.logger.Debug("failed to find relay", zap.String("relay", relay), zap.Error(err))
			continue
		}

		rh.Peerstore().AddAddrs(relayID, relayAddrs, peerstore.TempAddrTTL)
	}

	// if we're here, we got some addrs. let's use our wrapped host to connect.
	pi.Addrs = addrs
	return rh.host.Connect(ctx, pi)
}

func (rh *RoutedHost) findPeerAddrs(ctx context.Context, id peer.ID) ([]ma.Multiaddr, error) {
	pi, err := rh.route.FindPeer(ctx, id)
	if err != nil {
		return nil, err // couldnt find any :(
	}

	if pi.ID != id {
		err = fmt.Errorf("routing failure: provided addrs for different peer")
		return nil, err
	}

	return pi.Addrs, nil
}

// ID returns the ID of the routed host
func (rh *RoutedHost) ID() peer.ID {
	return rh.host.ID()
}

// Peerstore returns the routed host's underlying peerstore
func (rh *RoutedHost) Peerstore() peerstore.Peerstore {
	return rh.host.Peerstore()
}

// Addrs returns all multiaddresses of the routed host
func (rh *RoutedHost) Addrs() []ma.Multiaddr {
	return rh.host.Addrs()
}

// Network returns the network object of the routed host
func (rh *RoutedHost) Network() network.Network {
	return rh.host.Network()
}

// Mux returns the routed hosts's protocol muxer
func (rh *RoutedHost) Mux() protocol.Switch {
	return rh.host.Mux()
}

// EventBus returns the routed host's event bus
func (rh *RoutedHost) EventBus() event.Bus {
	return rh.host.EventBus()
}

// SetStreamHandler is used to add a supported protocol, and a handler for the protocol streams
func (rh *RoutedHost) SetStreamHandler(pid protocol.ID, handler network.StreamHandler) {
	rh.host.SetStreamHandler(pid, handler)
}

// SetStreamHandlerMatch is used to set a match func for protocol stream handlers
func (rh *RoutedHost) SetStreamHandlerMatch(pid protocol.ID, m func(string) bool, handler network.StreamHandler) {
	rh.host.SetStreamHandlerMatch(pid, m, handler)
}

// RemoveStreamHandler is used to remove all stream handlers for the given protocol
func (rh *RoutedHost) RemoveStreamHandler(pid protocol.ID) {
	rh.host.RemoveStreamHandler(pid)
}

// NewStream is used to create a new stream
func (rh *RoutedHost) NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error) {
	// Ensure we have a connection, with peer addresses resolved by the routing system (#207)
	// It is not sufficient to let the underlying host connect, it will most likely not have
	// any addresses for the peer without any prior connections.
	// If the caller wants to prevent the host from dialing, it should use the NoDial option.
	if nodial, _ := network.GetNoDial(ctx); !nodial {
		err := rh.Connect(ctx, peer.AddrInfo{ID: p})
		if err != nil {
			return nil, err
		}
	}

	return rh.host.NewStream(ctx, p, pids...)
}

// Close returns the routed host's libp2px host object
func (rh *RoutedHost) Close() error {
	// no need to close IpfsRouting. we dont own it.
	return rh.host.Close()
}

// ConnManager returns the underlying connection manager
func (rh *RoutedHost) ConnManager() connmgr.ConnManager {
	return rh.host.ConnManager()
}

var _ (host.Host) = (*RoutedHost)(nil)
