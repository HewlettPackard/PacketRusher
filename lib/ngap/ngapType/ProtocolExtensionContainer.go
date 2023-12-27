package ngapType

import freeNgapType "github.com/free5gc/ngap/ngapType"

// Need to import "free5gc/lib/aper" if it uses "aper"

/* Sequence of = 35, FULL Name = struct ProtocolExtensionContainer_6500P129 */
/* QosFlowItemExtIEs */
type ProtocolExtensionContainerQosFlowItemExtIEs struct {
	List []QosFlowItemExtIEs `aper:"sizeLB:1,sizeUB:65535"`
}

type QosFlowItemExtIEs struct {
	Id             freeNgapType.ProtocolExtensionID
	Criticality    freeNgapType.Criticality
	ExtensionValue QosFlowItemExtIEsExtensionValue `aper:"openType,referenceFieldName:Id"`
}

type QosFlowItemExtIEsExtensionValue struct {
	Present int
}
