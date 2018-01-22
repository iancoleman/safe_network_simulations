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

// Client definitions

type ConsistentClient struct {
	ConsistentUploader
	ConsistentDownloader
	ConsistentOperator
}

type InconsistentClient struct {
	InconsistentUploader
	InconsistentDownloader
	InconsistentOperator
}


// Client constructors

func NewConsistentClient() *ConsistentClient {
	c := ConsistentClient{}
	c.ConsistentOperator.Vaults = []*Vault{}
	return &c
}

func NewInconsistentClient() *InconsistentClient {
	c := InconsistentClient{}
	c.InconsistentOperator.Vaults = []*Vault{}
	return &c
}

// Client methods

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

type InconsistentUploader struct{}

func (i InconsistentUploader) MbPutPerDay() float64 {
	return float64(prng.Intn(20))
}

type InconsistentDownloader struct{}

func (i *InconsistentDownloader) MbGetPerDay() float64 {
	return float64(prng.Intn(2000))
}

type InconsistentOperator struct {
	Vaults []*Vault
}

func (o *InconsistentOperator) NewVaultsToStart() []*Vault {
	newVaults := []*Vault{}
	totalNewVaults := prng.Intn(4) + 1
	for i := 0; i < totalNewVaults; i++ {
		v := NewVault()
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

func NewRandomClient() Client {
	if prng.Float64() < 0.5 {
		return NewConsistentClient()
	} else {
		return NewInconsistentClient()
	}
}
