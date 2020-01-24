package discovery

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/RTradeLtd/libp2p-core/host"
	"github.com/RTradeLtd/libp2p-core/peer"
	"go.uber.org/zap"

	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
	"github.com/whyrusleeping/mdns"
)

func init() {
	// don't let mdns use logging...
	mdns.DisableLogging = true
}

// ServiceTag specifies the mdns service tag
const ServiceTag = "_ipfs-discovery._udp"

// Service is an interface that one must satisfy to provide mdns discovery services
type Service interface {
	io.Closer
	RegisterNotifee(Notifee)
	UnregisterNotifee(Notifee)
}

// Notifee is an interface that allows notifying other goroutines that we found a peer
type Notifee interface {
	HandlePeerFound(peer.AddrInfo)
}

type mdnsService struct {
	server  *mdns.Server
	service *mdns.MDNSService
	host    host.Host
	tag     string

	lk       sync.Mutex
	notifees []Notifee
	interval time.Duration

	logger *zap.Logger
}

func getDialableListenAddrs(ph host.Host) ([]*net.TCPAddr, error) {
	var out []*net.TCPAddr
	addrs, err := ph.Network().InterfaceListenAddresses()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		na, err := manet.ToNetAddr(addr)
		if err != nil {
			continue
		}
		tcp, ok := na.(*net.TCPAddr)
		if ok {
			out = append(out, tcp)
		}
	}
	if len(out) == 0 {
		return nil, errors.New("failed to find good external addr from peerhost")
	}
	return out, nil
}

// NewMdnsService starts a new mdns service
func NewMdnsService(ctx context.Context, peerhost host.Host, interval time.Duration, logger *zap.Logger, serviceTag string) (Service, error) {

	var ipaddrs []net.IP
	port := 4001

	addrs, err := getDialableListenAddrs(peerhost)
	if err != nil {
		return nil, err
	}
	port = addrs[0].Port
	for _, a := range addrs {
		ipaddrs = append(ipaddrs, a.IP)
	}

	myid := peerhost.ID().Pretty()

	info := []string{myid}
	if serviceTag == "" {
		serviceTag = ServiceTag
	}
	service, err := mdns.NewMDNSService(myid, serviceTag, "", "", port, ipaddrs, info)
	if err != nil {
		return nil, err
	}

	// Create the mDNS server, defer shutdown
	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		return nil, err
	}

	s := &mdnsService{
		server:   server,
		service:  service,
		host:     peerhost,
		interval: interval,
		tag:      serviceTag,
		logger:   logger.Named("mdns"),
	}

	go s.pollForEntries(ctx)

	return s, nil
}

func (m *mdnsService) Close() error {
	return m.server.Shutdown()
}

func (m *mdnsService) pollForEntries(ctx context.Context) {

	ticker := time.NewTicker(m.interval)
	for {
		//execute mdns query right away at method call and then with every tick
		entriesCh := make(chan *mdns.ServiceEntry, 16)
		go func() {
			for entry := range entriesCh {
				m.handleEntry(entry)
			}
		}()
		qp := &mdns.QueryParam{
			Domain:  "local",
			Entries: entriesCh,
			Service: m.tag,
			Timeout: time.Second * 5,
		}

		err := mdns.Query(qp)
		if err != nil {
			m.logger.Error("mdns lookup failure", zap.Error(err))
		}
		close(entriesCh)

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (m *mdnsService) handleEntry(e *mdns.ServiceEntry) {
	mpeer, err := peer.IDB58Decode(e.Info)
	if err != nil {
		m.logger.Warn("failed parsig mdns entry peer id", zap.Error(err))
		return
	}

	if mpeer == m.host.ID() {
		return
	}

	maddr, err := manet.FromNetAddr(&net.TCPAddr{
		IP:   e.AddrV4,
		Port: e.Port,
	})
	if err != nil {
		m.logger.Warn("failed parsing mdns entry multiaddr", zap.Error(err))
		return
	}

	pi := peer.AddrInfo{
		ID:    mpeer,
		Addrs: []ma.Multiaddr{maddr},
	}

	m.lk.Lock()
	for _, n := range m.notifees {
		go n.HandlePeerFound(pi)
	}
	m.lk.Unlock()
}

func (m *mdnsService) RegisterNotifee(n Notifee) {
	m.lk.Lock()
	m.notifees = append(m.notifees, n)
	m.lk.Unlock()
}

func (m *mdnsService) UnregisterNotifee(n Notifee) {
	m.lk.Lock()
	found := -1
	for i, notif := range m.notifees {
		if notif == n {
			found = i
			break
		}
	}
	if found != -1 {
		m.notifees = append(m.notifees[:found], m.notifees[found+1:]...)
	}
	m.lk.Unlock()
}
