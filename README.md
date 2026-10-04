# PacketRusher

![PacketRusher Logo](docs/media/img/PacketRusher.png)

----
## Description
#### Now with an eBPF user plane: as fast as gtp5g or faster, with no kernel module to build!
See the [changelog](CHANGELOG.md) for what is new in v2.0.0.

PacketRusher is a tool dedicated to the performance testing and automatic validation of 5G Core Networks using simulated UE (user equipment) and gNodeB (5G base station).

If you have questions or comments, feel free to open an issue after **a careful** review of existing closed issues.

PacketRusher borrows libraries and data structures from the [free5gc project](https://github.com/free5gc/free5gc).

## Features
* Simulate multiple UEs and gNodeB from a single tool
  * We tested up to 10k UEs!
* Supports both N2 (NGAP) and N1 (NAS) interfaces for stress testing
* --pcap parameter to capture pcap of N1/N2 traffic
* Implements main control plane procedures:
  * SUCI Concealing/Deconcealment (Null-Scheme, Profile A (X25519), Profile B (P-256))
  * UE attach/detach (registration/identity request/authentication/security mode) procedures
  * Create/Delete PDU Sessions, up to 15 PDU Sessions per UE
  * Xn handover: UE handover between simulated gNodeB (PathSwitchRequest)
  * N2 handover: UE handover between simulated gNodeB (HandoverRequired)
  * UE Enter/Exit CM-IDLE procedures (Service Request) 
  * GUTI Re-registration
  * Supports 5G roaming: Tested with new https://github.com/open5gs/open5gs/issues/2194 Roaming feature
* Implements high-performant N3 (GTP-U) interface
  * No kernel module needed: eBPF programs carry the user plane in the kernel, with a plain userspace fallback
    * The gtp5g kernel module remains supported
  * IPv4, IPv6 and dual-stack PDU sessions
  * Generic tunnel supporting all kind of traffic (TCP, UDP, Video…)
    * We tested iperf3 traffic, and Youtube traffic through PacketRusher
    * With eBPF, a UE carries 3 to 5 Gbit/s per TCP stream and four UEs 10 to 14 Gbit/s in total, as much as with the gtp5g kernel module or more
* Runtime control of the UEs (idle, handover, deregistration…) through a local socket, and JSON scenarios
* Procedure results as a JSON report and Prometheus metrics
* Integrated all-in-one mocked 5GC/AMF for PacketRusher's integration testing

## Installation
### Quick start guide
The following is a quick start guide, for more details on the installation, configuration or usage, you may refer to the [wiki](https://github.com/HewlettPackard/PacketRusher/wiki).

### Requirements
- Ubuntu 20.04-24.04
  - All Linux distributions with kernel from 5.4 up to the 7.0.x series should work, but untested.
  - There might be issues with frankenstein kernel from RHEL/CentOS/Rocky, feel free to open a bug if you encounter one!
- Windows is not supported (Windows does not support SCTP)
- Root privilege
- Linux 6.6 or more recent for the eBPF user plane: older kernels use gtp5g or the userspace backend

### Download a release
```bash
$ sudo apt install wget tar linux-modules-extra-$(uname -r) # the last one provides SCTP
$ wget https://github.com/HewlettPackard/PacketRusher/releases/download/v2.0.0/packetrusher-v2.0.0-linux-amd64.tar.gz
$ tar -xzf packetrusher-v2.0.0-linux-amd64.tar.gz
$ ./packetrusher --help
```
An arm64 archive is published too, and a container image, `ghcr.io/hewlettpackard/packetrusher:v2.0.0`: see
[docker/README.md](docker/README.md). The host must provide SCTP.

### Or build from source
It needs Go 1.26.2 or more recent:
```bash
$ sudo apt install git wget tar linux-modules-extra-$(uname -r)
# Warning this command will remove your existing local Go installation if you have one:
$ wget https://go.dev/dl/go1.26.2.linux-amd64.tar.gz && sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.26.2.linux-amd64.tar.gz
# Add go binary to the executable PATH variable:
$ echo 'export PATH=$PATH:/usr/local/go/bin' >> $HOME/.profile
$ git clone https://github.com/HewlettPackard/PacketRusher # or download the ZIP from https://github.com/HewlettPackard/PacketRusher/archive/refs/heads/main.zip and upload it to your Linux server
$ cd PacketRusher && echo "export PACKETRUSHER=$PWD" >> $HOME/.profile
$ source $HOME/.profile
$ go mod download
$ go build -o packetrusher ./cmd
$ ./packetrusher --help
```

### Build free5gc's gtp5g kernel module (optional)
User-plane tunnels work without it, with the `ebpf` and `userspace` backends: see [GTP-U tunnel backends](#gtp-u-tunnel-backends).
```bash
$ sudo apt install build-essential linux-headers-generic make
$ cd $PACKETRUSHER/lib/gtp5g
$ make clean && make && sudo make install
# Make sure you have Secure boot disabled if you are unable to install the custom Kernel module
```

### Run it
Edit the configuration in `config/config.yml` as specified [here](https://github.com/HewlettPackard/PacketRusher/wiki/Configuration), and then run a basic scenario:
```bash
$ sudo ./packetrusher ue                       # one gNB and one UE, with its PDU session and tunnel
$ sudo ./packetrusher multi-ue -n 10 --tunnel   # ten UEs, see ./packetrusher multi-ue --help
```
For more details on the installation, configuration or usage, you may refer to the [wiki](https://github.com/HewlettPackard/PacketRusher/wiki).

## Usage

### GTP-U tunnel backends

`ue` and `multi-ue --tunnel` give each UE a network interface carrying its PDU session. The global
`--tunnel-backend` flag selects what carries this user plane:

- `ebpf`: eBPF programs short-cutting the userspace backend in the kernel. It needs Linux 6.6, the rights to
  load eBPF programs (root); no kernel module is needed.
- `gtp5g`: the [gtp5g](https://github.com/free5gc/gtp5g) kernel module.
- `userspace`: PacketRusher itself, with one TUN device per UE; no kernel module is needed.
- `auto` (default): the first of these that is available on the host.

Only `auto` falls back: a backend requested by name that is not available is an error.

```bash
sudo ./packetrusher --tunnel-backend userspace multi-ue -n 10 --tunnel
```

`ebpf` and `gtp5g` scale with the number of UEs, with or without `--dedicatedGnb`. With `userspace`, the UEs of a
gNB share one socket: use `--dedicatedGnb` when several UEs send traffic at once. The userspace backend also keeps
one TUN device, hence one file descriptor, open per UE.

As an indication, TCP throughput of iperf3 in Gbit/s, uplink / downlink, through a free5GC UPF, all on one 16-core
host (Linux 7.0):

| Backend | 1 UE | 4 UEs |
|---|---|---|
| `ebpf` | 3.4 / 4.7 | 10 / 14 |
| `gtp5g` | 2.9 / 2.8 | 11 / 10.5 |
| `userspace` | 1.1 / 1.4 | 0.9 / 1.3 |
| `userspace` with `--dedicatedGnb` | 1.1 / 1.4 | 3.6 / 6.2 |

One TCP stream is bound by one core: a single UE with four streams reaches the figures of four UEs. `test/real-core/throughput.sh`
measures your own setup: see its header.

Each UE has its address on a `val<MSIN>` device and its packets go through `gtp0<MSIN>` or `gtp1<MSIN>`, which
alternate at each handover: the UE keeps its address and its connections. To send traffic, bind to the address of
the UE (`ping -I <UE IP>`, `iperf3 -B <UE IP>`), not to its `val` device; with `--tunnel-vrf`, run the command in
the VRF of the UE instead: `sudo ip vrf exec vrf<MSIN> <command>`. With gtp5g and without `--dedicatedGnb`, the
UEs of a gNB share one device, `valgnb<N3 address in hexadecimal>`, which also holds their addresses.

The MTU of a tunnel is the MTU of the N3 interface minus the 44 bytes of the GTP-U encapsulation, 1456 on
Ethernet. Set `ue.tunnelmtu` in the configuration for a smaller one, that of the UPF for instance.

A UE whose tunnel cannot be set up logs the error and keeps running without it. What a killed run left on the
host, devices and routing rules, is removed by the next run of the same UEs.

### IPv6 and dual-stack PDU sessions

`ue.pdusessiontype` in the configuration file is the type of the PDU sessions the UEs request: `IPv4` (default), `IPv6` or `IPv4v6`.

The IPv6 user plane needs the `ebpf` or `userspace` tunnel backend: `auto` selects one of them for `IPv6`, whereas on `gtp5g` an `IPv4v6` session carries IPv4 only.
The SMF only allocates an interface identifier: the UE sends a Router Solicitation through its tunnel, and takes its address in the /64 prefix of the Router Advertisement with which the UPF, or the SMF through it, answers. PacketRusher then logs the address to bind to.

```bash
# config.yml: "pdusessiontype: IPv4v6" under "ue:"
sudo ./packetrusher --tunnel-backend userspace ue
```

### Runtime UE controls and JSON scenarios

`multi-ue --control-socket` creates a Unix socket to trigger procedures on the UEs of a running test, and
`--number-of-gnbs` additional gNBs to hand UEs over to, each with its own N2/N3 IP. From another terminal,
`control` runs one action, waits for its completion and prints the state of the UE as JSON:

```bash
./packetrusher multi-ue -n 2 --number-of-gnbs 2 --control-socket /tmp/packetrusher.sock
./packetrusher control --socket /tmp/packetrusher.sock --ue 1 --action wait --timeout 45s
./packetrusher control --socket /tmp/packetrusher.sock --ue 1 --action xn-handover --target 000009
./packetrusher run-scenario --socket /tmp/packetrusher.sock --scenario scenario.json
```

The actions are `inspect` (the default, of all the UEs without `--ue`), `wait` (until registered with its PDU
sessions), `idle`, `reconnect`, `xn-handover` and `ng-handover` (to the `--target` gNB ID), `deregister` and
`register`. `run-scenario` runs them in order from a JSON file, each step once the previous one completed:

```json
{"steps": [
  {"ue": 1, "action": "wait", "timeout_ms": 45000},
  {"ue": 1, "action": "xn-handover", "target": "000009"},
  {"ue": 1, "action": "deregister"}
]}
```

### Procedure results

`--report-json <file>` writes the results of the registrations and PDU session establishments when PacketRusher
stops, replacing the file. `--metrics-addr <ip:port>` serves them live at `/metrics` for Prometheus. Both are global flags:

```bash
./packetrusher --config config/config.yml --report-json results.json --metrics-addr 127.0.0.1:9090 multi-ue -n 100
```

Each of the report's `procedures` (`registration`, `pdu_session_establishment`) has `started`, `success`, `failure`, `pending`
and, for the successes, `latency_count` and `latency_seconds_sum`, `_min`, `_max` and `_mean`. The metrics, labelled by `procedure`, are
`packetrusher_procedure_started_total`, `packetrusher_procedure_completed_total` (by `outcome`), `packetrusher_procedure_pending`
and the histogram `packetrusher_procedure_duration_seconds`.

### gNB and NR cell identities

`gnodeb.plmnlist.gnbid` is a hexadecimal gNB ID. `gnbidlength` sets its width in bits (22 to 32, default 24) and
`cellid` the cell suffix filling the remaining bits of the 36-bit NR cell identity (default 0):

```yaml
gnodeb:
  plmnlist:
    gnbid: "01ABCDE"
    gnbidlength: 25
    cellid: 3
```

### Boolean flags and UE distribution

Boolean flags take no separate value: `--tunnel` or `--tunnel=true` enables one and `--tunnel-vrf=false` disables
one, whereas `--tunnel-vrf false` is rejected, `false` being a positional argument.

The first UE uses the configured gNB ID and N2/N3 addresses, and each following UE the next gNB, back to the first
one after the last. Handovers advance through the same sequence. With `--dedicatedGnb`, ascending MSINs therefore
use ascending gNB IDs and N2/N3 addresses.

### Versions and releases

`packetrusher --version` prints the release tag, or the source commit for other builds. Docker builds have no
Git metadata: pass `--build-arg VERSION=$(git describe --tags --always)` to identify them.

Pushing a `v*` or `YYYYMMDD` tag publishes a GitHub release with Linux amd64/arm64 archives and checksums, and a
`ghcr.io/hewlettpackard/packetrusher:<tag>` image. [CHANGELOG.md](CHANGELOG.md) lists the changes of each release,
and its section for the tag becomes the release notes.

### Checks against real cores

The `Real cores` workflow registers a UE against Open5GS, with a PDU session and traffic through its UPF on the
`userspace` and `ebpf` tunnel backends, and against free5GC, registration only as its UPF needs gtp5g.
`test/real-core/ue.sh <packetrusher> <config.yml> [<address to ping>]` runs the same check against a core
you started yourself.

## Contributing
We're thrilled that you'd like to contribute to this project. Your help is essential for keeping it great!   
You can review our [contributing guide](CONTRIBUTING.md).

### Developer's Certificate of Origin
All contributions must include acceptance of the [DCO](DCO.md).

#### Sign your work
To accept the DCO, simply add this line to each commit message with your name and email address (*git commit -s* will do this for you):

    Signed-off-by: Jane Example <jane@example.com>

For legal reasons, no anonymous or pseudonymous contributions are accepted.

## Citation
If you use this software, you may cite it as below:
```latex
@software{PacketRusher,
  author = {D'Emmanuele, Valentin and Raguideau, Akiya},
  doi = {10.5281/zenodo.10446651},
  month = nov,
  title = {{PacketRusher: High performance 5G UE/gNB Simulator and CP/UP load tester}},
  url = {https://github.com/HewlettPackard/PacketRusher},
  version = {1.0.0},
  year = {2023}
}
```

## License
© Copyright 2023 Hewlett Packard Enterprise Development LP

© Copyright 2024-2026 Valentin D'Emmanuele

This project is under the [Apache 2.0 License](LICENSE) license.

By contributing here, [you agree](DCO.md) to license your contribution under the terms of the Apache 2.0 License. All files are released with the Apache License 2.0.

PacketRusher borrows libraries and data structures from the [free5gc project](https://github.com/free5gc/free5gc), and is originally based upon [my5G-RANTester](https://github.com/my5G/my5G-RANTester).
