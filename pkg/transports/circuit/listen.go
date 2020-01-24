package relay

import (
	"net"

	pb "github.com/RTradeLtd/libp2px/pkg/transports/circuit/pb"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
)

var _ manet.Listener = (*Listener)(nil)

// Listener is ???
type Listener Relay

// Relay returns the underlying relay
func (l *Listener) Relay() *Relay {
	return (*Relay)(l)
}

// Listener returns the listener
func (r *Relay) Listener() *Listener {
	// TODO: Only allow one!
	return (*Listener)(r)
}

// Accept is used to accept ??
func (l *Listener) Accept() (manet.Conn, error) {
	select {
	case c := <-l.incoming:
		err := l.Relay().writeResponse(c.stream, pb.CircuitRelay_SUCCESS)
		if err != nil {
			c.stream.Reset()
			return nil, err
		}

		c.tagHop()
		return c, nil
	case <-l.ctx.Done():
		return nil, l.ctx.Err()
	}
}

// Addr returns the net addr
func (l *Listener) Addr() net.Addr {
	return &NetAddr{
		Relay:  "any",
		Remote: "any",
	}
}

// Multiaddr returns the multiaddr of ?
func (l *Listener) Multiaddr() ma.Multiaddr {
	return circuitAddr
}

// Close is a noop closer
func (l *Listener) Close() error {
	// TODO: noop?
	return nil
}
