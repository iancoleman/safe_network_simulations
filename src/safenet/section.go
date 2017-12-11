package safenet

import (
	"math"
	"math/big"
	"sort"
)

type Section struct {
	Prefix           Prefix
	TotalAdults      int
	Vaults           []*Vault // all vaults, including elders
	Elders           []*Vault
	LeftTotalAdults  int
	RightTotalAdults int
	IsAttacked       bool
}

// Returns a slice of sections since as vaults age they may cascade into
// multiple sections.
func newSection(prefix Prefix, vaults []*Vault) *NetworkEvent {
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
		if v.IsAdult() {
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
		return s.split()
	}
	// track elders but do not respond to network events since this new
	// section is already the result of that
	s.updateEldersWithoutNetworkEvent()
	// track attack
	s.checkIfAttacked()
	// return the section
	ne := NewNetworkEvent()
	ne.NewSections = []*Section{&s}
	return ne
}

func (s *Section) split() *NetworkEvent {
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
	ne0 := newSection(leftPrefix, left)
	ne1 := newSection(rightPrefix, right)
	ne := NewNetworkEvent()
	ne.NewSections = []*Section{}
	ne.NewSections = append(ne.NewSections, ne0.NewSections...)
	ne.NewSections = append(ne.NewSections, ne1.NewSections...)
	return ne
}

func (s *Section) shouldSplit() bool {
	return s.LeftTotalAdults >= SplitSize && s.RightTotalAdults >= SplitSize
}

func (s *Section) addVault(v *Vault) *NetworkEvent {
	v.SetPrefix(s.Prefix)
	s.Vaults = append(s.Vaults, v)
	// track hypothetical future section
	if v.IsAdult() {
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
		return s.split()
	}
	// track elders
	ne := s.updateElders()
	// track attack
	s.checkIfAttacked()
	// no split so return zero new sections
	return ne
}

func (s *Section) removeVault(v *Vault) *NetworkEvent {
	// remove from section
	for i, vault := range s.Vaults {
		if vault == v {
			s.Vaults = append(s.Vaults[:i], s.Vaults[i+1:]...)
			break
		}
	}
	// track hypothetical future section
	if v.IsAdult() {
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
	// track elders
	ne := s.updateElders()
	// track attack
	s.checkIfAttacked()
	// merge is handled by network using NetworkEvent ne
	return ne
}

func (s *Section) getElders() []*Vault {
	// get elders
	// see https://forum.safedev.org/t/data-chains-deeper-dive/1209
	// the GROUP_SIZE oldest peers in the section
	sort.Sort(ByAge(s.Vaults))
	// if there aren't enough vaults, return all of them
	if len(s.Vaults) <= GroupSize {
		return s.Vaults
	}
	newElders := s.Vaults[len(s.Vaults)-GroupSize:]
	// include any vaults of the same age as the youngest elder
	for i := len(s.Vaults) - GroupSize - 1; i >= 0; i-- {
		if s.Vaults[i].Age == newElders[0].Age {
			newElders = append([]*Vault{s.Vaults[i]}, newElders...)
		} else {
			break
		}
	}
	return newElders
}

func (s *Section) updateElders() *NetworkEvent {
	// get new elders
	newElders := s.getElders()
	// check if elders has changed
	// firstly based on number of elders being different
	// secondly based on membership of elders being different
	eldersHasChanged := len(newElders) != len(s.Elders)
	if !eldersHasChanged {
		for i, e := range newElders {
			if s.Elders[i] != e {
				eldersHasChanged = true
				break
			}
		}
	}
	// cache the elders for future comparison
	s.Elders = newElders
	// create new network event if needed
	// see https://forum.safedev.org/t/data-chains-deeper-dive/1209
	var ne *NetworkEvent
	if eldersHasChanged {
		// see if this event causes a vault relocation
		ne = NewNetworkEvent()
		v := s.vaultForRelocation(ne)
		if v != nil {
			// set vault for relocation
			ne.VaultToRelocate = v
		}
	}
	return ne
}

func (s *Section) updateEldersWithoutNetworkEvent() {
	// get new elders
	newElders := s.getElders()
	// cache the elders for future comparison
	s.Elders = newElders
}

func (s *Section) checkIfAttacked() {
	// check if enough elders to control quorum
	totalAttackingElders := 0
	for _, v := range s.Elders {
		if v.IsAttacker {
			totalAttackingElders = totalAttackingElders + 1
		}
	}
	// use integer arithmetic to check quorum
	// see https://github.com/maidsafe/routing/blob/da462bfebfd47dd16cb0c7523359d219bb097a3e/src/lib.rs#L213
	attackingVotes := totalAttackingElders
	voters := len(s.Elders)
	quorumAttacked := attackingVotes*QuorumDenominator > voters*QuorumNumerator
	s.IsAttacked = quorumAttacked
}

func (s *Section) GetRandomVault() *Vault {
	i := prng.Intn(len(s.Vaults))
	return s.Vaults[i]
}

func (s *Section) vaultForRelocation(ne *NetworkEvent) *Vault {
	// find vault to relocate based on a randomly generated 'event hash'
	// see https://forum.safedev.org/t/data-chains-deeper-dive/1209
	// As we receive/form a valid block of Live for non-infant peers, we take
	// the Hash of the event H. Then if H % 2^age == 0 for any peer (sorted by
	// age ascending) in our section, we relocate this node to the neighbour
	// that has the lowest number of peers.
	youngestAge := math.MaxUint32
	smallestTiebreaker := big.NewInt(0).SetBytes(largestHashValue)
	var v *Vault
	for _, w := range s.Vaults {
		if w.Age > youngestAge {
			continue
		} else if w.Age < youngestAge {
			// calculate divisor as 2^age
			divisor := big.NewInt(1)
			divisor.Lsh(divisor, uint(w.Age))
			if ne.HashModIsZero(divisor) {
				youngestAge = w.Age
				v = w
				// track xordistance for potenital future tiebreaker
				xordistance := big.NewInt(0)
				xordistance.Xor(w.Name.bigint, ne.hash)
				xordistance.Abs(xordistance)
				smallestTiebreaker = xordistance
			}
		} else if w.Age == youngestAge {
			// calculate divisor as 2^age
			divisor := big.NewInt(1)
			divisor.Lsh(divisor, uint(w.Age))
			if ne.HashModIsZero(divisor) {
				// tiebreaker
				// If there are multiple peers of the same age then XOR their
				// public keys together and find the one XOR closest to it.
				xordistance := big.NewInt(0)
				xordistance.Xor(w.Name.bigint, ne.hash)
				xordistance.Abs(xordistance)
				if xordistance.Cmp(smallestTiebreaker) == -1 {
					smallestTiebreaker = xordistance
					v = w
				}
			}
		}
	}
	return v
}
