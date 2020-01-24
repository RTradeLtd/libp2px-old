package tcp

import (
	"context"
	"net"
	"time"

	"github.com/RTradeLtd/libp2px-core/peer"
	"github.com/RTradeLtd/libp2px-core/transport"
	rtpt "github.com/RTradeLtd/libp2px/pkg/transports/reuseport"
	tptu "github.com/RTradeLtd/libp2px/pkg/transports/upgrader"

	ma "github.com/multiformats/go-multiaddr"
	mafmt "github.com/multiformats/go-multiaddr-fmt"
	manet "github.com/multiformats/go-multiaddr-net"
)

// DefaultConnectTimeout is the (default) maximum amount of time the TCP
// transport will spend on the initial TCP connect before giving up.
var DefaultConnectTimeout = 5 * time.Second

// try to set linger on the connection, if possible.
func tryLinger(conn net.Conn, sec int) {
	type canLinger interface {
		SetLinger(int) error
	}

	if lingerConn, ok := conn.(canLinger); ok {
		_ = lingerConn.SetLinger(sec)
	}
}

type lingerListener struct {
	manet.Listener
	sec int
}

func (ll *lingerListener) Accept() (manet.Conn, error) {
	c, err := ll.Listener.Accept()
	if err != nil {
		return nil, err
	}
	tryLinger(c, ll.sec)
	return c, nil
}

// Transport is the TCP transport.
type Transport struct {
	// Connection upgrader for upgrading insecure stream connections to
	// secure multiplex connections.
	Upgrader *tptu.Upgrader

	// Explicitly disable reuseport.
	DisableReuseport bool

	// TCP connect timeout
	ConnectTimeout time.Duration

	reuse rtpt.Transport
}

var _ transport.Transport = &Transport{}

// NewTCPTransport creates a tcp transport object that tracks dialers and listeners
// created. It represents an entire tcp stack (though it might not necessarily be)
func NewTCPTransport(upgrader *tptu.Upgrader) *Transport {
	return &Transport{Upgrader: upgrader, ConnectTimeout: DefaultConnectTimeout}
}

// CanDial returns true if this transport believes it can dial the given
// multiaddr.
func (t *Transport) CanDial(addr ma.Multiaddr) bool {
	return mafmt.TCP.Matches(addr)
}

func (t *Transport) maDial(ctx context.Context, raddr ma.Multiaddr) (manet.Conn, error) {
	// Apply the deadline iff applicable
	if t.ConnectTimeout > 0 {
		deadline := time.Now().Add(t.ConnectTimeout)
		if d, ok := ctx.Deadline(); !ok || deadline.Before(d) {
			var cancel func()
			ctx, cancel = context.WithDeadline(ctx, deadline)
			defer cancel()
		}
	}

	if t.UseReuseport() {
		return t.reuse.DialContext(ctx, raddr)
	}
	var d manet.Dialer
	return d.DialContext(ctx, raddr)
}

// Dial dials the peer at the remote address.
func (t *Transport) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (transport.CapableConn, error) {
	conn, err := t.maDial(ctx, raddr)
	if err != nil {
		return nil, err
	}
	// Set linger to 0 so we never get stuck in the TIME-WAIT state. When
	// linger is 0, connections are _reset_ instead of closed with a FIN.
	// This means we can immediately reuse the 5-tuple and reconnect.
	tryLinger(conn, 0)
	return t.Upgrader.UpgradeOutbound(ctx, t, conn, p)
}

// UseReuseport returns true if reuseport is enabled and available.
func (t *Transport) UseReuseport() bool {
	return !t.DisableReuseport && ReuseportIsAvailable()
}

func (t *Transport) maListen(laddr ma.Multiaddr) (manet.Listener, error) {
	if t.UseReuseport() {
		return t.reuse.Listen(laddr)
	}
	return manet.Listen(laddr)
}

// Listen listens on the given multiaddr.
func (t *Transport) Listen(laddr ma.Multiaddr) (transport.Listener, error) {
	list, err := t.maListen(laddr)
	if err != nil {
		return nil, err
	}
	list = &lingerListener{list, 0}
	return t.Upgrader.UpgradeListener(t, list), nil
}

// Protocols returns the list of terminal protocols this transport can dial.
func (t *Transport) Protocols() []int {
	return []int{ma.P_TCP}
}

// Proxy always returns false for the TCP transport.
func (t *Transport) Proxy() bool {
	return false
}

func (t *Transport) String() string {
	return "TCP"
}
