package ngapType

import "my5G-RANTester/lib/aper"

// Need to import "free5gc/lib/aper" if it uses "aper"

const (
	IntegrityProtectionIndicationPresentRequired  aper.Enumerated = 0
	IntegrityProtectionIndicationPresentPreferred aper.Enumerated = 1
	IntegrityProtectionIndicationPresentNotNeeded aper.Enumerated = 2
)

type IntegrityProtectionIndication struct {
	Value aper.Enumerated `aper:"valueExt,valueLB:0,valueUB:2"`
}
