#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Generate isolated real-core fixtures; JSON is valid YAML for both cores."""
import argparse
import copy
import json
from pathlib import Path

HERE = Path(__file__).resolve().parent
CORE_IP, RAN_IP, UE_IP, DN_IP = "172.30.5.20", "172.30.5.30", "10.45.0.2", "10.45.0.1"
UE_IPV6, DN_IPV6 = "2001:db8:cafe:1::2", "2001:db8:cafe::1"
IMSI, KEY, OPC = "208930000000120", "00112233445566778899AABBCCDDEEFF", "00112233445566778899AABBCCDDEEFF"
FREE_IPS = {"nrf": "172.30.5.11", "ausf": "172.30.5.12", "udm": "172.30.5.13", "udr": "172.30.5.14", "nssf": "172.30.5.15", "pcf": "172.30.5.16", "amf": CORE_IP}
SMF_IP = "172.30.5.17"


def profile_options(core, sessions=None, backend="userspace", upf=None, native=False, pdu_type="IPv4"):
    sessions = (0 if core == "free5gc" else 1) if sessions is None else sessions
    if core not in {"free5gc", "open5gs"} or sessions not in {0, 1}:
        raise ValueError("profiles require a supported real core and zero or one PDU")
    if backend not in {"userspace", "ebpf"} and not (native and backend == "gtp5g"):
        raise ValueError("gtp5g requires an explicit native profile; hosted profiles use userspace or ebpf")
    if core == "open5gs" and sessions != 1:
        raise ValueError("Open5GS acceptance requires a real PDU and traffic")
    if sessions == 0 and (backend != "userspace" or upf is not None):
        raise ValueError("registration-only does not exercise a tunnel backend or UPF")
    upf = ("free5gc" if core == "free5gc" else "open5gs") if sessions and upf is None else upf
    if sessions and (upf not in {"free5gc", "open5gs"} or (core == "open5gs" and upf != "open5gs")):
        raise ValueError("unsupported real UPF implementation")
    if pdu_type not in {"IPv4", "IPv6", "IPv4v6"}:
        raise ValueError("unsupported requested PDU address family")
    if pdu_type != "IPv4" and (not native or core != "open5gs" or sessions != 1 or backend not in {"userspace", "ebpf"}):
        raise ValueError("IPv6/dual-stack acceptance requires an explicit native Open5GS portable-backend profile; free5GC v4.3 SMF supports IPv4")
    return sessions, backend, upf

def write(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("x") as file:
        json.dump(value, file, indent=2)
        file.write("\n")

def replace(value, mapping):
    if isinstance(value, dict):
        return {k: replace(v, mapping) for k, v in value.items() if k not in {"tls", "cert", "rootcert", "nrfCertPem"}}
    if isinstance(value, list):
        return [replace(v, mapping) for v in value]
    if isinstance(value, str):
        for source, target in mapping.items():
            value = value.replace(source, target)
    return value

def generate(core, output, native=False, prefix="/opt/open5gs", sessions=None, backend="userspace", upf=None, pdu_type="IPv4"):
    sessions, backend, upf = profile_options(core, sessions, backend, upf, native, pdu_type)
    output = Path(output).resolve()
    output.mkdir(parents=True, exist_ok=True)
    config = output / "config"
    config.mkdir(exist_ok=True)
    logs = str(output / "core") if native else "/artifacts/core"
    (output / "core").mkdir(exist_ok=True)
    mongo = "127.0.0.1" if native else "172.30.5.10"
    plmn = {"mcc": "208", "mnc": "93"}
    slice_ = {"sst": 1, "sd": "010203"}
    session = {"subnet": "10.45.0.0/16", "gateway": DN_IP, "dnn": "internet"}
    session6 = {"subnet": "2001:db8:cafe::/48", "gateway": DN_IPV6, "dnn": "internet"}
    session_pools = ([session] if pdu_type != "IPv6" else []) + ([session6] if pdu_type != "IPv4" else [])
    if core == "free5gc":
        values = json.loads((HERE / "free5gc-configs.json").read_text())
        addresses = dict(FREE_IPS) if not native else {nf: f"127.0.0.{i+11}" for i, nf in enumerate(FREE_IPS)} | {"amf": CORE_IP}
        if sessions:
            values.update(json.loads((HERE / "free5gc-userplane-configs.json").read_text()))
            addresses["smf"] = "127.0.0.18" if native else SMF_IP
        # UPF is not an SBI NF and must not be treated as an NRF/listener entry.
        mapping = {f"{nf}.free5gc.org": ip for nf, ip in addresses.items()}
        mapping["upf.free5gc.org"] = CORE_IP
        mapping["mongodb://db:27017"] = f"mongodb://{mongo}:27017"
        values = replace(values, mapping)
        values["nrf"]["configuration"]["sbi"]["oauth"] = False
        if sessions:
            smf = values["smf"]["configuration"]
            smf['plmnList'] = [plmn]
            smf["snssaiInfos"] = [{"sNssai": slice_, "dnnInfos": [{"dnn": "internet", "dns": {"ipv4": DN_IP}}]}]
            node = smf["userplaneInformation"]["upNodes"]["UPF"]
            node["sNssaiUpfInfos"] = [{"sNssai": slice_, "dnnUpfInfoList": [{"dnn": "internet", "pools": [{"cidr": "10.45.1.0/24"}], "staticPools": [{"cidr": "10.45.0.0/24"}]}]}]
            # Open5GS follows the FQDN-encoded Network Instance IE; native
            # free5UPF uses the release's ordinary Network Instance encoding.
            smf["nwInstFqdnEncoding"] = upf == "open5gs"
            values["upf"]["dnnList"] = [{"dnn": "internet", "cidr": "10.45.0.0/16"}]
        for nf, value in values.items():
            if nf == "upf" and upf == "open5gs":
                continue
            write(config / f"{nf}cfg.yaml", value)
        if sessions:
            # SMF requires this auxiliary file even with the default single-UPF
            # topology. Empty optional overrides preserve upNodes/links routing.
            write(config/'uerouting.yaml',{'info':{'version':'1.0.7','description':'Default configured single-UPF path'},'ueRoutingInfo':{}})
        if sessions and upf == "open5gs":
            write(config / "upf.yaml", {"logger": {"file": {"path": f"{logs}/upf.log"}}, "global": {"max": {"ue": 16}}, "upf": {"pfcp": {"server": [{"address": CORE_IP}]}, "gtpu": {"server": [{"address": CORE_IP}]}, "session": [session | {"dev": "ogstun"}]}})
        supi = "imsi-" + IMSI
        auth = {"ueId": supi, "authenticationMethod": "5G_AKA", "encPermanentKey": KEY, "encOpcKey": OPC, "authenticationManagementField": "8000", "sequenceNumber": {"sqnScheme": "GENERAL", "sqn": "000000000020"}}
        am = {"ueId": supi, "servingPlmnId": "20893", "gpsis": ["msisdn-0900000000"], "nssai": {"defaultSingleNssais": [slice_], "singleNssais": [slice_]}, "subscribedUeAmbr": {"uplink": "1 Gbps", "downlink": "1 Gbps"}}
        selection = {"ueId": supi, "servingPlmnId": "20893", "subscribedSnssaiInfos": {"01010203": {"dnnInfos": [{"dnn": "internet", "defaultDnnIndicator": True}]}}}
        documents = {
            "subscriptionData.authenticationData.authenticationSubscription": auth,
            "subscriptionData.provisionedData.amData": am,
            "subscriptionData.provisionedData.smfSelectionSubscriptionData": selection,
            "policyData.ues.amData": {"ueId": supi, "subscCats": ["free5gc"]},
        }
        if sessions:
            # Pinned WebUI api_sample.go and api_webui.go's SM/SM-policy schema.
            documents["subscriptionData.provisionedData.smData"] = {"ueId": supi, "servingPlmnId": "20893", "singleNssai": slice_, "dnnConfigurations": {"internet": {"pduSessionTypes": {"defaultSessionType": "IPV4", "allowedSessionTypes": ["IPV4"]}, "sscModes": {"defaultSscMode": "SSC_MODE_1", "allowedSscModes": ["SSC_MODE_1"]}, "sessionAmbr": {"uplink": "1 Gbps", "downlink": "1 Gbps"}, "5gQosProfile": {"5qi": 9, "arp": {"priorityLevel": 8}, "priorityLevel": 8}, "staticIpAddress": [{"ipv4Addr": UE_IP}]}}}
            documents["policyData.ues.smData"] = {"ueId": supi, "smPolicySnssaiData": {"01010203": {"snssai": slice_, "smPolicyDnnData": {"internet": {"dnn": "internet"}}}}}
        statements = [f'db.getCollection({json.dumps(name)}).insertOne({json.dumps(value)});' for name, value in documents.items()]
        oam = f'http://{CORE_IP}:8000/namf-oam/v1/registered-ue-context'
        nf_addresses = addresses
    else:
        nf_addresses = {"nrf": "127.0.0.10", "ausf": "127.0.0.11", "udm": "127.0.0.12", "udr": "127.0.0.20", "pcf": "127.0.0.13", "nssf": "127.0.0.14", "bsf": "127.0.0.15", "smf": "127.0.0.4", "amf": "127.0.0.5"}
        for nf, addr in nf_addresses.items():
            sbi = {"server": [{"address": addr, "port": 7777}]}
            if nf != "nrf":
                sbi["client"] = {"nrf": [{"uri": "http://127.0.0.10:7777"}]}
            value = {"logger": {"file": {"path": f"{logs}/{nf}.log"}, "level": "info"}, "global": {"max": {"ue": 16}}, nf: {"sbi": sbi}}
            if nf in {"udr", "pcf"}:
                value["db_uri"] = f"mongodb://{mongo}:27017/open5gs"
            if nf == "nrf":
                value[nf]["serving"] = [{"plmn_id": plmn}]
            if nf == "amf":
                # Required by v2.8.0; use the official AMF template's 540s.
                value[nf]["time"] = {"t3512": {"value": 540}}
                value[nf].update({"ngap": {"server": [{"address": CORE_IP}]}, "metrics": {"server": [{"address": CORE_IP, "port": 9090}]}, "guami": [{"plmn_id": plmn, "amf_id": {"region": 2, "set": 1}}], "tai": [{"plmn_id": plmn, "tac": 1}], "plmn_support": [{"plmn_id": plmn, "s_nssai": [slice_]}], "security": {"integrity_order": ["NIA2"], "ciphering_order": ["NEA0"]}, "amf_name": "packetrusher-ci", "network_name": {"full": "Open5GS"}})
            if nf == "nssf":
                sbi["client"]["nsi"] = [{"uri": "http://127.0.0.10:7777", "s_nssai": slice_}]
            if nf == "smf":
                value[nf].update({"pfcp": {"server": [{"address": addr}], "client": {"upf": [{"address": CORE_IP}]}}, "gtpc": {"server": [{"address": addr}]}, "gtpu": {"server": [{"address": addr}]}, "session": session_pools, "info": [{"s_nssai": [slice_ | {"dnn": ["internet"]}]}], "dns": [DN_IPV6] if pdu_type == "IPv6" else [DN_IP], "mtu": 1400, "freeDiameter": str(Path(prefix) / "etc/freeDiameter/smf.conf")})
            write(config / f"{nf}.yaml", value)
        write(config / "upf.yaml", {"logger": {"file": {"path": f"{logs}/upf.log"}}, "global": {"max": {"ue": 16}}, "upf": {"pfcp": {"server": [{"address": CORE_IP}]}, "gtpu": {"server": [{"address": CORE_IP}]}, "session": [pool | {"dev": "ogstun"} for pool in session_pools]}})
        qos = {"index": 9, "arp": {"priority_level": 8, "pre_emption_capability": 1, "pre_emption_vulnerability": 2}}
        ambr = {"uplink": {"value": 1, "unit": 3}, "downlink": {"value": 1, "unit": 3}}
        subscriber = {"schema_version": 1, "imsi": IMSI, "security": {"k": KEY, "opc": OPC, "op": None, "amf": "8000"}, "ambr": ambr, "slice": [slice_ | {"default_indicator": True, "session": [{"name": "internet", "type": 1, "qos": qos, "ambr": ambr, "ue": {"ipv4": UE_IP}, "pcc_rule": []}]}], "access_restriction_data": 32, "network_access_mode": 0, "subscriber_status": 0, "operator_determined_barring": 0, "subscribed_rau_tau_timer": 12}
        if pdu_type != "IPv4":
            subscriber_session = subscriber["slice"][0]["session"][0]
            subscriber_session["type"] = {"IPv6": 2, "IPv4v6": 3}[pdu_type]
            subscriber_session["ue"] = ({"ipv4": UE_IP} if pdu_type == "IPv4v6" else {}) | {"ipv6": UE_IPV6}
        statements = [f'db.subscribers.insertOne({json.dumps(subscriber)});']
        oam = f"http://{CORE_IP}:9090/metrics"
    with (output / "subscriber.js").open("x") as file:
        file.write(f'if (db.getName() !== {json.dumps(core)}) throw new Error("unexpected provisioning database");\n' + "\n".join(statements) + "\n")
    ue = {"hplmn": plmn, "msin": "0000000120", "routingindicator": "0000", "protectionScheme": 0, "homeNetworkPublicKeyID": 0, "key": KEY, "opc": OPC, "amf": "8000", "sqn": "000000000000", "dnn": "internet", "snssai": {"sst": "01", "sd": "010203"}, "pdusessiontype": pdu_type, "tunnelbackend": backend, "tunnelmtu": 1400, "integrity": {"nia2": True}, "ciphering": {"nea0": True}}
    write(config / "packetrusher.yaml", {"gnodeb": {"controlif": {"ip": RAN_IP, "port": 9487}, "dataif": {"ip": RAN_IP, "port": 2152}, "plmnlist": plmn | {"tac": "000001", "gnbid": "000008", "gnbidlength": 24}, "slicesupportlist": {"sst": "01", "sd": "010203"}}, "ue": ue, "amfif": [{"ip": CORE_IP, "port": 38412}], "logs": {"level": 4}})
    profile = {"core": core, "oam": oam, "nf_addresses": nf_addresses, "native": native, "sessions": sessions, "tunnel_backend": backend, "upf_implementation": upf, "pfcp_smf_ip": nf_addresses.get("smf"), "pfcp_upf_ip": CORE_IP if sessions else None}
    if pdu_type != "IPv4":
        profile.update(pdu_session_type=pdu_type, user_plane_families=([{"version": 4, "ue": UE_IP, "dn": DN_IP}] if pdu_type == "IPv4v6" else []) + [{"version": 6, "ue": UE_IPV6, "dn": DN_IPV6}])
    write(output / "profile.json", profile)

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--core", choices=["free5gc", "open5gs"], required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--native", action="store_true")
    parser.add_argument("--prefix", default="/opt/open5gs")
    parser.add_argument("--sessions", type=int, choices=[0, 1])
    parser.add_argument("--backend", choices=["userspace", "ebpf", "gtp5g"], default="userspace")
    parser.add_argument("--upf", choices=["free5gc", "open5gs"])
    parser.add_argument("--pdu-session-type", choices=["IPv4", "IPv6", "IPv4v6"], default="IPv4", help="IPv6/dual-stack: explicit native Open5GS profiles only")
    args = parser.parse_args()
    generate(args.core, args.output, args.native, args.prefix, args.sessions, args.backend, args.upf, args.pdu_session_type)
