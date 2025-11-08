package ngap
/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package ngap

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"net/netip"
	"testing"

	"github.com/free5gc/ngap/ngapType"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestGNBContext() *context.GNBContext {
	gnb := &context.GNBContext{}
	gnb.NewRanGnbContext("test-gnb", "001", "01", "000001", "1", "000001",
		netip.MustParseAddrPort("127.0.0.1:9999"),
		netip.MustParseAddrPort("127.0.0.1:2152"))
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
	
	// Test with UE in Down state
	retrievedUe := getUeFromContext(gnb, ranUeId, amfUeId)
	assert.Nil(t, retrievedUe, "Should return nil for UE in Down state")
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
