# PacketRusher

![PacketRusher Logo](docs/media/img/PacketRusher.png)

----
## Description
#### Now with SUCI Concealing/Deconcealment (Null-Scheme, Profile A (X25519), Profile B (P-256))!

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
  * UE attach/detach (registration/identity request/authentification/security mode) procedures
  * Create/Delete PDU Sessions, up to 15 PDU Sessions per UE
  * Xn handover: UE handover between simulated gNodeB (PathSwitchRequest)
  * N2 handover: UE handover between simulated gNodeB (HandoverRequired)
  * UE Enter/Exit CM-IDLE procedures (Service Request) 
  * GUTI Re-registration
  * Supports 5G roaming: Tested with new https://github.com/open5gs/open5gs/issues/2194 Roaming feature
* Implements high-performant N3 (GTP-U) interface
  * Generic tunnel supporting all kind of traffic (TCP, UDP, Video…) — see [tunnel ownership and handover](docs/TunnelLifecycle.md).
    * We tested iperf3 traffic, and Youtube traffic through PacketRusher
    * We roughly reach 5 GB/s per UE, which is more than what a real UE can achieve.
* Integrated all-in-one mocked 5GC/AMF for PacketRusher's integration testing

## Installation
### Quick start guide
The following is a quick start guide, for more details on the installation, configuration or usage, you may refer to the [wiki](https://github.com/HewlettPackard/PacketRusher/wiki).

### Requirements
- Ubuntu 20.04-24.04
  - All Linux distributions with kernel from 5.4 up to the 7.0.x series should work, but untested.
  - There might be issues with frankenstein kernel from RHEL/CentOS/Rocky, feel free to open a bug if you encounter one!
- Windows is not supported (Windows does not support SCTP)
- Go 1.26.2 or more recent
- Root privilege
- Secure boot disabled (for custom kernel module)

A Linux container workflow is available in [docker/README.md](docker/README.md).
The host provides SCTP and, for user-plane tunnels, the gtp5g kernel module.

### Dependencies
```bash
$ sudo apt install build-essential linux-headers-generic make git wget tar linux-modules-extra-$(uname -r)
# Warning this command will remove your existing local Go installation if you have one:
$ wget https://go.dev/dl/go1.26.2.linux-amd64.tar.gz && sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.26.2.linux-amd64.tar.gz
# Add go binary to the executable PATH variable:
$ echo 'export PATH=$PATH:/usr/local/go/bin' >> $HOME/.profile
```

### Download PacketRusher source code
```bash
$ git clone https://github.com/HewlettPackard/PacketRusher # or download the ZIP from https://github.com/HewlettPackard/PacketRusher/archive/refs/heads/master.zip and upload it to your Linux server
$ cd PacketRusher && echo "export PACKETRUSHER=$PWD" >> $HOME/.profile
$ source $HOME/.profile
```

### Build free5gc's gtp5g kernel module
```bash
$ cd $PACKETRUSHER/lib/gtp5g
$ make clean && make && sudo make install
# Make sure you have Secure boot disabled if you are unable to install the custom Kernel module
```

### Build PacketRusher CLI
```bash
$ cd $PACKETRUSHER
$ go mod download
$ go build -o packetrusher ./cmd
$ ./packetrusher --help
```

You can edit the configuration in $PACKETRUSHER/config/config.yml as specified [here](https://github.com/HewlettPackard/PacketRusher/wiki/Configuration), and then run a basic scenario using `sudo ./packetrusher ue` while in the $PACKETRUSHER folder.   
More complex scenarios are possible using `sudo ./packetrusher multi-ue`, see `./packetrusher multi-ue --help` for more details.   
For more details on the installation, configuration or usage, you may refer to the [wiki](https://github.com/HewlettPackard/PacketRusher/wiki).

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

© Copyright 2024-2025 Valentin D'Emmanuele

This project is under the [Apache 2.0 License](LICENSE) license.

By contributing here, [you agree](DCO.md) to license your contribution under the terms of the Apache 2.0 License. All files are released with the Apache License 2.0.

PacketRusher borrows libraries and data structures from the [free5gc project](https://github.com/free5gc/free5gc), and is originally based upon [my5G-RANTester](https://github.com/my5G/my5G-RANTester).

For JSON/CSV procedure reports and live Prometheus metrics, see [Load-test results](docs/load-test-results.md).

### Boolean flags and UE distribution

Boolean flags take no separate value. Enable a flag with `--tunnel` or
`--tunnel=true`, and disable a flag with `--tunnel-vrf=false`. For example:

```bash
./packetrusher --config config/config.yml multi-ue -n 2 --tunnel -d --tunnel-vrf=false
```

Flags may appear in either order. Do not write `--tunnel true` or
`--tunnel-vrf false`; those values are positional arguments and are rejected
before configuration loading or network setup.

UE IDs start at 1. The first UE uses the configured gNB ID and N2/N3 addresses;
the next UE uses the next gNB when multiple gNBs are present. Selection wraps
back to the first gNB after the last one, and handovers advance through that
same sequence. With `--dedicatedGnb`, ascending MSINs therefore use ascending
gNB IDs and N2/N3 addresses.

For automatic tunnel MTU calculation and the `ue.tunnelmtu` override, see [Tunnel MTU](docs/tunnel-mtu.md).

For the current codec APIs and validation commands, see [Dependency migration](docs/dependency-migration.md).

### gNB and NR cell identities

`gnodeb.plmnlist.gnbid` is a hexadecimal gNB ID. `gnbidlength` sets its width in bits (22 to 32, default 24) and
`cellid` the cell suffix filling the remaining bits of the 36-bit NR cell identity (default 0):

```yaml
gnodeb:
  plmnlist:
    gnbid: "01ABCDE"
    gnbidlength: 25
    cellid: 3
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
### GTP-U tunnel backends

The global `--tunnel-backend` flag selects what carries the user plane of the UEs:

- `gtp5g`: the [gtp5g](https://github.com/free5gc/gtp5g) kernel module.
- `userspace`: PacketRusher itself, with one TUN device per UE; no kernel module is needed.
- `auto` (default): `gtp5g` when its module is loaded, `userspace` otherwise.

Only `auto` falls back: a backend requested by name that is not available is an error.

```bash
sudo ./packetrusher --tunnel-backend userspace multi-ue -n 10 --tunnel
```
The userspace backend is slower than gtp5g and keeps one TUN device, hence one file descriptor, open per UE.
