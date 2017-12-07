package safenet

type Vault struct {
	Name       XorName
	Prefix     Prefix
	Age        uint
	IsAttacker bool
	IsAdult    bool
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
	if v.Age > 4 {
		v.IsAdult = true
	} else {
		v.IsAdult = false
	}
}

func (v *Vault) Rename() {
	v.Name = NewXorName()
}
