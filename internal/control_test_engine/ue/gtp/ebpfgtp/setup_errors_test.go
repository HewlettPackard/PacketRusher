//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
	"net/netip"
	"testing"
)

func TestSetupUnavailableRequiresCompleteOwnedRollback(t *testing.T) {
	for _, dirty := range []bool{false, true} {
		t.Run(map[bool]string{false: "clean", true: "retained endpoint"}[dirty], func(t *testing.T) {
			r, _, c := isolatedRegistry()
			endpoint := &fakeCloser{}
			if dirty {
				endpoint.err = unix.EBUSY
			}
			r.attachEndpoint = func(int) (closer, error) { return endpoint, unix.EOPNOTSUPP }
			s, err := r.Open(c)
			require.Nil(t, s)
			require.ErrorIs(t, err, ErrUnavailable)
			require.ErrorIs(t, err, unix.EOPNOTSUPP)
			require.Equal(t, dirty, errors.Is(err, ErrCleanupIncomplete))
			if dirty {
				require.ErrorIs(t, err, unix.EBUSY)
				require.Len(t, r.orphans, 1)
				require.NotEmpty(t, r.locals, "retained endpoint keeps its source lease")
				endpoint.err = nil
				r.reap()
			}
			require.Empty(t, r.orphans)
			require.Empty(t, r.locals)
			require.Empty(t, r.links)
		})
	}
}

func TestIngressCapabilityFailureDoesNotHideSocketCloseError(t *testing.T) {
	r, _, c := isolatedRegistry()
	socket := &fakeCloser{err: unix.EIO}
	r.listen = func(netip.Addr) (closer, error) { return socket, nil }
	r.attach = func(int) (closer, error) { return nil, unix.EPERM }
	_, err := r.Open(c)
	require.ErrorIs(t, err, ErrUnavailable)
	require.ErrorIs(t, err, ErrCleanupIncomplete)
	require.ErrorIs(t, err, unix.EPERM)
	require.ErrorIs(t, err, unix.EIO)
	require.Same(t, socket, r.ports[c.Local])
	socket.err = nil
	r.reap()
	require.Empty(t, r.ports)
}

func TestInvalidConfigAndOwnershipCollisionAreNeverUnavailable(t *testing.T) {
	r, _, c := isolatedRegistry()
	bad := c
	bad.QFI = 64
	_, err := r.Open(bad)
	require.Error(t, err)
	require.NotErrorIs(t, err, ErrUnavailable)
	s, err := r.Open(c)
	require.NoError(t, err)
	defer s.Close()
	_, err = r.Open(c)
	require.Error(t, err)
	require.NotErrorIs(t, err, ErrUnavailable)
	require.NotErrorIs(t, attachmentUnavailable(unix.EEXIST), ErrUnavailable)
}
