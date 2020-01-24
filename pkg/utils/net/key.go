package testutil

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"

	ic "github.com/RTradeLtd/libp2px-core/crypto"
	pb "github.com/RTradeLtd/libp2px-core/crypto/pb"

	"github.com/RTradeLtd/libp2px-core/peer"

	tnet "github.com/RTradeLtd/libp2px/pkg/testing/net"

	ma "github.com/multiformats/go-multiaddr"
)

// TestBogusPrivateKey is a key used for testing (to avoid expensive keygen)
type TestBogusPrivateKey []byte

// TestBogusPublicKey is a key used for testing (to avoid expensive keygen)
type TestBogusPublicKey []byte

// Verify is used to verify the data
func (pk TestBogusPublicKey) Verify(data, sig []byte) (bool, error) {
	return bytes.Equal(data, reverse(sig)), nil
}

// Bytes returns the key bytes
func (pk TestBogusPublicKey) Bytes() ([]byte, error) {
	return []byte(pk), nil
}

// Encrypt encrypts the contents
func (pk TestBogusPublicKey) Encrypt(b []byte) ([]byte, error) {
	return reverse(b), nil
}

// Equals checks whether this key is equal to another
func (pk TestBogusPublicKey) Equals(k ic.Key) bool {
	return ic.KeyEqual(pk, k)
}

// Raw returns the raw bytes of the key (not wrapped in the
// libp2p-crypto protobuf).
func (pk TestBogusPublicKey) Raw() ([]byte, error) {
	return pk, nil
}

// Type returns the protobof key type.
func (pk TestBogusPublicKey) Type() pb.KeyType {
	return pb.KeyType_RSA
}

// GenSecret returns a secret
func (sk TestBogusPrivateKey) GenSecret() []byte {
	return []byte(sk)
}

// Sign signs the message
func (sk TestBogusPrivateKey) Sign(message []byte) ([]byte, error) {
	return reverse(message), nil
}

// GetPublic returns the public key
func (sk TestBogusPrivateKey) GetPublic() ic.PubKey {
	return TestBogusPublicKey(sk)
}

// Decrypt decrypts the contents
func (sk TestBogusPrivateKey) Decrypt(b []byte) ([]byte, error) {
	return reverse(b), nil
}

// Bytes returns the bytes of the key
func (sk TestBogusPrivateKey) Bytes() ([]byte, error) {
	return []byte(sk), nil
}

// Equals checks whether this key is equal to another
func (sk TestBogusPrivateKey) Equals(k ic.Key) bool {
	return ic.KeyEqual(sk, k)
}

// Raw returns the raw bytes of the key (not wrapped in the
// libp2p-crypto protobuf).
func (sk TestBogusPrivateKey) Raw() ([]byte, error) {
	return sk, nil
}

// Type returns the protobof key type.
func (sk TestBogusPrivateKey) Type() pb.KeyType {
	return pb.KeyType_RSA
}

// RandTestBogusPrivateKey returns a new test private key
func RandTestBogusPrivateKey() (TestBogusPrivateKey, error) {
	k := make([]byte, 5)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil, err
	}
	return TestBogusPrivateKey(k), nil
}

// RandTestBogusPublicKey returns a new test public key
func RandTestBogusPublicKey() (TestBogusPublicKey, error) {
	k, err := RandTestBogusPrivateKey()
	return TestBogusPublicKey(k), err
}

// RandTestBogusPrivateKeyOrFatal errors out if we cant generate a private key
func RandTestBogusPrivateKeyOrFatal(t *testing.T) TestBogusPrivateKey {
	k, err := RandTestBogusPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	return k
}

// RandTestBogusPublicKeyOrFatal errors out if we cant generate a public key
func RandTestBogusPublicKeyOrFatal(t *testing.T) TestBogusPublicKey {
	k, err := RandTestBogusPublicKey()
	if err != nil {
		t.Fatal(err)
	}
	return k
}

// RandTestBogusIdentity returns a new test identity
func RandTestBogusIdentity() (tnet.Identity, error) {
	k, err := RandTestBogusPrivateKey()
	if err != nil {
		return nil, err
	}

	id, err := peer.IDFromPrivateKey(k)
	if err != nil {
		return nil, err
	}

	return &identity{
		k:  k,
		id: id,
		a:  tnet.RandLocalTCPAddress(),
	}, nil
}

// RandTestBogusIdentityOrFatal errors out if we cant generate a new identity
func RandTestBogusIdentityOrFatal(t *testing.T) tnet.Identity {
	k, err := RandTestBogusIdentity()
	if err != nil {
		t.Fatal(err)
	}
	return k
}

// identity is a temporary shim to delay binding of PeerNetParams.
type identity struct {
	k  TestBogusPrivateKey
	id peer.ID
	a  ma.Multiaddr
}

func (p *identity) ID() peer.ID {
	return p.id
}

func (p *identity) Address() ma.Multiaddr {
	return p.a
}

func (p *identity) PrivateKey() ic.PrivKey {
	return p.k
}

func (p *identity) PublicKey() ic.PubKey {
	return p.k.GetPublic()
}

func reverse(a []byte) []byte {
	b := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		b[i] = a[len(a)-1-i]
	}
	return b
}
