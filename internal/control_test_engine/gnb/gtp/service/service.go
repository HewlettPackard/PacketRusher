package service

import (
	create "context"
	"fmt"
	gtpv1 "github.com/wmnsk/go-gtp/v1"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"net"
)

func InitGTPTunnel(gnb *context.GNBContext) error {

	// get UPF ip and UPF port.
	remote := fmt.Sprintf("%s:%d", gnb.GetUpfIp(), gnb.GetUpfPort())
	remoteUdp, err := net.ResolveUDPAddr("udp", remote)
	if err != nil {
		return fmt.Errorf("Error in resolving UPF address for GTP/UDP tunnel", err)
	}

	// get GNB Data plane ip and GNB Data plane port.
	local := fmt.Sprintf("%s:%d", gnb.GetGnbIpByData(), gnb.GetGnbPortByData())
	localUdp, err := net.ResolveUDPAddr("udp", local)
	if err != nil {
		return fmt.Errorf("Error in resolving GNB address for GTP/UDP tunnel", err)
	}

	// make dial for UPF.
	conte := create.TODO()
	userPlane, err := gtpv1.DialUPlane(conte, localUdp, remoteUdp)
	if err != nil {
		return fmt.Errorf("Error in dial between GNB and UPF", err)
	}

	// successful established GTP/UDP tunnel.
	gnb.SetUserPlane(userPlane)

	go gtpListen(gnb)

	return nil
}

func gtpListen(gnb *context.GNBContext) {

	buf := make([]byte, 65535)
	conn := gnb.GetUserPlane()

	defer func() {
		err := conn.Close()
		if err != nil {
			fmt.Printf("Error in closing GTP/UDP tunnel\n")
		}
	}()

	for {

		n, addr, teid, err := conn.ReadFromGTP(buf)
		if err != nil {
			fmt.Println("Error read in GTP tunnel")
			return
		}

		fmt.Println("Address read is", addr)

		forwardData := make([]byte, n)
		copy(forwardData, buf[:n])

		// find owner of  the Data Plane.
		ue, err := gnb.GetGnbUeByTeid(teid)
		if err != nil || ue == nil {
			fmt.Println("Invalid Downlink Teid. UE is not found in UE Pool")
			return
		}

		// handling data plane.
		go processingData(ue, forwardData)
	}

}

func processingData(ue *context.GNBUe, packet []byte) {

	// send data plane to UE.
}
