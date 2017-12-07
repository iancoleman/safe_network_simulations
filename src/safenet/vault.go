package safenet

type Vault struct {
	Name       XorName
	Prefix     Prefix
	Age        uint
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

func (v *Vault) IncrementAge() {
	v.Age = v.Age + 1
}

func (v *Vault) IsAdult() bool {
	return v.Age > 4
}
