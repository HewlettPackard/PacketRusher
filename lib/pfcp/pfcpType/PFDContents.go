package pfcpType

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// PFDContents - describe in TS 29.244 Figure 8.2.39-1: PFD Contents
type PFDContents struct {
	FlowDescription  string
	URL              string
	DomainName       string
	CustomPFDContent []byte
}

func (p *PFDContents) MarshalBinary() (data []byte, err error) {
	buf := bytes.NewBuffer(nil)
	var presenceByte, spareByte byte
	// set presence header
	if p.FlowDescription != "" {
		presenceByte |= BitMask1
	}
	if p.URL != "" {
		presenceByte |= BitMask2
	}
	if p.DomainName != "" {
		presenceByte |= BitMask3
	}
	if p.CustomPFDContent != nil {
		presenceByte |= BitMask4
	}

	binary.Write(buf, binary.BigEndian, presenceByte)
	binary.Write(buf, binary.BigEndian, spareByte)

	if p.FlowDescription != "" {
		binary.Write(buf, binary.BigEndian, uint16(len(p.FlowDescription)))
		buf.WriteString(p.FlowDescription)
	}

	if p.URL != "" {
		binary.Write(buf, binary.BigEndian, uint16(len(p.URL)))
		buf.WriteString(p.URL)
	}

	if p.DomainName != "" {
		binary.Write(buf, binary.BigEndian, uint16(len(p.DomainName)))
		buf.WriteString(p.DomainName)
	}

	if p.CustomPFDContent != nil {
		binary.Write(buf, binary.BigEndian, uint16(len(p.CustomPFDContent)))
		binary.Write(buf, binary.BigEndian, p.CustomPFDContent)
	}

	return buf.Bytes(), nil
}

func (p *PFDContents) UnmarshalBinary(data []byte) error {
	length := uint16(len(data))

	var idx uint16 = 0

	// Octet 5
	if length < idx+1 {
		return fmt.Errorf("Inadequate TLV length: %d", length)
	}

	var presenceByte = data[idx]
	// presenceByte & spareByte
	idx = idx + 2

	// flow Description presence
	if utob(presenceByte & BitMask1) {
		if length < idx+2 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		flowDescriptionLen := binary.BigEndian.Uint16(data[idx:])
		idx = idx + 2

		if length < idx+flowDescriptionLen {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		p.FlowDescription = string(data[idx : idx+flowDescriptionLen])
		idx = idx + flowDescriptionLen
	}

	// URL presence
	if utob(presenceByte & BitMask2) {
		if length < idx+2 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		urlLen := binary.BigEndian.Uint16(data[idx:])
		idx = idx + 2

		if length < idx+urlLen {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		p.URL = string(data[idx : idx+urlLen])
		idx = idx + urlLen
	}

	// domain name presence
	if utob(presenceByte & BitMask3) {
		if length < idx+2 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		domainNameLen := binary.BigEndian.Uint16(data[idx:])
		idx = idx + 2

		if length < idx+domainNameLen {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		p.DomainName = string(data[idx : idx+domainNameLen])
		idx = idx + domainNameLen
	}

	// custom PFD content presence
	if utob(presenceByte & BitMask4) {
		if length < idx+2 {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		custemPFDContentLen := binary.BigEndian.Uint16(data[idx:])
		idx = idx + 2

		if length < idx+custemPFDContentLen {
			return fmt.Errorf("Inadequate TLV length: %d", length)
		}
		p.CustomPFDContent = data[idx : idx+custemPFDContentLen]
		idx = idx + custemPFDContentLen
	}

	return nil
}
