package autonat

import (
	"time"

	"github.com/RTradeLtd/libp2px-core/network"
	"go.uber.org/zap"

	ma "github.com/multiformats/go-multiaddr"
)

var _ network.Notifiee = (*AmbientAutoNAT)(nil)

// AutoNATIdentifyDelay is the default delay before running autonat identification
var AutoNATIdentifyDelay = 5 * time.Second

// Listen satisfies the network.Notifiee interface
func (as *AmbientAutoNAT) Listen(net network.Network, a ma.Multiaddr) {}

// ListenClose satisfies the network.Notifiee interface
func (as *AmbientAutoNAT) ListenClose(net network.Network, a ma.Multiaddr) {}

// OpenedStream satisfies the network.Notifiee interface
func (as *AmbientAutoNAT) OpenedStream(net network.Network, s network.Stream) {}

// ClosedStream satisfies the network.Notifiee interface
func (as *AmbientAutoNAT) ClosedStream(net network.Network, s network.Stream) {}

// Connected satisfies the network.Notifiee interface
func (as *AmbientAutoNAT) Connected(net network.Network, c network.Conn) {
	p := c.RemotePeer()

	go func() {
		// add some delay for identify
		time.Sleep(AutoNATIdentifyDelay)

		protos, err := as.host.Peerstore().SupportsProtocols(p, AutoNATProto)
		if err != nil {
			as.logger.Error("failed to retrieve supported protocols", zap.Error(err), zap.String("peer.id", p.String()))
			return
		}
		if len(protos) > 0 {
			as.logger.Info("discovered autonat peer", zap.String("peer.id", p.String()))
			as.mx.Lock()
			as.peers[p] = as.host.Peerstore().Addrs(p)
			as.mx.Unlock()
		}
	}()
}

// Disconnected satisfies the network.Notifiee interface
func (as *AmbientAutoNAT) Disconnected(net network.Network, c network.Conn) {}
