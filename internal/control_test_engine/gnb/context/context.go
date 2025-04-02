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
	"time"

	"github.com/free5gc/aper"
	"github.com/free5gc/nas/nasType"
	"github.com/free5gc/ngap/ngapConvert"
	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/openapi/models"
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
	idUeGenerator  int64  // ran UE id.
	idAmfGenerator int64  // ran amf id
	teidGenerator  uint32 // ran UE downlink Teid
	ueIpGenerator  uint8  // ran ue ip.
	pagedUEs       []PagedUE
	pagedUELock    sync.Mutex
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
	n2             *sctp.SCTPConn
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
	gnb.controlInfo.inboundChannel = make(chan UEMessage, 1)
	gnb.sliceInfo.sd = sd
	gnb.sliceInfo.sst = sst
	gnb.idUeGenerator = 1
	gnb.idAmfGenerator = 1
	gnb.controlInfo.gnbIpPort = n2
	gnb.teidGenerator = 1
	gnb.ueIpGenerator = 3
	gnb.dataInfo.gnbIpPort = n3
}

func (gnb *GNBContext) NewGnBUe(gnbTx chan UEMessage, gnbRx chan UEMessage, prUeId int64, tmsi *nasType.GUTI5G) (*GNBUe, error) {

	// TODO if necessary add more information for UE.
	// TODO implement mutex

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

	// store UE in the UE Pool of GNB.
	gnb.uePool.Store(ranId, ue)
	if prUeId != 0 {
		gnb.prUePool.Store(prUeId, ue)
	}

	// select AMF with Capacity is more than 0.
	amf := gnb.selectAmFByActive()
	if amf == nil {
		return nil, errors.New("no AMF available for this UE")
	}

	// set amfId and SCTP association for UE.
	ue.SetAmfId(amf.GetAmfId())
	ue.SetSCTP(amf.GetSCTPConn())

	// return UE Context.
	return ue, nil
}

func (gnb *GNBContext) GetInboundChannel() chan UEMessage {
	return gnb.controlInfo.inboundChannel
}

func (gnb *GNBContext) GetN3GnbIp() netip.Addr {
	return gnb.dataInfo.gnbIpPort.Addr()
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
	return ue.(*GNBUe), nil
}

func (gnb *GNBContext) GetGnbUeByPrUeId(pRUeId int64) (*GNBUe, error) {
	ue, err := gnb.prUePool.Load(pRUeId)
	if !err {
		return nil, fmt.Errorf("UE is not find in GNB PR UE POOL")
	}
	return ue.(*GNBUe), nil
}

func (gnb *GNBContext) GetGnbUeByTeid(teid uint32) (*GNBUe, error) {
	ue, err := gnb.teidPool.Load(teid)
	if !err {
		return nil, fmt.Errorf("UE is not find in GNB UE POOL using TEID")
	}
	return ue.(*GNBUe), nil
}

func (gnb *GNBContext) NewGnBAmf(ipPort netip.AddrPort) *GNBAmf {

	// TODO if necessary add more information for AMF.
	// TODO implement mutex

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
		if amf.GetState() == Active && maxWeightFactor < amf.tnla.tnlaWeightFactor {
			maxWeightFactor = amf.tnla.tnlaWeightFactor
			amfSelect = amf
		}
	}
	return amfSelect
}

func (gnb *GNBContext) getRanUeId() int64 {

	// TODO implement mutex

	id := gnb.idUeGenerator

	// increment RanUeId
	gnb.idUeGenerator++

	return id
}

func (gnb *GNBContext) GetUeTeid(ue *GNBUe) uint32 {

	// TODO implement mutex

	id := gnb.teidGenerator

	// store UE in the TEID Pool of GNB.
	gnb.teidPool.Store(id, ue)

	// increment UE teid.
	gnb.teidGenerator++

	return id
}

// for AMFs Pools.
func (gnb *GNBContext) getRanAmfId() int64 {

	// TODO implement mutex

	id := gnb.idAmfGenerator

	// increment Amf Id
	gnb.idAmfGenerator++

	return id
}

func (gnb *GNBContext) SetN2(n2 *sctp.SCTPConn) {
	gnb.controlInfo.n2 = n2
}

func (gnb *GNBContext) GetN2() *sctp.SCTPConn {
	return gnb.controlInfo.n2
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

func (gnb *GNBContext) GetPLMNIdentity() ngapType.PLMNIdentity {
	return ngapConvert.PlmnIdToNgap(models.PlmnId{Mcc: gnb.controlInfo.mcc, Mnc: gnb.controlInfo.mnc})
}

func (gnb *GNBContext) GetNRCellIdentity() ngapType.NRCellIdentity {
	nci := gnb.GetGnbIdInBytes()
	var slice = make([]byte, 2)

	return ngapType.NRCellIdentity{
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
	var res string

	// reverse mcc and mnc
	mcc := reverse(gnb.controlInfo.mcc)
	mnc := reverse(gnb.controlInfo.mnc)

	if len(mnc) == 2 {
		res = fmt.Sprintf("%c%cf%c%c%c", mcc[1], mcc[2], mcc[0], mnc[0], mnc[1])
	} else {
		res = fmt.Sprintf("%c%c%c%c%c%c", mcc[1], mcc[2], mnc[2], mcc[0], mnc[0], mnc[1])
	}

	resu, _ := hex.DecodeString(res)
	return resu
}

func (gnb *GNBContext) Terminate() {

	// close all connections
	close(gnb.GetInboundChannel())
	log.Info("[GNB][UE] NAS channel Terminated")

	n2 := gnb.GetN2()
	if n2 != nil {
		log.Info("[GNB][AMF] N2/TNLA Terminated")
		n2.Close()
	}

	log.Info("GNB Terminated")
}

func reverse(s string) string {
	// reverse string.
	var aux string
	for _, valor := range s {
		aux = string(valor) + aux
	}
	return aux
}
