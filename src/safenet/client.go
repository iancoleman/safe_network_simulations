package safenet

// clients do the following activities:
// upload data
// fetch data
// exchange coins for cash and vice versa
// manage vaults

type Client interface {
	Uploader
	Downloader
	Operator
}

type Uploader interface {
	MbPutPerDay() float64
	Id() string
}

type Downloader interface {
	MbGetPerDay() float64
}

type Operator interface {
	NewVaultsToStart() []*Vault
	ExistingVaultsToStop() []*Vault
	AllocateSafecoins(int32)
	TotalSafecoins() int32
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
	idBytes := make([]byte, 8)
	prng.Read(idBytes)
	c.ConsistentUploader.IdStr = string(idBytes)
	c.ConsistentOperator.Vaults = []*Vault{}
	return &c
}

func NewInconsistentClient() *InconsistentClient {
	c := InconsistentClient{}
	idBytes := make([]byte, 8)
	prng.Read(idBytes)
	c.InconsistentUploader.IdStr = string(idBytes)
	c.InconsistentOperator.Vaults = []*Vault{}
	return &c
}

func NewRandomClient() Client {
	if prng.Float64() < 0.5 {
		return NewConsistentClient()
	} else {
		return NewInconsistentClient()
	}
}

// Client methods

type ConsistentUploader struct {
	IdStr string
}

func (c ConsistentUploader) MbPutPerDay() float64 {
	return 10
}

func (c ConsistentUploader) Id() string {
	return c.IdStr
}

type ConsistentDownloader struct{}

func (c *ConsistentDownloader) MbGetPerDay() float64 {
	return 1000
}

type ConsistentOperator struct {
	Vaults    []*Vault
	Safecoins int32
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

func (o *ConsistentOperator) AllocateSafecoins(safecoins int32) {
	o.Safecoins = o.Safecoins + safecoins
}

func (o *ConsistentOperator) TotalSafecoins() int32 {
	return o.Safecoins
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
