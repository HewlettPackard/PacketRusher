# Real user-plane profiles

The registration-only free5GC job still uses zero PDUs. It cannot claim tunnel-backend or UPF coverage. Full native profiles explicitly select one PDU and a backend:

```sh
test/external-cores/native.sh --core open5gs --prefix /opt/open5gs \
  --packetrusher /path/to/packetrusher --state /absolute/fresh/open5gs-ebpf \
  --sessions 1 --backend ebpf
test/external-cores/native.sh --core free5gc --prefix /opt/free5gc \
  --packetrusher /path/to/packetrusher --state /absolute/fresh/free5gc-ebpf \
  --sessions 1 --backend ebpf --upf free5gc
```

Run these in a disposable Linux guest or test VM. Both require TUN and root's namespace/network permissions. The eBPF profile needs Linux 6.6 or newer, an assigned IPv4 Ethernet N3 address and UDP 2152; the runner disables checksum/GSO offloads only on its two newly owned veth endpoints. Source policy routing is used without a VRF. PacketRusher's custom eBPF backend does not provide UPF functionality.

Manual native runs may select `--backend gtp5g` with the guest module loaded. Configuration generation rejects this selection for Docker/hosted profiles, and zero-PDU registration still cannot claim any backend. Manual benchmarks can pass `--stop-capture-when-ready` to the core supervisor: it retains the actual accepted startup PFCP capture and stops that owned capture before timed data. Ordinary CI/acceptance keeps its captures enabled.

The genuine free5GC v4.3.0 UPF requires a separate gtp5g module. Exact release [UPF source](https://github.com/free5gc/go-upf/blob/616a4ff6cbd2e7de2bccd39399f37425f373bcc8/internal/forwarder/gtp5g.go) accepts versions 0.9.3 through versions below 0.10.3; `versions.json` pins [v0.10.2](https://github.com/free5gc/gtp5g/tree/952fb419130f5fc44cac1874e8183312006b746c). Install/load that module only inside the disposable guest. The runner refuses a genuine-UPF profile when the module is absent.

The full free5GC profile adds real SMF/UPF release configurations and static IPv4 SM subscription plus SM-policy records using the pinned [official WebUI schema](https://github.com/free5gc/webconsole/blob/c40b94b6896109e90e46e59d57ebc5f7fa4c32a2/backend/WebUI/api_sample.go). `--upf open5gs --upf-prefix /opt/open5gs` explicitly selects a hybrid and uses FQDN Network Instance encoding. Configuration generation is supported; hybrid interoperability is not presumed.

A passing result requires all configured SBI NFs to register, an actual accepted PFCP association packet from the configured peers, current association state, AMF counts 0 → 1 → 0, protected NAS readiness, exactly one completed registration and PDU, and three distinct DN UDP echoes with allocated UE source address and bidirectional GTP-U capture. After the positive echoes, the runner closes only its DN echo socket: the PDU and AMF must remain active, a fresh nonce must still appear on uplink, and no UDP echo may return. Failed NAS/PDU attempts are never retried or accepted as success. All core logs, captures, reports and ownership-scoped cleanup survive failure.
