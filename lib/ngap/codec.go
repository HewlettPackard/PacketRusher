package ngap

import (
	"encoding/binary"
	"fmt"
	"net/netip"

	"github.com/free5gc/ngap/aper"
	"github.com/free5gc/ngap/ie"
)

// Marshal and Unmarshal encode NGAP transfer IEs with the current APER API.
func Marshal(value interface{ Write(*aper.PerBitData) error }) ([]byte, error) {
	data := aper.NewPerBitData(nil)
	if err := value.Write(data); err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}
func Unmarshal(payload any, value interface{ Read(*aper.PerBitData) error }) error {
	var data []byte
	switch encoded := payload.(type) {
	case []byte:
		data = encoded
	case aper.OctetString:
		data = encoded
	case *aper.OctetString:
		if encoded == nil {
			return fmt.Errorf("missing NGAP transfer")
		}
		data = *encoded
	default:
		return fmt.Errorf("unsupported NGAP transfer %T", payload)
	}
	return value.Read(aper.NewPerBitData(data))
}
func IPAddressToNgap(ipv4, ipv6 string) *ie.TransportLayerAddress {
	var encoded []byte
	for _, text := range []string{ipv4, ipv6} {
		if text == "" {
			continue
		}
		addr, err := netip.ParseAddr(text)
		if err != nil {
			return nil
		}
		encoded = append(encoded, addr.AsSlice()...)
	}
	return &ie.TransportLayerAddress{Value: aper.BitString{Bytes: encoded, BitLength: uint64(len(encoded) * 8)}}
}
func IPAddressToString(address *ie.TransportLayerAddress) (string, string) {
	if address == nil {
		return "", ""
	}
	data := address.Value.Bytes
	var v4, v6 string
	if len(data) == 4 || len(data) == 20 {
		addr, ok := netip.AddrFromSlice(data[:4])
		if ok {
			v4 = addr.String()
		}
		data = data[4:]
	}
	if len(data) == 16 {
		addr, ok := netip.AddrFromSlice(data)
		if ok {
			v6 = addr.String()
		}
	}
	return v4, v6
}
func Tunnel(value *ie.UPTransportLayerInformation) (*ie.GTPTunnel, error) {
	if value == nil {
		return nil, fmt.Errorf("missing UP transport information")
	}
	tunnel, ok := value.Choice.(*ie.GTPTunnel)
	if !ok || tunnel == nil || tunnel.GTPTEID == nil || len(tunnel.GTPTEID.Value) != 4 || tunnel.TransportLayerAddress == nil {
		return nil, fmt.Errorf("missing or invalid GTP tunnel")
	}
	return tunnel, nil
}

func UPTransport(ip netip.Addr, teid uint32) *ie.UPTransportLayerInformation {
	value := make([]byte, 4)
	binary.BigEndian.PutUint32(value, teid)
	return &ie.UPTransportLayerInformation{Choice: &ie.GTPTunnel{
		TransportLayerAddress: IPAddressToNgap(ip.String(), ""),
		GTPTEID:               &ie.GTPTEID{Value: value},
	}}
}
func UserLocation(plmn *ie.PLMNIdentity, cell *ie.NRCellIdentity, tac []byte) *ie.UserLocationInformation {
	return &ie.UserLocationInformation{Choice: &ie.UserLocationInformationNR{
		NRCGI: &ie.NRCGI{PLMNIdentity: plmn, NRCellIdentity: cell},
		TAI:   &ie.TAI{PLMNIdentity: plmn, TAC: &ie.TAC{Value: tac}},
	}}
}
func Slice(sst, sd []byte) *ie.SNSSAI {
	value := &ie.SNSSAI{SST: &ie.SST{Value: sst}}
	if len(sd) > 0 {
		value.SD = &ie.SD{Value: sd}
	}
	return value
}
func Octets(value []byte) *aper.OctetString {
	encoded := aper.OctetString(value)
	return &encoded
}
func CPAddress(value *ie.CPTransportLayerInformation) *ie.TransportLayerAddress {
	if value == nil {
		return nil
	}
	address, _ := value.Choice.(*ie.TransportLayerAddress)
	return address
}
