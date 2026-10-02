package ie

import "github.com/pkg/errors"

// Cause5GMM is detailed in 9.11.3.2 5GMM cause, 24.501
type Cause5GMM struct {
	// Name, uint8, Bits, Octet
	Value uint8 // 8 -> 1 ,   2 -> 2
}

const (
	Cause5GMM_None                                     uint8 = 0x00
	Cause5GMM_IllegalUE                                uint8 = 0x03
	Cause5GMM_PEINotAccepted                           uint8 = 0x05
	Cause5GMM_IllegalME                                uint8 = 0x06
	Cause5GMM_5GSSvcNotAllowed                         uint8 = 0x07
	Cause5GMM_UEIdCannotBeDerivedByTheNw               uint8 = 0x09
	Cause5GMM_ImplicitlyDeregistered                   uint8 = 0x0a
	Cause5GMM_PLMNNotAllowed                           uint8 = 0x0b
	Cause5GMM_TrackingAreaNotAllowed                   uint8 = 0x0c
	Cause5GMM_RoamingNotAllowedInThisTrackingArea      uint8 = 0x0d
	Cause5GMM_NoSuitableCellsInTrackingArea            uint8 = 0x0f
	Cause5GMM_MACFailure                               uint8 = 0x14
	Cause5GMM_SynchFailure                             uint8 = 0x15
	Cause5GMM_Congestion                               uint8 = 0x16
	Cause5GMM_UESecCapabilitiesMismatch                uint8 = 0x17
	Cause5GMM_SecModeRejected                          uint8 = 0x18
	Cause5GMM_Non5GAuthUnacceptable                    uint8 = 0x1a
	Cause5GMM_N1ModeNotAllowed                         uint8 = 0x1b
	Cause5GMM_RestrictedSvcArea                        uint8 = 0x1c
	Cause5GMM_RedirToEPCRequired                       uint8 = 0x1f
	Cause5GMM_IABNodeOperationNotAutho                 uint8 = 0x24
	Cause5GMM_LADNNotAvailable                         uint8 = 0x2b
	Cause5GMM_NoNwSlicesAvailable                      uint8 = 0x3e
	Cause5GMM_MaxNumOfPDUSessionsReached               uint8 = 0x41
	Cause5GMM_InsufRsrcForSpecificSliceAndDNN          uint8 = 0x43
	Cause5GMM_InsufRsrcForSpecificSlice                uint8 = 0x45
	Cause5GMM_NgksiAlreadyInUse                        uint8 = 0x47
	Cause5GMM_Non3GPPAccessTo5GCNNotAllowed            uint8 = 0x48
	Cause5GMM_ServingNwNotAutho                        uint8 = 0x49
	Cause5GMM_TempNotAuthoForThisSNPN                  uint8 = 0x4a
	Cause5GMM_PermNotAuthoForThisSNPN                  uint8 = 0x4b
	Cause5GMM_NotAuthoForThisCAGOrAuthoForCAGCellsOnly uint8 = 0x4c
	Cause5GMM_WirelineAccessAreaNotAllowed             uint8 = 0x4d
	Cause5GMM_PLMNNotAllowed_UELocation                uint8 = 0x4e
	Cause5GMM_UASSvcNotAllowed                         uint8 = 0x4f
	Cause5GMM_DisasterRoamingNotAllowed                uint8 = 0x50
	Cause5GMM_PayloadWasNotForwarded                   uint8 = 0x5a
	Cause5GMM_DNNNotSupportedOrNotSubscribedInTheSlice uint8 = 0x5b
	Cause5GMM_InsufUserPlaneRsrcForThePDUSess          uint8 = 0x5c
	Cause5GMM_OnboardingSvcTerm                        uint8 = 0x5d
	Cause5GMM_SemanticallyIncorrectMsg                 uint8 = 0x5f
	Cause5GMM_InvalidMandatoryInfo                     uint8 = 0x60
	Cause5GMM_MsgTypeNonExistentOrNotImpl              uint8 = 0x61
	Cause5GMM_MsgTypeNotCompatibleWithTheProtState     uint8 = 0x62
	Cause5GMM_InfoElementNonExistentOrNotImpl          uint8 = 0x63
	Cause5GMM_ConditionalIEErr                         uint8 = 0x64
	Cause5GMM_MsgNotCompatibleWithTheProtState         uint8 = 0x65
	Cause5GMM_ProtError                                uint8 = 0x6f
)

var cause5GSMMStr = map[uint8]string{
	Cause5GMM_None:                                     "None",
	Cause5GMM_IllegalUE:                                "Illegal UE",
	Cause5GMM_PEINotAccepted:                           "PEI Not Accepted",
	Cause5GMM_IllegalME:                                "Illegal ME",
	Cause5GMM_5GSSvcNotAllowed:                         "5GS Service Not Allowed",
	Cause5GMM_UEIdCannotBeDerivedByTheNw:               "UE Identity Cannot Be Derived By The Network",
	Cause5GMM_ImplicitlyDeregistered:                   "Implicitly Deregistered",
	Cause5GMM_PLMNNotAllowed:                           "PLMN Not Allowed",
	Cause5GMM_TrackingAreaNotAllowed:                   "Tracking Area Not Allowed",
	Cause5GMM_RoamingNotAllowedInThisTrackingArea:      "Roaming Not Allowed In This Tracking Area",
	Cause5GMM_NoSuitableCellsInTrackingArea:            "No Suitable Cells In Tracking Area",
	Cause5GMM_MACFailure:                               "MAC Failure",
	Cause5GMM_SynchFailure:                             "Synch Failure",
	Cause5GMM_Congestion:                               "Congestion",
	Cause5GMM_UESecCapabilitiesMismatch:                "UE Security Capabilities Mismatch",
	Cause5GMM_SecModeRejected:                          "Security Mode Rejected",
	Cause5GMM_Non5GAuthUnacceptable:                    "Non5G Authentication Unacceptable",
	Cause5GMM_N1ModeNotAllowed:                         "N1 Mode Not Allowed",
	Cause5GMM_RestrictedSvcArea:                        "Restricted Service Area",
	Cause5GMM_RedirToEPCRequired:                       "Recirection To EPC Required",
	Cause5GMM_IABNodeOperationNotAutho:                 "IAB-node operation not authorized",
	Cause5GMM_LADNNotAvailable:                         "LADN Not Available",
	Cause5GMM_NoNwSlicesAvailable:                      "No Network Slices Available",
	Cause5GMM_MaxNumOfPDUSessionsReached:               "Maximum Number Of PDU Sessions Reached",
	Cause5GMM_InsufRsrcForSpecificSliceAndDNN:          "Insufficient Resources For Specific Slice And DNN",
	Cause5GMM_InsufRsrcForSpecificSlice:                "Insufficient Resources For Specific Slice",
	Cause5GMM_NgksiAlreadyInUse:                        "ngksi Already In Use",
	Cause5GMM_Non3GPPAccessTo5GCNNotAllowed:            "Non 3GPP Access To 5GCN Not Allowed",
	Cause5GMM_ServingNwNotAutho:                        "Serving Network Not Authorized",
	Cause5GMM_TempNotAuthoForThisSNPN:                  "Temporarily Not Authorized For This SNPN",
	Cause5GMM_PermNotAuthoForThisSNPN:                  "Permanently Not Authorized For This SNPN",
	Cause5GMM_NotAuthoForThisCAGOrAuthoForCAGCellsOnly: "Not Authorized For This CAG Or Authorized For CAG Cells Only",
	Cause5GMM_WirelineAccessAreaNotAllowed:             "Wireline Access Area Not Allowed",
	Cause5GMM_PLMNNotAllowed_UELocation:                "PLMN not allowed to operate at the present UE location",
	Cause5GMM_UASSvcNotAllowed:                         "UAS services not allowed",
	Cause5GMM_PayloadWasNotForwarded:                   "Payload Was Not Forwarded",
	Cause5GMM_DisasterRoamingNotAllowed: "Disaster roaming for the determined PLMN with " +
		"disaster condition not allowed",
	Cause5GMM_DNNNotSupportedOrNotSubscribedInTheSlice: "DNN Not Supported Or Not Subscribed In The Slice",
	Cause5GMM_InsufUserPlaneRsrcForThePDUSess:          "Insufficient User Plane resources For The PDU Sess",
	Cause5GMM_OnboardingSvcTerm:                        "Onboarding services terminated",
	Cause5GMM_SemanticallyIncorrectMsg:                 "Semantically Incorrect Message",
	Cause5GMM_InvalidMandatoryInfo:                     "Invalid Mandatory Information",
	Cause5GMM_MsgTypeNonExistentOrNotImpl:              "Message Type Non-Existent Or Not Implemented",
	Cause5GMM_MsgTypeNotCompatibleWithTheProtState:     "Message Type Not Compatible With The Protocol State",
	Cause5GMM_InfoElementNonExistentOrNotImpl:          "Information Element Non-Existent Or Not Implemented",
	Cause5GMM_ConditionalIEErr:                         "Conditional IE Err",
	Cause5GMM_MsgNotCompatibleWithTheProtState:         "Message Not Compatible With The Protocol State",
	Cause5GMM_ProtError:                                "Protocol Error, unspecified",
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *Cause5GMM) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("the Cause5GMM IE length(%d) is incorrect", len(b))
	}
	if i.Value = b[0]; i.String() == "" {
		i.Value = Cause5GMM_ProtError
	}
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *Cause5GMM) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = i.Value

	return b, nil
}

func (i *Cause5GMM) String() string {
	return cause5GSMMStr[i.Value]
}
