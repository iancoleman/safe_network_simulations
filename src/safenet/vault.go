package safenet

type Vault struct {
	Name       XorName
	Prefix     Prefix
	IsAttacker bool
}

func NewVault() *Vault {
	return &Vault{
		Name: NewXorName(),
	}
}

func (v *Vault) SetPrefix(p Prefix) {
	v.Prefix = p
}
