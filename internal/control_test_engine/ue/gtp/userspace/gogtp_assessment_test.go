// SPDX-License-Identifier: Apache-2.0
// Assessment only: no production call site uses go-gtp.
package userspace

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"testing"

	"github.com/wmnsk/go-gtp/gtpv1/message"
)

// goGTPAssessmentDecode retains PacketRusher's fixed-header, work-bound,
// allocation and inner-IP checks around an upstream generic codec. Current
// peer, destination ownership and handover generation remain caller policy.
func goGTPAssessmentDecode(wire []byte, qfi uint8) (uint32, []byte, error) {
	return goGTPAssessmentDecodeMethod(wire, qfi, false)
}

func goGTPAssessmentHeaderDecode(wire []byte, qfi uint8) (uint32, []byte, error) {
	return goGTPAssessmentDecodeMethod(wire, qfi, true)
}

func goGTPAssessmentDecodeMethod(wire []byte, qfi uint8, reuseHeader bool) (uint32, []byte, error) {
	id, err := TPDUTEID(wire)
	if err != nil || qfi > 63 {
		return 0, nil, errors.New("invalid fixed header or allocated QFI")
	}
	if wire[0]&7 != 0 {
		if len(wire) < 12 || (wire[0]&4 == 0 && wire[11] != 0) {
			return 0, nil, errors.New("invalid optional header")
		}
		// Bound allocation/work before entering the generic extension decoder.
		offset, next := 12, wire[11]
		for count := 0; next != 0; count++ {
			if count >= 32 || offset >= len(wire) {
				return 0, nil, errors.New("extension work limit")
			}
			size := int(wire[offset]) * 4
			if size < 4 || size > len(wire)-offset {
				return 0, nil, errors.New("invalid extension size")
			}
			next = wire[offset+size-1]
			offset += size
		}
	}
	var header *message.Header
	var reused message.Header
	if reuseHeader {
		err = reused.UnmarshalBinary(wire)
		header = &reused
	} else {
		var pdu *message.TPDU
		pdu, err = message.ParseTPDU(wire)
		if err == nil {
			header = pdu.Header
		}
	}
	if err != nil {
		return 0, nil, err
	}
	seen := false
	for _, extension := range header.ExtensionHeaders {
		if extension.Type != message.ExtHeaderTypePDUSessionContainer {
			continue
		}
		body := extension.Content
		if seen || len(body) < 2 || body[0]>>4 != 0 || body[1]&63 != qfi {
			return 0, nil, errors.New("duplicate or unallocated downlink container")
		}
		seen = true
		required := 2
		if body[1]&0x80 != 0 {
			required++
		}
		if body[0]&8 != 0 {
			required += 8
		}
		if body[0]&4 != 0 {
			required += 3
		}
		if body[0]&2 != 0 {
			required += 4
		}
		if required > len(body) {
			return 0, nil, errors.New("truncated container metadata")
		}
	}
	if _, _, err := IPAddresses(header.Payload); err != nil {
		return 0, nil, err
	}
	return id, header.Payload, nil
}

func assessmentIP(version, size int) []byte {
	header := 20
	if version == 6 {
		header = 40
	}
	ip := make([]byte, header+size)
	if version == 4 {
		ip[0], ip[8], ip[9] = 0x45, 64, 17
		binary.BigEndian.PutUint16(ip[2:4], uint16(len(ip)))
		copy(ip[12:16], []byte{192, 0, 2, 1})
		copy(ip[16:20], []byte{10, 0, 0, 1})
	} else {
		ip[0], ip[6], ip[7] = 0x60, 17, 64
		binary.BigEndian.PutUint16(ip[4:6], uint16(size))
		ip[8], ip[24], ip[39] = 0x20, 0x20, 1
	}
	for i := header; i < len(ip); i++ {
		ip[i] = byte(i)
	}
	return ip
}

func assessmentWire(flags byte, optional, ip []byte) []byte {
	wire := append([]byte{flags, 255, 0, 0, 1, 2, 3, 4}, optional...)
	wire = append(wire, ip...)
	binary.BigEndian.PutUint16(wire[2:4], uint16(len(wire)-8))
	return wire
}

type assessmentFixture struct {
	name string
	wire []byte
}

func assessmentFixtures() []assessmentFixture {
	ip4, ip6 := assessmentIP(4, 1200), assessmentIP(6, 1200)
	chain := []byte{0, 0, 0, 0x40}
	for i := 0; i < 31; i++ {
		chain = append(chain, 1, 0, 0, 0x40)
	}
	chain[len(chain)-1] = 0x85
	chain = append(chain, 1, 0, 9, 0)
	return []assessmentFixture{
		{"BareIPv4_64", assessmentWire(0x30, nil, assessmentIP(4, 64))},
		{"PSCIPv4_64", assessmentWire(0x34, []byte{0, 0, 0, 0x85, 1, 0, 9, 0}, assessmentIP(4, 64))},
		{"BareIPv4_1200", assessmentWire(0x30, nil, ip4)},
		{"PSCIPv4_1200", assessmentWire(0x34, []byte{0, 0, 0, 0x85, 1, 0, 9, 0}, ip4)},
		{"PSCIPv6_1200", assessmentWire(0x34, []byte{0, 0, 0, 0x85, 1, 0, 9, 0}, ip6)},
		{"SequenceNPDUChain", assessmentWire(0x37, []byte{0x12, 0x34, 7, 0x40, 1, 8, 0x68, 0x85, 1, 0, 9, 0}, ip4)},
		{"PSCAllMetadata", assessmentWire(0x34, []byte{0, 0, 0, 0x85, 5, 0x0f, 0xc9, 0xe0, 1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 1, 2, 3, 4, 0}, ip4)},
		{"Bounded32Extensions", assessmentWire(0x34, chain, ip4)},
	}
}

func TestGoGTPAssessmentValidWireAndPolicyParity(t *testing.T) {
	for _, fixture := range assessmentFixtures() {
		t.Run(fixture.name, func(t *testing.T) {
			id, inner, err := DecodeDownlink(fixture.wire, 9)
			if err != nil {
				t.Fatal(err)
			}
			for _, decode := range []func([]byte, uint8) (uint32, []byte, error){goGTPAssessmentDecode, goGTPAssessmentHeaderDecode} {
				gotID, gotInner, err := decode(fixture.wire, 9)
				if err != nil || gotID != id || !bytes.Equal(gotInner, inner) {
					t.Fatalf("adapter differs: id %d, payload equal %v, error %v", gotID, bytes.Equal(gotInner, inner), err)
				}
			}
		})
	}
	// Independent upstream encoder must reproduce our uplink PSC wire bytes.
	ip := assessmentIP(4, 1200)
	upstream := message.NewTPDUWithExtentionHeader(0x01020304, ip, message.NewExtensionHeader(0x85, []byte{0x10, 9}, 0))
	got, err := upstream.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	want, err := Encode(0x01020304, 9, ip)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("upstream uplink differs: %v", err)
	}
}

func TestGoGTPAssessmentGenericDecodeNeedsAdmissionPolicy(t *testing.T) {
	valid := assessmentWire(0x34, []byte{0, 0, 0, 0x85, 1, 0, 9, 0}, assessmentIP(4, 64))
	mutate := func(f func([]byte)) []byte { w := bytes.Clone(valid); f(w); return w }
	fixtures := []assessmentFixture{
		{"TrailingDatagramBytes", append(bytes.Clone(valid), 0, 0, 0, 0)},
		{"ExtensionPointerWithoutE", assessmentWire(0x32, []byte{0, 0, 0, 0x85}, assessmentIP(4, 64))},
		{"WrongVersion", mutate(func(w []byte) { w[0] = 0x74 })},
		{"AbsentPT", mutate(func(w []byte) { w[0] &^= 0x10 })},
		{"ReservedBit", mutate(func(w []byte) { w[0] |= 8 })},
		{"ZeroTEID", mutate(func(w []byte) { clear(w[4:8]) })},
		{"WrongAllocatedQFI", mutate(func(w []byte) { w[14] = 8 })},
		{"UplinkContainer", mutate(func(w []byte) { w[13] = 0x10 })},
		{"TruncatedPSCMetadata", mutate(func(w []byte) { w[13] |= 8 })},
		{"InvalidInnerIPLength", mutate(func(w []byte) { w[19]-- })},
		{"DuplicatePSC", assessmentWire(0x34, []byte{0, 0, 0, 0x85, 1, 0, 9, 0x85, 1, 0, 9, 0}, assessmentIP(4, 64))},
	}
	chain33 := []byte{0, 0, 0, 0x40}
	for i := 0; i < 33; i++ {
		chain33 = append(chain33, 1, 0, 0, 0x40)
	}
	chain33[len(chain33)-1] = 0
	fixtures = append(fixtures, assessmentFixture{"MoreThan32Extensions", assessmentWire(0x34, chain33, assessmentIP(4, 64))})
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			if _, err := message.ParseTPDU(fixture.wire); err != nil {
				t.Fatalf("generic decoder behavior changed: %v", err)
			}
			if _, _, err := DecodeDownlink(fixture.wire, 9); err == nil {
				t.Fatal("current allocation policy accepted negative")
			}
			for _, decode := range []func([]byte, uint8) (uint32, []byte, error){goGTPAssessmentDecode, goGTPAssessmentHeaderDecode} {
				if _, _, err := decode(fixture.wire, 9); err == nil {
					t.Fatal("guarded adapter accepted negative")
				}
			}
		})
	}
}

func TestGoGTPAssessmentFlaggedNoExtensionDifference(t *testing.T) {
	wire := assessmentWire(0x34, []byte{0, 0, 0, 0}, assessmentIP(4, 64))
	if _, _, err := DecodeDownlink(wire, 9); err != nil {
		t.Fatal(err)
	}
	if _, err := message.ParseTPDU(wire); err == nil {
		t.Fatal("expected upstream to attempt extension parsing with E set")
	}
	// This compatibility difference must be resolved explicitly before a real
	// migration; the assessment adapter deliberately does not normalize flags.
}

var assessmentSinkID uint32
var assessmentSinkPayload []byte

func BenchmarkGoGTPAssessmentDecode(b *testing.B) {
	for _, fixture := range assessmentFixtures() {
		parsers := []struct {
			name   string
			decode func([]byte) (uint32, []byte, error)
		}{
			{"SharedPolicy", func(w []byte) (uint32, []byte, error) { return DecodeDownlink(w, 9) }},
			{"GoGTPGuardedPolicy", func(w []byte) (uint32, []byte, error) { return goGTPAssessmentDecode(w, 9) }},
			{"GoGTPHeaderGuardedPolicy", func(w []byte) (uint32, []byte, error) { return goGTPAssessmentHeaderDecode(w, 9) }},
			{"GoGTPRaw_NotEquivalent", func(w []byte) (uint32, []byte, error) {
				pdu, err := message.ParseTPDU(w)
				if err != nil {
					return 0, nil, err
				}
				return pdu.TEID(), pdu.Payload, nil
			}},
		}
		for _, parser := range parsers {
			b.Run(fmt.Sprintf("%s/%s", fixture.name, parser.name), func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(fixture.wire)))
				for i := 0; i < b.N; i++ {
					id, payload, err := parser.decode(fixture.wire)
					if err != nil {
						b.Fatal(err)
					}
					assessmentSinkID, assessmentSinkPayload = id, payload
				}
			})
		}
	}
}
