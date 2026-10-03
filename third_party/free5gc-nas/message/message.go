package message

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"

	"github.com/pkg/errors"
)

type SecHdrType uint8

const (
	SecHdrTypePlainNas                                        SecHdrType = 0x00
	SecHdrTypeIntegrityProtected                              SecHdrType = 0x01
	SecHdrTypeIntegrityProtectedAndCiphered                   SecHdrType = 0x02
	SecHdrTypeIntegrityProtectedWithNew5gNasSecCtx            SecHdrType = 0x03
	SecHdrTypeIntegrityProtectedAndCipheredWithNew5gNasSecCtx SecHdrType = 0x04
	// Customized value to indicate any error happened when getting SecHdrType
	SecHdrTypeErr SecHdrType = 0xff
)

var secHdrTypeString = []string{
	"SecHdrTypePlainNas",
	"SecHdrTypeIntegrityProtected",
	"SecHdrTypeIntegrityProtectedAndCiphered",
	"SecHdrTypeIntegrityProtectedWithNew5gNasSecCtx",
	"SecHdrTypeIntegrityProtectedAndCipheredWithNew5gNasSecCtx",
	"SecHdrTypeErr",
}

func (t SecHdrType) String() string {
	return secHdrTypeString[t]
}

type Epd uint8

const (
	Epd5GSSessMgmtMsg     Epd = 0x2E
	Epd5GSMobilityMgmtMsg Epd = 0x7E
)

type MsgType uint8

const (
	MsgTypeRegReq                      MsgType = 65
	MsgTypeRegAccept                   MsgType = 66
	MsgTypeRegComplete                 MsgType = 67
	MsgTypeRegRej                      MsgType = 68
	MsgTypeDeregReqUEOrig              MsgType = 69
	MsgTypeDeregAcceptUEOrig           MsgType = 70
	MsgTypeDeregReqUETerm              MsgType = 71
	MsgTypeDeregAcceptUETerm           MsgType = 72
	MsgTypeSvcReq                      MsgType = 76
	MsgTypeSvcRej                      MsgType = 77
	MsgTypeSvcAccept                   MsgType = 78
	MsgTypeCtrlPlaneSvcReq             MsgType = 79
	MsgTypeNwSliceSpecificAuthCmd      MsgType = 80
	MsgTypeNwSliceSpecificAuthComplete MsgType = 81
	MsgTypeNwSliceSpecificAuthResult   MsgType = 82
	MsgTypeCfgUpdateCmd                MsgType = 84
	MsgTypeCfgUpdateComplete           MsgType = 85
	MsgTypeAuthReq                     MsgType = 86
	MsgTypeAuthRsp                     MsgType = 87
	MsgTypeAuthRej                     MsgType = 88
	MsgTypeAuthFailure                 MsgType = 89
	MsgTypeAuthResult                  MsgType = 90
	MsgTypeIdReq                       MsgType = 91
	MsgTypeIdRsp                       MsgType = 92
	MsgTypeSecModeCmd                  MsgType = 93
	MsgTypeSecModeComplete             MsgType = 94
	MsgTypeSecModeRej                  MsgType = 95
	MsgTypeStatus5GMM                  MsgType = 100
	MsgTypeNotif                       MsgType = 101
	MsgTypeNotifRsp                    MsgType = 102
	MsgTypeULNASTransport              MsgType = 103
	MsgTypeDLNASTransport              MsgType = 104
	MsgTypeRelayKeyReq                 MsgType = 105
	MsgTypeRelayKeyAccept              MsgType = 106
	MsgTypeRelayKeyRej                 MsgType = 107
	MsgTypeRelayAuthReq                MsgType = 108
	MsgTypeRelayAuthRsp                MsgType = 109

	MsgTypePDUSessEstReq       MsgType = 193
	MsgTypePDUSessEstAccept    MsgType = 194
	MsgTypePDUSessEstRej       MsgType = 195
	MsgTypePDUSessAuthCmd      MsgType = 197
	MsgTypePDUSessAuthComplete MsgType = 198
	MsgTypePDUSessAuthResult   MsgType = 199
	MsgTypePDUSessModReq       MsgType = 201
	MsgTypePDUSessModRej       MsgType = 202
	MsgTypePDUSessModCmd       MsgType = 203
	MsgTypePDUSessModComplete  MsgType = 204
	MsgTypePDUSessModCmdRej    MsgType = 205
	MsgTypePDUSessRelReq       MsgType = 209
	MsgTypePDUSessRelRej       MsgType = 210
	MsgTypePDUSessRelCmd       MsgType = 211
	MsgTypePDUSessRelComplete  MsgType = 212
	MsgTypeStatus5GSM          MsgType = 214
	MsgTypeSvcLvlAuthCmd       MsgType = 216
	MsgTypeSvcLvlAuthComplete  MsgType = 217
	MsgTypeRemoteUEReport      MsgType = 218
	MsgTypeRemoteUEReportRsp   MsgType = 219
)

var msgTypeMap = map[MsgType]string{
	MsgTypeRegReq:                      "RegReq",
	MsgTypeRegAccept:                   "RegAccept",
	MsgTypeRegComplete:                 "RegComplete",
	MsgTypeRegRej:                      "RegRej",
	MsgTypeDeregReqUEOrig:              "DeregReqUEOrig",
	MsgTypeDeregAcceptUEOrig:           "DeregAcceptUEOrig",
	MsgTypeDeregReqUETerm:              "DeregReqUETerm",
	MsgTypeDeregAcceptUETerm:           "DeregAcceptUETerm",
	MsgTypeSvcReq:                      "SvcReq",
	MsgTypeSvcRej:                      "SvcRej",
	MsgTypeSvcAccept:                   "SvcAccept",
	MsgTypeCtrlPlaneSvcReq:             "CtrlPlaneSvcReq",
	MsgTypeNwSliceSpecificAuthCmd:      "NwSliceSpecificAuthCmd",
	MsgTypeNwSliceSpecificAuthComplete: "NwSliceSpecificAuthComplete",
	MsgTypeNwSliceSpecificAuthResult:   "NwSliceSpecificAuthResult",
	MsgTypeCfgUpdateCmd:                "CfgUpdateCmd",
	MsgTypeCfgUpdateComplete:           "CfgUpdateComplete",
	MsgTypeAuthReq:                     "AuthReq",
	MsgTypeAuthRsp:                     "AuthRsp",
	MsgTypeAuthRej:                     "AuthRej",
	MsgTypeAuthFailure:                 "AuthFailure",
	MsgTypeAuthResult:                  "AuthResult",
	MsgTypeIdReq:                       "IdReq",
	MsgTypeIdRsp:                       "IdRsp",
	MsgTypeSecModeCmd:                  "SecModeCmd",
	MsgTypeSecModeComplete:             "SecModeComplete",
	MsgTypeSecModeRej:                  "SecModeRej",
	MsgTypeStatus5GMM:                  "Status5GMM",
	MsgTypeNotif:                       "Notif",
	MsgTypeNotifRsp:                    "NotifRsp",
	MsgTypeULNASTransport:              "ULNASTransport",
	MsgTypeDLNASTransport:              "DLNASTransport",
	MsgTypeRelayKeyReq:                 "RelayKeyReq",
	MsgTypeRelayKeyAccept:              "RelayKeyAccept",
	MsgTypeRelayKeyRej:                 "RelayKeyRej",
	MsgTypeRelayAuthReq:                "RelayAuthReq",
	MsgTypeRelayAuthRsp:                "RelayAuthRsp",
	MsgTypePDUSessEstReq:               "PDUSessEstReq",
	MsgTypePDUSessEstAccept:            "PDUSessEstAccept",
	MsgTypePDUSessEstRej:               "PDUSessEstRej",
	MsgTypePDUSessAuthCmd:              "PDUSessAuthCmd",
	MsgTypePDUSessAuthComplete:         "PDUSessAuthComplete",
	MsgTypePDUSessAuthResult:           "PDUSessAuthResult",
	MsgTypePDUSessModReq:               "PDUSessModReq",
	MsgTypePDUSessModRej:               "PDUSessModRej",
	MsgTypePDUSessModCmd:               "PDUSessModCmd",
	MsgTypePDUSessModComplete:          "PDUSessModComplete",
	MsgTypePDUSessModCmdRej:            "PDUSessModCmdRej",
	MsgTypePDUSessRelReq:               "PDUSessRelReq",
	MsgTypePDUSessRelRej:               "PDUSessRelRej",
	MsgTypePDUSessRelCmd:               "PDUSessRelCmd",
	MsgTypePDUSessRelComplete:          "PDUSessRelComplete",
	MsgTypeStatus5GSM:                  "Status5GSM",
	MsgTypeSvcLvlAuthCmd:               "SvcLvlAuthCmd",
	MsgTypeSvcLvlAuthComplete:          "SvcLvlAuthComplete",
	MsgTypeRemoteUEReport:              "RemoteUEReport",
	MsgTypeRemoteUEReportRsp:           "RemoteUEReportRsp",
}

func (t MsgType) String() string {
	return msgTypeMap[t]
}

const (
	GmmHdrLen uint8 = 3
	GsmHdrLen uint8 = 4
)

type IELengthType uint8

const (
	IELen8Bits  IELengthType = 1
	IELen16Bits IELengthType = 2
)

func getEpd(b []byte) (Epd, error) {
	if len(b) == 0 {
		return 0, errors.Errorf("getEpd(): len(b) is 0")
	}
	return Epd(b[0]), nil
}

// will return 0xff if len(b) < 2
func GetSecHdrType(b []byte) SecHdrType {
	if len(b) < 2 {
		return SecHdrTypeErr
	}
	return SecHdrType(b[1] & 0x0f)
}

func GetHdrInfo(b []byte) string {
	if len(b) == 0 {
		return "no nas msg"
	}
	secHdrType := GetSecHdrType(b)
	if secHdrType == SecHdrTypePlainNas || secHdrType == SecHdrTypeErr {
		return secHdrType.String()
	}
	return fmt.Sprintf("%v, SEQ[%d]", secHdrType, b[6])
}

type Message interface {
	ExtendedProtocolDiscriminator() Epd
	MsgType() MsgType

	MarshalBinary() ([]byte, error)
	UnmarshalBinary([]byte) error
}

type GSMMessage interface {
	Message

	PDUSessionID() uint8
	ProcedureTransactionID() uint8
}

type MACFailure struct {
	Expected []byte
	Received []byte
}

type Error struct {
	MACFailure *MACFailure
	IEToDoList []byte
}

func (e *Error) Error() string {
	var errs []string
	if e.MACFailure != nil {
		errs = append(errs,
			fmt.Sprintf("MAC err, (got: 0x%08x, want: 0x%08x)",
				e.MACFailure.Received, e.MACFailure.Expected))
	}
	if len(e.IEToDoList) > 0 {
		errs = append(errs,
			fmt.Sprintf("IE ToDo List: %q", e.IEToDoList))
	}
	return strings.Join(errs, ";")
}

func extractIEI(ieiByte uint8) uint8 {
	var tmpIeiN uint8
	if ieiByte >= 0x80 {
		tmpIeiN = (ieiByte & 0xf0)
	} else {
		tmpIeiN = ieiByte
	}
	return tmpIeiN
}

func getIeLen(reader *bytes.Buffer, octets IELengthType) (uint16, error) {
	ieLen := reader.Next(int(octets))
	if len(ieLen) != int(octets) {
		return 0, errors.Errorf("data is not enough")
	}
	var length uint16
	if octets == IELen8Bits {
		length = uint16(ieLen[0])
	} else {
		length = binary.BigEndian.Uint16(ieLen)
	}
	if int(length) > reader.Len() {
		return 0, errors.Errorf("IE length %d exceeds remaining data %d", length, reader.Len())
	}
	return length, nil
}

func Marshal(m Message, sc *SecCtx, st SecHdrType) ([]byte, error) {
	if st == SecHdrTypePlainNas {
		if sc != nil {
			return nil, errors.Errorf(
				"%s marshal with plain text, but SecCtx exists",
				m.MsgType().String())
		}
		return m.MarshalBinary()
	}

	if sc == nil {
		return nil, errors.Errorf("%s Marshal() SecCtx should not be nil",
			m.MsgType().String())
	}

	h := new(SecProtectedHdr)
	h.SecHdrType = st

	var direction Direction
	if sc.Side == CoreNetworkSide {
		direction = DirectionDownlink
	} else {
		direction = DirectionUplink
	}

	ciphered := false
	switch h.SecHdrType {
	case SecHdrTypeIntegrityProtected:
	case SecHdrTypeIntegrityProtectedAndCiphered:
		ciphered = true
	case SecHdrTypeIntegrityProtectedWithNew5gNasSecCtx:
	case SecHdrTypeIntegrityProtectedAndCipheredWithNew5gNasSecCtx:
		ciphered = true
		sc.CountReset(direction)
	default:
		return nil, errors.Errorf("%s Marshal() bad Sec Hdr type: 0x%0x",
			m.MsgType().String(), h.SecHdrType)
	}

	innerMsg, err := m.MarshalBinary()
	if err != nil {
		return nil, err
	}

	if ciphered {
		if innerMsg, err = sc.NASEncrypt(direction, innerMsg); err != nil {
			return nil, errors.Wrapf(err, "%s Marshal() NASEncrypt",
				m.MsgType().String())
		}
	}

	h.SequenceNumber = sc.GetCountSQN(direction)
	// count.AddOne()
	SeqAndInnerMsg := []byte{h.SequenceNumber}
	SeqAndInnerMsg = append(SeqAndInnerMsg, innerMsg...)
	h.MAC, err = sc.NASMacCalculate(direction, SeqAndInnerMsg)
	if err != nil {
		return nil, errors.Wrapf(err, "%s Marshal() NASMacCalculate",
			m.MsgType().String())
	}

	secMsg, err := h.MarshalBinary()
	if err != nil {
		return nil, errors.Wrapf(err, "%s Marshal() SecHdr err",
			m.MsgType().String())
	}
	secMsg = append(secMsg, innerMsg...)

	sc.CountAddOne(direction)
	return secMsg, nil
}

func Parse(b []byte, sc *SecCtx) (m Message, err error) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			err = errors.Errorf("Parse(): panic: %s", debug.Stack())
		}
	}()

	bufLen := len(b)
	if bufLen < 3 {
		return nil, errors.Errorf("Parse() NAS Msg length < 3")
	}

	epd, err := getEpd(b)
	if err != nil {
		return nil, errors.Wrap(err, "Parse()")
	}

	if epd == Epd5GSSessMgmtMsg || GetSecHdrType(b) == SecHdrTypePlainNas {
		return parsePlainMsg(b)
	}

	h := new(SecProtectedHdr)
	if err = h.UnmarshalBinary(b); err != nil {
		return nil, errors.Wrap(err, "Parse() SecHdr err")
	}

	var count *Count
	var direction Direction

	toDecrypt := false
	toCheckMAC := false
	resetCounter := false
	switch h.SecHdrType {
	case SecHdrTypeIntegrityProtected:
		toCheckMAC = true
	case SecHdrTypeIntegrityProtectedAndCiphered:
		toCheckMAC = true
		toDecrypt = true
	case SecHdrTypeIntegrityProtectedWithNew5gNasSecCtx:
		toCheckMAC = true
	case SecHdrTypeIntegrityProtectedAndCipheredWithNew5gNasSecCtx:
		toCheckMAC = true
		toDecrypt = true
		resetCounter = true
	default:
		return nil, errors.Errorf("Parse() bad SecHdr type: 0x%0x", h.SecHdrType)
	}

	var macFail *MACFailure
	var oldSqn uint8
	var oldOverflow uint16
	if sc == nil {
		// unable to calc MAC nor to decrypt; try to parse the msg as plaintext
		toCheckMAC = false
		macFail = &MACFailure{
			Expected: []byte{0x00, 0x00, 0x00, 0x00},
			Received: h.MAC,
		}
		toDecrypt = false
	} else {
		if sc.Side == CoreNetworkSide {
			count = sc.UplinkCount
			direction = DirectionUplink
		} else {
			count = sc.DownlinkCount
			direction = DirectionDownlink
		}
		// store count for recovery if MAC failed later
		oldSqn = count.SQN()
		oldOverflow = count.Overflow()

		if resetCounter {
			count.Set(0, 0)
		}
		if count.SQN() > h.SequenceNumber {
			count.SetOverflow(count.Overflow() + 1)
		}
		count.SetSQN(h.SequenceNumber)

		// TS 33.501 D.1: All processing performed in association with integrity (except for replay protection)
		// shall be exactly the same as with any of the integrity algorithms specified in this annex
		// except that the receiver does not check the received MAC.
		if sc.IntegrityAlg == AlgIntegrity128NIA0 {
			toCheckMAC = false
		}
	}

	// Verify the MAC of security protected header
	// TS 24.501: the integrity protection shall include octet 7 to n, i.e.
	// the Sequence number IE and the NAS message IE
	SeqAndInnerMsg := b[6:]

	if toCheckMAC {
		var mac32 []byte
		mac32, err = sc.NASMacCalculate(direction, SeqAndInnerMsg)
		if err != nil {
			return nil, errors.Wrap(err, "Parse() NASMacCalculate")
		}
		if !reflect.DeepEqual(mac32, h.MAC) {
			macFail = &MACFailure{
				Expected: mac32,
				Received: h.MAC,
			}

			// recover secCtx since mac failed
			count.Set(oldOverflow, oldSqn)
		}
	}

	innerMsg := SeqAndInnerMsg[1:]
	if toDecrypt {
		if innerMsg, err = sc.NASEncrypt(direction, innerMsg); err != nil {
			return nil, errors.Wrap(err, "Parse() NASEncrypt")
		}
	}

	msg, err := parsePlainMsg(innerMsg)
	if err != nil {
		e, ok := err.(*Error)
		if ok && macFail != nil {
			e.MACFailure = macFail
		}
		return nil, errors.Wrap(err, "Parse() parsePlainMsg")
	}

	if macFail != nil {
		return msg, &Error{MACFailure: macFail}
	}
	return msg, nil
}

func parsePlainMsg(b []byte) (Message, error) {
	epd, err := getEpd(b)
	if err != nil {
		return nil, errors.Wrap(err, "parsePlainMsg()")
	}
	switch epd {
	case Epd5GSMobilityMgmtMsg:
		return ParseGMM(b)
	case Epd5GSSessMgmtMsg:
		return ParseGSM(b)
	default:
		return nil, errors.Errorf("unknown epd type: %d", epd)
	}
}

func ParseGMM(b []byte) (m Message, err error) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			err = errors.Errorf("ParseGMM(): panic: %s", debug.Stack())
		}
	}()
	if len(b) < int(GmmHdrLen) {
		return nil, errors.Errorf("GMM length < %d", GmmHdrLen)
	}
	msgType := b[2]
	switch MsgType(msgType) {
	case MsgTypeRegReq:
		m = &RegReq{}
	case MsgTypeRegAccept:
		m = &RegAccept{}
	case MsgTypeRegComplete:
		m = &RegComplete{}
	case MsgTypeRegRej:
		m = &RegRej{}
	case MsgTypeDeregReqUEOrig:
		m = &DeregReqUEOrig{}
	case MsgTypeDeregAcceptUEOrig:
		m = &DeregAcceptUEOrig{}
	case MsgTypeDeregReqUETerm:
		m = &DeregReqUETerm{}
	case MsgTypeDeregAcceptUETerm:
		m = &DeregAcceptUETerm{}
	case MsgTypeSvcReq:
		m = &SvcReq{}
	case MsgTypeSvcRej:
		m = &SvcRej{}
	case MsgTypeSvcAccept:
		m = &SvcAccept{}
	case MsgTypeCfgUpdateCmd:
		m = &CfgUpdateCmd{}
	case MsgTypeCfgUpdateComplete:
		m = &CfgUpdateComplete{}
	case MsgTypeAuthReq:
		m = &AuthReq{}
	case MsgTypeAuthRsp:
		m = &AuthRsp{}
	case MsgTypeAuthRej:
		m = &AuthRej{}
	case MsgTypeAuthFailure:
		m = &AuthFailure{}
	case MsgTypeAuthResult:
		m = &AuthResult{}
	case MsgTypeIdReq:
		m = &IdReq{}
	case MsgTypeIdRsp:
		m = &IdRsp{}
	case MsgTypeSecModeCmd:
		m = &SecModeCmd{}
	case MsgTypeSecModeComplete:
		m = &SecModeComplete{}
	case MsgTypeSecModeRej:
		m = &SecModeRej{}
	case MsgTypeStatus5GMM:
		m = &Status5GMM{}
	case MsgTypeNotif:
		m = &Notif{}
	case MsgTypeNotifRsp:
		m = &NotifRsp{}
	case MsgTypeULNASTransport:
		m = &ULNASTransport{}
	case MsgTypeDLNASTransport:
		m = &DLNASTransport{}
	case MsgTypeRelayKeyReq:
		m = &RelayKeyReq{}
	case MsgTypeRelayKeyAccept:
		m = &RelayKeyAccept{}
	case MsgTypeRelayKeyRej:
		m = &RelayKeyRej{}
	case MsgTypeRelayAuthReq:
		m = &RelayAuthReq{}
	case MsgTypeRelayAuthRsp:
		m = &RelayAuthRsp{}
	default:
		return nil, errors.Errorf("unknown GMM Msg type: %d (0x%x)", msgType, msgType)
	}

	if err := m.UnmarshalBinary(b); err != nil {
		if umerr, ok := err.(*Error); ok {
			if len(umerr.IEToDoList) > 0 {
				return m, umerr
			}
		}
		// IEI Unknown (thus IE length unknown), or other msg unmarshl errors
		return nil, errors.Wrap(err, "ParseGMM() m.UnmarshalBinary()")
	}
	// no error
	return m, nil
}

func ParseGSM(b []byte) (m GSMMessage, err error) {
	defer func() {
		if p := recover(); p != nil {
			// Print stack for panic to log. Fatalf() will let program exit.
			err = errors.Errorf("ParseGSM(): panic: %s", debug.Stack())
		}
	}()
	if len(b) < int(GsmHdrLen) {
		return nil, errors.Errorf("GSM length < %d", GsmHdrLen)
	}

	msgType := b[3]
	switch MsgType(msgType) {
	case MsgTypePDUSessEstReq:
		m = &PDUSessEstReq{}
	case MsgTypePDUSessEstAccept:
		m = &PDUSessEstAccept{}
	case MsgTypePDUSessEstRej:
		m = &PDUSessEstRej{}
	case MsgTypePDUSessAuthCmd:
		m = &PDUSessAuthCmd{}
	case MsgTypePDUSessAuthComplete:
		m = &PDUSessAuthComplete{}
	case MsgTypePDUSessAuthResult:
		m = &PDUSessAuthResult{}
	case MsgTypePDUSessModReq:
		m = &PDUSessModReq{}
	case MsgTypePDUSessModRej:
		m = &PDUSessModRej{}
	case MsgTypePDUSessModCmd:
		m = &PDUSessModCmd{}
	case MsgTypePDUSessModComplete:
		m = &PDUSessModComplete{}
	case MsgTypePDUSessModCmdRej:
		m = &PDUSessModCmdRej{}
	case MsgTypePDUSessRelReq:
		m = &PDUSessRelReq{}
	case MsgTypePDUSessRelRej:
		m = &PDUSessRelRej{}
	case MsgTypePDUSessRelCmd:
		m = &PDUSessRelCmd{}
	case MsgTypePDUSessRelComplete:
		m = &PDUSessRelComplete{}
	case MsgTypeStatus5GSM:
		m = &Status5GSM{}
	case MsgTypeSvcLvlAuthCmd:
		m = &SvcLvlAuthCmd{}
	case MsgTypeSvcLvlAuthComplete:
		m = &SvcLvlAuthComplete{}
	case MsgTypeRemoteUEReport:
		m = &RemoteUEReport{}
	case MsgTypeRemoteUEReportRsp:
		m = &RemoteUEReportRsp{}
	default:
		return nil, errors.Errorf("unknown GSM Msg type: %d", msgType)
	}

	if err := m.UnmarshalBinary(b); err != nil {
		if umerr, ok := err.(*Error); ok {
			if len(umerr.IEToDoList) > 0 {
				return m, umerr
			}
		}
		// IEI Unknown (thus IE length unknown), or other msg unmarshl errors
		return nil, errors.Wrap(err, "ParseGSM()")
	}
	// no error
	return m, nil
}
