package basichost

import (
	"context"
	"net"
	"strconv"
	"sync"

	"github.com/RTradeLtd/libp2p-core/network"
	inat "github.com/RTradeLtd/libp2px/pkg/utils/nat"
	ma "github.com/multiformats/go-multiaddr"
)

// NATManager is a simple interface to manage NAT devices.
type NATManager interface {

	// Get the NAT device managed by the NAT manager.
	NAT() *inat.NAT

	// Receive a notification when the NAT device is ready for use.
	Ready() <-chan struct{}

	// Close all resources associated with a NAT manager.
	Close() error
}

// NewNATManager returns a newly constructed NATManager
func NewNATManager(ctx context.Context, net network.Network) NATManager {
	return newNatManager(ctx, net)
}

// natManager takes care of adding + removing port mappings to the nat.
// Initialized with the host if it has a NATPortMap option enabled.
// natManager receives signals from the network, and check on nat mappings:
//  * natManager listens to the network and adds or closes port mappings
//    as the network signals Listen() or ListenClose().
//  * closing the natManager closes the nat and its mappings.
type natManager struct {
	net   network.Network
	natmu sync.RWMutex
	nat   *inat.NAT

	ready chan struct{} // closed once the nat is ready to process port mappings

	syncMu sync.Mutex
}

func newNatManager(ctx context.Context, net network.Network) *natManager {
	nmgr := &natManager{
		net:   net,
		ready: make(chan struct{}),
	}
	// discover the nat.
	nmgr.discoverNAT(ctx)
	return nmgr
}

// Close closes the natManager, closing the underlying nat
// and unregistering from network events.
func (nmgr *natManager) Close() error {
	nmgr.net.StopNotify((*nmgrNetNotifiee)(nmgr))
	return nil
}

// Ready returns a channel which will be closed when the NAT has been found
// and is ready to be used, or the search process is done.
func (nmgr *natManager) Ready() <-chan struct{} {
	return nmgr.ready
}

func (nmgr *natManager) discoverNAT(ctx context.Context) {
	go func() {
		// inat.DiscoverNAT blocks until the nat is found or a timeout
		// is reached. we unfortunately cannot specify timeouts-- the
		// library we're using just blocks.
		//
		// Note: on early shutdown, there may be a case where we're trying
		// to close before DiscoverNAT() returns. Since we cant cancel it
		// (library) we can choose to (1) drop the result and return early,
		// or (2) wait until it times out to exit. For now we choose (2),
		// to avoid leaking resources in a non-obvious way. the only case
		// this affects is when the daemon is being started up and _immediately_
		// asked to close. other services are also starting up, so ok to wait.

		natInstance, err := inat.DiscoverNAT(ctx)
		if err != nil {
			close(nmgr.ready)
			return
		}

		nmgr.natmu.Lock()
		nmgr.nat = natInstance
		nmgr.natmu.Unlock()
		close(nmgr.ready)
		// sign natManager up for network notifications
		// we need to sign up here to avoid missing some notifs
		// before the NAT has been found.
		nmgr.net.Notify((*nmgrNetNotifiee)(nmgr))
		nmgr.sync()
	}()
}

// syncs the current NAT mappings, removing any outdated mappings and adding any
// new mappings.
func (nmgr *natManager) sync() {
	nat := nmgr.NAT()
	if nat == nil {
		// Nothing to do.
		return
	}
	go func() {
		nmgr.syncMu.Lock()
		defer nmgr.syncMu.Unlock()

		ports := map[string]map[int]bool{
			"tcp": {},
			"udp": {},
		}
		for _, maddr := range nmgr.net.ListenAddresses() {
			// Strip the IP
			maIP, rest := ma.SplitFirst(maddr)
			if maIP == nil || rest == nil {
				continue
			}

			switch maIP.Protocol().Code {
			case ma.P_IP6, ma.P_IP4:
			default:
				continue
			}

			// Only bother if we're listening on a
			// unicast/unspecified IP.
			ip := net.IP(maIP.RawValue())
			if !(ip.IsGlobalUnicast() || ip.IsUnspecified()) {
				continue
			}

			// Extract the port/protocol
			proto, _ := ma.SplitFirst(rest)
			if proto == nil {
				continue
			}

			var protocol string
			switch proto.Protocol().Code {
			case ma.P_TCP:
				protocol = "tcp"
			case ma.P_UDP:
				protocol = "udp"
			default:
				continue
			}

			port, err := strconv.ParseUint(proto.Value(), 10, 16)
			if err != nil {
				// bug in multiaddr
				panic(err)
			}
			ports[protocol][int(port)] = false
		}

		var wg sync.WaitGroup
		defer wg.Wait()

		// Close old mappings
		for _, m := range nat.Mappings() {
			mappedPort := m.InternalPort()
			if _, ok := ports[m.Protocol()][mappedPort]; !ok {
				// No longer need this mapping.
				wg.Add(1)
				go func(m inat.Mapping) {
					defer wg.Done()
					m.Close()
				}(m)
			} else {
				// already mapped
				ports[m.Protocol()][mappedPort] = true
			}
		}

		// Create new mappings.
		for proto, pports := range ports {
			for port, mapped := range pports {
				if mapped {
					continue
				}
				wg.Add(1)
				go func(proto string, port int) {
					defer wg.Done()
					_, _ = nat.NewMapping(proto, port)
				}(proto, port)
			}
		}
	}()
}

// NAT returns the natManager's nat object. this may be nil, if
// (a) the search process is still ongoing, or (b) the search process
// found no nat. Clients must check whether the return value is nil.
func (nmgr *natManager) NAT() *inat.NAT {
	nmgr.natmu.Lock()
	defer nmgr.natmu.Unlock()
	return nmgr.nat
}

type nmgrNetNotifiee natManager

func (nn *nmgrNetNotifiee) natManager() *natManager {
	return (*natManager)(nn)
}

func (nn *nmgrNetNotifiee) Listen(n network.Network, addr ma.Multiaddr) {
	nn.natManager().sync()
}

func (nn *nmgrNetNotifiee) ListenClose(n network.Network, addr ma.Multiaddr) {
	nn.natManager().sync()
}

func (nn *nmgrNetNotifiee) Connected(network.Network, network.Conn)      {}
func (nn *nmgrNetNotifiee) Disconnected(network.Network, network.Conn)   {}
func (nn *nmgrNetNotifiee) OpenedStream(network.Network, network.Stream) {}
func (nn *nmgrNetNotifiee) ClosedStream(network.Network, network.Stream) {}
