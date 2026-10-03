/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"my5G-RANTester/lib/ngap/ngapSctp"
	"my5G-RANTester/test/aio5gc/lib/types"

	"errors"
	"github.com/mohae/deepcopy"
	"sync"

	ngapType "github.com/free5gc/ngap/ie"
	"github.com/free5gc/openapi/models"
	"github.com/ishidawataru/sctp"
)

type GNBContext struct {
	mu               sync.RWMutex
	globalRanNodeID  models.GlobalRanNodeId
	ranNodename      string
	suportedTAList   []types.Tai
	defautlPagingDRX ngapType.PagingDRX
	conn             *sctp.SCTPConn
	localEndpoint    string
	remoteEndpoint   string
}

func (gnb *GNBContext) SetGlobalRanNodeID(globalRanNodeID models.GlobalRanNodeId) {
	gnb.mu.Lock()
	defer gnb.mu.Unlock()
	gnb.globalRanNodeID = deepcopy.Copy(globalRanNodeID).(models.GlobalRanNodeId)
}

func (gnb *GNBContext) SetRanNodename(ranNodename string) {
	gnb.mu.Lock()
	defer gnb.mu.Unlock()
	gnb.ranNodename = ranNodename
}

func (gnb *GNBContext) SetSuportedTAList(suportedTAList []types.Tai) {
	gnb.mu.Lock()
	defer gnb.mu.Unlock()
	gnb.suportedTAList = types.CloneTaiList(suportedTAList)
}

func (gnb *GNBContext) SetDefautlPagingDRX(defautlPagingDRX ngapType.PagingDRX) {
	gnb.mu.Lock()
	defer gnb.mu.Unlock()
	gnb.defautlPagingDRX = defautlPagingDRX
}

func (gnb *GNBContext) GetGlobalRanNodeID() *models.GlobalRanNodeId {
	gnb.mu.RLock()
	defer gnb.mu.RUnlock()
	copy := deepcopy.Copy(gnb.globalRanNodeID).(models.GlobalRanNodeId)
	return &copy
}

func (gnb *GNBContext) GetRanNodename() string {
	gnb.mu.RLock()
	defer gnb.mu.RUnlock()
	return gnb.ranNodename
}

func (gnb *GNBContext) GetSuportedTAList() []types.Tai {
	gnb.mu.RLock()
	defer gnb.mu.RUnlock()
	return types.CloneTaiList(gnb.suportedTAList)
}

func (gnb *GNBContext) GetDefautlPagingDRX() ngapType.PagingDRX {
	gnb.mu.RLock()
	defer gnb.mu.RUnlock()
	return gnb.defautlPagingDRX
}

func (gnb *GNBContext) SetSCTPConn(conn *sctp.SCTPConn) {
	gnb.mu.Lock()
	defer gnb.mu.Unlock()
	gnb.conn = conn
	if conn != nil {
		if addr := conn.LocalAddr(); addr != nil {
			gnb.localEndpoint = addr.String()
		}
		if addr := conn.RemoteAddr(); addr != nil {
			gnb.remoteEndpoint = addr.String()
		}
	}
}

func (gnb *GNBContext) GetSCTPConn() *sctp.SCTPConn {
	gnb.mu.RLock()
	defer gnb.mu.RUnlock()
	return gnb.conn
}

func (gnb *GNBContext) SendMsg(packet []byte) error {
	info := &sctp.SndRcvInfo{
		Stream: uint16(0),
		PPID:   ngapSctp.NGAP_PPID,
	}
	if packet != nil {
		conn := gnb.GetSCTPConn()
		if conn == nil {
			return errors.New("[5GC] Cannot send before SCTP association is established")
		}
		_, err := conn.SCTPWrite(packet, info)
		if err != nil {
			return err
		}
	}
	return nil
}

// Endpoints are captured at association acceptance, before teardown can close
// the socket. They describe real bound addresses, including ephemeral ports.
func (gnb *GNBContext) Endpoints() (local, remote string) {
	gnb.mu.RLock()
	defer gnb.mu.RUnlock()
	return gnb.localEndpoint, gnb.remoteEndpoint
}
