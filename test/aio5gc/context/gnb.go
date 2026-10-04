/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2026 Valentin D'Emmanuele
 */
package context

import (
	"my5G-RANTester/lib/ngap/ngapSctp"
	"my5G-RANTester/test/aio5gc/lib/types"

	"sync"

	ngapType "github.com/free5gc/ngap/ie"
	"github.com/free5gc/openapi/models"
	"github.com/ishidawataru/sctp"
	log "my5G-RANTester/internal/log"
)

type GNBContext struct {
	globalRanNodeID  models.GlobalRanNodeId
	ranNodename      string
	suportedTAList   []types.Tai
	defautlPagingDRX ngapType.PagingDRX
	conn             *sctp.SCTPConn
	// connMu guards conn: a gNB that re-establishes from the same address gets its
	// record back, and the accept loop replaces conn while others read it.
	connMu sync.RWMutex
}

func (gnb *GNBContext) SetGlobalRanNodeID(globalRanNodeID models.GlobalRanNodeId) {
	gnb.globalRanNodeID = globalRanNodeID
}

func (gnb *GNBContext) SetRanNodename(ranNodename string) {
	gnb.ranNodename = ranNodename
}

func (gnb *GNBContext) SetSuportedTAList(suportedTAList []types.Tai) {
	gnb.suportedTAList = suportedTAList
}

func (gnb *GNBContext) SetDefautlPagingDRX(defautlPagingDRX ngapType.PagingDRX) {
	gnb.defautlPagingDRX = defautlPagingDRX
}

func (gnb *GNBContext) GetGlobalRanNodeID() *models.GlobalRanNodeId {
	return &gnb.globalRanNodeID
}

func (gnb *GNBContext) GetRanNodename() string {
	return gnb.ranNodename
}

func (gnb *GNBContext) GetSuportedTAList() []types.Tai {
	return gnb.suportedTAList
}

func (gnb *GNBContext) GetDefautlPagingDRX() ngapType.PagingDRX {
	return gnb.defautlPagingDRX
}

func (gnb *GNBContext) SetSCTPConn(conn *sctp.SCTPConn) {
	gnb.connMu.Lock()
	defer gnb.connMu.Unlock()
	gnb.conn = conn
}

func (gnb *GNBContext) GetSCTPConn() *sctp.SCTPConn {
	gnb.connMu.RLock()
	defer gnb.connMu.RUnlock()
	return gnb.conn
}

func (gnb *GNBContext) SendMsg(packet []byte) {
	info := &sctp.SndRcvInfo{
		Stream: uint16(0),
		PPID:   ngapSctp.NGAP_PPID,
	}
	if packet != nil {
		conn := gnb.GetSCTPConn()
		if conn == nil {
			log.Error("[5GC] Cannot send before SCTP association is established")
			return
		}
		_, err := conn.SCTPWrite(packet, info)
		if err != nil {
			log.Infof("[5GC] write failed: %v", err)
			return
		}
	}
}
