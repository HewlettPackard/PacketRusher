package ngapType

import "my5G-RANTester/lib/aper"

// Need to import "free5gc/lib/aper" if it uses "aper"

const (
	RRCStatePresentInactive  aper.Enumerated = 0
	RRCStatePresentConnected aper.Enumerated = 1
)

type RRCState struct {
	Value aper.Enumerated `aper:"valueExt,valueLB:0,valueUB:1"`
}
