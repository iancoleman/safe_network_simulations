package safenet

import (
	"math/big"
)

type Vault struct {
	Name       XorName
	Prefix     Prefix
	Age        int
	IsAttacker bool
}

func NewVault() *Vault {
	return &Vault{
		Name: NewXorName(),
		Age:  1,
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

type forEldership []*Vault

func (v forEldership) Len() int      { return len(v) }
func (v forEldership) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v forEldership) Less(i, j int) bool {
	// ties in age are resolved by XOR their public keys together and find the
	// one XOR closest to it
	// see https://forum.safedev.org/t/data-chains-deeper-dive/1209
	// in this case the vault xorname is used as the public key
	if v[i].Age == v[j].Age {
		x := big.NewInt(0)
		x.Xor(v[i].Name.bigint, v[j].Name.bigint)
		xi := big.NewInt(0)
		xi.Xor(v[i].Name.bigint, x)
		xj := big.NewInt(0)
		xj.Xor(v[j].Name.bigint, x)
		// if xi is larger than xj then i should be lower in the sort order
		// than j since i is further away.
		return xi.Cmp(xj) == 1
	}
	return v[i].Age < v[j].Age
}
