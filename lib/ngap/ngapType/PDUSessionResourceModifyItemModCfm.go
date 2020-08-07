package ngapType

import "my5G-RANTester/lib/aper"

// Need to import "free5gc/lib/aper" if it uses "aper"

type PDUSessionResourceModifyItemModCfm struct {
	PDUSessionID                            PDUSessionID
	PDUSessionResourceModifyConfirmTransfer aper.OctetString
	IEExtensions                            *ProtocolExtensionContainerPDUSessionResourceModifyItemModCfmExtIEs `aper:"optional"`
}
