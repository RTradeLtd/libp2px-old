# libp2px

`libp2px` is a fork of `github.com/libp2p/go-libp2p`, intended for use with TemporalX's enterprise IPFS node. As such it is a suitable libp2p library for projects that want a more performant and well tested libp2p codebase

# Project Goals

* A more maintainable and approachable codebase 
* Thoroughly tested code base
* Performance and efficiency
* Privacy as long as it doesn't compromise performance
  * To this end we have disabled the built-in identify, and ping service. Eventually these will be available as modules

We will try to maintain compatability with `go-libp2p` as much as possible, but we expect certain things will not work.

# Compatability Issues

## Confirmed

None

## Suspected

One possible compatability issue is with secp256k1 keys being incompatible between libp2px and go-libp2p.

# Repository Structure

* `pkg` is where all the extra libp2p repositories are. For example things like `go-libp2p-loggables`, `go-libp2p-buffer-pool`, and all transports are here.
* `p2p` is equivalent
