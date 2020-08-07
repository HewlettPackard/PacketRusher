package ngapType

import "my5G-RANTester/lib/aper"

// Need to import "free5gc/lib/aper" if it uses "aper"

type PDUSessionResourceModifyItemModInd struct {
	PDUSessionID                               PDUSessionID
	PDUSessionResourceModifyIndicationTransfer aper.OctetString
	IEExtensions                               *ProtocolExtensionContainerPDUSessionResourceModifyItemModIndExtIEs `aper:"optional"`
}
