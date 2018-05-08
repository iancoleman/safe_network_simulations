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
	Vaults     []*Vault
	Safecoins  int32
	PutBalance float64
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

func (o *ConsistentOperator) DeductPutBalance(amount float64, n *Network) {
	// if not enough put balance, sell some coins to buy some puts
	// TODO allow this strategy to be varied rather than strictly on demand
	for o.PutBalance < amount && o.Safecoins > 0 {
		// sell a coin to the network
		puts := n.BuyPuts(1)
		// TODO network should manage operator balances
		o.Safecoins = o.Safecoins - 1
		o.PutBalance = o.PutBalance + puts
	}
	// deduct the amount
	// TODO validate that put balance does not go below zero
	o.PutBalance = o.PutBalance - amount
}
