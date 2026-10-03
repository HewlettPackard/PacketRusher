/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"fmt"
	"sync"
	"sync/atomic"

	nasType "github.com/free5gc/nas/ie"
	ngapType "github.com/free5gc/ngap/ie"
	"github.com/ishidawataru/sctp"
)

// UE main states in the GNB Context.
const Initialized = 0x00
const Ongoing = 0x01
const Ready = 0x02
const Down = 0x03

type GNBUe struct {
	ranUeNgapId    int64 // Identifier for UE in GNB Context.
	amfUeNgapId    atomic.Int64
	amfUeIDSet     atomic.Bool    // Identifier for UE in AMF Context.
	amfId          int64          // Identifier for AMF in UE/GNB Context.
	state          atomic.Int64   // State of UE in NAS/GNB Context.
	sctpConnection *sctp.SCTPConn // Sctp association in using by the UE.
	gnbRx          chan UEMessage
	gnbTx          chan UEMessage
	pRueId         int64 // PacketRusher unique UE ID
	tmsi           *nasType.MobileId5GS
	context        Context
	// contextMu protects session membership and the payload copied during Xn
	// handover. It is independent of the processing and release-request locks.
	contextMu        sync.RWMutex
	lock             sync.Mutex
	processingLock   sync.Mutex
	txLock           sync.Mutex
	delivery         *ueDelivery
	pendingDelivery  *[]ueMessageDelivery // Scoped to the current downlink handler, guarded by txLock.
	connectionLost   chan struct{}
	connectionFailed bool
	newGnb           atomic.Pointer[GNBContext]
	releaseRequested bool // Set when UE Context Release Request is sent to AMF
}

type Context struct {
	mobilityInfo           mobility
	maskedIMEISV           string
	pduSession             [16]*GnbPDUSession
	allowedSst             []string
	allowedSd              []string
	ueSecurityCapabilities *ngapType.UESecurityCapabilities
}

type GnbPDUSession struct {
	// Xn handover copies session pointers. Only tunnel endpoint fields mutate
	// after publication, and their access must be safe across both gNB lanes.
	tunnelMu     sync.RWMutex
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
	ue.contextMu.Lock()
	defer ue.contextMu.Unlock()

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
	// Take an independent membership snapshot, retaining the existing shallow
	// session/metadata copy. Never hold two UE locks, including for self-copy.
	oldUeContext.contextMu.RLock()
	previous := oldUeContext.context
	oldUeContext.contextMu.RUnlock()
	ue.contextMu.Lock()
	ue.context = previous
	ue.contextMu.Unlock()
}

func (ue *GNBUe) CreatePduSession(pduSessionId int64, upfIp string, sst string, sd string, pduType uint64,
	qosId int64, priArp int64, fiveQi int64, ulTeid uint32, dlTeid uint32) (*GnbPDUSession, error) {

	if pduSessionId < 1 || pduSessionId > 15 {
		return nil, fmt.Errorf("PDU session ID must be between 1 and 15, id: %d", pduSessionId)
	}
	ue.contextMu.Lock()
	defer ue.contextMu.Unlock()

	if ue.context.pduSession[pduSessionId-1] != nil {
		return nil, fmt.Errorf("unable to create PDU Session %d as such PDU Session already exists", pduSessionId)
	}

	var pduSession = new(GnbPDUSession)
	pduSession.pduSessionId = pduSessionId
	pduSession.upfIp = upfIp
	if !ue.isWantedNssai(sst, sd) {
		return nil, fmt.Errorf("unable to create PDU Session, slice sst:%s sd:%s is not selected for current UE", sst, sd)
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
	if pduSessionId < 1 || pduSessionId > 15 {
		return nil, fmt.Errorf("PDU session ID must be between 1 and 15, id: %d", pduSessionId)
	}
	ue.contextMu.RLock()
	defer ue.contextMu.RUnlock()

	return ue.context.pduSession[pduSessionId-1], nil
}

func (ue *GNBUe) GetPduSessions() [16]*GnbPDUSession {
	ue.contextMu.RLock()
	defer ue.contextMu.RUnlock()
	return ue.context.pduSession
}

func (ue *GNBUe) SetPduSessions(pduSessions [16]*GnbPDUSession) {
	ue.contextMu.Lock()
	defer ue.contextMu.Unlock()
	ue.context.pduSession = pduSessions
}

func (ue *GNBUe) DeletePduSession(pduSessionId int64) error {
	if pduSessionId < 1 || pduSessionId > 15 {
		return fmt.Errorf("PDU session ID must be between 1 and 15, id: %d", pduSessionId)
	}
	ue.contextMu.Lock()
	defer ue.contextMu.Unlock()

	ue.context.pduSession[pduSessionId-1] = nil

	return nil
}

func (ue *GNBUe) GetUeMobility() (string, string) {
	ue.contextMu.RLock()
	defer ue.contextMu.RUnlock()
	return ue.context.mobilityInfo.mcc, ue.context.mobilityInfo.mnc
}

func (ue *GNBUe) GetUeMaskedImeiSv() string {
	ue.contextMu.RLock()
	defer ue.contextMu.RUnlock()
	return ue.context.maskedIMEISV
}

func (ue *GNBUe) GetSelectedNssai(pduSessionId int64) (string, string) {
	if pduSessionId < 1 || pduSessionId > 15 {
		return "NSSAI was not selected", "NSSAI was not selected"
	}
	ue.contextMu.RLock()
	defer ue.contextMu.RUnlock()
	pduSession := ue.context.pduSession[pduSessionId-1]
	if pduSession != nil {
		return pduSession.sst, pduSession.sd
	}

	return "NSSAI was not selected", "NSSAI was not selected"
}

func (ue *GNBUe) GetUESecurityCapabilities() *ngapType.UESecurityCapabilities {
	ue.contextMu.RLock()
	defer ue.contextMu.RUnlock()
	return ue.context.ueSecurityCapabilities
}

// isWantedNssai requires contextMu; it is only called while creating a session.
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
	return int(ue.state.Load())
}

func (ue *GNBUe) SetStateInitialized() {
	ue.state.Store(Initialized)
}

func (ue *GNBUe) SetStateOngoing() {
	ue.state.Store(Ongoing)
}

func (ue *GNBUe) SetStateReady() {
	ue.state.Store(Ready)
}

func (ue *GNBUe) SetStateDown() {
	ue.state.Store(Down)
}

func (ue *GNBUe) SetHandoverGnodeB(gnb *GNBContext) {
	ue.newGnb.Store(gnb)
}

func (ue *GNBUe) GetHandoverGnodeB() *GNBContext {
	return ue.newGnb.Load()
}

func (ue *GNBUe) GetGnbRx() chan UEMessage {
	return ue.gnbRx
}

func (ue *GNBUe) SetGnbRx(gnbRx chan UEMessage) {
	ue.gnbRx = gnbRx
}

func (ue *GNBUe) GetGnbTx() chan UEMessage {
	ue.txLock.Lock()
	defer ue.txLock.Unlock()
	return ue.gnbTx
}

func (ue *GNBUe) SetGnbTx(gnbTx chan UEMessage) {
	ue.txLock.Lock()
	if ue.gnbTx == gnbTx {
		ue.txLock.Unlock()
		return
	}
	previous := ue.delivery
	ue.gnbTx = gnbTx
	if gnbTx != nil {
		ue.delivery = &ueDelivery{channel: gnbTx, done: make(chan struct{})}
	} else {
		ue.delivery = nil
	}
	if previous != nil {
		close(previous.done)
	}
	ue.txLock.Unlock()
	// The swap prevents later batches from enrolling in the old generation.
	// Already enrolled sends need cancellation before its channel can close.
	if previous != nil {
		previous.senders.Wait()
		close(previous.channel)
	}
}

func (ue *GNBUe) SetPrUeId(pRueId int64) {
	ue.pRueId = pRueId
}

func (ue *GNBUe) GetPrUeId() int64 {
	return ue.pRueId
}

func (ue *GNBUe) SetTMSI(tmsi *nasType.MobileId5GS) {
	ue.tmsi = tmsi
}

func (ue *GNBUe) GetTMSI() *nasType.MobileId5GS {
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
	pduSession.tunnelMu.RLock()
	defer pduSession.tunnelMu.RUnlock()
	return pduSession.upfIp
}

func (pduSession *GnbPDUSession) SetUpfIp(upfIp string) {
	pduSession.tunnelMu.Lock()
	defer pduSession.tunnelMu.Unlock()
	pduSession.upfIp = upfIp
}

func (pduSession *GnbPDUSession) GetTeidUplink() uint32 {
	pduSession.tunnelMu.RLock()
	defer pduSession.tunnelMu.RUnlock()
	return pduSession.uplinkTeid
}

func (pduSession *GnbPDUSession) SetTeidUplink(teidUplink uint32) {
	pduSession.tunnelMu.Lock()
	defer pduSession.tunnelMu.Unlock()
	pduSession.uplinkTeid = teidUplink
}

func (pduSession *GnbPDUSession) GetTeidDownlink() uint32 {
	pduSession.tunnelMu.RLock()
	defer pduSession.tunnelMu.RUnlock()
	return pduSession.downlinkTeid
}

func (pduSession *GnbPDUSession) SetTeidDownlink(teidDownlink uint32) {
	pduSession.tunnelMu.Lock()
	defer pduSession.tunnelMu.Unlock()
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
	return ue.amfUeNgapId.Load()
}

func (ue *GNBUe) SetAmfUeId(amfUeId int64) {
	ue.amfUeNgapId.Store(amfUeId)
	ue.amfUeIDSet.Store(true)
}

func (ue *GNBUe) SetReleaseRequested(val bool) {
	ue.lock.Lock()
	defer ue.lock.Unlock()
	ue.releaseRequested = val
}

func (ue *GNBUe) GetReleaseRequested() bool {
	ue.lock.Lock()
	defer ue.lock.Unlock()
	return ue.releaseRequested
}

// LockProcessing serializes uplink NAS and downlink NGAP changes to this UE.
func (ue *GNBUe) LockProcessing()   { ue.processingLock.Lock() }
func (ue *GNBUe) UnlockProcessing() { ue.processingLock.Unlock() }

// ProcessDownlink serializes context changes, then delivers the handler's messages
// outside the processing lock. The ordered NGAP worker waits for delivery before
// running the next handler, so a release cannot overtake a preceding NAS message.
// Meanwhile uplink processing can drain RX even when the UE's TX buffer is full.
func (ue *GNBUe) ProcessDownlink(process func()) {
	var pending []ueMessageDelivery
	func() {
		ue.LockProcessing()
		ue.txLock.Lock()
		ue.pendingDelivery = &pending
		ue.txLock.Unlock()
		defer func() {
			ue.txLock.Lock()
			ue.pendingDelivery = nil
			ue.txLock.Unlock()
			ue.UnlockProcessing()
		}()
		process()
	}()
	for _, message := range pending {
		ue.deliver(message)
	}
}

type ueMessageDelivery struct {
	delivery *ueDelivery
	message  UEMessage
}

type ueDelivery struct {
	channel chan UEMessage
	done    chan struct{}
	senders sync.WaitGroup
}

// DeliverToUE keeps channel ownership with the gNB. Closing first cancels blocked
// sends, then waits for them to leave before closing the receive channel.
func (ue *GNBUe) DeliverToUE(message UEMessage) bool {
	ue.txLock.Lock()
	delivery := ue.delivery
	if delivery == nil {
		ue.txLock.Unlock()
		return false
	}
	if ue.pendingDelivery != nil {
		*ue.pendingDelivery = append(*ue.pendingDelivery, ueMessageDelivery{delivery, message})
		ue.txLock.Unlock()
		return true
	}
	ue.txLock.Unlock()
	return ue.deliver(ueMessageDelivery{delivery, message})
}

func (ue *GNBUe) deliver(message ueMessageDelivery) bool {
	ue.txLock.Lock()
	delivery := message.delivery
	// A released or replaced connection must never receive an old batch.
	if delivery != ue.delivery {
		ue.txLock.Unlock()
		return false
	}
	delivery.senders.Add(1)
	ue.txLock.Unlock()
	defer delivery.senders.Done()
	select {
	case <-delivery.done:
		return false
	default:
	}
	select {
	case delivery.channel <- message.message:
		return true
	case <-delivery.done:
		return false
	}
}

func (ue *GNBUe) CloseUEChannel() {
	ue.txLock.Lock()
	delivery := ue.delivery
	ue.delivery = nil
	ue.gnbTx = nil
	if delivery != nil {
		close(delivery.done)
	}
	ue.txLock.Unlock()
	if delivery != nil {
		delivery.senders.Wait()
		close(delivery.channel)
	}
}

func (ue *GNBUe) HasAmfUeId() bool { return ue.amfUeIDSet.Load() }

// Normal context release (including Idle) closes TX without failing the UE.
// Association loss uses a separate signal that cannot be blocked by a full TX.
func (ue *GNBUe) SetConnectionLost(lost chan struct{}) {
	ue.txLock.Lock()
	defer ue.txLock.Unlock()
	if ue.connectionLost == lost {
		return
	}
	ue.connectionLost = lost
	if ue.connectionFailed && lost != nil {
		close(lost)
	}
}

func (ue *GNBUe) FailUEChannel() {
	ue.txLock.Lock()
	if !ue.connectionFailed {
		ue.connectionFailed = true
		if ue.connectionLost != nil {
			close(ue.connectionLost)
		}
	}
	ue.txLock.Unlock()
	ue.CloseUEChannel()
}
