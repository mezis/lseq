package uid

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

type Uid uint64

const Bits uint = 64

func Generate() Uid {
	var buf [8]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		panic("could not generate random ID")
	}
	var val big.Int
	val.SetBytes(buf[:])
	return Uid(val.Uint64())
}

func New(val *big.Int) Uid {
	return Uid(val.Uint64())
}

// func (uid Uid) ToBig() *big.Int {
// 	return new(big.Int).SetUint64(uint64(uid))
// }

func (uid Uid) ToBig(val *big.Int) *big.Int {
	return val.SetUint64(uint64(uid))
}

func (uid Uid) GoString() string {
	return fmt.Sprintf("@%X", uint64(uid))
}

func (uid Uid) String() string {
	return fmt.Sprintf("%X", uint64(uid))
}
