package ngapType

import "my5G-RANTester/lib/aper"

// Need to import "free5gc/lib/aper" if it uses "aper"

type PDUSessionResourceFailedToSetupItemHOAck struct {
	PDUSessionID                                   PDUSessionID
	HandoverResourceAllocationUnsuccessfulTransfer aper.OctetString
	IEExtensions                                   *ProtocolExtensionContainerPDUSessionResourceFailedToSetupItemHOAckExtIEs `aper:"optional"`
}
