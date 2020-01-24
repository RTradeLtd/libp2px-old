package pubsub

import (
	"fmt"

	pb "github.com/RTradeLtd/libp2p-pubsub/pb"
)

// NewMessageCache creates a sliding window cache that remembers messages for as
// long as `history` slots.
//
// When queried for messages to advertise, the cache only returns messages in
// the last `gossip` slots.
//
// The `gossip` parameter must be smaller or equal to `history`, or this
// function will panic.
//
// The slack between `gossip` and `history` accounts for the reaction time
// between when a message is advertised via IHAVE gossip, and the peer pulls it
// via an IWANT command.
func NewMessageCache(gossip, history int) *MessageCache {
	if gossip > history {
		err := fmt.Errorf("invalid parameters for message cache; gossip slots (%d) cannot be larger than history slots (%d)",
			gossip, history)
		panic(err)
	}
	return &MessageCache{
		msgs:    make(map[string]*pb.Message),
		history: make([][]CacheEntry, history),
		gossip:  gossip,
		msgID:   DefaultMsgIdFn,
	}
}

type MessageCache struct {
	msgs    map[string]*pb.Message
	history [][]CacheEntry
	gossip  int
	msgID   MsgIdFunction
}

func (mc *MessageCache) SetMsgIdFn(msgID MsgIdFunction) {
	mc.msgID = msgID
}

type CacheEntry struct {
	mid    string
	topics []string
}

func (mc *MessageCache) Put(msg *pb.Message) {
	mid := mc.msgID(msg)
	mc.msgs[mid] = msg
	mc.history[0] = append(mc.history[0], CacheEntry{mid: mid, topics: msg.GetTopicIDs()})
}

func (mc *MessageCache) Get(mid string) (*pb.Message, bool) {
	m, ok := mc.msgs[mid]
	return m, ok
}

func (mc *MessageCache) GetGossipIDs(topic string) []string {
	var mids []string
	for _, entries := range mc.history[:mc.gossip] {
		for _, entry := range entries {
			for _, t := range entry.topics {
				if t == topic {
					mids = append(mids, entry.mid)
					break
				}
			}
		}
	}
	return mids
}

func (mc *MessageCache) Shift() {
	last := mc.history[len(mc.history)-1]
	for _, entry := range last {
		delete(mc.msgs, entry.mid)
	}
	for i := len(mc.history) - 2; i >= 0; i-- {
		mc.history[i+1] = mc.history[i]
	}
	mc.history[0] = nil
}
