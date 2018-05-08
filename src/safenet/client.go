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
