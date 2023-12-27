/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"my5G-RANTester/lib/ngap/ngapSctp"
	"my5G-RANTester/test/aio5gc/lib/types"
	"os"

	"github.com/free5gc/ngap/ngapType"
	"github.com/free5gc/openapi/models"
	"github.com/ishidawataru/sctp"
	log "github.com/sirupsen/logrus"
)

type GNBContext struct {
	globalRanNodeID  models.GlobalRanNodeId
	ranNodename      string
	suportedTAList   []types.Tai
	defautlPagingDRX ngapType.PagingDRX
	conn             *sctp.SCTPConn
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
	gnb.conn = conn
}

func (gnb *GNBContext) SendMsg(packet []byte) {
	info := &sctp.SndRcvInfo{
		Stream: uint16(0),
		PPID:   ngapSctp.NGAP_PPID,
	}
	if packet != nil {
		_, err := gnb.conn.SCTPWrite(packet, info)
		if err != nil {
			log.Printf("[5GC] write failed: %v", err)
			os.Exit(1)
		}
	}
}
