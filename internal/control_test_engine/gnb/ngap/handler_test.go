/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package ngap

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"net/netip"
	"testing"

	"github.com/free5gc/aper"
	"github.com/free5gc/ngap/ngapType"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestGNBContext() *context.GNBContext {
	gnb := &context.GNBContext{}
	gnb.NewRanGnbContext("test-gnb", "001", "01", "000001", "1", "000001",
		netip.MustParseAddrPort("127.0.0.1:9999"),
		netip.MustParseAddrPort("127.0.0.1:2152"))

	// Create and activate an AMF for UE operations
	amf := gnb.NewGnBAmf(netip.MustParseAddrPort("127.0.0.1:38412"))
	if amf != nil {
		amf.SetStateActive()
	}

	return gnb
}

func createTestUE(gnb *context.GNBContext, prUeId int64) *context.GNBUe {
	gnbTx := make(chan context.UEMessage, 10)
	gnbRx := make(chan context.UEMessage, 10)
	ue, _ := gnb.NewGnBUe(gnbTx, gnbRx, prUeId, nil)
	return ue
}

func TestGetUeFromContext_ValidUE(t *testing.T) {
	gnb := createTestGNBContext()
	ue := createTestUE(gnb, 12345)

	ranUeId := ue.GetRanUeId()
	amfUeId := int64(67890)

	// Test with valid UE
	retrievedUe := getUeFromContext(gnb, ranUeId, amfUeId)
	assert.NotNil(t, retrievedUe, "Should retrieve valid UE")
	assert.Equal(t, ue, retrievedUe, "Retrieved UE should match original")
	assert.Equal(t, amfUeId, retrievedUe.GetAmfUeId(), "AMF UE ID should be set")
}

func TestGetUeFromContext_NonExistentUE(t *testing.T) {
	gnb := createTestGNBContext()

	// Test with non-existent UE ID
	nonExistentRanUeId := int64(99999)
	amfUeId := int64(67890)

	retrievedUe := getUeFromContext(gnb, nonExistentRanUeId, amfUeId)
	assert.Nil(t, retrievedUe, "Should return nil for non-existent UE")
}

func TestGetUeFromContext_DownStateUE(t *testing.T) {
	gnb := createTestGNBContext()
	ue := createTestUE(gnb, 12345)

	ranUeId := ue.GetRanUeId()
	amfUeId := int64(67890)

	// Set UE to Down state
	ue.SetStateDown()

	// getUeFromContext must still return the UE even in Down state: individual
	// handlers (e.g. HandlerPduSessionReleaseCommand) must be able to act on
	// mid-teardown UEs so that the AMF receives a proper response.
	retrievedUe := getUeFromContext(gnb, ranUeId, amfUeId)
	assert.NotNil(t, retrievedUe, "Should still return a Down-state UE so handlers can send a response")
}

func TestGetUeFromContext_DeletedUE(t *testing.T) {
	gnb := createTestGNBContext()
	ue := createTestUE(gnb, 12345)

	ranUeId := ue.GetRanUeId()
	amfUeId := int64(67890)

	// Delete the UE
	gnb.DeleteGnBUe(ue)

	// Test with deleted UE
	retrievedUe := getUeFromContext(gnb, ranUeId, amfUeId)
	assert.Nil(t, retrievedUe, "Should return nil for deleted UE")
}

func TestHandlerUeContextReleaseCommand_ValidUE(t *testing.T) {
	gnb := createTestGNBContext()
	ue := createTestUE(gnb, 12345)

	ranUeId := ue.GetRanUeId()

	// Create UE Context Release Command message
	message := &ngapType.NGAPPDU{
		Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentUEContextReleaseCommand,
				UEContextReleaseCommand: &ngapType.UEContextReleaseCommand{
					ProtocolIEs: ngapType.ProtocolIEContainerUEContextReleaseCommandIEs{
						List: []ngapType.UEContextReleaseCommandIEs{
							{
								Id: ngapType.ProtocolIEID{
									Value: ngapType.ProtocolIEIDUENGAPIDs,
								},
								Value: ngapType.UEContextReleaseCommandIEsValue{
									Present: ngapType.UEContextReleaseCommandIEsPresentUENGAPIDs,
									UENGAPIDs: &ngapType.UENGAPIDs{
										Present: ngapType.UENGAPIDsPresentUENGAPIDPair,
										UENGAPIDPair: &ngapType.UENGAPIDPair{
											RANUENGAPID: ngapType.RANUENGAPID{
												Value: ranUeId,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Verify UE exists before release
	retrievedUe, err := gnb.GetGnbUe(ranUeId)
	require.NoError(t, err, "UE should exist before release")
	require.NotNil(t, retrievedUe, "UE should not be nil before release")

	// Call handler
	HandlerUeContextReleaseCommand(gnb, message)

	// Verify UE is deleted after release
	retrievedUe, err = gnb.GetGnbUe(ranUeId)
	assert.Error(t, err, "UE should not exist after release")
	assert.Nil(t, retrievedUe, "UE should be nil after release")
}

func TestHandlerUeContextReleaseCommand_NonExistentUE(t *testing.T) {
	gnb := createTestGNBContext()

	nonExistentRanUeId := int64(99999)

	// Create UE Context Release Command message for non-existent UE
	message := &ngapType.NGAPPDU{
		Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentUEContextReleaseCommand,
				UEContextReleaseCommand: &ngapType.UEContextReleaseCommand{
					ProtocolIEs: ngapType.ProtocolIEContainerUEContextReleaseCommandIEs{
						List: []ngapType.UEContextReleaseCommandIEs{
							{
								Id: ngapType.ProtocolIEID{
									Value: ngapType.ProtocolIEIDUENGAPIDs,
								},
								Value: ngapType.UEContextReleaseCommandIEsValue{
									Present: ngapType.UEContextReleaseCommandIEsPresentUENGAPIDs,
									UENGAPIDs: &ngapType.UENGAPIDs{
										Present: ngapType.UENGAPIDsPresentUENGAPIDPair,
										UENGAPIDPair: &ngapType.UENGAPIDPair{
											RANUENGAPID: ngapType.RANUENGAPID{
												Value: nonExistentRanUeId,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Call handler - should not panic or cause issues
	HandlerUeContextReleaseCommand(gnb, message)

	// Verify no UE was affected
	retrievedUe, err := gnb.GetGnbUe(nonExistentRanUeId)
	assert.Error(t, err, "Non-existent UE should remain non-existent")
	assert.Nil(t, retrievedUe, "Non-existent UE should remain nil")
}

func TestHandlerUeContextReleaseCommand_MissingUEID(t *testing.T) {
	gnb := createTestGNBContext()

	// Create UE Context Release Command message without UE ID
	message := &ngapType.NGAPPDU{
		Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentUEContextReleaseCommand,
				UEContextReleaseCommand: &ngapType.UEContextReleaseCommand{
					ProtocolIEs: ngapType.ProtocolIEContainerUEContextReleaseCommandIEs{
						List: []ngapType.UEContextReleaseCommandIEs{
							// Empty list - no UE ID provided
						},
					},
				},
			},
		},
	}

	// Call handler - should not panic
	HandlerUeContextReleaseCommand(gnb, message)

	// No specific assertions needed - the test passes if no panic occurs
}

func TestNGAPHandlers_ConcurrentProcessing(t *testing.T) {
	// Test concurrent processing of NGAP messages to ensure thread safety
	gnb := createTestGNBContext()

	const numUEs = 50
	ues := make([]*context.GNBUe, numUEs)

	// Create multiple UEs
	for i := 0; i < numUEs; i++ {
		ues[i] = createTestUE(gnb, int64(i+1000))
	}

	// Concurrently process UE context release commands
	done := make(chan bool, numUEs)

	for i := 0; i < numUEs; i++ {
		go func(ueIndex int) {
			ue := ues[ueIndex]
			ranUeId := ue.GetRanUeId()

			// Create and process UE Context Release Command
			message := &ngapType.NGAPPDU{
				Present: ngapType.NGAPPDUPresentInitiatingMessage,
				InitiatingMessage: &ngapType.InitiatingMessage{
					Value: ngapType.InitiatingMessageValue{
						Present: ngapType.InitiatingMessagePresentUEContextReleaseCommand,
						UEContextReleaseCommand: &ngapType.UEContextReleaseCommand{
							ProtocolIEs: ngapType.ProtocolIEContainerUEContextReleaseCommandIEs{
								List: []ngapType.UEContextReleaseCommandIEs{
									{
										Id: ngapType.ProtocolIEID{
											Value: ngapType.ProtocolIEIDUENGAPIDs,
										},
										Value: ngapType.UEContextReleaseCommandIEsValue{
											Present: ngapType.UEContextReleaseCommandIEsPresentUENGAPIDs,
											UENGAPIDs: &ngapType.UENGAPIDs{
												Present: ngapType.UENGAPIDsPresentUENGAPIDPair,
												UENGAPIDPair: &ngapType.UENGAPIDPair{
													RANUENGAPID: ngapType.RANUENGAPID{
														Value: ranUeId,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}

			HandlerUeContextReleaseCommand(gnb, message)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numUEs; i++ {
		<-done
	}

	// Verify all UEs were deleted
	for i := 0; i < numUEs; i++ {
		ue := ues[i]
		ranUeId := ue.GetRanUeId()

		retrievedUe, err := gnb.GetGnbUe(ranUeId)
		assert.Error(t, err, "UE %d should be deleted", i)
		assert.Nil(t, retrievedUe, "UE %d should be nil after deletion", i)
	}
}

// pduSessionResourceSetupRequest builds a request for one PDU session on the given slice,
// carrying the transfer IEs the handler reads (UL tunnel, QoS flow, session type).
func pduSessionResourceSetupRequest(t *testing.T, ue *context.GNBUe, sst []byte, sd []byte) *ngapType.NGAPPDU {
	return pduSessionResourceSetupRequestWith(t, ue, sst, sd, true)
}

func pduSessionResourceSetupRequestWith(t *testing.T, ue *context.GNBUe, sst []byte, sd []byte, withUlTunnel bool) *ngapType.NGAPPDU {
	t.Helper()

	transfer := ngapType.PDUSessionResourceSetupRequestTransfer{}
	transfer.ProtocolIEs.List = []ngapType.PDUSessionResourceSetupRequestTransferIEs{
		{
			Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDULNGUUPTNLInformation},
			Value: ngapType.PDUSessionResourceSetupRequestTransferIEsValue{
				Present: ngapType.PDUSessionResourceSetupRequestTransferIEsPresentULNGUUPTNLInformation,
				ULNGUUPTNLInformation: &ngapType.UPTransportLayerInformation{
					Present: ngapType.UPTransportLayerInformationPresentGTPTunnel,
					GTPTunnel: &ngapType.GTPTunnel{
						TransportLayerAddress: ngapType.TransportLayerAddress{
							Value: aper.BitString{Bytes: []byte{10, 0, 0, 1}, BitLength: 32},
						},
						GTPTEID: ngapType.GTPTEID{Value: []byte{0, 0, 0, 1}},
					},
				},
			},
		},
		{
			Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDPDUSessionType},
			Value: ngapType.PDUSessionResourceSetupRequestTransferIEsValue{
				Present:        ngapType.PDUSessionResourceSetupRequestTransferIEsPresentPDUSessionType,
				PDUSessionType: &ngapType.PDUSessionType{Value: ngapType.PDUSessionTypePresentIpv4},
			},
		},
		{
			Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDQosFlowSetupRequestList},
			Value: ngapType.PDUSessionResourceSetupRequestTransferIEsValue{
				Present: ngapType.PDUSessionResourceSetupRequestTransferIEsPresentQosFlowSetupRequestList,
				QosFlowSetupRequestList: &ngapType.QosFlowSetupRequestList{
					List: []ngapType.QosFlowSetupRequestItem{
						{
							QosFlowIdentifier: ngapType.QosFlowIdentifier{Value: 1},
							QosFlowLevelQosParameters: ngapType.QosFlowLevelQosParameters{
								QosCharacteristics: ngapType.QosCharacteristics{
									Present:       ngapType.QosCharacteristicsPresentNonDynamic5QI,
									NonDynamic5QI: &ngapType.NonDynamic5QIDescriptor{FiveQI: ngapType.FiveQI{Value: 9}},
								},
								AllocationAndRetentionPriority: ngapType.AllocationAndRetentionPriority{
									PriorityLevelARP: ngapType.PriorityLevelARP{Value: 1},
								},
							},
						},
					},
				},
			},
		},
	}
	if !withUlTunnel {
		transfer.ProtocolIEs.List = transfer.ProtocolIEs.List[1:]
	}
	encodedTransfer, err := aper.MarshalWithParams(transfer, "valueExt")
	require.NoError(t, err)

	return &ngapType.NGAPPDU{
		Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentPDUSessionResourceSetupRequest,
				PDUSessionResourceSetupRequest: &ngapType.PDUSessionResourceSetupRequest{
					ProtocolIEs: ngapType.ProtocolIEContainerPDUSessionResourceSetupRequestIEs{
						List: []ngapType.PDUSessionResourceSetupRequestIEs{
							{
								Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDAMFUENGAPID},
								Value: ngapType.PDUSessionResourceSetupRequestIEsValue{
									Present:     ngapType.PDUSessionResourceSetupRequestIEsPresentAMFUENGAPID,
									AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 67890},
								},
							},
							{
								Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDRANUENGAPID},
								Value: ngapType.PDUSessionResourceSetupRequestIEsValue{
									Present:     ngapType.PDUSessionResourceSetupRequestIEsPresentRANUENGAPID,
									RANUENGAPID: &ngapType.RANUENGAPID{Value: ue.GetRanUeId()},
								},
							},
							{
								Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDPDUSessionResourceSetupListSUReq},
								Value: ngapType.PDUSessionResourceSetupRequestIEsValue{
									Present: ngapType.PDUSessionResourceSetupRequestIEsPresentPDUSessionResourceSetupListSUReq,
									PDUSessionResourceSetupListSUReq: &ngapType.PDUSessionResourceSetupListSUReq{
										List: []ngapType.PDUSessionResourceSetupItemSUReq{
											{
												PDUSessionID: ngapType.PDUSessionID{Value: 1},
												SNSSAI: ngapType.SNSSAI{
													SST: ngapType.SST{Value: sst},
													SD:  &ngapType.SD{Value: sd},
												},
												PDUSessionResourceSetupRequestTransfer: encodedTransfer,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// When every requested session is skipped -- here because the slice is not one the UE
// selected -- the Setup Response would be built from an empty list, which the encoder
// rejects. That used to be fatal to the whole simulator. The handler must return without
// sending anything and without exiting.
func TestHandlerPduSessionResourceSetupRequest_NoSessionSetUp(t *testing.T) {
	gnb := createTestGNBContext()
	ue := createTestUE(gnb, 12345)
	ue.CreateUeContext("not informed", "", []string{"01"}, []string{"010203"}, nil)

	HandlerPduSessionResourceSetupRequest(gnb, pduSessionResourceSetupRequest(t, ue, []byte{0x02}, []byte{0x01, 0x02, 0x03}))

	pduSession, err := ue.GetPduSession(1)
	require.NoError(t, err)
	assert.Nil(t, pduSession, "no PDU session should exist for an unselected slice")
	assert.NotEqual(t, context.Ready, ue.GetState(), "no Setup Response should have been built")
	assert.Empty(t, ue.GetGnbTx(), "nothing should have been sent to the UE")
}

// The guard above must not stop a legitimate Setup Response: with one session set up,
// the response is built (the UE becomes Ready) and the session is handed to the UE.
func TestHandlerPduSessionResourceSetupRequest_SessionSetUp(t *testing.T) {
	gnb := createTestGNBContext()
	ue := createTestUE(gnb, 12345)
	ue.CreateUeContext("not informed", "", []string{"01"}, []string{"010203"}, nil)

	HandlerPduSessionResourceSetupRequest(gnb, pduSessionResourceSetupRequest(t, ue, []byte{0x01}, []byte{0x01, 0x02, 0x03}))

	pduSession, err := ue.GetPduSession(1)
	require.NoError(t, err)
	require.NotNil(t, pduSession, "the PDU session should have been created")
	assert.Equal(t, context.Ready, ue.GetState(), "the Setup Response should have been built")
	assert.Len(t, ue.GetGnbTx(), 1, "the session should have been handed to the UE")
}

// A transfer without an UL NG-U tunnel leaves the UPF address empty. That used to be indexed
// unconditionally, and the panic ended the process; the session must be skipped instead.
func TestHandlerPduSessionResourceSetupRequest_NoUlTunnel(t *testing.T) {
	gnb := createTestGNBContext()
	ue := createTestUE(gnb, 12345)
	ue.CreateUeContext("not informed", "", []string{"01"}, []string{"010203"}, nil)

	HandlerPduSessionResourceSetupRequest(gnb, pduSessionResourceSetupRequestWith(t, ue, []byte{0x01}, []byte{0x01, 0x02, 0x03}, false))

	pduSession, err := ue.GetPduSession(1)
	require.NoError(t, err)
	assert.Nil(t, pduSession, "a session without an UL tunnel should be skipped")
	assert.NotEqual(t, context.Ready, ue.GetState(), "no Setup Response should have been built")
	assert.Empty(t, ue.GetGnbTx(), "nothing should have been sent to the UE")
}

// NAS-PDU is optional in a PDU Session Resource Release Command. An IE that is present but
// empty used to be dereferenced after being logged, and the panic ended the process; it
// must be treated as absent.
func TestHandlerPduSessionReleaseCommand_EmptyNasPdu(t *testing.T) {
	gnb := createTestGNBContext()
	ue := createTestUE(gnb, 12345)

	message := &ngapType.NGAPPDU{
		Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentPDUSessionResourceReleaseCommand,
				PDUSessionResourceReleaseCommand: &ngapType.PDUSessionResourceReleaseCommand{
					ProtocolIEs: ngapType.ProtocolIEContainerPDUSessionResourceReleaseCommandIEs{
						List: []ngapType.PDUSessionResourceReleaseCommandIEs{
							{
								Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDAMFUENGAPID},
								Value: ngapType.PDUSessionResourceReleaseCommandIEsValue{
									Present:     ngapType.PDUSessionResourceReleaseCommandIEsPresentAMFUENGAPID,
									AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 67890},
								},
							},
							{
								Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDRANUENGAPID},
								Value: ngapType.PDUSessionResourceReleaseCommandIEsValue{
									Present:     ngapType.PDUSessionResourceReleaseCommandIEsPresentRANUENGAPID,
									RANUENGAPID: &ngapType.RANUENGAPID{Value: ue.GetRanUeId()},
								},
							},
							{
								Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDNASPDU},
								Value: ngapType.PDUSessionResourceReleaseCommandIEsValue{
									Present: ngapType.PDUSessionResourceReleaseCommandIEsPresentNASPDU,
								},
							},
							{
								Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDPDUSessionResourceToReleaseListRelCmd},
								Value: ngapType.PDUSessionResourceReleaseCommandIEsValue{
									Present:                               ngapType.PDUSessionResourceReleaseCommandIEsPresentPDUSessionResourceToReleaseListRelCmd,
									PDUSessionResourceToReleaseListRelCmd: &ngapType.PDUSessionResourceToReleaseListRelCmd{},
								},
							},
						},
					},
				},
			},
		},
	}

	assert.NotPanics(t, func() { HandlerPduSessionReleaseCommand(gnb, message) })
}
