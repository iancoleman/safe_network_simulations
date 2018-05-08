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

func (o *UniversalOperator) ConvertCoinsToPutBalance(currentDay int, u Uploader, n *Network) {
	// TODO this should be overridden by specific client types
	// but for now it's done on demand by all client types
	mbToPutToday := u.MbPutForDay(currentDay)
	for mbToPutToday > o.PutBalance && o.Safecoins > 0 {
		n.BuyPuts(1, o)
	}
}

func (o *UniversalOperator) AllocateSafecoins(safecoins int32) {
	o.Safecoins = o.Safecoins + safecoins
}

func (o *UniversalOperator) TotalSafecoins() int32 {
	return o.Safecoins
}

func (o *UniversalOperator) AllocatePuts(puts float64) {
	o.PutBalance = o.PutBalance + puts
}

func (o *UniversalOperator) TotalPutBalance() float64 {
	return o.PutBalance
}

func (o *UniversalOperator) NewVaultsToStart() []*Vault {
	return []*Vault{}
}

func (o *UniversalOperator) ExistingVaultsToStop() []*Vault {
	return []*Vault{}
}
