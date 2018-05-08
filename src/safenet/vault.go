package safenet

import (
	"fmt"
	"math/big"
)

// the starting storage space for a vault is chosen randomly from this list
var startingStorageSizesMb = []int64{
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
	Chunks     []XorName
	TotalMb    int64
	Operator   Operator
}

func NewVault() *Vault {
	return &Vault{
		Name:    NewXorName(),
		Age:     1,
		TotalMb: 0,
		Chunks:  []XorName{},
	}
}

func NewVaultForOperator(o Operator) *Vault {
	return &Vault{
		Name:     NewXorName(),
		Age:      1,
		TotalMb:  randomStorageSize(),
		Chunks:   []XorName{},
		Operator: o,
	}
}

func (v *Vault) SetPrefix(p Prefix) {
	// check new prefix
	if !p.Matches(v.Name) {
		fmt.Println("Warning: Tried to set prefix that doesn't match vault name")
		fmt.Println("This vault will not have their prefix changed")
		fmt.Println("Consider if using vault.renameWithPrefix is suitable")
		return
	}
	// if the new prefix is longer, some chunks will be dead
	hasDeadChunks := len(p.bits) > len(v.Prefix.bits)
	// set new prefix
	v.Prefix = p
	// remove dead chunks
	if hasDeadChunks {
		v.removeDeadChunks()
	}
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
	v.Prefix = p
	v.removeDeadChunks()
}

func (v *Vault) removeDeadChunks() {
	// drop chunks that don't match the vault prefix.
	// TODO should drop chunks that aren't within GROUP_SIZE vaults close to
	// this vaults name.
	newChunks := []XorName{}
	for _, existingChunk := range v.Chunks {
		if v.Prefix.Matches(existingChunk) {
			newChunks = append(newChunks, existingChunk)
		}
	}
	v.Chunks = newChunks
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

func (v *Vault) StoreChunk(chunk XorName) bool {
	didStore := false
	// check if there's enough space to store the chunk
	spareSpace := v.SpareMb()
	if spareSpace <= 0 {
		return didStore
	}
	// store it
	v.Chunks = append(v.Chunks, chunk)
	// let the section know it was stored
	didStore = true
	return didStore
}

func (v *Vault) SpareMb() int64 {
	// Assumes all chunks are 1 MB in size.
	// TODO change this assumption since some chunks will be less.
	return v.TotalMb - int64(len(v.Chunks))
}

func randomStorageSize() int64 {
	// most vaults have smaller storage size
	n := len(startingStorageSizesMb)
	i := prng.Intn(n)
	return startingStorageSizesMb[i]
}
