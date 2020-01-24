package test

import (
	"math/rand"
	"sync/atomic"

	ci "github.com/RTradeLtd/libp2px-core/crypto"
)

var globalSeed int64

func RandTestKeyPair(typ, bits int) (ci.PrivKey, ci.PubKey, error) {
	// workaround for low time resolution
	seed := atomic.AddInt64(&globalSeed, 1)
	return SeededTestKeyPair(typ, bits, seed)
}

func SeededTestKeyPair(typ, bits int, seed int64) (ci.PrivKey, ci.PubKey, error) {
	r := rand.New(rand.NewSource(seed))
	return ci.GenerateKeyPairWithReader(typ, bits, r)
}
