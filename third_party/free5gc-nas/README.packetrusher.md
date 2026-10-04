# free5gc/nas v1.3.0, patched

Copy of [github.com/free5gc/nas v1.3.0](https://github.com/free5gc/nas/tree/v1.3.0) (commit `ce9d4b4`) that
the root `go.mod` uses in place of the upstream module. As `go mod vendor` would, it leaves out the upstream
tests, except the two that the patch changes. It differs from upstream by:

- Profile B SUCI: the encoded identity was 4 bytes too long. Profile A and B also refuse an ephemeral key or
  a MAC of the wrong length.
- Codecs for optional IEs that upstream reports as an error, so that the message carrying one was dropped:
  MICO indication, NSSAI inclusion mode, non-3GPP NW policies, SMS indication, always-on PDU session
  indication, allowed SSC mode, 5GSM congestion re-attempt indicator and, kept as opaque bytes
  (`ie/opaque_value.go`), the (extended) emergency number list, SOR transparent container, operator-defined
  access categories, EPS bearer context status, EPS NAS security algorithms, S1 UE security capability,
  additional information and mapped EPS bearer contexts.
- Network names in UCS-2 (Configuration Update Command), where upstream only has GSM 7-bit.
- An IE whose declared length exceeds the remaining bytes is an error (`message/message.go`), and the N1 SM
  payload container checks its length.
- QoS packet filter ports outside 0-65535 are refused (`ie/qos_rules.go`).

To update: copy `ie`, `message` and `internal` from the new tag without their tests, re-apply what
`diff -r` against v1.3.0 shows is still needed, and run `go test ./...` here. Remove the directory and the
`replace` once upstream has these fixes.
