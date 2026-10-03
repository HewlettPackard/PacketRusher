/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package context

import (
	"errors"
	"log"
	"net"
	"net/netip"
	"sync"

	"github.com/free5gc/openapi/models"
)

type SessionContext struct {
	dataNetworks      []DataNetwork
	sessionsRules     []*models.Pcf_SMPolCtrl_SessionRule
	lastAllocatedIP   net.IP
	lastAllocatedIPv6 uint64
	n3                net.IP
	ipMtx             sync.Mutex
}

type DataNetwork struct {
	Dnn string
	Dns DNS
}

type DNS struct {
	IPv4Addr net.IP
	IPv6Addr net.IP
}

func (s *SessionContext) NewSessionContext() {
	//TODO parametrize session data
	s.dataNetworks = []DataNetwork{
		{
			Dnn: "internet",
			Dns: DNS{
				IPv4Addr: net.ParseIP("8.8.8.8"),
				IPv6Addr: net.ParseIP("2001:4860:4860::8888"),
			},
		},
	}

	s.sessionsRules = []*models.Pcf_SMPolCtrl_SessionRule{{
		AuthSessAmbr: &models.Ambr{
			Uplink:   "1 Gbps",
			Downlink: "1 Gbps",
		},
		AuthDefQos: &models.Pcf_SMPolCtrl_AuthorizedDefaultQos{
			Var5qi: 6,
			Arp: &models.Arp{
				PriorityLevel: 8,
			},
			PriorityLevel: 8,
		},
		SessRuleId: "SessRuleId-1",
	}}

	s.lastAllocatedIP = net.ParseIP("10.0.0.1")
	s.n3 = net.ParseIP("127.0.0.1").To4()
}

func (s *SessionContext) GetN3() net.IP {
	return s.n3
}

func (s *SessionContext) GetDnnList() []string {
	dnn := []string{}
	for _, dn := range s.dataNetworks {
		dnn = append(dnn, dn.Dnn)
	}
	return dnn
}

func (s *SessionContext) GetSessionRules() []*models.Pcf_SMPolCtrl_SessionRule {
	return s.sessionsRules
}

func (s *SessionContext) GetDataNetwork(dnn string) (DataNetwork, error) {
	for _, dn := range s.dataNetworks {
		if dn.Dnn == dnn {
			return dn, nil
		}
	}
	return DataNetwork{}, errors.New("[5GC] Could not find requested datanetwork")
}

func (s *SessionContext) GetUnallocatedIP() net.IP {
	s.ipMtx.Lock()
	defer s.ipMtx.Unlock()
	next := netip.MustParseAddr(s.lastAllocatedIP.String()).Next()
	var err error
	if !netip.MustParsePrefix("10.0.0.0/8").Contains(next) {
		err = errors.New("mock IPv4 pool exhausted")
	}
	ip := next.String()
	if err != nil {
		log.Fatal("[5GC][NAS] Error while allocating ip for PDU session: " + err.Error())
	}

	s.lastAllocatedIP = net.ParseIP(ip).To4()
	return s.lastAllocatedIP
}

// GetUnallocatedIPv6 allocates one deterministic /64 per mock PDU session.
// The prefix is advertised by a fake UPF in dataplane tests, not encoded in NAS.
func (s *SessionContext) GetUnallocatedIPv6() net.IP {
	s.ipMtx.Lock()
	defer s.ipMtx.Unlock()
	s.lastAllocatedIPv6++
	bytes := netip.MustParseAddr("2001:db8::1").As16()
	bytes[4], bytes[5], bytes[6], bytes[7] = byte(s.lastAllocatedIPv6>>24), byte(s.lastAllocatedIPv6>>16), byte(s.lastAllocatedIPv6>>8), byte(s.lastAllocatedIPv6)
	return net.IP(bytes[:])
}
