package libp2pquic

import (
	"github.com/RTradeLtd/libp2px-core/mux"

	quic "github.com/lucas-clemente/quic-go"
)

type stream struct {
	quic.Stream
}

var _ mux.MuxedStream = &stream{}

func (s *stream) Reset() error {
	s.Stream.CancelRead(0)
	s.Stream.CancelWrite(0)
	return nil
}
