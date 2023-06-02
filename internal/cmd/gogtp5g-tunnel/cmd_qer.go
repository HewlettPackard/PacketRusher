package tunnel

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"syscall"

	"github.com/free5gc/go-gtp5gnl"
	"github.com/khirono/go-nl"
)

func ParseQEROptions(args []string) ([]nl.Attr, error) {
	var attrs []nl.Attr
	var mbrv nl.AttrList
	var gbrv nl.AttrList
	p := NewCmdParser(args)
	for {
		opt, ok := p.GetToken()
		if !ok {
			break
		}
		switch opt {
		case "--gate-status":
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 8)
			if err != nil {
				return attrs, err
			}
			attrs = append(attrs, nl.Attr{
				Type:  gtp5gnl.QER_GATE,
				Value: nl.AttrU8(v),
			})
		case "--mbr-uhigh":
			return attrs, fmt.Errorf("option %q is deprecated", opt)
		case "--mbr-ulow":
			return attrs, fmt.Errorf("option %q is deprecated", opt)
		case "--mbr-dhigh":
			return attrs, fmt.Errorf("option %q is deprecated", opt)
		case "--mbr-dlow":
			return attrs, fmt.Errorf("option %q is deprecated", opt)
		case "--gbr-uhigh":
			return attrs, fmt.Errorf("option %q is deprecated", opt)
		case "--gbr-ulow":
			return attrs, fmt.Errorf("option %q is deprecated", opt)
		case "--gbr-dhigh":
			return attrs, fmt.Errorf("option %q is deprecated", opt)
		case "--gbr-dlow":
			return attrs, fmt.Errorf("option %q is deprecated", opt)
		case "--mbr-ul":
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 40)
			if err != nil {
				return attrs, err
			}
			mbrv = append(mbrv, nl.Attr{
				Type:  gtp5gnl.QER_MBR_UL_HIGH32,
				Value: nl.AttrU32(v >> 8),
			})
			mbrv = append(mbrv, nl.Attr{
				Type:  gtp5gnl.QER_MBR_UL_LOW8,
				Value: nl.AttrU8(v),
			})
		case "--mbr-dl":
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 40)
			if err != nil {
				return attrs, err
			}
			mbrv = append(mbrv, nl.Attr{
				Type:  gtp5gnl.QER_MBR_DL_HIGH32,
				Value: nl.AttrU32(v >> 8),
			})
			mbrv = append(mbrv, nl.Attr{
				Type:  gtp5gnl.QER_MBR_DL_LOW8,
				Value: nl.AttrU8(v),
			})
		case "--gbr-ul":
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 40)
			if err != nil {
				return attrs, err
			}
			gbrv = append(gbrv, nl.Attr{
				Type:  gtp5gnl.QER_GBR_UL_HIGH32,
				Value: nl.AttrU32(v >> 8),
			})
			gbrv = append(gbrv, nl.Attr{
				Type:  gtp5gnl.QER_GBR_UL_LOW8,
				Value: nl.AttrU8(v),
			})
		case "--gbr-dl":
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 40)
			if err != nil {
				return attrs, err
			}
			gbrv = append(gbrv, nl.Attr{
				Type:  gtp5gnl.QER_GBR_DL_HIGH32,
				Value: nl.AttrU32(v >> 8),
			})
			gbrv = append(gbrv, nl.Attr{
				Type:  gtp5gnl.QER_GBR_DL_LOW8,
				Value: nl.AttrU8(v),
			})
		case "--qer-corr-id":
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 32)
			if err != nil {
				return attrs, err
			}
			attrs = append(attrs, nl.Attr{
				Type:  gtp5gnl.QER_CORR_ID,
				Value: nl.AttrU32(v),
			})
		case "--rqi":
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 8)
			if err != nil {
				return attrs, err
			}
			attrs = append(attrs, nl.Attr{
				Type:  gtp5gnl.QER_RQI,
				Value: nl.AttrU8(v),
			})
		case "--qfi":
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 8)
			if err != nil {
				return attrs, err
			}
			attrs = append(attrs, nl.Attr{
				Type:  gtp5gnl.QER_QFI,
				Value: nl.AttrU8(v),
			})
		case "--ppi":
			arg, ok := p.GetToken()
			if !ok {
				return attrs, fmt.Errorf("option requires argument %q", opt)
			}
			v, err := strconv.ParseUint(arg, 0, 8)
			if err != nil {
				return attrs, err
			}
			attrs = append(attrs, nl.Attr{
				Type:  gtp5gnl.QER_PPI,
				Value: nl.AttrU8(v),
			})
		case "--rcsr":
			return attrs, fmt.Errorf("option %q is deprecated", opt)
		default:
			return attrs, fmt.Errorf("unknown option %q", opt)
		}
	}

	if len(mbrv) != 0 {
		attrs = append(attrs, nl.Attr{
			Type:  gtp5gnl.QER_MBR,
			Value: mbrv,
		})
	}
	if len(gbrv) != 0 {
		attrs = append(attrs, nl.Attr{
			Type:  gtp5gnl.QER_GBR,
			Value: gbrv,
		})
	}

	return attrs, nil
}

// add qer <ifname> <oid> [options...]
func CmdAddQER(args []string) error {
	if len(args) < 2 {
		return errors.New("too few parameter")
	}
	ifname := args[0]
	oid, err := ParseOID(args[1])
	if err != nil {
		return err
	}
	attrs, err := ParseQEROptions(args[2:])
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

	return gtp5gnl.CreateQEROID(c, link, oid, attrs)
}

// mod qer <ifname> <oid> [options...]
func CmdModQER(args []string) error {
	if len(args) < 2 {
		return errors.New("too few parameter")
	}
	ifname := args[0]
	oid, err := ParseOID(args[1])
	if err != nil {
		return err
	}
	attrs, err := ParseQEROptions(args[2:])
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

	return gtp5gnl.UpdateQEROID(c, link, oid, attrs)
}

// delete far <ifname> <oid>
func CmdDeleteQER(args []string) error {
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

	return gtp5gnl.RemoveQEROID(c, link, oid)
}

// get qer <ifname> <oid>
func CmdGetQER(args []string) error {
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

	qer, err := gtp5gnl.GetQEROID(c, link, oid)
	if err != nil {
		return err
	}

	j, err := json.MarshalIndent(qer, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", j)
	return nil
}

// list qer
func CmdListQER(args []string) error {
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

	qers, err := gtp5gnl.GetQERAll(c)
	if err != nil {
		return err
	}

	j, err := json.MarshalIndent(qers, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", j)
	return nil
}
