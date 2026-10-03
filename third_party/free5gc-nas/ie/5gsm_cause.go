package ie

import "github.com/pkg/errors"

// Cause5GSM is detailed in 9.11.4.2 5GSM cause, 24.501
type Cause5GSM struct {
	// Name, uint8, Bits, Octet
	Value uint8 // 8 -> 1 ,   2 -> 2
}

const (
	Cause5GSM_OperatorDeterminedBarring                               uint8 = 0x08
	Cause5GSM_InsufRsrc                                               uint8 = 0x1a
	Cause5GSM_MissingOrUnknownDNN                                     uint8 = 0x1b
	Cause5GSM_UnknownPDUSessType                                      uint8 = 0x1c
	Cause5GSM_UserAuthOrAuthorizationFailed                           uint8 = 0x1d
	Cause5GSM_ReqRejected                                             uint8 = 0x1f
	Cause5GSM_SvcOptNotSupported                                      uint8 = 0x20
	Cause5GSM_ReqSvcOptNotSubscribed                                  uint8 = 0x21
	Cause5GSM_PTIAlreadyInUse                                         uint8 = 0x23
	Cause5GSM_RegularDeactivation                                     uint8 = 0x24
	Cause5GSM_QosNotAccepted                                          uint8 = 0x25
	Cause5GSM_NwFailure                                               uint8 = 0x26
	Cause5GSM_ReactivationReq                                         uint8 = 0x27
	Cause5GSM_SemanticErrInTheTFTOperation                            uint8 = 0x29
	Cause5GSM_SyntacticalErrInTheTFTOperation                         uint8 = 0x2a
	Cause5GSM_InvalidPDUSessId                                        uint8 = 0x2b
	Cause5GSM_SemanticErrorsInPktFilterS                              uint8 = 0x2c
	Cause5GSM_SyntacticalErrInPktFilterS                              uint8 = 0x2d
	Cause5GSM_OutOfLADNSvcArea                                        uint8 = 0x2e
	Cause5GSM_PTIMismatch                                             uint8 = 0x2f
	Cause5GSM_PDUSessTypeIpv4OnlyAllowed                              uint8 = 0x32
	Cause5GSM_PDUSessTypeIpv6OnlyAllowed                              uint8 = 0x33
	Cause5GSM_PDUSessDoesNotExist                                     uint8 = 0x36
	Cause5GSM_PDUSessTypeIpv4V6OnlyAllowed                            uint8 = 0x39
	Cause5GSM_PDUSessTypeUnstructuredOnlyAllowed                      uint8 = 0x3a
	Cause5GSM_Unsupported5QIValue                                     uint8 = 0x3b
	Cause5GSM_PDUSessTypeEthOnlyAllowed                               uint8 = 0x3d
	Cause5GSM_InsufRsrcForSpecificSliceAndDNN                         uint8 = 0x43
	Cause5GSM_NotSupportedSSCMode                                     uint8 = 0x44
	Cause5GSM_InsufRsrcForSpecificSlice                               uint8 = 0x45
	Cause5GSM_MissingOrUnknownDNNInASlice                             uint8 = 0x46
	Cause5GSM_InvalidPTIValue                                         uint8 = 0x51
	Cause5GSM_MaxDataRatePerUEForUserPlaneIntegrityProtectionIsTooLow uint8 = 0x52
	Cause5GSM_SemanticErrInTheQosOperation                            uint8 = 0x53
	Cause5GSM_SyntacticalErrInTheQosOperation                         uint8 = 0x54
	Cause5GSM_InvalidMappedEPSBearerId                                uint8 = 0x55
	Cause5GSM_UASSvcNotAllowed                                        uint8 = 0x56
	Cause5GSM_SemanticallyIncorrectMsg                                uint8 = 0x5f
	Cause5GSM_InvalidMandatoryInfo                                    uint8 = 0x60
	Cause5GSM_MsgTypeNonExistentOrNotImpl                             uint8 = 0x61
	Cause5GSM_MsgTypeNotCompatibleWithTheProtState                    uint8 = 0x62
	Cause5GSM_InfoElementNonExistentOrNotImpl                         uint8 = 0x63
	Cause5GSM_ConditionalIEErr                                        uint8 = 0x64
	Cause5GSM_MsgNotCompatibleWithTheProtState                        uint8 = 0x65
	Cause5GSM_ProtError                                               uint8 = 0x6f
	Cause5GSM_Unspecified                                             uint8 = 0x6f
)

var cause5GSMStr = map[uint8]string{
	Cause5GSM_OperatorDeterminedBarring:            "Operator determined barring",
	Cause5GSM_InsufRsrc:                            "Insufficient resources",
	Cause5GSM_MissingOrUnknownDNN:                  "Missing or unknown DNN",
	Cause5GSM_UnknownPDUSessType:                   "Unknown PDU session type",
	Cause5GSM_UserAuthOrAuthorizationFailed:        "User authentication or authorization failed",
	Cause5GSM_ReqRejected:                          "Request rejected",
	Cause5GSM_SvcOptNotSupported:                   "Service option not supported",
	Cause5GSM_ReqSvcOptNotSubscribed:               "Requested service option not subscribed",
	Cause5GSM_PTIAlreadyInUse:                      "PTI already in use",
	Cause5GSM_RegularDeactivation:                  "Regular deactivation",
	Cause5GSM_NwFailure:                            "Network failure",
	Cause5GSM_ReactivationReq:                      "Reactivation required",
	Cause5GSM_SemanticErrInTheTFTOperation:         "Semantic error in the TFT operation",
	Cause5GSM_SyntacticalErrInTheTFTOperation:      "Syntactical error in the TFT operation",
	Cause5GSM_InvalidPDUSessId:                     "Invalid PDU session ID",
	Cause5GSM_SemanticErrorsInPktFilterS:           "Semantic errors in packet filter(s)",
	Cause5GSM_SyntacticalErrInPktFilterS:           "Syntactical error in packet filter(s)",
	Cause5GSM_OutOfLADNSvcArea:                     "Out of LADN service area",
	Cause5GSM_PTIMismatch:                          "PTI mismatch",
	Cause5GSM_PDUSessTypeIpv4OnlyAllowed:           "PDU session type IPv4 only allowed",
	Cause5GSM_PDUSessTypeIpv6OnlyAllowed:           "PDU session type IPv6 only allowed",
	Cause5GSM_PDUSessDoesNotExist:                  "PDU session does not exist",
	Cause5GSM_PDUSessTypeIpv4V6OnlyAllowed:         "PDU session type IPv4v6 only allowed",
	Cause5GSM_PDUSessTypeUnstructuredOnlyAllowed:   "PDU session type unstructured only allowed",
	Cause5GSM_Unsupported5QIValue:                  "Unsupported 5QI value",
	Cause5GSM_PDUSessTypeEthOnlyAllowed:            "PDU session type Ethernet only allowed",
	Cause5GSM_InsufRsrcForSpecificSliceAndDNN:      "Insufficient resources for specific slice and DNN",
	Cause5GSM_NotSupportedSSCMode:                  "Not supported SSC mode",
	Cause5GSM_InsufRsrcForSpecificSlice:            "Insufficient resources for specific slice",
	Cause5GSM_MissingOrUnknownDNNInASlice:          "Missing or unknown DNN in a slice",
	Cause5GSM_InvalidPTIValue:                      "Invalid PTI value",
	Cause5GSM_SemanticErrInTheQosOperation:         "Semantic error in the QoS operation",
	Cause5GSM_SyntacticalErrInTheQosOperation:      "Syntactical error in the QoS operation",
	Cause5GSM_InvalidMappedEPSBearerId:             "Invalid mapped EPS bearer ID",
	Cause5GSM_SemanticallyIncorrectMsg:             "Semantically incorrect message",
	Cause5GSM_InvalidMandatoryInfo:                 "Invalid mandatory information",
	Cause5GSM_MsgTypeNonExistentOrNotImpl:          "Message type non-existent or not implemented",
	Cause5GSM_MsgTypeNotCompatibleWithTheProtState: "Message type not compatible with the protocol state",
	Cause5GSM_InfoElementNonExistentOrNotImpl:      "Information element non-existent or not implemented",
	Cause5GSM_ConditionalIEErr:                     "Conditional IE error",
	Cause5GSM_MsgNotCompatibleWithTheProtState:     "Message not compatible with the protocol state",
	Cause5GSM_ProtError:                            "Protocol error/Unspecified",

	// avoids line too long
	Cause5GSM_MaxDataRatePerUEForUserPlaneIntegrityProtectionIsTooLow: "Maximum data rate per UE" +
		" for user plane integrity protection is too low",
}

// UnmarshalBinary handles the value part, not including IEI and length.
func (i *Cause5GSM) UnmarshalBinary(b []byte) error {
	if len(b) != 1 {
		return errors.Errorf("The Cause5GSM IE length(%d) is incorrect", len(b))
	}
	i.Value = b[0]
	return nil
}

// MarshalBinary returns the value part, not including IEI and length.
func (i *Cause5GSM) MarshalBinary() ([]byte, error) {
	b := make([]byte, 1)
	b[0] = i.Value

	return b, nil
}

func (i *Cause5GSM) String() string {
	if s, ok := cause5GSMStr[i.Value]; ok {
		return s
	}
	return "Protocol error/Unspecified"
}
