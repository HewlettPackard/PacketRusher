package nasConvert

import (
	"my5G-RANTester/lib/nas/nasMessage"
	"my5G-RANTester/lib/openapi/models"
)

func PDUSessionTypeToModels(nasPduSessType uint8) (pduSessType models.PduSessionType) {
	switch nasPduSessType {
	case nasMessage.PDUSessionTypeIPv4:
		pduSessType = models.PduSessionType_IPV4
	case nasMessage.PDUSessionTypeIPv6:
		pduSessType = models.PduSessionType_IPV6
	case nasMessage.PDUSessionTypeIPv4IPv6:
		pduSessType = models.PduSessionType_IPV4_V6
	case nasMessage.PDUSessionTypeUnstructured:
		pduSessType = models.PduSessionType_UNSTRUCTURED
	case nasMessage.PDUSessionTypeEthernet:
		pduSessType = models.PduSessionType_ETHERNET
	}

	return
}

func ModelsToPDUSessionType(pduSessType models.PduSessionType) (nasPduSessType uint8) {
	switch pduSessType {
	case models.PduSessionType_IPV4:
		nasPduSessType = nasMessage.PDUSessionTypeIPv4
	case models.PduSessionType_IPV6:
		nasPduSessType = nasMessage.PDUSessionTypeIPv6
	case models.PduSessionType_IPV4_V6:
		nasPduSessType = nasMessage.PDUSessionTypeIPv4IPv6
	case models.PduSessionType_UNSTRUCTURED:
		nasPduSessType = nasMessage.PDUSessionTypeUnstructured
	case models.PduSessionType_ETHERNET:
		nasPduSessType = nasMessage.PDUSessionTypeEthernet
	}
	return
}
