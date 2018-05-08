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
	c.InconsistentUploader.PutHistory = []float64{}
	c.InconsistentDownloader.GetHistory = []float64{}
	c.InconsistentOperator.Vaults = []*Vault{}
	return &c
}

type InconsistentUploader struct {
	UniversalUploader
	PutHistory []float64
}

func (i InconsistentUploader) MbPutForDay(day int) float64 {
	maxPutsPerDay := 20
	for len(i.PutHistory) <= day {
		i.PutHistory = append(i.PutHistory, float64(prng.Intn(maxPutsPerDay)))
	}
	return i.PutHistory[day]
}

type InconsistentDownloader struct {
	GetHistory []float64
}

func (i *InconsistentDownloader) MbGetForDay(day int) float64 {
	maxGetsPerDay := 2000
	for len(i.GetHistory) <= day {
		i.GetHistory = append(i.GetHistory, float64(prng.Intn(maxGetsPerDay)))
	}
	return i.GetHistory[day]
}

type InconsistentOperator struct {
	UniversalOperator
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
