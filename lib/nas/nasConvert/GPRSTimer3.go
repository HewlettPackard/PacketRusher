package nasConvert

import (
	"my5G-RANTester/lib/nas/nasMessage"
)

// TS 24.008 10.5.7.4a
func GPRSTimer3ToNas(timerValue int) (timerValueNas uint8) {

	if timerValue <= 2*31 {
		t := uint8(timerValue / 2)
		timerValueNas = (nasMessage.GPRSTimer3UnitMultiplesOf2Seconds << 5) + t
	} else if timerValue <= 30*31 {
		t := uint8(timerValue / 30)
		timerValueNas = (nasMessage.GPRSTimer3UnitMultiplesOf30Seconds << 5) + t
	} else if timerValue <= 60*31 {
		t := uint8(timerValue / 60)
		timerValueNas = (nasMessage.GPRSTimer3UnitMultiplesOf1Minute << 5) + t
	} else if timerValue <= 600*31 {
		t := uint8(timerValue / 600)
		timerValueNas = (nasMessage.GPRSTimer3UnitMultiplesOf10Minutes << 5) + t
	} else if timerValue <= 3600*31 {
		t := uint8(timerValue / 3600)
		timerValueNas = (nasMessage.GPRSTimer3UnitMultiplesOf1Hour << 5) + t
	} else {
		t := uint8(timerValue / (36000))
		timerValueNas = (nasMessage.GPRSTimer3UnitMultiplesOf10Hours << 5) + t
	}

	return
}
