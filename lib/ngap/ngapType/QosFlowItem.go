package ngapType

import freeNgapType "github.com/free5gc/ngap/ngapType"

// Need to import "free5gc/lib/aper" if it uses "aper"

type QosFlowItem struct {
	QosFlowIdentifier freeNgapType.QosFlowIdentifier
	Cause             freeNgapType.Cause                           `aper:"valueLB:0,valueUB:5"`
	IEExtensions      *ProtocolExtensionContainerQosFlowItemExtIEs `aper:"optional"`
}
