package ngapType

import "my5G-RANTester/lib/aper"

// Need to import "free5gc/lib/aper" if it uses "aper"

type PDUSessionResourceItemHORqd struct {
	PDUSessionID             PDUSessionID
	HandoverRequiredTransfer aper.OctetString
	IEExtensions             *ProtocolExtensionContainerPDUSessionResourceItemHORqdExtIEs `aper:"optional"`
}
