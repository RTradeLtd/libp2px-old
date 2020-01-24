package stream

import (
	"fmt"

	"github.com/RTradeLtd/libp2p-core/mux"
	"github.com/RTradeLtd/libp2p-core/network"
	"github.com/RTradeLtd/libp2p-core/transport"
)

type transportConn struct {
	mux.MuxedConn
	network.ConnMultiaddrs
	network.ConnSecurity
	transport transport.Transport
}

func (t *transportConn) Transport() transport.Transport {
	return t.transport
}

func (t *transportConn) String() string {
	ts := ""
	if s, ok := t.transport.(fmt.Stringer); ok {
		ts = "[" + s.String() + "]"
	}
	return fmt.Sprintf(
		"<stream.Conn%s %s (%s) <-> %s (%s)>",
		ts,
		t.LocalMultiaddr(),
		t.LocalPeer(),
		t.RemoteMultiaddr(),
		t.RemotePeer(),
	)
}
