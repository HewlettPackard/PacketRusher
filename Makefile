all:
	go build -o packetrusher ./cmd
clean:
	rm ./packetrusher

# Rebuilds the eBPF object embedded in the binary; needs clang and the libbpf headers.
# CI checks that the committed object is the one clang 18 builds.
EBPF = internal/control_test_engine/ue/gtp/ebpfgtp
ebpf:
	cd $(EBPF) && clang -O2 -g -Wall -Werror -target bpf -fdebug-prefix-map=$(CURDIR)/$(EBPF)=. -I/usr/include/$(shell uname -m)-linux-gnu -c bpf/gtpu.c -o gtpu_bpfel.o
	llvm-strip -g $(EBPF)/gtpu_bpfel.o
