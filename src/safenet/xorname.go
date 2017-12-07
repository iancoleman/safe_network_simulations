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
