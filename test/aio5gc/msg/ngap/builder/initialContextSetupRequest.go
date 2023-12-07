/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package builder

import (
	"my5G-RANTester/lib/ngap"
	"my5G-RANTester/lib/ngap/ngapConvert"
	"my5G-RANTester/lib/ngap/ngapType"
	"my5G-RANTester/test/aio5gc/context"
)

func InitialContextSetupRequest(nasPdu []byte, ue *context.UEContext, amf context.AMFContext) ([]byte, error) {
	message, err := buildInitialContextSetupRequest(nasPdu, ue, amf)
	if err != nil {
		return []byte{}, err
	}
	return ngap.Encoder(message)
}

func buildInitialContextSetupRequest(nasPdu []byte, ue *context.UEContext, amf context.AMFContext) (ngapType.NGAPPDU, error) {
	var pdu ngapType.NGAPPDU

	pdu.Present = ngapType.NGAPPDUPresentInitiatingMessage
	pdu.InitiatingMessage = new(ngapType.InitiatingMessage)

	initiatingMessage := pdu.InitiatingMessage
	initiatingMessage.ProcedureCode.Value = ngapType.ProcedureCodeInitialContextSetup
	initiatingMessage.Criticality.Value = ngapType.CriticalityPresentReject

	initiatingMessage.Value.Present = ngapType.InitiatingMessagePresentInitialContextSetupRequest
	initiatingMessage.Value.InitialContextSetupRequest = new(ngapType.InitialContextSetupRequest)

	initialContextSetupRequest := initiatingMessage.Value.InitialContextSetupRequest
	initialContextSetupRequestIEs := &initialContextSetupRequest.ProtocolIEs

	// AMF UE NGAP ID
	ie := ngapType.InitialContextSetupRequestIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDAMFUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.InitialContextSetupRequestIEsPresentAMFUENGAPID
	ie.Value.AMFUENGAPID = new(ngapType.AMFUENGAPID)

	aMFUENGAPID := ie.Value.AMFUENGAPID
	aMFUENGAPID.Value = ue.GetAmfNgapId()

	initialContextSetupRequestIEs.List = append(initialContextSetupRequestIEs.List, ie)

	// RAN UE NGAP ID
	ie = ngapType.InitialContextSetupRequestIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDRANUENGAPID
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.InitialContextSetupRequestIEsPresentRANUENGAPID
	ie.Value.RANUENGAPID = new(ngapType.RANUENGAPID)

	rANUENGAPID := ie.Value.RANUENGAPID
	rANUENGAPID.Value = ue.GetRanNgapId()

	initialContextSetupRequestIEs.List = append(initialContextSetupRequestIEs.List, ie)

	// GUAMI
	ie = ngapType.InitialContextSetupRequestIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDGUAMI
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.InitialContextSetupRequestIEsPresentGUAMI
	ie.Value.GUAMI = new(ngapType.GUAMI)

	guami := ie.Value.GUAMI
	plmnID := &guami.PLMNIdentity
	amfRegionID := &guami.AMFRegionID
	amfSetID := &guami.AMFSetID
	amfPtrID := &guami.AMFPointer

	servedGuami := amf.GetServedGuami()[0]

	*plmnID = ngapConvert.PlmnIdToNgap(*servedGuami.PlmnId)
	amfRegionID.Value, amfSetID.Value, amfPtrID.Value = ngapConvert.AmfIdToNgap(servedGuami.AmfId)

	initialContextSetupRequestIEs.List = append(initialContextSetupRequestIEs.List, ie)

	// Allowed NSSAI
	ie = ngapType.InitialContextSetupRequestIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDAllowedNSSAI
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.InitialContextSetupRequestIEsPresentAllowedNSSAI
	ie.Value.AllowedNSSAI = new(ngapType.AllowedNSSAI)

	allowedNSSAI := ie.Value.AllowedNSSAI
	// Supporting only one plmn for now
	for _, allowedSnssai := range amf.GetSupportedPlmnSnssai()[0].SNssaiList {
		allowedNSSAIItem := ngapType.AllowedNSSAIItem{}
		ngapSnssai := ngapConvert.SNssaiToNgap(allowedSnssai)
		allowedNSSAIItem.SNSSAI = ngapSnssai
		allowedNSSAI.List = append(allowedNSSAI.List, allowedNSSAIItem)
	}

	initialContextSetupRequestIEs.List = append(initialContextSetupRequestIEs.List, ie)

	// UE Security Capabilities
	ie = ngapType.InitialContextSetupRequestIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDUESecurityCapabilities
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.InitialContextSetupRequestIEsPresentUESecurityCapabilities
	ie.Value.UESecurityCapabilities = new(ngapType.UESecurityCapabilities)

	ueSecurityCapabilities := ie.Value.UESecurityCapabilities
	nrEncryptionAlgorighm := []byte{0x00, 0x00}

	nrEncryptionAlgorighm[0] |= ue.GetSecurityCapability().GetEA1_128_5G() << 7
	nrEncryptionAlgorighm[0] |= ue.GetSecurityCapability().GetEA2_128_5G() << 6
	nrEncryptionAlgorighm[0] |= ue.GetSecurityCapability().GetEA3_128_5G() << 5
	ueSecurityCapabilities.NRencryptionAlgorithms.Value = ngapConvert.ByteToBitString(nrEncryptionAlgorighm, 16)

	nrIntegrityAlgorithm := []byte{0x00, 0x00}

	nrIntegrityAlgorithm[0] |= ue.GetSecurityCapability().GetIA1_128_5G() << 7
	nrIntegrityAlgorithm[0] |= ue.GetSecurityCapability().GetIA2_128_5G() << 6
	nrIntegrityAlgorithm[0] |= ue.GetSecurityCapability().GetIA3_128_5G() << 5

	ueSecurityCapabilities.NRintegrityProtectionAlgorithms.Value = ngapConvert.ByteToBitString(nrIntegrityAlgorithm, 16)

	// only support NR algorithms
	eutraEncryptionAlgorithm := []byte{0x00, 0x00}
	ueSecurityCapabilities.EUTRAencryptionAlgorithms.Value = ngapConvert.ByteToBitString(eutraEncryptionAlgorithm, 16)

	eutraIntegrityAlgorithm := []byte{0x00, 0x00}
	ueSecurityCapabilities.EUTRAintegrityProtectionAlgorithms.Value = ngapConvert.
		ByteToBitString(eutraIntegrityAlgorithm, 16)

	initialContextSetupRequestIEs.List = append(initialContextSetupRequestIEs.List, ie)

	// Security Key
	ie = ngapType.InitialContextSetupRequestIEs{}
	ie.Id.Value = ngapType.ProtocolIEIDSecurityKey
	ie.Criticality.Value = ngapType.CriticalityPresentReject
	ie.Value.Present = ngapType.InitialContextSetupRequestIEsPresentSecurityKey
	ie.Value.SecurityKey = new(ngapType.SecurityKey)

	securityKey := ie.Value.SecurityKey
	securityKey.Value = ngapConvert.ByteToBitString(ue.GetSecurityContext().GetKGNB(), 256)

	initialContextSetupRequestIEs.List = append(initialContextSetupRequestIEs.List, ie)

	// NAS-PDU (optional)
	if nasPdu != nil {
		ie = ngapType.InitialContextSetupRequestIEs{}
		ie.Id.Value = ngapType.ProtocolIEIDNASPDU
		ie.Criticality.Value = ngapType.CriticalityPresentIgnore
		ie.Value.Present = ngapType.InitialContextSetupRequestIEsPresentNASPDU
		ie.Value.NASPDU = new(ngapType.NASPDU)

		ie.Value.NASPDU.Value = nasPdu

		initialContextSetupRequestIEs.List = append(initialContextSetupRequestIEs.List, ie)
	}

	return pdu, nil
}
