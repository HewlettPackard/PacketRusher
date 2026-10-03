#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Generate isolated real-core fixtures; JSON is valid YAML for both cores."""
import argparse
import copy
import json
from pathlib import Path

HERE = Path(__file__).resolve().parent
CORE_IP, RAN_IP, UE_IP, DN_IP = "172.30.5.20", "172.30.5.30", "10.45.0.2", "10.45.0.1"
IMSI, KEY, OPC = "208930000000120", "00112233445566778899AABBCCDDEEFF", "00112233445566778899AABBCCDDEEFF"
FREE_IPS = {"nrf": "172.30.5.11", "ausf": "172.30.5.12", "udm": "172.30.5.13", "udr": "172.30.5.14", "nssf": "172.30.5.15", "pcf": "172.30.5.16", "amf": CORE_IP}

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

def generate(core, output, native=False, prefix="/opt/open5gs"):
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
    if core == "free5gc":
        values = json.loads((HERE / "free5gc-configs.json").read_text())
        addresses = FREE_IPS if not native else {nf: f"127.0.0.{i+11}" for i, nf in enumerate(FREE_IPS)} | {"amf": CORE_IP}
        mapping = {f"{nf}.free5gc.org": ip for nf, ip in addresses.items()}
        mapping["mongodb://db:27017"] = f"mongodb://{mongo}:27017"
        values = replace(values, mapping)
        values["nrf"]["configuration"]["sbi"]["oauth"] = False
        for nf, value in values.items():
            write(config / f"{nf}cfg.yaml", value)
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
        statements = [f'db.getCollection({json.dumps(name)}).insertOne({json.dumps(value)});' for name, value in documents.items()]
        oam = f'http://{CORE_IP}:8000/namf-oam/v1/registered-ue-context'
        nf_addresses = addresses
    else:
        nf_addresses = {"nrf": "127.0.0.10", "ausf": "127.0.0.11", "udm": "127.0.0.12", "udr": "127.0.0.20", "pcf": "127.0.0.13", "nssf": "127.0.0.14", "smf": "127.0.0.4", "amf": "127.0.0.5"}
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
                value[nf].update({"pfcp": {"server": [{"address": addr}], "client": {"upf": [{"address": CORE_IP}]}}, "gtpc": {"server": [{"address": addr}]}, "gtpu": {"server": [{"address": addr}]}, "session": [session], "info": [{"s_nssai": [slice_ | {"dnn": ["internet"]}]}], "dns": [DN_IP], "mtu": 1400, "freeDiameter": str(Path(prefix) / "etc/freeDiameter/smf.conf")})
            write(config / f"{nf}.yaml", value)
        write(config / "upf.yaml", {"logger": {"file": {"path": f"{logs}/upf.log"}}, "global": {"max": {"ue": 16}}, "upf": {"pfcp": {"server": [{"address": CORE_IP}]}, "gtpu": {"server": [{"address": CORE_IP}]}, "session": [session | {"dev": "ogstun"}]}})
        qos = {"index": 9, "arp": {"priority_level": 8, "pre_emption_capability": 1, "pre_emption_vulnerability": 2}}
        ambr = {"uplink": {"value": 1, "unit": 3}, "downlink": {"value": 1, "unit": 3}}
        subscriber = {"schema_version": 1, "imsi": IMSI, "security": {"k": KEY, "opc": OPC, "op": None, "amf": "8000"}, "ambr": ambr, "slice": [slice_ | {"default_indicator": True, "session": [{"name": "internet", "type": 1, "qos": qos, "ambr": ambr, "ue": {"ipv4": UE_IP}, "pcc_rule": []}]}], "access_restriction_data": 32, "network_access_mode": 0, "subscriber_status": 0, "operator_determined_barring": 0, "subscribed_rau_tau_timer": 12}
        statements = [f'db.subscribers.insertOne({json.dumps(subscriber)});']
        oam = f"http://{CORE_IP}:9090/metrics"
    with (output / "subscriber.js").open("x") as file:
        file.write(f'if (db.getName() !== {json.dumps(core)}) throw new Error("unexpected provisioning database");\n' + "\n".join(statements) + "\n")
    ue = {"hplmn": plmn, "msin": "0000000120", "routingindicator": "0000", "protectionScheme": 0, "homeNetworkPublicKeyID": 0, "key": KEY, "opc": OPC, "amf": "8000", "sqn": "000000000000", "dnn": "internet", "snssai": {"sst": "01", "sd": "010203"}, "pdusessiontype": "IPv4", "tunnelbackend": "userspace", "tunnelmtu": 1400, "integrity": {"nia2": True}, "ciphering": {"nea0": True}}
    write(config / "packetrusher.yaml", {"gnodeb": {"controlif": {"ip": RAN_IP, "port": 9487}, "dataif": {"ip": RAN_IP, "port": 2152}, "plmnlist": plmn | {"tac": "000001", "gnbid": "000008", "gnbidlength": 24}, "slicesupportlist": {"sst": "01", "sd": "010203"}}, "ue": ue, "amfif": [{"ip": CORE_IP, "port": 38412}], "logs": {"level": 4}})
    write(output / "profile.json", {"core": core, "oam": oam, "nf_addresses": nf_addresses, "native": native})

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--core", choices=["free5gc", "open5gs"], required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--native", action="store_true")
    parser.add_argument("--prefix", default="/opt/open5gs")
    args = parser.parse_args()
    generate(args.core, args.output, args.native, args.prefix)
