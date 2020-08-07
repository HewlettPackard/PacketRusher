package nasConvert

import (
	"my5G-RANTester/lib/nas/nasType"
	"strings"
)

func LocalTimeZoneToNas(timezone string) (nasTimezone nasType.LocalTimeZone) {

	time := 0

	if timezone[0] == '-' {
		time = 64 //0x80
	}

	if timezone[1] == '1' {
		time += (10 * 4) // expressed in quarters of an hour
	}

	for i := 0; i < 10; i++ {
		if int(timezone[2]) == (i + 0x30) {
			time += i * 4
		}
	}

	for i := 1; i <= 4; i++ {
		if int(timezone[4]) == (i + 0x30) {
			time += i
		}
	}

	nasTimezone.SetTimeZone(uint8(time))
	return
}

func DaylightSavingTimeToNas(timezone string) (nasDaylightSavingTimeToNas nasType.NetworkDaylightSavingTime) {

	value := 0

	if strings.Contains(timezone, "+1h") {
		value = 1
	}

	if strings.Contains(timezone, "+2h") {
		value = 2
	}

	nasDaylightSavingTimeToNas.SetLen(1)
	nasDaylightSavingTimeToNas.Setvalue(uint8(value))
	return
}
