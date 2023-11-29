package context

import (
	"errors"
	"slices"
	"sync"

	"github.com/free5gc/nas/nasConvert"
	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/openapi/models"
	"github.com/mohae/deepcopy"
	log "github.com/sirupsen/logrus"
)

type Session struct {
	Dnn []string
}

type SmContext struct {
	mu sync.RWMutex // protect the following fields

	// pdu session information
	pduSessionID int32
	smContextRef string
	snssai       models.Snssai
	dnn          string
	nsInstance   string
	userLocation models.NrLocation
	plmnID       models.PlmnId

	// for duplicate pdu session id handling
	ulNASTransport *nasMessage.ULNASTransport
	duplicated     bool
}

func NewSmContext(pduSessionID int32) *SmContext {
	c := &SmContext{pduSessionID: pduSessionID}
	return c
}

func (c *SmContext) SetSnssai(snssai models.Snssai) {
	c.snssai = snssai
}

func (c *SmContext) SetDnn(dnn string) {
	c.dnn = dnn
}

func (c *SmContext) SetUserLocation(location models.NrLocation) {
	c.userLocation = location
}

func CreatePDUSession(ulNasTransport *nasMessage.ULNASTransport,
	ue *UEContext,
	amf *AMFContext,
	pduSessionID int32,
	smMessage []uint8,
) (err error) {
	var (
		snssai models.Snssai
		dnn    string
	)
	// If the S-NSSAI IE is not included, select a default snssai
	if ulNasTransport.SNSSAI != nil {
		snssai = nasConvert.SnssaiToModels(ulNasTransport.SNSSAI)
	} else {
		snssai = ue.GetNssai()
	}

	if ulNasTransport.DNN != nil {
		if !slices.Contains(amf.GetDnnList(), ulNasTransport.DNN.GetDNN()) {
			return errors.New("[AMF] Unknown DNN requested")
		}
		dnn = ulNasTransport.DNN.GetDNN()

	} else {
		dnn = amf.GetDnnList()[0]
	}

	newSmContext := NewSmContext(pduSessionID)
	newSmContext.SetSnssai(snssai)
	newSmContext.SetDnn(dnn)

	newSmContext.SetUserLocation(deepcopy.Copy(ue.GetUserLocationInfo()).(models.NrLocation))
	ue.AddSmContext(newSmContext)
	log.Infof("[AMF] create smContext[pduSessionID: %d] Success", pduSessionID)
	return nil
}
