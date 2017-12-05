# Safe Network Simulations

Simulates some scenarios for vault joining / leaving on the [safe network](https://safenetwork.org/).

## Section Size Distribution

Outputs the number of groups of various sizes in the simulated network.

Read more [on the safenet forum](https://safenetforum.org/t/explaining-group-splits-and-merges/18383)

### Usage

```
$ cd /path/to/safe_network_simulations
$ export GOPATH=/path/to/safe_network_simulations
$ go run section_size_distribution.go
```

Once the simulation is complete a report will be printed to stdout.
