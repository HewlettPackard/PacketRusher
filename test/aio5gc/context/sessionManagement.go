/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"errors"
	"net"

	"github.com/free5gc/nas/nasConvert"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/openapi/models"
	"github.com/mohae/deepcopy"
	log "github.com/sirupsen/logrus"
)

type SmContext struct {
	// pdu session information
	pduSessionID                 int32
	snssai                       models.Snssai
	pduAddress                   net.IP
	dataNetwork                  DataNetwork
	userLocation                 models.NrLocation
	plmnID                       models.PlmnId
	pti                          uint8
	sessionType                  uint8
	ProtocolConfigurationOptions *ProtocolConfigurationOptions
	sessionRule                  *models.SessionRule
	defQosQFI                    uint8
}

type ProtocolConfigurationOptions struct {
	DNSIPv4Request     bool
	DNSIPv6Request     bool
	PCSCFIPv4Request   bool
	IPv4LinkMTURequest bool
}

func NewSmContext(pduSessionID int32) *SmContext {
	c := &SmContext{pduSessionID: pduSessionID}
	c.ProtocolConfigurationOptions = &ProtocolConfigurationOptions{}
	return c
}

func (c *SmContext) GetPduSessionId() int32 {
	return c.pduSessionID
}

func (c *SmContext) SetSnssai(snssai models.Snssai) {
	c.snssai = snssai
}

func (c *SmContext) SetPDUAddress(ip net.IP) {
	c.pduAddress = ip
}

func (c *SmContext) GetSnnsai() models.Snssai {
	return c.snssai
}

func (c *SmContext) SetDataNetwork(dn DataNetwork) {
	c.dataNetwork = dn
}

func (c *SmContext) GetDataNetwork() DataNetwork {
	return c.dataNetwork
}

func (c *SmContext) SetUserLocation(location models.NrLocation) {
	c.userLocation = location
}

func (c *SmContext) SetPti(pti uint8) {
	c.pti = pti
}

func (c *SmContext) GetPti() uint8 {
	return c.pti
}

func (c *SmContext) SetPduSessionType(sType uint8) {
	c.sessionType = sType
}

func (c *SmContext) GetPduSessionType() uint8 {
	return c.sessionType
}

func (c *SmContext) GetSessionRule() *models.SessionRule {
	return c.sessionRule
}

func (c *SmContext) SetSessionRule(sessionRule *models.SessionRule) {
	c.sessionRule = sessionRule
}

func (c *SmContext) GetDefQosQFI() uint8 {
	return c.defQosQFI
}

func (c *SmContext) SetDefQosQFI(defQosQFI uint8) {
	c.defQosQFI = defQosQFI
}

func (smContext *SmContext) PDUAddressToNAS() ([12]byte, uint8) {
	var addr [12]byte
	var addrLen uint8
	copy(addr[:], smContext.pduAddress)
	switch smContext.sessionType {
	case nasMessage.PDUSessionTypeIPv4:
		addrLen = 4 + 1
	case nasMessage.PDUSessionTypeIPv6:
	case nasMessage.PDUSessionTypeIPv4IPv6:
		addrLen = 12 + 1
	}
	return addr, addrLen
}

func CreatePDUSession(sessionRequest *nasMessage.PDUSessionEstablishmentRequest,
	ue *UEContext,
	session *SessionContext,
	pduSessionID int32,
	snssai models.Snssai,
	dnn string,
) (smContext *SmContext, err error) {

	newSmContext := NewSmContext(pduSessionID)
	newSmContext.SetSnssai(snssai)
	dn, err := session.GetDataNetwork(dnn)
	if err != nil {
		return nil, err
	}
	newSmContext.SetDataNetwork(dn)

	locationCopy := deepcopy.Copy(*ue.GetUserLocationInfo()).(models.NrLocation)
	newSmContext.SetUserLocation(locationCopy)

	newSmContext.SetPti(sessionRequest.GetPTI())
	newSmContext.SetPduSessionType(sessionRequest.GetPDUSessionTypeValue())
	newSmContext.SetSessionRule(session.GetSessionRules()[0])
	newSmContext.SetDefQosQFI(uint8(1))

	newSmContext.SetPDUAddress(session.GetUnallocatedIP())
	EPCOContents := sessionRequest.ExtendedProtocolConfigurationOptions.GetExtendedProtocolConfigurationOptionsContents()
	protocolConfigurationOptions := nasConvert.NewProtocolConfigurationOptions()
	err = protocolConfigurationOptions.UnMarshal(EPCOContents)
	if err != nil {
		return nil, errors.New("[5GC][NAS] Error while decoding protocol configuration options : " + err.Error())
	}
	for _, container := range protocolConfigurationOptions.ProtocolOrContainerList {
		switch container.ProtocolOrContainerID {
		case nasMessage.DNSServerIPv6AddressRequestUL:
			newSmContext.ProtocolConfigurationOptions.DNSIPv6Request = true
		case nasMessage.PCSCFIPv4AddressRequestUL:
			newSmContext.ProtocolConfigurationOptions.PCSCFIPv4Request = true
		case nasMessage.DNSServerIPv4AddressRequestUL:
			newSmContext.ProtocolConfigurationOptions.DNSIPv4Request = true
		case nasMessage.IPv4LinkMTURequestUL:
			newSmContext.ProtocolConfigurationOptions.IPv4LinkMTURequest = true
		}
	}

	err = ue.AddSmContext(newSmContext)
	if err != nil {
		return nil, err
	}
	log.Infof("[5GC] Create smContext[pduSessionID: %d] Success", pduSessionID)
	return newSmContext, nil
}
