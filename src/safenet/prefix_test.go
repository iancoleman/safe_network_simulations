package safenet

import (
	"testing"
)

func TestNewBlankPrefix(t *testing.T) {
	p := NewBlankPrefix()
	if len(p.id) != 0 {
		t.Error("NewBlankPrefix length is not 0")
	}
}

func TestPrefixExtendLeft(t *testing.T) {
	p := NewBlankPrefix()
	p = p.extendLeft()
	if len(p.id) != 1 {
		t.Error("extendLeft once length")
	}
	if p.id[0] != false {
		t.Error("extendLeft once value")
	}
	p = p.extendLeft()
	if len(p.id) != 2 {
		t.Error("extendLeft twice length")
	}
	if p.id[0] != false && p.id[1] != false {
		t.Error("extendLeft twice value")
	}
}

func TestPrefixExtendRight(t *testing.T) {
	p := NewBlankPrefix()
	p = p.extendRight()
	if len(p.id) != 1 {
		t.Error("extendRight once length")
	}
	if p.id[0] != true {
		t.Error("extendRight once value")
	}
	p = p.extendRight()
	if len(p.id) != 2 {
		t.Error("extendRight twice length")
	}
	if p.id[0] != true && p.id[1] != true {
		t.Error("extendRight twice value")
	}
}

func TestPrefixExtendLeftAndRight(t *testing.T) {
	p := NewBlankPrefix()
	p = p.extendLeft()
	p = p.extendRight()
	if len(p.id) != 2 {
		t.Error("extend Left and Right length")
	}
	if p.id[0] != false && p.id[1] != true {
		t.Error("extend Left and Right values")
	}
}

func TestPrefixParent(t *testing.T) {
	p := NewBlankPrefix()
	p.id = []bool{false, false, false, true}
	p = p.parent()
	if len(p.id) != 3 {
		t.Error("parent length")
	}
	if p.id[2] != false {
		t.Error("parent value")
	}
}

func TestPrefixSibling(t *testing.T) {
	p := NewBlankPrefix()
	p.id = []bool{false, false, false, true}
	p = p.sibling()
	if len(p.id) != 4 {
		t.Error("sibling length")
	}
	if p.id[3] != false {
		t.Error("sibling value")
	}
}

func TestPrefixSetKey(t *testing.T) {
	p := NewBlankPrefix()
	p = p.extendLeft()
	q := NewBlankPrefix()
	q = q.extendLeft()
	q = q.extendLeft()
	if p.Key == q.Key {
		t.Error("setKey for 0 matches 00")
	}
	// key characters come directly from bytes
	q = q.extendLeft()
	q = q.extendLeft()
	q = q.extendLeft()
	q = q.extendLeft()
	q = q.extendLeft()
	q = q.extendLeft()
	if len(q.Key) != 2 {
		t.Error("setKey length for eight bits")
	}
	q = q.extendLeft()
	if len(q.Key) != 3 {
		t.Error("setKey length for nine bits")
	}
	q = q.extendLeft()
	if len(q.Key) != 4 {
		t.Error("setKey length for ten bits")
	}
}

func TestPrefixMatches(t *testing.T) {
	// 0000 0100 0000 0010
	x := XorName{4, 2}
	p := NewBlankPrefix()
	if !p.Matches(x) {
		t.Error("blank prefix match")
	}
	p = p.extendLeft()
	if !p.Matches(x) {
		t.Error("one bit prefix match")
	}
	p = p.extendLeft()
	if !p.Matches(x) {
		t.Error("two bits prefix match")
	}
	// make mismatch
	p = p.extendRight()
	if p.Matches(x) {
		t.Error("three bits prefix match")
	}
	// return to 2 bits prefix that matches
	p = p.parent()
	// test second bit
	p = p.extendLeft()
	p = p.extendLeft()
	p = p.extendLeft()
	p = p.extendRight()
	p = p.extendLeft()
	p = p.extendLeft()
	if len(p.id) != 8 {
		t.Error("eight bits match length")
	}
	if !p.Matches(x) {
		t.Error("eight bits prefix match")
	}
	p = p.extendLeft()
	if !p.Matches(x) {
		t.Error("nine bits prefix match")
	}
	// make second bit mismatch
	p = p.extendRight()
	if p.Matches(x) {
		t.Error("ten bits prefix match")
	}
	// return to nine bits matching
	p = p.parent()
	p = p.extendLeft()
	p = p.extendLeft()
	p = p.extendLeft()
	p = p.extendLeft()
	p = p.extendLeft()
	p = p.extendRight()
	p = p.extendLeft()
	if !p.Matches(x) {
		t.Error("sixteen bits prefix match")
	}
	p = p.extendLeft()
	if p.Matches(x) {
		t.Error("seventeen bits prefix match on sixteen bit name")
	}
}

func TestPrefixTotalBytes(t *testing.T) {
	p := NewBlankPrefix()
	if p.totalBytes() != 0 {
		t.Error("totalBytes for 0 bits")
	}
	p = p.extendLeft()
	if p.totalBytes() != 1 {
		t.Error("totalBytes for 1 bit")
	}
	p = p.extendLeft()
	if p.totalBytes() != 1 {
		t.Error("totalBytes for 2 bits")
	}
	p = p.extendLeft()
	p = p.extendLeft()
	p = p.extendLeft()
	p = p.extendLeft()
	p = p.extendLeft()
	p = p.extendLeft()
	if p.totalBytes() != 1 {
		t.Error("totalBytes for 8 bits")
	}
	p = p.extendLeft()
	if p.totalBytes() != 2 {
		t.Error("totalBytes for 9 bits")
	}
}
