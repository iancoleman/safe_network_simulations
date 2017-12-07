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
			if v == nil {
				fmt.Println("Warning: No vault for GetRandomVault")
				continue
			}
			network.RemoveVault(v)
		}
	}
	fmt.Println("   100%\n")
	// report
	sizes := map[int]uint{}
	sizeKeys := []int{}
	for _, s := range network.Sections {
		size := int(s.TotalVaults)
		_, exists := sizes[size]
		if !exists {
			sizes[size] = 0
			sizeKeys = append(sizeKeys, size)
		}
		sizes[size] = sizes[size] + 1
	}
	fmt.Println("size", "count")
	sort.Sort(sort.IntSlice(sizeKeys))
	for _, size := range sizeKeys {
		fmt.Println(size, sizes[size])
	}
	fmt.Println()
	fmt.Println(network.TotalVaults, "total vaults")
	fmt.Println(network.TotalJoins, "total joins")
	fmt.Println(network.TotalDepartures, "total departures")
	fmt.Println(network.TotalVaultEvents, "total vault events")
	fmt.Println(network.TotalSections, "total sections")
	fmt.Println(network.TotalSectionEvents, "total section events")
	fmt.Println(network.TotalSplits, "total splits")
	fmt.Println(network.TotalMerges, "total merges")
}
