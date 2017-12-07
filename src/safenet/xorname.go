package safenet

import (
	"fmt"
)

type XorName []byte

const xornameBits = 256

func NewXorName() XorName {
	// create a name from prng
	name := make([]byte, xornameBits/8)
	prng.Read(name)
	return name
}

func (x XorName) BinaryString() string {
	s := ""
	for _, b := range x {
		s = s + fmt.Sprintf("%08b", b)
	}
	return s
}

// XorNames are compared by their reverse-byte-order
// so the order is independent of the prefix.
func (x XorName) IsBefore(y XorName) bool {
	for i := len(x) - 1; i >= 0; i-- {
		if x[i] == y[i] {
			continue
		}
		return x[i] < y[i]
	}
	return false
}
