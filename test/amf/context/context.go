package context

import (
	"errors"
	"math"
	"strconv"
	"sync"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/idgenerator"
	log "github.com/sirupsen/logrus"
)

var (
	tmsiGenerator *idgenerator.IDGenerator = nil
	ueMutex       sync.Mutex
	scMutex       sync.Mutex
	gnbMutex      sync.Mutex
)

type AMFContext struct {
	name                string
	id                  string
	supportedPlmnSnssai []models.PlmnSnssai
	servedGuami         []models.Guami
	relativeCapacity    int64
	gnbs                map[string]GNBContext
	ues                 []UEContext
	securityContext     []SecurityContext
	idUeGenerator       int64
}

func init() {
	tmsiGenerator = idgenerator.NewGenerator(1, math.MaxInt32)
}

func (c *AMFContext) NewAmfContext(name string, id string, supportedPlmnSnssai []models.PlmnSnssai, servedGuami []models.Guami, relativeCapacity int64) {
	c.name = name
	c.id = id
	c.supportedPlmnSnssai = supportedPlmnSnssai
	c.servedGuami = servedGuami
	c.relativeCapacity = relativeCapacity
	c.gnbs = make(map[string]GNBContext)
	c.ues = []UEContext{}
	c.securityContext = []SecurityContext{}
	c.idUeGenerator = 0
}

func (c *AMFContext) TmsiAllocate() int32 {
	tmsi, err := tmsiGenerator.Allocate()
	if err != nil {
		log.Errorf("[AMF] Allocate TMSI error: %+v", err)
		return -1
	}
	return int32(tmsi)
}

func (c *AMFContext) GetName() string {
	return c.name
}

func (c *AMFContext) GetId() string {
	return c.id
}

func (c *AMFContext) FindSecurityContextByMsin(msin string) (SecurityContext, error) {
	scMutex.Lock()
	for sub := range c.securityContext {
		if c.securityContext[sub].msin == msin {
			scMutex.Unlock()
			return c.securityContext[sub], nil
		}
	}
	scMutex.Unlock()
	return SecurityContext{}, errors.New("[AMF] UE with msin " + msin + "not found")
}

func (c *AMFContext) FindUEById(id int64) (*UEContext, error) {
	ueMutex.Lock()
	for ue := range c.ues {
		if c.ues[ue].amfNgapId == id {
			ueMutex.Unlock()
			return &c.ues[ue], nil
		}
	}
	ueMutex.Unlock()

	return &UEContext{}, errors.New("[AMF] UE with amfNgapId " + strconv.Itoa(int(id)) + "not found")
}

func (c *AMFContext) FindUEByRanId(id int64) (*UEContext, error) {
	ueMutex.Lock()
	for ue := range c.ues {
		if c.ues[ue].ranNgapId == id {
			ueMutex.Unlock()
			return &c.ues[ue], nil
		}
	}
	ueMutex.Unlock()

	return &UEContext{}, errors.New("[AMF] UE with RanNgapId " + strconv.Itoa(int(id)) + "not found")
}

func (c *AMFContext) NewSecurityContext(sub SecurityContext) error {
	_, notExist := c.FindSecurityContextByMsin(sub.msin)
	if notExist == nil {
		return errors.New("[AMF] Cannot create new subscriber: subscriber with msin " + sub.msin + " already exist")
	}
	scMutex.Lock()
	c.securityContext = append(c.securityContext, sub)
	scMutex.Unlock()
	return nil
}

func (c *AMFContext) NewUE(ueRanNgapId int64) *UEContext {
	newUE := UEContext{}
	newUE.SetRanNgapId(ueRanNgapId)
	newUE.SetAmfNgapId(c.getAmfUeId())
	newUE.SecurityContextAvailable = false
	ueMutex.Lock()
	c.ues = append(c.ues, newUE)
	ueMutex.Unlock()
	ue, _ := c.FindUEById(newUE.amfNgapId)
	return ue
}

func (c *AMFContext) GetServedGuami() []models.Guami {
	return c.servedGuami
}

func (c *AMFContext) GetServedGuamiPlmns(plmnIds []models.PlmnId) []models.Guami {
	guamis := []models.Guami{}
	for i := range c.servedGuami {
		for j := range plmnIds {
			if *c.servedGuami[i].PlmnId == plmnIds[j] {
				guamis = append(guamis, c.servedGuami[i])
			}
		}
	}
	return guamis
}

func (c *AMFContext) GetSupportedPlmnSnssai() []models.PlmnSnssai {
	return c.supportedPlmnSnssai
}

func (c *AMFContext) GetRelativeCapacity() int64 {
	return c.relativeCapacity
}

func (c *AMFContext) GetGnb(Addr string) (GNBContext, error) {
	gnbMutex.Lock()
	gnb, exist := c.gnbs[Addr]
	gnbMutex.Unlock()
	if !exist {
		return gnb, errors.New("GNB with address " + Addr + " not found in AMF")
	}
	return gnb, nil
}

func (c *AMFContext) AddGnb(gnbAddr string, gnb GNBContext) error {
	gnbMutex.Lock()
	c.gnbs[gnbAddr] = gnb
	gnbMutex.Unlock()
	return nil
}

func (c *AMFContext) getAmfUeId() int64 {

	// TODO implement mutex

	id := c.idUeGenerator

	// increment RanUeId
	c.idUeGenerator++

	return id
}
