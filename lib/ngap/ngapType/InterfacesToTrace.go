package ngapType

import "my5G-RANTester/lib/aper"

// Need to import "free5gc/lib/aper" if it uses "aper"

type InterfacesToTrace struct {
	Value aper.BitString `aper:"sizeLB:8,sizeUB:8"`
}
