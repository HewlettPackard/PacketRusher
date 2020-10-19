package data_test_engine

import (
	"encoding/hex"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"strconv"
	"time"
)

// get source ping.
func getSrcPing(i int) string {
	basePing := "60.60.0."
	srcPing := basePing + strconv.Itoa(i)
	return srcPing
}

func ipv4HeaderChecksum(hdr *ipv4.Header) uint32 {
	var Checksum uint32
	Checksum += uint32((hdr.Version<<4|(20>>2&0x0f))<<8 | hdr.TOS)
	Checksum += uint32(hdr.TotalLen)
	Checksum += uint32(hdr.ID)
	Checksum += uint32((hdr.FragOff & 0x1fff) | (int(hdr.Flags) << 13))
	Checksum += uint32((hdr.TTL << 8) | (hdr.Protocol))

	src := hdr.Src.To4()
	Checksum += uint32(src[0])<<8 | uint32(src[1])
	Checksum += uint32(src[2])<<8 | uint32(src[3])
	dst := hdr.Dst.To4()
	Checksum += uint32(dst[0])<<8 | uint32(dst[1])
	Checksum += uint32(dst[2])<<8 | uint32(dst[3])
	return ^(Checksum&0xffff0000>>16 + Checksum&0xffff)
}

func pingUE(connN3 *net.UDPConn, gtpHeader string, ipUe string) error {

	// gtpheader = GTP-TEIDs for the RAN-UPF tunnels(uplink)
	gtpHdr, err := hex.DecodeString(gtpHeader)
	if err != nil {
		return fmt.Errorf("Error getting gtp header")
	}
	icmpData, err := hex.DecodeString("8c870d0000000000101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f3031323334353637")
	if err != nil {
		return fmt.Errorf("Error getting icmp data")
	}

	// included ping source.
	ipv4hdr := ipv4.Header{
		Version:  4,
		Len:      20,
		Protocol: 1,
		Flags:    0,
		TotalLen: 48,
		TTL:      64,
		Src:      net.ParseIP(ipUe).To4(),
		Dst:      net.ParseIP("60.60.0.101").To4(),
		ID:       1,
	}
	checksum := ipv4HeaderChecksum(&ipv4hdr)
	ipv4hdr.Checksum = int(checksum)

	v4HdrBuf, err := ipv4hdr.Marshal()
	if err != nil {
		return fmt.Errorf("Error in ping!")
	}
	tt := append(gtpHdr, v4HdrBuf...)
	if err != nil {
		return fmt.Errorf("Error in ping!")
	}

	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: 12394, Seq: 1,
			Data: icmpData,
		},
	}
	b, err := m.Marshal(nil)
	if err != nil {
		return fmt.Errorf("Error in ping!")
	}
	b[2] = 0xaf
	b[3] = 0x88

	_, err = connN3.Write(append(tt, b...))
	if err != nil {
		return fmt.Errorf("Error in ping!")
	}
	time.Sleep(1 * time.Second)

	// function worked fine.
	return nil
}
