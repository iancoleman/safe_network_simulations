package safenet

type Section struct {
	Prefix              Prefix
	TotalAdults         uint
	Vaults              []*Vault
	LeftTotalAdults     uint
	RightTotalAdults    uint
	TotalAttackedElders uint
	IsAttacked          bool
}

// Returns a slice of sections since as vaults age they may cascade into
// multiple sections.
func newSection(prefix Prefix, vaults []*Vault) []*Section {
	s := Section{
		Prefix: prefix,
		Vaults: []*Vault{},
	}
	// add each existing vault for new section data
	for _, v := range vaults {
		// increment the age
		v.IncrementAge()
		// add to section
		v.SetPrefix(s.Prefix)
		s.Vaults = append(s.Vaults, v)
		// track hypothetical future section
		if v.IsAdult {
			s.TotalAdults = s.TotalAdults + 1
			leftPrefix := s.Prefix.extendLeft()
			if leftPrefix.Matches(v.Name) {
				s.LeftTotalAdults = s.LeftTotalAdults + 1
			} else {
				s.RightTotalAdults = s.RightTotalAdults + 1
			}
		}
	}
	// split into two sections if needed
	if s.shouldSplit() {
		leftPrefix := s.Prefix.extendLeft()
		rightPrefix := s.Prefix.extendRight()
		left := []*Vault{}
		right := []*Vault{}
		for _, v := range s.Vaults {
			if leftPrefix.Matches(v.Name) {
				left = append(left, v)
			} else {
				right = append(right, v)
			}
		}
		s1 := newSection(leftPrefix, left)
		s2 := newSection(rightPrefix, right)
		sections := []*Section{}
		sections = append(sections, s1...)
		sections = append(sections, s2...)
		return sections
	}
	// track attacking elder stats
	for _, v := range s.Vaults {
		if v.IsAttacker {
			if s.VaultIsElder(v) {
				s.TotalAttackedElders = s.TotalAttackedElders + 1
			}
		}
	}
	// track attack
	s.checkIfAttacked()
	// return the section
	return []*Section{&s}
}

func (s *Section) shouldSplit() bool {
	return s.LeftTotalAdults >= SplitSize && s.RightTotalAdults >= SplitSize
}

func (s *Section) addVault(v *Vault) []*Section {
	v.SetPrefix(s.Prefix)
	s.Vaults = append(s.Vaults, v)
	// track hypothetical future section
	if v.IsAdult {
		s.TotalAdults = s.TotalAdults + 1
		leftPrefix := s.Prefix.extendLeft()
		if leftPrefix.Matches(v.Name) {
			s.LeftTotalAdults = s.LeftTotalAdults + 1
		} else {
			s.RightTotalAdults = s.RightTotalAdults + 1
		}
	}
	// split into two sections if needed
	// details are handled by network upon returning two new sections
	if s.shouldSplit() {
		leftPrefix := s.Prefix.extendLeft()
		rightPrefix := s.Prefix.extendRight()
		left := []*Vault{}
		right := []*Vault{}
		for _, v := range s.Vaults {
			if leftPrefix.Matches(v.Name) {
				left = append(left, v)
			} else {
				right = append(right, v)
			}
		}
		s1 := newSection(leftPrefix, left)
		s2 := newSection(rightPrefix, right)
		sections := []*Section{}
		sections = append(sections, s1...)
		sections = append(sections, s2...)
		return sections
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
	for i, vault := range s.Vaults {
		if vault == v {
			s.Vaults = append(s.Vaults[:i], s.Vaults[i+1:]...)
			break
		}
	}
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
	for _, sv := range s.Vaults {
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
