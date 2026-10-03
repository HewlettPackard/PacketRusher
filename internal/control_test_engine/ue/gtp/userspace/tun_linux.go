//go:build linux

// SPDX-License-Identifier: Apache-2.0
package userspace

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"os"
)

// NewTUN creates a nonpersistent L3 TUN. Closing its descriptor removes only this
// process's interface; an existing interface name is never deleted or adopted.
func NewTUN(name string) (PacketPort, netlink.Link, error) {
	if _, err := netlink.LinkByName(name); err == nil {
		return nil, nil, fmt.Errorf("interface %s already exists", name)
	}
	fd, err := unix.Open("/dev/net/tun", unix.O_RDWR|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("open TUN (requires /dev/net/tun and CAP_NET_ADMIN): %w", err)
	}
	req, err := unix.NewIfreq(name)
	if err != nil {
		unix.Close(fd)
		return nil, nil, err
	}
	req.SetUint16(unix.IFF_TUN | unix.IFF_NO_PI | unix.IFF_TUN_EXCL)
	if err = unix.IoctlIfreq(fd, unix.TUNSETIFF, req); err != nil {
		unix.Close(fd)
		return nil, nil, fmt.Errorf("create TUN %s: %w", name, err)
	}
	file := os.NewFile(uintptr(fd), name)
	link, err := netlink.LinkByName(req.Name())
	if err != nil {
		file.Close()
		return nil, nil, err
	}
	return file, link, nil
}
