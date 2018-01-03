package safenet

import (
	"fmt"
	"math/big"
	"sort"
)

type Section struct {
	Prefix Prefix
	Vaults []*Vault
}

// Returns a slice of sections since as vaults age they may cascade into
// multiple sections.
func newSection(prefix Prefix, vaults []*Vault) *NetworkEvent {
	s := Section{
		Prefix: prefix,
		Vaults: []*Vault{},
	}
	// add each existing vault to new section
	for _, v := range vaults {
		v.SetPrefix(s.Prefix)
		s.Vaults = append(s.Vaults, v)
	}
	// split into two sections if needed.
	// there is no vault relocation here.
	if s.shouldSplit() {
		return s.split()
	}
	// return the section as a network event.
	// there is a vault relocation here.
	ne := NewNetworkEvent()
	ne.NewSections = []*Section{&s}
	v := s.vaultForRelocation(ne)
	if v != nil {
		ne.VaultToRelocate = v
	}
	return ne
}

func (s *Section) addVault(v *Vault) (*NetworkEvent, bool) {
	// disallow more than one node aged 1 per section if the section is
	// complete (all elders are adults)
	// see https://github.com/fizyk20/ageing_sim/blob/53829350daa372731c9b8080488b2a75c72f60bb/src/network/section.rs#L198
	isDisallowed := false
	if v.Age == 1 && s.hasVaultAgedOne() && s.isComplete() {
		isDisallowed = true
		return nil, isDisallowed
	}
	v.SetPrefix(s.Prefix)
	s.Vaults = append(s.Vaults, v)
	// split into two sections if needed
	// details are handled by network upon returning two new sections
	if s.shouldSplit() {
		return s.split(), isDisallowed
	}
	// no split so return zero new sections
	// but a new vault added triggers a network event which may lead to vault
	// relocation
	ne := NewNetworkEvent()
	r := s.vaultForRelocation(ne)
	if r != nil {
		ne.VaultToRelocate = r
	}
	return ne, isDisallowed
}

func (s *Section) removeVault(v *Vault) *NetworkEvent {
	// remove from section
	for i, vault := range s.Vaults {
		if vault == v {
			s.Vaults = append(s.Vaults[:i], s.Vaults[i+1:]...)
			break
		}
	}
	// merge is handled by network using NetworkEvent ne
	// which includes a vault relocation
	ne := NewNetworkEvent()
	r := s.vaultForRelocation(ne)
	if r != nil {
		ne.VaultToRelocate = r
	}
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
	if s.isComplete() {
		left := s.leftAdultCount()
		right := s.rightAdultCount()
		return left >= SplitSize && right >= SplitSize
	} else {
		left := s.leftVaultCount()
		right := s.rightVaultCount()
		return left >= SplitSize && right >= SplitSize
	}
}

func (s *Section) isComplete() bool {
	// GROUP_SIZE peers with age >4 in a section
	return s.TotalAdults() >= GroupSize
}

func (s *Section) hasVaultAgedOne() bool {
	for _, v := range s.Vaults {
		if v.Age == 1 {
			return true
		}
	}
	return false
}

func (s *Section) elders() []*Vault {
	// get elders
	// see https://forum.safedev.org/t/data-chains-deeper-dive/1209
	// the GROUP_SIZE oldest peers in the section
	// tiebreakers are handled by the sort algorithm
	sort.Sort(oldestFirst(s.Vaults))
	// if there aren't enough vaults, use all of them
	elders := s.Vaults
	// otherwise get the GroupSize oldest vaults
	if len(s.Vaults) > GroupSize {
		elders = s.Vaults[:GroupSize]
	}
	return elders
}

func (s *Section) IsAttacked() bool {
	// check if enough attacking elders to control quorum
	// and if attackers control 50% of the age
	// see https://github.com/maidsafe/rfcs/blob/master/text/0045-node-ageing/0045-node-ageing.md#consensus-measurement
	// A group consensus will require >50% of nodes and >50% of the age of the whole group.
	elders := s.elders()
	totalVotes := len(elders)
	totalAge := 0
	attackingVotes := 0
	attackingAge := 0
	for _, v := range elders {
		totalAge = totalAge + v.Age
		if v.IsAttacker {
			attackingVotes = attackingVotes + 1
			attackingAge = attackingAge + v.Age
		}
	}
	// use integer arithmetic to check quorum
	// see https://github.com/maidsafe/routing/blob/da462bfebfd47dd16cb0c7523359d219bb097a3e/src/lib.rs#L213
	votesAttacked := attackingVotes*QuorumDenominator > totalVotes*QuorumNumerator
	// compare ages
	ageAttacked := attackingAge*QuorumDenominator > totalAge*QuorumNumerator
	return votesAttacked && ageAttacked
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
	oldestAge := 0
	smallestTiebreaker := big.NewInt(0).SetBytes(largestHashValue)
	var v *Vault
	for _, w := range s.Vaults {
		if w.Age < oldestAge {
			continue
		} else if w.Age > oldestAge {
			// calculate divisor as 2^age
			divisor := big.NewInt(1)
			divisor.Lsh(divisor, uint(w.Age))
			if ne.HashModIsZero(divisor) {
				oldestAge = w.Age
				v = w
				// track xordistance for potential future tiebreaker
				xordistance := big.NewInt(0)
				xordistance.Xor(w.Name.bigint, ne.hash)
				smallestTiebreaker = xordistance
			}
		} else if w.Age == oldestAge {
			// calculate divisor as 2^age
			divisor := big.NewInt(1)
			divisor.Lsh(divisor, uint(w.Age))
			if ne.HashModIsZero(divisor) {
				// tiebreaker
				// If there are multiple peers of the same age then XOR their
				// public keys together and find the one XOR closest to it.
				// TODO this isn't done correctly, since it only XORs the two
				// keys when it should XOR all keys of this age.
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
	return s.TotalAdults() <= GroupSize
}

func (s *Section) TotalAdults() int {
	adults := 0
	for _, v := range s.Vaults {
		if v.IsAdult() {
			adults = adults + 1
		}
	}
	return adults
}

func (s *Section) TotalElders() int {
	return len(s.elders())
}

func (s *Section) leftVaultCount() int {
	leftPrefix := s.Prefix.extendLeft()
	return s.vaultCountForExtendedPrefix(leftPrefix)
}

func (s *Section) rightVaultCount() int {
	rightPrefix := s.Prefix.extendRight()
	return s.vaultCountForExtendedPrefix(rightPrefix)
}

func (s *Section) vaultCountForExtendedPrefix(p Prefix) int {
	vaults := 0
	for _, v := range s.Vaults {
		if p.Matches(v.Name) {
			vaults = vaults + 1
		}
	}
	return vaults
}

func (s *Section) leftAdultCount() int {
	leftPrefix := s.Prefix.extendLeft()
	return s.adultCountForExtendedPrefix(leftPrefix)
}

func (s *Section) rightAdultCount() int {
	rightPrefix := s.Prefix.extendRight()
	return s.adultCountForExtendedPrefix(rightPrefix)
}

func (s *Section) adultCountForExtendedPrefix(p Prefix) int {
	adults := 0
	for _, v := range s.Vaults {
		if v.IsAdult() && p.Matches(v.Name) {
			adults = adults + 1
		}
	}
	return adults
}
