package ngapType

import "my5G-RANTester/lib/aper"

// Need to import "free5gc/lib/aper" if it uses "aper"

type PDUSessionResourceReleasedItemNot struct {
	PDUSessionID                             PDUSessionID
	PDUSessionResourceNotifyReleasedTransfer aper.OctetString
	IEExtensions                             *ProtocolExtensionContainerPDUSessionResourceReleasedItemNotExtIEs `aper:"optional"`
}
