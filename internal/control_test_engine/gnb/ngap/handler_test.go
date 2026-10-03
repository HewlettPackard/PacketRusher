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

	"github.com/free5gc/ngap/aper"
	ngapType "github.com/free5gc/ngap/ie"
	ngapmsg "github.com/free5gc/ngap/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ngapConvert "my5G-RANTester/lib/ngap"
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
	message := &ngapmsg.UEContextReleaseCommand{
		UENGAPIDs: &ngapType.UENGAPIDs{
			Choice: &ngapType.UENGAPIDPair{
				RANUENGAPID: &ngapType.RANUENGAPID{
					Value: ranUeId,
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
	message := &ngapmsg.UEContextReleaseCommand{
		UENGAPIDs: &ngapType.UENGAPIDs{
			Choice: &ngapType.UENGAPIDPair{
				RANUENGAPID: &ngapType.RANUENGAPID{
					Value: nonExistentRanUeId,
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
	message := &ngapmsg.UEContextReleaseCommand{}

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
			message := &ngapmsg.UEContextReleaseCommand{
				UENGAPIDs: &ngapType.UENGAPIDs{
					Choice: &ngapType.UENGAPIDPair{
						RANUENGAPID: &ngapType.RANUENGAPID{
							Value: ranUeId,
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
func pduSessionResourceSetupRequest(t *testing.T, ue *context.GNBUe, sst, sd []byte) *ngapmsg.PDUSessionResourceSetupRequest {
	return pduSessionResourceSetupRequestWith(t, ue, sst, sd, true)
}

func pduSessionResourceSetupRequestWith(t *testing.T, ue *context.GNBUe, sst, sd []byte, withUlTunnel bool) *ngapmsg.PDUSessionResourceSetupRequest {
	t.Helper()
	transfer := &ngapType.PDUSessionResourceSetupRequestTransfer{
		ProtocolIEs: &ngapType.ProtocolIEContainerPDUSessionResourceSetupRequestTransferIEs{List: []ngapType.PDUSessionResourceSetupRequestTransferIEs{
			{ULNGUUPTNLInformation: ngapConvert.UPTransport(netip.MustParseAddr("10.0.0.1"), 1)},
			{PDUSessionType: &ngapType.PDUSessionType{Value: ngapType.PDUSessionTypePresentIpv4}},
			{QosFlowSetupRequestList: &ngapType.QosFlowSetupRequestList{List: []ngapType.QosFlowSetupRequestItem{{
				QosFlowIdentifier: &ngapType.QosFlowIdentifier{Value: 1},
				QosFlowLevelQosParameters: &ngapType.QosFlowLevelQosParameters{
					QosCharacteristics: &ngapType.QosCharacteristics{Choice: &ngapType.NonDynamic5QIDescriptor{FiveQI: &ngapType.FiveQI{Value: 9}}},
					AllocationAndRetentionPriority: &ngapType.AllocationAndRetentionPriority{
						PriorityLevelARP:        &ngapType.PriorityLevelARP{Value: 1},
						PreEmptionCapability:    &ngapType.PreEmptionCapability{Value: 0},
						PreEmptionVulnerability: &ngapType.PreEmptionVulnerability{Value: 0},
					},
				},
			}}}},
		}},
	}
	if !withUlTunnel {
		transfer.ProtocolIEs.List = transfer.ProtocolIEs.List[1:]
	}
	encoded, err := ngapConvert.Marshal(transfer)
	require.NoError(t, err)
	return &ngapmsg.PDUSessionResourceSetupRequest{
		AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 67890}, RANUENGAPID: &ngapType.RANUENGAPID{Value: ue.GetRanUeId()},
		PDUSessionResourceSetupListSUReq: &ngapType.PDUSessionResourceSetupListSUReq{List: []ngapType.PDUSessionResourceSetupItemSUReq{{
			PDUSessionID: &ngapType.PDUSessionID{Value: 1}, SNSSAI: ngapConvert.Slice(sst, sd), PDUSessionResourceSetupRequestTransfer: ptrOctets(encoded),
		}}},
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

	message := &ngapmsg.PDUSessionResourceReleaseCommand{
		AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 67890},

		RANUENGAPID: &ngapType.RANUENGAPID{Value: ue.GetRanUeId()},

		PDUSessionResourceToReleaseListRelCmd: &ngapType.PDUSessionResourceToReleaseListRelCmd{},
	}

	assert.NotPanics(t, func() { HandlerPduSessionReleaseCommand(gnb, message) })
}

func errorIndication(ue *context.GNBUe) *ngapmsg.ErrorIndication {
	return &ngapmsg.ErrorIndication{RANUENGAPID: &ngapType.RANUENGAPID{Value: ue.GetRanUeId()}}
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
	message.PDUSessionResourceSetupListSUReq = nil

	HandlerPduSessionResourceSetupRequest(gnb, message)

	assert.NotEqual(t, context.Ready, ue.GetState(), "no Setup Response should have been built")
	assert.Empty(t, ue.GetGnbTx(), "nothing should have been sent to the UE")
}

// UE Paging Identity is mandatory. An absent one was dereferenced and panicked.
func TestHandlerPaging_AbsentIdentity(t *testing.T) {
	gnb := createTestGNBContext()

	message := &ngapmsg.Paging{}

	HandlerPaging(gnb, message)

	assert.Empty(t, gnb.GetPagedUEs(), "nothing should have been paged")
}

// UE NGAP IDs is a CHOICE, and the AMF may identify the UE by its AMF UE NGAP ID alone. That
// left the ID pair nil, and dereferencing it panicked.
func TestHandlerUeContextReleaseCommand_AmfUeIdOnly(t *testing.T) {
	gnb := createTestGNBContext()
	ue := createTestUE(gnb, 12345)
	ue.SetAmfUeId(67890)

	message := &ngapmsg.UEContextReleaseCommand{
		UENGAPIDs: &ngapType.UENGAPIDs{
			Choice: &ngapType.AMFUENGAPID{Value: 67890},
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
	item := setup.PDUSessionResourceSetupListSUReq.List[0]

	message := &ngapmsg.InitialContextSetupRequest{
		AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 67890},

		RANUENGAPID: &ngapType.RANUENGAPID{Value: ue.GetRanUeId()},

		PDUSessionResourceSetupListCxtReq: &ngapType.PDUSessionResourceSetupListCxtReq{
			List: []ngapType.PDUSessionResourceSetupItemCxtReq{
				{
					PDUSessionID:                           item.PDUSessionID,
					SNSSAI:                                 item.SNSSAI,
					PDUSessionResourceSetupRequestTransfer: item.PDUSessionResourceSetupRequestTransfer,
				},
			},
		},
	}

	HandlerInitialContextSetupRequest(gnb, message)

	pduSession, err := ue.GetPduSession(1)
	require.NoError(t, err)
	assert.Nil(t, pduSession, "a session without an UL tunnel should be skipped")
}

func handoverRequest(t *testing.T, gnb *context.GNBContext, prUeId int64, withList bool) *ngapmsg.HandoverRequest {
	t.Helper()
	cell := &ngapType.NGRANCGI{Choice: &ngapType.NRCGI{
		PLMNIdentity:   &ngapType.PLMNIdentity{Value: []byte{0, 0xf1, 0x10}},
		NRCellIdentity: &ngapType.NRCellIdentity{Value: aper.BitString{Bytes: []byte{0, 0, 0, 0x10, 0}, BitLength: 36}},
	}}
	container := &ngapType.SourceNGRANNodeToTargetNGRANNodeTransparentContainer{
		RRCContainer: &ngapType.RRCContainer{Value: []byte{0, 0, 0x11}}, IndexToRFSP: &ngapType.IndexToRFSP{Value: prUeId}, TargetCellID: cell,
		UEHistoryInformation: &ngapType.UEHistoryInformation{List: []ngapType.LastVisitedCellItem{{
			LastVisitedCellInformation: &ngapType.LastVisitedCellInformation{Choice: &ngapType.LastVisitedNGRANCellInformation{
				GlobalCellID: cell, CellType: &ngapType.CellType{CellSize: &ngapType.CellSize{Value: 0}}, TimeUEStayedInCell: &ngapType.TimeUEStayedInCell{Value: 0},
			}},
		}}},
	}
	encoded, err := ngapConvert.Marshal(container)
	require.NoError(t, err)
	request := &ngapmsg.HandoverRequest{AMFUENGAPID: &ngapType.AMFUENGAPID{Value: 67890}, SourceToTargetTransparentContainer: &ngapType.SourceToTargetTransparentContainer{Value: encoded}}
	if withList {
		ue := createTestUE(gnb, prUeId+1)
		setup := pduSessionResourceSetupRequestWith(t, ue, []byte{1}, []byte{1, 2, 3}, false)
		item := setup.PDUSessionResourceSetupListSUReq.List[0]
		request.PDUSessionResourceSetupListHOReq = &ngapType.PDUSessionResourceSetupListHOReq{List: []ngapType.PDUSessionResourceSetupItemHOReq{{PDUSessionID: item.PDUSessionID, SNSSAI: item.SNSSAI, HandoverRequestTransfer: item.PDUSessionResourceSetupRequestTransfer}}}
	}
	return request
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
	item := &message.PDUSessionResourceSetupListSUReq.List[0]
	transfer := ngapType.PDUSessionResourceSetupRequestTransfer{}
	require.NoError(t, ngapConvert.Unmarshal(*item.PDUSessionResourceSetupRequestTransfer, &transfer))
	transfer.ProtocolIEs.List[2].QosFlowSetupRequestList.List[0].QosFlowLevelQosParameters.QosCharacteristics = &ngapType.QosCharacteristics{
		Choice: &ngapType.Dynamic5QIDescriptor{
			PriorityLevelQos:  &ngapType.PriorityLevelQos{Value: 1},
			PacketDelayBudget: &ngapType.PacketDelayBudget{Value: 100},
			PacketErrorRate:   &ngapType.PacketErrorRate{PERScalar: ptrInt64(1), PERExponent: ptrInt64(6)},
			FiveQI:            &ngapType.FiveQI{Value: 7},
		},
	}
	encoded, err := ngapConvert.Marshal(&transfer)
	require.NoError(t, err)
	item.PDUSessionResourceSetupRequestTransfer = ptrOctets(encoded)

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

			Choice: &ngapType.GTPTunnel{
				TransportLayerAddress: &ngapType.TransportLayerAddress{
					Value: aper.BitString{Bytes: []byte{10, 0, 0, 1}, BitLength: 32},
				},
				GTPTEID: &ngapType.GTPTEID{Value: []byte{0, 0, 0, 1}},
			},
		},
	}
	encoded, err := ngapConvert.Marshal(&transfer)
	require.NoError(t, err)

	message := &ngapmsg.PathSwitchRequestAcknowledge{
		RANUENGAPID: &ngapType.RANUENGAPID{Value: ue.GetRanUeId()},

		PDUSessionResourceSwitchedList: &ngapType.PDUSessionResourceSwitchedList{
			List: []ngapType.PDUSessionResourceSwitchedItem{
				{
					PDUSessionID:                         &ngapType.PDUSessionID{Value: 5},
					PathSwitchRequestAcknowledgeTransfer: ptrOctets(encoded),
				},
			},
		},
	}

	HandlerPathSwitchRequestAcknowledge(gnb, message)

	assert.Empty(t, ue.GetGnbTx(), "nothing should have been sent to the UE for an unknown session")
}

func ptrOctets(value []byte) *aper.OctetString {
	octets := aper.OctetString(value)
	return &octets
}

func ptrInt64(value int64) *int64 { return &value }
