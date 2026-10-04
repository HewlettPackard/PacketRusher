/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */
package ngap

import (
	"encoding"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"testing"
	"time"

	ngapType "github.com/free5gc/ngap/ie"
	ngapmsg "github.com/free5gc/ngap/message"
	"github.com/stretchr/testify/require"
)

// dispatch hands a message to the gNB as the reader of the AMF's association does.
func dispatch(t *testing.T, gnb *context.GNBContext, message encoding.BinaryMarshaler) {
	t.Helper()
	payload, err := message.MarshalBinary()
	require.NoError(t, err)
	Dispatch(nil, gnb, payload)
}

func downlinkNAS(ue *context.GNBUe, amfUeId int64, nas byte) *ngapmsg.DownlinkNASTransport {
	return &ngapmsg.DownlinkNASTransport{
		AMFUENGAPID: &ngapType.AMFUENGAPID{Value: amfUeId},
		RANUENGAPID: &ngapType.RANUENGAPID{Value: ue.GetRanUeId()},
		NASPDU:      &ngapType.NASPDU{Value: []byte{nas}},
	}
}

func receive(t *testing.T, tx <-chan context.UEMessage) (context.UEMessage, bool) {
	t.Helper()
	select {
	case message, open := <-tx:
		return message, open
	case <-time.After(time.Second):
		t.Fatal("the UE's channel should have received a message, or been closed")
		return context.UEMessage{}, false
	}
}

// The messages of a UE are handled in the order they were received (#102): a release
// naming the UE by its AMF UE NGAP ID alone, an ID no handler has seen yet, does not
// overtake the NAS messages before it. Meanwhile that UE, slow to take its messages,
// holds back neither the reader nor another UE.
func TestDispatchHandlesTheMessagesOfAUEInOrder(t *testing.T) {
	gnb := createTestGNBContext()
	slowTx, otherTx := make(chan context.UEMessage), make(chan context.UEMessage, 1)
	slow, err := gnb.NewGnBUe(slowTx, nil, 1, nil)
	require.NoError(t, err)
	other, err := gnb.NewGnBUe(otherTx, nil, 2, nil)
	require.NoError(t, err)

	dispatch(t, gnb, downlinkNAS(slow, 42, 1))
	dispatch(t, gnb, downlinkNAS(slow, 42, 2))
	dispatch(t, gnb, &ngapmsg.UEContextReleaseCommand{
		UENGAPIDs: &ngapType.UENGAPIDs{Choice: &ngapType.AMFUENGAPID{Value: 42}},
		Cause:     &ngapType.Cause{Choice: &ngapType.CauseNas{Value: ngapType.CauseNasPresentNormalRelease}},
	})
	dispatch(t, gnb, downlinkNAS(other, 43, 3))

	message, _ := receive(t, otherTx)
	require.Equal(t, []byte{3}, message.Nas)
	for _, nas := range []byte{1, 2} {
		message, open := receive(t, slowTx)
		require.True(t, open, "the release overtook NAS message %d", nas)
		require.Equal(t, []byte{nas}, message.Nas)
	}
	_, open := receive(t, slowTx)
	require.False(t, open, "the release should close the UE's channel, after its NAS messages")
}
