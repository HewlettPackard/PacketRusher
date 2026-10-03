package ie

import "github.com/pkg/errors"

const (
	eapMsgValueMiniOctets int = 4
	eapMsgValueMaxOctets  int = 1500
)

type EAPMsg struct {
	Eap []byte
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *EAPMsg) UnmarshalBinary(b []byte) error {
	dataLen := len(b)
	if dataLen < eapMsgValueMiniOctets || dataLen > eapMsgValueMaxOctets {
		return errors.Errorf("EAPMsg Unmarshal err: data length (%d) should be >= %d and <= %d",
			dataLen, eapMsgValueMiniOctets, eapMsgValueMaxOctets)
	}
	i.Eap = b
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *EAPMsg) MarshalBinary() ([]byte, error) {
	dataLen := len(i.Eap)
	if dataLen < eapMsgValueMiniOctets || dataLen > eapMsgValueMaxOctets {
		return nil, errors.Errorf("EAPMsg Marshal err: data length(%d) should be >= %d and <= %d",
			dataLen, eapMsgValueMiniOctets, eapMsgValueMaxOctets)
	}
	return i.Eap, nil
}
