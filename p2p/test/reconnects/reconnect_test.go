package reconnect

import (
	"context"
	"io"
	"math/rand"
	"sync"
	"testing"
	"time"

	bhost "github.com/RTradeLtd/libp2px/p2p/host/basic"
	"go.uber.org/zap/zaptest"

	"github.com/RTradeLtd/libp2px-core/host"
	"github.com/RTradeLtd/libp2px-core/network"
	protocol "github.com/RTradeLtd/libp2px-core/protocol"
	swarmt "github.com/RTradeLtd/libp2px/pkg/swarm/testing"
	u "github.com/ipfs/go-ipfs-util"
)

func EchoStreamHandler(stream network.Stream) {
	go func() {
		_, err := io.Copy(stream, stream)
		if err == nil {
			stream.Close()
		} else {
			stream.Reset()
		}
	}()
}

type sendChans struct {
	send   chan struct{}
	sent   chan struct{}
	read   chan struct{}
	close_ chan struct{}
	closed chan struct{}
}

func newSendChans() sendChans {
	return sendChans{
		send:   make(chan struct{}),
		sent:   make(chan struct{}),
		read:   make(chan struct{}),
		close_: make(chan struct{}),
		closed: make(chan struct{}),
	}
}

func newSender() (chan sendChans, func(s network.Stream)) {
	scc := make(chan sendChans)
	return scc, func(s network.Stream) {
		sc := newSendChans()
		scc <- sc

		defer func() {
			if s != nil {
				s.Close()
			}
			sc.closed <- struct{}{}
		}()

		buf := make([]byte, 65536)
		buf2 := make([]byte, 65536)
		u.NewTimeSeededRand().Read(buf)

		for {
			select {
			case <-sc.close_:
				return
			case <-sc.send:
			}

			// send a randomly sized subchunk
			from := rand.Intn(len(buf) / 2)
			to := rand.Intn(len(buf) / 2)
			sendbuf := buf[from : from+to]

			_, err := s.Write(sendbuf)
			if err != nil {
				return
			}

			sc.sent <- struct{}{}

			if _, err = io.ReadFull(s, buf2[:len(sendbuf)]); err != nil {
				return
			}

			sc.read <- struct{}{}
		}
	}
}

// TestReconnect tests whether hosts are able to disconnect and reconnect.
func TestReconnect2(t *testing.T) {
	ctx := context.Background()
	s1, closer1 := swarmt.GenSwarm(t, ctx)
	defer closer1()
	s2, closer2 := swarmt.GenSwarm(t, ctx)
	defer closer2()
	h1 := bhost.New(ctx, s1, zaptest.NewLogger(t))
	h2 := bhost.New(ctx, s2, zaptest.NewLogger(t))
	hosts := []host.Host{h1, h2}

	h1.SetStreamHandler(protocol.TestingID, EchoStreamHandler)
	h2.SetStreamHandler(protocol.TestingID, EchoStreamHandler)

	rounds := 8
	if testing.Short() {
		rounds = 4
	}
	for i := 0; i < rounds; i++ {
		SubtestConnSendDisc(t, hosts)
	}
}

// TestReconnect tests whether hosts are able to disconnect and reconnect.
func TestReconnect5(t *testing.T) {
	ctx := context.Background()
	s1, closer1 := swarmt.GenSwarm(t, ctx)
	defer closer1()
	s2, closer2 := swarmt.GenSwarm(t, ctx)
	defer closer2()
	s3, closer3 := swarmt.GenSwarm(t, ctx)
	defer closer3()
	s4, closer4 := swarmt.GenSwarm(t, ctx)
	defer closer4()
	s5, closer5 := swarmt.GenSwarm(t, ctx)
	defer closer5()
	h1 := bhost.New(ctx, s1, zaptest.NewLogger(t))
	h2 := bhost.New(ctx, s2, zaptest.NewLogger(t))
	h3 := bhost.New(ctx, s3, zaptest.NewLogger(t))
	h4 := bhost.New(ctx, s4, zaptest.NewLogger(t))
	h5 := bhost.New(ctx, s5, zaptest.NewLogger(t))
	hosts := []host.Host{h1, h2, h3, h4, h5}

	h1.SetStreamHandler(protocol.TestingID, EchoStreamHandler)
	h2.SetStreamHandler(protocol.TestingID, EchoStreamHandler)
	h3.SetStreamHandler(protocol.TestingID, EchoStreamHandler)
	h4.SetStreamHandler(protocol.TestingID, EchoStreamHandler)
	h5.SetStreamHandler(protocol.TestingID, EchoStreamHandler)

	rounds := 4
	if testing.Short() {
		rounds = 2
	}
	for i := 0; i < rounds; i++ {
		SubtestConnSendDisc(t, hosts)
	}
}

func SubtestConnSendDisc(t *testing.T, hosts []host.Host) {

	ctx := context.Background()
	numStreams := 3 * len(hosts)
	numMsgs := 10

	if testing.Short() {
		numStreams = 5 * len(hosts)
		numMsgs = 4
	}

	ss, sF := newSender()

	for _, h1 := range hosts {
		for _, h2 := range hosts {
			if h1.ID() >= h2.ID() {
				continue
			}

			h2pi := h2.Peerstore().PeerInfo(h2.ID())
			if err := h1.Connect(ctx, h2pi); err != nil {
				t.Fatal("Failed to connect:", err)
			}
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < numStreams; i++ {
		h1 := hosts[i%len(hosts)]
		h2 := hosts[(i+1)%len(hosts)]
		s, err := h1.NewStream(context.Background(), h2.ID(), protocol.TestingID)
		if err != nil {
			t.Error(err)
		}

		wg.Add(1)
		go func(j int) {
			defer wg.Done()

			go sF(s)
			sc := <-ss // wait to get handle.

			for k := 0; k < numMsgs; k++ {
				sc.send <- struct{}{}
				<-sc.sent
				<-sc.read
			}
			sc.close_ <- struct{}{}
			<-sc.closed
		}(i)
	}
	wg.Wait()

	for _, h1 := range hosts {
		// close connection
		cs := h1.Network().Conns()
		for _, c := range cs {
			if c.LocalPeer() > c.RemotePeer() {
				continue
			}
			c.Close()
		}
	}

	<-time.After(20 * time.Millisecond)

	for i, h := range hosts {
		if len(h.Network().Conns()) > 0 {
			t.Fatalf("host %d %s has %d conns! not zero.", i, h.ID(), len(h.Network().Conns()))
		}
	}
}
