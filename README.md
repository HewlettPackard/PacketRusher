# PacketRusher

![PacketRusher Logo](docs/media/img/PacketRusher.png)

----
## Description

PacketRusher is a tool, based upon [my5G-RANTester](https://github.com/my5G/my5G-RANTester), dedicated to the performance testing and automatic validation of 5G Core Networks using simulated UE (user equipment) and gNodeB (5G base station).

If you have questions or comments, please email us: [PacketRusher team](mailto:packetrusher@hpe.com).

PacketRusher borrows libraries and data structures from the [free5gc project](https://github.com/free5gc/free5gc).

## Installation 
### Requirements
- Ubuntu 20.04 or RHEL 8
- Windows is not supported (Windows does not support SCTP)
- Go 1.21.0 or more recent
- Root privilege
- Secure boot disabled (for custom kernel module)

### Dependencies
```bash
$ sudo apt install build-essentials linux-headers-generic make git wget tar
# Warning this command will remove your existing local Go installation if you have one
$ wget https://go.dev/dl/go1.21.3.linux-amd64.tar.gz && sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.3.linux-amd64.tar.gz
# Download PacketRusher source code
$ git clone https://github.hpe.com/valentin-demmanuele/PacketRusher # or download the ZIP from https://github.hpe.com/valentin-demmanuele/PacketRusher/archive/refs/heads/master.zip and upload it to your Linux server
$ cd PacketRusher && export PACKETRUSHER=$PWD
```

### Build free5gc's gtp5g kernel module
```bash
$ cd $PACKETRUSHER/lib/gtp5g
$ make clean && make && sudo make install
```

### Build PacketRusher CLI
```bash
$ cd $PACKETRUSHER
$ go mod download
$ go build cmd/packetrusher.go
$ ./packetrusher --help
```

You can edit the configuration in $PACKETRUSHER/config/config.yml, and then run a basic scenario using `./packetrusher multi-ue -n 1` while in the $PACKETRUSHER folder.