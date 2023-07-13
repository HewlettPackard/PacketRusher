package link

import (
	"fmt"
	"net"
	"os"
	"path"
	"sync"
	"syscall"

	"github.com/free5gc/go-gtp5gnl"
	"github.com/khirono/go-nl"
	"github.com/khirono/go-rtnllink"
)

func CmdDel(ifname string) error {
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

	conn, err := nl.Open(syscall.NETLINK_ROUTE)
	if err != nil {
		return err
	}
	defer conn.Close()

	c := nl.NewClient(conn, mux)

	err = rtnllink.Remove(c, ifname)
	if err != nil {
		return err
	}

	return nil
}

func CmdAdd(ifname string, role int, ip string, stopSignal chan bool) error {
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

	conn, err := nl.Open(syscall.NETLINK_ROUTE)
	if err != nil {
		return err
	}
	defer conn.Close()

	c := nl.NewClient(conn, mux)

	laddr, err := net.ResolveUDPAddr("udp", ip+":2152")
	if err != nil {
		return err
	}
	conn2, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return err
	}
	defer conn2.Close()
	f, err := conn2.File()
	if err != nil {
		return err
	}
	defer f.Close()

	linkinfo := &nl.Attr{
		Type: syscall.IFLA_LINKINFO,
		Value: nl.AttrList{
			{
				Type:  rtnllink.IFLA_INFO_KIND,
				Value: nl.AttrString("gtp5g"),
			},
			{
				Type: rtnllink.IFLA_INFO_DATA,
				Value: nl.AttrList{
					{
						Type:  gtp5gnl.IFLA_FD1,
						Value: nl.AttrU32(f.Fd()),
					},
					{
						Type:  gtp5gnl.IFLA_HASHSIZE,
						Value: nl.AttrU32(131072),
					},
					{
						Type:  gtp5gnl.IFLA_ROLE,
						Value: nl.AttrU32(role),
					},
				},
			},
		},
	}
	err = rtnllink.Create(c, ifname, linkinfo)
	if err != nil {
		return err
	}

	err = rtnllink.Up(c, ifname)
	if err != nil {
		return err
	}

	<-stopSignal

	return nil
}

func usage(prog string) {
	fmt.Fprintf(os.Stderr, "usage: %v <add|del> <ifname> [--ran]\n", prog)
}

func main() {
	prog := path.Base(os.Args[0])
	if len(os.Args) < 4 {
		usage(prog)
		os.Exit(1)
	}
	cmd := os.Args[1]
	ifname := os.Args[2]
	ip := os.Args[3]
	var role int
	if len(os.Args) > 4 && os.Args[4] == "--ran" {
		role = 1
	}

	switch cmd {
	case "add":
		err := CmdAdd(ifname, role, ip, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: %v\n", prog, err)
			os.Exit(1)
		}
	case "del":
		err := CmdDel(ifname)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: %v\n", prog, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "%v: unknown command %q\n", prog, cmd)
		os.Exit(1)
	}
}
