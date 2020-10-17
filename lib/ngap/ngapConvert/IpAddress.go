package ngapConvert

import (
	"my5G-RANTester/lib/aper"
	"my5G-RANTester/lib/ngap/ngapType"
	"net"
)

func IPAddressToString(ipAddr ngapType.TransportLayerAddress) (ipv4Addr, ipv6Addr string) {

	ip := ipAddr.Value

	// Described in 38.414
	switch ip.BitLength {
	case 32: // ipv4
		netIP := net.IPv4(ip.Bytes[0], ip.Bytes[1], ip.Bytes[2], ip.Bytes[3])
		ipv4Addr = netIP.String()
	case 128: // ipv6
		netIP := net.IP{}
		for i := range ip.Bytes {
			netIP = append(netIP, ip.Bytes[i])
		}
		ipv6Addr = netIP.String()
	case 160: // ipv4 + ipv6, and ipv4 is contained in the first 32 bits
		netIPv4 := net.IPv4(ip.Bytes[0], ip.Bytes[1], ip.Bytes[2], ip.Bytes[3])
		netIPv6 := net.IP{}
		for i := range ip.Bytes {
			netIPv6 = append(netIPv6, ip.Bytes[i+4])
		}
		ipv4Addr = netIPv4.String()
		ipv6Addr = netIPv6.String()
	}
	return
}

func IPAddressToNgap(ipv4Addr, ipv6Addr string) (ipAddr ngapType.TransportLayerAddress) {

	if ipv4Addr == "" && ipv6Addr == "" {
		//logger.NgapLog.Warningln("IPAddressToNgap: Both ipv4 & ipv6 are nil string")
		return
	}

	if ipv4Addr != "" && ipv6Addr != "" { // Both ipv4 & ipv6
		ipv4NetIP := net.ParseIP(ipv4Addr).To4()
		ipv6NetIP := net.ParseIP(ipv6Addr).To16()

		ipBytes := []byte{ipv4NetIP[0], ipv4NetIP[1], ipv4NetIP[2], ipv4NetIP[3]}
		for i := 0; i < 16; i++ {
			ipBytes = append(ipBytes, ipv6NetIP[i])
		}

		ipAddr.Value = aper.BitString{
			Bytes:     ipBytes,
			BitLength: 160,
		}
	} else if ipv4Addr != "" && ipv6Addr == "" { // ipv4
		ipv4NetIP := net.ParseIP(ipv4Addr).To4()

		ipBytes := []byte{ipv4NetIP[0], ipv4NetIP[1], ipv4NetIP[2], ipv4NetIP[3]}

		ipAddr.Value = aper.BitString{
			Bytes:     ipBytes,
			BitLength: 32,
		}
	} else { // ipv6
		ipv6NetIP := net.ParseIP(ipv6Addr).To16()

		ipBytes := []byte{}
		for i := 0; i < 16; i++ {
			ipBytes = append(ipBytes, ipv6NetIP[i])
		}

		ipAddr.Value = aper.BitString{
			Bytes:     ipBytes,
			BitLength: 128,
		}
	}

	return
}
