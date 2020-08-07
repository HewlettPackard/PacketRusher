package pfcpType

import (
	"encoding/binary"
	"fmt"
	"time"
)

type RecoveryTimeStamp struct {
	RecoveryTimeStamp time.Time
}

func (r *RecoveryTimeStamp) MarshalBinary() (data []byte, err error) {
	// Octet 5 to 8
	duration := r.RecoveryTimeStamp.Sub(BASE_DATE_NTP_ERA0).Seconds()
	data = make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(duration))

	return data, nil
}

func (r *RecoveryTimeStamp) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5 to 8
	if length < idx+4 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	duration := binary.BigEndian.Uint32(data[idx:])
	r.RecoveryTimeStamp = BASE_DATE_NTP_ERA0.Add(time.Duration(duration) * time.Second)
	idx = idx + 4

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
