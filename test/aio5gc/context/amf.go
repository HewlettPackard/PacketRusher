/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
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
	log "my5G-RANTester/internal/log"
)

type AMFContext struct {
	amfName             string
	id                  string
	supportedPlmnSnssai []models.Nrf_NFMgmt_PlmnSnssai
	servedGuami         []models.Guami
	relativeCapacity    int64
	gnbs                map[string]*GNBContext
	ues                 []*UEContext
	idUeGenerator       int64
	networkName         NetworkName
	provisionedData     map[string]provisionedData
	ueFsm               *fsm.FSM
	pduFsm              *fsm.FSM
	tmsiGenerator       *idgenerator.IDGenerator
	ueMutex             sync.Mutex
	scMutex             sync.Mutex
	gnbMutex            sync.Mutex
	ueIdMutex           sync.Mutex
}

type NetworkName struct {
	Full  string
	Short string
}

func (c *AMFContext) NewAmfContext(amfName string, id string, supportedPlmnSnssai []models.Nrf_NFMgmt_PlmnSnssai, servedGuami []models.Guami, relativeCapacity int64, ueFsm *fsm.FSM, pduFsm *fsm.FSM) {
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
	c.tmsiGenerator = idgenerator.NewGenerator(1, math.MaxInt32)
}

func (c *AMFContext) TmsiAllocate() int32 {
	tmsi, err := c.tmsiGenerator.Allocate()
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
	c.scMutex.Lock()
	defer c.scMutex.Unlock()
	data, ok := c.provisionedData[msin]
	if !ok {
		return provisionedData{}, errors.New("[5GC] UE with msin " + msin + "not found")
	}
	return data, nil
}

func (c *AMFContext) FindUEById(id int64) (*UEContext, error) {
	c.ueMutex.Lock()
	defer c.ueMutex.Unlock()
	for ue := range c.ues {
		if c.ues[ue].amfNgapId == id {
			return c.ues[ue], nil
		}
	}
	return nil, errors.New("[5GC] UE with amfNgapId " + strconv.Itoa(int(id)) + "not found")
}

func (c *AMFContext) FindUEByRanId(id int64) (*UEContext, error) {
	c.ueMutex.Lock()
	defer c.ueMutex.Unlock()
	for ue := range c.ues {
		if c.ues[ue].ranNgapId == id {
			return c.ues[ue], nil
		}
	}

	return nil, errors.New("[5GC] UE with RanNgapId " + strconv.Itoa(int(id)) + "not found")
}

func (c *AMFContext) FindRegisteredUEByMsin(msin string) (*UEContext, error) {
	c.ueMutex.Lock()
	defer c.ueMutex.Unlock()
	for ue := range c.ues {
		if c.ues[ue].GetState().Is(Registered) && c.ues[ue].GetSecurityContext() != nil && c.ues[ue].GetSecurityContext().msin == msin {
			return c.ues[ue], nil
		}
	}
	return nil, errors.New("[5GC] Registered UE with msin " + msin + "not found")
}

func (c *AMFContext) ExecuteForAllUe(function func(ue *UEContext)) {
	c.ueMutex.Lock()
	ues := append([]*UEContext(nil), c.ues...)
	c.ueMutex.Unlock()
	for _, ue := range ues {
		function(ue)
	}
}

func (c *AMFContext) Provision(nssai models.Snssai, securityContext SecurityContext) error {
	c.scMutex.Lock()
	defer c.scMutex.Unlock()
	_, ok := c.provisionedData[securityContext.msin]
	if ok {
		return errors.New("[5GC] Cannot create new subscriber: subscriber with msin " + securityContext.msin + " already exist")
	}
	c.provisionedData[securityContext.msin] = provisionedData{defaultSNssai: nssai, securityContext: securityContext}
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
	c.ueMutex.Lock()
	c.ues = append(c.ues, &newUE)
	c.ueMutex.Unlock()
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
			if c.servedGuami[i].PlmnId.Mcc == plmnIds[j].Mcc && c.servedGuami[i].PlmnId.Mnc == plmnIds[j].Mnc {
				guamis = append(guamis, c.servedGuami[i])
			}
		}
	}
	return guamis
}

func (c *AMFContext) GetSupportedPlmnSnssai() []models.Nrf_NFMgmt_PlmnSnssai {
	return c.supportedPlmnSnssai
}

func (c *AMFContext) GetRelativeCapacity() int64 {
	return c.relativeCapacity
}

func (c *AMFContext) GetGnb(Addr string) (*GNBContext, error) {
	c.gnbMutex.Lock()
	gnb, exist := c.gnbs[Addr]
	c.gnbMutex.Unlock()
	if !exist {
		return gnb, errors.New("GNB with address " + Addr + " not found in AMF")
	}
	return gnb, nil
}

func (c *AMFContext) FindGnbById(globalRanNodeID models.GlobalRanNodeId) (*GNBContext, error) {
	c.gnbMutex.Lock()
	defer c.gnbMutex.Unlock()
	for _, gnb := range c.gnbs {
		if gnb.globalRanNodeID == globalRanNodeID {
			return gnb, nil
		}
	}
	return nil, errors.New("GNB with matching global RanNode ID not found in AMF")

}

func (c *AMFContext) AddGnb(gnbAddr string, gnb *GNBContext) error {
	c.gnbMutex.Lock()
	c.gnbs[gnbAddr] = gnb
	c.gnbMutex.Unlock()
	return nil
}

func (c *AMFContext) getAmfUeId() int64 {
	c.ueIdMutex.Lock()
	defer c.ueIdMutex.Unlock()
	id := c.idUeGenerator

	// increment UeId
	c.idUeGenerator++

	return id
}

func (c *AMFContext) GetNetworkName() NetworkName {
	return c.networkName
}
