# Load-test results and metrics

Enable collection with one or more global options before the command:

```bash
./packetrusher --config config/config.yml \
  --report-json results.json --report-csv results.csv \
  --metrics-addr 127.0.0.1:9090 multi-ue -n 100
```

Scrape `http://127.0.0.1:9090/metrics` while the scenario runs. Use a different
listen address to expose the endpoint outside the host. Collection is disabled
when none of these options is present. These options also work with `ue` and
custom scenarios that create UE instances.

Send Ctrl-C for the normal shutdown path. JSON and CSV reports are written after
the scenario returns, including when the action returns an error. Existing report
files are preserved: choose new paths for each run. Abrupt process termination or
an existing fatal-exit path can prevent final output; use live metrics for those
cases. The endpoint stops when the scenario exits.

## Measurements

Both registration and PDU session establishment have these counters:

- `started`: procedure attempts started. Retransmits while an attempt is pending
  do not increment it; retries after a reject and registration loops do.
  Delayed session retries are cancelled when the session is deleted, accepted
  or released, or its UE terminates; stale queued retries cannot start attempts.
- `success`: registration reaches the registered state or a PDU session reaches
  the active state.
- `failure`: an explicit authentication/registration or PDU establishment reject,
  including a matching pending establishment request returned in DL NAS Transport
  with a 5GMM refusal cause (for example, an unsupported or unsubscribed DNN).
  This accounting does not introduce automatic retries or backoff policy.
- `cancelled`: an unfinished attempt ended by session deletion, UE termination or scenario shutdown.
- `pending`: attempts still waiting for completion in a live snapshot.

Successful-procedure latency runs from the procedure trigger to the corresponding
successful state transition. Failed and cancelled attempts do not affect these
latencies. The report contains count, mean, minimum, maximum and (JSON) sum in
seconds. The Prometheus endpoint also exposes a cumulative latency histogram.

Metrics are `packetrusher_procedure_started_total`,
`packetrusher_procedure_completed_total`, `packetrusher_procedure_pending`, and
`packetrusher_procedure_duration_seconds`. Labels are bounded to procedure and
outcome. Subscriber identities, SIM keys, and NAS payloads are not exported.
Aggregation uses fixed histogram buckets and retains only currently pending
attempts, so loop runs do not retain a record of every completed UE.

JSON uses `schema_version: 1`, UTC timestamps and a sorted
`procedures` array. CSV has one row per procedure and a header with explicit
units. A procedure with no successes has latency count and values zero.

These measurements describe UE control-plane procedure completion. They do not
measure user-plane throughput, prove that kernel tunnel rules installed, or
measure handover/deregistration latency. A missing response remains pending
until termination; it is not silently classified as a protocol rejection.

Transport refusals must match the current session and establishment transaction.
An unscoped 5GMM Status without a session ID and PTI cannot safely identify a
PDU establishment attempt. It is logged, but does not complete a per-session
outcome; a still-pending attempt is cancelled when its scenario shuts down.
The existing NAS sender uses PTI 1 for each establishment; a delayed wire response
that is identical to a new request after session-ID reuse cannot be distinguished
without changing that transaction allocation. Non-pending sessions, stale context
objects, mismatched IDs/PTIs and duplicate completed attempts are not counted.

The latest free5GC util NAS/NGAP helpers count messages. Procedure attempts and
latencies here are separate measurements and work with both old and migrated
codec APIs; message counters can be added alongside them later.
