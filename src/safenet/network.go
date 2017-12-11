package safenet

import (
	"fmt"
	"math/rand"
)

const GroupSize = 8
const SplitBuffer = 3
const QuorumNumerator = 1
const QuorumDenominator = 2
const SplitSize = GroupSize + SplitBuffer

var prng = rand.New(rand.NewSource(0))

type Network struct {
	Sections           map[string]*Section
	TotalVaults        int
	TotalSections      int
	TotalMerges        int
	TotalSplits        int
	TotalVaultEvents   int
	TotalSectionEvents int
	TotalJoins         int
	TotalDepartures    int
	TotalRelocations   int
}

func NewNetwork() Network {
	return Network{
		Sections: map[string]*Section{},
	}
}

func NewNetworkFromSeed(seed int64) Network {
	prng = rand.New(rand.NewSource(seed))
	return NewNetwork()
}

func (n *Network) AddVault(v *Vault) {
	// track stats for total vaults
	n.TotalVaults = n.TotalVaults + 1
	n.TotalJoins = n.TotalJoins + 1
	n.TotalVaultEvents = n.TotalVaultEvents + 1
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
				n.TotalSections = n.TotalSections + 1
			}
		}
	}
	// add the vault to the section
	ne := section.addVault(v)
	// if there was a split
	if ne != nil && len(ne.NewSections) > 0 {
		n.TotalSplits = n.TotalSplits + 1
		n.TotalSectionEvents = n.TotalSectionEvents + 1
		// add new sections
		for _, s := range ne.NewSections {
			n.Sections[s.Prefix.Key] = s
			n.TotalSections = n.TotalSections + 1
		}
		// remove old section
		delete(n.Sections, section.Prefix.Key)
		n.TotalSections = n.TotalSections - 1
	}
	// relocate vault if there is one to relocate
	if ne != nil && ne.VaultToRelocate != nil {
		n.relocateVault(ne.VaultToRelocate)
	}
}

func (n *Network) RemoveVault(v *Vault) {
	n.TotalVaults = n.TotalVaults - 1
	n.TotalDepartures = n.TotalDepartures + 1
	n.TotalVaultEvents = n.TotalVaultEvents + 1
	section, exists := n.Sections[v.Prefix.Key]
	if !exists {
		fmt.Println("Warning: No section for removeVault", v.Prefix)
		return
	}
	// remove the vault from the section
	ne := section.removeVault(v)
	// merge if needed
	if section.TotalAdults < GroupSize && n.TotalSections > 1 {
		n.TotalMerges = n.TotalMerges + 1
		n.TotalSectionEvents = n.TotalSectionEvents + 1
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
			n.TotalSections = n.TotalSections - 1
		} else {
			// get child vaults
			childPrefixes := n.getChildPrefixes(siblingPrefix)
			for _, childPrefix := range childPrefixes {
				// merge child vault
				childVaults := n.Sections[childPrefix.Key].Vaults
				parentVaults = append(parentVaults, childVaults...)
				delete(n.Sections, childPrefix.Key)
				n.TotalSections = n.TotalSections - 1
			}
		}
		// remove the merged section
		delete(n.Sections, section.Prefix.Key)
		n.TotalSections = n.TotalSections - 1
		// create the new section
		ne := newSection(parentPrefix, parentVaults)
		if ne != nil {
			for _, s := range ne.NewSections {
				n.Sections[s.Prefix.Key] = s
				n.TotalSections = n.TotalSections + 1
			}
		}
	}
	// relocate vault if there is one to relocate
	if ne != nil && ne.VaultToRelocate != nil {
		n.relocateVault(ne.VaultToRelocate)
	}
}

func (n *Network) relocateVault(v *Vault) {
	v.IncrementAge()
	n.TotalRelocations = n.TotalRelocations + 1
	// TODO acutally relocate this vault to neighbour with fewest vaults
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
	// get random section
	s := n.GetRandomSection()
	// get random vault from section
	return s.GetRandomVault()
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
	} else if len(prefix.id) < 256 {
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
	xornameBitIndex := 0
	xornameByteIndex := 0
	for !exists && len(prefix.id) < xornameBits {
		// get the next bit of the xorname prefix
		xornameByteIndex = xornameBitIndex / 8
		maskBitIndex := uint(xornameBitIndex % 8)
		maskByte := byte(0x80) >> maskBitIndex
		// AND the byte with the mask to give 0 if the bit is 0 or maskByte if
		// it's 1
		bit := x.ByteAtIndex(xornameByteIndex) & maskByte
		// extend the prefix depending on the bit of the xorname
		if bit == 0 {
			prefix = prefix.extendLeft()
		} else {
			prefix = prefix.extendRight()
		}
		_, exists = n.Sections[prefix.Key]
		// update the next bit to check in the xorname
		xornameBitIndex = xornameBitIndex + 1
	}
	if !exists && n.TotalVaults > 1 {
		fmt.Println("Warning: No prefix for xorname")
		return NewBlankPrefix()
	}
	return prefix
}
