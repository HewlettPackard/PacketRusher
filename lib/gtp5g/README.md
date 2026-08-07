# gtp5g - 5G compatible GTP kernel module
gtp5g is a customized Linux kernel module gtp5g to handle packet by PFCP IEs such as PDR and FAR.
For detailed information, please reference to 3GPP specification TS 29.281 and TS 29.244.

## Notice
Due to the evolution of Linux kernel, this module would not work with every kernel version.
It is known to build and run on kernels from `5.4` (Ubuntu 20.04) or RHEL8 up to and including the `7.0.x` series. Kernels older than that baseline are not supported.

### Kernel 6.18 and 7.0 support
Two networking API changes are handled behind `LINUX_VERSION_CODE` guards, so older
kernels keep building unchanged:

- **`struct flowi4` DSCP field** (kernel 6.18): the `flowi4_tos` member was replaced by
  `flowi4_dscp` (typed `dscp_t`). Route lookups now set it via `inet_sk_dscp()` and
  express the on-link flag through `flowi4_scope`, matching upstream `drivers/net/gtp.c`.
- **`proto_ops->connect` address type** (kernel 7.0): the address argument changed from
  `struct sockaddr *` to `struct sockaddr_unsized *`. The AF_UNIX buffering client casts
  to the matching type per kernel version.

Kernels `7.1` and newer have not been tested.

## Usage
### Clone
#### The latest version
```
git clone https://github.com/free5gc/gtp5g.git
```
#### The specific version
```
# git clone -b {version} https://github.com/free5gc/gtp5g.git
git clone -b v0.8.10 https://github.com/free5gc/gtp5g.git
```
### Install Required Packages
```
sudo apt -y update
sudo apt -y install gcc g++ cmake autoconf libtool pkg-config libmnl-dev libyaml-dev
```

### Compile
```
cd gtp5g
make clean && make
```

### Install kernel module
Install the module to the system and load automatically at boot
```
sudo make install
```

### Remove kernel module
Remove the kernel module from the system
```
sudo make uninstall
```

### DKMS support
When updating the kernel gtp5g needs to get rebuilt against the current kernel source.
This can be automated using [DKMS](https://github.com/dell/dkms).

To use the DKMS support the following steps are required:
1. Copy the repository to `/usr/src/gtp5g-<VERSION>` (e.g., `/usr/src/gtp5g-0.9.5/`).
1. Run the following command to add the DKMS module to the module tree:
   ```
   # sudo dkms add -m gtp5g -v <VERSION>
   sudo dkms add -m gtp5g -v 0.9.5
   ```
1. Run this command to install the DKMS module:
   ```
   # sudo dkms install -m gtp5g -v <VERSION>
   sudo dkms install -m gtp5g -v 0.9.5
   ```

After a reboot of the system everything should be set up.
Whether the kernel module is loaded can be checked by running:
```
lsmod | grep gtp
```
Which should result in a similar output to:
```
gtp5g                 200704  0
udp_tunnel             28672  2 gtp5g,sctp
```

### Check Rules
Get PDR/FAR/QER information by "/proc"
```
# if UPF used legacy netlink struct without SEID, need use #SEID=0 to query related info as below:
echo #interfaceName #SEID #PDRID > /proc/gtp5g/pdr
echo #interfaceName #SEID #FARID > /proc/gtp5g/far
echo #interfaceName #SEID #QERID > /proc/gtp5g/qer
```
```
cat /proc/gtp5g/pdr
cat /proc/gtp5g/far
cat /proc/gtp5g/qer
```

### Enable/Disable QoS (Default is disabled)
Support Session AMBR and MFBR.

1) Check whether QoS is enabled or disabled. (1 means enabled, 0 means disabled)
    ```
    cat /proc/gtp5g/qos
    ```
2) Enable or disable QoS
   + enable
        ```
        echo 1 >  /proc/gtp5g/qos
        ```
   + disable
        ```
        echo 0 >  /proc/gtp5g/qos
        ```

### Enable/Disable GTP-U Sequence Number (Default is enabled)
Support GTP-U Sequence Number.

1) Check whether GTP-U Sequence Number is enabled or disabled. (1 means enabled, 0 means disabled)
    ```
    cat /proc/gtp5g/seq
    ```
2) Enable or disable GTP-U Sequence Number
   + enable
        ```
        echo 1 >  /proc/gtp5g/seq
        ```
   + disable
        ```
        echo 0 >  /proc/gtp5g/seq
        ```

### Log Related
1) check log
    ```
    dmesg
    ```
1) update log level
    ```
    # [number] is 0~4 
    # e.g. echo 4 > /proc/gtp5g/dbg
    echo [number] > /proc/gtp5g/dbg
    ```
### Tools
+ [go-gtp5gnl](https://github.com/free5gc/go-gtp5gnl)
