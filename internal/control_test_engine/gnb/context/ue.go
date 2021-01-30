package context

import (
	"github.com/ishidawataru/sctp"
	"net"
)

// UE main states in the GNB Context.
const Initialized = 0x00
const Ongoing = 0x01
const Ready = 0x02

type GNBUe struct {
	ranUeNgapId          int64          // Identifier for UE in GNB Context.
	amfUeNgapId          int64          // Identifier for UE in AMF Context.
	amfId                int64          // Identifier for AMF in UE/GNB Context.
	state                int            // State of UE in NAS/GNB Context.
	sctpConnection       *sctp.SCTPConn // Sctp association in using by the UE.
	unixSocketConnection net.Conn       // Unix sockets association in using by the UE.
	context              Context
}

type Context struct {
	mobilityInfo mobility
	maskedIMEISV string
	pduSession   PDUSession // PDU Session.
}

type PDUSession struct {
	pduSessionId int64
	ranUeIP      net.IP
	uplinkTeid   uint32
	downlinkTeid uint32
	slices       *SliceSupported
	lenSlice     int
}

type mobility struct {
	mcc string
	mnc string
}

func (ue *GNBUe) CreateUeContext(plmn string, imeisv string, sst []string, sd []string) {

	ue.context.pduSession.lenSlice = 0
	ue.context.mobilityInfo.mcc, ue.context.mobilityInfo.mnc = convertMccMnc(plmn)
	ue.context.maskedIMEISV = imeisv

	for i := 0; i < len(sst); i++ {
		ue.addedSlice(sst[i], sd[i])
	}
}

func (ue *GNBUe) GetUeMobility() (string, string) {
	return ue.context.mobilityInfo.mcc, ue.context.mobilityInfo.mnc
}

func (ue *GNBUe) GetUeMaskedImeiSv() string {
	return ue.context.maskedIMEISV
}

func (ue *GNBUe) GetLenSlice() int {
	return ue.context.pduSession.lenSlice
}

func (ue *GNBUe) addedSlice(sst string, sd string) {

	if ue.context.pduSession.lenSlice == 0 {
		newElem := &SliceSupported{}
		newElem.sst = sst
		newElem.sd = sd
		newElem.next = nil

		// update list
		ue.context.pduSession.slices = newElem
		ue.context.pduSession.lenSlice++
		return
	}

	mov := ue.context.pduSession.slices
	for i := 0; i < ue.context.pduSession.lenSlice; i++ {

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
	ue.context.pduSession.lenSlice++
}

func (ue *GNBUe) GetAllowedNSSAI(index int) (string, string) {

	mov := ue.context.pduSession.slices
	for i := 0; i < index; i++ {
		mov = mov.next
	}

	return mov.sst, mov.sd
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

func (ue *GNBUe) GetUnixSocket() net.Conn {
	return ue.unixSocketConnection
}

func (ue *GNBUe) SetUnixSocket(conn net.Conn) {
	ue.unixSocketConnection = conn
}

func (ue *GNBUe) GetPduSessionId() int64 {
	return ue.context.pduSession.pduSessionId
}

func (ue *GNBUe) SetPduSessionId(id int64) {
	ue.context.pduSession.pduSessionId = id
}

func (ue *GNBUe) GetTeidUplink() uint32 {
	return ue.context.pduSession.uplinkTeid
}

func (ue *GNBUe) SetTeidUplink(teidUplink uint32) {
	ue.context.pduSession.uplinkTeid = teidUplink
}

func (ue *GNBUe) GetTeidDownlink() uint32 {
	return ue.context.pduSession.downlinkTeid
}

func (ue *GNBUe) SetTeidDownlink(teidDownlink uint32) {
	ue.context.pduSession.downlinkTeid = teidDownlink
}

func (ue *GNBUe) GetRanUeId() int64 {
	return ue.ranUeNgapId
}

func (ue *GNBUe) SetRanUeId(id int64) {
	ue.ranUeNgapId = id
}

func (ue *GNBUe) SetIp(ueIp uint8) {
	ue.context.pduSession.ranUeIP = net.IPv4(127, 0, 0, ueIp)
}

func (ue *GNBUe) GetIp() net.IP {
	return ue.context.pduSession.ranUeIP

}

func (ue *GNBUe) GetAmfUeId() int64 {
	// fmt.Println(amfUeId)
	return ue.amfUeNgapId
}

func (ue *GNBUe) SetAmfUeId(amfUeId int64) {
	// fmt.Println(amfUeId)
	ue.amfUeNgapId = amfUeId
}
