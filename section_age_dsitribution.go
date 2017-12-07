package main

import (
	"flag"
	"fmt"
	"safenet"
	"sort"
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
	for i := 0; i < totalEvents; i++ {
		// logging
		if i%pctStep == 0 {
			progress := int(float64(i) / float64(totalEvents) * 100.0)
			fmt.Print("   ", progress, "%\r")
		}
		// create new vault
		v := safenet.NewVault()
		network.AddVault(v)
		// relocate existing vault
		v = network.GetRandomVault()
		network.RelocateVault(v)
		// remove existing vault
		if i >= netsize {
			v := network.GetRandomVault()
			if v == nil {
				fmt.Println("Warning: No vault for GetRandomVault")
				continue
			}
			network.RemoveVault(v)
		}
	}
	fmt.Println("   100%\n")
	// report
	// age distribution for all vaults
	ageCount := map[int]int{}
	ageKeys := []int{}
	children := 0
	adults := 0
	for _, s := range network.Sections {
		for v := range s.Vaults {
			age := int(v.Age)
			// track distribution
			_, exists := ageCount[age]
			if !exists {
				ageCount[age] = 0
				ageKeys = append(ageKeys, age)
			}
			ageCount[age] = ageCount[age] + 1
			// track category
			if v.IsAdult {
				adults = adults + 1
			} else {
				children = children + 1
			}
		}
	}
	sort.Sort(sort.IntSlice(ageKeys))
	fmt.Println("age", "vaults")
	for _, age := range ageKeys {
		fmt.Println(age, ageCount[age])
	}
	fmt.Println()
	// section distribution of adults
	adultsCount := map[int]int{}
	adultsKeys := []int{}
	for _, s := range network.Sections {
		adults := int(s.TotalAdults)
		// track distribution
		_, exists := adultsCount[adults]
		if !exists {
			adultsCount[adults] = 0
			adultsKeys = append(adultsKeys, adults)
		}
		adultsCount[adults] = adultsCount[adults] + 1
	}
	sort.Sort(sort.IntSlice(adultsKeys))
	fmt.Println("adults", "sections")
	for _, adults := range adultsKeys {
		fmt.Println(adults, adultsCount[adults])
	}
	fmt.Println()
	// network stats
	fmt.Println(children, "children")
	fmt.Println(adults, "adults")
	fmt.Println(network.TotalSections, "total sections")
}
