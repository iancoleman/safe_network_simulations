package safenet

type Section struct {
	Prefix              Prefix
	TotalAdults         uint
	Vaults              map[*Vault]bool
	LeftTotalAdults     uint
	RightTotalAdults    uint
	TotalAttackedElders uint
	IsAttacked          bool
}

func newSection(prefix Prefix, vaults map[*Vault]bool) *Section {
	s := Section{
		Prefix: prefix,
		Vaults: map[*Vault]bool{},
	}
	// add each existing vault for new section data
	for v := range vaults {
		// increment the age
		v.IncrementAge()
		// add to section
		s.addVault(v)
	}
	// set total attacked elders now that all vaults are populated
	// because elder status depends on all other vaults
	for v := range s.Vaults {
		if v.IsAttacker {
			if s.VaultIsElder(v) {
				s.TotalAttackedElders = s.TotalAttackedElders + 1
			}
		}
	}
	// set stats
	s.checkIfAttacked()
	return &s
}

func (s *Section) addVault(v *Vault) []*Section {
	v.SetPrefix(s.Prefix)
	s.Vaults[v] = true
	// track hypothetical future group
	if v.IsAdult {
		s.TotalAdults = s.TotalAdults + 1
		leftPrefix := s.Prefix.extendLeft()
		if leftPrefix.Matches(v.Name) {
			s.LeftTotalAdults = s.LeftTotalAdults + 1
		} else {
			s.RightTotalAdults = s.RightTotalAdults + 1
		}
	}
	// split into two groups if needed
	// details are handled by network upon returning two new sections
	if s.LeftTotalAdults >= SplitSize && s.RightTotalAdults >= SplitSize {
		leftPrefix := s.Prefix.extendLeft()
		rightPrefix := s.Prefix.extendRight()
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
		s2 := newSection(rightPrefix, right)
		return []*Section{s1, s2}
	}
	// track attacking elder stats
	if v.IsAttacker {
		if s.VaultIsElder(v) {
			s.TotalAttackedElders = s.TotalAttackedElders + 1
		}
	}
	// track attack
	s.checkIfAttacked()
	// no split so return zero new sections
	return []*Section{}
}

func (s *Section) removeVault(v *Vault) {
	// track attacking elder stats
	if v.IsAttacker {
		if s.VaultIsElder(v) {
			s.TotalAttackedElders = s.TotalAttackedElders - 1
		}
	}
	// remove from section
	delete(s.Vaults, v)
	// track hypothetical future section
	if v.IsAdult {
		if s.TotalAdults > 0 {
			s.TotalAdults = s.TotalAdults - 1
		}
		leftPrefix := s.Prefix.extendLeft()
		if leftPrefix.Matches(v.Name) {
			s.LeftTotalAdults = s.LeftTotalAdults - 1
		} else {
			s.RightTotalAdults = s.RightTotalAdults - 1
		}
	}
	// track attack
	s.checkIfAttacked()
	// merge is handled by network
}

// The oldest GroupSize (8) adults are elders
func (s *Section) VaultIsElder(v *Vault) bool {
	if !v.IsAdult {
		return false
	}
	olderAdults := 0
	for sv := range s.Vaults {
		if sv.IsAdult {
			if sv.Age > v.Age {
				olderAdults = olderAdults + 1
			} else if sv.Age == v.Age {
				// TODO xorname tiebreaker for elder status
			}
		}
		if olderAdults >= GroupSize {
			return false
		}
	}
	return true
}

func (s *Section) checkIfAttacked() {
	// check if enough elders to control quorum
	s.IsAttacked = s.TotalAttackedElders > QuorumSize
}
