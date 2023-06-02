package main

import (
	"fmt"
	"os"
	"path"
	"strings"
)

func usage(prog string) {
	fmt.Fprintf(os.Stderr, `Usage:
    %v <add|mod> <pdr|far|qer> <ifname> <oid> [<options>...]
    %v <delete|get> <pdr|far|qer> <ifname> <oid>
    %v list <pdr|far|qer>

OID format:
    <id>
    or
    <seid>:<id>

PDR Options:
    --pcd <precedence>
    --hdr-rm <outer-header-removal>
    --far-id <existed-far-id>
    --ue-ipv4 <pdi-ue-ipv4>
    --f-teid <i-teid> <local-gtpu-ipv4>
    --sdf-desp <description-string>
    --sdf-tos-traff-cls <tos-traffic-class>
    --sdf-scy-param-idx <security-param-idx>
    --sdf-flow-label <flow-label>
    --sdf-id <id>
    --qer-id <id>
    --gtpu-src-ip <gtpu-src-ip>
    --buffer-usock-path <AF_UNIX-sock-path>

FAR Options:
    --action <apply-action>
    --hdr-creation <description> <o-teid> <peer-ipv4> <peer-port>
    --fwd-policy <mark set in iptable>

QER Options:
    --gate-status <gate-status>
    --mbr-ul <mbr-uplink>
    --mbr-dl <mbr-downlink>
    --gbr-ul <gbr-uplink>
    --gbr-dl <gbr-downlink>
    --qer-corr-id <qer-corr-id>
    --rqi <rqi>
    --qfi <qfi>
    --ppi <ppi>
`, prog, prog, prog)
}

func main() {
	prog := path.Base(os.Args[0])
	if len(os.Args) < 3 {
		usage(prog)
		os.Exit(1)
	}

	cmd, args := CmdTree.Find(os.Args[1:])
	if cmd == nil {
		fmt.Fprintf(os.Stderr, "%v: unknown command %q\n", prog, strings.Join(os.Args[1:], " "))
		os.Exit(1)
	}

	err := cmd(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: %v\n", prog, err)
		os.Exit(1)
	}
}
