package main

import (
	"flag"
	"fmt"
	"safenet"
)

func main() {
	// get user variables
	var seedPtr *int64
	seedPtr = flag.Int64("seed", 0, "seed for the prng")
	var netsizePtr *int
	netsizePtr = flag.Int("netsize", 100000, "number of vaults in the final network")
	flag.Parse()
	seed := *seedPtr
	netsize := *netsizePtr
	// create network
	network := safenet.NewNetworkFromSeed(seed)
	totalEvents := netsize * 5
	pctStep := totalEvents / 100
	// Create initial network
	fmt.Println("Building initial network")
	for i := 0; i < totalEvents; i++ {
		// logging
		if i%pctStep == 0 {
			progress := int(float64(i) / float64(totalEvents) * 100.0)
			fmt.Print("   ", progress, "%\r")
		}
		// create a new vault
		v := safenet.NewVault()
		network.AddVault(v)
		// remove existing vaults until network is back to capacity
		for network.TotalVaults() > netsize {
			e := network.GetRandomVault()
			network.RemoveVault(e)
		}
	}
	fmt.Println("   100%\n")
	fmt.Println(network.TotalVaults(), "vaults before attack")
	// atack the network until the attacker owns a section
	attackVaultCount := 0
	for true {
		// logging
		if attackVaultCount%1000 == 0 {
			fmt.Print(attackVaultCount, " attacking vaults added\r")
		}
		// add an attacking vault
		a := safenet.NewVault()
		a.IsAttacker = true
		network.AddVault(a)
		attackVaultCount = attackVaultCount + 1
		// check if attack has worked
		s := network.Sections[a.Prefix.Key]
		if s.IsAttacked() {
			break
		}
		// TODO edge case: if section just split it may have
		// caused the sibling section to be attacked so
		// should check the sibling section
		// add one normal vault for every ten attacking
		if attackVaultCount%10 == 0 {
			v := safenet.NewVault()
			network.AddVault(v)
		}
		// remove a non-attacking vault for every ten attacking
		if attackVaultCount%10 == 0 {
			e := network.GetRandomVault()
			for e.IsAttacker {
				e = network.GetRandomVault()
			}
			network.RemoveVault(e)
		}
	}
	// report
	fmt.Println(attackVaultCount, "attacking vaults added to own a section")
	fmt.Println(network.TotalVaults(), "vaults after attack")
	fmt.Println(network.TotalSections(), "sections after attack")
	pctOwned := float64(attackVaultCount) / float64(network.TotalVaults()) * 100
	fmt.Println(pctOwned, "percent of total network owned by attacker")
}
