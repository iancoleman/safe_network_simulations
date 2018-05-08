package safenet

type ConsistentClient struct {
	ConsistentUploader
	ConsistentDownloader
	ConsistentOperator
}

func NewConsistentClient() *ConsistentClient {
	c := ConsistentClient{}
	idBytes := make([]byte, 8)
	prng.Read(idBytes)
	c.ConsistentUploader.IdStr = string(idBytes)
	c.ConsistentOperator.Vaults = []*Vault{}
	return &c
}

type ConsistentUploader struct {
	UniversalUploader
}

func (c ConsistentUploader) MbPutPerDay() float64 {
	return 10
}

type ConsistentDownloader struct{}

func (c *ConsistentDownloader) MbGetPerDay() float64 {
	return 1000
}

type ConsistentOperator struct {
	UniversalOperator
}

func (o *ConsistentOperator) NewVaultsToStart() []*Vault {
	newVaults := []*Vault{}
	totalNewVaults := 2
	for i := 0; i < totalNewVaults; i++ {
		v := NewVaultForOperator(o)
		newVaults = append(newVaults, v)
		o.Vaults = append(o.Vaults, v)
	}
	return newVaults
}

func (o *ConsistentOperator) ExistingVaultsToStop() []*Vault {
	if len(o.Vaults) == 0 {
		return []*Vault{}
	}
	toStop := o.Vaults[0:1]
	o.Vaults = o.Vaults[1:len(o.Vaults)]
	return toStop
}
