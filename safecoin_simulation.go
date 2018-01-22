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
	}
	fmt.Println()
	// initialize report
	report := "day,totalSafecoin,mbPerSafecoin,totalSections,totalVaults\n"
	fmt.Print(report)
	// simulate the network activity by creating clients
	days := 100000
	newClientsPerDay := 10
	clients := []safenet.Client{}
	for day := 0; day < days; day++ {
		// create new clients
		for i := 0; i < newClientsPerDay; i++ {
			c := safenet.NewRandomClient()
			clients = append(clients, c)
		}
		// do each client activity
		for _, c := range clients {
			// make new vaults
			newVaults := c.NewVaultsToStart()
			for _, v := range newVaults {
				n.AddVault(v)
			}
			// stop existing vaults
			stopVaults := c.ExistingVaultsToStop()
			for _, v := range stopVaults {
				n.RemoveVault(v)
			}
			// do puts
			totalPuts := c.MbPutPerDay()
			for p := 0.0; p < totalPuts; p++ {
				n.DoRandomPut()
			}
			// do gets
			totalGets := c.MbGetPerDay()
			for g := 0.0; g < totalGets; g++ {
				n.DoRandomGet()
			}
		}
		// calculate average mb per safecoin
		mbPerSafecoin := 1.0 / n.AvgSafecoinPerMb()
		// add day to report
		line := fmt.Sprintf("%d,%d,%f,%d,%d\n", day, n.TotalSafecoins, mbPerSafecoin, n.TotalSections(), n.TotalVaults())
		fmt.Print(line)
		report = report + line
	}
	//fmt.Println(report)
}
