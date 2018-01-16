package main

import (
	"fmt"
	"safenet"
	"sort"
)

func main() {
	// get user variables
	seed := safenet.LoadConfigInt("config_section_size_distribution.json", "seed", 0)
	netsize := safenet.LoadConfigInt("config_section_size_distribution.json", "netsize", 100000)
	// create network
	network := safenet.NewNetworkFromSeed(int64(seed))
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
	sizes := map[int]int{}
	sizeKeys := []int{}
	for _, s := range network.Sections {
		size := len(s.Vaults)
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
	fmt.Println(network.TotalVaults(), "total vaults")
	fmt.Println(network.TotalJoins, "total joins")
	fmt.Println(network.TotalDepartures, "total departures")
	fmt.Println(network.TotalSections(), "total sections")
	fmt.Println(network.TotalSplits, "total splits")
	fmt.Println(network.TotalMerges, "total merges")
}
