package safenet

// Common methods for all clients

type UniversalUploader struct {
	IdStr string
}

func (c UniversalUploader) Id() string {
	return c.IdStr
}

type UniversalOperator struct {
	Vaults     []*Vault
	Safecoins  int32
	PutBalance float64
}

func (o *UniversalOperator) AllocateSafecoins(safecoins int32) {
	o.Safecoins = o.Safecoins + safecoins
}

func (o *UniversalOperator) TotalSafecoins() int32 {
	return o.Safecoins
}

func (o *UniversalOperator) DeductPutBalance(amount float64, n *Network) {
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
