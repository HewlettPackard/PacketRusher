package service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/ipv4"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"net"
)

func InitGatewayGnb(gnb *context.GNBContext) error {

	// get ip for GNB gateway for data plane.
	ipGateway := gnb.GetGatewayGnbIp()

	conn, err := net.ListenPacket("ip4:4", ipGateway)
	if err != nil {
		return fmt.Errorf("[GNB][DATA] Error setting listen gateway GNB", err)
	}

	dataPlaneConn, err := ipv4.NewRawConn(conn)
	if err != nil {
		return fmt.Errorf("[GNB][DATA] Error setting data plane communication with UEs", err)
	}

	// successful established GNB/UE tunnel.
	gnb.SetUePlane(dataPlaneConn)

	go gatewayListen(gnb)

	return nil
}

func gatewayListen(gnb *context.GNBContext) {

	buffer := make([]byte, 65535)
	conn := gnb.GetUePlane()

	defer func() {
		err := conn.Close()
		if err != nil {
			log.Info("[GNB][DATA] Error in closing GNB/UE tunnel\n")
		}
	}()

	for {

		ipHeader, payload, _, err := conn.ReadFrom(buffer)
		// log.Info(" [GNB][DATA] Read %d bytes in GNB/UE tunnel", len(payload))
		if err != nil {
			log.Info("[GNB][DATA] Error in reading from GNB/UE tunnel: %+v", err)
			return
		}

		forwardData := make([]byte, len(payload[:]))
		copy(forwardData, payload[:])

		// find owner of  the Data Plane.
		ue, err := gnb.GetGnbUeByIp(ipHeader.Src.String())
		if err != nil || ue == nil {
			log.Info("[GNB][DATA] Invalid GNB UE IP. UE is not found in GNB UE IP Pool")
			return
		}

		go processingData(ue, gnb, forwardData)
	}
}

func processingData(ue *context.GNBUe, gnb *context.GNBContext, packet []byte) {

	// get GTP/UDP connection.
	conn := gnb.GetN3Plane()
	if conn == nil {
		log.Info("[GNB][GTP] N3 GTP/UDP is not setting")
		return
	}

	// send Data plane with GTP header.
	teidUplink := ue.GetTeidUplink()

	remote := fmt.Sprintf("%s:%d", gnb.GetUpfIp(), gnb.GetUpfPort())
	upfAddr, err := net.ResolveUDPAddr("udp", remote)
	if err != nil {
		log.Info("[GNB][GTP] Error resolving UPF address for GTP/UDP tunnel", err)
		return
	}

	// send Data plane with GTP header.
	_, err = conn.WriteToGTP(teidUplink, packet, upfAddr)
	if err != nil {
		log.Info("[GNB][GTP] Error sending data plane in GTP/UDP tunnel")
	}

	//log.Info("[GNB][GTP] Send %d bytes in GNB->UPF tunnel\n", n)
}
