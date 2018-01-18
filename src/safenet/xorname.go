package safenet

import (
	"fmt"
	"math/big"
)

type XorName struct {
	bigint *big.Int
	bits   []bool
}

const xornameBits = 256

func NewXorName() XorName {
	// create a name from prng
	nameBits := make([]bool, xornameBits)
	binaryStr := ""
	for i := 0; i < xornameBits; i++ {
		bit := prng.Intn(2)
		if bit == 0 {
			nameBits[i] = false
			binaryStr = binaryStr + "0"
		} else if bit == 1 {
			nameBits[i] = true
			binaryStr = binaryStr + "1"
		} else {
			fmt.Println("Warning: NewXorName generated a number not 0 or 1")
		}
	}
	nameBigint := big.NewInt(0)
	nameBigint.SetString(binaryStr, 2)
	x := XorName{
		bigint: nameBigint,
		bits:   nameBits,
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

func (x *XorName) SetBit(i int, b bool) {
	x.bits[i] = b
	v := uint(0)
	if b {
		v = 1
	}
	x.bigint.SetBit(x.bigint, i, v)
}
func (x *XorName) GetBit(i int) bool {
	return x.bits[i]
}

func bigIntModInt64IsZero(bi *big.Int, i int64) bool {
	b64 := big.NewInt(i)
	result := big.NewInt(1)
	result.Mod(bi, b64)
	return result.Cmp(big.NewInt(0)) == 0
}
