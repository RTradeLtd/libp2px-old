// Package secio is used to encrypt `go-libp2p-conn` connections. Connections wrapped by secio use secure sessions provided by this package to encrypt all traffic. A TLS-like handshake is used to setup the communication channel.
package secio

import (
	"context"
	"net"
	"time"

	ci "github.com/RTradeLtd/libp2px-core/crypto"
	"github.com/RTradeLtd/libp2px-core/peer"
	"github.com/RTradeLtd/libp2px-core/sec"
)

// ID is secio's protocol ID (used when negotiating with multistream)
const ID = "/secio/1.0.0"

// Transport constructs secure communication sessions for a peer.
type Transport struct {
	LocalID    peer.ID
	PrivateKey ci.PrivKey
}

// New returns a new SecIO transport
func New(sk ci.PrivKey) (*Transport, error) {
	id, err := peer.IDFromPrivateKey(sk)
	if err != nil {
		return nil, err
	}
	return &Transport{
		LocalID:    id,
		PrivateKey: sk,
	}, nil
}

var _ sec.SecureTransport = (*Transport)(nil)

// SecureInbound secures the inbound connection
func (sg *Transport) SecureInbound(ctx context.Context, insecure net.Conn) (sec.SecureConn, error) {
	return newSecureSession(ctx, sg.LocalID, sg.PrivateKey, insecure, "")
}

// SecureOutbound secures the outbound connection
func (sg *Transport) SecureOutbound(ctx context.Context, insecure net.Conn, p peer.ID) (sec.SecureConn, error) {
	return newSecureSession(ctx, sg.LocalID, sg.PrivateKey, insecure, p)
}

// SetReadDeadline sets the read deadline
func (s *secureSession) SetReadDeadline(t time.Time) error {
	return s.insecure.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline
func (s *secureSession) SetWriteDeadline(t time.Time) error {
	return s.insecure.SetWriteDeadline(t)
}

// SetDeadlien sets the deadline
func (s *secureSession) SetDeadline(t time.Time) error {
	return s.insecure.SetDeadline(t)
}

func (s *secureSession) RemoteAddr() net.Addr {
	return s.insecure.RemoteAddr()
}

func (s *secureSession) LocalAddr() net.Addr {
	return s.insecure.LocalAddr()
}

// LocalPeer retrieves the local peer.
func (s *secureSession) LocalPeer() peer.ID {
	return s.localPeer
}

// LocalPrivateKey retrieves the local peer's PrivateKey
func (s *secureSession) LocalPrivateKey() ci.PrivKey {
	return s.localKey
}

// RemotePeer retrieves the remote peer.
func (s *secureSession) RemotePeer() peer.ID {
	return s.remotePeer
}

// RemotePublicKey retrieves the remote public key.
func (s *secureSession) RemotePublicKey() ci.PubKey {
	return s.remote.permanentPubKey
}
