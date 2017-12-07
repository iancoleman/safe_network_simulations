package safenet

type Section struct {
	Prefix              Prefix
	TotalVaults         uint
	Vaults              map[*Vault]bool
	LeftTotalAdults     uint
	RightTotalAdults    uint
	TargetVault         *Vault
	TotalAttackedVaults uint
	IsAttacked          bool
	TotalAdults         uint
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
		if v.IsAdult {
			if leftPrefix.Matches(v.Name) {
				s.LeftTotalAdults = s.LeftTotalAdults + 1
			} else {
				s.RightTotalAdults = s.RightTotalAdults + 1
			}
		}
		// increment the age
		v.IncrementAge()
		// track adults
		if v.IsAdult {
			s.TotalAdults = s.TotalAdults + 1
		}
	}
	// set target vault
	s.setRandomTargetVault()
	// set stats
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
		s.LeftTotalAdults = s.LeftTotalAdults + 1
	} else {
		s.RightTotalAdults = s.RightTotalAdults + 1
	}
	// split into two groups if needed
	// details are handled by network upon returning two new sections
	if s.LeftTotalAdults >= SplitSize && s.RightTotalAdults >= SplitSize {
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
	s.setRandomTargetVault()
	// no split so return zero new sections
	return []*Section{}
}

func (s *Section) removeVault(v *Vault) {
	// remove from section
	s.TotalVaults = s.TotalVaults - 1
	delete(s.Vaults, v)
	// track hypothetical future section
	leftPrefix := s.Prefix.extendLeft()
	if v.IsAdult {
		if leftPrefix.Matches(v.Name) {
			s.LeftTotalAdults = s.LeftTotalAdults - 1
		} else {
			s.RightTotalAdults = s.RightTotalAdults - 1
		}
	}
	// track attack
	if v.IsAttacker {
		s.TotalAttackedVaults = s.TotalAttackedVaults - 1
		s.checkIfAttacked()
	}
	// set new target vault
	s.setRandomTargetVault()
	// merge is handled by network
}

func (s *Section) setRandomTargetVault() {
	testBytes := NewXorName()
	smallestDiff := 0
	isFirst := true
	for v := range s.Vaults {
		diff := 0
		for i := len(v.Name) - 1; i >= 0; i-- {
			if testBytes[i] > v.Name[i] {
				diff = diff + int(testBytes[i]-v.Name[i])
			} else {
				diff = diff + int(v.Name[i]-testBytes[i])
			}
		}
		if isFirst || diff < smallestDiff {
			s.TargetVault = v
			smallestDiff = diff
			isFirst = false
		}
	}
}

func (s *Section) checkIfAttacked() {
	attackPct := float64(s.TotalAttackedVaults) / float64(s.TotalVaults)
	quorumPct := float64(QuorumSize) / float64(GroupSize)
	s.IsAttacked = attackPct >= quorumPct
}
