# PacketRusher

![PacketRusher Logo](docs/media/img/PacketRusher.png)

----
## Description
PacketRusher is a tool, based upon [my5G-RANTester](https://github.com/my5G/my5G-RANTester), dedicated to the performance testing and automatic validation of 5G Core Networks using simulated UE (user equipment) and gNodeB (5G base station).

If you have questions or comments, feel free to open an issue.

PacketRusher borrows libraries and data structures from the [free5gc project](https://github.com/free5gc/free5gc).

## Features
* Simulate multiple UEs and gNodeB from a single tool
  * We tested up to 10k UEs!
* Supports both N2 (NGAP) and N1 (NAS) interfaces for stress testing
* --pcap parameter to capture pcap of N1/N2 traffic
* Implements main control plane procedures:
  * Supports UE attach/detach (registration/authentifcation/security mode) procedures
  * Supports Create/Delete PDU Sessions,  up to 15 PDU Sessions per UE
  * Supports Xn handover: UE handover between simulated gNodeB (PathSwitchRequest)
  * Supports 5G roaming: Tested with new https://github.com/open5gs/open5gs/issues/2194 Roaming feature
* Implements high-performant N3 (GTP-U) interface
  * Generic tunnel supporting all kind of traffic (TCP, UDP, Video…)
    * We tested iperf3 traffic, and Youtube traffic through PacketRusher
    * We roughly reach 5 GB/s per UE, which is more than what a real UE can achieve.
* Integrated all-in-one mocked 5GC/AMF for PacketRusher's integration testing

## Installation
### Quick start guide
The following is a quick start guide, for more details on the installation, configuration or usage, you may refer to the [wiki](https://github.com/HewlettPackard/PacketRusher/wiki).

### Requirements
- Ubuntu 20.04 or RHEL 
  - /!\ gtp5g kernel module does not work on Ubuntu 22.04 / RHEL 9, work is under way to fix that.
- Windows is not supported (Windows does not support SCTP)
- Go 1.21.0 or more recent
- Root privilege
- Secure boot disabled (for custom kernel module)

PacketRusher is not yet supported on Docker.

### Dependencies
```bash
$ sudo apt install build-essentials linux-headers-generic make git wget tar
# Warning this command will remove your existing local Go installation if you have one
$ wget https://go.dev/dl/go1.21.3.linux-amd64.tar.gz && sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.3.linux-amd64.tar.gz
# Download PacketRusher source code
$ git clone https://github.com/HewlettPackard/PacketRusher # or download the ZIP from https://github.com/HewlettPackard/PacketRusher/archive/refs/heads/master.zip and upload it to your Linux server
$ cd PacketRusher && export PACKETRUSHER=$PWD
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
$ go build cmd/packetrusher.go
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

## License
© Copyright 2023 Hewlett Packard Enterprise Development LP

This project is under the [Apache 2.0 License](LICENSE) license.

By contributing here, [you agree](DCO.md) to license your contribution under the terms of the Apache 2.0 License. All files are released with the Apache License 2.0.

PacketRusher borrows libraries and data structures from the [free5gc project](https://github.com/free5gc/free5gc), and is originally based upon [my5G-RANTester](https://github.com/my5G/my5G-RANTester).
