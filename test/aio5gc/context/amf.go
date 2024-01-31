/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"errors"
	"math"
	"strconv"
	"sync"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/fsm"
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
	idUeGenerator       int64
	networkName         NetworkName
	provisionedData     map[string]provisionedData
	ueFsm               *fsm.FSM
	pduFsm              *fsm.FSM
}

type NetworkName struct {
	Full  string
	Short string
}

func init() {
	tmsiGenerator = idgenerator.NewGenerator(1, math.MaxInt32)
}

func (c *AMFContext) NewAmfContext(amfName string, id string, supportedPlmnSnssai []models.PlmnSnssai, servedGuami []models.Guami, relativeCapacity int64, ueFsm *fsm.FSM, pduFsm *fsm.FSM) {
	c.amfName = amfName
	c.id = id
	c.supportedPlmnSnssai = supportedPlmnSnssai
	c.servedGuami = servedGuami
	c.relativeCapacity = relativeCapacity
	c.gnbs = make(map[string]*GNBContext)
	c.ues = []*UEContext{}
	c.provisionedData = map[string]provisionedData{}
	c.idUeGenerator = 0
	c.networkName = NetworkName{
		Full:  "NtwFull",
		Short: "Ntwshrt",
	}
	c.ueFsm = ueFsm
	c.pduFsm = pduFsm
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

func (c *AMFContext) FindProvisionedData(msin string) (provisionedData, error) {
	scMutex.Lock()
	defer scMutex.Unlock()
	data, ok := c.provisionedData[msin]
	if !ok {
		return provisionedData{}, errors.New("[5GC] UE with msin " + msin + "not found")
	}
	return data, nil
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

func (c *AMFContext) FindRegisteredUEByMsin(msin string) (*UEContext, error) {
	ueMutex.Lock()
	defer ueMutex.Unlock()
	for ue := range c.ues {
		if c.ues[ue].securityContext.msin == msin && c.ues[ue].GetState().Is(Registered) {
			return c.ues[ue], nil
		}
	}
	return nil, errors.New("[5GC] Registered UE with msin " + msin + "not found")
}

func (c *AMFContext) ExecuteForAllUe(function func(ue *UEContext)) {
	ueMutex.Lock()
	defer ueMutex.Unlock()
	for ue := range c.ues {
		function(c.ues[ue])
	}
}

func (c *AMFContext) Provision(nssai models.Snssai, securityContext SecurityContext) error {
	_, ok := c.provisionedData[securityContext.msin]
	if ok {
		return errors.New("[5GC] Cannot create new subscriber: subscriber with msin " + securityContext.msin + " already exist")
	}
	scMutex.Lock()
	c.provisionedData[securityContext.msin] = provisionedData{defaultSNssai: nssai, securityContext: securityContext}
	scMutex.Unlock()
	return nil
}

func (c *AMFContext) NewUE(ueRanNgapId int64) *UEContext {
	newUE := UEContext{}
	newUE.SetRanNgapId(ueRanNgapId)
	newUE.SetAmfNgapId(c.getAmfUeId())
	newUE.smContexts = make(map[int32]*SmContext)
	newUE.state = fsm.NewState(Deregistered)
	newUE.ueFsm = c.ueFsm
	newUE.pduFsm = c.pduFsm
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
