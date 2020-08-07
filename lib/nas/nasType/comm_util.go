package nasType

// GetBitMask number, pos is shift bit
// >= lb
// < up
// TODO　exception check
func GetBitMask(ub uint8, lb uint8) (bitMask uint8) {
	// fmt.Println("%x", number)
	// fmt.Println("%x", 1<<number)
	bitMask = ((1<<(ub-lb) - 1) << (lb))
	return bitMask
}
