# libp2px

[![GoDoc](https://godoc.org/github.com/RTradeLtd/libp2px?status.svg)](https://godoc.org/github.com/RTradeLtd/libp2px) [![Build Status](https://travis-ci.com/RTradeLtd/libp2px.svg?branch=master)](https://travis-ci.com/RTradeLtd/libp2px) [![codecov](https://codecov.io/gh/RTradeLtd/libp2px/branch/master/graph/badge.svg)](https://codecov.io/gh/RTradeLtd/libp2px) [![Maintainability](https://api.codeclimate.com/v1/badges/eb5732a9c3200416782f/maintainability)](https://codeclimate.com/github/RTradeLtd/libp2px/maintainability)

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

### secp256k1 issues

One possible compatability issue is with secp256k1 keys being incompatible between libp2px and go-libp2p.

## Needs Investigation

### TestStBackpressureStreamWrite And TestProtoDowngrade Failures

Notice that in [this commit](https://github.com/RTradeLtd/libp2px/commit/b45de2ae197cb95aacb150c8a53490d81cacfdf7) the TravisCI builds passed. The important thing to take note of is that this commit uses the [IPFS ci helper scripts](https://github.com/ipfs/ci-helpers/blob/master/travis-ci/run-standard-tests.sh). However if you notice in [this commit](https://github.com/RTradeLtd/libp2px/commit/1e9958227c15fbfc446f356b4660a317b9e6efc9) when we switched to a different method of executing golang test tooling, we encounter build failures. I'm not yet sure why but this is repatable behavior. This needs investigation.

# Repository Structure

* `pkg` is where all the extra libp2p repositories are. For example things like `go-libp2p-loggables`, `go-libp2p-buffer-pool`, and all transports are here.
* `p2p` is equivalent
