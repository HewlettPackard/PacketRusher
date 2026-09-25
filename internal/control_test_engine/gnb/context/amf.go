/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"fmt"
	"net/netip"
	"sync"

	"github.com/free5gc/aper"
	"github.com/ishidawataru/sctp"
)

// AMF main states in the GNB Context.
const Inactive = 0x00
const Active = 0x01
const Overload = 0x02

type GNBAmf struct {
	// mu guards state, tnla, name, relativeAmfCapacity and the PLMN and slice lists: NG
	// Setup rewrites them on every re-established association, and AMF Configuration
	// Update rewrites the TNLA fields, while UE and dispatch goroutines read them.
	mu                  sync.RWMutex
	amfIpPort           netip.AddrPort // AMF ip and port
	amfId               int64          // AMF id
	tnla                TNLAssociation // AMF sctp associations
	relativeAmfCapacity int64          // AMF capacity
	state               int
	name                string // amf name.
	regionId            aper.BitString
	setId               aper.BitString
	pointer             aper.BitString
	plmns               *PlmnSupported
	slices              *SliceSupported
	lenSlice            int
	lenPlmn             int
	backupAMF           string
	// TODO implement the other fields of the AMF Context
}

type TNLAssociation struct {
	sctpConn         *sctp.SCTPConn
	tnlaWeightFactor int64
	usage            aper.Enumerated
	streams          uint16
}

type SliceSupported struct {
	sst    string
	sd     string
	status string
	next   *SliceSupported
}

type PlmnSupported struct {
	mcc  string
	mnc  string
	next *PlmnSupported
}

func (amf *GNBAmf) GetSliceSupport(index int) (string, string) {
	amf.mu.RLock()
	defer amf.mu.RUnlock()

	mov := amf.slices
	for i := 0; i < index && mov != nil; i++ {
		mov = mov.next
	}
	if mov == nil {
		return "", ""
	}

	return mov.sst, mov.sd
}

func (amf *GNBAmf) GetPlmnSupport(index int) (string, string) {
	amf.mu.RLock()
	defer amf.mu.RUnlock()

	mov := amf.plmns
	for i := 0; i < index && mov != nil; i++ {
		mov = mov.next
	}
	if mov == nil {
		return "", ""
	}

	return mov.mcc, mov.mnc
}

// ResetSupported empties the PLMN and slice lists before an NG Setup Response refills
// them, so that each re-established association does not append the same entries again.
func (amf *GNBAmf) ResetSupported() {
	amf.mu.Lock()
	defer amf.mu.Unlock()

	amf.plmns, amf.lenPlmn = nil, 0
	amf.slices, amf.lenSlice = nil, 0
}

func convertMccMnc(plmn string) (mcc string, mnc string) {
	if plmn[2] == 'f' {
		mcc = fmt.Sprintf("%c%c%c", plmn[1], plmn[0], plmn[3])
		mnc = fmt.Sprintf("%c%c", plmn[5], plmn[4])
	} else {
		mcc = fmt.Sprintf("%c%c%c", plmn[1], plmn[0], plmn[3])
		mnc = fmt.Sprintf("%c%c%c", plmn[2], plmn[5], plmn[4])
	}

	return mcc, mnc
}

func (amf *GNBAmf) AddedPlmn(plmn string) {
	amf.mu.Lock()
	defer amf.mu.Unlock()

	if amf.lenPlmn == 0 {
		newElem := &PlmnSupported{}

		// newElem.info = plmn
		newElem.next = nil
		newElem.mcc, newElem.mnc = convertMccMnc(plmn)
		// update list
		amf.plmns = newElem
		amf.lenPlmn++
		return
	}

	mov := amf.plmns
	for i := 0; i < amf.lenPlmn; i++ {

		// end of the list
		if mov.next == nil {

			newElem := &PlmnSupported{}
			newElem.mcc, newElem.mnc = convertMccMnc(plmn)
			newElem.next = nil

			mov.next = newElem

		} else {
			mov = mov.next
		}
	}

	amf.lenPlmn++
}

func (amf *GNBAmf) AddedSlice(sst string, sd string) {
	amf.mu.Lock()
	defer amf.mu.Unlock()

	if amf.lenSlice == 0 {
		newElem := &SliceSupported{}
		newElem.sst = sst
		newElem.sd = sd
		newElem.next = nil

		// update list
		amf.slices = newElem
		amf.lenSlice++
		return
	}

	mov := amf.slices
	for i := 0; i < amf.lenSlice; i++ {

		// end of the list
		if mov.next == nil {

			newElem := &SliceSupported{}
			newElem.sst = sst
			newElem.sd = sd
			newElem.next = nil

			mov.next = newElem

		} else {
			mov = mov.next
		}
	}
	amf.lenSlice++
}

func (amf *GNBAmf) GetTNLA() TNLAssociation {
	amf.mu.RLock()
	defer amf.mu.RUnlock()
	return amf.tnla
}
func (tnla *TNLAssociation) GetSCTP() *sctp.SCTPConn {
	return tnla.sctpConn
}

func (tnla *TNLAssociation) GetWeightFactor() int64 {
	return tnla.tnlaWeightFactor
}

func (tnla *TNLAssociation) GetUsage() aper.Enumerated {
	return tnla.usage
}

func (tnla *TNLAssociation) Release() error {
	return tnla.sctpConn.Close()
}

func (amf *GNBAmf) SetStateInactive() {
	amf.mu.Lock()
	amf.state = Inactive
	amf.mu.Unlock()
}

func (amf *GNBAmf) SetStateActive() {
	amf.mu.Lock()
	amf.state = Active
	amf.mu.Unlock()
}

func (amf *GNBAmf) SetStateOverload() {
	amf.mu.Lock()
	amf.state = Overload
	amf.mu.Unlock()
}

func (amf *GNBAmf) GetState() int {
	amf.mu.RLock()
	defer amf.mu.RUnlock()
	return amf.state
}

func (amf *GNBAmf) GetSCTPConn() *sctp.SCTPConn {
	amf.mu.RLock()
	defer amf.mu.RUnlock()
	return amf.tnla.sctpConn
}

func (amf *GNBAmf) SetSCTPConn(conn *sctp.SCTPConn) {
	amf.mu.Lock()
	amf.tnla.sctpConn = conn
	amf.mu.Unlock()
}

func (amf *GNBAmf) SetTNLAWeight(weight int64) {
	amf.mu.Lock()
	amf.tnla.tnlaWeightFactor = weight
	amf.mu.Unlock()
}

func (amf *GNBAmf) GetTNLAWeight() int64 {
	amf.mu.RLock()
	defer amf.mu.RUnlock()
	return amf.tnla.tnlaWeightFactor
}

func (amf *GNBAmf) SetTNLAUsage(usage aper.Enumerated) {
	amf.mu.Lock()
	amf.tnla.usage = usage
	amf.mu.Unlock()
}

// CloseIfNotActive closes conn if it is still this AMF's association and NG Setup has
// not completed on it, deciding under the same lock SetStateActive takes, so that a
// setup completing at the deadline is not closed after the fact.
func (amf *GNBAmf) CloseIfNotActive(conn *sctp.SCTPConn) bool {
	amf.mu.Lock()
	defer amf.mu.Unlock()

	if amf.state == Active || amf.tnla.sctpConn != conn {
		return false
	}
	_ = conn.Close()
	return true
}

func (amf *GNBAmf) SetTNLAStreams(streams uint16) {
	amf.tnla.streams = streams
}

func (amf *GNBAmf) GetTNLAStreams() uint16 {
	return amf.tnla.streams
}

func (amf *GNBAmf) GetAmfIpPort() netip.AddrPort {
	return amf.amfIpPort
}

func (amf *GNBAmf) SetAmfIpPort(ap netip.AddrPort) {
	amf.amfIpPort = ap
}

func (amf *GNBAmf) GetAmfId() int64 {
	return amf.amfId
}

func (amf *GNBAmf) setAmfId(id int64) {
	amf.amfId = id
}

func (amf *GNBAmf) GetAmfName() string {
	amf.mu.RLock()
	defer amf.mu.RUnlock()
	return amf.name
}

func (amf *GNBAmf) GetRegionId() aper.BitString {
	return amf.regionId
}

func (amf *GNBAmf) SetRegionId(regionId aper.BitString) {
	amf.regionId = regionId
}

func (amf *GNBAmf) GetSetId() aper.BitString {
	return amf.setId
}

func (amf *GNBAmf) SetSetId(setId aper.BitString) {
	amf.setId = setId
}

func (amf *GNBAmf) GetPointer() aper.BitString {
	return amf.pointer
}

func (amf *GNBAmf) SetPointer(pointer aper.BitString) {
	amf.pointer = pointer
}

func (amf *GNBAmf) SetAmfName(name string) {
	amf.mu.Lock()
	amf.name = name
	amf.mu.Unlock()
}

func (amf *GNBAmf) GetAmfCapacity() int64 {
	amf.mu.RLock()
	defer amf.mu.RUnlock()
	return amf.relativeAmfCapacity
}

func (amf *GNBAmf) SetAmfCapacity(capacity int64) {
	amf.mu.Lock()
	amf.relativeAmfCapacity = capacity
	amf.mu.Unlock()
}

func (amf *GNBAmf) GetLenPlmns() int {
	amf.mu.RLock()
	defer amf.mu.RUnlock()
	return amf.lenPlmn
}

func (amf *GNBAmf) GetLenSlice() int {
	amf.mu.RLock()
	defer amf.mu.RUnlock()
	return amf.lenSlice
}

func (amf *GNBAmf) SetLenPlmns(value int) {
	amf.mu.Lock()
	defer amf.mu.Unlock()
	amf.lenPlmn = value
}

func (amf *GNBAmf) SetLenSlice(value int) {
	amf.mu.Lock()
	defer amf.mu.Unlock()
	amf.lenSlice = value
}

func (amf *GNBAmf) GetBackupAMF() string {
	return amf.backupAMF
}

func (amf *GNBAmf) SetBackupAMF(backupAMF string) {
	amf.backupAMF = backupAMF
}
