package ie

// PayloadCntr is detailed in 9.11.3.39 Payload container, 24.501
type PayloadCntr struct {
	// Name, uint8, Bits, Octet
	Pct      uint8 // ConstPayloadCntrType
	Contents []byte
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *PayloadCntr) UnmarshalBinary(b []byte, pct uint8) error {
	switch pct {
	case PayloadCntrType_MultiplePayloads, PayloadCntrType_EventNotif:
		var e error = &IEToDo{
			IEName: "(Multiple)PayloadCntr",
		}
		return e
	}
	value, err := copyIEValue("PayloadCntr", b, 1, 65535)
	if err != nil {
		return err
	}
	i.Pct = pct
	i.Contents = value
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *PayloadCntr) MarshalBinary() ([]byte, error) {
	switch i.Pct {
	case PayloadCntrType_MultiplePayloads, PayloadCntrType_EventNotif:
		var e error = &IEToDo{
			IEName: "(Multiple)PayloadCntr",
		}
		return nil, e
	}
	return copyIEValue("PayloadCntr", i.Contents, 1, 65535)
}
