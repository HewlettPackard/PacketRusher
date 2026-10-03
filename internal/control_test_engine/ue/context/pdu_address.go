// SPDX-License-Identifier: Apache-2.0
package context

import (
	"fmt"
	"net/netip"

	"github.com/free5gc/nas/ie"
)

// Done closes after UE termination. Nil is safe for contexts used without a
// running scenario and disables that select arm.
func (ue *UEContext) Done() <-chan struct{} { return ue.done }

func (s *UEPDUSession) RequestedSessionType() uint8 {
	if s.requestedSessionType == 0 {
		return ie.PDUSessType_IPv4
	}
	return s.requestedSessionType
}

// SetPDUAddress retains the IPv6 interface identifier, not an invented prefix.
// TS 24.501 9.11.4.10 allocates the prefix later through user-plane IPv6 RA.
func (s *UEPDUSession) SetPDUAddress(selected uint8, address *ie.PDUAddr) error {
	if address == nil {
		return fmt.Errorf("missing PDU address")
	}
	requested := s.RequestedSessionType()
	if selected < ie.PDUSessType_IPv4 || selected > ie.PDUSessType_IPv4v6 || (requested != ie.PDUSessType_IPv4v6 && selected != requested) {
		return fmt.Errorf("selected PDU type %d incompatible with requested type %d", selected, requested)
	}
	want4, want6 := selected != ie.PDUSessType_IPv6, selected != ie.PDUSessType_IPv4
	if (want4 && len(address.IPv4) != 4) || (!want4 && len(address.IPv4) != 0) || (want6 && len(address.IPv6IfId) != 8) || (!want6 && len(address.IPv6IfId) != 0) {
		return fmt.Errorf("PDU address fields do not match selected type %d", selected)
	}
	var ipv4 netip.Addr
	if want4 {
		ipv4 = netip.AddrFrom4([4]byte(address.IPv4))
		if ipv4.IsUnspecified() || ipv4.IsMulticast() {
			return fmt.Errorf("invalid allocated IPv4 address %s", ipv4)
		}
	}
	var iid [8]byte
	if want6 {
		copy(iid[:], address.IPv6IfId)
	}
	if len(address.SMFIPv6LLA) != 0 {
		addr, ok := netip.AddrFromSlice(address.SMFIPv6LLA)
		if !want6 || !ok || !addr.Is6() || !addr.IsLinkLocalUnicast() {
			return fmt.Errorf("invalid SMF IPv6 link-local address")
		}
	}
	s.addressLock.Lock()
	defer s.addressLock.Unlock()
	s.selectedSessionType = selected
	s.ueIP = ""
	if ipv4.IsValid() {
		s.ueIP = ipv4.String()
	}
	s.ipv6InterfaceID = iid
	s.ipv6Address = netip.Addr{}
	return nil
}

func (s *UEPDUSession) GetIPv4() netip.Addr {
	ip, _ := netip.ParseAddr(s.GetIp())
	return ip
}

func (s *UEPDUSession) GetIPv6InterfaceID() ([8]byte, bool) {
	s.addressLock.RLock()
	defer s.addressLock.RUnlock()
	return s.ipv6InterfaceID, s.selectedSessionType == ie.PDUSessType_IPv6 || s.selectedSessionType == ie.PDUSessType_IPv4v6
}

func (s *UEPDUSession) GetIPv6() netip.Addr {
	s.addressLock.RLock()
	defer s.addressLock.RUnlock()
	return s.ipv6Address
}

func (s *UEPDUSession) SetIPv6(address netip.Addr) error {
	s.addressLock.Lock()
	defer s.addressLock.Unlock()
	if !address.Is6() || address.Is4In6() || address.IsLinkLocalUnicast() || address.IsMulticast() || address.IsUnspecified() {
		return fmt.Errorf("invalid UE IPv6 address %s", address)
	}
	bytes := address.As16()
	if [8]byte(bytes[8:]) != s.ipv6InterfaceID || (s.selectedSessionType != ie.PDUSessType_IPv6 && s.selectedSessionType != ie.PDUSessType_IPv4v6) {
		return fmt.Errorf("IPv6 address does not match allocated interface identifier")
	}
	s.ipv6Address = address
	return nil
}

// ClearIPv6 withdraws a user-plane prefix without deleting the NAS allocation.
func (s *UEPDUSession) ClearIPv6() {
	s.addressLock.Lock()
	defer s.addressLock.Unlock()
	s.ipv6Address = netip.Addr{}
}
