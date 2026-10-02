/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"encoding/hex"
	"errors"
	"fmt"
	"iter"
	"net/netip"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"my5G-RANTester/internal/control_test_engine/gnb/gtp"

	nasType "github.com/free5gc/nas/ie"
	"github.com/free5gc/ngap/aper"

	ngapType "github.com/free5gc/ngap/ie"

	"github.com/ishidawataru/sctp"
	log "github.com/sirupsen/logrus"
)

type GNBContext struct {
	dataInfo       DataInfo    // gnb data plane information
	controlInfo    ControlInfo // gnb control plane information
	uePool         sync.Map    // map[int64]*GNBUe, UeRanNgapId as key
	prUePool       sync.Map    // map[int64]*GNBUe, PrUeId as key
	amfPool        sync.Map    // map[int64]*GNBAmf, AmfId as key
	teidPool       sync.Map    // map[uint32]*GNBUe, downlinkTeid as key
	sliceInfo      Slice
	idUeGenerator  atomic.Int64  // ran UE id, incremented atomically.
	idAmfGenerator atomic.Int64  // ran amf id, incremented atomically.
	teidGenerator  atomic.Uint32 // ran UE downlink Teid, incremented atomically.
	ueIpGenerator  uint8         // ran ue ip.
	pagedUEs       []PagedUE
	pagedUELock    sync.Mutex
	gtpDevice      atomic.Pointer[gtp.Device] // GTP-U device shared by this gNB's UEs, if any
}

type DataInfo struct {
	gnbIpPort netip.AddrPort // gnb ip and port for data plane.
}

type Slice struct {
	sd  string
	sst string
}

type ControlInfo struct {
	mcc            string
	mnc            string
	tac            string
	gnbId          string
	gnbIpPort      netip.AddrPort
	inboundChannel chan UEMessage
	n2             atomic.Pointer[sctp.SCTPConn]
	terminated     atomic.Bool
	// lifecycle orders publishing a new association against Terminate and against
	// removing its AMF, so that either the dialler sees it is no longer wanted, or the
	// closer sees the association and closes it. This holds for each AMF, because
	// Terminate closes every AMF's association, not only the last one set as N2.
	lifecycle sync.Mutex
}

type PagedUE struct {
	FiveGSTMSI *ngapType.FiveGSTMSI
	Timestamp  time.Time
}

func (gnb *GNBContext) NewRanGnbContext(gnbId, mcc, mnc, tac, sst, sd string, n2, n3 netip.AddrPort) {
	gnb.controlInfo.mcc = mcc
	gnb.controlInfo.mnc = mnc
	gnb.controlInfo.tac = tac
	gnb.controlInfo.gnbId = gnbId
	gnb.controlInfo.inboundChannel = make(chan UEMessage, 100)
	gnb.sliceInfo.sd = sd
	gnb.sliceInfo.sst = sst
	gnb.controlInfo.gnbIpPort = n2
	gnb.ueIpGenerator = 3
	gnb.dataInfo.gnbIpPort = n3
}

func (gnb *GNBContext) NewGnBUe(gnbTx chan UEMessage, gnbRx chan UEMessage, prUeId int64, tmsi *nasType.MobileId5GS) (*GNBUe, error) {

	// TODO if necessary add more information for UE.

	// new instance of ue.
	ue := &GNBUe{}

	// set ran UE Ngap Id.
	ranId := gnb.getRanUeId()
	ue.SetRanUeId(ranId)

	ue.SetAmfUeId(0)

	// Connect gNB and UE's channels
	ue.SetGnbRx(gnbRx)
	ue.SetGnbTx(gnbTx)
	ue.SetPrUeId(prUeId)
	ue.SetTMSI(tmsi)

	// set state to UE.
	ue.SetStateInitialized()

	// select AMF with Capacity is more than 0. Done before the UE is stored, so that a UE
	// arriving while no association is up is not left behind in the pools.
	amf := gnb.selectAmFByActive()
	if amf == nil {
		return nil, errors.New("no AMF available for this UE")
	}

	// set amfId and SCTP association for UE.
	ue.SetAmfId(amf.GetAmfId())
	ue.SetSCTP(amf.GetSCTPConn())

	// store UE in the UE Pool of GNB.
	gnb.uePool.Store(ranId, ue)
	if prUeId != 0 {
		gnb.prUePool.Store(prUeId, ue)
	}

	// return UE Context.
	return ue, nil
}

func (gnb *GNBContext) GetInboundChannel() chan UEMessage {
	return gnb.controlInfo.inboundChannel
}

func (gnb *GNBContext) GetN3GnbIp() netip.Addr {
	return gnb.dataInfo.gnbIpPort.Addr()
}

// SetGtpDevice gives this gNB the GTP-U device its UEs' tunnels share. The gNB owns
// it: it is closed when the gNB terminates. It is set after the NGAP receiver has
// started, and read from there, hence atomic.
func (gnb *GNBContext) SetGtpDevice(dev *gtp.Device) {
	gnb.gtpDevice.Store(dev)
}

// GetGtpDevice returns the GTP-U device this gNB's UEs share, or nil when each UE
// has its own or there are no tunnels.
func (gnb *GNBContext) GetGtpDevice() *gtp.Device {
	return gnb.gtpDevice.Load()
}

// CloseGtpDevice waits for this gNB's UEs to give back what they hold on its shared
// GTP-U device, then removes the device. Their policy rules and routes are host-wide
// and outlive the process, so the gNB lets them go first. The wait ends once stall
// passes with no UE releasing, or after limit.
func (gnb *GNBContext) CloseGtpDevice(stall, limit time.Duration) {
	dev := gnb.GetGtpDevice()
	if dev == nil {
		return
	}

	log.Info("[GNB][GTP] Waiting up to ", limit, " for UEs to release their tunnels on ", dev.Name())
	if held := dev.WaitIdle(stall, limit); held > 0 {
		log.Warn("[GNB][GTP] Removing shared GTP-U device ", dev.Name(), " while ", held, " UEs still hold rules on it")
	}

	dev.Close()
	log.Info("[GNB][GTP] Shared GTP-U device ", dev.Name(), " removed")
}

func (gnb *GNBContext) GetUePool() *sync.Map {
	return &gnb.uePool
}

func (gnb *GNBContext) GetPrUePool() *sync.Map {
	return &gnb.prUePool
}

func (gnb *GNBContext) DeleteGnBUe(ue *GNBUe) {
	gnb.uePool.Delete(ue.ranUeNgapId)
	gnb.prUePool.CompareAndDelete(ue.GetPrUeId(), ue)
	for _, pduSession := range ue.context.pduSession {
		if pduSession != nil {
			gnb.teidPool.Delete(pduSession.GetTeidDownlink())
		}
	}
	ue.Lock()
	if ue.gnbTx != nil {
		close(ue.gnbTx)
		ue.gnbTx = nil
	}
	ue.Unlock()
}

func (gnb *GNBContext) GetGnbUe(ranUeId int64) (*GNBUe, error) {
	ue, err := gnb.uePool.Load(ranUeId)
	if !err {
		return nil, fmt.Errorf("UE is not find in GNB UE POOL")
	}
	gnbUe, ok := ue.(*GNBUe)
	if !ok || gnbUe == nil {
		return nil, fmt.Errorf("UE is nil or invalid type in GNB UE POOL")
	}
	return gnbUe, nil
}

func (gnb *GNBContext) GetGnbUeByAmfUeId(amfUeId int64) (*GNBUe, error) {
	var found *GNBUe
	gnb.uePool.Range(func(key, value any) bool {
		ue := value.(*GNBUe)
		if ue.GetAmfUeId() == amfUeId {
			found = ue
			return false
		}
		return true
	})
	if found == nil {
		return nil, fmt.Errorf("UE is not found in GNB UE POOL using AMF UE ID")
	}
	return found, nil
}

func (gnb *GNBContext) GetGnbUeByPrUeId(pRUeId int64) (*GNBUe, error) {
	ue, err := gnb.prUePool.Load(pRUeId)
	if !err {
		return nil, fmt.Errorf("UE is not find in GNB PR UE POOL")
	}
	gnbUe, ok := ue.(*GNBUe)
	if !ok || gnbUe == nil {
		return nil, fmt.Errorf("UE is nil or invalid type in GNB PR UE POOL")
	}
	return gnbUe, nil
}

func (gnb *GNBContext) GetGnbUeByTeid(teid uint32) (*GNBUe, error) {
	ue, err := gnb.teidPool.Load(teid)
	if !err {
		return nil, fmt.Errorf("UE is not find in GNB UE POOL using TEID")
	}
	gnbUe, ok := ue.(*GNBUe)
	if !ok || gnbUe == nil {
		return nil, fmt.Errorf("UE is nil or invalid type in GNB UE POOL using TEID")
	}
	return gnbUe, nil
}

func (gnb *GNBContext) NewGnBAmf(ipPort netip.AddrPort) *GNBAmf {

	// TODO if necessary add more information for AMF.

	amf := &GNBAmf{}

	// set id for AMF.
	amfId := gnb.getRanAmfId()
	amf.setAmfId(amfId)

	// set AMF ip and AMF port.
	amf.SetAmfIpPort(ipPort)

	// set state to AMF.
	amf.SetStateInactive()

	// store AMF in the AMF Pool of GNB.
	gnb.amfPool.Store(amfId, amf)

	// Plmns and slices supported by AMF initialized.
	amf.SetLenPlmns(0)
	amf.SetLenSlice(0)

	// return AMF Context
	return amf
}

func (gnb *GNBContext) IterGnbAmf() iter.Seq[*GNBAmf] {
	return func(yield func(*GNBAmf) bool) {
		gnb.amfPool.Range(func(key, value any) bool {
			return yield(value.(*GNBAmf))
		})
	}
}

func (gnb *GNBContext) FindGnbAmfByIpPort(ipPort netip.AddrPort) *GNBAmf {
	for amf := range gnb.IterGnbAmf() {
		if amf.GetAmfIpPort() == ipPort {
			return amf
		}
	}
	return nil
}

func (gnb *GNBContext) DeleteGnBAmf(amfId int64) {
	gnb.amfPool.Delete(amfId)
}

// RemoveGnbAmf stops serving through the AMF and closes its association, including one a
// re-establishment publishes concurrently.
func (gnb *GNBContext) RemoveGnbAmf(amf *GNBAmf) {
	gnb.controlInfo.lifecycle.Lock()
	gnb.DeleteGnBAmf(amf.GetAmfId())
	conn := amf.GetSCTPConn()
	gnb.controlInfo.lifecycle.Unlock()

	if conn != nil {
		_ = conn.Close()
	}
}

// PublishAssociation makes conn the AMF's association, unless the gNB has been terminated or
// no longer serves through the AMF; the caller then owns conn and must close it.
func (gnb *GNBContext) PublishAssociation(amf *GNBAmf, conn *sctp.SCTPConn) bool {
	gnb.controlInfo.lifecycle.Lock()
	defer gnb.controlInfo.lifecycle.Unlock()

	if gnb.IsTerminated() || !gnb.HasGnbAmf(amf.GetAmfId()) {
		return false
	}
	amf.SetSCTPConn(conn)
	gnb.SetN2(conn)
	return true
}

// HasGnbAmf reports whether the AMF is still one this gNB serves through, as opposed to
// one InitGnb abandoned or an AMF Configuration Update removed.
func (gnb *GNBContext) HasGnbAmf(amfId int64) bool {
	_, ok := gnb.amfPool.Load(amfId)
	return ok
}

func (gnb *GNBContext) selectAmFByCapacity() *GNBAmf {
	var amfSelect *GNBAmf
	var maxWeightFactor int64 = -1
	for amf := range gnb.IterGnbAmf() {
		if amf.relativeAmfCapacity > 0 {
			if maxWeightFactor < amf.tnla.tnlaWeightFactor {
				// select AMF
				maxWeightFactor = amf.tnla.tnlaWeightFactor
				amfSelect = amf
			}
		}
	}
	return amfSelect
}

func (gnb *GNBContext) selectAmFByActive() *GNBAmf {
	var amfSelect *GNBAmf
	var maxWeightFactor int64 = -1
	for amf := range gnb.IterGnbAmf() {
		if weight := amf.GetTNLAWeight(); amf.GetState() == Active && maxWeightFactor < weight {
			maxWeightFactor = weight
			amfSelect = amf
		}
	}
	return amfSelect
}

func (gnb *GNBContext) getRanUeId() int64 {
	return gnb.idUeGenerator.Add(1)
}

func (gnb *GNBContext) GetUeTeid(ue *GNBUe) uint32 {
	id := gnb.teidGenerator.Add(1)
	gnb.teidPool.Store(id, ue)
	return id
}

// for AMFs Pools.
func (gnb *GNBContext) getRanAmfId() int64 {
	return gnb.idAmfGenerator.Add(1)
}

func (gnb *GNBContext) SetN2(n2 *sctp.SCTPConn) {
	gnb.controlInfo.n2.Store(n2)
}

func (gnb *GNBContext) GetN2() *sctp.SCTPConn {
	return gnb.controlInfo.n2.Load()
}

// IsTerminated reports whether Terminate has been called, so that the loss of an association
// the gNB closed itself is not mistaken for one to re-establish.
func (gnb *GNBContext) IsTerminated() bool {
	return gnb.controlInfo.terminated.Load()
}

// ReleaseUesOfAmf deletes the context of every UE served through the given AMF and returns how
// many there were. A UE-associated logical NG-connection does not outlive its TNL association,
// so once the association is gone these contexts refer to nothing on the AMF side.
func (gnb *GNBContext) ReleaseUesOfAmf(amfId int64) int {
	released := 0
	gnb.uePool.Range(func(_, value any) bool {
		if ue, ok := value.(*GNBUe); ok && ue.GetAmfId() == amfId {
			gnb.DeleteGnBUe(ue)
			released++
		}
		return true
	})
	return released
}

func (gnb *GNBContext) setGnbId(id string) {
	gnb.controlInfo.gnbId = id
}

func (gnb *GNBContext) setTac(tac string) {
	gnb.controlInfo.tac = tac
}

func (gnb *GNBContext) setMnc(mnc string) {
	gnb.controlInfo.mnc = mnc
}

func (gnb *GNBContext) setMcc(mcc string) {
	gnb.controlInfo.mcc = mcc
}

func (gnb *GNBContext) GetGnbId() string {
	return gnb.controlInfo.gnbId
}

func (gnb *GNBContext) GetGnbIpPort() netip.AddrPort {
	return gnb.controlInfo.gnbIpPort
}

func (gnb *GNBContext) AddPagedUE(tmsi *ngapType.FiveGSTMSI) {
	gnb.pagedUELock.Lock()
	defer gnb.pagedUELock.Unlock()

	pagedUE := PagedUE{
		FiveGSTMSI: tmsi,
		Timestamp:  time.Now(),
	}
	gnb.pagedUEs = append(gnb.pagedUEs, pagedUE)

	go func() {
		time.Sleep(time.Second)
		gnb.pagedUELock.Lock()
		i := slices.Index(gnb.pagedUEs, pagedUE)
		if i == -1 {
			return
		}
		gnb.pagedUEs = slices.Delete(gnb.pagedUEs, i, i)
		gnb.pagedUELock.Unlock()
	}()
}

func (gnb *GNBContext) GetPagedUEs() []PagedUE {
	gnb.pagedUELock.Lock()
	defer gnb.pagedUELock.Unlock()

	return gnb.pagedUEs[:]
}

func (gnb *GNBContext) GetGnbIdInBytes() []byte {
	// changed for bytes.
	resu, err := hex.DecodeString(gnb.controlInfo.gnbId)
	if err != nil {
		fmt.Println(err)
	}
	return resu
}

func (gnb *GNBContext) getTac() string {
	return gnb.controlInfo.tac
}

func (gnb *GNBContext) GetTacInBytes() []byte {
	// changed for bytes.
	resu, err := hex.DecodeString(gnb.controlInfo.tac)
	if err != nil {
		fmt.Println(err)
	}
	return resu
}

func (gnb *GNBContext) getSlice() (string, string) {
	return gnb.sliceInfo.sst, gnb.sliceInfo.sd
}

func (gnb *GNBContext) GetSliceInBytes() ([]byte, []byte) {
	sstBytes, err := hex.DecodeString(gnb.sliceInfo.sst)
	if err != nil {
		fmt.Println(err)
	}

	if gnb.sliceInfo.sd != "" {
		sdBytes, err := hex.DecodeString(gnb.sliceInfo.sd)
		if err != nil {
			fmt.Println(err)
		}
		return sstBytes, sdBytes
	}
	return sstBytes, nil
}

func (gnb *GNBContext) GetPLMNIdentity() *ngapType.PLMNIdentity {
	return &ngapType.PLMNIdentity{Value: gnb.GetMccAndMncInOctets()}
}

func (gnb *GNBContext) GetNRCellIdentity() *ngapType.NRCellIdentity {
	nci := gnb.GetGnbIdInBytes()
	var slice = make([]byte, 2)

	return &ngapType.NRCellIdentity{
		Value: aper.BitString{
			Bytes:     append(nci, slice...),
			BitLength: 36,
		},
	}
}

func (gnb *GNBContext) GetMccAndMnc() (string, string) {
	return gnb.controlInfo.mcc, gnb.controlInfo.mnc
}

func (gnb *GNBContext) GetMccAndMncInOctets() []byte {
	mcc, mnc := gnb.GetMccAndMnc()
	if len(mcc) != 3 || (len(mnc) != 2 && len(mnc) != 3) {
		log.Error("[GNB] MCC must have three digits and MNC two or three digits")
		return nil
	}
	for _, digit := range mcc + mnc {
		if digit < '0' || digit > '9' {
			log.Error("[GNB] MCC and MNC must contain decimal digits")
			return nil
		}
	}
	// Use the NAS PLMN codec for NGAP too: MNC digit 3 belongs in the high
	// nibble of octet 2, and a two-digit MNC uses the filler nibble 0xf.
	plmn := nasType.PlmnId{MCC: mcc, MNC: mnc}
	encoded := make([]byte, nasType.PlmnIdPktSz)
	if err := plmn.MarshalBinary(encoded); err != nil {
		log.Errorf("[GNB] Could not encode PLMN: %v", err)
		return nil
	}
	return encoded
}

func (gnb *GNBContext) Terminate() {
	gnb.controlInfo.lifecycle.Lock()
	gnb.controlInfo.terminated.Store(true)
	// N2 is also the conn of the AMF that published it; collect each conn once.
	conns := []*sctp.SCTPConn{gnb.GetN2()}
	for amf := range gnb.IterGnbAmf() {
		if conn := amf.GetSCTPConn(); !slices.Contains(conns, conn) {
			conns = append(conns, conn)
		}
	}
	gnb.controlInfo.lifecycle.Unlock()

	// close all connections
	close(gnb.GetInboundChannel())
	log.Info("[GNB][UE] NAS channel Terminated")

	for _, conn := range conns {
		if conn != nil {
			log.Info("[GNB][AMF] N2/TNLA Terminated")
			conn.Close()
		}
	}

	gnb.CloseGtpDevice(0, 0)

	log.Info("GNB Terminated")
}
