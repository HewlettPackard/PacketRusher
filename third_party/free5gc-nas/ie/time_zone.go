package ie

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
)

// "[+-]HH:MM[+][1-2]", Refer to TS 29.571 - 5.2.2 Simple Data Types
// e.g. "-08:00+1": 8hr behind UTC, +1hr adjustment for Daylight Saving Time
type TimeZone struct {
	Value string
}

func (i *TimeZone) String() string {
	return i.Value
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *TimeZone) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("TimeZone bad IE len(%d)", len(b))
	}
	val := b[0]

	// reverse semi octet
	val = toSemiOctet(val)

	// get signed bit
	signed := "+"
	signedBit := val >> 7
	if signedBit == 1 {
		signed = "-"
	}

	// removed signed bit and get value from BCD
	val &= 0x7f
	numQuarters := fromBinaryCodedDecimal(val)

	hr := numQuarters / 4
	if hr > 12 {
		return errors.Errorf("TimeZone UnmarshalBinary(): invalid value[%d] for hr", hr)
	}
	minute := numQuarters % 4 * 15

	// fill i.Value with string
	i.Value = fmt.Sprintf("%s%02d:%02d", signed, hr, minute)
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *TimeZone) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	re := regexp.MustCompile(`(?m)^([+-])([0-9]{2}):([0-9]{2})(\+[12])?$`)
	valid := re.MatchString(i.String())
	if !valid {
		return nil, errors.Errorf("TimeZone MarshalBinary(): invalid format[%s]", i.String())
	}

	matches := re.FindStringSubmatch(i.String())
	// e.g. "+12:20+2" -> ["+12:20+2" "+" "12" "20" "+2"]
	// e.g. "+12:20" -> ["+12:20" "+" "12" "20"]
	if len(matches) < 4 {
		return nil, errors.Errorf("TimeZone MarshalBinary(): unexpected matches result[%v]", matches)
	}

	// Parse hour
	h, err := strconv.ParseUint(matches[2], 10, 8)
	if err != nil {
		return nil, errors.Wrap(err, "TimeZone MarshalBinary(): Parse Hour")
	}
	if h > 12 {
		return nil, errors.Errorf("TimeZone MarshalBinary(): invalid format[%s] (hr: 0~12)", i.String())
	}

	// Parse minute
	m, err := strconv.ParseUint(matches[3], 10, 8)
	if err != nil {
		return nil, errors.Wrap(err, "TimeZone MarshalBinary(): Parse Minute")
	}
	if m > 59 {
		return nil, errors.Errorf("TimeZone MarshalBinary(): invalid format[%s] (min: 0~59)", i.String())
	}

	offsetInQuarter := int(h*4 + m/15)
	if matches[1] == "-" {
		offsetInQuarter *= -1 // decide negative/positive value for DST addition
	}

	// Daylight saving time is provided
	if len(matches) == 5 {
		switch matches[4] {
		case "+1":
			offsetInQuarter += 4 // 1 (hr) * 4 (quarters)
		case "+2":
			offsetInQuarter += 8 // 2 (hr) * 4 (quarters)
		}
	}

	isNegativeTz := false
	resultVal := uint8(0) // expressed in quarters of an hour
	if offsetInQuarter < 0 {
		// decide isNegativeTz and recover offsetInQuarter to positive value (uint8) for BCD
		isNegativeTz = true
		offsetInQuarter *= -1
	}

	// Convert decimal to binary-coded decimal
	resultVal = toBinaryCodedDecimal(uint8(offsetInQuarter))

	// Add signed number
	if isNegativeTz {
		resultVal |= 0x80
	}

	// Reform to SemiOctet
	b[0] = toSemiOctet(resultVal)

	return b, nil
}

// Offset returns the time offset of the timezone in seconds, which is used by time.FixedZone().
func (i *TimeZone) Offset() (int, error) {
	re := regexp.MustCompile(`(?m)^([+-])([0-9]{2}):([0-9]{2})(\+[12])?$`)
	valid := re.MatchString(i.String())
	if !valid {
		return 0, errors.Errorf("TimeZone Offset(): invalid format[%s]", i.String())
	}

	matches := re.FindStringSubmatch(i.String())
	// e.g. "+12:20+2" -> ["+12:20+2" "+" "12" "20" "+2"]
	// e.g. "+12:20" -> ["+12:20" "+" "12" "20"]
	if len(matches) < 4 {
		return 0, errors.Errorf("TimeZone Offset(): unexpected matches result[%v]", matches)
	}

	// Parse hour
	h, err := strconv.ParseUint(matches[2], 10, 8)
	if err != nil {
		return 0, errors.Wrap(err, "TimeZone Offset(): Parse Hour")
	}
	if h > 12 {
		return 0, errors.Errorf("TimeZone Offset(): invalid format[%s] (hr: 0~12)", i.String())
	}

	// Parse minute
	m, err := strconv.ParseUint(matches[3], 10, 8)
	if err != nil {
		return 0, errors.Wrap(err, "TimeZone Offset(): Parse Minute")
	}
	if m > 59 {
		return 0, errors.Errorf("TimeZone Offset(): invalid format[%s] (min: 0~59)", i.String())
	}

	offsetInSeconds := int(h*60*60 + m*60)
	if matches[1] == "-" {
		offsetInSeconds *= -1 // decide negative/positive value for DST addition
	}

	// Daylight saving time is provided
	if len(matches) == 5 {
		switch matches[4] {
		case "+1":
			offsetInSeconds += 60 * 60 // 1 hr
		case "+2":
			offsetInSeconds += 2 * 60 * 60 // 2 hr
		}
	}

	return offsetInSeconds, nil
}
