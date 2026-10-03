package ie

import "github.com/pkg/errors"

type GPRSTimerUnitType uint8

const (
	TimerIncIn_2Secs     GPRSTimerUnitType = 0
	TimerIncIn_1Min      GPRSTimerUnitType = 1
	TimerIncIn_Decihours GPRSTimerUnitType = 2
	Timer_Deactivated    GPRSTimerUnitType = 7
)

// GPRSTimer is detailed in 10.5.7.3 GPRS Timer, 24.008
type GPRSTimer struct {
	Unit       GPRSTimerUnitType // 8 -> 6 ,   2 -> 2
	TimerValue uint8             // 5 -> 1 ,   2 -> 2
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *GPRSTimer) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("Bad GPRSTimer IE length(%d)", len(b))
	}
	i.Unit = GPRSTimerUnitType(Get3Bits86(b[0]))
	i.TimerValue = Get5Bits51(b[0])
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *GPRSTimer) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = Set3Bits86(b[0], uint8(i.Unit))
	b[0] = Set5Bits51(b[0], i.TimerValue)
	return b, nil
}

func GPRSTmr1ToSec(b uint8) uint32 {
	unit := Get3Bits86(b)
	value := Get5Bits51(b)
	switch GPRSTimerUnitType(unit) {
	case TimerIncIn_2Secs:
		return uint32(value) * 2
	case TimerIncIn_Decihours:
		return uint32(value) * 360
	case Timer_Deactivated:
		return 0
	}
	// Other values shall be interpreted as multiples of
	// 1 minute in this version of the protocol.
	return uint32(value) * 60
}

func SecToGPRSTmr1(sec uint32) uint8 {
	if sec == 0 {
		return uint8(Timer_Deactivated) << 5
	}
	if sec < 0x20*2 {
		return uint8(sec / 2)
	}
	if sec < 0x20*60 {
		return uint8(TimerIncIn_1Min)<<5 + uint8(sec/60)
	}
	if sec < 0x20*360 {
		return uint8(TimerIncIn_Decihours)<<5 + uint8(sec/360)
	}
	return uint8(TimerIncIn_Decihours)<<5 + 0x1F
}
