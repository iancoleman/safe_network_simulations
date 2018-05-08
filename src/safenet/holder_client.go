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
	Vaults    []*Vault
	Safecoins int32
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
