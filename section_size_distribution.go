package main

import (
	"fmt"
	"safenet"
	"sort"
)

func main() {
	network := safenet.NewNetwork()
	netsize := 100000
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
	sizes := map[int]int{}
	sizeKeys := []int{}
	for _, s := range network.Sections {
		i := s.TotalVaults
		_, exists := sizes[i]
		if !exists {
			sizes[i] = 0
			sizeKeys = append(sizeKeys, i)
		}
		sizes[i] = sizes[i] + 1
	}
	fmt.Println("size", "count")
	sort.Sort(sort.IntSlice(sizeKeys))
	for _, i := range sizeKeys {
		fmt.Println(i, sizes[i])
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
