package testing

import (
	"context"
	"testing"

	"github.com/RTradeLtd/libp2px-core/metrics"
	"github.com/RTradeLtd/libp2px-core/network"
	"github.com/RTradeLtd/libp2px-core/peerstore"
	tnet "github.com/RTradeLtd/libp2px/pkg/testing/net"
	"github.com/RTradeLtd/libp2px/pkg/transports/tcp"
	"go.uber.org/zap/zaptest"

	pstoremem "github.com/RTradeLtd/libp2px/pkg/peerstore/pstoremem"
	csms "github.com/RTradeLtd/libp2px/pkg/transports/conn-security-multistream"
	secio "github.com/RTradeLtd/libp2px/pkg/transports/secio"
	msmux "github.com/RTradeLtd/libp2px/pkg/transports/stream-muxer-multistream"
	tptu "github.com/RTradeLtd/libp2px/pkg/transports/upgrader"
	yamux "github.com/RTradeLtd/libp2px/pkg/transports/yamux"

	swarm "github.com/RTradeLtd/libp2px/pkg/swarm"
)

type config struct {
	disableReuseport bool
	dialOnly         bool
}

// Option is an option that can be passed when constructing a test swarm.
type Option func(*testing.T, *config)

// OptDisableReuseport disables reuseport in this test swarm.
var OptDisableReuseport Option = func(_ *testing.T, c *config) {
	c.disableReuseport = true
}

// OptDialOnly prevents the test swarm from listening.
var OptDialOnly Option = func(_ *testing.T, c *config) {
	c.dialOnly = true
}

// GenUpgrader creates a new connection upgrader for use with this swarm.
func GenUpgrader(n *swarm.Swarm) *tptu.Upgrader {
	id := n.LocalPeer()
	pk := n.Peerstore().PrivKey(id)
	secMuxer := new(csms.SSMuxer)
	secMuxer.AddTransport(secio.ID, &secio.Transport{
		LocalID:    id,
		PrivateKey: pk,
	})

	stMuxer := msmux.NewBlankTransport()
	stMuxer.AddTransport("/yamux/1.0.0", yamux.DefaultTransport)

	return &tptu.Upgrader{
		Secure:  secMuxer,
		Muxer:   stMuxer,
		Filters: n.Filters,
	}

}

// GenSwarm generates a new test swarm.
func GenSwarm(t *testing.T, ctx context.Context, opts ...Option) (*swarm.Swarm, func() error) {
	var cfg config
	for _, o := range opts {
		o(t, &cfg)
	}

	p := tnet.RandPeerNetParamsOrFatal(t)

	ps := pstoremem.NewPeerstore(ctx)
	ps.AddPubKey(p.ID, p.PubKey)
	ps.AddPrivKey(p.ID, p.PrivKey)
	s := swarm.NewSwarm(ctx, zaptest.NewLogger(t), p.ID, ps, metrics.NewBandwidthCounter())
	tcpTransport := tcp.NewTCPTransport(GenUpgrader(s))
	tcpTransport.DisableReuseport = cfg.disableReuseport

	if err := s.AddTransport(tcpTransport); err != nil {
		t.Fatal(err)
	}

	if !cfg.dialOnly {
		if err := s.Listen(p.Addr); err != nil {
			t.Fatal(err)
		}

		s.Peerstore().AddAddrs(p.ID, s.ListenAddresses(), peerstore.PermanentAddrTTL)
	}

	return s, ps.Close
}

// DivulgeAddresses adds swarm a's addresses to swarm b's peerstore.
func DivulgeAddresses(a, b network.Network) {
	id := a.LocalPeer()
	addrs := a.Peerstore().Addrs(id)
	b.Peerstore().AddAddrs(id, addrs, peerstore.PermanentAddrTTL)
}
