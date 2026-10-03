// SPDX-License-Identifier: Apache-2.0
package message_test

import (
	"bytes"
	"github.com/free5gc/ngap/ie"
	"github.com/free5gc/ngap/message"
	interfaces "my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/interface_management"
	mobility "my5G-RANTester/internal/control_test_engine/gnb/ngap/message/ngap_control/ue_mobility_management"
	codec "my5G-RANTester/lib/ngap"
	"testing"
)

func TestConfiguredGNBIdentityOnNGSetupAndHandoverWire(t *testing.T) {
	for _, bits := range []uint8{22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32} {
		source, ue, _ := fixture(t, "000001")
		target, _, _ := fixture(t, "000002")
		if err := source.ConfigureIdentity(bits, 3); err != nil {
			t.Fatal(err)
		}
		if err := target.ConfigureIdentity(bits, 4); err != nil {
			t.Fatal(err)
		}
		wire, err := interfaces.NGSetupRequest(target, "width-test")
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := message.Parse(wire)
		if err != nil {
			t.Fatal(err)
		}
		global := decoded.(*message.NGSetupRequest).GlobalRANNodeID.Choice.(*ie.GlobalGNBID)
		gnb := global.GNBID.Choice.(*ie.GNBIDForGNBID).Value
		if gnb.BitLength != uint64(bits) || !bytes.Equal(gnb.Bytes, target.GetGnbIdInBytes()) {
			t.Fatalf("NG Setup lost %d-bit identity", bits)
		}
		wire, err = mobility.HandoverRequired(source, target, ue)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err = message.Parse(wire)
		if err != nil {
			t.Fatal(err)
		}
		ho := decoded.(*message.HandoverRequired)
		global = ho.TargetID.Choice.(*ie.TargetRANNodeID).GlobalRANNodeID.Choice.(*ie.GlobalGNBID)
		gnb = global.GNBID.Choice.(*ie.GNBIDForGNBID).Value
		if gnb.BitLength != uint64(bits) || !bytes.Equal(gnb.Bytes, target.GetGnbIdInBytes()) {
			t.Fatalf("handover lost %d-bit identity", bits)
		}
		transfer := &ie.SourceNGRANNodeToTargetNGRANNodeTransparentContainer{}
		if err := codec.Unmarshal(ho.SourceToTargetTransparentContainer.Value, transfer); err != nil {
			t.Fatal(err)
		}
		nci := transfer.TargetCellID.Choice.(*ie.NRCGI).NRCellIdentity.Value
		if nci.BitLength != 36 || !bytes.Equal(nci.Bytes, target.GetNRCellIdentity().Value.Bytes) {
			t.Fatalf("handover target NCI mismatch at %d bits", bits)
		}
	}
}
