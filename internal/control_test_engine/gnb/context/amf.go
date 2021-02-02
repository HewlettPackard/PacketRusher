package context

import (
	"github.com/ishidawataru/sctp"
)

// AMF main states in the GNB Context.
const Inactive = 0x00
const Active = 0x01
const Overload = 0x02

type GNBAmf struct {
	amfIp               string         // AMF ip
	amfPort             int            // AMF port
	amfId               int64          // AMF id
	tnla                TNLAssociation // AMF sctp associations
	relativeAmfCapacity int64          // AMF capacity
	state               int
	name                string // amf name.
	regionId            byte
	setId               byte
	pointer             byte
	plmns               *PlmnSupported
	slices              *SliceSupported
	lenSlice            int
	lenPlmn             int
	// TODO implement the other fields of the AMF Context
}

type TNLAssociation struct {
	sctpConn         *sctp.SCTPConn
	tnlaWeightFactor int64
	usage            bool
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

	mov := amf.slices
	for i := 0; i < index; i++ {
		mov = mov.next
	}

	return mov.sst, mov.sd
}

func (amf *GNBAmf) GetPlmnSupport(index int) (string, string) {

	mov := amf.plmns
	for i := 0; i < index; i++ {
		mov = mov.next
	}

	return mov.mcc, mov.mnc
}

func convertMccMnc(plmn string) (mcc string, mnc string) {

	// mcc digit 1 e 2 -- invert 0 1 string
	mcc12 := string(plmn[0:2])

	// mnc digit 3
	mnc3 := string(plmn[2])

	// mcc digit 3
	mcc3 := string(plmn[3])

	// mnc digit 1 2.
	mnc12 := string(plmn[4:])

	// make mcc and mnc.
	if mcc3 != "f" {
		mcc = reverse(mcc12) + mcc3
	} else {
		mcc = reverse(mcc12)
	}

	if mnc3 != "f" {
		mnc = reverse(mnc12) + mnc3
	} else {
		mnc = reverse(mnc12)
	}

	return mcc, mnc
}

func (amf *GNBAmf) AddedPlmn(plmn string) {

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

func (amf *GNBAmf) getTNLAs() TNLAssociation {
	return amf.tnla
}

func (amf *GNBAmf) SetStateInactive() {
	amf.state = Inactive
}

func (amf *GNBAmf) SetStateActive() {
	amf.state = Active
}

func (amf *GNBAmf) SetStateOverload() {
	amf.state = Overload
}

func (amf *GNBAmf) GetState() int {
	return amf.state
}

func (amf *GNBAmf) GetSCTPConn() *sctp.SCTPConn {
	return amf.tnla.sctpConn
}

func (amf *GNBAmf) SetSCTPConn(conn *sctp.SCTPConn) {
	amf.tnla.sctpConn = conn
}

func (amf *GNBAmf) setTNLAWeight(weight int64) {
	amf.tnla.tnlaWeightFactor = weight
}

func (amf *GNBAmf) setTNLAUsage(usage bool) {
	amf.tnla.usage = usage
}

func (amf *GNBAmf) SetTNLAStreams(streams uint16) {
	amf.tnla.streams = streams
}

func (amf *GNBAmf) GetTNLAStreams() uint16 {
	return amf.tnla.streams
}

func (amf *GNBAmf) GetAmfIp() string {
	return amf.amfIp
}

func (amf *GNBAmf) SetAmfIp(ip string) {
	amf.amfIp = ip
}

func (amf *GNBAmf) GetAmfPort() int {
	return amf.amfPort
}

func (amf *GNBAmf) setAmfPort(port int) {
	amf.amfPort = port
}

func (amf *GNBAmf) GetAmfId() int64 {
	return amf.amfId
}

func (amf *GNBAmf) setAmfId(id int64) {
	amf.amfId = id
}

func (amf *GNBAmf) GetAmfName() string {
	return amf.name
}

func (amf *GNBAmf) SetAmfName(name string) {
	amf.name = name
}

func (amf *GNBAmf) GetAmfCapacity() int64 {
	return amf.relativeAmfCapacity
}

func (amf *GNBAmf) SetAmfCapacity(capacity int64) {
	amf.relativeAmfCapacity = capacity
}

func (amf *GNBAmf) GetLenPlmns() int {
	return amf.lenPlmn
}

func (amf *GNBAmf) GetLenSlice() int {
	return amf.lenSlice
}

func (amf *GNBAmf) SetLenPlmns(value int) {
	amf.lenPlmn = value
}

func (amf *GNBAmf) SetLenSlice(value int) {
	amf.lenSlice = value
}
