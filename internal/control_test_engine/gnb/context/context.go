package context

import (
	"encoding/hex"
	"fmt"
	"github.com/ishidawataru/sctp"
	gtpv1 "github.com/wmnsk/go-gtp/v1"
	"net"
	"sync"
)

// UE main states in the GNB Context.
const initialized = 0x00
const ongoing = 0x01
const ready = 0x02

// AMF main states in the GNB Context.
const inactive = 0x00
const active = 0x01
const overload = 0x02

type GNBContext struct {
	dataInfo       DataPlane
	gnbInfo        GNBInfo
	uePool         sync.Map // map[RanUeNgapId]*GNBUe, UeRanNgapId as key
	amfPool        sync.Map // map[int64]*GNBAmf, AmfId as key
	sliceInfo      Slice
	idUeGenerator  int64
	idAmfGenerator int64
}

type DataPlane struct {
	upfIp     string            // upf ip
	upfPort   string            // upf port
	userPlane *gtpv1.UPlaneConn // N3 connection
}

type Slice struct {
	st  string
	sst string
}

type GNBInfo struct {
	mcc     string
	mnc     string
	tac     string
	gnbId   string
	gnbIp   string
	gnbPort string
}

func (gnb *GNBContext) NewRanGnbContext(gnbId, mcc, mnc, tac, st, sst string) {
	gnb.gnbInfo.mcc = mcc
	gnb.gnbInfo.mnc = mnc
	gnb.gnbInfo.tac = tac
	gnb.gnbInfo.gnbId = gnbId
	gnb.sliceInfo.st = st
	gnb.sliceInfo.sst = sst
	gnb.idUeGenerator = 1
	gnb.idAmfGenerator = 1
}

func (gnb *GNBContext) NewGnBUe(conn net.Conn) *GNBUe {

	// TODO if necessary add more information for UE.
	// TODO implement mutex

	// new instance of ue.
	ue := &GNBUe{}

	// set ranUeNgapId for UE.
	ranId := gnb.getRanUeId()
	ue.SetRanUeId(ranId)

	// set unix connection for UE.
	ue.SetUnixSocket(conn)

	// set state to UE.
	ue.SetState(initialized)

	// store UE in the UE Pool of GNB.
	gnb.uePool.Store(ranId, ue)

	// return UE Context
	return ue
}

func (gnb *GNBContext) DeleteGnBUe(ranUeId int64) {
	gnb.uePool.Delete(ranUeId)
}

func (gnb *GNBContext) getGnbUe(ranUeId int64) (*GNBUe, error) {
	ue, err := gnb.uePool.Load(ranUeId)
	if !err {
		return nil, fmt.Errorf("UE is not find in GNB UE POOL ")
	}
	return ue.(*GNBUe), nil
}

func (gnb *GNBContext) NewGnBAmf(conn sctp.SCTPConn) *GNBAmf {

	// TODO if necessary add more information for AMF.
	// TODO implement mutex

	amf := &GNBAmf{}

	// set id for AMF.
	amfId := gnb.getRanAmfId()
	amf.setAmfId(amfId)

	// set SCTP association for AMF
	amf.setSCTPConn(conn)

	// set state to AMF.
	amf.setState(active)

	// store AMF in the AMF Pool of GNB.
	gnb.amfPool.Store(amfId, amf)

	// return AMF Context
	return amf
}

func (gnb *GNBContext) deleteGnBAmf(amfId int64) {
	gnb.amfPool.Delete(amfId)
}

func (gnb *GNBContext) getGnbAmf(amfId int64) (*GNBAmf, error) {
	amf, err := gnb.amfPool.Load(amfId)
	if !err {
		return nil, fmt.Errorf("AMF is not find in GNB AMF POOL ")
	}
	return amf.(*GNBAmf), nil
}

func (gnb *GNBContext) getRanUeId() int64 {

	// TODO implement mutex

	id := gnb.idUeGenerator

	// increment RanUeId
	gnb.idUeGenerator++

	return id
}

func (gnb *GNBContext) getRanAmfId() int64 {

	// TODO implement mutex

	id := gnb.idAmfGenerator

	// increment Amf Id
	gnb.idAmfGenerator++

	return id
}

func (gnb *GNBContext) setUpfIp(ip string) {
	gnb.dataInfo.upfIp = ip
}

func (gnb *GNBContext) setUpfPort(port string) {
	gnb.dataInfo.upfPort = port
}

func (gnb *GNBContext) setUserPlane(n3 *gtpv1.UPlaneConn) {
	gnb.dataInfo.userPlane = n3
}

func (gnb *GNBContext) getUpfIp() string {
	return gnb.dataInfo.upfIp
}

func (gnb *GNBContext) getUpfPort() string {
	return gnb.dataInfo.upfPort
}

func (gnb *GNBContext) getUserPlane() *gtpv1.UPlaneConn {
	return gnb.dataInfo.userPlane
}

func (gnb *GNBContext) setGnbPort(port string) {
	gnb.gnbInfo.gnbPort = port
}

func (gnb *GNBContext) setGnbIp(ip string) {
	gnb.gnbInfo.gnbIp = ip
}

func (gnb *GNBContext) setGnbId(id string) {
	gnb.gnbInfo.gnbId = id
}

func (gnb *GNBContext) setTac(tac string) {
	gnb.gnbInfo.tac = tac
}

func (gnb *GNBContext) setMnc(mnc string) {
	gnb.gnbInfo.mnc = mnc
}

func (gnb *GNBContext) setMcc(mcc string) {
	gnb.gnbInfo.mcc = mcc
}

func (gnb *GNBContext) GetGnbId() string {
	return gnb.gnbInfo.gnbId
}

func (gnb *GNBContext) GetGnbIp() string {
	return gnb.gnbInfo.gnbIp
}

func (gnb *GNBContext) GetGnbPort() string {
	return gnb.gnbInfo.gnbPort
}

func (gnb *GNBContext) GetGnbIdInBytes() []byte {
	// changed for bytes.
	resu, err := hex.DecodeString(gnb.gnbInfo.gnbId)
	if err != nil {
		fmt.Println(err)
	}
	return resu
}

func (gnb *GNBContext) getTac() string {
	return gnb.gnbInfo.tac
}

func (gnb *GNBContext) GetTacInBytes() []byte {
	// changed for bytes.
	resu, err := hex.DecodeString(gnb.gnbInfo.tac)
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
	return gnb.gnbInfo.mcc, gnb.gnbInfo.mnc
}

func (gnb *GNBContext) GetMccAndMncInOctets() []byte {

	// reverse mcc and mnc
	mcc := reverse(gnb.gnbInfo.mcc)
	mnc := reverse(gnb.gnbInfo.mnc)

	// include mcc and mnc in octets
	oct5 := mcc[1:3]
	var oct6 string
	var oct7 string
	if len(gnb.gnbInfo.mnc) == 2 {
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

func reverse(s string) string {
	// reverse string.
	var aux string
	for _, valor := range s {
		aux = string(valor) + aux
	}
	return aux
}
