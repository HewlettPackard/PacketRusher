/** SPDX-License-Identifier: Apache-2.0 */
package nas

import (
	"bytes"
	"fmt"
	"github.com/free5gc/nas/ie"
	nas "github.com/free5gc/nas/message"
	log "github.com/sirupsen/logrus"
	"my5G-RANTester/internal/common/auth"
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas/handler"
)

// DecodeNAS verifies protected NAS before committing negotiated algorithms or counters.
func DecodeNAS(ue *context.UEContext, packet []byte) (nas.Message, error) {
	if ue == nil {
		return nil, fmt.Errorf("UE context is nil")
	}
	// Plain 5GSM messages carry a session ID where 5GMM carries its security header.
	if len(packet) > 0 && packet[0] == byte(nas.Epd5GSSessMgmtMsg) {
		return nas.Parse(packet, nil)
	}
	st := nas.GetSecHdrType(packet)
	if st == nas.SecHdrTypePlainNas {
		msg, err := nas.Parse(packet, nil)
		if err != nil {
			return nil, err
		}
		if err := validateSecurityModeCommandHeader(msg, st); err != nil {
			return nil, err
		}
		return msg, nil
	}
	if len(packet) < int(nas.SecHdrLen) {
		return nil, fmt.Errorf("truncated protected NAS header")
	}
	if st == nas.SecHdrTypeIntegrityProtectedAndCipheredWithNew5gNasSecCtx {
		return nil, fmt.Errorf("downlink NAS uses security header reserved for Security Mode Complete")
	}
	candidate := ue.NASSecurityContext().Clone()
	if st == nas.SecHdrTypeIntegrityProtectedWithNew5gNasSecCtx {
		preliminary, err := nas.Parse(packet[nas.SecHdrLen:], nil)
		if err != nil {
			return nil, err
		}
		command, ok := preliminary.(*nas.SecModeCmd)
		if !ok || command.SelectedNASSecAlgos == nil {
			return nil, fmt.Errorf("new NAS security context requires Security Mode Command")
		}
		if err := validateSecurityModeCommand(ue, command); err != nil {
			return nil, err
		}
		candidate.CipheringAlg = command.SelectedNASSecAlgos.CipheringAlgo
		candidate.IntegrityAlg = command.SelectedNASSecAlgos.MsgIntAlgo
		candidate.DownlinkCount.Set(0, 0)
		if err = auth.AlgorithmKeyDerivation(uint8(candidate.CipheringAlg), ue.UeSecurity.Kamf, &candidate.KnasEnc, uint8(candidate.IntegrityAlg), &candidate.KnasInt); err != nil {
			return nil, err
		}
	}
	msg, err := nas.Parse(packet, candidate)
	if err != nil {
		return nil, err
	}
	if err := validateSecurityModeCommandHeader(msg, st); err != nil {
		return nil, err
	}
	ue.UeSecurity.ULCount, ue.UeSecurity.DLCount = *candidate.UplinkCount, *candidate.DownlinkCount
	ue.UeSecurity.CipheringAlg, ue.UeSecurity.IntegrityAlg = uint8(candidate.CipheringAlg), uint8(candidate.IntegrityAlg)
	ue.UeSecurity.KnasEnc, ue.UeSecurity.KnasInt = candidate.KnasEnc, candidate.KnasInt
	return msg, nil
}

// TS 24.501 section 5.4.2.2 requires this outer header even for rekeying.
func validateSecurityModeCommandHeader(msg nas.Message, st nas.SecHdrType) error {
	if _, ok := msg.(*nas.SecModeCmd); ok && st != nas.SecHdrTypeIntegrityProtectedWithNew5gNasSecCtx {
		return fmt.Errorf("Security Mode Command requires integrity protection with a new NAS security context")
	}
	return nil
}

func validateSecurityModeCommand(ue *context.UEContext, command *nas.SecModeCmd) error {
	capability := ue.GetUeSecurityCapability()
	if capability == nil || command.ReplayedUESecCapabilities == nil || command.Ngksi == nil || command.Ngksi.Ksi == ie.NASKeyNA {
		return fmt.Errorf("Security Mode Command is missing the UE security capability or valid key identifier")
	}
	ciphering := [...]bool{capability.EA05G, capability.EA1_128_5G, capability.EA2_128_5G, capability.EA3_128_5G}
	integrity := [...]bool{capability.IA05G, capability.IA1_128_5G, capability.IA2_128_5G, capability.IA3_128_5G}
	selected := command.SelectedNASSecAlgos
	if int(selected.CipheringAlgo) >= len(ciphering) || !ciphering[selected.CipheringAlgo] || int(selected.MsgIntAlgo) >= len(integrity) || !integrity[selected.MsgIntAlgo] {
		return fmt.Errorf("Security Mode Command selected an algorithm the UE did not offer")
	}
	offered, err := capability.MarshalBinary()
	if err != nil {
		return err
	}
	replayed, err := command.ReplayedUESecCapabilities.MarshalBinary()
	if err != nil {
		return err
	}
	if !bytes.Equal(offered, replayed) {
		return fmt.Errorf("Security Mode Command changed the replayed UE security capability")
	}
	return nil
}

func DispatchNas(ue *context.UEContext, packet []byte) {
	msg, err := DecodeNAS(ue, packet)
	if err != nil {
		log.Errorf("[UE][NAS] Decode failed: %v", err)
		return
	}
	log.Infof("[UE][NAS] Receive %s", msg.MsgType())
	switch m := msg.(type) {
	case *nas.AuthReq:
		handler.HandlerAuthenticationRequest(ue, m)
	case *nas.AuthRej:
		handler.HandlerAuthenticationReject(ue, m)
	case *nas.IdReq:
		handler.HandlerIdentityRequest(ue, m)
	case *nas.SecModeCmd:
		handler.HandlerSecurityModeCommand(ue, m)
	case *nas.RegAccept:
		handler.HandlerRegistrationAccept(ue, m)
	case *nas.CfgUpdateCmd:
		handler.HandlerConfigurationUpdateCommand(ue, m)
	case *nas.DLNASTransport:
		if m.Cause5GMM != nil {
			log.Errorf("[UE][NAS] 5GMM failure: %s", m.Cause5GMM)
		}
		handler.HandlerDlNasTransportPduaccept(ue, m)
	case *nas.SvcAccept:
		handler.HandlerServiceAccept(ue, m)
	case *nas.SvcRej:
		log.Errorf("[UE][NAS] Service Reject: %s", m.Cause5GMM)
	case *nas.RegRej:
		ue.RegistrationFailed()
		log.Errorf("[UE][NAS] Registration Reject: %s", m.Cause5GMM)
	case *nas.Status5GMM:
		log.Errorf("[UE][NAS] 5GMM status: %s", m.Cause5GMM)
	case *nas.Status5GSM:
		log.Errorf("[UE][NAS] 5GSM status: %s", m.Cause5GSM)
	default:
		log.Warnf("[UE][NAS] Unsupported message %s", msg.MsgType())
	}
}
