// SPDX-License-Identifier: Apache-2.0
package service

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

type allocationFixture struct {
	mu        sync.Mutex
	addresses map[string]bool
	rules     map[string]bool
	router    bool
	claim     bool
	forwarded netip.Addr
	revoked   int
	deleteErr error
	ops       ipv6NetworkOperations
}

func newAllocationFixture() *allocationFixture {
	f := &allocationFixture{addresses: make(map[string]bool), rules: make(map[string]bool)}
	addAddress := func(_ netlink.Link, address *netlink.Addr) error {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.addresses[address.IP.String()] = true
		return nil
	}
	f.ops = ipv6NetworkOperations{
		addressAdd: addAddress, addressReplace: addAddress,
		addressDel: func(_ netlink.Link, address *netlink.Addr) error {
			f.mu.Lock()
			defer f.mu.Unlock()
			if !address.IP.IsLinkLocalUnicast() && f.deleteErr != nil {
				return f.deleteErr
			}
			delete(f.addresses, address.IP.String())
			return nil
		},
		ruleAdd: func(rule *netlink.Rule) error {
			f.mu.Lock()
			defer f.mu.Unlock()
			f.rules[rule.Src.String()] = true
			return nil
		},
		ruleDel: func(rule *netlink.Rule) error {
			f.mu.Lock()
			defer f.mu.Unlock()
			delete(f.rules, rule.Src.String())
			return nil
		},
		routeAdd: func(route *netlink.Route) error {
			f.mu.Lock()
			defer f.mu.Unlock()
			if route.Type == unix.RTN_BLACKHOLE {
				f.claim = true
			} else {
				f.router = true
			}
			return nil
		},
		routeDel: func(route *netlink.Route) error {
			f.mu.Lock()
			defer f.mu.Unlock()
			if route.Type == unix.RTN_BLACKHOLE {
				f.claim = false
			} else {
				f.router = false
			}
			return nil
		},
	}
	return f
}

func (f *allocationFixture) update(address netip.Addr) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.forwarded = address
	if !address.IsValid() {
		f.revoked++
	}
	return nil
}

func allocationAdvertisement(t *testing.T, prefix uint16, valid uint32, router uint16) []byte {
	t.Helper()
	packet, err := hex.DecodeString(raWire)
	require.NoError(t, err)
	binary.BigEndian.PutUint16(packet[76:78], prefix)
	binary.BigEndian.PutUint32(packet[60:64], valid)
	binary.BigEndian.PutUint32(packet[64:68], valid)
	binary.BigEndian.PutUint16(packet[46:48], router)
	packet[42], packet[43] = 0, 0
	binary.BigEndian.PutUint16(packet[42:44], udp6Checksum(packet))
	return packet
}

func (f *allocationFixture) inspect(assertions func()) {
	f.mu.Lock()
	defer f.mu.Unlock()
	assertions()
}

func TestIPv6AllocationCallbackStagesBeforePublicationAndRevokesPartialFailure(t *testing.T) {
	ue, session := ipv6TestSession(t)
	f := newAllocationFixture()
	link := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Index: 77, MTU: 1456}}
	advertisements := make(chan []byte, 1)
	advertisements <- allocationAdvertisement(t, 0x1234, 30, 30)
	injected := errors.New("forwarding map update failed")
	var staged bool
	reject := func(address netip.Addr) error {
		if !address.IsValid() {
			return nil
		}
		f.inspect(func() {
			staged = f.addresses[address.String()] && f.router && len(f.rules) == 1
		})
		require.False(t, session.GetIPv6().IsValid(), "an unaccepted allocation must not be published")
		return injected
	}
	cleanup, err := setupIPv6Session(ue, session, link, 1000, func([]byte) error { return nil }, advertisements, f.ops, time.Millisecond, f.update, reject)
	require.ErrorIs(t, err, injected)
	require.True(t, staged, "forwarding notification follows address and routing staging")
	require.False(t, session.GetIPv6().IsValid())
	require.NotNil(t, cleanup)
	require.NoError(t, cleanup())
	f.mu.Lock()
	defer f.mu.Unlock()
	require.False(t, f.forwarded.IsValid(), "the first owner must be revoked when a later owner fails")
	require.Positive(t, f.revoked)
	require.Empty(t, f.addresses)
	require.Empty(t, f.rules)
	require.False(t, f.router)
	require.False(t, f.claim)
}

func TestIPv6AllocationCallbackRejectsRenumberingWithoutPublishingCandidate(t *testing.T) {
	ue, session := ipv6TestSession(t)
	f := newAllocationFixture()
	link := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Index: 77, MTU: 1456}}
	advertisements := make(chan []byte, 2)
	advertisements <- allocationAdvertisement(t, 0x1234, 30, 30)
	rejected := make(chan netip.Addr, 1)
	callback := func(address netip.Addr) error {
		if address == netip.MustParseAddr("2001:db8:5678::7") {
			rejected <- session.GetIPv6()
			return errors.New("reject new prefix")
		}
		return f.update(address)
	}
	cleanup, err := setupIPv6Session(ue, session, link, 1000, func([]byte) error { return nil }, advertisements, f.ops, time.Millisecond, callback)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, cleanup()) })
	advertisements <- allocationAdvertisement(t, 0x5678, 30, 30)
	select {
	case old := <-rejected:
		require.Equal(t, netip.MustParseAddr("2001:db8:1234::7"), old)
	case <-time.After(time.Second):
		t.Fatal("candidate did not reach the forwarding owner")
	}
	require.Eventually(t, func() bool { return !session.GetIPv6().IsValid() }, time.Second, time.Millisecond)
	f.inspect(func() {
		require.False(t, f.forwarded.IsValid())
		require.Empty(t, f.rules)
		require.False(t, f.addresses["2001:db8:1234::7"])
		require.False(t, f.addresses["2001:db8:5678::7"])
	})
}

func TestIPv6AllocationExpiryRevokesDespiteDeletionFailureAndRetainsPolicy(t *testing.T) {
	ue, session := ipv6TestSession(t)
	f := newAllocationFixture()
	link := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Index: 77, MTU: 1456}}
	advertisements := make(chan []byte, 1)
	advertisements <- allocationAdvertisement(t, 0x1234, 1, 5)
	cleanup, err := setupIPv6Session(ue, session, link, 1000, func([]byte) error { return nil }, advertisements, f.ops, time.Millisecond, f.update)
	require.NoError(t, err)
	injected := errors.New("address deletion failed")
	f.inspect(func() {
		f.deleteErr = injected
	})
	t.Cleanup(func() {
		f.inspect(func() {
			f.deleteErr = nil
		})
		require.NoError(t, cleanup())
	})
	require.Eventually(t, func() bool { return !session.GetIPv6().IsValid() }, 2*time.Second, time.Millisecond)
	f.inspect(func() {
		require.False(t, f.forwarded.IsValid())
		require.True(t, f.addresses["2001:db8:1234::7"])
		require.True(t, f.rules["2001:db8:1234::7/128"], "failed address deletion must retain its source claim")
		require.True(t, f.claim)
	})
	require.ErrorIs(t, cleanup(), injected)
	f.inspect(func() {
		f.deleteErr = nil
	})
	require.NoError(t, cleanup())
	f.inspect(func() {
		require.Empty(t, f.addresses)
		require.Empty(t, f.rules)
		require.False(t, f.claim)
	})
}

func TestIPv6RouterOnlyExpiryPreservesForwardingAllocation(t *testing.T) {
	ue, session := ipv6TestSession(t)
	f := newAllocationFixture()
	link := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Index: 77, MTU: 1456}}
	advertisements := make(chan []byte, 1)
	advertisements <- allocationAdvertisement(t, 0x1234, 30, 1)
	cleanup, err := setupIPv6Session(ue, session, link, 1000, func([]byte) error { return nil }, advertisements, f.ops, time.Millisecond, f.update)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, cleanup()) })
	require.Eventually(t, func() bool { f.mu.Lock(); defer f.mu.Unlock(); return !f.router }, 2*time.Second, time.Millisecond)
	require.Equal(t, netip.MustParseAddr("2001:db8:1234::7"), session.GetIPv6())
	f.inspect(func() {
		require.Equal(t, session.GetIPv6(), f.forwarded)
		require.Zero(t, f.revoked)
		require.True(t, f.claim)
		require.True(t, f.rules["2001:db8:1234::7/128"])
	})
}

func TestIPv6AllocationRevocationErrorClearsContextAndRemainsRetryable(t *testing.T) {
	ue, session := ipv6TestSession(t)
	f := newAllocationFixture()
	link := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Index: 77, MTU: 1456}}
	advertisements := make(chan []byte, 1)
	advertisements <- allocationAdvertisement(t, 0x1234, 30, 30)
	injected := errors.New("forwarding revocation failed after deactivation")
	var errorMu sync.Mutex
	failRevocation := true
	callback := func(address netip.Addr) error {
		if err := f.update(address); err != nil {
			return err
		}
		errorMu.Lock()
		defer errorMu.Unlock()
		if !address.IsValid() && failRevocation {
			return injected
		}
		return nil
	}
	cleanup, err := setupIPv6Session(ue, session, link, 1000, func([]byte) error { return nil }, advertisements, f.ops, time.Millisecond, callback)
	require.NoError(t, err)
	t.Cleanup(func() {
		errorMu.Lock()
		failRevocation = false
		errorMu.Unlock()
		require.NoError(t, cleanup())
	})
	require.ErrorIs(t, cleanup(), injected)
	require.False(t, session.GetIPv6().IsValid())
	f.inspect(func() {
		require.False(t, f.forwarded.IsValid())
		require.Empty(t, f.addresses)
		require.Empty(t, f.rules)
		require.True(t, f.claim, "retain the table claim until forwarding retirement succeeds")
	})
	errorMu.Lock()
	failRevocation = false
	errorMu.Unlock()
	require.NoError(t, cleanup())
	f.inspect(func() {
		require.False(t, f.claim)
	})
}

func TestIPv6ClosedControlTransportRevokesAndCleanupJoinsNotification(t *testing.T) {
	ue, session := ipv6TestSession(t)
	f := newAllocationFixture()
	link := &netlink.Tuntap{LinkAttrs: netlink.LinkAttrs{Index: 77, MTU: 1456}}
	advertisements := make(chan []byte, 1)
	advertisements <- allocationAdvertisement(t, 0x1234, 30, 30)
	revoking, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	callback := func(address netip.Addr) error {
		if !address.IsValid() {
			once.Do(func() { close(revoking); <-release })
		}
		return f.update(address)
	}
	cleanup, err := setupIPv6Session(ue, session, link, 1000, func([]byte) error { return nil }, advertisements, f.ops, time.Millisecond, callback)
	require.NoError(t, err)
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }); require.NoError(t, cleanup()) })
	close(advertisements)
	select {
	case <-revoking:
	case <-time.After(time.Second):
		t.Fatal("closed control transport retained its allocation")
	}
	done := make(chan error, 1)
	go func() { done <- cleanup() }()
	select {
	case err := <-done:
		t.Fatalf("cleanup returned before its forwarding notification: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	releaseOnce.Do(func() { close(release) })
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("cleanup did not join the forwarding worker")
	}
	require.False(t, session.GetIPv6().IsValid())
	f.inspect(func() {
		require.False(t, f.forwarded.IsValid())
		require.Empty(t, f.addresses)
		require.Empty(t, f.rules)
	})
}
