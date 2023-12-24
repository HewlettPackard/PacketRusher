// Copyright 2023 Edgecom LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ebpf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"reflect"
	"testing"
	"unsafe"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func testArp(bpfObjects *BpfObjects) error {

	packetArp := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(packetArp, gopacket.SerializeOptions{},
		&layers.Ethernet{
			SrcMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 10},
			DstMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 20},
			EthernetType: layers.EthernetTypeARP,
		},
		&layers.ARP{},
	); err != nil {
		return fmt.Errorf("serializing input packet failed: %v", err)
	}

	bpfRet, bufOut, err := bpfObjects.UpfIpEntrypointFunc.Test(packetArp.Bytes())
	if err != nil {
		return fmt.Errorf("test run failed: %v", err)
	}
	if bpfRet != 2 { // XDP_PASS
		return fmt.Errorf("unexpected return value: %d", bpfRet)
	}
	if !reflect.DeepEqual(bufOut, packetArp.Bytes()) {
		return errors.New("unexpected data modification")
	}

	return nil
}

func testArpBenchmark(bpfObjects *BpfObjects, repeat int) (int64, error) {

	packetArp := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(packetArp, gopacket.SerializeOptions{},
		&layers.Ethernet{
			SrcMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 10},
			DstMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 20},
			EthernetType: layers.EthernetTypeARP,
		},
		&layers.ARP{},
	); err != nil {
		return 0, fmt.Errorf("serializing input packet failed: %v", err)
	}

	_, duration, err := bpfObjects.UpfIpEntrypointFunc.Benchmark(packetArp.Bytes(), repeat, func() {})
	if err != nil {
		return 0, fmt.Errorf("benchmark run failed: %v", err)
	}

	return duration.Nanoseconds(), nil
}

func testGtpBenchmark(bpfObjects *BpfObjects, repeat int) (int64, error) {

	packet := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(packet, gopacket.SerializeOptions{},
		&layers.Ethernet{
			SrcMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 10},
			DstMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 20},
			EthernetType: layers.EthernetTypeIPv4,
		},
		&layers.IPv4{
			Version:  4,
			DstIP:    net.IP{10, 3, 0, 10},
			SrcIP:    net.IP{10, 3, 0, 20},
			Protocol: layers.IPProtocolUDP,
			IHL:      5,
		},
		&layers.UDP{
			DstPort: 2152,
			SrcPort: 2152,
		},
		&layers.GTPv1U{
			Version:        1,
			MessageType:    255, // GTPU_G_PDU
			TEID:           1,
			SequenceNumber: 0,
		},
		&layers.IPv4{
			Version:  4,
			DstIP:    net.IP{1, 1, 1, 1},
			SrcIP:    net.IP{10, 60, 0, 1},
			Protocol: layers.IPProtocolICMPv4,
			IHL:      5,
		},
		&layers.ICMPv4{
			TypeCode: layers.ICMPv4TypeEchoRequest,
			Id:       0,
			Seq:      0,
		},
	); err != nil {
		return 0, fmt.Errorf("serializing input packet failed: %v", err)
	}

	_, duration, err := bpfObjects.UpfIpEntrypointFunc.Benchmark(packet.Bytes(), repeat, func() {})
	if err != nil {
		return 0, fmt.Errorf("benchmark run failed: %v", err)
	}

	return duration.Nanoseconds(), nil
}

func testGtpWithPDRBenchmark(bpfObjects *BpfObjects, repeat int) (int64, error) {

	teid := uint32(1)

	packet := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(packet, gopacket.SerializeOptions{},
		&layers.Ethernet{
			SrcMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 10},
			DstMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 20},
			EthernetType: layers.EthernetTypeIPv4,
		},
		&layers.IPv4{
			Version:  4,
			DstIP:    net.IP{10, 3, 0, 10},
			SrcIP:    net.IP{10, 3, 0, 20},
			Protocol: layers.IPProtocolUDP,
			IHL:      5,
		},
		&layers.UDP{
			DstPort: 2152,
			SrcPort: 2152,
		},
		&layers.GTPv1U{
			Version:        1,
			MessageType:    255, // GTPU_G_PDU
			TEID:           teid,
			SequenceNumber: 0,
		},
		&layers.IPv4{
			Version:  4,
			DstIP:    net.IP{1, 1, 1, 1},
			SrcIP:    net.IP{10, 60, 0, 1},
			Protocol: layers.IPProtocolICMPv4,
			IHL:      5,
		},
		&layers.ICMPv4{
			TypeCode: layers.ICMPv4TypeEchoRequest,
			Id:       0,
			Seq:      0,
		},
	); err != nil {
		return 0, fmt.Errorf("serializing input packet failed: %v", err)
	}

	pdr := PdrInfo{OuterHeaderRemoval: 0, FarId: 1, QerId: 1}
	far := FarInfo{Action: 2, OuterHeaderCreation: 1, RemoteIP: 1, LocalIP: 2, Teid: 2, TransportLevelMarking: 0}
	qer := QerInfo{GateStatusUL: 0, GateStatusDL: 0, Qfi: 0, MaxBitrateUL: 1000000, MaxBitrateDL: 100000, StartUL: 0, StartDL: 0}

	if err := bpfObjects.FarMap.Put(uint32(1), unsafe.Pointer(&far)); err != nil {
		return 0, fmt.Errorf("benchmark run failed: %v", err)
	}
	if err := bpfObjects.QerMap.Put(uint32(1), unsafe.Pointer(&qer)); err != nil {
		return 0, fmt.Errorf("benchmark run failed: %v", err)
	}
	if err := bpfObjects.PdrMapUplinkIp4.Put(teid, unsafe.Pointer(&pdr)); err != nil {
		return 0, fmt.Errorf("benchmark run failed: %v", err)
	}

	_, duration, err := bpfObjects.UpfIpEntrypointFunc.Benchmark(packet.Bytes(), repeat, func() {})
	if err != nil {
		return 0, fmt.Errorf("benchmark run failed: %v", err)
	}

	return duration.Nanoseconds(), nil
}

func testGtpEcho(bpfObjects *BpfObjects) error {

	packetArp := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(packetArp, gopacket.SerializeOptions{},
		&layers.Ethernet{
			SrcMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 10},
			DstMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 20},
			EthernetType: layers.EthernetTypeIPv4,
		},
		&layers.IPv4{
			Version:  4,
			DstIP:    net.IP{10, 3, 0, 10},
			SrcIP:    net.IP{10, 3, 0, 20},
			Protocol: layers.IPProtocolUDP,
			IHL:      5,
		},
		&layers.UDP{
			DstPort: 2152,
			SrcPort: 2152,
		},
		&layers.GTPv1U{
			Version:        1,
			MessageType:    1, // GTPU_ECHO_REQUEST
			TEID:           0,
			SequenceNumber: 0,
		},
	); err != nil {
		return fmt.Errorf("serializing input packet failed: %v", err)
	}

	bpfRet, bufOut, err := bpfObjects.UpfIpEntrypointFunc.Test(packetArp.Bytes())
	if err != nil {
		return fmt.Errorf("test run failed: %v", err)
	}
	if bpfRet != 3 { // XDP_TX
		return fmt.Errorf("unexpected return value: %d", bpfRet)
	}

	response := gopacket.NewPacket(bufOut, layers.LayerTypeEthernet, gopacket.Default)
	if gtpLayer := response.Layer(layers.LayerTypeGTPv1U); gtpLayer != nil {
		gtp, _ := gtpLayer.(*layers.GTPv1U)

		if gtp.MessageType != 2 { //GTPU_ECHO_RESPONSE
			return fmt.Errorf("unexpected gtp response: %d", gtp.MessageType)
		}
		if gtp.SequenceNumber != 0 {
			return fmt.Errorf("unexpected gtp sequence: %d", gtp.SequenceNumber)
		}
		if gtp.TEID != 0 {
			return fmt.Errorf("unexpected gtp TEID: %d", gtp.TEID)
		}
	} else {
		return errors.New("unexpected response")
	}

	return nil
}

func testGtpWithSDFFilter(bpfObjects *BpfObjects) error {

	teid := uint32(1)

	packet := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(packet, gopacket.SerializeOptions{},
		&layers.Ethernet{
			SrcMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 10},
			DstMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 20},
			EthernetType: layers.EthernetTypeIPv4,
		},
		&layers.IPv4{
			Version:  4,
			DstIP:    net.IP{10, 3, 0, 10},
			SrcIP:    net.IP{10, 3, 0, 20},
			Protocol: layers.IPProtocolUDP,
			IHL:      5,
		},
		&layers.UDP{
			DstPort: 2152,
			SrcPort: 2152,
		},
		&layers.GTPv1U{
			Version:        1,
			MessageType:    255, // GTPU_G_PDU
			TEID:           teid,
			SequenceNumber: 0,
		},
		&layers.IPv4{
			Version:  4,
			DstIP:    net.IP{1, 1, 1, 1},
			SrcIP:    net.IP{10, 60, 0, 1},
			Protocol: layers.IPProtocolICMPv4,
			IHL:      5,
		},
		&layers.ICMPv4{
			TypeCode: layers.ICMPv4TypeEchoRequest,
			Id:       0,
			Seq:      0,
		},
	); err != nil {
		return fmt.Errorf("serializing input packet failed: %v", err)
	}

	pdr := PdrInfo{OuterHeaderRemoval: 0, FarId: 1, QerId: 1}
	farForward := FarInfo{Action: 2, OuterHeaderCreation: 1, RemoteIP: 1, LocalIP: 2, Teid: 2, TransportLevelMarking: 0}
	farDrop := FarInfo{Action: 1, OuterHeaderCreation: 1, RemoteIP: 1, LocalIP: 2, Teid: 2, TransportLevelMarking: 0}
	qer := QerInfo{GateStatusUL: 0, GateStatusDL: 0, Qfi: 0, MaxBitrateUL: 1000000, MaxBitrateDL: 100000, StartUL: 0, StartDL: 0}

	if err := bpfObjects.FarMap.Put(uint32(1), unsafe.Pointer(&farForward)); err != nil {
		return fmt.Errorf("can't set FAR: %v", err)
	}
	if err := bpfObjects.FarMap.Put(uint32(2), unsafe.Pointer(&farDrop)); err != nil {
		return fmt.Errorf("can't set FAR: %v", err)
	}
	if err := bpfObjects.QerMap.Put(uint32(1), unsafe.Pointer(&qer)); err != nil {
		return fmt.Errorf("can't set QER: %v", err)
	}

	if err := bpfObjects.PutPdrUplink(teid, pdr); err != nil {
		return fmt.Errorf("can't set uplink PDR: %v", err)
	}

	sdf := SdfFilter{
		Protocol:     1,
		SrcAddress:   IpWMask{Type: 1, Ip: net.IP{10, 60, 0, 1}, Mask: net.IPMask{255, 255, 255, 255}},
		DstAddress:   IpWMask{Type: 1, Ip: net.IP{1, 1, 1, 1}, Mask: net.IPMask{255, 255, 255, 255}},
		SrcPortRange: PortRange{LowerBound: 0, UpperBound: 65535},
		DstPortRange: PortRange{LowerBound: 0, UpperBound: 65535},
	}
	pdr.SdfFilter = &sdf
	pdr.FarId = 2
	if err := bpfObjects.PutPdrUplink(teid, pdr); err != nil {
		return fmt.Errorf("can't set uplink PDR: %v", err)
	}

	bpfRet, _, err := bpfObjects.UpfIpEntrypointFunc.Test(packet.Bytes())
	if err != nil {
		return fmt.Errorf("ebpf2 run failed: %v", err)
	}

	if bpfRet != 1 { // XDP_DROP
		return fmt.Errorf("unexpected return value: %d", bpfRet)
	}

	return nil
}

func testGtpExtHeader(bpfObjects *BpfObjects) error {

	teid := uint32(2)

	packet := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(packet, gopacket.SerializeOptions{},
		&layers.Ethernet{
			SrcMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 10},
			DstMAC:       net.HardwareAddr{1, 0, 0, 3, 0, 20},
			EthernetType: layers.EthernetTypeIPv4,
		},
		&layers.IPv4{
			Version:  4,
			SrcIP:    net.IP{1, 1, 1, 1},
			DstIP:    net.IP{10, 60, 0, 1},
			Protocol: layers.IPProtocolICMPv4,
			IHL:      5,
		},
		&layers.ICMPv4{
			TypeCode: layers.ICMPv4TypeEchoReply,
			Id:       0,
			Seq:      0,
		},
	); err != nil {
		return fmt.Errorf("serializing input packet failed: %v", err)
	}

	pdr := PdrInfo{OuterHeaderRemoval: 0, FarId: 1, QerId: 1}
	farForward := FarInfo{
		Action:                2,
		OuterHeaderCreation:   1,
		RemoteIP:              binary.LittleEndian.Uint32(net.IP{10, 3, 0, 10}),
		LocalIP:               binary.LittleEndian.Uint32(net.IP{10, 3, 0, 20}),
		Teid:                  teid,
		TransportLevelMarking: 0}
	qer := QerInfo{GateStatusUL: 0, GateStatusDL: 0, Qfi: 5, MaxBitrateUL: 1000000, MaxBitrateDL: 100000, StartUL: 0, StartDL: 0}

	if err := bpfObjects.FarMap.Put(uint32(1), unsafe.Pointer(&farForward)); err != nil {
		return fmt.Errorf("can't set FAR: %v", err)
	}
	if err := bpfObjects.QerMap.Put(uint32(1), unsafe.Pointer(&qer)); err != nil {
		return fmt.Errorf("can't set QER: %v", err)
	}

	if err := bpfObjects.PutPdrDownlink(net.IP{10, 60, 0, 1}, pdr); err != nil {
		return fmt.Errorf("can't set downlink PDR: %v", err)
	}

	bpfRet, bufOut, err := bpfObjects.UpfIpEntrypointFunc.Test(packet.Bytes())
	if err != nil {
		return fmt.Errorf("ebpf2 run failed: %v", err)
	}

	if bpfRet != 4 { // XDP_REDIRECT
		return fmt.Errorf("unexpected return value: %d", bpfRet)
	}

	response := gopacket.NewPacket(bufOut, layers.LayerTypeEthernet, gopacket.Default)
	if gtpLayer := response.Layer(layers.LayerTypeGTPv1U); gtpLayer != nil {
		gtp, _ := gtpLayer.(*layers.GTPv1U)

		if gtp.MessageType != 255 { //GTPU_G_PDU
			return fmt.Errorf("unexpected gtp response: %d", gtp.MessageType)
		}
		if gtp.ExtensionHeaderFlag != true {
			return fmt.Errorf("unexpected gtp extention flag: %t", gtp.ExtensionHeaderFlag)
		}
		if gtp.TEID != teid {
			return fmt.Errorf("unexpected gtp TEID: %d", gtp.TEID)
		}
		if len(gtp.GTPExtensionHeaders) != 1 {
			return fmt.Errorf("unexpected gtp extention header count: %d", len(gtp.GTPExtensionHeaders))
		}

		extHeader := gtp.GTPExtensionHeaders[0]
		if extHeader.Type != 0x85 {
			return fmt.Errorf("unexpected gtp extention header: %d", gtp.GTPExtensionHeaders[0].Type)
		}
		if len(extHeader.Content) != 2 {
			return fmt.Errorf("unexpected gtp extention header len: %d", len(extHeader.Content))
		}

		if extHeader.Content[1] != 5 {
			return fmt.Errorf("unexpected gtp extention header QFI: %d %d", extHeader.Content[0], extHeader.Content[1])
		}
	} else {
		return fmt.Errorf("unexpected response: %v", response)
	}

	return nil
}

func TestEntrypoint(t *testing.T) {

	if err := IncreaseResourceLimits(); err != nil {
		t.Fatalf("Can't increase resource limits: %s", err.Error())
	}

	bpfObjects := &BpfObjects{}

	if err := bpfObjects.Load(); err != nil {
		t.Fatalf("Loading bpf objects failed: %s", err.Error())
	}

	defer bpfObjects.Close()

	t.Run("Arp test", func(t *testing.T) {
		err := testArp(bpfObjects)
		if err != nil {
			t.Fatalf("test failed: %s", err)
		}
	})

	t.Run("GTP-U Echo test", func(t *testing.T) {
		err := testGtpEcho(bpfObjects)
		if err != nil {
			t.Fatalf("test failed: %s", err)
		}
	})

	t.Run("SDF filter test", func(t *testing.T) {
		err := testGtpWithSDFFilter(bpfObjects)
		if err != nil {
			t.Fatalf("test failed: %s", err)
		}
	})

	t.Run("GTP Extention Header test", func(t *testing.T) {
		err := testGtpExtHeader(bpfObjects)
		if err != nil {
			t.Fatalf("test failed: %s", err)
		}
	})

}

func TestEntrypointBenchmark(t *testing.T) {

	if err := IncreaseResourceLimits(); err != nil {
		t.Fatalf("Can't increase resource limits: %s", err.Error())
	}

	bpfObjects := &BpfObjects{}

	if err := bpfObjects.Load(); err != nil {
		t.Fatalf("Loading bpf objects failed: %s", err.Error())
	}

	defer bpfObjects.Close()

	t.Run("Arp (x1)) benchmark", func(t *testing.T) {
		duration, err := testArpBenchmark(bpfObjects, 1)
		if err != nil {
			t.Fatalf("test failed: %s", err)
		}

		t.Logf("%s result: %d ns", t.Name(), duration)
	})

	t.Run("Arp (x1000000) benchmark", func(t *testing.T) {
		duration, err := testArpBenchmark(bpfObjects, 1000000)
		if err != nil {
			t.Fatalf("test failed: %s", err)
		}

		t.Logf("%s result: %d ns", t.Name(), duration)
	})

	t.Run("Gtp (x1)) benchmark", func(t *testing.T) {
		duration, err := testGtpBenchmark(bpfObjects, 1)
		if err != nil {
			t.Fatalf("test failed: %s", err)
		}

		t.Logf("%s result: %d ns", t.Name(), duration)
	})

	t.Run("Gtp (x1000000) benchmark", func(t *testing.T) {
		duration, err := testGtpBenchmark(bpfObjects, 1000000)
		if err != nil {
			t.Fatalf("test failed: %s", err)
		}

		t.Logf("%s result: %d ns", t.Name(), duration)
	})

	t.Run("Gtp with PDR (x1)) benchmark", func(t *testing.T) {
		duration, err := testGtpWithPDRBenchmark(bpfObjects, 1)
		if err != nil {
			t.Fatalf("test failed: %s", err)
		}

		t.Logf("%s result: %d ns", t.Name(), duration)
	})

	t.Run("Gtp with PDR (x1000000) benchmark", func(t *testing.T) {
		duration, err := testGtpWithPDRBenchmark(bpfObjects, 1000000)
		if err != nil {
			t.Fatalf("test failed: %s", err)
		}

		t.Logf("%s result: %d ns", t.Name(), duration)
	})
}
