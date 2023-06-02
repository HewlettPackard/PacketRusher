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

func ParsePDROptions(args []string) ([]nl.Attr, error) {
	var attrs []nl.Attr
	var sdfv nl.AttrList
	var pdiv nl.AttrList
	p := NewCmdParser(args)
	for {
		opt, ok := p.GetToken()
		if !ok {
			break
		}
		switch opt {
		case "--pcd":
			// --pcd <precedence>
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 32)
			if err != nil {
				return attrs, err
			}
			attrs = append(attrs, nl.Attr{
				Type:  gtp5gnl.PDR_PRECEDENCE,
				Value: nl.AttrU32(v),
			})
		case "--hdr-rm":
			// --hdr-rm <outer-header-removal>
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 8)
			if err != nil {
				return attrs, err
			}
			attrs = append(attrs, nl.Attr{
				Type:  gtp5gnl.PDR_OUTER_HEADER_REMOVAL,
				Value: nl.AttrU8(v),
			})
		case "--far-id":
			// --far-id <existed-far-id>
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 32)
			if err != nil {
				return attrs, err
			}
			attrs = append(attrs, nl.Attr{
				Type:  gtp5gnl.PDR_FAR_ID,
				Value: nl.AttrU32(v),
			})
		case "--ue-ipv4":
			// --ue-ipv4 <pdi-ue-ipv4>
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v := net.ParseIP(arg)
			if v == nil {
				return attrs, fmt.Errorf("invalid IP address %q", arg)
			}
			v = v.To4()
			if v == nil {
				return attrs, fmt.Errorf("option value is not IPv4 %q", arg)
			}
			pdiv = append(pdiv, nl.Attr{
				Type:  gtp5gnl.PDI_UE_ADDR_IPV4,
				Value: nl.AttrBytes(v),
			})
		case "--f-teid":
			// --f-teid <i-teid> <local-gtpu-ipv4>
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			teid, err := strconv.ParseUint(arg, 0, 32)
			if err != nil {
				return attrs, err
			}
			arg2, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			addr := net.ParseIP(arg2)
			if addr == nil {
				return attrs, fmt.Errorf("invalid IP address %q", arg2)
			}
			addr = addr.To4()
			if addr == nil {
				return attrs, fmt.Errorf("option value is not IPv4 %q", arg2)
			}
			pdiv = append(pdiv, nl.Attr{
				Type: gtp5gnl.PDI_F_TEID,
				Value: nl.AttrList{
					{
						Type:  gtp5gnl.F_TEID_I_TEID,
						Value: nl.AttrU32(teid),
					},
					{
						Type:  gtp5gnl.F_TEID_GTPU_ADDR_IPV4,
						Value: nl.AttrBytes(addr),
					},
				},
			})
		case "--sdf-desp":
			// --sdf-desp <description-string>
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			fd, err := ParseFlowDesc(arg)
			if err != nil {
				return attrs, err
			}
			sdfv = append(sdfv, nl.Attr{
				Type:  gtp5gnl.SDF_FILTER_FLOW_DESCRIPTION,
				Value: fd,
			})
		case "--sdf-tos-traff-cls":
			// --sdf-tos-traff-cls <tos-traffic-class>
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 16)
			if err != nil {
				return attrs, err
			}
			sdfv = append(sdfv, nl.Attr{
				Type:  gtp5gnl.SDF_FILTER_TOS_TRAFFIC_CLASS,
				Value: nl.AttrU16(v),
			})
		case "--sdf-scy-param-idx":
			// --sdf-scy-param-idx <security-param-idx>
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 32)
			if err != nil {
				return attrs, err
			}
			sdfv = append(sdfv, nl.Attr{
				Type:  gtp5gnl.SDF_FILTER_SECURITY_PARAMETER_INDEX,
				Value: nl.AttrU32(v),
			})
		case "--sdf-flow-label":
			// --sdf-flow-label <flow-label>
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 32)
			if err != nil {
				return attrs, err
			}
			sdfv = append(sdfv, nl.Attr{
				Type:  gtp5gnl.SDF_FILTER_FLOW_LABEL,
				Value: nl.AttrU32(v),
			})
		case "--sdf-id":
			// --sdf-id <id>
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 32)
			if err != nil {
				return attrs, err
			}
			sdfv = append(sdfv, nl.Attr{
				Type:  gtp5gnl.SDF_FILTER_SDF_FILTER_ID,
				Value: nl.AttrU32(v),
			})
		case "--qer-id":
			// --qer-id <id>
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 32)
			if err != nil {
				return attrs, err
			}
			attrs = append(attrs, nl.Attr{
				Type:  gtp5gnl.PDR_QER_ID,
				Value: nl.AttrU32(v),
			})
		case "--gtpu-src-ip":
			// --gtpu-src-ip <gtpu-src-ip>
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v := net.ParseIP(arg)
			if v == nil {
				return attrs, fmt.Errorf("invalid IP address %q", arg)
			}
			v = v.To4()
			if v == nil {
				return attrs, fmt.Errorf("option value is not IPv4 %q", arg)
			}
			attrs = append(attrs, nl.Attr{
				Type:  gtp5gnl.PDR_ROLE_ADDR_IPV4,
				Value: nl.AttrBytes(v),
			})
		case "--buffer-usock-path":
			// --buffer-usock-path <AF_UNIX-sock-path>
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			attrs = append(attrs, nl.Attr{
				Type:  gtp5gnl.PDR_UNIX_SOCKET_PATH,
				Value: nl.AttrString(arg),
			})
		default:
			return attrs, fmt.Errorf("unknown option %q", opt)
		}
	}

	if len(sdfv) != 0 {
		pdiv = append(pdiv, nl.Attr{
			Type:  gtp5gnl.PDI_SDF_FILTER,
			Value: sdfv,
		})
	}

	if len(pdiv) != 0 {
		attrs = append(attrs, nl.Attr{
			Type:  gtp5gnl.PDR_PDI,
			Value: pdiv,
		})
	}

	return attrs, nil
}

// add pdr <ifname> <oid> [options...]
func CmdAddPDR(args []string) error {
	if len(args) < 2 {
		return errors.New("too few parameter")
	}
	ifname := args[0]
	oid, err := ParseOID(args[1])
	if err != nil {
		return err
	}
	attrs, err := ParsePDROptions(args[2:])
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

	return gtp5gnl.CreatePDROID(c, link, oid, attrs)
}

// mod pdr <ifname> <oid> [options...]
func CmdModPDR(args []string) error {
	if len(args) < 2 {
		return errors.New("too few parameter")
	}
	ifname := args[0]
	oid, err := ParseOID(args[1])
	if err != nil {
		return err
	}
	attrs, err := ParsePDROptions(args[2:])
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

	return gtp5gnl.UpdatePDROID(c, link, oid, attrs)
}

// delete pdr <ifname> <oid>
func CmdDeletePDR(args []string) error {
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

	return gtp5gnl.RemovePDROID(c, link, oid)
}

// get pdr <ifname> <oid>
func CmdGetPDR(args []string) error {
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

	pdr, err := gtp5gnl.GetPDROID(c, link, oid)
	if err != nil {
		return err
	}

	j, err := json.MarshalIndent(pdr, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", j)
	return nil
}

// list pdr
func CmdListPDR(args []string) error {
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

	pdrs, err := gtp5gnl.GetPDRAll(c)
	if err != nil {
		return err
	}

	j, err := json.MarshalIndent(pdrs, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", j)
	return nil
}
