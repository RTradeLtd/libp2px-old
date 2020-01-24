package autonat

import (
	"context"
	"fmt"

	"github.com/RTradeLtd/libp2px-core/helpers"
	pb "github.com/RTradeLtd/libp2px/pkg/autonat/pb"

	"github.com/RTradeLtd/libp2px-core/host"
	"github.com/RTradeLtd/libp2px-core/network"
	"github.com/RTradeLtd/libp2px-core/peer"
	ggio "github.com/gogo/protobuf/io"
	ma "github.com/multiformats/go-multiaddr"
)

// NATClient is a stateless client interface to AutoNAT peers
type NATClient interface {
	// DialBack requests from a peer providing AutoNAT services to test dial back
	// and report the address on a successful connection.
	DialBack(ctx context.Context, p peer.ID) (ma.Multiaddr, error)
}

// Error is the class of errors signalled by AutoNAT services
type Error struct {
	Status pb.Message_ResponseStatus
	Text   string
}

// GetAddrs is a function that returns the addresses to dial back
type GetAddrs func() []ma.Multiaddr

// NewAutoNATClient creates a fresh instance of an AutoNATClient
// If getAddrs is nil, h.Addrs will be used
func NewAutoNATClient(h host.Host, getAddrs GetAddrs) NATClient {
	if getAddrs == nil {
		getAddrs = h.Addrs
	}
	return &client{h: h, getAddrs: getAddrs}
}

type client struct {
	h        host.Host
	getAddrs GetAddrs
}

func (c *client) DialBack(ctx context.Context, p peer.ID) (ma.Multiaddr, error) {
	s, err := c.h.NewStream(ctx, p, AutoNATProto)
	if err != nil {
		return nil, err
	}
	// Might as well just reset the stream. Once we get to this point, we
	// don't care about being nice.
	defer helpers.FullClose(s)

	r := ggio.NewDelimitedReader(s, network.MessageSizeMax)
	w := ggio.NewDelimitedWriter(s)

	req := newDialMessage(peer.AddrInfo{ID: c.h.ID(), Addrs: c.getAddrs()})
	err = w.WriteMsg(req)
	if err != nil {
		s.Reset()
		return nil, err
	}

	var res pb.Message
	err = r.ReadMsg(&res)
	if err != nil {
		s.Reset()
		return nil, err
	}

	if res.GetType() != pb.Message_DIAL_RESPONSE {
		return nil, fmt.Errorf("unexpected response: %s", res.GetType().String())
	}

	status := res.GetDialResponse().GetStatus()
	switch status {
	case pb.Message_OK:
		addr := res.GetDialResponse().GetAddr()
		return ma.NewMultiaddrBytes(addr)

	default:
		return nil, Error{Status: status, Text: res.GetDialResponse().GetStatusText()}
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("AutoNAT error: %s (%s)", e.Text, e.Status.String())
}

// IsDialError returns whether or not the error is a dial error
func (e Error) IsDialError() bool {
	return e.Status == pb.Message_E_DIAL_ERROR
}

// IsDialRefused returns whether or not the error is a dial refused error
func (e Error) IsDialRefused() bool {
	return e.Status == pb.Message_E_DIAL_REFUSED
}

// IsDialError returns true if the AutoNAT peer signalled an error dialing back
func IsDialError(e error) bool {
	ae, ok := e.(Error)
	return ok && ae.IsDialError()
}

// IsDialRefused returns true if the AutoNAT peer signalled refusal to dial back
func IsDialRefused(e error) bool {
	ae, ok := e.(Error)
	return ok && ae.IsDialRefused()
}
