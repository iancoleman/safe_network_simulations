package safenet

const coinNameByteLen = 4 // 4 bytes * 8 bits per byte = 32 bits = 2^32 coins

func randomSafecoinAddress() string {
	coinname := make([]byte, coinNameByteLen)
	prng.Read(coinname)
	return string(coinname)
}
