package safenet

import (
	"strconv"
)

type Prefix struct {
	bits []bool
	Key  string
}

func NewBlankPrefix() Prefix {
	p := Prefix{
		bits: []bool{},
	}
	p.setKey()
	return p
}

func (p Prefix) extendLeft() Prefix {
	newBits := make([]bool, len(p.bits)+1)
	for i, _ := range p.bits {
		newBits[i] = p.bits[i]
	}
	l := Prefix{
		bits: newBits,
	}
	l.setKey()
	return l
}

func (p Prefix) extendRight() Prefix {
	newBits := make([]bool, len(p.bits)+1)
	for i, _ := range p.bits {
		newBits[i] = p.bits[i]
	}
	newBits[len(newBits)-1] = true
	r := Prefix{
		bits: newBits,
	}
	r.setKey()
	return r
}

func (p Prefix) sibling() Prefix {
	s := Prefix{
		bits: p.bits,
	}
	s.bits[len(s.bits)-1] = !s.bits[len(s.bits)-1]
	s.setKey()
	return s
}

func (p Prefix) parent() Prefix {
	a := Prefix{
		bits: p.bits[:len(p.bits)-1],
	}
	a.setKey()
	return a
}

func (p *Prefix) totalBytes() int {
	// int division does floor automatically
	totalBytes := len(p.bits) / 8
	// but totalBytes should be ceil so do that here
	if len(p.bits) > 0 && len(p.bits)%8 != 0 {
		totalBytes = totalBytes + 1
	}
	return totalBytes
}

func (p *Prefix) setKey() {
	totalBytes := p.totalBytes()
	// preallocate bytes to avoid append
	bytes := make([]byte, totalBytes)
	for i := 0; i < totalBytes; i++ {
		var thisByte byte
		startBit := i * 8
		endBit := (i + 1) * 8
		for j := startBit; j < endBit; j++ {
			thisByte = thisByte << 1
			if j < len(p.bits) && p.bits[j] {
				thisByte = thisByte + 1
			}
		}
		bytes[i] = thisByte
	}
	p.Key = strconv.Itoa(len(p.bits)) + string(bytes)
}

func (p Prefix) BinaryString() string {
	pb := ""
	for _, b := range p.bits {
		if b {
			pb = pb + "1"
		} else {
			pb = pb + "0"
		}
	}
	return pb
}

func (p Prefix) Equals(q Prefix) bool {
	return p.Key == q.Key
}

func (p Prefix) Matches(x XorName) bool {
	if len(p.bits) > len(x.bits) {
		return false
	}
	for i := 0; i < len(p.bits); i++ {
		if p.bits[i] != x.bits[i] {
			return false
		}
	}
	return true
}
