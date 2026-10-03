package ie

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// UniversalTimeAndLocalTimeZone 9.11.3.53
// Year Row, sBit, len = [0, 0], 8 , 8
// Month Row, sBit, len = [1, 1], 8 , 8
// Day Row, sBit, len = [2, 2], 8 , 8
// Hour Row, sBit, len = [3, 3], 8 , 8
// Minute Row, sBit, len = [4, 4], 8 , 8
// Second Row, sBit, len = [5, 5], 8 , 8
// TimeZone Row, sBit, len = [6, 6], 8 , 8

type TimeZoneAndTime struct {
	Time time.Time
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *TimeZoneAndTime) UnmarshalBinary(b []byte) error {
	if len(b) != 7 {
		return errors.Errorf("TimeZoneAndTime UnmarshalBinary: len should be 7")
	}
	year := 2000 + int(fromBinaryCodedDecimal(toSemiOctet(b[0])))
	month := fromBinaryCodedDecimal(toSemiOctet(b[1]))
	day := fromBinaryCodedDecimal(toSemiOctet(b[2]))
	hour := fromBinaryCodedDecimal(toSemiOctet(b[3]))
	minute := fromBinaryCodedDecimal(toSemiOctet(b[4]))
	second := fromBinaryCodedDecimal(toSemiOctet(b[5]))
	timezone := &TimeZone{}
	err := timezone.UnmarshalBinary([]byte{b[6]})
	if err != nil {
		return errors.Wrap(err, "TimeZoneAndTime UnmarshalBinary()")
	}
	offset, err := timezone.Offset()
	if err != nil {
		return errors.Wrap(err, "TimeZoneAndTime UnmarshalBinary()")
	}

	location := time.FixedZone("", offset)
	i.Time = time.Date(
		year, time.Month(month), int(day), int(hour), int(minute), int(second), 0, location)
	return nil
}

// Get TimeZone from time.Time structure
func (i *TimeZoneAndTime) GetTimeZone() *TimeZone {
	t := i.Time
	timezone := ""
	_, offset := t.Zone() // diff to UTC in second
	if t.IsDST() {
		// Adjust one hour to get the orignal time
		offset -= 3600
	}
	if offset < 0 {
		timezone += "-"
		offset = 0 - offset
	} else {
		timezone += "+"
	}
	timezone += fmt.Sprintf("%02d:%02d", offset/3600, (offset%3600)/60)
	if t.IsDST() {
		timezone += "+1"
	}

	return &TimeZone{
		Value: timezone,
	}
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *TimeZoneAndTime) MarshalBinary() ([]byte, error) {
	universalTime := i.Time
	year := toSemiOctet(toBinaryCodedDecimal(uint8(universalTime.Year() % 100)))
	month := toSemiOctet(toBinaryCodedDecimal(uint8(universalTime.Month())))
	day := toSemiOctet(toBinaryCodedDecimal(uint8(universalTime.Day())))
	hour := toSemiOctet(toBinaryCodedDecimal(uint8(universalTime.Hour())))
	minute := toSemiOctet(toBinaryCodedDecimal(uint8(universalTime.Minute())))
	second := toSemiOctet(toBinaryCodedDecimal(uint8(universalTime.Second())))
	timezone, err := i.GetTimeZone().MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "TimeZoneAndTime MarshalBinary()")
	}
	if len(timezone) != 1 {
		return nil, errors.Errorf("TimeZoneAndTime MarshalBinary(): TimeZone should be 1 byte")
	}

	return []byte{
		year, month, day, hour, minute, second, timezone[0],
	}, nil
}
