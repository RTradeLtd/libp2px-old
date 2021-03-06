<h1 align="center">LibP2P<u>X</u> 🌌</h1>

<p align="center">
  <a href="#about"><strong>About</strong></a> ·
  <a href="#goals"><strong>Goals</strong></a> · 
  <a href="#differences-from-libp2p"><strong>Differences From LibP2P</strong></a> ·
  <a href="#repository-structure"><strong>Repository Structure</strong></a> ·
  <a href="#license"><strong>License</strong></a> ·
</p>

<p align="center">
  <a href="https://godoc.org/github.com/RTradeLtd/libp2px">
    <img src="https://godoc.org/github.com/RTradeLtd/libp2px?status.svg"
       alt="GoDocs available" />
  </a>
  <a href="https://travis-ci.com/RTradeLtd/libp2px">
    <img src="https://travis-ci.com/RTradeLtd/libp2px.svg?branch=master"
      alt="Travis Build Status" />
  </a>
  <a href="https://github.com/RTradeLtd/libp2px/releases">
    <img src="https://img.shields.io/github/release-pre/RTradeLtd/libp2px.svg"
      alt="Release" />
  </a>
  </br>
  <a href="https://codecov.io/gh/RTradeLtd/libp2px">
    <img src="https://codecov.io/gh/RTradeLtd/libp2px/branch/master/graph/badge.svg" 
        alt="Code Coverage"/>
  </a>
  <a href="https://codeclimate.com/github/RTradeLtd/libp2px/maintainability">
    <img src="https://api.codeclimate.com/v1/badges/eb5732a9c3200416782f/maintainability" 
        alt="Maintanability"/>
  </a>
  <a href="https://goreportcard.com/report/github.com/RTradeLtd/libp2px">
    <img src="https://goreportcard.com/badge/github.com/RTradeLtd/libp2px"
      alt="Clean code" />
  </a>
</p>
</p>

# About

> **status: work in progress, not recomended for use in production**

`libp2px` is a complete fork of the libp2p stack, intended for use with TemporalX's enterprise IPFS node, buit suitable for people who need a [performance focused alternative to existing libp2p implementations. We will try to remain backwards compatable as much as possible with `go-libp2p`, and the rest of the network, but this is neither a design goal, nor an outright priority. This does not currently contain a version of `go-libp2p-kad-dht` so you'll need to BYOD, and use an existing implementation. For now we recommend using `go-libp2p-kad-dht`.

We have a fully fork version of `libp2p/go-libp2p-core` at [RTradeLtd/libp2px-core](https://github.com/RTradeLtd/libp2px-core), and a fully forked version of `libp2p/go-openssl` at [RTradeLtd/libp2px-openssl](https://github.com/RTradeLtd/libp2px-openssl).

# Goals

* A more maintainable and approachable codebase 
* Thoroughly tested code base
* Performance and efficiency
* Privacy as long as it doesn't compromise performance
  * To this end we have disabled the built-in identify, and ping service. Eventually these will be available as modules

# Differences From LibP2P

* No default ping and identify service
* Complete removal of `goprocess` which at scale becomes a significant resource hog.
  * We replace this with idomatic, and stdlib friendly context usage
* Removal of `go-log` replaced with pure zap logging
* Transports have no logging as they were relying on gobally initialized loggers
  * With the current method of using transports, it's impossible to use logging there with non-global loggers, at some point in time this may be refactored and changed.
* All libp2p dependencies from the libp2p organization have been forked, and stored in this repository as a "mono repo"
  * The exceptions to this are the core, and openssl repos.

## Compatability Issues

### Confirmed

None

### Suspected

#### secp256k1 issues

One possible compatability issue is with secp256k1 keys being incompatible between libp2px and go-libp2p. The reason being is that during the fork of `libp2p/go-libp2p-core` tests broke when using the pre-generated secp256k1 test data, and we needed to regenerate it to fix.

### Needs Investigation

#### TestStBackpressureStreamWrite TestProtoDowngrade,TestHostProtoPreference, TestDefaultListenAddrs, TestNewDialOld, TestValidateOverload, TestPeerTopicReporting Failures

Notice that in [this commit](https://github.com/RTradeLtd/libp2px/commit/b45de2ae197cb95aacb150c8a53490d81cacfdf7) the TravisCI builds passed. The important thing to take note of is that this commit uses the [IPFS ci helper scripts](https://github.com/ipfs/ci-helpers/blob/master/travis-ci/run-standard-tests.sh). However if you notice in [this commit](https://github.com/RTradeLtd/libp2px/commit/1e9958227c15fbfc446f356b4660a317b9e6efc9) when we switched to a different method of executing golang test tooling, we encounter build failures. I'm not yet sure why but this is repatable behavior. This needs investigation.

The names of all tests that fail when not using the ipfs ci helper script are listed in this markdown header. All but 

# Support

In terms of support for using this library from RTrade, we will be more than happy to address github issues for deficiencies in functionality that impact performance, but that is where the level of support will end. If you have issues with integrating this code, want explanations about the code, etc... that isn't publicly available please contact us privately.

# Repository Structure

* `pkg` is where all the extra libp2p repositories are. For example things like `go-libp2p-loggables`, `go-libp2p-buffer-pool`, and all transports are here.
* `p2p` is equivalent

## pkg

`pkg` contains various packages that may be useful to other users of `libp2px` including:

| path | description |
|------|-------------|
| `pkg/autonat` | an autonat service implementation |
| `pkg/blankhost` | a bare libp2px host implementation | 
| `pkg/buffer-pool` | a memory buffer pool |
| `pkg/discovery` | a service to discovert things |
| `pkg/kbucket` | TODO | 
| `pkg/mdns` | TODO |
| `pkg/metrics` | TODO |
| `pkg/msgio` | TODO |
| `pkg/nat` | TODO | 
| `pkg/peerstore` | a storage system for libp2px peers |
| `pkg/pnet` | TODO |
| `pkg/pubsub` | a libp2px pubsub implementation supporting gossipsub, floodsub, and randomsub |
| `pkg/reuseport` | TODO |
| `pkg/swarm` | a libp2px swarm manager | 
| `pkg/transports` | contains a variety of libp2px transports, responsible for defining methods of connecting two peers |
| `pkg/muxers` | contains a variety of connection multiplexers |
  
# License

All original code is licensed under MIT+Apache, and we've included all the previous licenses. New code (aka, newly added transports, etc...) will be added with AGPLv3 licenses.