package pfcp

import (
	"fmt"
	"free5gc/lib/tlv"
)

func (m *Message) Unmarshal(data []byte) (err error) {
	_ = m.Header.UnmarshalBinary(data)
	// Check Message Length field in header
	if int(m.Header.MessageLength) != len(data)-4 {
		return fmt.Errorf("Incorrect Message Length: Expected %d, got %d", m.Header.MessageLength, len(data)-4)
	}
	switch m.Header.MessageType {
	case PFCP_HEARTBEAT_REQUEST:
		Body := HeartbeatRequest{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_HEARTBEAT_RESPONSE:
		Body := HeartbeatResponse{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_PFD_MANAGEMENT_REQUEST:
		Body := PFCPPFDManagementRequest{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_PFD_MANAGEMENT_RESPONSE:
		Body := PFCPPFDManagementResponse{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_ASSOCIATION_SETUP_REQUEST:
		Body := PFCPAssociationSetupRequest{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_ASSOCIATION_SETUP_RESPONSE:
		Body := PFCPAssociationSetupResponse{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_ASSOCIATION_UPDATE_REQUEST:
		Body := PFCPAssociationUpdateRequest{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_ASSOCIATION_UPDATE_RESPONSE:
		Body := PFCPAssociationUpdateResponse{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_ASSOCIATION_RELEASE_REQUEST:
		Body := PFCPAssociationReleaseRequest{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_ASSOCIATION_RELEASE_RESPONSE:
		Body := PFCPAssociationReleaseResponse{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_NODE_REPORT_REQUEST:
		Body := PFCPNodeReportRequest{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_NODE_REPORT_RESPONSE:
		Body := PFCPNodeReportResponse{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_SESSION_SET_DELETION_REQUEST:
		Body := PFCPSessionSetDeletionRequest{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_SESSION_SET_DELETION_RESPONSE:
		Body := PFCPSessionSetDeletionResponse{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_SESSION_ESTABLISHMENT_REQUEST:
		Body := PFCPSessionEstablishmentRequest{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_SESSION_ESTABLISHMENT_RESPONSE:
		Body := PFCPSessionEstablishmentResponse{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_SESSION_MODIFICATION_REQUEST:
		Body := PFCPSessionModificationRequest{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_SESSION_MODIFICATION_RESPONSE:
		Body := PFCPSessionModificationResponse{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_SESSION_DELETION_REQUEST:
		Body := PFCPSessionDeletionRequest{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_SESSION_DELETION_RESPONSE:
		Body := PFCPSessionDeletionResponse{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_SESSION_REPORT_REQUEST:
		Body := PFCPSessionReportRequest{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	case PFCP_SESSION_REPORT_RESPONSE:
		Body := PFCPSessionReportResponse{}
		err = tlv.Unmarshal(data[m.Header.Len():], &Body)
		m.Body = Body
	default:
		return fmt.Errorf("not support m type %d", m.Header.MessageType)
	}
	return err
}
