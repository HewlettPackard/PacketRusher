package context

import (
	"github.com/ishidawataru/sctp"
)

type GNBAmf struct {
	amfIp               string         // AMF ip
	amfPort             string         // AMF port
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
	streams          int64
}

func (amf *GNBAmf) getTNLAs() TNLAssociation {
	return amf.tnla
}

func (amf *GNBAmf) setState(state int) {
	amf.state = state
}

func (amf *GNBAmf) getState() int {
	return amf.state
}

func (amf *GNBAmf) getSCTPConn() *sctp.SCTPConn {
	return amf.tnla.sctpConn
}

func (amf *GNBAmf) setSCTPConn(conn *sctp.SCTPConn) {
	amf.tnla.sctpConn = conn
}

func (amf *GNBAmf) setTNLAWeight(weight int64) {
	amf.tnla.tnlaWeightFactor = weight
}

func (amf *GNBAmf) setTNLAUsage(usage bool) {
	amf.tnla.usage = usage
}

func (amf *GNBAmf) setTNLAStreams(streams int64) {
	amf.tnla.streams = streams
}

func (amf *GNBAmf) getAmfIp() string {
	return amf.amfIp
}

func (amf *GNBAmf) setAmfIp(ip string) {
	amf.amfIp = ip
}

func (amf *GNBAmf) getAmfPort() string {
	return amf.amfPort
}

func (amf *GNBAmf) setAmfPort(port string) {
	amf.amfIp = port
}

func (amf *GNBAmf) getAmfId() int64 {
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
