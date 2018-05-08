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
	MbPutForDay(int) float64
	Id() string
}

type Downloader interface {
	MbGetForDay(int) float64
}

type Operator interface {
	NewVaultsToStart() []*Vault
	ExistingVaultsToStop() []*Vault
	ConvertCoinsToPutBalance(int, Uploader, *Network)
	AllocateSafecoins(int32)
	TotalSafecoins() int32
	AllocatePuts(float64)
	TotalPutBalance() float64
}

func NewRandomClient() Client {
	var clientTypes float64 = 2 // update this if adding new client type
	classifier := prng.Float64()
	if classifier < 1/clientTypes {
		return NewConsistentClient()
		//} else if classifier < 2/clientTypes {
		//	return NewTemplateClient()
		//} else if classifier < 3/clientTypes {
		//	return NewYourInterestingClient()
	} else {
		return NewInconsistentClient()
	}
}
