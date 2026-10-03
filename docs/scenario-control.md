# Runtime procedure controls

Start a normal load test with an optional local control socket:

```sh
./packetrusher multi-ue -n 2 --number-of-gnbs 2 --control-socket /tmp/packetrusher.sock
```

Each gNB needs its own N2 and N3 address, assigned on the host and consecutive
from the configured addresses. `--number-of-gnbs` distributes UEs across shared
gNBs; it cannot be combined with `--dedicatedGnb`. Tunnels remain disabled unless
`--tunnel` is selected, so these controls need no kernel module or network
administration privileges for control-plane tests. The load-test process keeps
running until SIGINT.

The socket is accessible only to its owner's account (`0600`). PacketRusher
refuses an existing socket or file and removes its own socket on shutdown.

In another terminal, inspect attachments or select a UE and a procedure:

```sh
./packetrusher control --socket /tmp/packetrusher.sock
./packetrusher control --socket /tmp/packetrusher.sock --ue 1 --action wait
./packetrusher control --socket /tmp/packetrusher.sock --ue 1 --action xn-handover --target 000009
./packetrusher control --socket /tmp/packetrusher.sock --ue 1 --action ng-handover --target 000008
./packetrusher control --socket /tmp/packetrusher.sock --ue 1 --action idle
./packetrusher control --socket /tmp/packetrusher.sock --ue 1 --action reconnect
./packetrusher control --socket /tmp/packetrusher.sock --ue 1 --action deregister
./packetrusher control --socket /tmp/packetrusher.sock --ue 1 --action register
```

UE IDs start at one and are independent of gNB IDs. Targets must match configured
gNB IDs reported by `inspect`. Inspect omits subscriber identities, security keys
and addresses. Example response:

```json
{
  "ues": [{
    "ue": 1,
    "generation": 1,
    "connection_generation": 2,
    "state": "registered",
    "gnb": "000009",
    "connected": true,
    "ready": true,
    "active_pdu_sessions": [1]
  }],
  "gnbs": [{"id": "000008", "ready": true}, {"id": "000009", "ready": true}]
}
```

| Action | Readiness and completion |
| --- | --- |
| `inspect` | Return current local UE state and successful NG Setup state. Omit `--ue` to inspect all UEs started so far. |
| `wait` | Wait for registration, all configured PDU sessions and gNB readiness. |
| `idle` | Wait for readiness, request context release and wait for the gNB connection to close normally. |
| `reconnect` | Wait for idle release, reconnect with a service request (registration fallback without a GUTI), and wait for readiness. |
| `xn-handover` | Wait for UE and target readiness, connect to the target and wait for Path Switch acknowledgement and local session setup. |
| `ng-handover` | Send Handover Required for the selected target, then wait for the core's handover command and local target session setup. |
| `deregister` | Send switch-off deregistration, clean up local session resources and park the scenario. Parking is local completion; switch-off has no NAS acknowledgement. |
| `register` | Rearm a parked scenario with fresh registration and PDU sessions. |

Actions on one UE run serially on its existing event loop. Other UEs can progress
and be inspected independently. `--timeout` defaults to 30 seconds, accepts 1 ms
to 2 minutes and covers both readiness and completion. Rejected, cancelled or
stale requests cannot execute during a later registration generation. A
procedure already sent to the core may complete after a timeout; inspect its
state before retrying. Association loss stops that UE and is reported to callers. An unavailable or
full target connection queue is rejected without blocking the source dispatcher.

The legacy deregistration/idle/handover timers start after registration and use
these same readiness checks. Automatic deregistration waits for every configured
PDU session to be active before releasing sessions and ending that iteration.
Its readiness wait is bounded to 30 seconds and cancelled when the iteration
ends. If the core never accepts every PDU, the same UE generation and connection
are terminated at that deadline so registration loops can progress; unfinished
attempts are recorded as cancelled and the deadline is logged. Manual `deregister` continues to park immediately.
Avoid mixing timer-driven procedures with a manually sequenced
scenario on the same UE if precise ordering matters.

NG handover carries PacketRusher's virtual UE identity in simulator metadata
inside the opaque RRC container, allowing UE IDs greater than 256. The marker is
decoded with fixed length and integer bounds, and older PacketRusher containers
using the RFSP index for UE IDs 1–256 remain accepted. This is an exchange between
PacketRusher gNB simulators; its RRC container is simulator metadata. Validation
exercises NGAP encoding and the opaque transfer, without implementing a radio RRC
codec or claiming interoperability with a physical gNB's RRC implementation.

## JSON scenarios

Save ordered steps as `scenario.json`:

```json
{
  "steps": [
    {"ue": 1, "action": "wait", "timeout_ms": 30000},
    {"ue": 1, "action": "xn-handover", "target": "000009"},
    {"ue": 1, "action": "idle"},
    {"ue": 1, "action": "reconnect"},
    {"ue": 1, "action": "deregister"},
    {"ue": 1, "action": "register"}
  ]
}
```

```sh
./packetrusher run-scenario --socket /tmp/packetrusher.sock --scenario scenario.json
```

The complete document is validated before any request is sent. Steps wait for
observed completion, stop on the first error and report the failing step. Unknown
fields, unsupported actions, invalid IDs/targets and timeout values are rejected.
For a churn wave, list cohort UEs in order with `deregister` steps and later
`register` steps; each UE remains parked between waves. The existing WASM
`custom-scenario` command remains available.

The control socket carries HTTP `POST /control` with the same JSON as one step.
Inspection is `{"action":"inspect"}`. All requests have finite deadlines and
return JSON containing `ues`, `gnbs` and an `error` field on failure. No TCP
listener is opened for procedure control.
