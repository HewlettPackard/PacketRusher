package context

import (
	"fmt"
	"github.com/ishidawataru/sctp"
)

type UEContext struct {
	RanUeNgapId int64
	AmfUeNgapId int64
	// Snssai               models.Snssai
	PduSession           PDUSession
	AmfId                int64
	State                int64
	sctpConnection       sctp.SCTPConn
	unixSocketConnection string // TODO implement type for unix socket
}

type PDUSession struct {
	PduSessionId int64
	UeIp         string
	UplinkTeid   uint8
	DownlinkTeid uint8
}

func (ue *UEContext) GetUeTeid() uint8 {
	return ue.PduSession.UplinkTeid
}

func (ue *UEContext) SetUeTeid(teid uint8) {
	ue.PduSession.UplinkTeid = teid
}

func (ue *UEContext) SetIp(ip [12]uint8) {
	ue.PduSession.UeIp = fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

func (ue *UEContext) GetIp() string {
	return ue.PduSession.UeIp
}

func (ue *UEContext) SetAmfNgapId(amfUeId int64) {
	// fmt.Println(amfUeId)
	ue.AmfUeNgapId = amfUeId
}
