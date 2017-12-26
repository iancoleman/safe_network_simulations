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
		// remove existing vault
		if i >= netsize {
			v := network.GetRandomVault()
			network.RemoveVault(v)
		}
	}
	fmt.Println("   100%\n")
	// report
	// age distribution for all vaults
	ageCount, ageKeys := network.ReportAges()
	fmt.Println("age", "vaults")
	for _, age := range ageKeys {
		fmt.Println(age, ageCount[age])
	}
	fmt.Println()
	// section distribution of adults
	adultsCount := map[int]int{}
	adultsKeys := []int{}
	for _, s := range network.Sections {
		adults := s.TotalAdults()
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
	fmt.Println(network.TotalVaults(), "total vaults")
	fmt.Println(network.TotalSections(), "total sections")
}
