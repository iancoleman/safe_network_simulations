package safenet

import (
	"strconv"
)

type Prefix struct {
	id  []bool
	Key string
}

func NewBlankPrefix() Prefix {
	p := Prefix{
		id: []bool{},
	}
	p.setKey()
	return p
}

func (p Prefix) extendLeft() Prefix {
	newId := make([]bool, len(p.id)+1)
	for i, _ := range p.id {
		newId[i] = p.id[i]
	}
	l := Prefix{
		id: newId,
	}
	l.setKey()
	return l
}

func (p Prefix) extendRight() Prefix {
	newId := make([]bool, len(p.id)+1)
	for i, _ := range p.id {
		newId[i] = p.id[i]
	}
	newId[len(newId)-1] = true
	r := Prefix{
		id: newId,
	}
	r.setKey()
	return r
}

func (p Prefix) sibling() Prefix {
	s := Prefix{
		id: p.id,
	}
	s.id[len(s.id)-1] = !s.id[len(s.id)-1]
	s.setKey()
	return s
}

func (p Prefix) parent() Prefix {
	a := Prefix{
		id: p.id[:len(p.id)-1],
	}
	a.setKey()
	return a
}

func (p *Prefix) totalBytes() int {
	// int division does floor automatically
	totalBytes := len(p.id) / 8
	// but totalBytes should be ceil so do that here
	if len(p.id) > 0 && len(p.id)%8 != 0 {
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
			if j < len(p.id) && p.id[j] {
				thisByte = thisByte + 1
			}
		}
		bytes[i] = thisByte
	}
	p.Key = strconv.Itoa(len(p.id)) + string(bytes)
}

func (p Prefix) BinaryString() string {
	pb := ""
	for _, b := range p.id {
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
	totalBytes := p.totalBytes()
	if totalBytes > xornameBytes {
		return false
	}
	for i := 0; i < totalBytes; i++ {
		xornameByte := x.ByteAtIndex(i)
		startBit := i * 8
		endBit := (i + 1) * 8
		var thisByte byte
		for j := startBit; j < endBit; j++ {
			if j < len(p.id) {
				thisByte = thisByte << 1
				if p.id[j] {
					thisByte = thisByte + 1
				}
			} else {
				xornameByte = xornameByte >> 1
			}
		}
		if thisByte != xornameByte {
			return false
		}
	}
	return true
}
