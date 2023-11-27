package context

import (
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/test/amf/lib/types"

	"github.com/free5gc/openapi/models"
)

type GNBContext struct {
	globalRanNodeID  models.GlobalRanNodeId
	ranNodename      string
	suportedTAList   []types.Tai
	defautlPagingDRX ngapType.PagingDRX
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
