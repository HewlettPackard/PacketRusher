# Scoped native NAS v1.3.0 patch

This directory contains the native `ie`, `message`, and `internal` packages from
[`github.com/free5gc/nas` v1.3.0](https://github.com/free5gc/nas/tree/v1.3.0), including
upstream tests and the original [Apache-2.0 license](LICENSE). The legacy root,
`nasType`, and `nasMessage` APIs are not copied or used. PacketRusher's root
`go.mod` replaces the pinned module with this directory, so production and aio5gc
use the same corrected native codecs.

Upstream tag commit: `ce9d4b4139dd7cf9a697a43490e637dcbcb82cf8`.
Original module checksum: `h1:r42lNApvMtAwaHwpb50ZpOXZabGaQvFRy/q3JxS0o/g=`.
`UPSTREAM.json` records the module-cache origin and retained paths.

## Changes

- Correct Profile B SUCI allocation from `12 + 33 + 8 + len(CipherVal)` to
  `8 + 33 + 8 + len(CipherVal)`. Native registration, identity response, and
  embedded registration encoding now retain the original ciphertext and MAC.
  Profile A/B encoding also checks exact ephemeral-key and MAC lengths instead
  of silently padding or truncating incorrect native identity fields.
- Implement all 16 stub optional IEs that the NAS v1.1.3 codec supported in
  PacketRusher's dispatched MM responses and nested SM accept/reject/release
  responses. An IEI comparison against the old message definitions determined
  this surface; it is broader than the three Registration Accept reproductions.
- Support UCS-2 Network Name values in Configuration Update Command, including
  complete big-endian characters and malformed-text rejection. The previous
  native implementation only accepted GSM 7-bit text.
- Validate declared LV/TLV and LV-E/TLV-E value lengths against remaining bytes
  before reading or skipping the value, including repeated optional IEs. The
  consumed N1 SM payload container also checks its 1–65535-octet value bounds
  and replaces its previous value on decoder reuse.
- Update two upstream message test expectations for the newly implemented fields
  and the earlier truncated-envelope error; keep the remaining upstream tests.
- Validate local and remote QoS packet-filter ports in `ie/qos_rules.go` against
  0–65535 before converting parsed integers to the 16-bit wire value. Decimal
  parsing and range tokenization retain the upstream contract.

There are no exceptions to parse errors in PacketRusher or aio5gc. MAC failures,
malformed input, and remaining unsupported-IE warnings still fail. Both protected
receive paths commit a cloned counter/context only after successful parsing and
integrity verification. Type-1 codecs require exactly one octet, decode the value
bits, ignore spare bits, and emit zero spare bits.

## Compatibility audit

The audit covered Registration Accept/Reject, Configuration Update Command,
Service Accept/Reject, Authentication Request/Reject, Identity Request, Security
Mode Command, DL NAS Transport, 5GMM/5GSM Status, and nested PDU Session
Establishment Accept/Reject and Release Command/Reject. No baseline-supported
optional IE in the other audited messages corresponds to an unimplemented
native codec.

| Restored native IE | Responses using it | Representation / value length |
| --- | --- | --- |
| MICOInd | Registration Accept, Configuration Update | Typed SPRTI/RAAI bits, 1 octet |
| NSSAIInclusionMode | Registration Accept | Typed two-bit mode, 1 octet |
| Non3GppNWProvidedPolicies | Registration Accept | Typed value bit, 1 octet |
| SMSInd | Configuration Update | Typed SAI bit, 1 octet |
| AlwaysonPDUSessInd | PDU Establishment Accept | Typed APSI bit, 1 octet |
| AllowedSSCMode | PDU Establishment Reject | Typed SSC1/SSC2/SSC3 bits, 1 octet |
| CongestionReattemptIndicator5GSM | PDU Establishment Reject, Release Command | Typed ABO/CATBO bits, 1 octet |
| EmergNumList | Registration Accept | Owned opaque value, 3–48 octets |
| ExtendedEmergNumList | Registration Accept | Owned opaque value, 4–65535 octets |
| SORTransparentCntr | Registration Accept | Owned opaque value, 17–65535 octets |
| OperatorDefinedAccessCategoryDefs | Registration Accept, Configuration Update | Owned opaque value, 0–8320 octets |
| EPSBearerCtxStatus | Registration Accept | Owned opaque value, 2 octets |
| EPSNASSecAlgos | Security Mode Command | Owned opaque value, 1 octet |
| S1UESecCapability | Security Mode Command | Owned opaque value, 2–5 octets |
| AdditionalInfo | DL NAS Transport | Owned opaque value, 1–255 octets |
| MappedEPSBearerCtxs | PDU Establishment Accept | Owned opaque value, 4–65535 octets |
| NwName (partial upstream codec) | Configuration Update full/short name | Typed UCS-2 text, in addition to existing GSM 7-bit text |

Lengths above exclude IEI and length octets. Bounds follow the relevant native
message definitions and [TS 24.501](https://www.etsi.org/deliver/etsi_ts/124500_124599/124501/17.14.00_60/ts_124501v171400p.pdf)
/ [TS 24.008](https://www.etsi.org/deliver/etsi_ts/124000_124099/124008/18.08.00_60/ts_124008v180800p.pdf)
IE formats. Variable TLV-E containers are limited to the 16-bit length envelope;
SOR no longer inherits the old codec's obsolete 2045-octet maximum. Opaque IEs are
not consumed by PacketRusher: their complete, bounded value is retained without
claiming to interpret inner containers or perform SOR, USIM, EPS bearer, or
operator-category procedures. This matches the baseline's raw-value handling.
Opaque decoders and encoders own their buffers. PacketRusher's consumed NAS IEs
continue using typed upstream codecs.

The 24 additional stub codecs below occur in the audited current messages but
were absent from their v1.1.3 definitions. They remain explicitly unsupported:
`UERadioCapabilityID`, `UERadioCapabilityIDDelInd`, `CipheringKeyData`,
`CAGInfoList`, `Truncated5GSTMSICfg`, `WUSAssistanceInfo`, `NBN1ModeDRXParams`,
`ExtendedRejectedNSSAI`, `SvcLvlAACntr`, `PEIPSAssistanceInfo`, `NSSRGInfo`,
`RegWaitRange`, `DisasterPlmnList`, `ExtendedCAGInfoList`, `NSAGInfo`,
`AdditionalCfgInd`, `PriorityIndicator`, `ServingPLMNRateCtrl`, `ATSSSCntr`,
`CtrlPlaneOnlyInd`, `IPHdrCompressionCfg`, `EthHdrCompressionCfg`,
`ReceivedMBSCntr`, and `ReattemptIndicator`.

Multiple/event-notification payload containers remain unsupported upstream;
PacketRusher dispatches only N1 SM payload containers. Unimplemented IEs outside
this audited response surface are unchanged.

## Validation and updates

From the PacketRusher repository root, run the tests using the root dependency
graph (CI runs this separately because `./...` does not traverse nested modules):

```sh
go test -race github.com/free5gc/nas/ie/... github.com/free5gc/nas/message/... github.com/free5gc/nas/internal/...
go test -race ./internal/control_test_engine/ue/nas/... ./test/aio5gc/msg/nas/...
```

PacketRusher tests construct exact inner NAS bytes and apply AES-CTR/CMAC directly,
without the native message encoders or mock response builders. Every restored IE
and both UCS-2 name fields are checked through production and aio5gc receive
paths. Tests also cover nested SM preservation, invalid-MAC and malformed-envelope
counter rollback, unsupported-warning rejection, and actual Registration Accept
state transition plus Registration Complete transmission. Profile A/B tests use
actual production registration and identity builders, including the protected
Security Mode Complete registration container, and independently authenticate /
deconceal the received wire identity using SIDF with known home-network keys.
QoS packet-filter tests assert exact single-port and range bytes at the 0/65535
bounds and reject overflowing, negative, and malformed port endpoints.

On an upstream update, compare these native packages with the new tag, reapply
only corrections still needed, rerun the audit and both test commands, and update
this note plus `UPSTREAM.json`. Remove the local replace once a compatible
upstream release includes these fixes. This patch does not establish live-core
or kernel user-plane interoperability.
