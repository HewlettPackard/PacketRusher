# BPF source and embedded object

`gtpu.c` is PacketRusher's Apache-2.0 source, written for the Linux BPF UAPI. It
contains LWT IPv4 encapsulation and owned TCX IPv4 decapsulation. No eUPF source
or UPF/PFCP implementation is included. Linux UAPI headers are build inputs;
Go runtime loading uses Apache-2.0 `github.com/cilium/ebpf` v0.22.0.

Rebuild the adjacent `gtpu_bpfel.o` from the repository root with
`./scripts/build-ebpf.sh`. Normal Go/Docker/release builds embed that object and
do not require clang or libbpf. The object uses little-endian BPF ISA v3; Linux
6.6+ TCX is the limiting runtime prerequisite. Map keys/values are explicit
32-bit native-little-endian fields: IPv4 address bytes retain network order,
TEIDs are host integers, and the 32-byte canonical binding is shared by both
programs. `downlinks` is an alias index, not an independent forwarding authority.

The initially reviewed object was built with Ubuntu clang 18.1.8
(`1:18.1.8-20ubuntu8`) extracted locally, without installing host packages, and
Linux UAPI headers already present. The rebuild command fixes its working-directory
DWARF prefix to `.`. Preserve source/object together and run the actual native
verifier/packet tests whenever rebuilding. Compiler/kernel versions can change
object bytes; reproducibility comparisons must use the same compiler and headers.
