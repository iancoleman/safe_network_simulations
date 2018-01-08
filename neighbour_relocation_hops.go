package main

import (
	"flag"
	"fmt"
	"safenet"
	"sort"
)

// When a vault is relocated, it goes to the best neighbourhood.
// Finding the best neighbourhood is a recursive process, so the vault may end
// up several neighbourhoods away.
// This test looks at how far vaults are relocated, as in, 1 neighbourhood away
// or 2 neighbourhoods away etc.

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
	totalEvents := netsize * 12 / 10
	pctStep := totalEvents / 1000
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
	fmt.Println(network.TotalVaults(), "total vaults")
	// report
	occurences := map[int]int{}
	hopKeys := []int{}
	for _, hops := range network.NeighbourhoodHops {
		count, exists := occurences[hops]
		if !exists {
			count = 0
			occurences[hops] = count
			hopKeys = append(hopKeys, hops)
		}
		occurences[hops] = count + 1
	}
	sort.Sort(sort.IntSlice(hopKeys))
	fmt.Println("hops", "occurances")
	for _, hops := range hopKeys {
		fmt.Println(hops, occurences[hops])
	}
	fmt.Println(network.TotalRelocations, "total relocations")
}
