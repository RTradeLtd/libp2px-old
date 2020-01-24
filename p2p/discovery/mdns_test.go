package discovery

import (
	"context"
	"testing"
	"time"

	"github.com/RTradeLtd/libp2px-core/host"
	"github.com/RTradeLtd/libp2px-core/peer"
	"go.uber.org/zap/zaptest"

	bhost "github.com/RTradeLtd/libp2px/p2p/host/basic"
	swarmt "github.com/RTradeLtd/libp2px/pkg/swarm/testing"
)

type DiscoveryNotifee struct {
	h host.Host
}

func (n *DiscoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.h.Connect(context.Background(), pi)
}

func TestMdnsDiscovery(t *testing.T) {
	//TODO: re-enable when the new lib will get integrated
	t.Skip("TestMdnsDiscovery fails randomly with current lib")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s1, closer1 := swarmt.GenSwarm(t, ctx)
	defer closer1()
	s2, closer2 := swarmt.GenSwarm(t, ctx)
	defer closer2()
	a := bhost.New(ctx, s1, zaptest.NewLogger(t))
	b := bhost.New(ctx, s2, zaptest.NewLogger(t))

	sa, err := NewMdnsService(ctx, a, time.Second, zaptest.NewLogger(t), "someTag")
	if err != nil {
		t.Fatal(err)
	}

	sb, err := NewMdnsService(ctx, b, time.Second, zaptest.NewLogger(t), "someTag")
	if err != nil {
		t.Fatal(err)
	}

	_ = sb

	n := &DiscoveryNotifee{a}

	sa.RegisterNotifee(n)

	time.Sleep(time.Second * 2)

	err = a.Connect(ctx, peer.AddrInfo{ID: b.ID()})
	if err != nil {
		t.Fatal(err)
	}
}
