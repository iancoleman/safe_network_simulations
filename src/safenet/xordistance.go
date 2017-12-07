package safenet

type XorDistance []byte

func (x XorDistance) IsLessThan(y XorDistance) bool {
	if len(x) < len(y) {
		return true
	} else if len(x) > len(y) {
		return false
	}
	for i := 0; i < len(x); i++ {
		if x[i] < y[i] {
			return true
		} else if x[i] == y[i] {
			continue
		} else {
			return false
		}
	}
	return false
}

func (x XorDistance) IsZeroValue() bool {
	return len(x) == 0
}
