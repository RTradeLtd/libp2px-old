<h1 align="center">LibP2PX ☄️</h1>

<p align="center">
  <a href="#about"><strong>About</strong></a> ·
  <a href="#goals"><strong>Goals</strong></a> · 
  <a href="#differences-from-libp2p"><strong>Differences From LibP2P</strong></a> ·
  <a href="#repository-structure"><strong>Repository Structure</strong></a> ·
  <a href="#license"><strong>License</strong></a> ·
</p>

[![GoDoc](https://godoc.org/github.com/RTradeLtd/libp2px?status.svg)](https://godoc.org/github.com/RTradeLtd/libp2px) [![Build Status](https://travis-ci.com/RTradeLtd/libp2px.svg?branch=master)](https://travis-ci.com/RTradeLtd/libp2px) [![codecov](https://codecov.io/gh/RTradeLtd/libp2px/branch/master/graph/badge.svg)](https://codecov.io/gh/RTradeLtd/libp2px) [![Maintainability](https://api.codeclimate.com/v1/badges/eb5732a9c3200416782f/maintainability)](https://codeclimate.com/github/RTradeLtd/libp2px/maintainability)

# About

> **status: work in progress, not recomended for use in production**

`libp2px` is a fork of `github.com/libp2p/go-libp2p`, intended for use with TemporalX's enterprise IPFS node, but suitable as a performance optimized replacement of `go-libp2p`. We will try to remain backwards compatable as much as possible with `go-libp2p`, and the rest of the network, but this is neither a design goal, nor an outright priority. This does not currently contain a version of `go-libp2p-kad-dht` so you'll need to BYOD, and use an existing implementation which should probably be `go-libp2p-kad-dht`.

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
## Compatability Issues

### Confirmed

None

### Suspected

#### secp256k1 issues

One possible compatability issue is with secp256k1 keys being incompatible between libp2px and go-libp2p. The reason being is that during the fork of `libp2p/go-libp2p-core` tests broke when using the pre-generated secp256k1 test data, and we needed to regenerate it to fix.

### Needs Investigation

#### TestStBackpressureStreamWrite TestProtoDowngrade,TestHostProtoPreference, TestDefaultListenAddrs, TestNewDialOld Failures

Notice that in [this commit](https://github.com/RTradeLtd/libp2px/commit/b45de2ae197cb95aacb150c8a53490d81cacfdf7) the TravisCI builds passed. The important thing to take note of is that this commit uses the [IPFS ci helper scripts](https://github.com/ipfs/ci-helpers/blob/master/travis-ci/run-standard-tests.sh). However if you notice in [this commit](https://github.com/RTradeLtd/libp2px/commit/1e9958227c15fbfc446f356b4660a317b9e6efc9) when we switched to a different method of executing golang test tooling, we encounter build failures. I'm not yet sure why but this is repatable behavior. This needs investigation.

The names of all tests that fail when not using the ipfs ci helper script are listed in this markdown header. All but 

# Support

In terms of support for using this library from RTrade, we will be more than happy to address github issues for deficiencies in functionality that impact performance, but that is where the level of support will end. If you have issues with integrating this code, want explanations about the code, etc... that isn't publicly available please contact us privately.

# Repository Structure

* `pkg` is where all the extra libp2p repositories are. For example things like `go-libp2p-loggables`, `go-libp2p-buffer-pool`, and all transports are here.
* `p2p` is equivalent


# License

All original code is licensed under MIT+Apache, and we've included all the previous licenses. New code and modifications are licensed under3 AGPLv3