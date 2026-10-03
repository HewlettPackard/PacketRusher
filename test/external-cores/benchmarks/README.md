# Three-backend real-core iperf3 comparison

This manual benchmark uses the exact genuine free5GC v4.3.0 control plane and UPF with guest-only gtp5g0.10.2. PacketRusher switches among userspace, its own eBPF program and gtp5g using one frozen CLI. It does not change the no-module hosted CI profiles. Only the benchmark subscription raises both UE/session AMBR to10Gbps so the acceptance fixture's1Gbps limit cannot mask capacity.

Run only inside the disposable Ubuntu24/Linux6.8 guest after loading the pinned module. `run.sh` creates private mount/network namespaces; its RAN child namespace and all processes belong to this run. A fresh private Mongo/subscription, actual NF registration and accepted PFCP capture establish the core. All three backends use UE10.45.0.2, MTU1400, source policy routing without VRF, veth offloads disabled, CPU affinity0–3, GOMAXPROCS2 and info logging. Per cohort, actual AMF counts0→1→1→0, protected PDU readiness, three source-bound echoes with typed N3 capture, PDU retirement and the completed procedure report are mandatory.

Startup PFCP and preflight N3 captures stop before timing. The benchmark uses iperf3's [documented client/server, source binding, reverse direction and JSON output](https://software.es.net/iperf/invoking.html). TCP MSS and UDP payload are both1200 bytes below the UE MTU. Separate one-second warmups precede five-second measured TCP uplink/downlink and UDP50/200/800/2000Mbit/s uplink/downlink flows. Three seeded, shuffled backend cohorts per repetition retain every raw JSON, stderr, failure and schedule.

```sh
test/external-cores/benchmarks/run.sh --fixture /path/to/code/test/external-cores \
  --packetrusher /path/to/one/exact/packetrusher \
  --prefix /home/tester/packetrusher-real-cores/free5gc \
  --state /absolute/fresh/benchmark --repetitions 3 --seconds 5
python3 test/external-cores/benchmarks/summarize.py /absolute/fresh/benchmark
```

Raw `/proc/stat`/`/proc/softirqs` and live NF/PacketRusher/iperf-server process ticks bracket every measured flow; reaped-client CPU comes from child rusage. CPU percentages use one logical CPU as100%, so aggregate guest usage can exceed100%. These intervals include client startup/teardown, while iperf throughput uses its own receiver interval. Process CPU does not include softirq work; guest counters expose that separate cost. Steal is reported separately from consumed CPU. Source/binary/core hashes, module/runtime versions, routes, MTU and offload settings are retained. This VM, one stream, CPU affinity, real UPF and application/core CPU can limit the result; it does not establish absolute production throughput. Failed or missing trials remain explicit in summaries; a failed lifecycle/report cohort cannot contribute to accepted medians. `--pilot` explicitly labels setup-only executions with fewer repetitions; final comparison requires at least3.

No timed results have been claimed until the execution evidence and final report are written.
