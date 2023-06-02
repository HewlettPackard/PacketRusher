package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"syscall"

	"github.com/free5gc/go-gtp5gnl"
	"github.com/khirono/go-nl"
)

func ParseFAROptions(args []string) ([]nl.Attr, error) {
	var attrs []nl.Attr
	var paramv nl.AttrList
	p := NewCmdParser(args)
	for {
		opt, ok := p.GetToken()
		if !ok {
			break
		}
		switch opt {
		case "--action":
			// --action <apply-action>
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 8)
			if err != nil {
				return attrs, err
			}
			attrs = append(attrs, nl.Attr{
				Type:  gtp5gnl.FAR_APPLY_ACTION,
				Value: nl.AttrU16(v),
			})
		case "--hdr-creation":
			// --hdr-creation <description> <o-teid> <peer-ipv4> <peer-port>
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			desc, err := strconv.ParseUint(arg, 0, 16)
			if err != nil {
				return attrs, err
			}
			arg2, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			teid, err := strconv.ParseUint(arg2, 0, 32)
			if err != nil {
				return attrs, err
			}
			arg3, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			addr := net.ParseIP(arg3)
			if addr == nil {
				return attrs, fmt.Errorf("invalid IP address %q", arg3)
			}
			addr = addr.To4()
			if addr == nil {
				return attrs, fmt.Errorf("option value is not IPv4 %q", arg3)
			}
			arg4, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			port, err := strconv.ParseUint(arg4, 0, 16)
			if err != nil {
				return attrs, err
			}
			paramv = append(paramv, nl.Attr{
				Type: gtp5gnl.FORWARDING_PARAMETER_OUTER_HEADER_CREATION,
				Value: nl.AttrList{
					{
						Type:  gtp5gnl.OUTER_HEADER_CREATION_DESCRIPTION,
						Value: nl.AttrU8(desc),
					},
					{
						Type:  gtp5gnl.OUTER_HEADER_CREATION_O_TEID,
						Value: nl.AttrU32(teid),
					},
					{
						Type:  gtp5gnl.OUTER_HEADER_CREATION_PEER_ADDR_IPV4,
						Value: nl.AttrBytes(addr),
					},
					{
						Type:  gtp5gnl.OUTER_HEADER_CREATION_PORT,
						Value: nl.AttrU16(port),
					},
				},
			})
		case "--fwd-policy":
			// --fwd-policy <mark set in iptable>
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			paramv = append(paramv, nl.Attr{
				Type:  gtp5gnl.FORWARDING_PARAMETER_FORWARDING_POLICY,
				Value: nl.AttrString(arg),
			})
		default:
			return attrs, fmt.Errorf("unknown option %q", opt)
		}
	}

	if len(paramv) != 0 {
		attrs = append(attrs, nl.Attr{
			Type:  gtp5gnl.FAR_FORWARDING_PARAMETER,
			Value: paramv,
		})
	}

	return attrs, nil
}

// add far <ifname> <oid> [options...]
func CmdAddFAR(args []string) error {
	if len(args) < 2 {
		return errors.New("too few parameter")
	}
	ifname := args[0]
	oid, err := ParseOID(args[1])
	if err != nil {
		return err
	}
	attrs, err := ParseFAROptions(args[2:])
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	mux, err := nl.NewMux()
	if err != nil {
		return err
	}
	defer func() {
		mux.Close()
		wg.Wait()
	}()
	wg.Add(1)
	go func() {
		mux.Serve()
		wg.Done()
	}()

	conn, err := nl.Open(syscall.NETLINK_GENERIC)
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := gtp5gnl.NewClient(conn, mux)
	if err != nil {
		return err
	}

	link, err := gtp5gnl.GetLink(ifname)
	if err != nil {
		return err
	}

	return gtp5gnl.CreateFAROID(c, link, oid, attrs)
}

// mod far <ifname> <oid> [options...]
func CmdModFAR(args []string) error {
	if len(args) < 2 {
		return errors.New("too few parameter")
	}
	ifname := args[0]
	oid, err := ParseOID(args[1])
	if err != nil {
		return err
	}
	attrs, err := ParseFAROptions(args[2:])
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	mux, err := nl.NewMux()
	if err != nil {
		return err
	}
	defer func() {
		mux.Close()
		wg.Wait()
	}()
	wg.Add(1)
	go func() {
		mux.Serve()
		wg.Done()
	}()

	conn, err := nl.Open(syscall.NETLINK_GENERIC)
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := gtp5gnl.NewClient(conn, mux)
	if err != nil {
		return err
	}

	link, err := gtp5gnl.GetLink(ifname)
	if err != nil {
		return err
	}

	return gtp5gnl.UpdateFAROID(c, link, oid, attrs)
}

// delete far <ifname> <oid>
func CmdDeleteFAR(args []string) error {
	if len(args) < 2 {
		return errors.New("too few parameter")
	}
	ifname := args[0]
	oid, err := ParseOID(args[1])
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	mux, err := nl.NewMux()
	if err != nil {
		return err
	}
	defer func() {
		mux.Close()
		wg.Wait()
	}()
	wg.Add(1)
	go func() {
		mux.Serve()
		wg.Done()
	}()

	conn, err := nl.Open(syscall.NETLINK_GENERIC)
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := gtp5gnl.NewClient(conn, mux)
	if err != nil {
		return err
	}

	link, err := gtp5gnl.GetLink(ifname)
	if err != nil {
		return err
	}

	return gtp5gnl.RemoveFAROID(c, link, oid)
}

// get far <ifname> <oid>
func CmdGetFAR(args []string) error {
	if len(args) < 2 {
		return errors.New("too few parameter")
	}
	ifname := args[0]
	oid, err := ParseOID(args[1])
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	mux, err := nl.NewMux()
	if err != nil {
		return err
	}
	defer func() {
		mux.Close()
		wg.Wait()
	}()
	wg.Add(1)
	go func() {
		mux.Serve()
		wg.Done()
	}()

	conn, err := nl.Open(syscall.NETLINK_GENERIC)
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := gtp5gnl.NewClient(conn, mux)
	if err != nil {
		return err
	}

	link, err := gtp5gnl.GetLink(ifname)
	if err != nil {
		return err
	}

	far, err := gtp5gnl.GetFAROID(c, link, oid)
	if err != nil {
		return err
	}

	j, err := json.MarshalIndent(far, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", j)
	return nil
}

// list far
func CmdListFAR(args []string) error {
	var wg sync.WaitGroup
	mux, err := nl.NewMux()
	if err != nil {
		return err
	}
	defer func() {
		mux.Close()
		wg.Wait()
	}()
	wg.Add(1)
	go func() {
		mux.Serve()
		wg.Done()
	}()

	conn, err := nl.Open(syscall.NETLINK_GENERIC)
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := gtp5gnl.NewClient(conn, mux)
	if err != nil {
		return err
	}

	fars, err := gtp5gnl.GetFARAll(c)
	if err != nil {
		return err
	}

	j, err := json.MarshalIndent(fars, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", j)
	return nil
}
