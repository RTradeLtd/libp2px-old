package nat

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	nat "github.com/RTradeLtd/libp2px/pkg/nat"
)

var (
	// ErrNoMapping signals no mapping exists for an address
	ErrNoMapping = errors.New("mapping not established")
)

// MappingDuration is a default port mapping duration.
// Port mappings are renewed every (MappingDuration / 3)
const MappingDuration = time.Second * 60

// CacheTime is the time a mapping will cache an external address for
const CacheTime = time.Second * 15

// DiscoverNAT looks for a NAT device in the network and
// returns an object that can manage port mappings.
func DiscoverNAT(ctx context.Context) (*NAT, error) {
	var (
		natInstance nat.NAT
		err         error
	)

	done := make(chan struct{})
	go func() {
		defer close(done)
		// This will abort in 10 seconds anyways.
		natInstance, err = nat.DiscoverGateway()
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	if err != nil {
		return nil, err
	}

	return newNAT(ctx, natInstance), nil
}

// NAT is an object that manages address port mappings in
// NATs (Network Address Translators). It is a long-running
// service that will periodically renew port mappings,
// and keep an up-to-date list of all the external addresses.
type NAT struct {
	natmu     sync.Mutex
	nat       nat.NAT
	ctx       context.Context
	cancel    context.CancelFunc
	mappingmu sync.RWMutex // guards mappings
	mappings  map[*mapping]struct{}
}

func newNAT(ctx context.Context, realNAT nat.NAT) *NAT {
	ctx, cancel := context.WithCancel(ctx)
	return &NAT{
		nat:      realNAT,
		ctx:      ctx,
		cancel:   cancel,
		mappings: make(map[*mapping]struct{}),
	}
}

// Close shuts down all port mappings. NAT can no longer be used.
func (nat *NAT) Close() error {
	nat.mappingmu.Lock()
	nat.cancel()
	for mp := range nat.mappings {
		nat.rmMapping(mp)
		mp.nat.natmu.Lock()
		mp.nat.nat.DeletePortMapping(mp.Protocol(), mp.InternalPort())
		mp.Close()
		// this might need to be done before mp.CLose()
		mp.nat.natmu.Unlock()
	}
	nat.mappingmu.Unlock()
	return nil
}

// Context is used to return the nat managers context
func (nat *NAT) Context() context.Context {
	return nat.ctx
}

// Mappings returns a slice of all NAT mappings
func (nat *NAT) Mappings() []Mapping {
	nat.mappingmu.Lock()
	maps2 := make([]Mapping, 0, len(nat.mappings))
	for m := range nat.mappings {
		maps2 = append(maps2, m)
	}
	nat.mappingmu.Unlock()
	return maps2
}

func (nat *NAT) addMapping(m *mapping) {
	nat.mappingmu.Lock()
	nat.mappings[m] = struct{}{}
	nat.mappingmu.Unlock()
}

func (nat *NAT) rmMapping(m *mapping) {
	nat.mappingmu.Lock()
	delete(nat.mappings, m)
	nat.mappingmu.Unlock()
}

// NewMapping attempts to construct a mapping on protocol and internal port
// It will also periodically renew the mapping until the returned Mapping
// -- or its parent NAT -- is Closed.
//
// May not succeed, and mappings may change over time;
// NAT devices may not respect our port requests, and even lie.
// Clients should not store the mapped results, but rather always
// poll our object for the latest mappings.
func (nat *NAT) NewMapping(protocol string, port int) (Mapping, error) {
	if nat == nil {
		return nil, fmt.Errorf("no nat available")
	}

	switch protocol {
	case "tcp", "udp":
	default:
		return nil, fmt.Errorf("invalid protocol: %s", protocol)
	}

	m := &mapping{
		intport: port,
		nat:     nat,
		proto:   protocol,
	}

	nat.addMapping(m)

	/* TODO(bonedaddy): replace with cronjob style function
	m.proc.AddChild(periodic.Every(MappingDuration/3, func(worker goprocess.Process) {
		nat.establishMapping(m)
	}))
	*/
	// do it once synchronously, so first mapping is done right away, and before exiting,
	// allowing users -- in the optimistic case -- to use results right after.
	nat.establishMapping(m)
	return m, nil
}

func (nat *NAT) establishMapping(m *mapping) {
	oldport := m.ExternalPort()

	comment := "libp2p"

	nat.natmu.Lock()
	newport, err := nat.nat.AddPortMapping(m.Protocol(), m.InternalPort(), comment, MappingDuration)
	if err != nil {
		// Some hardware does not support mappings with timeout, so try that
		newport, err = nat.nat.AddPortMapping(m.Protocol(), m.InternalPort(), comment, 0)
	}
	nat.natmu.Unlock()

	if err != nil || newport == 0 {
		m.setExternalPort(0) // clear mapping
		// we do not close if the mapping failed,
		// because it may work again next time.
		return
	}

	m.setExternalPort(newport)
	if oldport != 0 && newport != oldport {
	}
}
