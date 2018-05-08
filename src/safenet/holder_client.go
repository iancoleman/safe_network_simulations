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
	UniversalUploader
}

func (h HolderUploader) MbPutPerDay() float64 {
	return 0
}

type HolderDownloader struct{}

func (i *HolderDownloader) MbGetPerDay() float64 {
	return 0
}

type HolderOperator struct {
	UniversalOperator
}

func (o *HolderOperator) NewVaultsToStart() []*Vault {
	newVaults := []*Vault{}
	return newVaults
}

func (o *HolderOperator) ExistingVaultsToStop() []*Vault {
	return []*Vault{}
}
