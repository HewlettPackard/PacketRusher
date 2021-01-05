package context

import (
	"github.com/ishidawataru/sctp"
)

type GNBAmf struct {
	amfIp               string         // AMF ip
	amfPort             int            // AMF port
	amfId               int64          // AMF id
	tnla                TNLAssociation // AMF sctp associations
	relativeAmfCapacity int64          // AMF capacity
	state               int
	// TODO implement the other fields of the AMF Context
}

type TNLAssociation struct {
	sctpConn         *sctp.SCTPConn
	tnlaWeightFactor int64
	usage            bool
	streams          uint16
}

func (amf *GNBAmf) getTNLAs() TNLAssociation {
	return amf.tnla
}

func (amf *GNBAmf) SetState(state int) {
	amf.state = state
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

func (amf *GNBAmf) getAmfCapacity() int64 {
	return amf.relativeAmfCapacity
}

func (amf *GNBAmf) setAmfCapacity(capacity int64) {
	amf.relativeAmfCapacity = capacity
}
