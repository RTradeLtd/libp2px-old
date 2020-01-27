package pubsub

import (
	"context"

	pb "github.com/RTradeLtd/libp2px/pkg/pubsub/pb"

	"github.com/RTradeLtd/libp2px-core/host"
	"github.com/RTradeLtd/libp2px-core/peer"
	"github.com/RTradeLtd/libp2px-core/protocol"
)

const (
	FloodSubID              = protocol.ID("/floodsub/1.0.0")
	FloodSubTopicSearchSize = 5
)

// NewFloodsubWithProtocols returns a new floodsub-enabled PubSub objecting using the protocols specified in ps.
func NewFloodsubWithProtocols(ctx context.Context, h host.Host, ps []protocol.ID, opts ...Option) (*PubSub, error) {
	rt := &FloodSubRouter{
		protocols: ps,
	}
	return NewPubSub(ctx, h, rt, opts...)
}

// NewFloodSub returns a new PubSub object using the FloodSubRouter.
func NewFloodSub(ctx context.Context, h host.Host, opts ...Option) (*PubSub, error) {
	return NewFloodsubWithProtocols(ctx, h, []protocol.ID{FloodSubID}, opts...)
}

type FloodSubRouter struct {
	p         *PubSub
	protocols []protocol.ID
	tracer    *pubsubTracer
}

func (fs *FloodSubRouter) Protocols() []protocol.ID {
	return fs.protocols
}

func (fs *FloodSubRouter) Attach(p *PubSub) {
	fs.p = p
	fs.tracer = p.tracer
}

func (fs *FloodSubRouter) AddPeer(p peer.ID, proto protocol.ID) {
	fs.tracer.AddPeer(p, proto)
}

func (fs *FloodSubRouter) RemovePeer(p peer.ID) {
	fs.tracer.RemovePeer(p)
}

func (fs *FloodSubRouter) EnoughPeers(topic string, suggested int) bool {
	// check all peers in the topic
	tmap, ok := fs.p.topics[topic]
	if !ok {
		return false
	}

	if suggested == 0 {
		suggested = FloodSubTopicSearchSize
	}

	if len(tmap) >= suggested {
		return true
	}

	return false
}

func (fs *FloodSubRouter) HandleRPC(rpc *RPC) {}

func (fs *FloodSubRouter) Publish(from peer.ID, msg *pb.Message) {
	tosend := make(map[peer.ID]struct{})
	for _, topic := range msg.GetTopicIDs() {
		tmap, ok := fs.p.topics[topic]
		if !ok {
			continue
		}

		for p := range tmap {
			tosend[p] = struct{}{}
		}
	}

	out := rpcWithMessages(msg)
	for pid := range tosend {
		if pid == from || pid == peer.ID(msg.GetFrom()) {
			continue
		}

		mch, ok := fs.p.peers[pid]
		if !ok {
			continue
		}

		select {
		case mch <- out:
			fs.tracer.SendRPC(out, pid)
		default:
			fs.tracer.DropRPC(out, pid)
			// Drop it. The peer is too slow.
		}
	}
}

func (fs *FloodSubRouter) Join(topic string) {
	fs.tracer.Join(topic)
}

func (fs *FloodSubRouter) Leave(topic string) {
	fs.tracer.Join(topic)
}
