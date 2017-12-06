package safenet

type Section struct {
	Prefix              Prefix
	TotalVaults         int
	Vaults              map[*Vault]bool
	LeftTotalVaults     int
	RightTotalVaults    int
	TargetVault         *Vault
	TotalAttackedVaults int
	IsAttacked          bool
}

func newSection(prefix Prefix, vaults map[*Vault]bool) *Section {
	s := Section{
		Prefix:      prefix,
		TotalVaults: 0,
		Vaults:      vaults,
	}
	// update each vault for new section data
	leftPrefix := prefix.extendLeft()
	for v := range vaults {
		s.TotalVaults = s.TotalVaults + 1
		// set new prefix for vault
		v.SetPrefix(prefix)
		// track attack
		if v.IsAttacker {
			s.TotalAttackedVaults = s.TotalAttackedVaults + 1
		}
		// track hypothetical future groups
		if leftPrefix.Matches(v.Name) {
			s.LeftTotalVaults = s.LeftTotalVaults + 1
		} else {
			s.RightTotalVaults = s.RightTotalVaults + 1
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
	s.checkIfAttacked()
	return &s
}

func (s *Section) addVault(v *Vault) []*Section {
	v.SetPrefix(s.Prefix)
	s.Vaults[v] = true
	s.TotalVaults = s.TotalVaults + 1
	// track attack
	if v.IsAttacker {
		s.TotalAttackedVaults = s.TotalAttackedVaults + 1
	}
	s.checkIfAttacked()
	// track hypothetical future group
	leftPrefix := s.Prefix.extendLeft()
	if leftPrefix.Matches(v.Name) {
		s.LeftTotalVaults = s.LeftTotalVaults + 1
	} else {
		s.RightTotalVaults = s.RightTotalVaults + 1
	}
	// split into two groups if needed
	// details are handled by network upon returning two new sections
	if s.LeftTotalVaults >= SplitSize && s.RightTotalVaults >= SplitSize {
		left := map[*Vault]bool{}
		right := map[*Vault]bool{}
		for v := range s.Vaults {
			if leftPrefix.Matches(v.Name) {
				left[v] = true
			} else {
				right[v] = true
			}
		}
		s1 := newSection(leftPrefix, left)
		rightPrefix := s.Prefix.extendRight()
		s2 := newSection(rightPrefix, right)
		return []*Section{s1, s2}
	}
	// set target vault
	if s.TargetVault == nil || v.Name.IsBefore(s.TargetVault.Name) {
		s.TargetVault = v
	}
	// no split so return zero new sections
	return []*Section{}
}

func (s *Section) removeVault(v *Vault) {
	// remove from section
	s.TotalVaults = s.TotalVaults - 1
	delete(s.Vaults, v)
	// track hypothetical future section
	leftPrefix := s.Prefix.extendLeft()
	if leftPrefix.Matches(v.Name) {
		s.LeftTotalVaults = s.LeftTotalVaults - 1
	} else {
		s.RightTotalVaults = s.RightTotalVaults - 1
	}
	// track attack
	if v.IsAttacker {
		s.TotalAttackedVaults = s.TotalAttackedVaults - 1
		s.checkIfAttacked()
	}
	// set new target vault if needed
	if v == s.TargetVault {
		s.setNewTargetVault()
	}
	// merge is handled by network
}

func (s *Section) setNewTargetVault() {
	isFirst := true
	for v := range s.Vaults {
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

func (s *Section) checkIfAttacked() {
	attackPct := float64(s.TotalAttackedVaults) / float64(s.TotalVaults)
	quorumPct := float64(QuorumSize) / float64(GroupSize)
	s.IsAttacked = attackPct >= quorumPct
}
