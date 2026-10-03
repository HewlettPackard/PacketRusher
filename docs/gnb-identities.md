# gNB and NR cell identities

`gnodeb.plmnlist.gnbid` is an unsigned hexadecimal number without `0x`.
`gnbidlength` selects 22..32 significant bits; omitted or zero means 24 bits,
which preserves existing configurations and default NR cell identities.
`cellid` is an unsigned decimal cell suffix, default zero, fitting in the
remaining `36 - gnbidlength` bits of the NR cell identity.

```yaml
gnodeb:
  plmnlist:
    mcc: "208"
    mnc: "93"
    tac: "000001"
    gnbid: "01ABCDE"
    gnbidlength: 25
    cellid: 3
```

The hex number is right-aligned as a number in YAML. NGAP BIT STRINGs are
left-aligned on the wire and their unused low bits are zero. The same configured
width is used by NG Setup and target gNB identities. NR cell identities combine
that identifier with the cell suffix into exactly 36 bits, including transparent
handover containers and all user-location messages.

Invalid widths, non-hexadecimal identifiers and values outside the chosen width
fail before network setup. Multi-gNB runs validate the highest generated ID
before creating any gNB, so the allocation never wraps into another identifier.
IDs retain fixed-width uppercase hexadecimal formatting in the simulator.

This addresses the variable-length identifier item in issue #51. Converting all
NAS/NGAP builders to one uniform pattern is a separate item in that issue.
