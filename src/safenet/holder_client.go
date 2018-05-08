package safenet

type HolderClient struct {
	HolderUploader
	HolderDownloader
	HolderOperator
}

func NewHolderClient() *HolderClient {
	c := HolderClient{}
	idBytes := make([]byte, 8)
	prng.Read(idBytes)
	c.HolderUploader.IdStr = string(idBytes)
	c.HolderOperator.Vaults = []*Vault{}
	return &c
}

type HolderUploader struct {
	IdStr string
}

func (h HolderUploader) MbPutPerDay() float64 {
	return 0
}

func (i HolderUploader) Id() string {
	return i.IdStr
}

type HolderDownloader struct{}

func (i *HolderDownloader) MbGetPerDay() float64 {
	return 0
}

type HolderOperator struct {
	Vaults     []*Vault
	Safecoins  int32
	PutBalance float64
}

func (o *HolderOperator) NewVaultsToStart() []*Vault {
	newVaults := []*Vault{}
	return newVaults
}

func (o *HolderOperator) ExistingVaultsToStop() []*Vault {
	return []*Vault{}
}

func (o *HolderOperator) AllocateSafecoins(safecoins int32) {
	o.Safecoins = o.Safecoins + safecoins
}

func (o *HolderOperator) TotalSafecoins() int32 {
	return o.Safecoins
}

func (o *HolderOperator) DeductPutBalance(amount float64, n *Network) {
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
	o.PutBalance = o.PutBalance - amount
}
