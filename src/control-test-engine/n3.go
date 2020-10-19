package control_test_engine

import (
	"fmt"
	"net"
)

func connectToUpf(gnbIP, upfIP string, gnbPort, upfPort int) (*net.UDPConn, error) {
	upfAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", upfIP, upfPort))
	if err != nil {
		return nil, err
	}
	gnbAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", gnbIP, gnbPort))
	if err != nil {
		return nil, err
	}
	return net.DialUDP("udp", gnbAddr, upfAddr)
}
