## Install for old version of RHEL9
It is possible to run PacketRusher on an old version of RHEL9. It has been tested on 5.14.0-70.13.1 using Rocky-9.0-20220805.0-x86_64-minimal.iso from https://dl.rockylinux.org/vault/rocky/9.0/isos/x86_64/ using following install procedure:

### Dependencies
```bash
# Change yum repo to the archived 9.0 one
$ sudo find /etc/yum.repos.d -type f -exec sed -i -e "s,#baseurl=http://dl.rockylinux.org/\$contentdir/\$releasever,baseurl=https://download.rockylinux.org/vault/rocky/9.0,g" {} \;
$ sudo find /etc/yum.repos.d -type f -exec sed -i -e "s,mirrorlist,#mirrorlist,g" {} \;
$ sudo find /etc/yum.repos.d -type f -exec sed -i -e "s,/os/,/kickstart/,g" {} \;

# Install the required packages
$ sudo yum install git wget tar make automake gcc gcc-c++ kernel-devel-$(uname -r) kernel-modules-extra-$(uname -r) lksctp-tools lksctp-tools-devel lksctp-tools-doc mokutil

# Install sctp kernel module
$ echo sctp | sudo tee /etc/modules-load.d/sctp.conf
$ sudo sed -i 's/blacklist sctp/#blacklist sctp/g' /etc/modprobe.d/sctp-blacklist.conf
$ sudo sed -i 's/blacklist sctp_diag/#blacklist sctp_diag/g' /etc/modprobe.d/sctp_diag-blacklist.conf
$ sudo modprobe sctp
# If you are unable to install the custom Kernel module, make sure you have Secure boot disabled by running "mokutil --sb-state".

# Install Golang
# Warning this command will remove your existing local Go installation if you have one:
$ wget https://go.dev/dl/go1.21.3.linux-amd64.tar.gz && sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.3.linux-amd64.tar.gz
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
# If you are unable to install the custom Kernel module, make sure you have Secure boot disabled by running "mokutil --sb-state".
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