/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package ngap

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/ngap/trigger"
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

func errorIndication(ue *context.GNBUe) *ngapType.NGAPPDU {
	return &ngapType.NGAPPDU{
		Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentErrorIndication,
				ErrorIndication: &ngapType.ErrorIndication{
					ProtocolIEs: ngapType.ProtocolIEContainerErrorIndicationIEs{
						List: []ngapType.ErrorIndicationIEs{
							{
								Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDRANUENGAPID},
								Value: ngapType.ErrorIndicationIEsValue{
									Present:     ngapType.ErrorIndicationIEsPresentRANUENGAPID,
									RANUENGAPID: &ngapType.RANUENGAPID{Value: ue.GetRanUeId()},
								},
							},
						},
					},
				},
			},
		},
	}
}

// With a release request pending, an Error Indication for the UE is its answer and the
// context is released locally.
func TestHandlerErrorIndication_PendingReleaseDeletesUe(t *testing.T) {
	gnb := createTestGNBContext()
	ue := createTestUE(gnb, 12345)
	ue.SetReleaseRequested(true)

	HandlerErrorIndication(gnb, errorIndication(ue))

	_, err := gnb.GetGnbUe(ue.GetRanUeId())
	assert.Error(t, err, "the UE context should have been released locally")
}

// A release request that fails to send is not pending: a later Error Indication must not be
// taken as its answer and delete the UE. The test UE has no SCTP association, so the send fails.
func TestSendUeContextReleaseRequest_FailedSendIsNotPending(t *testing.T) {
	gnb := createTestGNBContext()
	ue := createTestUE(gnb, 12345)
	require.Nil(t, ue.GetSCTP())

	trigger.SendUeContextReleaseRequest(ue)
	assert.False(t, ue.GetReleaseRequested(), "a request that was not sent must not be pending")

	HandlerErrorIndication(gnb, errorIndication(ue))

	retrievedUe, err := gnb.GetGnbUe(ue.GetRanUeId())
	require.NoError(t, err, "the UE context must survive the Error Indication")
	assert.Equal(t, ue, retrievedUe)
}

// A failed send must not cancel an earlier request that was sent and is still pending.
func TestSendUeContextReleaseRequest_FailedSendKeepsEarlierRequest(t *testing.T) {
	gnb := createTestGNBContext()
	ue := createTestUE(gnb, 12345)
	ue.SetReleaseRequested(true)

	trigger.SendUeContextReleaseRequest(ue)

	assert.True(t, ue.GetReleaseRequested(), "the earlier request should still be pending")
}

// PDU SESSION RESOURCE SETUP LIST SU REQ is mandatory. An absent one left the list nil, and
// ranging over it panicked.
func TestHandlerPduSessionResourceSetupRequest_AbsentList(t *testing.T) {
	gnb := createTestGNBContext()
	ue := createTestUE(gnb, 12345)
	ue.CreateUeContext("not informed", "", []string{"01"}, []string{"010203"}, nil)

	message := pduSessionResourceSetupRequest(t, ue, []byte{0x01}, []byte{0x01, 0x02, 0x03})
	ies := &message.InitiatingMessage.Value.PDUSessionResourceSetupRequest.ProtocolIEs
	ies.List = ies.List[:2]

	HandlerPduSessionResourceSetupRequest(gnb, message)

	assert.NotEqual(t, context.Ready, ue.GetState(), "no Setup Response should have been built")
	assert.Empty(t, ue.GetGnbTx(), "nothing should have been sent to the UE")
}

// UE Paging Identity is mandatory. An absent one was dereferenced and panicked.
func TestHandlerPaging_AbsentIdentity(t *testing.T) {
	gnb := createTestGNBContext()

	message := &ngapType.NGAPPDU{
		Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentPaging,
				Paging:  &ngapType.Paging{},
			},
		},
	}

	HandlerPaging(gnb, message)

	assert.Empty(t, gnb.GetPagedUEs(), "nothing should have been paged")
}

// UE NGAP IDs is a CHOICE, and the AMF may identify the UE by its AMF UE NGAP ID alone. That
// left the ID pair nil, and dereferencing it panicked.
func TestHandlerUeContextReleaseCommand_AmfUeIdOnly(t *testing.T) {
	gnb := createTestGNBContext()
	ue := createTestUE(gnb, 12345)
	ue.SetAmfUeId(67890)

	message := &ngapType.NGAPPDU{
		Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentUEContextReleaseCommand,
				UEContextReleaseCommand: &ngapType.UEContextReleaseCommand{
					ProtocolIEs: ngapType.ProtocolIEContainerUEContextReleaseCommandIEs{
						List: []ngapType.UEContextReleaseCommandIEs{
							{
								Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDUENGAPIDs},
								Value: ngapType.UEContextReleaseCommandIEsValue{
									Present: ngapType.UEContextReleaseCommandIEsPresentUENGAPIDs,
									UENGAPIDs: &ngapType.UENGAPIDs{
										Present:     ngapType.UENGAPIDsPresentAMFUENGAPID,
										AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 67890},
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

	_, err := gnb.GetGnbUe(ue.GetRanUeId())
	assert.Error(t, err, "the UE identified by its AMF UE NGAP ID should have been released")
}

// A transfer without an UL NG-U tunnel in an Initial Context Setup Request left the uplink
// TEID empty, and reading it panicked. The session must be skipped.
func TestHandlerInitialContextSetupRequest_NoUlTunnel(t *testing.T) {
	gnb := createTestGNBContext()
	ue := createTestUE(gnb, 12345)

	// Reuse the transfer of a setup request built without an UL tunnel.
	setup := pduSessionResourceSetupRequestWith(t, ue, []byte{0x01}, []byte{0x01, 0x02, 0x03}, false)
	item := setup.InitiatingMessage.Value.PDUSessionResourceSetupRequest.ProtocolIEs.List[2].Value.PDUSessionResourceSetupListSUReq.List[0]

	message := &ngapType.NGAPPDU{
		Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentInitialContextSetupRequest,
				InitialContextSetupRequest: &ngapType.InitialContextSetupRequest{
					ProtocolIEs: ngapType.ProtocolIEContainerInitialContextSetupRequestIEs{
						List: []ngapType.InitialContextSetupRequestIEs{
							{
								Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDAMFUENGAPID},
								Value: ngapType.InitialContextSetupRequestIEsValue{
									Present:     ngapType.InitialContextSetupRequestIEsPresentAMFUENGAPID,
									AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 67890},
								},
							},
							{
								Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDRANUENGAPID},
								Value: ngapType.InitialContextSetupRequestIEsValue{
									Present:     ngapType.InitialContextSetupRequestIEsPresentRANUENGAPID,
									RANUENGAPID: &ngapType.RANUENGAPID{Value: ue.GetRanUeId()},
								},
							},
							{
								Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDPDUSessionResourceSetupListCxtReq},
								Value: ngapType.InitialContextSetupRequestIEsValue{
									Present: ngapType.InitialContextSetupRequestIEsPresentPDUSessionResourceSetupListCxtReq,
									PDUSessionResourceSetupListCxtReq: &ngapType.PDUSessionResourceSetupListCxtReq{
										List: []ngapType.PDUSessionResourceSetupItemCxtReq{
											{
												PDUSessionID:                           item.PDUSessionID,
												SNSSAI:                                 item.SNSSAI,
												PDUSessionResourceSetupRequestTransfer: item.PDUSessionResourceSetupRequestTransfer,
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

	HandlerInitialContextSetupRequest(gnb, message)

	pduSession, err := ue.GetPduSession(1)
	require.NoError(t, err)
	assert.Nil(t, pduSession, "a session without an UL tunnel should be skipped")
}

func handoverRequest(t *testing.T, gnb *context.GNBContext, prUeId int64, withList bool) *ngapType.NGAPPDU {
	t.Helper()

	// The handler reads only the IndexToRFSP, which carries the simulator's UE id; the
	// rest is the minimum the encoder accepts.
	cell := ngapType.NGRANCGI{
		Present: ngapType.NGRANCGIPresentNRCGI,
		NRCGI: &ngapType.NRCGI{
			PLMNIdentity:   ngapType.PLMNIdentity{Value: aper.OctetString("\x00\xf1\x10")},
			NRCellIdentity: ngapType.NRCellIdentity{Value: aper.BitString{Bytes: []byte{0, 0, 0, 0x10, 0}, BitLength: 36}},
		},
	}
	container := ngapType.SourceNGRANNodeToTargetNGRANNodeTransparentContainer{
		RRCContainer: ngapType.RRCContainer{Value: aper.OctetString("\x00\x00\x11")},
		IndexToRFSP:  &ngapType.IndexToRFSP{Value: prUeId},
		TargetCellID: cell,
		UEHistoryInformation: ngapType.UEHistoryInformation{
			List: []ngapType.LastVisitedCellItem{
				{
					LastVisitedCellInformation: ngapType.LastVisitedCellInformation{
						Present:   ngapType.LastVisitedCellInformationPresentNGRANCell,
						NGRANCell: &ngapType.LastVisitedNGRANCellInformation{GlobalCellID: cell},
					},
				},
			},
		},
	}
	encodedContainer, err := aper.MarshalWithParams(container, "valueExt")
	require.NoError(t, err)

	ies := []ngapType.HandoverRequestIEs{
		{
			Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDAMFUENGAPID},
			Value: ngapType.HandoverRequestIEsValue{
				Present:     ngapType.HandoverRequestIEsPresentAMFUENGAPID,
				AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 67890},
			},
		},
		{
			Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDSourceToTargetTransparentContainer},
			Value: ngapType.HandoverRequestIEsValue{
				Present: ngapType.HandoverRequestIEsPresentSourceToTargetTransparentContainer,
				SourceToTargetTransparentContainer: &ngapType.SourceToTargetTransparentContainer{
					Value: encodedContainer,
				},
			},
		},
	}
	if withList {
		// Reuse the transfer of a setup request built without an UL tunnel.
		ue := createTestUE(gnb, prUeId+1)
		setup := pduSessionResourceSetupRequestWith(t, ue, []byte{0x01}, []byte{0x01, 0x02, 0x03}, false)
		item := setup.InitiatingMessage.Value.PDUSessionResourceSetupRequest.ProtocolIEs.List[2].Value.PDUSessionResourceSetupListSUReq.List[0]
		ies = append(ies, ngapType.HandoverRequestIEs{
			Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDPDUSessionResourceSetupListHOReq},
			Value: ngapType.HandoverRequestIEsValue{
				Present: ngapType.HandoverRequestIEsPresentPDUSessionResourceSetupListHOReq,
				PDUSessionResourceSetupListHOReq: &ngapType.PDUSessionResourceSetupListHOReq{
					List: []ngapType.PDUSessionResourceSetupItemHOReq{
						{
							PDUSessionID:            item.PDUSessionID,
							SNSSAI:                  item.SNSSAI,
							HandoverRequestTransfer: item.PDUSessionResourceSetupRequestTransfer,
						},
					},
				},
			},
		})
	}

	return &ngapType.NGAPPDU{
		Present: ngapType.NGAPPDUPresentInitiatingMessage,
		InitiatingMessage: &ngapType.InitiatingMessage{
			Value: ngapType.InitiatingMessageValue{
				Present: ngapType.InitiatingMessagePresentHandoverRequest,
				HandoverRequest: &ngapType.HandoverRequest{
					ProtocolIEs: ngapType.ProtocolIEContainerHandoverRequestIEs{List: ies},
				},
			},
		},
	}
}

// PDU Session Resource Setup List HO Req is mandatory. An absent one left the list nil, and
// ranging over it panicked after a UE had already been created for the handover.
func TestHandlerHandoverRequest_AbsentList(t *testing.T) {
	gnb := createTestGNBContext()

	HandlerHandoverRequest(nil, gnb, handoverRequest(t, gnb, 777, false))

	_, err := gnb.GetGnbUeByPrUeId(777)
	assert.Error(t, err, "a rejected Handover Request should leave no UE behind")
}

// A transfer without an UL NG-U tunnel in a Handover Request left the uplink TEID empty, and
// reading it panicked. The session must be skipped.
func TestHandlerHandoverRequest_NoUlTunnel(t *testing.T) {
	gnb := createTestGNBContext()

	HandlerHandoverRequest(nil, gnb, handoverRequest(t, gnb, 777, true))

	ue, err := gnb.GetGnbUeByPrUeId(777)
	require.NoError(t, err)
	pduSession, err := ue.GetPduSession(1)
	require.NoError(t, err)
	assert.Nil(t, pduSession, "a session without an UL tunnel should be skipped")
}

// QoS Characteristics is a CHOICE. A flow with a dynamic 5QI left NonDynamic5QI nil, and
// reading its 5QI panicked.
func TestHandlerPduSessionResourceSetupRequest_Dynamic5QI(t *testing.T) {
	gnb := createTestGNBContext()
	ue := createTestUE(gnb, 12345)
	ue.CreateUeContext("not informed", "", []string{"01"}, []string{"010203"}, nil)

	message := pduSessionResourceSetupRequest(t, ue, []byte{0x01}, []byte{0x01, 0x02, 0x03})
	item := &message.InitiatingMessage.Value.PDUSessionResourceSetupRequest.ProtocolIEs.List[2].Value.PDUSessionResourceSetupListSUReq.List[0]
	transfer := ngapType.PDUSessionResourceSetupRequestTransfer{}
	require.NoError(t, aper.UnmarshalWithParams(item.PDUSessionResourceSetupRequestTransfer, &transfer, "valueExt"))
	transfer.ProtocolIEs.List[2].Value.QosFlowSetupRequestList.List[0].QosFlowLevelQosParameters.QosCharacteristics = ngapType.QosCharacteristics{
		Present: ngapType.QosCharacteristicsPresentDynamic5QI,
		Dynamic5QI: &ngapType.Dynamic5QIDescriptor{
			PriorityLevelQos:  ngapType.PriorityLevelQos{Value: 1},
			PacketDelayBudget: ngapType.PacketDelayBudget{Value: 100},
			PacketErrorRate:   ngapType.PacketErrorRate{PERScalar: 1, PERExponent: 6},
			FiveQI:            &ngapType.FiveQI{Value: 7},
		},
	}
	encoded, err := aper.MarshalWithParams(transfer, "valueExt")
	require.NoError(t, err)
	item.PDUSessionResourceSetupRequestTransfer = encoded

	HandlerPduSessionResourceSetupRequest(gnb, message)

	pduSession, err := ue.GetPduSession(1)
	require.NoError(t, err)
	require.NotNil(t, pduSession, "a session with a dynamic 5QI should be set up")
}

// An empty PDU session slot is (nil, nil). A Path Switch Request Acknowledge naming a session
// the UE does not have, with an UL tunnel, then wrote through the nil session and panicked.
func TestHandlerPathSwitchRequestAcknowledge_UnknownSession(t *testing.T) {
	gnb := createTestGNBContext()
	ue := createTestUE(gnb, 12345)

	transfer := ngapType.PathSwitchRequestAcknowledgeTransfer{
		ULNGUUPTNLInformation: &ngapType.UPTransportLayerInformation{
			Present: ngapType.UPTransportLayerInformationPresentGTPTunnel,
			GTPTunnel: &ngapType.GTPTunnel{
				TransportLayerAddress: ngapType.TransportLayerAddress{
					Value: aper.BitString{Bytes: []byte{10, 0, 0, 1}, BitLength: 32},
				},
				GTPTEID: ngapType.GTPTEID{Value: []byte{0, 0, 0, 1}},
			},
		},
	}
	encoded, err := aper.MarshalWithParams(transfer, "valueExt")
	require.NoError(t, err)

	message := &ngapType.NGAPPDU{
		Present: ngapType.NGAPPDUPresentSuccessfulOutcome,
		SuccessfulOutcome: &ngapType.SuccessfulOutcome{
			Value: ngapType.SuccessfulOutcomeValue{
				Present: ngapType.SuccessfulOutcomePresentPathSwitchRequestAcknowledge,
				PathSwitchRequestAcknowledge: &ngapType.PathSwitchRequestAcknowledge{
					ProtocolIEs: ngapType.ProtocolIEContainerPathSwitchRequestAcknowledgeIEs{
						List: []ngapType.PathSwitchRequestAcknowledgeIEs{
							{
								Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDRANUENGAPID},
								Value: ngapType.PathSwitchRequestAcknowledgeIEsValue{
									Present:     ngapType.PathSwitchRequestAcknowledgeIEsPresentRANUENGAPID,
									RANUENGAPID: &ngapType.RANUENGAPID{Value: ue.GetRanUeId()},
								},
							},
							{
								Id: ngapType.ProtocolIEID{Value: ngapType.ProtocolIEIDPDUSessionResourceSwitchedList},
								Value: ngapType.PathSwitchRequestAcknowledgeIEsValue{
									Present: ngapType.PathSwitchRequestAcknowledgeIEsPresentPDUSessionResourceSwitchedList,
									PDUSessionResourceSwitchedList: &ngapType.PDUSessionResourceSwitchedList{
										List: []ngapType.PDUSessionResourceSwitchedItem{
											{
												PDUSessionID:                         ngapType.PDUSessionID{Value: 5},
												PathSwitchRequestAcknowledgeTransfer: encoded,
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

	HandlerPathSwitchRequestAcknowledge(gnb, message)

	assert.Empty(t, ue.GetGnbTx(), "nothing should have been sent to the UE for an unknown session")
}
