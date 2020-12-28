package context

import (
	"encoding/hex"
	"fmt"
	gtpv1 "github.com/wmnsk/go-gtp/v1"
	"sync"
)

type GNBContext struct {
	info       GNBInfo
	sliceInfo  SLICE
	amfPool    sync.Map          // map[int64]*AMFContext, AmfId as key
	upfAddress string            // upf address
	userPlane  *gtpv1.UPlaneConn // N3 connection
	uePool     sync.Map          // map[RanUeNgapId]*UEContext, UeRanId as key
}

type GNBInfo struct {
	mcc     string
	mnc     string
	tac     string
	gnbId   string
	gnbIp   string
	gnbPort string
}

type SLICE struct {
	st  string
	sst string
}

func (gnb *GNBContext) NewRanGnbContext(gnbId, mcc, mnc, tac, st, sst string) {
	gnb.info.mcc = mcc
	gnb.info.mnc = mnc
	gnb.info.tac = tac
	gnb.info.gnbId = gnbId
	gnb.sliceInfo.st = st
	gnb.sliceInfo.sst = sst
}

func (gnb *GNBContext) GetGnbId() string {
	return gnb.info.gnbId
}

func (gnb *GNBContext) GetGnbIdInBytes() []byte {
	// changed for bytes.
	resu, err := hex.DecodeString(gnb.info.gnbId)
	if err != nil {
		fmt.Println(err)
	}
	return resu
}

func (gnb *GNBContext) getTac() string {
	return gnb.info.tac
}

func (gnb *GNBContext) GetTacInBytes() []byte {
	// changed for bytes.
	resu, err := hex.DecodeString(gnb.info.tac)
	if err != nil {
		fmt.Println(err)
	}
	return resu
}

func (gnb *GNBContext) getSlice() (string, string) {
	return gnb.sliceInfo.st, gnb.sliceInfo.sst
}

func (gnb *GNBContext) GetSliceInBytes() ([]byte, []byte) {
	stBytes, err := hex.DecodeString(gnb.sliceInfo.st)
	if err != nil {
		fmt.Println(err)
	}

	sstBytes, err := hex.DecodeString(gnb.sliceInfo.sst)
	if err != nil {
		fmt.Println(err)
	}
	return stBytes, sstBytes
}

func (gnb *GNBContext) getMccAndMnc() (string, string) {
	return gnb.info.mcc, gnb.info.mnc
}

func (gnb *GNBContext) GetMccAndMncInOctets() []byte {

	// reverse mcc and mnc
	mcc := reverse(gnb.info.mcc)
	mnc := reverse(gnb.info.mnc)

	// include mcc and mnc in octets
	oct5 := mcc[1:3]
	var oct6 string
	var oct7 string
	if len(gnb.info.mnc) == 2 {
		oct6 = "f" + string(mcc[0])
		oct7 = mnc
	} else {
		oct6 = string(mnc[0]) + string(mcc[0])
		oct7 = mnc[1:3]
	}

	// changed for bytes.
	resu, err := hex.DecodeString(oct5 + oct6 + oct7)
	if err != nil {
		fmt.Println(err)
	}

	return resu
}
