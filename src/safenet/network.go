package safenet

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

const GroupSize = 8
const SplitBuffer = 3
const QuorumNumerator = 1
const QuorumDenominator = 2
const SplitSize = GroupSize + SplitBuffer
const MaxSafecoins = 4294967296 // 2^32

var prng = rand.New(rand.NewSource(0))

type Network struct {
	Sections          map[string]*Section
	TotalMerges       int
	TotalSplits       int
	TotalJoins        int
	TotalDepartures   int
	TotalRelocations  int
	NeighbourhoodHops []int
	TotalSafecoins    int
}

func NewNetwork() Network {
	return Network{
		Sections:          map[string]*Section{},
		NeighbourhoodHops: []int{},
	}
}

func NewNetworkFromSeed(seed int64) Network {
	prng = rand.New(rand.NewSource(seed))
	return NewNetwork()
}

func (n *Network) AddVault(v *Vault) bool {
	// track stats
	n.TotalJoins = n.TotalJoins + 1
	// get prefix for vault
	prefix := n.getPrefixForXorname(v.Name)
	section, exists := n.Sections[prefix.Key]
	// get the section for this prefix
	if !exists {
		blankPrefix := NewBlankPrefix()
		ne := newSection(blankPrefix, []*Vault{})
		if ne != nil {
			for _, section = range ne.NewSections {
				n.Sections[section.Prefix.Key] = section
			}
		}
	}
	// add the vault to the section
	ne, disallowed := section.addVault(v)
	// if there was a split
	if ne != nil && len(ne.NewSections) > 0 {
		n.TotalSplits = n.TotalSplits + 1
		// add new sections
		for _, s := range ne.NewSections {
			n.Sections[s.Prefix.Key] = s
		}
		// remove old section
		delete(n.Sections, section.Prefix.Key)
	}
	// relocate vault if there is one to relocate
	if ne != nil && ne.VaultToRelocate != nil {
		n.relocateVault(ne)
	}
	return disallowed
}

func (n *Network) RemoveVault(v *Vault) {
	n.TotalDepartures = n.TotalDepartures + 1
	section, exists := n.Sections[v.Prefix.Key]
	if !exists {
		fmt.Println("Warning: No section for removeVault")
		return
	}
	// remove the vault from the section
	ne := section.removeVault(v)
	// merge if needed
	if section.shouldMerge() && n.HasMoreThanOneSection() {
		n.TotalMerges = n.TotalMerges + 1
		parentPrefix := section.Prefix.parent()
		// get sibling prefix
		siblingPrefix := v.Prefix.sibling()
		// get sibling vaults
		parentVaults := section.Vaults
		_, exists := n.Sections[siblingPrefix.Key]
		if exists {
			// merge sibling
			siblingVaults := n.Sections[siblingPrefix.Key].Vaults
			parentVaults = append(parentVaults, siblingVaults...)
			delete(n.Sections, siblingPrefix.Key)
		} else {
			// get child vaults
			childPrefixes := n.getChildPrefixes(siblingPrefix)
			for _, childPrefix := range childPrefixes {
				// merge child vault
				childVaults := n.Sections[childPrefix.Key].Vaults
				parentVaults = append(parentVaults, childVaults...)
				delete(n.Sections, childPrefix.Key)
			}
		}
		// remove the merged section
		delete(n.Sections, section.Prefix.Key)
		// create the new section
		ne := newSection(parentPrefix, parentVaults)
		if ne != nil {
			for _, s := range ne.NewSections {
				n.Sections[s.Prefix.Key] = s
			}
		}
	} else if ne != nil && ne.VaultToRelocate != nil {
		// if there is no merge but there is a vault to relocate,
		// relocate the vault
		n.relocateVault(ne)
	}
}

func (n *Network) relocateVault(ne *NetworkEvent) {
	// track stats for relocations
	n.TotalRelocations = n.TotalRelocations + 1
	// find the neighbour with shortest prefix or fewest vaults
	// default to the existing section, useful for zero-length prefix
	smallestNeighbour := n.Sections[ne.VaultToRelocate.Prefix.Key]
	minNeighbourPrefix := math.MaxUint32
	minNeighbourVaults := math.MaxUint32
	// get all neighbours
	for i := 0; i < len(ne.VaultToRelocate.Prefix.bits); i++ {
		// clone the prefix but flip the ith bit of the prefix
		neighbourPrefix := NewBlankPrefix()
		for j := 0; j < len(ne.VaultToRelocate.Prefix.bits); j++ {
			isZero := !ne.VaultToRelocate.Prefix.bits[j]
			if j == i {
				isZero = !isZero
			}
			if isZero {
				neighbourPrefix = neighbourPrefix.extendLeft()
			} else {
				neighbourPrefix = neighbourPrefix.extendRight()
			}
		}
		// get neighbouring prefixes from the network for this prefix
		// and repeat until we arrive at the 'best' neighbour prefix
		prevNeighbourPrefix := NewBlankPrefix()
		for !neighbourPrefix.Equals(prevNeighbourPrefix) {
			// track previous neighbour prefix
			prevNeighbourPrefix = neighbourPrefix
			// get potential new neighbour prefixes
			neighbourPrefixes := n.getMatchingPrefixes(neighbourPrefix)
			// check if these neighbours contain the 'best' neighbour
			// prioritise sections with shorter prefixes and having less nodes to balance the network
			for _, p := range neighbourPrefixes {
				s := n.Sections[p.Key]
				if len(p.bits) < minNeighbourPrefix {
					// prefer shorter prefixes
					neighbourPrefix = p
					minNeighbourPrefix = len(p.bits)
					smallestNeighbour = s
				} else if len(p.bits) == minNeighbourPrefix {
					// prefer less vaults if prefix length is same
					if len(s.Vaults) < minNeighbourVaults {
						neighbourPrefix = p
						minNeighbourVaults = len(s.Vaults)
						smallestNeighbour = s
					} else if len(s.Vaults) == minNeighbourVaults {
						// TODO tiebreaker for equal sized neighbours
						// see https://forum.safedev.org/t/data-chains-deeper-dive/1209
						// If all neighbours have the same number of peers we relocate
						// to the section closest to the H above (that is not us)
					}
				}
			}
		}
	}
	// track neighbourhood hops by comparing how many bits differ
	// between the new and the old prefix.
	neighbourhoodHops := 0
	prefixLength := len(smallestNeighbour.Prefix.Key)
	if len(ne.VaultToRelocate.Prefix.Key) < prefixLength {
		prefixLength = len(ne.VaultToRelocate.Prefix.Key)
	}
	for i := 0; i < prefixLength; i++ {
		newBit := smallestNeighbour.Prefix.Key[i]
		oldBit := ne.VaultToRelocate.Prefix.Key[i]
		if newBit != oldBit {
			neighbourhoodHops = neighbourhoodHops + 1
		}
	}
	n.NeighbourhoodHops = append(n.NeighbourhoodHops, neighbourhoodHops)
	// remove vault from current section (includes merge if needed)
	n.RemoveVault(ne.VaultToRelocate)
	// adjust vault name to match the neighbour section prefix
	ne.VaultToRelocate.renameWithPrefix(smallestNeighbour.Prefix)
	// age the relocated vault
	ne.VaultToRelocate.IncrementAge()
	// relocate the vault to the smallest neighbour (includes split if needed)
	disallowed := n.AddVault(ne.VaultToRelocate)
	if disallowed {
		fmt.Println("Warning: disallowed relocated vault")
	}
}

func (n *Network) GetRandomSection() *Section {
	x := NewXorName()
	p := n.getPrefixForXorname(x)
	s, _ := n.Sections[p.Key]
	return s
}

// Needs to be deterministic but also random.
// Iterating over keys of a map is not deterministic
func (n *Network) GetRandomVault() *Vault {
	s := n.GetRandomSection()
	return s.GetRandomVault()
}

// Returns the parent, prefix, or children that matches this prefix on the network
func (n *Network) getMatchingPrefixes(prefix Prefix) []Prefix {
	prefixes := []Prefix{}
	testPrefix := NewBlankPrefix()
	// find possible parents
	_, exists := n.Sections[testPrefix.Key]
	if exists {
		prefixes = append(prefixes, testPrefix)
	}
	for i := 0; i < len(prefix.bits); i++ {
		if !prefix.bits[i] {
			testPrefix = testPrefix.extendLeft()
		} else {
			testPrefix = testPrefix.extendRight()
		}
		_, exists := n.Sections[testPrefix.Key]
		if exists {
			prefixes = append(prefixes, testPrefix)
			// TODO can probably break here?
		}
	}
	// get child prefixes if no parent found
	if len(prefixes) == 0 {
		prefixes = n.getChildPrefixes(prefix)
	}
	return prefixes
}

func (n *Network) getChildPrefixes(prefix Prefix) []Prefix {
	prefixes := []Prefix{}
	leftPrefix := prefix.extendLeft()
	rightPrefix := prefix.extendRight()
	_, leftExists := n.Sections[leftPrefix.Key]
	_, rightExists := n.Sections[rightPrefix.Key]
	if leftExists && rightExists {
		prefixes = append(prefixes, leftPrefix, rightPrefix)
	} else if leftExists {
		prefixes = append(prefixes, leftPrefix)
		prefixes = append(prefixes, n.getChildPrefixes(rightPrefix)...)
	} else if rightExists {
		prefixes = append(prefixes, rightPrefix)
		prefixes = append(prefixes, n.getChildPrefixes(leftPrefix)...)
	} else if len(prefix.bits) < 256 {
		prefixes = append(prefixes, n.getChildPrefixes(leftPrefix)...)
		prefixes = append(prefixes, n.getChildPrefixes(rightPrefix)...)
	} else {
		fmt.Println("Warning: No children exist for prefix")
	}
	return prefixes
}

func (n *Network) getPrefixForXorname(x XorName) Prefix {
	prefix := NewBlankPrefix()
	_, exists := n.Sections[prefix.Key]
	for !exists && len(prefix.bits) < len(x.bits) {
		// get the next bit of the xorname prefix
		bit := x.bits[len(prefix.bits)]
		// extend the prefix depending on the bit of the xorname
		if bit == false {
			prefix = prefix.extendLeft()
		} else {
			prefix = prefix.extendRight()
		}
		_, exists = n.Sections[prefix.Key]
	}
	if !exists && n.HasMoreThanOneVault() {
		fmt.Println("Warning: No prefix for xorname")
		return NewBlankPrefix()
	}
	return prefix
}

func (n *Network) ReportAges() (map[int]int, []int) {
	ages := map[int]int{}
	ageKeys := []int{}
	count := 0
	for p := range n.Sections {
		for _, v := range n.Sections[p].Vaults {
			count = count + 1
			_, exists := ages[v.Age]
			if !exists {
				ages[v.Age] = 0
				ageKeys = append(ageKeys, v.Age)
			}
			ages[v.Age] = ages[v.Age] + 1
		}
	}
	sort.Sort(sort.IntSlice(ageKeys))
	return ages, ageKeys
}

func (n *Network) TotalVaults() int {
	vaults := 0
	for p := range n.Sections {
		vaults = vaults + len(n.Sections[p].Vaults)
	}
	return vaults
}

func (n *Network) TotalSections() int {
	sections := 0
	for range n.Sections {
		sections = sections + 1
	}
	return sections
}

func (n *Network) HasMoreThanOneVault() bool {
	vaults := 0
	for p := range n.Sections {
		vaults = vaults + len(n.Sections[p].Vaults)
		if vaults > 1 {
			return true
		}
	}
	return false
}

func (n *Network) HasMoreThanOneSection() bool {
	sections := 0
	for range n.Sections {
		sections = sections + 1
		if sections > 1 {
			return true
		}
	}
	return false
}

func (n *Network) HasOneSection() bool {
	sections := 0
	for range n.Sections {
		sections = sections + 1
		if sections > 1 {
			return false
		}
	}
	return sections == 1
}

func (n *Network) DoRandomPut() {
	chunkName := NewXorName()
	prefix := n.getPrefixForXorname(chunkName)
	section := n.Sections[prefix.Key]
	// the section knows it has stored 'some chunk' but doesn't care about the
	// details of that chunk. The name is only important for deciding the
	// target section.
	section.PutChunk()
}

func (n *Network) DoRandomGet() {
	// get triggers opportunity to farm.
	// check the opportunity passes the farm rate test.
	// see https://github.com/maidsafe/rfcs/blob/master/text/0012-safecoin-implementation/0012-safecoin-implementation.md#farm-request-calculation
	chunkName := NewXorName()
	prefix := n.getPrefixForXorname(chunkName)
	section := n.Sections[prefix.Key]
	farmDivisor := section.FarmDivisor()
	if farmDivisor > 0 {
		chunkHash := NewXorName() // simulated hash of PmidHolderName + chunkHame
		testPasses := bigIntModInt64IsZero(chunkHash.bigint, farmDivisor)
		if !testPasses {
			return
		}
	}
	// try creating the coin if it doesn't exist yet
	// do it statistically based on percent of total safecoin issued
	exists := prng.Float64() < float64(n.TotalSafecoins)/float64(MaxSafecoins)
	if !exists {
		n.TotalSafecoins = n.TotalSafecoins + 1
	}
}

func (n *Network) AvgSafecoinPerMb() float64 {
	var sum float64
	var sections float64
	for _, s := range n.Sections {
		sum = s.SafecoinPerMb()
		sections = sections + 1
	}
	return sum / sections
}

// Always works, does not test if it's possible or not due to maxsafecoins
func (n *Network) ForceCreateSafecoin() {
	n.TotalSafecoins = n.TotalSafecoins + 1
}
