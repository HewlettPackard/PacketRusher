// SPDX-License-Identifier: Apache-2.0
package userspace

import "my5G-RANTester/internal/control_test_engine/ue/gtp/ipv6"

func routerAdvertisement(packet []byte) bool { return ipv6.RouterAdvertisement(packet) }
