# Optional uplink CPU investigation

These switches collect controlled profiles separately from accepted throughput.
The ordinary three-backend benchmark defaults and strict real-core/lifecycle
checks remain unchanged. Run only inside the disposable guest. Use the same
immutable client, real core, UPF, UE/QFI/MTU, MSS, affinity and offloads.

```
test/external-cores/benchmarks/run.sh --fixture /path/to/test/external-cores \
  --packetrusher /path/to/frozen/client --prefix /path/to/free5gc \
  --state /absolute/fresh/uplink-baseline --repetitions 3 --seconds 5 \
  --backends ebpf gtp5g --directions uplink --tcp-only
test/external-cores/benchmarks/run.sh --fixture /path/to/test/external-cores \
  --packetrusher /path/to/the/same/client --prefix /path/to/free5gc \
  --state /absolute/fresh/uplink-profile --repetitions 3 --seconds 5 \
  --backends ebpf gtp5g --directions uplink --tcp-only --profile cpu-clock
```

`cpu-clock` uses perf's software event at199Hz and retains raw perf data, flat
symbol costs and caller graphs for each warmup/measurement. perf wraps the actual
source-bound sender in its owned process group, so cancellation joins both.
The profiled interval and child CPU include instrumentation overhead. Do not
substitute its rates for the original unprofiled comparison or call a profile
an optimization. PMU availability, virtual CPU scheduling and sample counts limit
interpretation.

`--program-runtime` additionally reads only the exact client's owned program IDs
from fdinfo and checks PID start time/program tags. It requires guest-only BPF
runtime statistics to have been explicitly enabled; the script changes no sysctl.
Raw program snapshots bracket measurements outside their CPU interval. Enabling
these statistics changes execution overhead: record that state and compare it
separately. Unavailable/reset/different-owner statistics fail the row visibly.

The [Linux6.8 transmit path](https://github.com/torvalds/linux/blob/v6.8/net/core/dev.c)
runs TCX egress before normal checksum/segmentation completion. Preserve the
owned completion boundary when examining candidate optimizations; validate
actual inner checksums, source/TEID/QFI ownership and native parity afterwards.
