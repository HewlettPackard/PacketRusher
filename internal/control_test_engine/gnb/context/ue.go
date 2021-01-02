package context

import (
	"fmt"
	"github.com/ishidawataru/sctp"
	"net"
)

type GNBUe struct {
	ranUeNgapId int64 // Identifier for UE in GNB Context.
	amfUeNgapId int64 // Identifier for UE in AMF Context.
	// Snssai            models.Snssai        // slice requested
	pduSession           PDUSession     // PDU Session.
	amfId                int64          // Identifier for AMF in UE/GNB Context.
	state                int            // State of UE in NAS/GNB Context.
	sctpConnection       *sctp.SCTPConn // Sctp association in using by the UE.
	unixSocketConnection net.Conn       // Unix sockets association in using by the UE.
}

type PDUSession struct {
	pduSessionId int64
	ueIp         string
	uplinkTeid   uint8
	downlinkTeid uint8
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

func (ue *GNBUe) SetState(state int) {
	ue.state = state
}

func (ue *GNBUe) GetUnixSocket() net.Conn {
	return ue.unixSocketConnection
}

func (ue *GNBUe) SetUnixSocket(conn net.Conn) {
	ue.unixSocketConnection = conn
}

func (ue *GNBUe) GetPduSessionId() int64 {
	return ue.pduSession.pduSessionId
}

func (ue *GNBUe) SetPduSessionId(id int64) {
	ue.pduSession.pduSessionId = id
}

func (ue *GNBUe) GetTeidUplink() uint8 {
	return ue.pduSession.uplinkTeid
}

func (ue *GNBUe) SetTeidUplink(teidUplink uint8) {
	ue.pduSession.uplinkTeid = teidUplink
}

func (ue *GNBUe) GetTeidDownlink() uint8 {
	return ue.pduSession.downlinkTeid
}

func (ue *GNBUe) SetTeidDownlink(teidDownlink uint8) {
	ue.pduSession.uplinkTeid = teidDownlink
}

func (ue *GNBUe) GetRanUeId() int64 {
	return ue.ranUeNgapId
}

func (ue *GNBUe) SetRanUeId(id int64) {
	ue.ranUeNgapId = id
}

func (ue *GNBUe) SetIp(ip [12]uint8) {
	ue.pduSession.ueIp = fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

func (ue *GNBUe) GetIp() string {
	return ue.pduSession.ueIp

}

func (ue *GNBUe) GetAmfUeId() int64 {
	// fmt.Println(amfUeId)
	return ue.amfUeNgapId
}

func (ue *GNBUe) SetAmfUeId(amfUeId int64) {
	// fmt.Println(amfUeId)
	ue.amfUeNgapId = amfUeId
}
