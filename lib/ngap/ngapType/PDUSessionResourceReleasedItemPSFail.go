package ngapType

import "my5G-RANTester/lib/aper"

// Need to import "free5gc/lib/aper" if it uses "aper"

type PDUSessionResourceReleasedItemPSFail struct {
	PDUSessionID                          PDUSessionID
	PathSwitchRequestUnsuccessfulTransfer aper.OctetString
	IEExtensions                          *ProtocolExtensionContainerPDUSessionResourceReleasedItemPSFailExtIEs `aper:"optional"`
}
