/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"errors"
	"math"
	"my5G-RANTester/test/aio5gc/lib/state"
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
	ueIdMutex     sync.Mutex
)

type AMFContext struct {
	amfName             string
	id                  string
	supportedPlmnSnssai []models.PlmnSnssai
	servedGuami         []models.Guami
	relativeCapacity    int64
	gnbs                map[string]*GNBContext
	ues                 []*UEContext
	securityContext     []SecurityContext
	idUeGenerator       int64
	networkName         NetworkName
}

type NetworkName struct {
	Full  string
	Short string
}

func init() {
	tmsiGenerator = idgenerator.NewGenerator(1, math.MaxInt32)
}

func (c *AMFContext) NewAmfContext(amfName string, id string, supportedPlmnSnssai []models.PlmnSnssai, servedGuami []models.Guami, relativeCapacity int64) {
	c.amfName = amfName
	c.id = id
	c.supportedPlmnSnssai = supportedPlmnSnssai
	c.servedGuami = servedGuami
	c.relativeCapacity = relativeCapacity
	c.gnbs = make(map[string]*GNBContext)
	c.ues = []*UEContext{}
	c.securityContext = []SecurityContext{}
	c.idUeGenerator = 0
	c.networkName = NetworkName{
		Full:  "NtwFull",
		Short: "Ntwshrt",
	}
}

func (c *AMFContext) TmsiAllocate() int32 {
	tmsi, err := tmsiGenerator.Allocate()
	if err != nil {
		log.Errorf("[5GC] Allocate TMSI error: %+v", err)
		return -1
	}
	return int32(tmsi)
}

func (c *AMFContext) GetName() string {
	return c.amfName
}

func (c *AMFContext) GetId() string {
	return c.id
}

func (c *AMFContext) FindSecurityContextByMsin(msin string) (SecurityContext, error) {
	scMutex.Lock()
	defer scMutex.Unlock()
	for sub := range c.securityContext {
		if c.securityContext[sub].msin == msin {
			return c.securityContext[sub], nil
		}
	}
	return SecurityContext{}, errors.New("[5GC] UE with msin " + msin + "not found")
}

func (c *AMFContext) FindUEById(id int64) (*UEContext, error) {
	ueMutex.Lock()
	defer ueMutex.Unlock()
	for ue := range c.ues {
		if c.ues[ue].amfNgapId == id {
			return c.ues[ue], nil
		}
	}
	return nil, errors.New("[5GC] UE with amfNgapId " + strconv.Itoa(int(id)) + "not found")
}

func (c *AMFContext) FindUEByRanId(id int64) (*UEContext, error) {
	ueMutex.Lock()
	defer ueMutex.Unlock()
	for ue := range c.ues {
		if c.ues[ue].ranNgapId == id {
			return c.ues[ue], nil
		}
	}

	return nil, errors.New("[5GC] UE with RanNgapId " + strconv.Itoa(int(id)) + "not found")
}

func (c *AMFContext) FindRegistredUEByMsin(msin string) (*UEContext, error) {
	ueMutex.Lock()
	defer ueMutex.Unlock()
	for ue := range c.ues {
		if c.ues[ue].securityContext.msin == msin && c.ues[ue].initialContextSetup {
			return c.ues[ue], nil
		}
	}
	return nil, errors.New("[5GC] Registred UE with msin " + msin + "not found")
}

func (c *AMFContext) ExecuteForAllUe(function func(ue *UEContext)) {
	ueMutex.Lock()
	defer ueMutex.Unlock()
	for ue := range c.ues {
		function(c.ues[ue])
	}
}

func (c *AMFContext) NewSecurityContext(sub SecurityContext) error {
	_, notExist := c.FindSecurityContextByMsin(sub.msin)
	if notExist == nil {
		return errors.New("[5GC] Cannot create new subscriber: subscriber with msin " + sub.msin + " already exist")
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
	newUE.smContexts = make(map[int32]*SmContext)
	newUE.state = &state.UEState{}
	newUE.state.Init()
	ueMutex.Lock()
	c.ues = append(c.ues, &newUE)
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

func (c *AMFContext) GetGnb(Addr string) (*GNBContext, error) {
	gnbMutex.Lock()
	gnb, exist := c.gnbs[Addr]
	gnbMutex.Unlock()
	if !exist {
		return gnb, errors.New("GNB with address " + Addr + " not found in AMF")
	}
	return gnb, nil
}

func (c *AMFContext) FindGnbById(globalRanNodeID models.GlobalRanNodeId) (GNBContext, error) {
	gnbMutex.Lock()
	defer gnbMutex.Unlock()
	for _, gnb := range c.gnbs {
		if gnb.globalRanNodeID == globalRanNodeID {
			return *gnb, nil
		}
	}
	return GNBContext{}, errors.New("GNB with matching global RanNode ID not found in AMF")

}

func (c *AMFContext) AddGnb(gnbAddr string, gnb *GNBContext) error {
	gnbMutex.Lock()
	c.gnbs[gnbAddr] = gnb
	gnbMutex.Unlock()
	return nil
}

func (c *AMFContext) getAmfUeId() int64 {
	ueIdMutex.Lock()
	defer ueIdMutex.Unlock()
	id := c.idUeGenerator

	// increment UeId
	c.idUeGenerator++

	return id
}

func (c *AMFContext) GetNetworkName() NetworkName {
	return c.networkName
}
