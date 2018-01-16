package main

import (
	"fmt"
	"safenet"
)

func main() {
	// get user variables
	seed := safenet.LoadConfigInt("config_google_attack_targeted.json", "seed", 0)
	netsize := safenet.LoadConfigInt("config_google_attack_targeted.json", "netsize", 100000)
	// create network
	network := safenet.NewNetworkFromSeed(int64(seed))
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
		disallowed := network.AddVault(v)
		for disallowed {
			v = safenet.NewVault()
			disallowed = network.AddVault(v)
		}
		// remove existing vaults until network is back to capacity
		for network.TotalVaults() > netsize {
			e := network.GetRandomVault()
			network.RemoveVault(e)
		}
	}
	fmt.Println("   100%\n")
	fmt.Println(network.TotalVaults(), "vaults before attack")
	// atack the network until the attacker owns a section
	// by adding vaults to a specific prefix
	// which will be relocated only to the neighbours which is
	// a fairly small subsection of the network
	attackPrefix := safenet.NewXorName()
	// TODO should set prefixBitCount to the current length of the
	// section prefix length. 64 is a sort-of suitable compromise
	// which is valid for networks with up to about 2^64 sections
	prefixBitCount := 64
	attackVaultCount := 0
	for true {
		// logging
		if attackVaultCount%1000 == 0 {
			fmt.Print(attackVaultCount, " attacking vaults added\r")
		}
		// add an attacking vault
		disallowed := true
		var a *safenet.Vault
		for disallowed {
			a = safenet.NewVault()
			// set vault to use the attack prefix
			for i := 0; i < prefixBitCount; i++ {
				// TODO vault / prefix abstraction seems wrong here, too messy
				a.Name.SetBit(i, attackPrefix.GetBit(i))
			}
			a.IsAttacker = true
			disallowed = network.AddVault(a)
		}
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
			disallowed := true
			for disallowed {
				v := safenet.NewVault()
				disallowed = network.AddVault(v)
			}
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
