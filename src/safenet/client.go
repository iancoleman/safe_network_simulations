package safenet

// clients do the following activities:
// upload data
// fetch data
// exchange coins for cash and vice versa
// manage vaults

type Client interface {
	// uploading methods
	MbPutPerDay() float64
	// fetching methods
	MbGetPerDay() float64
	// vault methods
	NewVaultsToStart() []*Vault
	ExistingVaultsToStop() []*Vault
}

type ConsistentClient struct {
	ConsistentUploader
	ConsistentDownloader
	ConsistentOperator
}

func NewConsistentClient() *ConsistentClient {
	c := ConsistentClient{}
	c.ConsistentOperator.Vaults = []*Vault{}
	return &c
}

type ConsistentUploader struct{}

func (c ConsistentUploader) MbPutPerDay() float64 {
	return 10
}

type ConsistentDownloader struct{}

func (c *ConsistentDownloader) MbGetPerDay() float64 {
	return 1000
}

type ConsistentOperator struct {
	Vaults []*Vault
}

func (c *ConsistentOperator) NewVaultsToStart() []*Vault {
	newVaults := []*Vault{}
	totalNewVaults := 2
	for i := 0; i < totalNewVaults; i++ {
		v := NewVault()
		newVaults = append(newVaults, v)
		c.Vaults = append(c.Vaults, v)
	}
	return newVaults
}

func (c *ConsistentOperator) ExistingVaultsToStop() []*Vault {
	if len(c.Vaults) == 0 {
		return []*Vault{}
	}
	toStop := c.Vaults[0:1]
	c.Vaults = c.Vaults[1:len(c.Vaults)]
	return toStop
}
