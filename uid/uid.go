package uid

import (
	"crypto/rand"
	"math/big"
)

type Uid uint64

const Bits uint = 64

func New() Uid {
	var buf [8]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		panic("could not generate random ID")
	}
	var val big.Int
	val.SetBytes(buf[:])
	return Uid(val.Uint64())
}

func (self Uid) ToBig() *big.Int {
	return new(big.Int).SetUint64(uint64(self))
}
