package main

import (
	"fmt"
	"safenet"
	"time"
)

const numIcoCoins = 452552412
const growthRate = 1.003 // 0.3% growth every day

func main() {
	// create network
	n := safenet.NewNetwork()
	// initialize ICO coins
	fmt.Println("Initializing ICO coins")
	initIcoCoins(&n)
	fmt.Println()
	// initialize 1000 MaidSafe vaults
	maidsafeClient := safenet.NewConsistentClient()
	n.AddClient(maidsafeClient)
	for n.TotalVaults() < 1000 {
		vaults := maidsafeClient.NewVaultsToStart()
		for _, v := range vaults {
			n.AddVault(v)
		}
	}
	// initialize report
	report := "endOfDay,totalSafecoin,mbPerSafecoin,farmDivisor,totalSections,totalVaults,totalClients,secondsToSimulate\n"
	// calculate average mb per safecoin
	mbPerSafecoin := 1.0 / n.AvgSafecoinPerMb()
	farmDivisor := n.AvgFarmDivisor()
	report = report + fmt.Sprintf("%d,%d,%f,%f,%d,%d,%d,%f\n", 0, n.TotalSafecoins(), mbPerSafecoin, farmDivisor, n.TotalSections(), n.TotalVaults(), n.TotalClients(), 0.0)
	// report current state
	fmt.Print(report)
	// simulate the network activity by creating clients
	days := 100000
	for day := 1; day < days; day++ {
		// create new clients
		startTimer := time.Now()
		newClientsForToday := int(float64(n.TotalClients()) * (growthRate - 1))
		for i := 0; i < newClientsForToday; i++ {
			c := safenet.NewRandomClient()
			n.AddClient(c)
		}
		// do each client activity
		// TODO interleave the activity so the early clients do not benefit more than
		// the later clients.
		for _, c := range n.Clients {
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
				n.DoRandomPut(c)
			}
			// do gets
			totalGets := c.MbGetPerDay()
			for g := 0.0; g < totalGets; g++ {
				n.DoRandomGet()
			}
		}
		// calculate average mb per safecoin
		mbPerSafecoin := 1.0 / n.AvgSafecoinPerMb()
		farmDivisor := n.AvgFarmDivisor()
		// get timing stats
		timeToSimulate := time.Now().Sub(startTimer).Seconds()
		// add day to report
		line := fmt.Sprintf("%d,%d,%f,%f,%d,%d,%d,%f\n", day, n.TotalSafecoins(), mbPerSafecoin, farmDivisor, n.TotalSections(), n.TotalVaults(), n.TotalClients(), timeToSimulate)
		fmt.Print(line)
		report = report + line
	}
	//fmt.Println(report)
}

func initIcoCoins(n *safenet.Network) {
	// create clients and distribute safecoins based on distribution at
	// https://omniexplorer.info/spstats.aspx?sp=3
	// create clients for 0-10 coins
	distribution := make([][]int, 0)
	// distribution is slice of [ clients, min, max ] coins per client
	distribution = append(distribution, []int{1271, 0, 10})
	distribution = append(distribution, []int{1850, 10, 100})
	distribution = append(distribution, []int{3865, 100, 1000})
	distribution = append(distribution, []int{3697, 1000, 10000})
	distribution = append(distribution, []int{1640, 10000, 100000})
	var totalDistributed int32
	for _, d := range distribution {
		fmt.Println("Distributing", d[1], "-", d[2], "coins to", d[0], "ICO clients")
		for i := 0; i < d[0]; i++ {
			c := safenet.NewRandomClient()
			// TODO use random distribution instead of average
			coins := int32((d[1] + d[2]) / 2)
			c.AllocateSafecoins(coins)
			n.AddClient(c)
			totalDistributed = totalDistributed + coins
		}
	}
	// distribute coins to top 50 clients
	topHolders := []int32{
		95000000,
		34983416,
		33136516,
		10912998,
		7096283,
		5000000,
		4849734,
		4491018,
		4097871,
		3968884,
		3569998,
		2949940,
		2808400,
		2709555,
		2602997,
		2463680,
		2276003,
		2023000,
		1592184,
		1585786,
		1500000,
		1450000,
		1444563,
		1309000,
		1300000,
		1250000,
		1213870,
		1200000,
		1190003,
		1190000,
		1189993,
		1153593,
		1076107,
		1043259,
		1010012,
		1000000,
		1000000,
		1000000,
		1000000,
		970410,
		952000,
		952000,
		938993,
		928200,
		912380,
		892873,
		888517,
		851111,
		800000,
		795000,
	}
	fmt.Println("Distributing to top 50 coin holders")
	for _, coins := range topHolders {
		c := safenet.NewRandomClient()
		c.AllocateSafecoins(coins)
		n.AddClient(c)
		totalDistributed = totalDistributed + coins
	}
	// distribute remaining coins between remaining rich clients
	remaining := numIcoCoins - totalDistributed
	totalRichClients := 450 - len(topHolders)
	coins := int32(float64(remaining) / float64(totalRichClients))
	fmt.Println("Distributing", coins, "coins each to", totalRichClients-1, "rich ICO clients")
	for i := 0; i < totalRichClients-1; i++ {
		c := safenet.NewRandomClient()
		c.AllocateSafecoins(coins)
		n.AddClient(c)
		totalDistributed = totalDistributed + coins
	}
	// distribute remaining coins to last rich client
	coins = numIcoCoins - totalDistributed
	fmt.Println("Distributing", coins, "remaining coins to final rich ICO client")
	c := safenet.NewRandomClient()
	c.AllocateSafecoins(coins)
	n.AddClient(c)
	totalDistributed = totalDistributed + coins
	// Log the result of issuing ico coins
	fmt.Println("Issued", totalDistributed, "of", numIcoCoins, "ICO coins to", n.TotalClients(), "clients")
}
