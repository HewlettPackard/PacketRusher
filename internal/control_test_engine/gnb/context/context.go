package context

import (
	"encoding/hex"
	"fmt"
	"github.com/ishidawataru/sctp"
	log "github.com/sirupsen/logrus"
	gtpv1 "github.com/wmnsk/go-gtp/v1"
	"golang.org/x/net/ipv4"
	"net"
	"sync"
)

type GNBContext struct {
	dataInfo       DataInfo    // gnb data plane information
	controlInfo    ControlInfo // gnb control plane information
	uePool         sync.Map    // map[in64]*GNBUe, UeRanNgapId as key
	amfPool        sync.Map    // map[int64]*GNBAmf, AmfId as key
	teidPool       sync.Map    // map[uint32]*GNBUe, downlinkTeid as key
	ueIpPool       sync.Map    // map[string]*GNBUe, ueGnbIp as key
	sliceInfo      Slice
	idUeGenerator  int64  // ran UE id.
	idAmfGenerator int64  // ran amf id
	teidGenerator  uint32 // ran UE downlink Teid
	ueIpGenerator  uint8  // ran ue ip.
}

type DataInfo struct {
	gnbIp        string            // gnb ip for data plane.
	gnbPort      int               // gnb port for data plane.
	upfIp        string            // upf ip
	upfPort      int               // upf port
	gtpPlane     *gtpv1.UPlaneConn // N3 connection
	gatewayGnbIp string            // IP gateway that communicates with UE data plane.
	uePlane      *ipv4.RawConn     // listen UE data plane
}

type Slice struct {
	sd  string
	sst string
}

type ControlInfo struct {
	mcc          string
	mnc          string
	tac          string
	gnbId        string
	gnbIp        string
	gnbPort      int
	unixlistener net.Listener
	n2           *sctp.SCTPConn
}

func (gnb *GNBContext) NewRanGnbContext(gnbId, mcc, mnc, tac, sst, sd, ip, ipData string, port, portData int) {
	gnb.controlInfo.mcc = mcc
	gnb.controlInfo.mnc = mnc
	gnb.controlInfo.tac = tac
	gnb.controlInfo.gnbId = gnbId
	gnb.sliceInfo.sd = sd
	gnb.sliceInfo.sst = sst
	gnb.idUeGenerator = 1
	gnb.idAmfGenerator = 1
	gnb.controlInfo.gnbIp = ip
	gnb.teidGenerator = 1
	gnb.ueIpGenerator = 3
	gnb.controlInfo.gnbPort = port
	gnb.dataInfo.upfPort = 2152
	gnb.dataInfo.gtpPlane = nil
	gnb.dataInfo.gatewayGnbIp = "127.0.0.2"
	gnb.dataInfo.upfIp = ""
	gnb.dataInfo.gnbIp = ipData
	gnb.dataInfo.gnbPort = portData
}

func (gnb *GNBContext) NewGnBUe(conn net.Conn) *GNBUe {

	// TODO if necessary add more information for UE.
	// TODO implement mutex

	// new instance of ue.
	ue := &GNBUe{}

	// set ran UE Ngap Id.
	ranId := gnb.getRanUeId()
	ue.SetRanUeId(ranId)

	ue.SetAmfUeId(0)

	// set unix connection for UE.
	ue.SetUnixSocket(conn)

	// set state to UE.
	ue.SetStateInitialized()

	// set downlinkTeid.
	teidDown := gnb.GetUeTeid()
	ue.SetTeidDownlink(teidDown)

	// store UE in the TEID Pool of GNB.
	gnb.teidPool.Store(teidDown, ue)

	// store UE in the UE Pool of GNB.
	gnb.uePool.Store(ranId, ue)

	// set ran UE IP
	ueIp := gnb.getRanUeIp()
	ue.SetIp(ueIp)

	// store UE in the GNB UE IP Pool.
	gnb.ueIpPool.Store(ue.GetIp().String(), ue)

	// select AMF with Capacity is more than 0.
	amf := gnb.selectAmFByActive()
	if amf == nil {
		log.Info("No AMF available for this UE")
		return nil
	}

	// set amfId and SCTP association for UE.
	ue.SetAmfId(amf.GetAmfId())
	ue.SetSCTP(amf.GetSCTPConn())

	// return UE Context.
	return ue
}

func (gnb *GNBContext) SetListener(conn net.Listener) {
	gnb.controlInfo.unixlistener = conn
}

func (gnb *GNBContext) GetListener() net.Listener {
	return gnb.controlInfo.unixlistener
}

func (gnb *GNBContext) GetGatewayGnbIp() string {
	return gnb.dataInfo.gatewayGnbIp
}

func (gnb *GNBContext) DeleteGnBUeByTeid(teid uint32) {
	gnb.teidPool.Delete(teid)
}

func (gnb *GNBContext) DeleteGnBUe(ranUeId int64) {
	gnb.uePool.Delete(ranUeId)
}

func (gnb *GNBContext) DeleteGnBUeByIp(ip string) {
	gnb.ueIpPool.Delete(ip)
}

func (gnb *GNBContext) GetGnbUe(ranUeId int64) (*GNBUe, error) {
	ue, err := gnb.uePool.Load(ranUeId)
	if !err {
		return nil, fmt.Errorf("UE is not find in GNB UE POOL")
	}
	return ue.(*GNBUe), nil
}

func (gnb *GNBContext) GetGnbUeByIp(ip string) (*GNBUe, error) {
	ue, err := gnb.ueIpPool.Load(ip)
	if !err {
		return nil, fmt.Errorf("UE is not find in GNB UE POOL using IP")
	}
	return ue.(*GNBUe), nil
}

func (gnb *GNBContext) GetGnbUeByTeid(teid uint32) (*GNBUe, error) {
	ue, err := gnb.teidPool.Load(teid)
	if !err {
		return nil, fmt.Errorf("UE is not find in GNB UE POOL using TEID")
	}
	return ue.(*GNBUe), nil
}

func (gnb *GNBContext) NewGnBAmf(ip string, port int) *GNBAmf {

	// TODO if necessary add more information for AMF.
	// TODO implement mutex

	amf := &GNBAmf{}

	// set id for AMF.
	amfId := gnb.getRanAmfId()
	amf.setAmfId(amfId)

	// set AMF ip and AMF port.
	amf.SetAmfIp(ip)
	amf.setAmfPort(port)

	// set state to AMF.
	amf.SetStateInactive()

	// store AMF in the AMF Pool of GNB.
	gnb.amfPool.Store(amfId, amf)

	// Plmns and slices supported by AMF initialized.
	amf.SetLenPlmns(0)
	amf.SetLenSlice(0)

	// return AMF Context
	return amf
}

func (gnb *GNBContext) deleteGnBAmf(amfId int64) {
	gnb.amfPool.Delete(amfId)
}

func (gnb *GNBContext) selectAmFByCapacity() *GNBAmf {
	var amfSelect *GNBAmf
	gnb.amfPool.Range(func(key, value interface{}) bool {
		amf := value.(*GNBAmf)
		if amf.relativeAmfCapacity > 0 {
			amfSelect = amf
			// select AMF and decrement capacity.
			amfSelect.relativeAmfCapacity--
			return false
		} else {
			return true
		}
	})

	return amfSelect
}

func (gnb *GNBContext) selectAmFByActive() *GNBAmf {
	var amfSelect *GNBAmf
	gnb.amfPool.Range(func(key, value interface{}) bool {
		amf := value.(*GNBAmf)
		if amf.GetState() == Active {
			amfSelect = amf
			return false
		} else {
			return true
		}
	})

	return amfSelect
}

func (gnb *GNBContext) getGnbAmf(amfId int64) (*GNBAmf, error) {
	amf, err := gnb.amfPool.Load(amfId)
	if !err {
		return nil, fmt.Errorf("AMF is not find in GNB AMF POOL ")
	}
	return amf.(*GNBAmf), nil
}

func (gnb *GNBContext) getRanUeIp() uint8 {

	// TODO implement mutex

	id := gnb.ueIpGenerator

	// increment Ue Ip Generator.
	gnb.ueIpGenerator++

	return id
}

func (gnb *GNBContext) getRanUeId() int64 {

	// TODO implement mutex

	id := gnb.idUeGenerator

	// increment RanUeId
	gnb.idUeGenerator++

	return id
}

func (gnb *GNBContext) GetUeTeid() uint32 {

	// TODO implement mutex

	id := gnb.teidGenerator

	// increment UE teid.
	gnb.teidGenerator++

	return id
}

// for AMFs Pools.
func (gnb *GNBContext) getRanAmfId() int64 {

	// TODO implement mutex

	id := gnb.idAmfGenerator

	// increment Amf Id
	gnb.idAmfGenerator++

	return id
}

func (gnb *GNBContext) SetUpfIp(ip string) {
	gnb.dataInfo.upfIp = ip
}

func (gnb *GNBContext) setUpfPort(port int) {
	gnb.dataInfo.upfPort = port
}

func (gnb *GNBContext) SetN2(n2 *sctp.SCTPConn) {
	gnb.controlInfo.n2 = n2
}

func (gnb *GNBContext) GetN2() *sctp.SCTPConn {
	return gnb.controlInfo.n2
}

func (gnb *GNBContext) SetN3Plane(n3 *gtpv1.UPlaneConn) {
	gnb.dataInfo.gtpPlane = n3
}

func (gnb *GNBContext) GetUpfIp() string {
	return gnb.dataInfo.upfIp
}

func (gnb *GNBContext) GetUpfPort() int {
	return gnb.dataInfo.upfPort
}

func (gnb *GNBContext) GetN3Plane() *gtpv1.UPlaneConn {
	return gnb.dataInfo.gtpPlane
}

func (gnb *GNBContext) GetUePlane() *ipv4.RawConn {
	return gnb.dataInfo.uePlane
}

func (gnb *GNBContext) SetUePlane(uePlane *ipv4.RawConn) {
	gnb.dataInfo.uePlane = uePlane
}

func (gnb *GNBContext) setGnbPort(port int) {
	gnb.controlInfo.gnbPort = port
}

func (gnb *GNBContext) setGnbIp(ip string) {
	gnb.controlInfo.gnbIp = ip
}

func (gnb *GNBContext) setGnbId(id string) {
	gnb.controlInfo.gnbId = id
}

func (gnb *GNBContext) setTac(tac string) {
	gnb.controlInfo.tac = tac
}

func (gnb *GNBContext) setMnc(mnc string) {
	gnb.controlInfo.mnc = mnc
}

func (gnb *GNBContext) setMcc(mcc string) {
	gnb.controlInfo.mcc = mcc
}

func (gnb *GNBContext) GetGnbId() string {
	return gnb.controlInfo.gnbId
}

func (gnb *GNBContext) GetGnbIpByData() string {
	return gnb.dataInfo.gnbIp
}

func (gnb *GNBContext) GetGnbPortByData() int {
	return gnb.dataInfo.gnbPort
}

func (gnb *GNBContext) GetGnbIp() string {
	return gnb.controlInfo.gnbIp
}

func (gnb *GNBContext) GetGnbPort() int {
	return gnb.controlInfo.gnbPort
}

func (gnb *GNBContext) GetGnbIdInBytes() []byte {
	// changed for bytes.
	resu, err := hex.DecodeString(gnb.controlInfo.gnbId)
	if err != nil {
		fmt.Println(err)
	}
	return resu
}

func (gnb *GNBContext) getTac() string {
	return gnb.controlInfo.tac
}

func (gnb *GNBContext) GetTacInBytes() []byte {
	// changed for bytes.
	resu, err := hex.DecodeString(gnb.controlInfo.tac)
	if err != nil {
		fmt.Println(err)
	}
	return resu
}

func (gnb *GNBContext) getSlice() (string, string) {
	return gnb.sliceInfo.sst, gnb.sliceInfo.sd
}

func (gnb *GNBContext) GetSliceInBytes() ([]byte, []byte) {
	sstBytes, err := hex.DecodeString(gnb.sliceInfo.sst)
	if err != nil {
		fmt.Println(err)
	}

	if gnb.sliceInfo.sd != "" {
		sdBytes, err := hex.DecodeString(gnb.sliceInfo.sd)
		if err != nil {
			fmt.Println(err)
		}
		return sstBytes, sdBytes
	}
	return sstBytes, nil
}

func (gnb *GNBContext) getMccAndMnc() (string, string) {
	return gnb.controlInfo.mcc, gnb.controlInfo.mnc
}

func (gnb *GNBContext) GetMccAndMncInOctets() []byte {

	// reverse mcc and mnc
	mcc := reverse(gnb.controlInfo.mcc)
	mnc := reverse(gnb.controlInfo.mnc)

	// include mcc and mnc in octets
	oct5 := mcc[1:3]
	var oct6 string
	var oct7 string
	if len(gnb.controlInfo.mnc) == 2 {
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

func (gnb *GNBContext) Terminate() {

	// close all connections
	ln := gnb.GetListener()
	if ln != nil {
		log.Info("[GNB][UE] UNIX/NAS Terminated")
		ln.Close()
	}

	n2 := gnb.GetN2()
	if n2 != nil {
		log.Info("[GNB][AMF] N2/TNLA Terminated")
		n2.Close()
	}

	// TODO: problem in close de N3 socket in gtp library
	/*
		n3 := gnb.GetN3Plane()
		if n3 != nil {
			n3.Close()
			log.Info("[GNB][UPF] N3/NG-U Terminated")
		}
	*/

	log.Info("GNB Terminated")
}

func reverse(s string) string {
	// reverse string.
	var aux string
	for _, valor := range s {
		aux = string(valor) + aux
	}
	return aux
}
