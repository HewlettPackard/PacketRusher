
RHEL8 := $(shell cat /etc/redhat-release 2>/dev/null | grep -c " 8." )
ifneq (,$(findstring 1, $(RHEL8)))
	RHEL8FLAG := -DRHEL8
endif

PWD := $(shell pwd)
MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MAKEFILE_DIR := $(shell dirname $(MAKEFILE_PATH))

CONFIG_MODULE_SIG=n
MODULE_NAME = gtp5g
MOD_KERNEL_PATH := kernel/drivers/net

ifeq ($(KVER),)
	KVER := $(shell uname -r)
endif

ifeq ($(KDIR),)
	KDIR := /lib/modules/$(KVER)/build
endif

ifneq ($(RHEL8FLAG),)
	INSTALL := $(MAKE) -C $(KDIR) M=$$PWD INSTALL_MOD_PATH=$(DESTDIR) INSTALL_MOD_DIR=$(MOD_KERNEL_PATH) modules_install
else
	INSTALL := cp $(MODULE_NAME).ko $(DESTDIR)/lib/modules/$(KVER)/$(MOD_KERNEL_PATH)
	RUN_DEPMOD := true
endif

ifneq ($(RUN_DEPMOD),)
	DEPMOD := /sbin/depmod -a
else
	DEPMOD := true
endif

MY_CFLAGS += -g -DDEBUG $(RHEL8FLAG)
# MY_CFLAGS += -DMATCH_IP # match IP address(in F-TEID) or not
EXTRA_CFLAGS += -Wno-misleading-indentation -Wuninitialized
CC += ${MY_CFLAGS}

EXTRA_CFLAGS += -I $(MAKEFILE_DIR)/include

5G_MOD := src/gtp5g.o

5G_LOG	:= src/log.o

5G_UTIL	:= src/util.o

5G_GTPU	:= src/gtpu/dev.o \
			src/gtpu/encap.o \
			src/gtpu/hash.o \
			src/gtpu/link.o \
			src/gtpu/net.o \
			src/gtpu/pktinfo.o \
			src/gtpu/trTCM.o

5G_GENL := src/genl/genl.o \
			src/genl/genl_version.o \
			src/genl/genl_pdr.o \
			src/genl/genl_far.o \
			src/genl/genl_qer.o \
			src/genl/genl_urr.o \
			src/genl/genl_report.o \
			src/genl/genl_bar.o

5G_PFCP := src/pfcp/api_version.o \
			src/pfcp/pdr.o \
			src/pfcp/far.o \
			src/pfcp/qer.o \
			src/pfcp/urr.o \
			src/pfcp/bar.o \
			src/pfcp/seid.o

5G_PROC := src/proc.o

# Build files
obj-m += $(MODULE_NAME).o
$(MODULE_NAME)-objs := $(5G_MOD) $(5G_LOG) $(5G_UTIL) $(5G_GTPU) \
						$(5G_GENL) $(5G_PFCP) $(5G_PROC)

default: module

module:
	$(MAKE) -C $(KDIR) M=$(PWD) modules
clean:
	$(MAKE) -C $(KDIR) M=$(PWD) clean
 
install:
	$(INSTALL)
	modprobe udp_tunnel
	$(DEPMOD)
	modprobe $(MODULE_NAME)
	echo "gtp5g" >> /etc/modules

uninstall:
	rm -f $(DESTDIR)/lib/modules/$(KVER)/$(MOD_KERNEL_PATH)/$(MODULE_NAME).ko
	$(DEPMOD)
	sed -zi "s/gtp5g\n//g" /etc/modules
	rmmod -f  $(MODULE_NAME)
