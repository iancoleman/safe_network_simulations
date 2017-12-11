package safenet

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

type ByAge []*Vault

func (a ByAge) Len() int      { return len(a) }
func (a ByAge) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByAge) Less(i, j int) bool {
	// ties in age are resolved by comparing the name itself
	if a[i].Age == a[j].Age {
		return a[i].Name.IsLessThan(a[j].Name)
	}
	return a[i].Age < a[j].Age
}
