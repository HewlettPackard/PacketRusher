package ngapType

import "my5G-RANTester/lib/aper"

// Need to import "free5gc/lib/aper" if it uses "aper"

type PDUSessionResourceHandoverItem struct {
	PDUSessionID            PDUSessionID
	HandoverCommandTransfer aper.OctetString
	IEExtensions            *ProtocolExtensionContainerPDUSessionResourceHandoverItemExtIEs `aper:"optional"`
}
