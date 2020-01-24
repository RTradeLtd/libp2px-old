package autonat

import (
	pb "github.com/RTradeLtd/libp2px/pkg/autonat/pb"

	"github.com/RTradeLtd/libp2px-core/peer"
)

// AutoNATProto is the libp2p autonat protocol name
const AutoNATProto = "/libp2p/autonat/1.0.0"

func newDialMessage(pi peer.AddrInfo) *pb.Message {
	msg := new(pb.Message)
	msg.Type = pb.Message_DIAL.Enum()
	msg.Dial = new(pb.Message_Dial)
	msg.Dial.Peer = new(pb.Message_PeerInfo)
	msg.Dial.Peer.Id = []byte(pi.ID)
	msg.Dial.Peer.Addrs = make([][]byte, len(pi.Addrs))
	for i, addr := range pi.Addrs {
		msg.Dial.Peer.Addrs[i] = addr.Bytes()
	}

	return msg
}
