package main

import (
	"fmt"
	"safenet"
)

const numIcoCoins = 452552412
const growthRate = 1.003 // 0.3% growth every step

func main() {
	// create network
	n := safenet.NewNetwork()
	// initialize ICO coins
	fmt.Println("Initializing ICO coins")
	for n.TotalSafecoins < numIcoCoins {
		n.ForceCreateSafecoin()
		if n.TotalSafecoins%100000 == 0 {
			pct := int64(float64(n.TotalSafecoins) / float64(numIcoCoins) * 100)
			fmt.Print(pct, "% - ", n.TotalSafecoins, " of ", numIcoCoins, "\n")
		}
	}
	fmt.Println()
	// initialize report
	report := "step,totalSafecoin,mbPerSafecoin,totalSections,totalVaults\n"
	fmt.Print(report)
	// simulate the network activity
	steps := 100000
	totalVaultJoinsPerStep := 110.0
	totalVaultDepartsPerStep := 10.0
	totalPutsPerStep := 1000.0
	totalGetsPerStep := 2000.0
	for step := 0; step < steps; step++ {
		// update joins based on growth rate
		// TODO base this on the economics of joining
		totalVaultJoinsPerStep = totalVaultJoinsPerStep * growthRate
		// update removals based on growth rate
		// TODO base this on the economics of remaining
		totalVaultDepartsPerStep = totalVaultDepartsPerStep * growthRate
		// update puts based on growth rate
		totalPutsPerStep = totalPutsPerStep * growthRate
		// update gets based on growth rate
		totalGetsPerStep = totalGetsPerStep * growthRate
		// track some variables for interleaving the events as much as possible
		sumJoins := 0.0
		sumDeparts := 0.0
		sumPuts := 0.0
		sumGets := 0.0
		mostIters := 0.0
		if totalVaultJoinsPerStep > mostIters {
			mostIters = totalVaultJoinsPerStep
		}
		if totalVaultDepartsPerStep > mostIters {
			mostIters = totalVaultDepartsPerStep
		}
		if totalPutsPerStep > mostIters {
			mostIters = totalPutsPerStep
		}
		if totalGetsPerStep > mostIters {
			mostIters = totalGetsPerStep
		}
		// interleave the events
		for i := 0.0; i < mostIters; i++ {
			pct := (i + 1) / mostIters
			// add new vaults
			expectedSumJoins := pct * totalVaultJoinsPerStep
			for sumJoins < expectedSumJoins {
				v := safenet.NewVault()
				n.AddVault(v)
				sumJoins = sumJoins + 1
			}
			// remove some vaults
			expectedSumDeparts := pct * totalVaultDepartsPerStep
			for sumDeparts < expectedSumDeparts {
				v := n.GetRandomVault()
				n.RemoveVault(v)
				sumDeparts = sumDeparts + 1
			}
			// do some puts
			expectedSumPuts := pct * totalPutsPerStep
			for sumPuts < expectedSumPuts {
				n.DoRandomPut()
				sumPuts = sumPuts + 1
			}
			// do gets
			expectedSumGets := pct * totalGetsPerStep
			for sumGets < expectedSumGets {
				n.DoRandomGet()
				sumGets = sumGets + 1
			}
		}
		// calculate average mb per safecoin
		mbPerSafecoin := 1.0 / n.AvgSafecoinPerMb()
		// add step to report
		line := fmt.Sprintf("%d,%d,%f,%d,%d\n", step, n.TotalSafecoins, mbPerSafecoin, n.TotalSections(), n.TotalVaults())
		fmt.Print(line)
		report = report + line
	}
	//fmt.Println(report)
}
