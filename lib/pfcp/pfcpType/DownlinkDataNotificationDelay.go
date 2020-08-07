package pfcpType

import (
	"fmt"
)

type DownlinkDataNotificationDelay struct {
	DelayValue uint8
}

func (d *DownlinkDataNotificationDelay) MarshalBinary() (data []byte, err error) {
	// Octet 5
	data = append([]byte(""), byte(d.DelayValue))

	return data, nil
}

func (d *DownlinkDataNotificationDelay) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0
	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}
	d.DelayValue = uint8(data[idx])
	idx = idx + 1

	if length != idx {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	return nil
}
