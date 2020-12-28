package context

import (
	"github.com/ishidawataru/sctp"
)

type AMFContext struct {
	amfIp               string            // AMF ip
	amfPort             string            // AMF port
	amfId               int64             // AMF id
	tnla                []*TNLAssociation // AMF sctp associations
	relativeAmfCapacity int64             // AMF capacity
	// TODO implement the other fields of the AMF Context
}

type TNLAssociation struct {
	sctpConn         sctp.SCTPConn
	tnlaWeightFactor int64
	usage            string
}
