package safenet

import (
	"fmt"
	"math"
	"math/big"
	"sort"
)

type Section struct {
	Prefix Prefix
	Vaults []*Vault // all vaults, including elders
	Elders []*Vault
}

// Returns a slice of sections since as vaults age they may cascade into
// multiple sections.
func newSection(prefix Prefix, vaults []*Vault) *NetworkEvent {
	s := Section{
		Prefix: prefix,
		Vaults: []*Vault{},
		Elders: []*Vault{},
	}
	// add each existing vault to new section
	for _, v := range vaults {
		// increment the age
		v.IncrementAge()
		// add to section
		v.SetPrefix(s.Prefix)
		s.Vaults = append(s.Vaults, v)
	}
	// track elders
	ne := s.updateElders()
	// split into two sections if needed
	if s.shouldSplit() {
		return s.split()
	}
	// return the section as a network event
	if ne == nil {
		ne = NewNetworkEvent()
	}
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
		} else if rightPrefix.Matches(v.Name) {
			right = append(right, v)
		} else {
			fmt.Println("Warning: Split has vault that doesn't match extended prefix")
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
	// use adults if there are enough adults to split
	if s.LeftTotalAdults() >= SplitSize && s.RightTotalAdults() >= SplitSize {
		return true
	}
	// all elders count as adults, which may help split young networks with
	// infant elders.
	if s.LeftTotalElders() >= SplitSize && s.RightTotalElders() >= SplitSize {
		return true
	}
	return false
}

func (s *Section) IsComplete() bool {
	return s.TotalAdults() >= GroupSize
}

func (s *Section) addVault(v *Vault) *NetworkEvent {
	// disallow more than one node aged 1 per section if the section is complete
	// (all elders are adults)
	if v.Age == 1 && s.hasVaultAgedOne() && s.IsComplete() {
		// TODO decide if vault relocation is appropriate here
		ne := NewNetworkEvent()
		ne.VaultToRelocate = s.vaultForRelocation(ne)
		return ne
	}
	v.SetPrefix(s.Prefix)
	s.Vaults = append(s.Vaults, v)
	// track elders
	ne := s.updateElders()
	// split into two sections if needed
	// details are handled by network upon returning two new sections
	if s.shouldSplit() {
		return s.split()
	}
	// no split so return zero new sections
	return ne
}

func (s *Section) hasVaultAgedOne() bool {
	for _, v := range s.Vaults {
		if v.Age == 1 {
			return true
		}
	}
	return false
}

func (s *Section) removeVault(v *Vault) *NetworkEvent {
	// remove from section
	for i, vault := range s.Vaults {
		if vault == v {
			s.Vaults = append(s.Vaults[:i], s.Vaults[i+1:]...)
			break
		}
	}
	// track elders
	ne := s.updateElders()
	// merge is handled by network using NetworkEvent ne
	return ne
}

func (s *Section) updateElders() *NetworkEvent {
	// get elders
	// see https://forum.safedev.org/t/data-chains-deeper-dive/1209
	// the GROUP_SIZE oldest peers in the section
	sort.Sort(ByAge(s.Vaults))
	// if there aren't enough vaults, use all of them
	newElders := s.Vaults
	// otherwise get the GroupSize oldest vaults
	if len(s.Vaults) > GroupSize {
		newElders = s.Vaults[len(s.Vaults)-GroupSize:]
	}
	// include any vaults of the same age as the youngest elder
	for i := len(s.Vaults) - GroupSize - 1; i >= 0; i-- {
		if s.Vaults[i].Age == newElders[0].Age {
			newElders = append([]*Vault{s.Vaults[i]}, newElders...)
		} else {
			break
		}
	}
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
	blockIsCreated := eldersHasChanged
	var ne *NetworkEvent
	if blockIsCreated {
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

func (s *Section) IsAttacked() bool {
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
	return quorumAttacked
}

func (s *Section) GetRandomVault() *Vault {
	totalVaults := len(s.Vaults)
	if totalVaults == 0 {
		fmt.Println("Warning: GetRandomVault for section with no vaults")
		return nil
	}
	i := prng.Intn(totalVaults)
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
				// track xordistance for potential future tiebreaker
				xordistance := big.NewInt(0)
				xordistance.Xor(w.Name.bigint, ne.hash)
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
				if xordistance.Cmp(smallestTiebreaker) == -1 {
					smallestTiebreaker = xordistance
					v = w
				}
			}
		}
	}
	return v
}

func (s *Section) shouldMerge() bool {
	return s.TotalAdults() < GroupSize
}

func (s *Section) TotalAdults() int {
	adults := 0
	for _, v := range s.Vaults {
		if v.IsAdult() {
			adults = adults + 1
		}
	}
	// all elders count as adults, which may help split young networks with
	// infant elders.
	elders := s.TotalElders()
	if elders > adults {
		adults = elders
	}
	return adults
}

func (s *Section) TotalElders() int {
	return len(s.Elders)
}

func (s *Section) LeftTotalAdults() int {
	adults := 0
	leftPrefix := s.Prefix.extendLeft()
	for _, v := range s.Vaults {
		if v.IsAdult() && leftPrefix.Matches(v.Name) {
			adults = adults + 1
		}
	}
	return adults
}

func (s *Section) RightTotalAdults() int {
	adults := 0
	rightPrefix := s.Prefix.extendRight()
	for _, v := range s.Vaults {
		if v.IsAdult() && rightPrefix.Matches(v.Name) {
			adults = adults + 1
		}
	}
	return adults
}

func (s *Section) LeftTotalElders() int {
	elders := 0
	leftPrefix := s.Prefix.extendLeft()
	for _, v := range s.Elders {
		if leftPrefix.Matches(v.Name) {
			elders = elders + 1
		}
	}
	return elders
}

func (s *Section) RightTotalElders() int {
	elders := 0
	rightPrefix := s.Prefix.extendRight()
	for _, v := range s.Elders {
		if rightPrefix.Matches(v.Name) {
			elders = elders + 1
		}
	}
	return elders
}
