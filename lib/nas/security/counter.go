package security

/*
TS 33.501 6.4.3.1
 COUNT (32 bits) := 0x00 || NAS COUNT (24 bits)
 NAS COUNT (24 bits) := NAS OVERFLOW (16 bits) || NAS SQN (8 bits)
*/
type Count struct {
	count uint32
}

func (counter *Count) maskTo24Bits() {
	counter.count &= 0x00ffffff
}

func (counter *Count) Set(overflow uint16, sqn uint8) {
	counter.SetOverflow(overflow)
	counter.SetSQN(sqn)
}

func (counter *Count) Get() uint32 {
	return counter.count
}

func (counter *Count) AddOne() {
	counter.count++
	counter.maskTo24Bits()
}

func (counter *Count) SQN() uint8 {
	return uint8(counter.count & 0x000000ff)
}

func (counter *Count) SetSQN(sqn uint8) {
	counter.count = (counter.count & 0xffffff00) | uint32(sqn)
}

func (counter *Count) Overflow() uint16 {
	return uint16((counter.count & 0x00ffff00) >> 8)
}

func (counter *Count) SetOverflow(overflow uint16) {
	counter.count = (counter.count & 0xff0000ff) | (uint32(overflow) << 8)
}
