package safenet

import (
	"fmt"
	"math/rand"
)

const seed = 0

var prng = rand.New(rand.NewSource(seed))

type XorName struct {
	binary string
	bytes  []byte
}

func NewXorName() XorName {
	// get an id from prng
	id := make([]byte, 32)
	prng.Read(id)
	// convert to binary
	binaryStr := ""
	for _, n := range id {
		s := fmt.Sprintf("%08b", n)
		binaryStr = binaryStr + s
	}
	// convert to XorName type
	return XorName{
		binary: binaryStr,
		bytes:  id,
	}
}

func (x XorName) StartsWith(p Prefix) bool {
	return x.binary[0:len(p)] == string(p)
}

// XorNames are compared by their reverse-byte-order
// so the order is independent of the prefix.
func (x XorName) IsBefore(y XorName) bool {
	for i := len(x.bytes) - 1; i >= 0; i-- {
		if x.bytes[i] == y.bytes[i] {
			continue
		}
		return x.bytes[i] < y.bytes[i]
	}
	return false
}
