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

func (x XorName) XorDistanceTo(y XorName) XorDistance {
	d := XorDistance{}
	if len(x) != len(y) {
		fmt.Println("Warning: xordistance for mismatched lengths")
	}
	for i := 0; i < len(x); i++ {
		d = append(d, x[i]^y[i])
	}
	return d
}
