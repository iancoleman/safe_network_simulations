package safenet

import (
	"fmt"
)

type Section struct {
	Prefix           Prefix
	TotalVaults      int
	Vaults           map[string]*Vault
	LeftPrefix       Prefix
	RightPrefix      Prefix
	LeftTotalVaults  int
	Left             map[string]*Vault
	RightTotalVaults int
	Right            map[string]*Vault
	TargetVault      *Vault
}

func newSection(prefix Prefix, vaults map[string]*Vault) *Section {
	s := Section{
		Prefix:      prefix,
		LeftPrefix:  prefix + "0",
		RightPrefix: prefix + "1",
		TotalVaults: 0,
		Vaults:      vaults,
		Left:        map[string]*Vault{},
		Right:       map[string]*Vault{},
	}
	// track target vault
	for _, v := range vaults {
		s.TotalVaults = s.TotalVaults + 1
		// set new prefix
		v.SetPrefix(prefix)
		// form hypothetical future groups
		if v.Name.StartsWith(s.LeftPrefix) {
			s.Left[v.Name.binary] = v
			s.LeftTotalVaults = s.LeftTotalVaults + 1
		} else if v.Name.StartsWith(s.RightPrefix) {
			s.Right[v.Name.binary] = v
			s.RightTotalVaults = s.RightTotalVaults + 1
		} else {
			fmt.Println("Warning: New section vault has no hypothetical group")
		}
		// set target vault if is smallest
		if s.TotalVaults == 1 {
			s.TargetVault = v
		} else {
			if v.Name.IsBefore(s.TargetVault.Name) {
				s.TargetVault = v
			}
		}
	}
	return &s
}

func (s *Section) addVault(v *Vault) []*Section {
	v.SetPrefix(s.Prefix)
	s.Vaults[v.Name.binary] = v
	s.TotalVaults = s.TotalVaults + 1
	// add to hypothetical future group
	if v.Name.StartsWith(s.LeftPrefix) {
		s.Left[v.Name.binary] = v
		s.LeftTotalVaults = s.LeftTotalVaults + 1
	} else if v.Name.StartsWith(s.RightPrefix) {
		s.Right[v.Name.binary] = v
		s.RightTotalVaults = s.RightTotalVaults + 1
	} else {
		fmt.Println("Warning: New vault has no hypothetical future group")
	}
	// split if needed
	if s.LeftTotalVaults >= SplitSize && s.RightTotalVaults >= SplitSize {
		s1 := newSection(s.LeftPrefix, s.Left)
		s2 := newSection(s.RightPrefix, s.Right)
		return []*Section{s1, s2}
	}
	// set target vault
	if s.TargetVault == nil || v.Name.IsBefore(s.TargetVault.Name) {
		s.TargetVault = v
	}
	return []*Section{}
}

func (s *Section) removeVault(v *Vault) {
	s.TotalVaults = s.TotalVaults - 1
	delete(s.Vaults, v.Name.binary)
	// set new target vault if needed
	if v == s.TargetVault {
		s.setNewTargetVault()
	}
	// merge is handled by network
}

func (s *Section) setNewTargetVault() {
	isFirst := true
	for _, v := range s.Vaults {
		if isFirst {
			s.TargetVault = v
			isFirst = false
		} else {
			if v.Name.IsBefore(s.TargetVault.Name) {
				s.TargetVault = v
			}
		}
	}
}
