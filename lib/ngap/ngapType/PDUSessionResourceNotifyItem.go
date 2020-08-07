package ngapType

import "my5G-RANTester/lib/aper"

// Need to import "free5gc/lib/aper" if it uses "aper"

type PDUSessionResourceNotifyItem struct {
	PDUSessionID                     PDUSessionID
	PDUSessionResourceNotifyTransfer aper.OctetString
	IEExtensions                     *ProtocolExtensionContainerPDUSessionResourceNotifyItemExtIEs `aper:"optional"`
}
