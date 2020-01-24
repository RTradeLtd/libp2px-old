package relay

import (
	"fmt"
	"net"
	"time"

	"github.com/RTradeLtd/libp2px-core/host"
	"github.com/RTradeLtd/libp2px-core/network"
	"github.com/RTradeLtd/libp2px-core/peer"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
)

// Conn is a circuit connection
type Conn struct {
	stream network.Stream
	remote peer.AddrInfo
	host   host.Host
}

// NetAddr ???
type NetAddr struct {
	Relay  string
	Remote string
}

// Network ??
func (n *NetAddr) Network() string {
	return "libp2p-circuit-relay"
}

func (n *NetAddr) String() string {
	return fmt.Sprintf("relay[%s-%s]", n.Remote, n.Relay)
}

// Close closes the connection
func (c *Conn) Close() error {
	c.untagHop()
	return c.stream.Reset()
}

func (c *Conn) Read(buf []byte) (int, error) {
	return c.stream.Read(buf)
}

func (c *Conn) Write(buf []byte) (int, error) {
	return c.stream.Write(buf)
}

// SetDeadline sets the deadline
func (c *Conn) SetDeadline(t time.Time) error {
	return c.stream.SetDeadline(t)
}

// SetReadDeadline sets the read deadline
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.stream.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.stream.SetWriteDeadline(t)
}

// RemoteAddr returns the address of the remote peer
func (c *Conn) RemoteAddr() net.Addr {
	return &NetAddr{
		Relay:  c.stream.Conn().RemotePeer().Pretty(),
		Remote: c.remote.ID.Pretty(),
	}
}

// Increment the underlying relay connection tag by 1, thus increasing its protection from
// connection pruning. This ensures that connections to relays are not accidentally closed,
// by the connection manager, taking with them all the relayed connections (that may themselves
// be protected).
func (c *Conn) tagHop() {
	c.host.ConnManager().UpsertTag(c.stream.Conn().RemotePeer(), "relay-hop-stream", incrementTag)
}

// Decrement the underlying relay connection tag by 1; this is performed when we close the
// relayed connection.
func (c *Conn) untagHop() {
	c.host.ConnManager().UpsertTag(c.stream.Conn().RemotePeer(), "relay-hop-stream", decrementTag)
}

// RemoteMultiaddr  TODO: is it okay to cast c.Conn().RemotePeer() into a multiaddr? might be "user input"
func (c *Conn) RemoteMultiaddr() ma.Multiaddr {
	// TODO: We should be able to do this directly without converting to/from a string.
	relayAddr, err := ma.NewComponent(
		ma.ProtocolWithCode(ma.P_P2P).Name,
		c.stream.Conn().RemotePeer().Pretty(),
	)
	if err != nil {
		panic(err)
	}
	return ma.Join(c.stream.Conn().RemoteMultiaddr(), relayAddr, circuitAddr)
}

// LocalMultiaddr returns th eaddress of the local connection (us)
func (c *Conn) LocalMultiaddr() ma.Multiaddr {
	return c.stream.Conn().LocalMultiaddr()
}

// LocalAddr returns our local multi address
func (c *Conn) LocalAddr() net.Addr {
	na, err := manet.ToNetAddr(c.stream.Conn().LocalMultiaddr())
	if err != nil {
		return nil
	}
	return na
}
