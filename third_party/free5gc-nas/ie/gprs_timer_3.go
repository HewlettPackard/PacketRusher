package ie

import "github.com/pkg/errors"

type GPRSTimer3 struct {
	Unit  GPRSTimer3UnitType // 8 -> 6 ,   1 -> 1
	Value uint8              // 5 -> 1 ,   1 -> 1
}

type GPRSTimer3UnitType uint8

const (
	TimerIncIn_10Minutes GPRSTimer3UnitType = 0x00
	TimerIncIn_1Hour     GPRSTimer3UnitType = 0x01
	TimerIncIn_10Hours   GPRSTimer3UnitType = 0x02
	TimerIncIn_2Seconds  GPRSTimer3UnitType = 0x03
	TimerIncIn_30Seconds GPRSTimer3UnitType = 0x04
	TimerIncIn_1Minute   GPRSTimer3UnitType = 0x05
	TimerIncIn_320Hours  GPRSTimer3UnitType = 0x06
	TimerDeactivated     GPRSTimer3UnitType = 0x07
)

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *GPRSTimer3) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("The GPRSTimer3  IE length(%d) is incorrect", len(b))
	}
	i.Unit = GPRSTimer3UnitType(Get3Bits86(b[0]))
	i.Value = Get5Bits51(b[0])

	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *GPRSTimer3) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = Set3Bits86(b[0], uint8(i.Unit))
	b[0] = Set5Bits51(b[0], i.Value)

	return b, nil
}

func (i *GPRSTimer3) Set(sec uint32) {
	if sec == 0 {
		i.Value, i.Unit = 0, GPRSTimer3UnitType(Timer_Deactivated)
		return
	}
	var secUnit uint32 = 2
	if sec < 0x20*secUnit {
		i.Value, i.Unit = uint8(sec/secUnit), TimerIncIn_2Seconds
		return
	}
	secUnit = 30
	if sec < 0x20*secUnit {
		i.Value, i.Unit = uint8(sec/secUnit), TimerIncIn_30Seconds
		return
	}
	secUnit = 60
	if sec < 0x20*secUnit {
		i.Value, i.Unit = uint8(sec/secUnit), TimerIncIn_1Minute
		return
	}
	secUnit = 600
	if sec < 0x20*secUnit {
		i.Value, i.Unit = uint8(sec/secUnit), TimerIncIn_10Minutes
		return
	}
	secUnit = 3600
	if sec < 0x20*secUnit {
		i.Value, i.Unit = uint8(sec/secUnit), TimerIncIn_1Hour
		return
	}
	secUnit = 3600 * 10
	if sec < 0x20*secUnit {
		i.Value, i.Unit = uint8(sec/secUnit), TimerIncIn_10Hours
		return
	}
	secUnit = 3600 * 320
	if sec < 0x20*secUnit {
		i.Value, i.Unit = uint8(sec/secUnit), TimerIncIn_320Hours
		return
	}
	i.Value, i.Unit = Get5Bits51(0xff), TimerIncIn_320Hours
}
