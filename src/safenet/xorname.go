package safenet

import (
	"math/big"
)

type XorName struct {
	bigint *big.Int
	bytes  []byte
}

const xornameBits = 256
const xornameBytes = xornameBits / 8

func NewXorName() XorName {
	// create a name from prng
	name := make([]byte, xornameBytes)
	prng.Read(name)
	x := XorName{
		bigint: big.NewInt(0).SetBytes(name),
		bytes:  name,
	}
	return x
}

func (x XorName) BinaryString() string {
	s := x.bigint.Text(2)
	for len(s) < xornameBits {
		s = "0" + s
	}
	return s
}

func (x XorName) IsLessThan(y XorName) bool {
	return x.bigint.Cmp(y.bigint) == -1
}

func (x XorName) ByteAtIndex(i int) byte {
	return x.bytes[i]
}
