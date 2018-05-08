package safenet

// Copy this template file to a new file for the new client type.
// Then modify the following functions to perform the behaviour for this
// specific client type.

func (c TemplateUploader) MbPutPerDay() float64 {
	// how many mb does this client PUT each day?
	return 10
}

func (c *TemplateDownloader) MbGetPerDay() float64 {
	// how many mb does this client GET each day?
	return 1000
}

func (o *TemplateOperator) NewVaultsToStart() []*Vault {
	// how many new vaults should this client start each day?
	totalNewVaults := 2
	// now create the vaults which will be added to the network
	// by whatever script is creating this client.
	newVaults := []*Vault{}
	for i := 0; i < totalNewVaults; i++ {
		v := NewVaultForOperator(o)
		newVaults = append(newVaults, v)
		o.Vaults = append(o.Vaults, v)
	}
	return newVaults
}

func (o *TemplateOperator) ExistingVaultsToStop() []*Vault {
	// Which vaults should this client stop and remove from the network?
	toStop := []*Vault{}
	toKeep := []*Vault{}
	for i, vault := range o.Vaults {
		// remove the first vault, keep the rest
		if i == 0 {
			toStop = append(toStop, vault)
		} else {
			toKeep = append(toKeep, vault)
		}
	}
	// update the vaults for the operator
	o.Vaults = toKeep
	// respond with the vaults to be stopped, which will be done by
	// whatever script is managing this client.
	return toStop
}

// Don't forget to update client.go:NewRandomClient() with this new client type!

// The rest is for setting up. Some parts may want to be changed but usually it
// can stay how it is.

type TemplateClient struct {
	TemplateUploader
	TemplateDownloader
	TemplateOperator
}

func NewTemplateClient() *TemplateClient {
	c := TemplateClient{}
	idBytes := make([]byte, 8)
	prng.Read(idBytes)
	c.TemplateUploader.IdStr = string(idBytes)
	c.TemplateOperator.Vaults = []*Vault{}
	return &c
}

type TemplateUploader struct {
	UniversalUploader
}

type TemplateDownloader struct{}

type TemplateOperator struct {
	UniversalOperator
}
