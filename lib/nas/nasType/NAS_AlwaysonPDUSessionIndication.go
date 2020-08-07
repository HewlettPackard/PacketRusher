package nasType

// AlwaysonPDUSessionIndication 9.11.4.3
// Iei Row, sBit, len = [0, 0], 8 , 4
// APSI Row, sBit, len = [0, 0], 1 , 1
type AlwaysonPDUSessionIndication struct {
	Octet uint8
}

func NewAlwaysonPDUSessionIndication(iei uint8) (alwaysonPDUSessionIndication *AlwaysonPDUSessionIndication) {
	alwaysonPDUSessionIndication = &AlwaysonPDUSessionIndication{}
	alwaysonPDUSessionIndication.SetIei(iei)
	return alwaysonPDUSessionIndication
}

// AlwaysonPDUSessionIndication 9.11.4.3
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *AlwaysonPDUSessionIndication) GetIei() (iei uint8) {
	return a.Octet & GetBitMask(8, 4) >> (4)
}

// AlwaysonPDUSessionIndication 9.11.4.3
// Iei Row, sBit, len = [0, 0], 8 , 4
func (a *AlwaysonPDUSessionIndication) SetIei(iei uint8) {
	a.Octet = (a.Octet & 15) + ((iei & 15) << 4)
}

// AlwaysonPDUSessionIndication 9.11.4.3
// APSI Row, sBit, len = [0, 0], 1 , 1
func (a *AlwaysonPDUSessionIndication) GetAPSI() (aPSI uint8) {
	return a.Octet & GetBitMask(1, 0)
}

// AlwaysonPDUSessionIndication 9.11.4.3
// APSI Row, sBit, len = [0, 0], 1 , 1
func (a *AlwaysonPDUSessionIndication) SetAPSI(aPSI uint8) {
	a.Octet = (a.Octet & 254) + (aPSI & 1)
}
