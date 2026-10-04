/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */
package nas

import (
	"testing"
	"time"

	"github.com/free5gc/nas/ie"
	"github.com/stretchr/testify/require"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/analytics"
	"my5G-RANTester/internal/common/sidf"
	gnbcontext "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/scenario"
)

func transportRefusalUE(t *testing.T) (*context.UEContext, *context.UEPDUSession) {
	t.Helper()
	ue := &context.UEContext{Results: analytics.NewRecorder()}
	ue.NewRanUeContext("0000000001", &ie.UESecCapability{Length: 2, EA2_128_5G: true, IA2_128_5G: true},
		"", "", "", "", "", "001", "01", sidf.HomeNetworkPublicKey{},
		"0000", "internet", 1, "", config.TunnelDisabled, make(chan scenario.ScenarioMessage, 16), nil, 1)
	ue.SetGnbRx(make(chan gnbcontext.UEMessage, 2))
	ue.UeSecurity.CipheringAlg, ue.UeSecurity.IntegrityAlg = 2, 2
	ue.UeSecurity.KnasEnc, ue.UeSecurity.KnasInt = [16]byte{5, 6, 7, 8}, [16]byte{1, 2, 3, 4}
	t.Cleanup(ue.Terminate)
	session, err := ue.CreatePDUSession()
	require.NoError(t, err)
	return ue, session
}

// Independent native wire: returned establishment request (C1), mandatory
// integrity-rate bytes, IPv4 type, matching session IE, and 5GMM DNN refusal 5B.
func returnedEstablishmentWire(innerID, outerID, pti uint8, cause bool) []byte {
	inner := []byte{0x2e, innerID, pti, 0xc1, 0xff, 0xff, 0x91}
	outer := []byte{0x7e, 0, 0x68, 1, 0, byte(len(inner))}
	outer = append(outer, inner...)
	outer = append(outer, 0x12, outerID)
	if cause {
		outer = append(outer, 0x58, 0x5b)
	}
	return outer
}

func TestTransportRefusalAccountsForInitialAndRetriedEstablishment(t *testing.T) {
	for _, retrying := range []bool{false, true} {
		for _, protected := range []bool{false, true} {
			name := "initial/plain"
			if retrying {
				name = "retry/plain"
			}
			if protected {
				name = name[:len(name)-len("plain")] + "protected"
			}
			t.Run(name, func(t *testing.T) {
				ue, session := transportRefusalUE(t)
				encode := func() ([]byte, error) { return []byte{1}, nil }
				require.NoError(t, ue.StartPduSessionRequest(session, encode))
				<-ue.GetGnbRx()
				if retrying {
					session.EstablishmentFailed()
					require.True(t, ue.SchedulePduSessionRetry(session))
					select {
					case retry := <-ue.PduSessionRetries():
						require.NoError(t, ue.StartPduSessionRetry(retry, encode))
					case <-time.After(3 * time.Second):
						t.Fatal("retry was not handed to the UE loop")
					}
					<-ue.GetGnbRx()
				}
				plain := returnedEstablishmentWire(session.Id, session.Id, 1, true)
				for i := 0; i < 2; i++ {
					packet := plain
					if protected {
						packet = protect(ue, plain, uint32(7+i))
					}
					DispatchNas(ue, packet)
				}
				want := uint64(1)
				if retrying {
					want = 2
				}
				for _, p := range ue.Results.Snapshot().Procedures {
					if p.Procedure == analytics.SessionEstablishment {
						require.Equal(t, want, p.Started)
						require.Equal(t, want, p.Failure)
						require.Zero(t, p.Pending)
						require.Zero(t, p.Success)
						require.Zero(t, p.Cancelled)
						require.Zero(t, p.LatencyCount)
					}
				}
				require.Equal(t, context.SM5G_PDU_SESSION_ACTIVE_PENDING, session.GetStateSM(), "accounting must not invent a NAS state transition")
				require.Empty(t, ue.GetGnbRx(), "accounting must not send a retry")
				select {
				case <-ue.PduSessionRetries():
					t.Fatal("transport accounting scheduled an automatic retry")
				case <-time.After(20 * time.Millisecond):
				}
				ue.Terminate()
				ue.Results.Close()
				for _, p := range ue.Results.Snapshot().Procedures {
					if p.Procedure == analytics.SessionEstablishment {
						require.Equal(t, want, p.Failure)
						require.Zero(t, p.Cancelled, "a refused request must not later become cancellation")
					}
				}
			})
		}
	}
}

func TestTransportRefusalIgnoresUnmatchedOrEndedEstablishment(t *testing.T) {
	for _, name := range []string{"no cause", "mismatched envelope", "wrong PTI", "unknown session", "missing session IE", "wrong payload type", "active", "inactive", "deleted", "terminated", "ID reused before request"} {
		t.Run(name, func(t *testing.T) {
			ue, session := transportRefusalUE(t)
			session.SetStateSM_PDU_SESSION_PENDING()
			packet := returnedEstablishmentWire(session.Id, session.Id, 1, true)
			switch name {
			case "no cause":
				packet = returnedEstablishmentWire(session.Id, session.Id, 1, false)
			case "mismatched envelope":
				packet = returnedEstablishmentWire(session.Id, 2, 1, true)
			case "wrong PTI":
				packet = returnedEstablishmentWire(session.Id, session.Id, 2, true)
			case "unknown session":
				packet = returnedEstablishmentWire(2, 2, 1, true)
			case "missing session IE":
				packet = append(packet[:len(packet)-4:len(packet)-4], 0x58, 0x5b)
			case "wrong payload type":
				packet[3] = 2
			case "active":
				session.SetStateSM_PDU_SESSION_ACTIVE()
			case "inactive":
				session.SetStateSM_PDU_SESSION_INACTIVE()
			case "deleted":
				require.NoError(t, ue.DeletePduSession(session.Id))
			case "terminated":
				ue.Terminate()
			case "ID reused before request":
				require.NoError(t, ue.DeletePduSession(session.Id))
				replacement, err := ue.CreatePDUSession()
				require.NoError(t, err)
				require.Equal(t, session.Id, replacement.Id)
			}
			before := ue.Results.Snapshot()
			state := session.GetStateSM()
			DispatchNas(ue, packet)
			require.Equal(t, before, ue.Results.Snapshot(), "unmatched/ended request must not change outcomes")
			require.Equal(t, state, session.GetStateSM(), "unmatched/ended request must not change session state")
		})
	}
}
