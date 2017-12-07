package safenet

import (
	"fmt"
	"math/rand"
)

const GroupSize = 8
const SplitBuffer = 3
const QuorumSize = 5
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
		section = newSection(blankPrefix, map[*Vault]bool{})
		n.Sections[section.Prefix.Key] = section
		n.TotalSections = n.TotalSections + 1
	}
	// add the vault to the section
	newSections := section.addVault(v)
	// if there was a split
	if len(newSections) > 0 {
		n.TotalSplits = n.TotalSplits + 1
		n.TotalSectionEvents = n.TotalSectionEvents + 1
		// add new sections
		for _, s := range newSections {
			n.Sections[s.Prefix.Key] = s
			n.TotalSections = n.TotalSections + 1
		}
		// remove old section
		delete(n.Sections, section.Prefix.Key)
		n.TotalSections = n.TotalSections - 1
	}
}

func (n *Network) RelocateVault(v *Vault) {
	// track stats
	n.TotalRelocations = n.TotalRelocations + 1
	n.TotalVaultEvents = n.TotalVaultEvents + 1
	v.IncrementAge()
	if n.TotalSections > 1 {
		// remove from current section
		n.RemoveVault(v)
		// rename it to give a new location on the network
		v.Rename()
		// add to appropriate section
		n.AddVault(v)
	}
}

func (n *Network) AddOrRelocateVault() {
	if prng.Float32() < 0.9 {
		v := NewVault()
		n.AddVault(v)
	} else {
		v := n.GetRandomVault()
		if v != nil {
			n.RelocateVault(v)
		}
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
	section.removeVault(v)
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
			for v := range siblingVaults {
				parentVaults[v] = true
			}
			delete(n.Sections, siblingPrefix.Key)
			n.TotalSections = n.TotalSections - 1
		} else {
			// get child vaults
			childPrefixes := n.getChildPrefixes(siblingPrefix)
			for _, childPrefix := range childPrefixes {
				// merge child vault
				childVaults := n.Sections[childPrefix.Key].Vaults
				for v := range childVaults {
					parentVaults[v] = true
				}
				delete(n.Sections, childPrefix.Key)
				n.TotalSections = n.TotalSections - 1
			}
		}
		// remove the merged section
		delete(n.Sections, section.Prefix.Key)
		n.TotalSections = n.TotalSections - 1
		// create the new section
		s := newSection(parentPrefix, parentVaults)
		n.Sections[s.Prefix.Key] = s
		n.TotalSections = n.TotalSections + 1
	}
}

func (n *Network) GetRandomVault() *Vault {
	x := NewXorName()
	p := n.getPrefixForXorname(x)
	s, exists := n.Sections[p.Key]
	if !exists {
		return nil
	}
	var min XorDistance
	var target *Vault
	for v := range s.Vaults {
		d := v.Name.XorDistanceTo(x)
		if min.IsZeroValue() || d.IsLessThan(min) {
			min = d
			target = v
		}
	}
	return target
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
	xornameBitIndex := uint(0)
	xornameByteIndex := uint(0)
	for !exists && len(prefix.id) < xornameBits {
		// get the next bit of the xorname prefix
		xornameByteIndex = xornameBitIndex / 8
		maskBitIndex := xornameBitIndex % 8
		maskByte := byte(0x80) >> maskBitIndex
		// AND the byte with the mask to give 0 if the bit is 0 or maskByte if
		// it's 1
		bit := x[xornameByteIndex] & maskByte
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
