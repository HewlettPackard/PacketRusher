package ngapType

import "my5G-RANTester/lib/aper"

// Need to import "free5gc/lib/aper" if it uses "aper"

const (
	MICOModeIndicationPresentTrue aper.Enumerated = 0
)

type MICOModeIndication struct {
	Value aper.Enumerated `aper:"valueExt,valueLB:0,valueUB:0"`
}
