// SPDX-License-Identifier: Apache-2.0

package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
)

func TestEndpointMTUCommitsWithRouteOrRestoresSource(t *testing.T) {
	for _, tc := range []struct {
		name                              string
		preserve, badSet, badRoute        bool
		badRestore                        bool
		wantMTU, wantSets, wantRouteCalls int
	}{
		{"retained commit", true, false, false, false, 1356, 1, 1},
		{"new endpoint commit", false, false, false, false, 1356, 1, 1},
		{"retained route failure", true, false, true, false, 1456, 2, 1},
		{"new endpoint route failure", false, false, true, false, 1356, 1, 1},
		{"retained partially applied MTU failure", true, true, false, false, 1456, 2, 0},
		{"new endpoint MTU failure", false, true, false, false, 1356, 1, 0},
		{"rollback failure surfaced", true, false, true, true, 1356, 2, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			oldSet, oldRoute := setUEEndpointMTU, replaceTunnelRoute
			t.Cleanup(func() { setUEEndpointMTU, replaceTunnelRoute = oldSet, oldRoute })
			endpoint := &netlink.Dummy{LinkAttrs: netlink.LinkAttrs{Name: "val123", MTU: 1456}}
			route := &netlink.Route{Table: 1000, LinkIndex: 2}
			setFailure, routeFailure, rollbackFailure := errors.New("MTU failed"), errors.New("route failed"), errors.New("rollback failed")
			kernelMTU, sets, routeCalls := 1456, 0, 0
			setUEEndpointMTU = func(link netlink.Link, mtu int) error {
				require.Same(t, endpoint, link)
				sets++
				if mtu == 1456 && tc.badRestore {
					return rollbackFailure
				}
				kernelMTU = mtu
				if mtu == 1356 && tc.badSet {
					return setFailure // An error may follow a partially applied update.
				}
				return nil
			}
			replaceTunnelRoute = func(target *netlink.Route) error {
				require.Same(t, route, target)
				require.Equal(t, 1356, kernelMTU, "endpoint MTU must be ready before routing switches")
				require.Equal(t, kernelMTU, endpoint.Attrs().MTU)
				routeCalls++
				if tc.badRoute {
					return routeFailure
				}
				return nil
			}
			err := replaceRouteWithEndpointMTU(route, endpoint, 1356, tc.preserve)
			if tc.badSet {
				require.ErrorIs(t, err, setFailure)
			} else if tc.badRoute {
				require.ErrorIs(t, err, routeFailure)
			} else {
				require.NoError(t, err)
			}
			if tc.badRestore {
				require.ErrorIs(t, err, rollbackFailure)
				require.ErrorIs(t, err, errTunnelRollback)
				require.ErrorContains(t, err, "source endpoint MTU rollback failed")
			}
			require.Equal(t, tc.wantMTU, kernelMTU)
			if !tc.badSet || tc.preserve {
				require.Equal(t, kernelMTU, endpoint.Attrs().MTU)
			}
			require.Equal(t, tc.wantSets, sets)
			require.Equal(t, tc.wantRouteCalls, routeCalls)
		})
	}
}
