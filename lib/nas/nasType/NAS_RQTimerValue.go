package nasType

// RQTimerValue 9.11.2.3
// Unit Row, sBit, len = [0, 0], 8 , 3
// TimerValue Row, sBit, len = [0, 0], 5 , 5
type RQTimerValue struct {
	Iei   uint8
	Octet uint8
}

func NewRQTimerValue(iei uint8) (rQTimerValue *RQTimerValue) {
	rQTimerValue = &RQTimerValue{}
	rQTimerValue.SetIei(iei)
	return rQTimerValue
}

// RQTimerValue 9.11.2.3
// Iei Row, sBit, len = [], 8, 8
func (a *RQTimerValue) GetIei() (iei uint8) {
	return a.Iei
}

// RQTimerValue 9.11.2.3
// Iei Row, sBit, len = [], 8, 8
func (a *RQTimerValue) SetIei(iei uint8) {
	a.Iei = iei
}

// RQTimerValue 9.11.2.3
// Unit Row, sBit, len = [0, 0], 8 , 3
func (a *RQTimerValue) GetUnit() (unit uint8) {
	return a.Octet & GetBitMask(8, 5) >> (5)
}

// RQTimerValue 9.11.2.3
// Unit Row, sBit, len = [0, 0], 8 , 3
func (a *RQTimerValue) SetUnit(unit uint8) {
	a.Octet = (a.Octet & 31) + ((unit & 7) << 5)
}

// RQTimerValue 9.11.2.3
// TimerValue Row, sBit, len = [0, 0], 5 , 5
func (a *RQTimerValue) GetTimerValue() (timerValue uint8) {
	return a.Octet & GetBitMask(5, 0)
}

// RQTimerValue 9.11.2.3
// TimerValue Row, sBit, len = [0, 0], 5 , 5
func (a *RQTimerValue) SetTimerValue(timerValue uint8) {
	a.Octet = (a.Octet & 224) + (timerValue & 31)
}
