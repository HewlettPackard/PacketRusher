package nas

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/lib/nas/nasMessage"
	"os"
)

var NasLogFile *os.File

// message TODO：description
type Message struct {
	SecurityHeader
	*GmmMessage
	*GsmMessage
}

// SecurityHeader TODO：description
type SecurityHeader struct {
	ProtocolDiscriminator     uint8
	SecurityHeaderType        uint8
	MessageAuthenticationCode uint32
	SequenceNumber            uint8
}

const (
	SecurityHeaderTypePlainNas                                                 uint8 = 0x00
	SecurityHeaderTypeIntegrityProtected                                       uint8 = 0x01
	SecurityHeaderTypeIntegrityProtectedAndCiphered                            uint8 = 0x02
	SecurityHeaderTypeIntegrityProtectedWithNew5gNasSecurityContext            uint8 = 0x03
	SecurityHeaderTypeIntegrityProtectedAndCipheredWithNew5gNasSecurityContext uint8 = 0x04
)

// NewMessage TODO:desc
func NewMessage() *Message {
	Message := &Message{}
	return Message
}

// NewGmmMessage TODO:desc
func NewGmmMessage() *GmmMessage {
	GmmMessage := &GmmMessage{}
	return GmmMessage
}

// NewGmmMessage TODO:desc
func NewGsmMessage() *GsmMessage {
	GsmMessage := &GsmMessage{}
	return GsmMessage
}

// GmmHeader Octet1 protocolDiscriminator securityHeaderType
//           Octet2 MessageType
type GmmHeader struct {
	Octet [3]uint8
}

type GsmHeader struct {
	Octet [4]uint8
}

// GetMessageType 9.8
func (a *GmmHeader) GetMessageType() (messageType uint8) {
	messageType = a.Octet[2]
	return messageType
}

// GetMessageType 9.8
func (a *GmmHeader) SetMessageType(messageType uint8) {
	a.Octet[2] = messageType
}

func (a *GmmHeader) GetExtendedProtocolDiscriminator() uint8 {
	return a.Octet[0]
}

func (a *GmmHeader) SetExtendedProtocolDiscriminator(epd uint8) {
	a.Octet[0] = epd
}

func (a *GsmHeader) GetExtendedProtocolDiscriminator() uint8 {
	return a.Octet[0]
}

func (a *GsmHeader) SetExtendedProtocolDiscriminator(epd uint8) {
	a.Octet[0] = epd
}

// GetMessageType 9.8
func (a *GsmHeader) GetMessageType() (messageType uint8) {
	messageType = a.Octet[3]
	return messageType
}

// GetMessageType 9.8
func (a *GsmHeader) SetMessageType(messageType uint8) {
	a.Octet[3] = messageType
}

func GetEPD(byteArray []byte) uint8 {
	return byteArray[0]
}

func GetSecurityHeaderType(byteArray []byte) uint8 {
	return byteArray[1]
}

type GmmMessage struct {
	GmmHeader
	*nasMessage.AuthenticationRequest                            //8.2.1
	*nasMessage.AuthenticationResponse                           //8.2.2
	*nasMessage.AuthenticationResult                             //8.2.3
	*nasMessage.AuthenticationFailure                            //8.2.4
	*nasMessage.AuthenticationReject                             //8.2.5
	*nasMessage.RegistrationRequest                              //8.2.6
	*nasMessage.RegistrationAccept                               //8.2.7
	*nasMessage.RegistrationComplete                             //8.2.8
	*nasMessage.RegistrationReject                               //8.2.9
	*nasMessage.ULNASTransport                                   //8.2.10
	*nasMessage.DLNASTransport                                   //8.2.11
	*nasMessage.DeregistrationRequestUEOriginatingDeregistration //8.2.12
	*nasMessage.DeregistrationAcceptUEOriginatingDeregistration  //8.2.13
	*nasMessage.DeregistrationRequestUETerminatedDeregistration  //8.2.14
	*nasMessage.DeregistrationAcceptUETerminatedDeregistration   //8.2.15
	*nasMessage.ServiceRequest                                   //8.2.16
	*nasMessage.ServiceAccept                                    //8.2.17
	*nasMessage.ServiceReject                                    //8.2.18
	*nasMessage.ConfigurationUpdateCommand                       //8.2.19
	*nasMessage.ConfigurationUpdateComplete                      //8.2.20
	*nasMessage.IdentityRequest                                  //8.2.21
	*nasMessage.IdentityResponse                                 //8.2.22
	*nasMessage.Notification                                     //8.2.23
	*nasMessage.NotificationResponse                             //8.2.24
	*nasMessage.SecurityModeCommand                              //8.2.25
	*nasMessage.SecurityModeComplete                             //8.2.26
	*nasMessage.SecurityModeReject                               //8.2.27
	*nasMessage.SecurityProtected5GSNASMessage                   //8.2.28
	*nasMessage.Status5GMM                                       //8.2.29
}

const (
	MsgTypeRegistrationRequest                              uint8 = 65
	MsgTypeRegistrationAccept                               uint8 = 66
	MsgTypeRegistrationComplete                             uint8 = 67
	MsgTypeRegistrationReject                               uint8 = 68
	MsgTypeDeregistrationRequestUEOriginatingDeregistration uint8 = 69
	MsgTypeDeregistrationAcceptUEOriginatingDeregistration  uint8 = 70
	MsgTypeDeregistrationRequestUETerminatedDeregistration  uint8 = 71
	MsgTypeDeregistrationAcceptUETerminatedDeregistration   uint8 = 72
	MsgTypeServiceRequest                                   uint8 = 76
	MsgTypeServiceReject                                    uint8 = 77
	MsgTypeServiceAccept                                    uint8 = 78
	MsgTypeConfigurationUpdateCommand                       uint8 = 84
	MsgTypeConfigurationUpdateComplete                      uint8 = 85
	MsgTypeAuthenticationRequest                            uint8 = 86
	MsgTypeAuthenticationResponse                           uint8 = 87
	MsgTypeAuthenticationReject                             uint8 = 88
	MsgTypeAuthenticationFailure                            uint8 = 89
	MsgTypeAuthenticationResult                             uint8 = 90
	MsgTypeIdentityRequest                                  uint8 = 91
	MsgTypeIdentityResponse                                 uint8 = 92
	MsgTypeSecurityModeCommand                              uint8 = 93
	MsgTypeSecurityModeComplete                             uint8 = 94
	MsgTypeSecurityModeReject                               uint8 = 95
	MsgTypeStatus5GMM                                       uint8 = 100
	MsgTypeNotification                                     uint8 = 101
	MsgTypeNotificationResponse                             uint8 = 102
	MsgTypeULNASTransport                                   uint8 = 103
	MsgTypeDLNASTransport                                   uint8 = 104
)

func (a *Message) PlainNasDecode(byteArray *[]byte) error {
	epd := GetEPD(*byteArray)
	switch epd {
	case nasMessage.Epd5GSMobilityManagementMessage:
		return a.GmmMessageDecode(byteArray)
	case nasMessage.Epd5GSSessionManagementMessage:
		return a.GsmMessageDecode(byteArray)
	}
	return fmt.Errorf("Extended Protocol Discriminator[%d] is not allowed in Nas message Deocde", epd)
}
func (a *Message) PlainNasEncode() ([]byte, error) {
	data := new(bytes.Buffer)
	if a.GmmMessage != nil {
		err := a.GmmMessageEncode(data)
		return data.Bytes(), err
	} else if a.GsmMessage != nil {
		err := a.GsmMessageEncode(data)
		return data.Bytes(), err
	}
	return nil, fmt.Errorf("Gmm/Gsm message are both empty in Nas message Encode")

}

func (a *Message) GmmMessageDecode(byteArray *[]byte) error {
	buffer := bytes.NewBuffer(*byteArray)
	a.GmmMessage = NewGmmMessage()
	binary.Read(buffer, binary.BigEndian, &a.GmmMessage.GmmHeader)

	log.Debug(spew.Sdump(a))

	switch a.GmmMessage.GmmHeader.GetMessageType() {
	case MsgTypeRegistrationRequest:
		a.GmmMessage.RegistrationRequest = nasMessage.NewRegistrationRequest(MsgTypeRegistrationRequest)
		a.GmmMessage.DecodeRegistrationRequest(byteArray)
	case MsgTypeRegistrationAccept:
		a.GmmMessage.RegistrationAccept = nasMessage.NewRegistrationAccept(MsgTypeRegistrationAccept)
		a.GmmMessage.DecodeRegistrationAccept(byteArray)
	case MsgTypeRegistrationComplete:
		a.GmmMessage.RegistrationComplete = nasMessage.NewRegistrationComplete(MsgTypeRegistrationComplete)
		a.GmmMessage.DecodeRegistrationComplete(byteArray)
	case MsgTypeRegistrationReject:
		a.GmmMessage.RegistrationReject = nasMessage.NewRegistrationReject(MsgTypeRegistrationReject)
		a.GmmMessage.DecodeRegistrationReject(byteArray)
	case MsgTypeDeregistrationRequestUEOriginatingDeregistration:
		a.GmmMessage.DeregistrationRequestUEOriginatingDeregistration = nasMessage.NewDeregistrationRequestUEOriginatingDeregistration(MsgTypeDeregistrationRequestUEOriginatingDeregistration)
		a.GmmMessage.DecodeDeregistrationRequestUEOriginatingDeregistration(byteArray)
	case MsgTypeDeregistrationAcceptUEOriginatingDeregistration:
		a.GmmMessage.DeregistrationAcceptUEOriginatingDeregistration = nasMessage.NewDeregistrationAcceptUEOriginatingDeregistration(MsgTypeDeregistrationAcceptUEOriginatingDeregistration)
		a.GmmMessage.DecodeDeregistrationAcceptUEOriginatingDeregistration(byteArray)
	case MsgTypeDeregistrationRequestUETerminatedDeregistration:
		a.GmmMessage.DeregistrationRequestUETerminatedDeregistration = nasMessage.NewDeregistrationRequestUETerminatedDeregistration(MsgTypeDeregistrationRequestUETerminatedDeregistration)
		a.GmmMessage.DecodeDeregistrationRequestUETerminatedDeregistration(byteArray)
	case MsgTypeDeregistrationAcceptUETerminatedDeregistration:
		a.GmmMessage.DeregistrationAcceptUETerminatedDeregistration = nasMessage.NewDeregistrationAcceptUETerminatedDeregistration(MsgTypeDeregistrationAcceptUETerminatedDeregistration)
		a.GmmMessage.DecodeDeregistrationAcceptUETerminatedDeregistration(byteArray)
	case MsgTypeServiceRequest:
		a.GmmMessage.ServiceRequest = nasMessage.NewServiceRequest(MsgTypeServiceRequest)
		a.GmmMessage.DecodeServiceRequest(byteArray)
	case MsgTypeServiceReject:
		a.GmmMessage.ServiceReject = nasMessage.NewServiceReject(MsgTypeServiceReject)
		a.GmmMessage.DecodeServiceReject(byteArray)
	case MsgTypeServiceAccept:
		a.GmmMessage.ServiceAccept = nasMessage.NewServiceAccept(MsgTypeServiceAccept)
		a.GmmMessage.DecodeServiceAccept(byteArray)
	case MsgTypeConfigurationUpdateCommand:
		a.GmmMessage.ConfigurationUpdateCommand = nasMessage.NewConfigurationUpdateCommand(MsgTypeConfigurationUpdateCommand)
		a.GmmMessage.DecodeConfigurationUpdateCommand(byteArray)
	case MsgTypeConfigurationUpdateComplete:
		a.GmmMessage.ConfigurationUpdateComplete = nasMessage.NewConfigurationUpdateComplete(MsgTypeConfigurationUpdateComplete)
		a.GmmMessage.DecodeConfigurationUpdateComplete(byteArray)
	case MsgTypeAuthenticationRequest:
		a.GmmMessage.AuthenticationRequest = nasMessage.NewAuthenticationRequest(MsgTypeAuthenticationRequest)
		a.GmmMessage.DecodeAuthenticationRequest(byteArray)
	case MsgTypeAuthenticationResponse:
		a.GmmMessage.AuthenticationResponse = nasMessage.NewAuthenticationResponse(MsgTypeAuthenticationResponse)
		a.GmmMessage.DecodeAuthenticationResponse(byteArray)
	case MsgTypeAuthenticationReject:
		a.GmmMessage.AuthenticationReject = nasMessage.NewAuthenticationReject(MsgTypeAuthenticationReject)
		a.GmmMessage.DecodeAuthenticationReject(byteArray)
	case MsgTypeAuthenticationFailure:
		a.GmmMessage.AuthenticationFailure = nasMessage.NewAuthenticationFailure(MsgTypeAuthenticationFailure)
		a.GmmMessage.DecodeAuthenticationFailure(byteArray)
	case MsgTypeAuthenticationResult:
		a.GmmMessage.AuthenticationResult = nasMessage.NewAuthenticationResult(MsgTypeAuthenticationResult)
		a.GmmMessage.DecodeAuthenticationResult(byteArray)
	case MsgTypeIdentityRequest:
		a.GmmMessage.IdentityRequest = nasMessage.NewIdentityRequest(MsgTypeIdentityRequest)
		a.GmmMessage.DecodeIdentityRequest(byteArray)
	case MsgTypeIdentityResponse:
		a.GmmMessage.IdentityResponse = nasMessage.NewIdentityResponse(MsgTypeIdentityResponse)
		a.GmmMessage.DecodeIdentityResponse(byteArray)
	case MsgTypeSecurityModeCommand:
		a.GmmMessage.SecurityModeCommand = nasMessage.NewSecurityModeCommand(MsgTypeSecurityModeCommand)
		a.GmmMessage.DecodeSecurityModeCommand(byteArray)
	case MsgTypeSecurityModeComplete:
		a.GmmMessage.SecurityModeComplete = nasMessage.NewSecurityModeComplete(MsgTypeSecurityModeComplete)
		a.GmmMessage.DecodeSecurityModeComplete(byteArray)
	case MsgTypeSecurityModeReject:
		a.GmmMessage.SecurityModeReject = nasMessage.NewSecurityModeReject(MsgTypeSecurityModeReject)
		a.GmmMessage.DecodeSecurityModeReject(byteArray)
	case MsgTypeStatus5GMM:
		a.GmmMessage.Status5GMM = nasMessage.NewStatus5GMM(MsgTypeStatus5GMM)
		a.GmmMessage.DecodeStatus5GMM(byteArray)
	case MsgTypeNotification:
		a.GmmMessage.Notification = nasMessage.NewNotification(MsgTypeNotification)
		a.GmmMessage.DecodeNotification(byteArray)
	case MsgTypeNotificationResponse:
		a.GmmMessage.NotificationResponse = nasMessage.NewNotificationResponse(MsgTypeNotificationResponse)
		a.GmmMessage.DecodeNotificationResponse(byteArray)
	case MsgTypeULNASTransport:
		a.GmmMessage.ULNASTransport = nasMessage.NewULNASTransport(MsgTypeULNASTransport)
		a.GmmMessage.DecodeULNASTransport(byteArray)
	case MsgTypeDLNASTransport:
		a.GmmMessage.DLNASTransport = nasMessage.NewDLNASTransport(MsgTypeDLNASTransport)
		a.GmmMessage.DecodeDLNASTransport(byteArray)
	default:
		return fmt.Errorf("NAS decode Fail: MsgType[%d] doesn't exist in GMM message", a.GmmMessage.GmmHeader.GetMessageType())
	}
	return nil
}

func (a *Message) GmmMessageEncode(buffer *bytes.Buffer) error {

	log.Debug(spew.Sdump(a))

	switch a.GmmMessage.GmmHeader.GetMessageType() {
	case MsgTypeRegistrationRequest:
		a.GmmMessage.EncodeRegistrationRequest(buffer)
	case MsgTypeRegistrationAccept:
		a.GmmMessage.EncodeRegistrationAccept(buffer)
	case MsgTypeRegistrationComplete:
		a.GmmMessage.EncodeRegistrationComplete(buffer)
	case MsgTypeRegistrationReject:
		a.GmmMessage.EncodeRegistrationReject(buffer)
	case MsgTypeDeregistrationRequestUEOriginatingDeregistration:
		a.GmmMessage.EncodeDeregistrationRequestUEOriginatingDeregistration(buffer)
	case MsgTypeDeregistrationAcceptUEOriginatingDeregistration:
		a.GmmMessage.EncodeDeregistrationAcceptUEOriginatingDeregistration(buffer)
	case MsgTypeDeregistrationRequestUETerminatedDeregistration:
		a.GmmMessage.EncodeDeregistrationRequestUETerminatedDeregistration(buffer)
	case MsgTypeDeregistrationAcceptUETerminatedDeregistration:
		a.GmmMessage.EncodeDeregistrationAcceptUETerminatedDeregistration(buffer)
	case MsgTypeServiceRequest:
		a.GmmMessage.EncodeServiceRequest(buffer)
	case MsgTypeServiceReject:
		a.GmmMessage.EncodeServiceReject(buffer)
	case MsgTypeServiceAccept:
		a.GmmMessage.EncodeServiceAccept(buffer)
	case MsgTypeConfigurationUpdateCommand:
		a.GmmMessage.EncodeConfigurationUpdateCommand(buffer)
	case MsgTypeConfigurationUpdateComplete:
		a.GmmMessage.EncodeConfigurationUpdateComplete(buffer)
	case MsgTypeAuthenticationRequest:
		a.GmmMessage.EncodeAuthenticationRequest(buffer)
	case MsgTypeAuthenticationResponse:
		a.GmmMessage.EncodeAuthenticationResponse(buffer)
	case MsgTypeAuthenticationReject:
		a.GmmMessage.EncodeAuthenticationReject(buffer)
	case MsgTypeAuthenticationFailure:
		a.GmmMessage.EncodeAuthenticationFailure(buffer)
	case MsgTypeAuthenticationResult:
		a.GmmMessage.EncodeAuthenticationResult(buffer)
	case MsgTypeIdentityRequest:
		a.GmmMessage.EncodeIdentityRequest(buffer)
	case MsgTypeIdentityResponse:
		a.GmmMessage.EncodeIdentityResponse(buffer)
	case MsgTypeSecurityModeCommand:
		a.GmmMessage.EncodeSecurityModeCommand(buffer)
	case MsgTypeSecurityModeComplete:
		a.GmmMessage.EncodeSecurityModeComplete(buffer)
	case MsgTypeSecurityModeReject:
		a.GmmMessage.EncodeSecurityModeReject(buffer)
	case MsgTypeStatus5GMM:
		a.GmmMessage.EncodeStatus5GMM(buffer)
	case MsgTypeNotification:
		a.GmmMessage.EncodeNotification(buffer)
	case MsgTypeNotificationResponse:
		a.GmmMessage.EncodeNotificationResponse(buffer)
	case MsgTypeULNASTransport:
		a.GmmMessage.EncodeULNASTransport(buffer)
	case MsgTypeDLNASTransport:
		a.GmmMessage.EncodeDLNASTransport(buffer)
	default:
		return fmt.Errorf("NAS Encode Fail: MsgType[%d] doesn't exist in GMM message", a.GmmMessage.GmmHeader.GetMessageType())
	}

	return nil
}

type GsmMessage struct {
	GsmHeader
	*nasMessage.PDUSessionEstablishmentRequest      //8.3.1
	*nasMessage.PDUSessionEstablishmentAccept       //8.3.2
	*nasMessage.PDUSessionEstablishmentReject       //8.3.3
	*nasMessage.PDUSessionAuthenticationCommand     //8.3.4
	*nasMessage.PDUSessionAuthenticationComplete    //8.3.5
	*nasMessage.PDUSessionAuthenticationResult      //8.3.6
	*nasMessage.PDUSessionModificationRequest       //8.3.7
	*nasMessage.PDUSessionModificationReject        //8.3.8
	*nasMessage.PDUSessionModificationCommand       //8.3.9
	*nasMessage.PDUSessionModificationComplete      //8.3.10
	*nasMessage.PDUSessionModificationCommandReject //8.3.11
	*nasMessage.PDUSessionReleaseRequest            //8.3.12
	*nasMessage.PDUSessionReleaseReject             //8.3.13
	*nasMessage.PDUSessionReleaseCommand            //8.3.14
	*nasMessage.PDUSessionReleaseComplete           //8.3.15
	*nasMessage.Status5GSM                          //8.3.16
}

const (
	MsgTypePDUSessionEstablishmentRequest      uint8 = 193
	MsgTypePDUSessionEstablishmentAccept       uint8 = 194
	MsgTypePDUSessionEstablishmentReject       uint8 = 195
	MsgTypePDUSessionAuthenticationCommand     uint8 = 197
	MsgTypePDUSessionAuthenticationComplete    uint8 = 198
	MsgTypePDUSessionAuthenticationResult      uint8 = 199
	MsgTypePDUSessionModificationRequest       uint8 = 201
	MsgTypePDUSessionModificationReject        uint8 = 202
	MsgTypePDUSessionModificationCommand       uint8 = 203
	MsgTypePDUSessionModificationComplete      uint8 = 204
	MsgTypePDUSessionModificationCommandReject uint8 = 205
	MsgTypePDUSessionReleaseRequest            uint8 = 209
	MsgTypePDUSessionReleaseReject             uint8 = 210
	MsgTypePDUSessionReleaseCommand            uint8 = 211
	MsgTypePDUSessionReleaseComplete           uint8 = 212
	MsgTypeStatus5GSM                          uint8 = 214
)

func (a *Message) GsmMessageDecode(byteArray *[]byte) error {
	buffer := bytes.NewBuffer(*byteArray)
	a.GsmMessage = NewGsmMessage()
	binary.Read(buffer, binary.BigEndian, &a.GsmMessage.GsmHeader)
	log.Debug(spew.Sdump(a))

	switch a.GsmMessage.GsmHeader.GetMessageType() {
	case MsgTypePDUSessionEstablishmentRequest:
		a.GsmMessage.PDUSessionEstablishmentRequest = nasMessage.NewPDUSessionEstablishmentRequest(MsgTypePDUSessionEstablishmentRequest)
		a.GsmMessage.DecodePDUSessionEstablishmentRequest(byteArray)
	case MsgTypePDUSessionEstablishmentAccept:
		a.GsmMessage.PDUSessionEstablishmentAccept = nasMessage.NewPDUSessionEstablishmentAccept(MsgTypePDUSessionEstablishmentAccept)
		a.GsmMessage.DecodePDUSessionEstablishmentAccept(byteArray)
	case MsgTypePDUSessionEstablishmentReject:
		a.GsmMessage.PDUSessionEstablishmentReject = nasMessage.NewPDUSessionEstablishmentReject(MsgTypePDUSessionEstablishmentReject)
		a.GsmMessage.DecodePDUSessionEstablishmentReject(byteArray)
	case MsgTypePDUSessionAuthenticationCommand:
		a.GsmMessage.PDUSessionAuthenticationCommand = nasMessage.NewPDUSessionAuthenticationCommand(MsgTypePDUSessionAuthenticationCommand)
		a.GsmMessage.DecodePDUSessionAuthenticationCommand(byteArray)
	case MsgTypePDUSessionAuthenticationComplete:
		a.GsmMessage.PDUSessionAuthenticationComplete = nasMessage.NewPDUSessionAuthenticationComplete(MsgTypePDUSessionAuthenticationComplete)
		a.GsmMessage.DecodePDUSessionAuthenticationComplete(byteArray)
	case MsgTypePDUSessionAuthenticationResult:
		a.GsmMessage.PDUSessionAuthenticationResult = nasMessage.NewPDUSessionAuthenticationResult(MsgTypePDUSessionAuthenticationResult)
		a.GsmMessage.DecodePDUSessionAuthenticationResult(byteArray)
	case MsgTypePDUSessionModificationRequest:
		a.GsmMessage.PDUSessionModificationRequest = nasMessage.NewPDUSessionModificationRequest(MsgTypePDUSessionModificationRequest)
		a.GsmMessage.DecodePDUSessionModificationRequest(byteArray)
	case MsgTypePDUSessionModificationReject:
		a.GsmMessage.PDUSessionModificationReject = nasMessage.NewPDUSessionModificationReject(MsgTypePDUSessionModificationReject)
		a.GsmMessage.DecodePDUSessionModificationReject(byteArray)
	case MsgTypePDUSessionModificationCommand:
		a.GsmMessage.PDUSessionModificationCommand = nasMessage.NewPDUSessionModificationCommand(MsgTypePDUSessionModificationCommand)
		a.GsmMessage.DecodePDUSessionModificationCommand(byteArray)
	case MsgTypePDUSessionModificationComplete:
		a.GsmMessage.PDUSessionModificationComplete = nasMessage.NewPDUSessionModificationComplete(MsgTypePDUSessionModificationComplete)
		a.GsmMessage.DecodePDUSessionModificationComplete(byteArray)
	case MsgTypePDUSessionModificationCommandReject:
		a.GsmMessage.PDUSessionModificationCommandReject = nasMessage.NewPDUSessionModificationCommandReject(MsgTypePDUSessionModificationCommandReject)
		a.GsmMessage.DecodePDUSessionModificationCommandReject(byteArray)
	case MsgTypePDUSessionReleaseRequest:
		a.GsmMessage.PDUSessionReleaseRequest = nasMessage.NewPDUSessionReleaseRequest(MsgTypePDUSessionReleaseRequest)
		a.GsmMessage.DecodePDUSessionReleaseRequest(byteArray)
	case MsgTypePDUSessionReleaseReject:
		a.GsmMessage.PDUSessionReleaseReject = nasMessage.NewPDUSessionReleaseReject(MsgTypePDUSessionReleaseReject)
		a.GsmMessage.DecodePDUSessionReleaseReject(byteArray)
	case MsgTypePDUSessionReleaseCommand:
		a.GsmMessage.PDUSessionReleaseCommand = nasMessage.NewPDUSessionReleaseCommand(MsgTypePDUSessionReleaseCommand)
		a.GsmMessage.DecodePDUSessionReleaseCommand(byteArray)
	case MsgTypePDUSessionReleaseComplete:
		a.GsmMessage.PDUSessionReleaseComplete = nasMessage.NewPDUSessionReleaseComplete(MsgTypePDUSessionReleaseComplete)
		a.GsmMessage.DecodePDUSessionReleaseComplete(byteArray)
	case MsgTypeStatus5GSM:
		a.GsmMessage.Status5GSM = nasMessage.NewStatus5GSM(MsgTypeStatus5GSM)
		a.GsmMessage.DecodeStatus5GSM(byteArray)
	default:
		return fmt.Errorf("NAS Decode Fail: MsgType[%d] doesn't exist in GSM message", a.GsmMessage.GsmHeader.GetMessageType())
	}
	return nil
}

func (a *Message) GsmMessageEncode(buffer *bytes.Buffer) error {

	log.Debug(spew.Sdump(a))

	switch a.GsmMessage.GsmHeader.GetMessageType() {
	case MsgTypePDUSessionEstablishmentRequest:
		a.GsmMessage.EncodePDUSessionEstablishmentRequest(buffer)
	case MsgTypePDUSessionEstablishmentAccept:
		a.GsmMessage.EncodePDUSessionEstablishmentAccept(buffer)
	case MsgTypePDUSessionEstablishmentReject:
		a.GsmMessage.EncodePDUSessionEstablishmentReject(buffer)
	case MsgTypePDUSessionAuthenticationCommand:
		a.GsmMessage.EncodePDUSessionAuthenticationCommand(buffer)
	case MsgTypePDUSessionAuthenticationComplete:
		a.GsmMessage.EncodePDUSessionAuthenticationComplete(buffer)
	case MsgTypePDUSessionAuthenticationResult:
		a.GsmMessage.EncodePDUSessionAuthenticationResult(buffer)
	case MsgTypePDUSessionModificationRequest:
		a.GsmMessage.EncodePDUSessionModificationRequest(buffer)
	case MsgTypePDUSessionModificationReject:
		a.GsmMessage.EncodePDUSessionModificationReject(buffer)
	case MsgTypePDUSessionModificationCommand:
		a.GsmMessage.EncodePDUSessionModificationCommand(buffer)
	case MsgTypePDUSessionModificationComplete:
		a.GsmMessage.EncodePDUSessionModificationComplete(buffer)
	case MsgTypePDUSessionModificationCommandReject:
		a.GsmMessage.EncodePDUSessionModificationCommandReject(buffer)
	case MsgTypePDUSessionReleaseRequest:
		a.GsmMessage.EncodePDUSessionReleaseRequest(buffer)
	case MsgTypePDUSessionReleaseReject:
		a.GsmMessage.EncodePDUSessionReleaseReject(buffer)
	case MsgTypePDUSessionReleaseCommand:
		a.GsmMessage.EncodePDUSessionReleaseCommand(buffer)
	case MsgTypePDUSessionReleaseComplete:
		a.GsmMessage.EncodePDUSessionReleaseComplete(buffer)
	case MsgTypeStatus5GSM:
		a.GsmMessage.EncodeStatus5GSM(buffer)
	default:
		return fmt.Errorf("NAS Encode Fail: MsgType[%d] doesn't exist in GSM message", a.GsmMessage.GsmHeader.GetMessageType())
	}
	return nil
}
