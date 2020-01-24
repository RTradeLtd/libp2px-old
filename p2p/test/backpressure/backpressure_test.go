package backpressure_tests

import (
	"context"
	"io"
	"math/rand"
	"testing"
	"time"

	bhost "github.com/RTradeLtd/libp2px/p2p/host/basic"
	"go.uber.org/zap/zaptest"

	"github.com/RTradeLtd/libp2px-core/host"
	"github.com/RTradeLtd/libp2px-core/network"
	"github.com/RTradeLtd/libp2px-core/peer"
	protocol "github.com/RTradeLtd/libp2px-core/protocol"
	swarmt "github.com/RTradeLtd/libp2px/pkg/swarm/testing"
	u "github.com/ipfs/go-ipfs-util"
)

// TestBackpressureStreamHandler tests whether mux handler
// ratelimiting works. Meaning, since the handler is sequential
// it should block senders.
//
// Important note: spdystream (which peerstream uses) has a set
// of n workers (n=spdsystream.FRAME_WORKERS) which handle new
// frames, including those starting new streams. So all of them
// can be in the handler at one time. Also, the sending side
// does not rate limit unless we call stream.Wait()
//
//
// Note: right now, this happens muxer-wide. the muxer should
// learn to flow control, so handlers cant block each other.
func TestBackpressureStreamHandler(t *testing.T) {
	t.Skip(`Sadly, as cool as this test is, it doesn't work
Because spdystream doesnt handle stream open backpressure
well IMO. I'll see about rewriting that part when it becomes
a problem.
`)

	// a number of concurrent request handlers
	limit := 10

	// our way to signal that we're done with 1 request
	requestHandled := make(chan struct{})

	// handler rate limiting
	receiverRatelimit := make(chan struct{}, limit)
	for i := 0; i < limit; i++ {
		receiverRatelimit <- struct{}{}
	}

	// sender counter of successfully opened streams
	senderOpened := make(chan struct{}, limit*100)

	// sender signals it's done (errored out)
	senderDone := make(chan struct{})

	// the receiver handles requests with some rate limiting
	receiver := func(s network.Stream) {

		<-receiverRatelimit // acquire
		go func() {
			// our request handler. can do stuff here. we
			// simulate something taking time by waiting
			// on requestHandled
			<-requestHandled
			receiverRatelimit <- struct{}{} // release
		}()
	}

	// the sender opens streams as fast as possible
	sender := func(host host.Host, remote peer.ID) {
		var s network.Stream
		var err error
		defer func() {
			t.Error(err)
			senderDone <- struct{}{}
		}()

		for {
			s, err = host.NewStream(context.Background(), remote, protocol.TestingID)
			if err != nil {
				return
			}

			_ = s
			// if err = s.SwarmStream().Stream().Wait(); err != nil {
			// 	return
			// }

			// "count" another successfully opened stream
			// (large buffer so shouldn't block in normal operation)
			senderOpened <- struct{}{}
		}
	}

	// count our senderOpened events
	countStreamsOpenedBySender := func(min int) int {
		opened := 0
		for opened < min {
			select {
			case <-senderOpened:
				opened++
			case <-time.After(10 * time.Millisecond):
			}
		}
		return opened
	}

	// count our received events
	// waitForNReceivedStreams := func(n int) {
	// 	for n > 0 {
	// 		log.Debugf("waiting for %d received streams...", n)
	// 		select {
	// 		case <-receiverRatelimit:
	// 			n--
	// 		}
	// 	}
	// }

	testStreamsOpened := func(expected int) {
		if n := countStreamsOpenedBySender(expected); n != expected {
			t.Fatalf("rate limiting did not work :( -- %d != %d", expected, n)
		}
	}

	// ok that's enough setup. let's do it!

	ctx := context.Background()
	s1, closer1 := swarmt.GenSwarm(t, ctx)
	defer closer1()
	s2, closer2 := swarmt.GenSwarm(t, ctx)
	defer closer2()
	h1 := bhost.New(ctx, s1, zaptest.NewLogger(t))
	h2 := bhost.New(ctx, s2, zaptest.NewLogger(t))

	// setup receiver handler
	h1.SetStreamHandler(protocol.TestingID, receiver)

	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	if err := h1.Connect(ctx, h2pi); err != nil {
		t.Fatal("Failed to connect:", err)
	}

	// launch sender!
	go sender(h2, h1.ID())

	// ok, what do we expect to happen? the receiver should
	// receive 10 requests and stop receiving, blocking the sender.
	// we can test this by counting 10x senderOpened requests

	<-senderOpened // wait for the sender to successfully open some.
	testStreamsOpened(limit - 1)

	// let's "handle" 3 requests.
	<-requestHandled
	<-requestHandled
	<-requestHandled
	// the sender should've now been able to open exactly 3 more.

	testStreamsOpened(3)

	// shouldn't have opened anything more
	testStreamsOpened(0)

	// let's "handle" 100 requests in batches of 5
	for i := 0; i < 20; i++ {
		<-requestHandled
		<-requestHandled
		<-requestHandled
		<-requestHandled
		<-requestHandled
		testStreamsOpened(5)
	}

	// success!

	// now for the sugar on top: let's tear down the receiver. it should
	// exit the sender.
	h1.Close()

	// shouldn't have opened anything more
	testStreamsOpened(0)

	select {
	case <-time.After(100 * time.Millisecond):
		t.Error("receiver shutdown failed to exit sender")
	case <-senderDone:
	}
}

// TestStBackpressureStreamWrite tests whether streams see proper
// backpressure when writing data over the network streams.
func TestStBackpressureStreamWrite(t *testing.T) {

	// senderWrote signals that the sender wrote bytes to remote.
	// the value is the count of bytes written.
	senderWrote := make(chan int, 10000)

	// sender signals it's done (errored out)
	senderDone := make(chan struct{})

	// writeStats lets us listen to all the writes and return
	// how many happened and how much was written
	writeStats := func() (int, int) {
		t.Helper()
		writes := 0
		bytes := 0
		for {
			select {
			case n := <-senderWrote:
				writes++
				bytes += n
			default:
				return bytes, writes
			}
		}
	}

	// sender attempts to write as fast as possible, signaling on the
	// completion of every write. This makes it possible to see how
	// fast it's actually writing. We pair this with a receiver
	// that waits for a signal to read.
	sender := func(s network.Stream) {
		defer func() {
			s.Close()
			senderDone <- struct{}{}
		}()

		// ready a buffer of random data
		buf := make([]byte, 65536)
		u.NewTimeSeededRand().Read(buf)

		for {
			// send a randomly sized subchunk
			from := rand.Intn(len(buf) / 2)
			to := rand.Intn(len(buf) / 2)
			sendbuf := buf[from : from+to]

			n, err := s.Write(sendbuf)
			if err != nil {
				return
			}

			select {
			case senderWrote <- n:
			default:
				t.Error("sender wrote channel full")
			}
		}
	}

	// receive a number of bytes from a stream.
	// returns the number of bytes written.
	receive := func(s network.Stream, expect int) {
		t.Helper()
		rbuf := make([]byte, expect)
		n, err := io.ReadFull(s, rbuf)
		if err != nil {
			t.Error("read failed:", err)
		}
		if expect != n {
			t.Errorf("read len differs: %d != %d", expect, n)
		}
	}

	// ok let's do it!

	// setup the networks
	ctx := context.Background()
	s1, closer1 := swarmt.GenSwarm(t, ctx)
	defer closer1()
	s2, closer2 := swarmt.GenSwarm(t, ctx)
	defer closer2()
	h1 := bhost.New(ctx, s1, zaptest.NewLogger(t))
	h2 := bhost.New(ctx, s2, zaptest.NewLogger(t))

	// setup sender handler on 1
	h1.SetStreamHandler(protocol.TestingID, sender)

	h2pi := h2.Peerstore().PeerInfo(h2.ID())
	if err := h1.Connect(ctx, h2pi); err != nil {
		t.Fatal("Failed to connect:", err)
	}

	// open a stream, from 2->1, this is our reader
	s, err := h2.NewStream(context.Background(), h1.ID(), protocol.TestingID)
	if err != nil {
		t.Fatal(err)
	}

	// let's make sure r/w works.
	testSenderWrote := func(bytesE int) {
		t.Helper()
		bytesA, writesA := writeStats()
		if bytesA != bytesE {
			t.Errorf("numbers failed: %d =?= %d bytes, via %d writes", bytesA, bytesE, writesA)
		}
	}

	// trigger lazy connection handshaking
	_, err = s.Read(nil)
	if err != nil {
		t.Fatal(err)
	}

	// 500ms rounds of lockstep write + drain
	roundsStart := time.Now()
	roundsTotal := 0
	for roundsTotal < (2 << 20) {
		// let the sender fill its buffers, it will stop sending.
		<-time.After(time.Second)
		b, _ := writeStats()
		testSenderWrote(0)
		<-time.After(100 * time.Millisecond)
		testSenderWrote(0)

		// drain it all, wait again
		receive(s, b)
		roundsTotal += b
	}
	roundsTime := time.Since(roundsStart)

	// now read continuously, while we measure stats.
	stop := make(chan struct{})
	contStart := time.Now()

	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				receive(s, 2<<15)
			}
		}
	}()

	contTotal := 0
	for contTotal < (2 << 20) {
		n := <-senderWrote
		contTotal += n
	}
	stop <- struct{}{}
	contTime := time.Since(contStart)

	// now compare! continuous should've been faster AND larger
	if roundsTime < contTime {
		t.Error("continuous should have been faster")
	}

	if roundsTotal < contTotal {
		t.Error("continuous should have been larger, too!")
	}

	// and a couple rounds more for good measure ;)
	for i := 0; i < 3; i++ {
		// let the sender fill its buffers, it will stop sending.
		<-time.After(time.Second)
		b, _ := writeStats()
		testSenderWrote(0)
		<-time.After(100 * time.Millisecond)
		testSenderWrote(0)

		// drain it all, wait again
		receive(s, b)
	}

	// this doesn't work :(:
	// // now for the sugar on top: let's tear down the receiver. it should
	// // exit the sender.
	// n1.Close()
	// testSenderWrote(0)
	// testSenderWrote(0)
	// select {
	// case <-time.After(2 * time.Second):
	// 	t.Error("receiver shutdown failed to exit sender")
	// case <-senderDone:
	// 	log.Info("handler backpressure works!")
	// }
}
