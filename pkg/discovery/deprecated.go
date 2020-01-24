package discovery

import (
	"time"

	core "github.com/RTradeLtd/libp2px-core/discovery"
)

// Advertiser is Deprecated: use github.com/RTradeLtd/libp2px-core.Advertiser instead.
type Advertiser = core.Advertiser

// Discoverer Deprecated: use github.com/RTradeLtd/libp2px-core.Discoverer instead.
type Discoverer = core.Discoverer

// Discovery is Deprecated: use github.com/RTradeLtd/libp2px-core.Discovery instead.
type Discovery = core.Discovery

// Option is Deprecated: use github.com/RTradeLtd/libp2px-core/discovery.Option instead.
type Option = core.Option

// Options is Deprecated: use github.com/RTradeLtd/libp2px-core/discovery.Options instead.
type Options = core.Options

// TTL is Deprecated: use github.com/RTradeLtd/libp2px-core/discovery.TTL instead.
func TTL(ttl time.Duration) core.Option {
	return core.TTL(ttl)
}

// Limit is Deprecated: use github.com/RTradeLtd/libp2px-core/discovery.Limit instead.
func Limit(limit int) core.Option {
	return core.Limit(limit)
}
