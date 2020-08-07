package nasConvert

func PSIToBooleanArray(buf []uint8) (array [16]bool) {
	if len(buf) < 2 {
		return
	}
	for i := uint8(0); i < 16; i++ {
		if (buf[i/8] & (1 << (i % 8))) > 0 {
			array[i] = true
		}
	}
	return
}

func PSIToBuf(array [16]bool) []uint8 {
	var buf [2]uint8
	for i := uint8(0); i < 16; i++ {
		if array[i] {
			buf[i/8] |= (1 << (i % 8))
		}
	}
	return buf[:]
}
