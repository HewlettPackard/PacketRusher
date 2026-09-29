//go:build linux

/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package context

import (
	"net/netip"
	"syscall"
	"testing"

	"github.com/ishidawataru/sctp"
	"github.com/stretchr/testify/require"
)

func newAssociationTestGnb() *GNBContext {
	gnb := &GNBContext{}
	gnb.NewRanGnbContext("test-gnb", "001", "01", "000001", "1", "000001",
		netip.MustParseAddrPort("127.0.0.1:9999"),
		netip.MustParseAddrPort("127.0.0.1:2152"))
	return gnb
}

// pairedConn returns an association and the fd of a peer that reads end-of-file once the
// association is closed.
func pairedConn(t *testing.T) (*sctp.SCTPConn, int) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	require.NoError(t, err)
	require.NoError(t, syscall.SetNonblock(fds[1], true))
	t.Cleanup(func() { _ = syscall.Close(fds[1]) })
	return sctp.NewSCTPConn(fds[0], nil), fds[1]
}

func isClosed(peer int) bool {
	buf := make([]byte, 1)
	n, err := syscall.Read(peer, buf)
	return n == 0 && err == nil
}

func TestPublishAssociationStoresConn(t *testing.T) {
	gnb := newAssociationTestGnb()
	amf := gnb.NewGnBAmf(netip.MustParseAddrPort("127.0.0.1:38412"))
	conn, peer := pairedConn(t)

	require.True(t, gnb.PublishAssociation(amf, conn))
	require.Same(t, conn, amf.GetSCTPConn())
	require.Same(t, conn, gnb.GetN2())
	require.False(t, isClosed(peer))
}

// The AMF is removed while its association is being dialled: the dialled one must not be
// published, since nothing that serves the AMF would close it.
func TestPublishAssociationRefusesRemovedAmf(t *testing.T) {
	gnb := newAssociationTestGnb()
	amf := gnb.NewGnBAmf(netip.MustParseAddrPort("127.0.0.1:38412"))
	gnb.DeleteGnBAmf(amf.GetAmfId())
	conn, _ := pairedConn(t)

	require.False(t, gnb.PublishAssociation(amf, conn))
	require.Nil(t, amf.GetSCTPConn())
	require.Nil(t, gnb.GetN2())
}

// Terminate runs while an association is being dialled: the dialled one must not be
// published, since Terminate has already closed what it could see.
func TestPublishAssociationRefusesAfterTerminate(t *testing.T) {
	gnb := newAssociationTestGnb()
	amf := gnb.NewGnBAmf(netip.MustParseAddrPort("127.0.0.1:38412"))
	gnb.Terminate()
	conn, _ := pairedConn(t)

	require.False(t, gnb.PublishAssociation(amf, conn))
	require.Nil(t, amf.GetSCTPConn())
	require.Nil(t, gnb.GetN2())
}

func TestRemoveGnbAmfClosesAssociationAndRefusesPublish(t *testing.T) {
	gnb := newAssociationTestGnb()
	amf := gnb.NewGnBAmf(netip.MustParseAddrPort("127.0.0.1:38412"))
	conn, peer := pairedConn(t)
	require.True(t, gnb.PublishAssociation(amf, conn))

	gnb.RemoveGnbAmf(amf)
	require.True(t, isClosed(peer), "removing the AMF should close its association")
	require.False(t, gnb.HasGnbAmf(amf.GetAmfId()))

	redialled, _ := pairedConn(t)
	require.False(t, gnb.PublishAssociation(amf, redialled))
	require.Same(t, conn, amf.GetSCTPConn())
}

// Only the association published last is the gNB's N2; Terminate must close every AMF's.
func TestTerminateClosesEveryAmfAssociation(t *testing.T) {
	gnb := newAssociationTestGnb()
	first := gnb.NewGnBAmf(netip.MustParseAddrPort("127.0.0.1:38412"))
	second := gnb.NewGnBAmf(netip.MustParseAddrPort("127.0.0.2:38412"))
	firstConn, firstPeer := pairedConn(t)
	secondConn, secondPeer := pairedConn(t)
	require.True(t, gnb.PublishAssociation(first, firstConn))
	require.True(t, gnb.PublishAssociation(second, secondConn))

	gnb.Terminate()
	require.True(t, isClosed(firstPeer), "the first AMF's association should be closed")
	require.True(t, isClosed(secondPeer), "the second AMF's association should be closed")
}
