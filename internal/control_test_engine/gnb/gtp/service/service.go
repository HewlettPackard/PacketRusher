package service

import (
	create "context"
	"fmt"
	log "github.com/sirupsen/logrus"
	gtpv1 "github.com/wmnsk/go-gtp/v1"
	"golang.org/x/net/ipv4"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"net"
	"syscall"
)

func InitGTPTunnel(gnb *context.GNBContext) error {

	// get UPF ip and UPF port.
	remote := fmt.Sprintf("%s:%d", gnb.GetUpfIp(), gnb.GetUpfPort())
	remoteUdp, err := net.ResolveUDPAddr("udp", remote)
	if err != nil {
		return fmt.Errorf("[GNB][GTP] Error in resolving UPF address for GTP/UDP tunnel", err)
	}

	// get GNB Data plane ip and GNB Data plane port.
	local := fmt.Sprintf("%s:%d", gnb.GetGnbIpByData(), gnb.GetGnbPortByData())
	localUdp, err := net.ResolveUDPAddr("udp", local)
	if err != nil {
		return fmt.Errorf("[GNB][GTP] Error in resolving GNB address for GTP/UDP tunnel", err)
	}

	// make dial for UPF.
	conte := create.TODO()
	userPlane, err := gtpv1.DialUPlane(conte, localUdp, remoteUdp)
	if err != nil {
		return fmt.Errorf("[GNB][GTP] Error in dial between GNB and UPF", err)
	}

	// successful established GTP/UDP tunnel.
	gnb.SetN3Plane(userPlane)

	go gtpListen(gnb)

	return nil
}

func gtpListen(gnb *context.GNBContext) {

	buf := make([]byte, 65535)
	conn := gnb.GetN3Plane()

	/*
		defer func() {
			err := conn.Close()
			if err != nil {
				log.Info("[GNB][GTP] Error in closing GTP/UDP tunnel\n")
			}
		}()
	*/

	for {

		n, _, teid, err := conn.ReadFromGTP(buf)
		if err != nil {
			log.Info("[GNB][GTP] Error read in GTP tunnel")
			break
		}

		forwardData := make([]byte, n)

		// check gtp header to see if has extension or not.
		if buf[3] == 133 {
			copy(forwardData, buf[8:n])
		} else {
			copy(forwardData, buf[:n])
		}

		// find owner of the Data Plane using downlink Teid.
		ue, err := gnb.GetGnbUeByTeid(teid)
		if err != nil || ue == nil {
			log.Info("[GNB][GTP] Invalid Downlink Teid. UE is not found in UE Pool")
			break
		}

		// log.Info("[GNB][GTP] Read ", n," bytes for UE ", ue.GetRanUeId()," in GTP tunnel\n")

		// handling data plane.
		go processingData(ue, gnb, forwardData)
	}

}

func processingData(ue *context.GNBUe, gnb *context.GNBContext, packet []byte) {

	// get connection UE with GNB.
	ueRawn := gnb.GetUePlane()

	// make ipv4 header.
	ipHeader := &ipv4.Header{
		Version:  4,
		Len:      20,
		TOS:      0,
		Flags:    ipv4.DontFragment,
		FragOff:  0,
		TTL:      64,
		Protocol: syscall.IPPROTO_IPIP,
	}

	ueIp := ue.GetIp()

	ipHeader.Dst = ueIp

	// len = ipv4 + packet.
	packetLenght := 20 + len(packet)
	ipHeader.TotalLen = packetLenght

	// send data plane to ue.
	if err := ueRawn.WriteTo(ipHeader, packet, nil); err != nil {
		log.Info("[GNB][GTP] Error in sending data plane to UE")
		return
	}
}
