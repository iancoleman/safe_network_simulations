package safenet

import (
	"math/big"
)

type NetworkEvent struct {
	hash            *big.Int
	NewSections     []*Section
	VaultToRelocate *Vault
}

const networkeventHashBits = 256
const networkeventHashBytes = networkeventHashBits / 8

var largestHashValue = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

func NewNetworkEvent() *NetworkEvent {
	ne := NetworkEvent{}
	// create a hash from prng
	b := make([]byte, networkeventHashBytes)
	prng.Read(b)
	ne.hash = big.NewInt(0)
	ne.hash.SetBytes(b)
	return &ne
}

// calculates x = b.hash % divisor and returns x == 0
func (ne NetworkEvent) HashModIsZero(divisor *big.Int) bool {
	x := big.NewInt(0)
	x.Mod(ne.hash, divisor)
	return x.Cmp(big.NewInt(0)) == 0
}
