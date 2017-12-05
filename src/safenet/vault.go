package safenet

type Vault struct {
	Name   XorName
	Prefix Prefix
}

func NewVault() *Vault {
	return &Vault{
		Name: NewXorName(),
	}
}

func (v *Vault) SetPrefix(p Prefix) {
	v.Prefix = p
}
