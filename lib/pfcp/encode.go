package pfcp

import (
	"fmt"
	"free5gc/lib/tlv"
)

func (m *Message) Marshal() (data []byte, err error) {
	var headerBuf []byte
	var bodyBuf []byte

	bodyBuf, err = tlv.Marshal(m.Body)
	if err != nil {
		fmt.Println(err)
	}
	if m.Header.S&1 != 0 {
		// 8 (SEID) + 3 (Sequence Number) + 1 (Message Priority and Spare)
		m.Header.MessageLength = 12
	} else {
		// 3 (Sequence Number) + 1 (Message Priority and Spare)
		m.Header.MessageLength = 4
	}
	m.Header.MessageLength += uint16(len(bodyBuf))
	headerBuf, _ = m.Header.MarshalBinary()
	return append(headerBuf, bodyBuf[:]...), nil
}
