package pubsub

import (
	"context"

	pb "github.com/RTradeLtd/libp2px/pkg/pubsub/pb"

	"github.com/RTradeLtd/libp2px-core/host"
	"github.com/RTradeLtd/libp2px-core/peer"
	"github.com/RTradeLtd/libp2px-core/protocol"
)

const (
	RandomSubID = protocol.ID("/randomsub/1.0.0")
)

var (
	RandomSubD = 6
)

// NewRandomSub returns a new PubSub object using RandomSubRouter as the router.
func NewRandomSub(ctx context.Context, h host.Host, opts ...Option) (*PubSub, error) {
	rt := &RandomSubRouter{
		peers: make(map[peer.ID]protocol.ID),
	}
	return NewPubSub(ctx, h, rt, opts...)
}

// RandomSubRouter is a router that implements a random propagation strategy.
// For each message, it selects RandomSubD peers and forwards the message to them.
type RandomSubRouter struct {
	p      *PubSub
	peers  map[peer.ID]protocol.ID
	tracer *pubsubTracer
}

func (rs *RandomSubRouter) Protocols() []protocol.ID {
	return []protocol.ID{RandomSubID, FloodSubID}
}

func (rs *RandomSubRouter) Attach(p *PubSub) {
	rs.p = p
	rs.tracer = p.tracer
}

func (rs *RandomSubRouter) AddPeer(p peer.ID, proto protocol.ID) {
	rs.tracer.AddPeer(p, proto)
	rs.peers[p] = proto
}

func (rs *RandomSubRouter) RemovePeer(p peer.ID) {
	rs.tracer.RemovePeer(p)
	delete(rs.peers, p)
}

func (rs *RandomSubRouter) EnoughPeers(topic string, suggested int) bool {
	// check all peers in the topic
	tmap, ok := rs.p.topics[topic]
	if !ok {
		return false
	}

	fsPeers := 0
	rsPeers := 0

	// count floodsub and randomsub peers
	for p := range tmap {
		switch rs.peers[p] {
		case FloodSubID:
			fsPeers++
		case RandomSubID:
			rsPeers++
		}
	}

	if suggested == 0 {
		suggested = RandomSubD
	}

	if fsPeers+rsPeers >= suggested {
		return true
	}

	if rsPeers >= RandomSubD {
		return true
	}

	return false
}

func (rs *RandomSubRouter) HandleRPC(rpc *RPC) {}

func (rs *RandomSubRouter) Publish(from peer.ID, msg *pb.Message) {
	tosend := make(map[peer.ID]struct{})
	rspeers := make(map[peer.ID]struct{})
	src := peer.ID(msg.GetFrom())

	for _, topic := range msg.GetTopicIDs() {
		tmap, ok := rs.p.topics[topic]
		if !ok {
			continue
		}

		for p := range tmap {
			if p == from || p == src {
				continue
			}

			if rs.peers[p] == FloodSubID {
				tosend[p] = struct{}{}
			} else {
				rspeers[p] = struct{}{}
			}
		}
	}

	if len(rspeers) > RandomSubD {
		xpeers := peerMapToList(rspeers)
		shufflePeers(xpeers)
		xpeers = xpeers[:RandomSubD]
		for _, p := range xpeers {
			tosend[p] = struct{}{}
		}
	} else {
		for p := range rspeers {
			tosend[p] = struct{}{}
		}
	}

	out := rpcWithMessages(msg)
	for p := range tosend {
		mch, ok := rs.p.peers[p]
		if !ok {
			continue
		}

		select {
		case mch <- out:
			rs.tracer.SendRPC(out, p)
		default:
			rs.tracer.DropRPC(out, p)
		}
	}
}

func (rs *RandomSubRouter) Join(topic string) {
	rs.tracer.Join(topic)
}

func (rs *RandomSubRouter) Leave(topic string) {
	rs.tracer.Join(topic)
}
