package main

import (
	"fmt"
	"safenet"
)

const numIcoCoins = 452552412
const growthRate = 1.003 // 0.3% growth every day

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
	report := "day,totalSafecoin,mbPerSafecoin,totalSections,totalVaults,putsToday,getsToday\n"
	fmt.Print(report)
	// simulate the network activity
	days := 100000
	totalVaultJoinsPerDay := 110.0
	totalVaultDepartsPerDay := 10.0
	totalPutsPerDay := 1000.0
	totalGetsPerDay := 2000.0
	for day := 0; day < days; day++ {
		// update joins based on growth rate
		// TODO base this on the economics of joining
		totalVaultJoinsPerDay = totalVaultJoinsPerDay * growthRate
		// update removals based on growth rate
		// TODO base this on the economics of remaining
		totalVaultDepartsPerDay = totalVaultDepartsPerDay * growthRate
		// update puts based on growth rate
		totalPutsPerDay = totalPutsPerDay * growthRate
		// update gets based on growth rate
		totalGetsPerDay = totalGetsPerDay * growthRate
		// track some variables for interleaving the events as much as possible
		sumJoins := 0.0
		sumDeparts := 0.0
		sumPuts := 0.0
		sumGets := 0.0
		mostIters := 0.0
		if totalVaultJoinsPerDay > mostIters {
			mostIters = totalVaultJoinsPerDay
		}
		if totalVaultDepartsPerDay > mostIters {
			mostIters = totalVaultDepartsPerDay
		}
		if totalPutsPerDay > mostIters {
			mostIters = totalPutsPerDay
		}
		if totalGetsPerDay > mostIters {
			mostIters = totalGetsPerDay
		}
		// interleave the events
		for i := 0.0; i < mostIters; i++ {
			pct := (i + 1) / mostIters
			// add new vaults
			expectedSumJoins := pct * totalVaultJoinsPerDay
			for sumJoins < expectedSumJoins {
				v := safenet.NewVault()
				n.AddVault(v)
				sumJoins = sumJoins + 1
			}
			// remove some vaults
			expectedSumDeparts := pct * totalVaultDepartsPerDay
			for sumDeparts < expectedSumDeparts {
				v := n.GetRandomVault()
				n.RemoveVault(v)
				sumDeparts = sumDeparts + 1
			}
			// do some puts
			expectedSumPuts := pct * totalPutsPerDay
			for sumPuts < expectedSumPuts {
				n.DoRandomPut()
				sumPuts = sumPuts + 1
			}
			// do gets
			expectedSumGets := pct * totalGetsPerDay
			for sumGets < expectedSumGets {
				n.DoRandomGet()
				sumGets = sumGets + 1
			}
		}
		// calculate average mb per safecoin
		mbPerSafecoin := 1.0 / n.AvgSafecoinPerMb()
		// add day to report
		line := fmt.Sprintf("%d,%d,%f,%d,%d,%d,%d\n", day, n.TotalSafecoins, mbPerSafecoin, n.TotalSections(), n.TotalVaults(), int64(sumPuts), int64(sumGets))
		fmt.Print(line)
		report = report + line
	}
	//fmt.Println(report)
}
