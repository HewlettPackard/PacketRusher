package ngapType

import "my5G-RANTester/lib/aper"

// Need to import "free5gc/lib/aper" if it uses "aper"

type PDUSessionResourceReleasedItemPSAck struct {
	PDUSessionID                          PDUSessionID
	PathSwitchRequestUnsuccessfulTransfer aper.OctetString
	IEExtensions                          *ProtocolExtensionContainerPDUSessionResourceReleasedItemPSAckExtIEs `aper:"optional"`
}
