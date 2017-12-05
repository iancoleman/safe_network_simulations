package safenet

import (
	"fmt"
)

const GroupSize = 8
const SplitBuffer = 3
const QuorumSize = 5
const SplitSize = GroupSize + SplitBuffer

type Network struct {
	Sections           map[Prefix]*Section
	TargetPrefix       Prefix
	TotalVaults        int
	TotalSections      int
	TotalMerges        int
	TotalSplits        int
	TotalVaultEvents   int
	TotalSectionEvents int
	TotalJoins         int
	TotalDepartures    int
}

func NewNetwork() Network {
	return Network{
		Sections: map[Prefix]*Section{},
	}
}

func (n *Network) AddVault(v *Vault) {
	// track stats for total vaults
	n.TotalVaults = n.TotalVaults + 1
	n.TotalJoins = n.TotalJoins + 1
	n.TotalVaultEvents = n.TotalVaultEvents + 1
	// get prefix for vault
	prefix := n.getPrefixForXorname(v.Name)
	section, exists := n.Sections[prefix]
	// get the section for this prefix
	if !exists {
		var blankPrefix Prefix
		section = newSection(blankPrefix, map[string]*Vault{})
		n.Sections[section.Prefix] = section
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
			n.Sections[s.Prefix] = s
			n.TotalSections = n.TotalSections + 1
		}
		// remove old section
		delete(n.Sections, section.Prefix)
		n.TotalSections = n.TotalSections - 1
	}
	n.updateTargetSection()
}

func (n *Network) RemoveVault(v *Vault) {
	n.TotalVaults = n.TotalVaults - 1
	n.TotalDepartures = n.TotalDepartures + 1
	n.TotalVaultEvents = n.TotalVaultEvents + 1
	section, exists := n.Sections[v.Prefix]
	if !exists {
		fmt.Println("Warning: No section for removeVault", v.Prefix)
		return
	}
	// remove the vault from the section
	section.removeVault(v)
	// merge if needed
	if section.TotalVaults < GroupSize {
		n.TotalMerges = n.TotalMerges + 1
		n.TotalSectionEvents = n.TotalSectionEvents + 1
		parentPrefix := section.Prefix[:len(section.Prefix)-1]
		// get sibling prefix
		siblingBit := Prefix("0")
		prefixLastBit := section.Prefix[len(section.Prefix)-1]
		if prefixLastBit == '0' {
			siblingBit = "1"
		}
		siblingPrefix := parentPrefix + siblingBit
		// get sibling vaults
		parentVaults := section.Vaults
		_, exists := n.Sections[siblingPrefix]
		if exists {
			// merge sibling
			siblingVaults := n.Sections[siblingPrefix].Vaults
			for _, v := range siblingVaults {
				parentVaults[v.Name.binary] = v
			}
			delete(n.Sections, siblingPrefix)
			n.TotalSections = n.TotalSections - 1
		} else {
			// get child vaults
			childPrefixes := n.getChildPrefixes(siblingPrefix)
			for _, childPrefix := range childPrefixes {
				// merge child vault
				childVaults := n.Sections[childPrefix].Vaults
				for _, v := range childVaults {
					parentVaults[v.Name.binary] = v
				}
				delete(n.Sections, childPrefix)
				n.TotalSections = n.TotalSections - 1
			}
		}
		// remove the merged section
		delete(n.Sections, section.Prefix)
		n.TotalSections = n.TotalSections - 1
		// create the new section
		s := newSection(parentPrefix, parentVaults)
		n.Sections[s.Prefix] = s
		n.TotalSections = n.TotalSections + 1
	}
	n.updateTargetSection()
}

func (n *Network) GetRandomVault() *Vault {
	return n.Sections[n.TargetPrefix].TargetVault
}

// Sets the target section to a new section
// by generating a new random xorname and setting the
// prefix for that new name as the target section
func (n *Network) updateTargetSection() {
	x := NewXorName()
	prefix := n.getPrefixForXorname(x)
	n.TargetPrefix = prefix
}

func (n *Network) getChildPrefixes(prefix Prefix) []Prefix {
	prefixes := []Prefix{}
	leftPrefix := prefix + "0"
	rightPrefix := prefix + "1"
	_, leftExists := n.Sections[leftPrefix]
	_, rightExists := n.Sections[rightPrefix]
	if leftExists && rightExists {
		prefixes = append(prefixes, leftPrefix, rightPrefix)
	} else if leftExists {
		prefixes = append(prefixes, leftPrefix)
		prefixes = append(prefixes, n.getChildPrefixes(rightPrefix)...)
	} else if rightExists {
		prefixes = append(prefixes, rightPrefix)
		prefixes = append(prefixes, n.getChildPrefixes(leftPrefix)...)
	} else if len(prefix) < 256 {
		prefixes = append(prefixes, n.getChildPrefixes(leftPrefix)...)
		prefixes = append(prefixes, n.getChildPrefixes(rightPrefix)...)
	} else {
		fmt.Println("Warning: No children exist for prefix")
	}
	return prefixes
}

func (n *Network) getPrefixForXorname(x XorName) Prefix {
	var prefix Prefix
	_, exists := n.Sections[prefix]
	for !exists && len(prefix) < len(x.binary) {
		prefix = prefix + Prefix(x.binary[len(prefix)])
		_, exists = n.Sections[prefix]
	}
	return prefix
}
