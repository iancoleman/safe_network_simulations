package safenet

import (
	"fmt"
	"math/big"
)

// the starting storage space for a vault is chosen randomly from this list
var startingStorageSizesMb = []float64{
	100,
	200,
	300,
	400,
	500,
}

type Vault struct {
	Name       XorName
	Prefix     Prefix
	Age        int
	IsAttacker bool
	UsedMb     float64
	SpareMb    float64
	Operator   Operator
}

func NewVault() *Vault {
	return &Vault{
		Name:    NewXorName(),
		Age:     1,
		UsedMb:  0,
		SpareMb: randomStorageSize(),
	}
}

func NewVaultForOperator(o Operator) *Vault {
	return &Vault{
		Name:     NewXorName(),
		Age:      1,
		UsedMb:   0,
		SpareMb:  randomStorageSize(),
		Operator: o,
	}
}

func (v *Vault) SetPrefix(p Prefix) {
	v.Prefix = p
}

func (v *Vault) IncrementAge() {
	v.Age = v.Age + 1
}

func (v *Vault) IsAdult() bool {
	return v.Age > 4
}

func (v *Vault) renameWithPrefix(p Prefix) {
	v.Name = NewXorName()
	for i, prefixBit := range p.bits {
		v.Name.SetBit(i, prefixBit)
	}
}

type oldestFirst []*Vault

func (v oldestFirst) Len() int      { return len(v) }
func (v oldestFirst) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v oldestFirst) Less(i, j int) bool {
	if v[i].Age == v[j].Age {
		return resolveAgeTiebreaker(v[i], v[j])
	}
	return v[i].Age > v[j].Age
}

func resolveAgeTiebreaker(vi, vj *Vault) bool {
	// ties in age are resolved by XOR their public keys together and find the
	// one XOR closest to it
	// see https://forum.safedev.org/t/data-chains-deeper-dive/1209
	// in this case the vault xorname is used as the public key
	x := big.NewInt(0)
	x.Xor(vi.Name.bigint, vj.Name.bigint)
	xi := big.NewInt(0)
	xi.Xor(vi.Name.bigint, x)
	xj := big.NewInt(0)
	xj.Xor(vj.Name.bigint, x)
	// if xi is larger than xj then i should be lower in the sort order
	// than j since i is further away.
	return xi.Cmp(xj) == 1
}

func (v *Vault) StoreChunk() bool {
	// check if there's enough space to store the chunk
	didStore := false
	if v.SpareMb < 0 {
		fmt.Println("Warning: vault has", v.SpareMb, "spare MB")
		return didStore
	} else if v.SpareMb == 0 {
		return didStore
	}
	// store it
	v.UsedMb = v.UsedMb + 1
	v.SpareMb = v.SpareMb - 1
	// let the section know it was stored
	didStore = true
	return didStore
}

func randomStorageSize() float64 {
	// most vaults have smaller storage size
	n := len(startingStorageSizesMb)
	i := prng.Intn(n)
	return startingStorageSizesMb[i]
}
