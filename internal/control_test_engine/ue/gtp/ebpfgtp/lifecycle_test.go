//go:build linux

// SPDX-License-Identifier: Apache-2.0
package ebpfgtp

import (
	"errors"
	"net/netip"
	"testing"

	"github.com/cilium/ebpf"
	"github.com/stretchr/testify/require"
)

type memoryStore struct {
	values map[string]map[any]any
	events []string
	fail   func(string, string, any, any) error
}

func (m *memoryStore) put(name string, k, v any, flags ebpf.MapUpdateFlags) error {
	m.events = append(m.events, "put:"+name)
	if m.fail != nil {
		if err := m.fail("put", name, k, v); err != nil {
			return err
		}
	}
	if m.values[name] == nil {
		m.values[name] = map[any]any{}
	}
	_, exists := m.values[name][k]
	if flags == ebpf.UpdateNoExist && exists {
		return errors.New("key exists")
	}
	if flags == ebpf.UpdateExist && !exists {
		return errors.New("missing key")
	}
	m.values[name][k] = v
	return nil
}
func (m *memoryStore) del(name string, k any) error {
	m.events = append(m.events, "del:"+name)
	if m.fail != nil {
		if err := m.fail("del", name, k, nil); err != nil {
			return err
		}
	}
	delete(m.values[name], k)
	return nil
}

type fakeCloser struct {
	closed int
	err    error
}

func (f *fakeCloser) Close() error { f.closed++; return f.err }
func isolatedRegistry() (*Registry, *memoryStore, Config) {
	r := NewRegistry()
	state := &memoryStore{values: map[string]map[any]any{}}
	r.state = state
	r.discover = func(Config) (int, error) { return 10, nil }
	r.listen = func(netip.Addr) (closer, error) { return &fakeCloser{}, nil }
	r.attach = func(int) (closer, error) { return &fakeCloser{}, nil }
	c := Config{Local: netip.MustParseAddr("10.88.0.1"), Remote: netip.MustParseAddr("10.88.0.2"), IPv4: netip.MustParseAddr("10.60.0.1"), UplinkTEID: 1001, DownlinkTEID: 2001, QFI: 9, EndpointIfIndex: 20, MTU: 1456}
	return r, state, c
}
func TestProfileRejectsUnsupportedOrAmbiguousBinding(t *testing.T) {
	_, _, c := isolatedRegistry()
	require.NoError(t, c.Validate())
	for _, mutate := range []func(*Config){func(c *Config) { c.Local = netip.IPv6Loopback() }, func(c *Config) { c.Remote = c.Local }, func(c *Config) { c.MTU = 1457 }, func(c *Config) { c.QFI = 64 }, func(c *Config) { c.UplinkTEID = 0 }, func(c *Config) { c.EndpointIfIndex = 0 }} {
		invalid := c
		mutate(&invalid)
		require.Error(t, invalid.Validate())
	}
}
func TestAtomicHandoverFailurePreservesSourceAndReleasesStaging(t *testing.T) {
	for _, failure := range []string{"listen", "canonical"} {
		t.Run(failure, func(t *testing.T) {
			r, state, c := isolatedRegistry()
			s, err := r.Open(c)
			require.NoError(t, err)
			sourceSocket := r.locals[c.Local].socket.(*fakeCloser)
			next := c
			next.Local = netip.MustParseAddr("10.88.0.3")
			next.UplinkTEID = 1002
			next.DownlinkTEID = 2002
			if failure == "listen" {
				r.listen = func(netip.Addr) (closer, error) { return nil, errors.New("occupied target port") }
			} else {
				state.fail = func(op, name string, k, v any) error {
					if op == "put" && name == "sessions" && v.(binding).Local == ipv4(next.Local) {
						return errors.New("canonical update failed")
					}
					return nil
				}
			}
			require.Error(t, s.Update(next))
			require.Equal(t, c, s.cfg)
			require.Equal(t, c.binding(), state.values["sessions"][ipv4(c.IPv4)])
			require.Equal(t, map[downKey]bool{c.downKey(): true}, s.keys)
			require.Len(t, r.locals, 1)
			require.Zero(t, sourceSocket.closed)
			state.fail = nil
			require.NoError(t, s.Close())
			require.Equal(t, 1, sourceSocket.closed)
			require.Empty(t, r.sessions)
			require.Empty(t, r.links)
		})
	}
}
func TestSharedSocketAndPeerTEIDCollisionOwnership(t *testing.T) {
	r, _, c := isolatedRegistry()
	a, err := r.Open(c)
	require.NoError(t, err)
	_, err = r.Open(c)
	require.Error(t, err)
	other := c
	other.IPv4 = netip.MustParseAddr("10.60.0.2")
	other.EndpointIfIndex = 21
	_, err = r.Open(other)
	require.Error(t, err, "same local/peer/downlink TEID cannot be shared")
	require.Equal(t, 1, r.locals[c.Local].refs)
	other.DownlinkTEID++
	other.UplinkTEID++
	b, err := r.Open(other)
	require.NoError(t, err)
	require.Equal(t, 2, r.locals[c.Local].refs)
	socket := r.locals[c.Local].socket.(*fakeCloser)
	require.NoError(t, a.Close())
	require.Zero(t, socket.closed)
	require.Len(t, r.links, 1)
	require.NoError(t, b.Close())
	require.Equal(t, 1, socket.closed)
	require.Empty(t, r.locals)
	require.Empty(t, r.links)
}
func TestCommittedHandoverRetainsCleanupOwnershipAndRetries(t *testing.T) {
	r, state, c := isolatedRegistry()
	s, err := r.Open(c)
	require.NoError(t, err)
	next := c
	next.Local = netip.MustParseAddr("10.88.0.3")
	next.DownlinkTEID++
	next.UplinkTEID++
	warnings := 0
	r.SetWarningHandler(func(error) { warnings++ })
	state.fail = func(op, name string, k, v any) error {
		if op == "del" && name == "locals" && k == ipv4(c.Local) {
			return errors.New("retirement failed")
		}
		return nil
	}
	require.NoError(t, s.Update(next), "canonical replacement has committed")
	require.Equal(t, next.binding(), state.values["sessions"][ipv4(c.IPv4)])
	require.Equal(t, next, s.cfg)
	require.Equal(t, 1, warnings)
	require.Len(t, r.locals, 2)
	require.Error(t, s.Close())
	require.False(t, s.closed)
	require.Contains(t, r.locals, c.Local, "old port must remain reserved while its ingress ownership remains")
	state.fail = nil
	require.NoError(t, s.Close())
	require.Empty(t, r.locals)
	require.Empty(t, r.links)
	require.Empty(t, r.sessions)
}
func TestFailedDeactivationCannotRetireEndpointOrPort(t *testing.T) {
	r, state, c := isolatedRegistry()
	s, err := r.Open(c)
	require.NoError(t, err)
	socket := r.locals[c.Local].socket.(*fakeCloser)
	state.fail = func(op, name string, k, v any) error {
		if name == "sessions" {
			return errors.New("map operation denied")
		}
		return nil
	}
	require.Error(t, s.Close())
	require.False(t, s.closed)
	require.Equal(t, c.binding(), state.values["sessions"][ipv4(c.IPv4)])
	require.Zero(t, socket.closed)
	require.Len(t, r.links, 1)
	state.fail = func(op, name string, k, v any) error {
		if op == "del" && name == "sessions" {
			return errors.New("delete denied")
		}
		return nil
	}
	require.Error(t, s.Close())
	require.Equal(t, binding{}, state.values["sessions"][ipv4(c.IPv4)], "fallback must fail closed")
	require.Zero(t, socket.closed)
	state.fail = nil
	require.NoError(t, s.Close())
	require.Equal(t, 1, socket.closed)
}
func TestFailedInitialCommitRetainsAndReapsOrphanedClaims(t *testing.T) {
	r, state, c := isolatedRegistry()
	state.fail = func(op, name string, k, v any) error {
		if op == "put" && name == "sessions" || op == "del" && name == "downlinks" {
			return errors.New("injected map failure")
		}
		return nil
	}
	_, err := r.Open(c)
	require.Error(t, err)
	require.Empty(t, state.values["sessions"])
	require.Len(t, r.orphans, 1)
	require.Len(t, state.values["downlinks"], 1)
	state.fail = nil
	s, err := r.Open(c)
	require.NoError(t, err)
	require.Empty(t, r.orphans)
	require.NoError(t, s.Close())
	require.Empty(t, state.values["downlinks"])
}

func TestFailedAcquireRetainsUnclosedAttachmentAndPortClaims(t *testing.T) {
	r, state, c := isolatedRegistry()
	attachment := &fakeCloser{err: errors.New("detach failed")}
	socket := &fakeCloser{err: errors.New("close failed")}
	r.attach = func(int) (closer, error) { return attachment, nil }
	r.listen = func(netip.Addr) (closer, error) { return socket, nil }
	state.fail = func(op, name string, k, v any) error {
		if op == "put" && name == "locals" {
			return errors.New("local map insert failed")
		}
		return nil
	}
	_, err := r.Open(c)
	require.Error(t, err)
	require.Same(t, attachment, r.links[10].attachment)
	require.Zero(t, r.links[10].refs)
	require.Same(t, socket, r.ports[c.Local])
	require.Empty(t, r.sessions)
	attachment.err = nil
	socket.err = nil
	state.fail = nil
	r.attach = func(int) (closer, error) { return &fakeCloser{}, nil }
	r.listen = func(netip.Addr) (closer, error) { return &fakeCloser{}, nil }
	s, err := r.Open(c)
	require.NoError(t, err)
	require.Empty(t, r.ports)
	require.GreaterOrEqual(t, attachment.closed, 2)
	require.GreaterOrEqual(t, socket.closed, 2)
	require.NoError(t, s.Close())
}
func TestFailedSocketCloseRetainsLeaseWithoutDoubleRetiringSharedLink(t *testing.T) {
	r, _, c := isolatedRegistry()
	a, err := r.Open(c)
	require.NoError(t, err)
	other := c
	other.Local = netip.MustParseAddr("10.88.0.3")
	other.IPv4 = netip.MustParseAddr("10.60.0.2")
	other.EndpointIfIndex++
	other.DownlinkTEID++
	other.UplinkTEID++
	b, err := r.Open(other)
	require.NoError(t, err)
	attachment := r.links[10].attachment.(*fakeCloser)
	socket := r.locals[c.Local].socket.(*fakeCloser)
	socket.err = errors.New("close failed")
	require.Error(t, a.Close())
	require.Contains(t, r.locals, c.Local)
	require.False(t, r.locals[c.Local].mapped)
	require.False(t, r.locals[c.Local].linked)
	require.Equal(t, 1, r.links[10].refs)
	require.Zero(t, attachment.closed)
	socket.err = nil
	require.NoError(t, a.Close())
	require.Equal(t, 1, r.links[10].refs, "retry must not release the other session's reference")
	require.Zero(t, attachment.closed)
	require.NoError(t, b.Close())
	require.Equal(t, 1, attachment.closed)
	require.Empty(t, r.locals)
}
func TestEchoResponseRequiresExactControlHeader(t *testing.T) {
	request := []byte{0x32, 1, 0, 4, 0, 0, 0, 0, 0x12, 0x34, 0, 0}
	require.Equal(t, []byte{0x32, 2, 0, 6, 0, 0, 0, 0, 0x12, 0x34, 0, 0, 14, 0}, echoResponse(request))
	for _, index := range []int{0, 1, 3, 4, 10, 11} {
		bad := append([]byte(nil), request...)
		bad[index] ^= 1
		require.Nil(t, echoResponse(bad))
	}
	require.Nil(t, echoResponse(request[:11]))
	require.Nil(t, echoResponse(append(request, 0)))
}
