/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"errors"
	"sync"

	"github.com/free5gc/nas/nasType"
	"github.com/free5gc/ngap/ngapType"
	"github.com/ishidawataru/sctp"
)

// UE main states in the GNB Context.
const Initialized = 0x00
const Ongoing = 0x01
const Ready = 0x02
const Down = 0x03

type GNBUe struct {
	ranUeNgapId    int64          // Identifier for UE in GNB Context.
	amfUeNgapId    int64          // Identifier for UE in AMF Context.
	amfId          int64          // Identifier for AMF in UE/GNB Context.
	state          int            // State of UE in NAS/GNB Context.
	sctpConnection *sctp.SCTPConn // Sctp association in using by the UE.
	gnbRx          chan UEMessage
	gnbTx          chan UEMessage
	pRueId         int64 // PacketRusher unique UE ID
	tmsi           *nasType.GUTI5G
	context        Context
	lock           sync.Mutex
	newGnb         *GNBContext
}

type Context struct {
	mobilityInfo           mobility
	maskedIMEISV           string
	pduSession             [16]*GnbPDUSession
	allowedSst             []string
	allowedSd              []string
	lenSlice               int
	ueSecurityCapabilities *ngapType.UESecurityCapabilities
}

type GnbPDUSession struct {
	pduSessionId int64
	upfIp        string
	sst          string
	sd           string
	uplinkTeid   uint32
	downlinkTeid uint32
	pduType      uint64
	qosId        int64
	fiveQi       int64
	priArp       int64
}

type mobility struct {
	mcc string
	mnc string
}

func (ue *GNBUe) CreateUeContext(plmn string, imeisv string, sst []string, sd []string, ueSecurityCapabilities *ngapType.UESecurityCapabilities) {
	if plmn != "not informed" {
		ue.context.mobilityInfo.mcc, ue.context.mobilityInfo.mnc = convertMccMnc(plmn)
	} else {
		ue.context.mobilityInfo.mcc = plmn
		ue.context.mobilityInfo.mnc = plmn
	}

	ue.context.maskedIMEISV = imeisv
	ue.context.allowedSst = sst
	ue.context.allowedSd = sd
	ue.context.ueSecurityCapabilities = ueSecurityCapabilities
}

func (ue *GNBUe) CopyFromPreviousContext(oldUeContext *GNBUe) {
	ue.SetAmfUeId(oldUeContext.GetAmfUeId())
	ue.context = oldUeContext.context
}

func (ue *GNBUe) CreatePduSession(pduSessionId int64, upfIp string, sst string, sd string, pduType uint64,
	qosId int64, priArp int64, fiveQi int64, ulTeid uint32, dlTeid uint32) (*GnbPDUSession, error) {

	if pduSessionId < 1 && pduSessionId > 16 {
		return nil, errors.New("PDU Session Id must lies between 0 and 15, id: " + string(pduSessionId))
	}

	if ue.context.pduSession[pduSessionId-1] != nil {
		return nil, errors.New("Unable to create PDU Session " + string(pduSessionId) + " as such PDU Session already exists")
	}

	var pduSession = new(GnbPDUSession)
	pduSession.pduSessionId = pduSessionId
	pduSession.upfIp = upfIp
	if !ue.isWantedNssai(sst, sd) {
		return nil, errors.New("Unable to create PDU Session, slice " + string(sst) + string(sd) + " is not selected for current UE")
	}
	pduSession.pduType = pduType
	pduSession.qosId = qosId
	pduSession.priArp = priArp
	pduSession.fiveQi = fiveQi
	pduSession.uplinkTeid = ulTeid
	pduSession.downlinkTeid = dlTeid
	pduSession.sst = sst
	pduSession.sd = sd

	ue.context.pduSession[pduSessionId-1] = pduSession

	return pduSession, nil
}

func (ue *GNBUe) GetPduSession(pduSessionId int64) (*GnbPDUSession, error) {
	if pduSessionId < 1 && pduSessionId > 16 {
		return nil, errors.New("PDU Session Id must lies between 1 and 16, id: " + string(pduSessionId))
	}

	return ue.context.pduSession[pduSessionId-1], nil
}

func (ue *GNBUe) GetPduSessions() [16]*GnbPDUSession {
	return ue.context.pduSession
}

func (ue *GNBUe) SetPduSessions(pduSessions [16]*GnbPDUSession) {
	ue.context.pduSession = pduSessions
}

func (ue *GNBUe) DeletePduSession(pduSessionId int64) error {
	if pduSessionId < 1 && pduSessionId > 16 {
		return errors.New("PDU Session Id must lies between 1 and 16, id: " + string(pduSessionId))
	}

	ue.context.pduSession[pduSessionId-1] = nil

	return nil
}

func (ue *GNBUe) GetUeMobility() (string, string) {
	return ue.context.mobilityInfo.mcc, ue.context.mobilityInfo.mnc
}

func (ue *GNBUe) GetUeMaskedImeiSv() string {
	return ue.context.maskedIMEISV
}

func (ue *GNBUe) GetSelectedNssai(pduSessionId int64) (string, string) {
	pduSession := ue.context.pduSession[pduSessionId-1]
	if pduSession != nil {
		return pduSession.sst, pduSession.sd
	}

	return "NSSAI was not selected", "NSSAI was not selected"
}

func (ue *GNBUe) GetUESecurityCapabilities() *ngapType.UESecurityCapabilities {
	return ue.context.ueSecurityCapabilities
}

func (ue *GNBUe) isWantedNssai(sst string, sd string) bool {
	if len(ue.context.allowedSst) == len(ue.context.allowedSd) {
		for i := range ue.context.allowedSst {
			if ue.context.allowedSst[i] == sst && ue.context.allowedSd[i] == sd {
				return true
			}
		}
	}

	return false
}

func (ue *GNBUe) GetAmfId() int64 {
	return ue.amfId
}

func (ue *GNBUe) SetAmfId(id int64) {
	ue.amfId = id
}

func (ue *GNBUe) GetSCTP() *sctp.SCTPConn {
	return ue.sctpConnection
}

func (ue *GNBUe) SetSCTP(conn *sctp.SCTPConn) {
	ue.sctpConnection = conn
}

func (ue *GNBUe) GetState() int {
	return ue.state
}

func (ue *GNBUe) SetStateInitialized() {
	ue.state = Initialized
}

func (ue *GNBUe) SetStateOngoing() {
	ue.state = Ongoing
}

func (ue *GNBUe) SetStateReady() {
	ue.state = Ready
}

func (ue *GNBUe) SetStateDown() {
	ue.state = Down
}

func (ue *GNBUe) SetHandoverGnodeB(gnb *GNBContext) {
	ue.newGnb = gnb
}

func (ue *GNBUe) GetHandoverGnodeB() *GNBContext {
	return ue.newGnb
}

func (ue *GNBUe) GetGnbRx() chan UEMessage {
	return ue.gnbRx
}

func (ue *GNBUe) SetGnbRx(gnbRx chan UEMessage) {
	ue.gnbRx = gnbRx
}

func (ue *GNBUe) GetGnbTx() chan UEMessage {
	return ue.gnbTx
}

func (ue *GNBUe) SetGnbTx(gnbTx chan UEMessage) {
	ue.gnbTx = gnbTx
}

func (ue *GNBUe) SetPrUeId(pRueId int64) {
	ue.pRueId = pRueId
}

func (ue *GNBUe) GetPrUeId() int64 {
	return ue.pRueId
}

func (ue *GNBUe) SetTMSI(tmsi *nasType.GUTI5G) {
	ue.tmsi = tmsi
}

func (ue *GNBUe) GetTMSI() *nasType.GUTI5G {
	return ue.tmsi
}

func (ue *GNBUe) Lock() {
	ue.lock.Lock()
}

func (ue *GNBUe) Unlock() {
	ue.lock.Unlock()
}

func (pduSession *GnbPDUSession) GetPduSessionId() int64 {
	return pduSession.pduSessionId
}

func (pduSession *GnbPDUSession) GetUpfIp() string {
	return pduSession.upfIp
}

func (pduSession *GnbPDUSession) SetUpfIp(upfIp string) {
	pduSession.upfIp = upfIp
}

func (pduSession *GnbPDUSession) GetTeidUplink() uint32 {
	return pduSession.uplinkTeid
}

func (pduSession *GnbPDUSession) SetTeidUplink(teidUplink uint32) {
	pduSession.uplinkTeid = teidUplink
}

func (pduSession *GnbPDUSession) GetTeidDownlink() uint32 {
	return pduSession.downlinkTeid
}

func (pduSession *GnbPDUSession) SetTeidDownlink(teidDownlink uint32) {
	pduSession.downlinkTeid = teidDownlink
}

func (pduSession *GnbPDUSession) GetQosId() int64 {
	return pduSession.qosId
}

func (pduSession *GnbPDUSession) GetFiveQI() int64 {
	return pduSession.fiveQi
}

func (pduSession *GnbPDUSession) GetPriorityARP() int64 {
	return pduSession.priArp
}

func (pduSession *GnbPDUSession) GetPduType() (valor string) {

	switch pduSession.pduType {
	case 0:
		valor = "ipv4"
	case 1:
		valor = "ipv6"
	case 2:
		valor = "Ipv4Ipv6"
	case 3:
		valor = "ethernet"

	}
	return
}

func (ue *GNBUe) GetRanUeId() int64 {
	return ue.ranUeNgapId
}

func (ue *GNBUe) SetRanUeId(id int64) {
	ue.ranUeNgapId = id
}

func (ue *GNBUe) GetAmfUeId() int64 {
	return ue.amfUeNgapId
}

func (ue *GNBUe) SetAmfUeId(amfUeId int64) {
	ue.amfUeNgapId = amfUeId
}
