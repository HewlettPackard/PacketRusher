/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"errors"
	"github.com/mohae/deepcopy"
	"math"
	"reflect"
	"strconv"
	"sync"

	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/fsm"
	"github.com/free5gc/util/idgenerator"
	log "github.com/sirupsen/logrus"
)

type AMFContext struct {
	ueMu                sync.RWMutex
	subscriberMu        sync.RWMutex
	gnbMu               sync.RWMutex
	idMu                sync.Mutex
	tmsiGenerator       *idgenerator.IDGenerator
	ueByID              map[int64]*UEContext
	onCreate            func(*UEContext)
	onAssociationChange func()
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
}

type NetworkName struct {
	Full  string
	Short string
}

func (c *AMFContext) NewAmfContext(amfName string, id string, supportedPlmnSnssai []models.Nrf_NFMgmt_PlmnSnssai, servedGuami []models.Guami, relativeCapacity int64, ueFsm *fsm.FSM, pduFsm *fsm.FSM) {
	c.amfName = amfName
	c.id = id
	c.supportedPlmnSnssai = deepcopy.Copy(supportedPlmnSnssai).([]models.Nrf_NFMgmt_PlmnSnssai)
	c.servedGuami = deepcopy.Copy(servedGuami).([]models.Guami)
	c.relativeCapacity = relativeCapacity
	c.gnbs = make(map[string]*GNBContext)
	c.ues = []*UEContext{}
	c.ueByID = make(map[int64]*UEContext)
	c.tmsiGenerator = idgenerator.NewGenerator(1, math.MaxInt32)
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
	c.idMu.Lock()
	defer c.idMu.Unlock()
	if c.tmsiGenerator == nil {
		c.tmsiGenerator = idgenerator.NewGenerator(1, math.MaxInt32)
	}
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
	c.subscriberMu.Lock()
	defer c.subscriberMu.Unlock()
	data, ok := c.provisionedData[msin]
	if !ok {
		return provisionedData{}, errors.New("[5GC] UE with msin " + msin + "not found")
	}
	return data, nil
}

func (c *AMFContext) FindUEById(id int64) (*UEContext, error) {
	c.ueMu.Lock()
	defer c.ueMu.Unlock()
	if ue, ok := c.ueByID[id]; ok {
		return ue, nil
	}
	return nil, errors.New("[5GC] UE with amfNgapId " + strconv.Itoa(int(id)) + "not found")
}

func (c *AMFContext) FindUEByRanId(id int64) (*UEContext, error) {
	c.ueMu.Lock()
	defer c.ueMu.Unlock()
	for ue := range c.ues {
		if c.ues[ue].GetRanNgapId() == id {
			return c.ues[ue], nil
		}
	}

	return nil, errors.New("[5GC] UE with RanNgapId " + strconv.Itoa(int(id)) + "not found")
}

func (c *AMFContext) FindRegisteredUEByMsin(msin string) (*UEContext, error) {
	c.ueMu.Lock()
	defer c.ueMu.Unlock()
	for ue := range c.ues {
		if c.ues[ue].GetState().Is(Registered) && c.ues[ue].GetSecurityContext() != nil && c.ues[ue].GetSecurityContext().msin == msin {
			return c.ues[ue], nil
		}
	}
	return nil, errors.New("[5GC] Registered UE with msin " + msin + "not found")
}

func (c *AMFContext) ExecuteForAllUe(function func(ue *UEContext)) {
	c.ueMu.Lock()
	ues := append([]*UEContext(nil), c.ues...)
	c.ueMu.Unlock()
	for _, ue := range ues {
		function(ue)
	}
}

func (c *AMFContext) Provision(nssai models.Snssai, securityContext SecurityContext) error {
	c.subscriberMu.Lock()
	defer c.subscriberMu.Unlock()
	_, ok := c.provisionedData[securityContext.msin]
	if ok {
		return errors.New("[5GC] Cannot create new subscriber: subscriber with msin " + securityContext.msin + " already exist")
	}
	if c.provisionedData == nil {
		c.provisionedData = make(map[string]provisionedData)
	}
	c.provisionedData[securityContext.msin] = provisionedData{defaultSNssai: nssai, securityContext: securityContext.clone()}
	return nil
}

func (c *AMFContext) NewUE(ueRanNgapId int64) *UEContext { return c.NewUEForGNB(ueRanNgapId, nil) }

// NewUEForGNB binds identity before publishing the context to other readers.
func (c *AMFContext) NewUEForGNB(ueRanNgapId int64, gnb *GNBContext) *UEContext {
	newUE := UEContext{}
	newUE.BindGNB(gnb, ueRanNgapId)
	newUE.SetAmfNgapId(c.getAmfUeId())
	newUE.smContexts = make(map[int32]*SmContext)
	newUE.state = fsm.NewState(Deregistered)
	newUE.ueFsm = c.ueFsm
	newUE.pduFsm = c.pduFsm
	c.ueMu.Lock()
	if c.ueByID == nil {
		c.ueByID = make(map[int64]*UEContext)
	}
	c.ueByID[newUE.amfNgapId] = &newUE
	c.ues = append(c.ues, &newUE)
	c.ueMu.Unlock()
	if c.onCreate != nil {
		c.onCreate(&newUE)
	}
	return &newUE
}

func (c *AMFContext) GetServedGuami() []models.Guami {
	return deepcopy.Copy(c.servedGuami).([]models.Guami)
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
	return deepcopy.Copy(guamis).([]models.Guami)
}

func (c *AMFContext) GetSupportedPlmnSnssai() []models.Nrf_NFMgmt_PlmnSnssai {
	return deepcopy.Copy(c.supportedPlmnSnssai).([]models.Nrf_NFMgmt_PlmnSnssai)
}

func (c *AMFContext) GetRelativeCapacity() int64 {
	return c.relativeCapacity
}

func (c *AMFContext) GetGnb(Addr string) (*GNBContext, error) {
	c.gnbMu.Lock()
	gnb, exist := c.gnbs[Addr]
	c.gnbMu.Unlock()
	if !exist {
		return gnb, errors.New("GNB with address " + Addr + " not found in AMF")
	}
	return gnb, nil
}

func (c *AMFContext) FindGnbById(globalRanNodeID models.GlobalRanNodeId) (*GNBContext, error) {
	c.gnbMu.Lock()
	defer c.gnbMu.Unlock()
	for _, gnb := range c.gnbs {
		if reflect.DeepEqual(*gnb.GetGlobalRanNodeID(), globalRanNodeID) {
			return gnb, nil
		}
	}
	return nil, errors.New("GNB with matching global RanNode ID not found in AMF")

}

func (c *AMFContext) AddGnb(gnbAddr string, gnb *GNBContext) error {
	c.gnbMu.Lock()
	if c.gnbs == nil {
		c.gnbs = make(map[string]*GNBContext)
	}
	c.gnbs[gnbAddr] = gnb
	c.gnbMu.Unlock()
	if c.onAssociationChange != nil {
		c.onAssociationChange()
	}
	return nil
}

func (c *AMFContext) getAmfUeId() int64 {
	c.idMu.Lock()
	defer c.idMu.Unlock()
	id := c.idUeGenerator

	// increment UeId
	c.idUeGenerator++

	return id
}

func (c *AMFContext) GetNetworkName() NetworkName {
	return c.networkName
}

// RemoveGnb cannot retire a newer association that reused the same peer address.
func (c *AMFContext) RemoveGnb(addr string, expected *GNBContext) {
	c.gnbMu.Lock()
	if c.gnbs[addr] == expected {
		delete(c.gnbs, addr)
	}
	c.gnbMu.Unlock()
	if c.onAssociationChange != nil {
		c.onAssociationChange()
	}
}
func (c *AMFContext) GNBCount() int { c.gnbMu.RLock(); defer c.gnbMu.RUnlock(); return len(c.gnbs) }

func (c *AMFContext) Associations() []AssociationRecord {
	c.gnbMu.RLock()
	defer c.gnbMu.RUnlock()
	out := make([]AssociationRecord, 0, len(c.gnbs))
	for _, gnb := range c.gnbs {
		local, remote := gnb.Endpoints()
		out = append(out, AssociationRecord{Local: local, Remote: remote})
	}
	return out
}
