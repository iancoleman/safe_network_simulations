package safenet

type InconsistentClient struct {
	InconsistentUploader
	InconsistentDownloader
	InconsistentOperator
}

func NewInconsistentClient() *InconsistentClient {
	c := InconsistentClient{}
	idBytes := make([]byte, 8)
	prng.Read(idBytes)
	c.InconsistentUploader.IdStr = string(idBytes)
	c.InconsistentOperator.Vaults = []*Vault{}
	return &c
}

type InconsistentUploader struct {
	IdStr string
}

func (i InconsistentUploader) MbPutPerDay() float64 {
	return float64(prng.Intn(20))
}

func (i InconsistentUploader) Id() string {
	return i.IdStr
}

type InconsistentDownloader struct{}

func (i *InconsistentDownloader) MbGetPerDay() float64 {
	return float64(prng.Intn(2000))
}

type InconsistentOperator struct {
	Vaults    []*Vault
	Safecoins int32
}

func (o *InconsistentOperator) NewVaultsToStart() []*Vault {
	newVaults := []*Vault{}
	totalNewVaults := prng.Intn(4) + 1
	for i := 0; i < totalNewVaults; i++ {
		v := NewVaultForOperator(o)
		newVaults = append(newVaults, v)
		o.Vaults = append(o.Vaults, v)
	}
	return newVaults
}

func (o *InconsistentOperator) ExistingVaultsToStop() []*Vault {
	if len(o.Vaults) == 0 {
		return []*Vault{}
	}
	i := prng.Intn(len(o.Vaults))
	if i == 0 {
		return []*Vault{}
	}
	toStop := o.Vaults[0:i]
	o.Vaults = o.Vaults[i:len(o.Vaults)]
	return toStop
}

func (o *InconsistentOperator) AllocateSafecoins(safecoins int32) {
	o.Safecoins = o.Safecoins + safecoins
}

func (o *InconsistentOperator) TotalSafecoins() int32 {
	return o.Safecoins
}
