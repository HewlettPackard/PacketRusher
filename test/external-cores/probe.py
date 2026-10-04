#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Fail closed on real-core NAS, core state, reporting and UPF packet proofs."""
import argparse
import json
import math
import os
import re
import signal
import socket
import struct
import subprocess
import time
import urllib.error
import urllib.request
from pathlib import Path
from prepare import CORE_IP, RAN_IP, UE_IP, DN_IP, IMSI
from startup import free_registered
from diagnostics import inspect
from ipv6 import checksum, router_advertisement_proof

# All HTTP targets are private fixtures, regardless of the invoking shell's proxies.
HTTP = urllib.request.build_opener(urllib.request.ProxyHandler({}))

def require(condition, message):
    if not condition:
        raise AssertionError(message)

def until(predicate, seconds, label):
    deadline = time.monotonic() + seconds
    last = None
    while time.monotonic() < deadline:
        try:
            value = predicate()
            if value:
                return value
        except (OSError, ValueError, AssertionError) as error:
            last = str(error)
        time.sleep(.1)
    raise TimeoutError(f"{label} did not complete within {seconds}s: {last}")

def open_registered_count(raw, initial=False):
    metric = rb"fivegs_amffunction_rm_registeredsubnbr"
    require(re.search(rb"^# TYPE "+metric+rb" gauge\s*$",raw,re.M) is not None, "Open5GS registration gauge missing")
    values = re.findall(rb"^"+metric+rb"(?:\{[^\n]*\})?\s+(\S+)\s*$", raw, re.M)
    # v2.8.0 creates per-slice samples on the first Registered transition.
    require(bool(values) or initial, "Open5GS registered gauge has no numeric sample after registration")
    numbers = [float(value) for value in values]
    require(all(math.isfinite(value) and value >= 0 and value.is_integer() for value in numbers), "invalid Open5GS registration gauge value")
    return sum(numbers)


def core_snapshot(profile, initial=False):
    with HTTP.open(profile["oam"], timeout=2) as response:
        raw = response.read(1024*1024)
    if profile["core"] == "free5gc":
        value = json.loads(raw)
        require(value is None or isinstance(value, list), "unexpected free5GC registered-UE response")
        require(all(ue.get("Supi") == "imsi-"+IMSI for ue in value or []), "AMF registered an unexpected subscriber")
        return len(value or []), raw
    return open_registered_count(raw,initial), raw


def await_core_count(profile, state, label, expected, seconds, initial=False):
    def matches():
        count, raw = core_snapshot(profile,initial)
        if count != expected:
            return False
        (state/f"core-{label}.txt").write_bytes(raw)
        return True
    until(matches,seconds,f"real AMF {label} count={expected}")

def validate_report(report, sessions):
    require(report.get("schema_version") == 1 and report.get("ended_at"), "report is not completed schema1")
    rows = report.get("procedures", [])
    require(len(rows) == 2 and len({r["procedure"] for r in rows}) == 2, "missing/duplicate procedure report")
    data = {r["procedure"]: r for r in rows}
    for name, count in [("registration", 1), ("pdu_session_establishment", sessions)]:
        row = data[name]
        require(row["started"] == count and row["success"] == count, f"{name} expected {count} actual successes: {row}")
        require(all(row[field] == 0 for field in ["failure", "cancelled", "pending"]), f"{name} has unfinished/failed attempts: {row}")

def gtpu_proof(path, nonce, expected_sequences=(0,1,2), bidirectional=True, family=4, ue_ip=UE_IP, dn_ip=DN_IP):
    require(family in {4,6},'unsupported expected inner IP family')
    packets = {"uplink": [], "downlink": []}
    sequences = {"uplink": set(), "downlink": set()}
    with Path(path).open("rb") as file:
        header = file.read(24)
        require(len(header) == 24, "empty/truncated N3 capture")
        endian = {b"\xd4\xc3\xb2\xa1": "<", b"\xa1\xb2\xc3\xd4": ">", b"\x4d\x3c\xb2\xa1": "<", b"\xa1\xb2\x3c\x4d": ">"}.get(header[:4])
        require(endian is not None and struct.unpack(endian+"I", header[20:24])[0] == 1, "N3 capture must use Ethernet PCAP")
        while record := file.read(16):
            require(len(record) == 16, "truncated PCAP record")
            size = struct.unpack(endian+"IIII", record)[2]
            require(size <= 65535, "unexpected PCAP packet size")
            frame = file.read(size)
            require(len(frame) == size, "truncated PCAP packet")
            if len(frame) < 14+20 or frame[12:14] != b"\x08\x00":
                continue
            ip = frame[14:]
            ihl = (ip[0]&15)*4
            if ip[9] != 17 or ihl < 20 or len(ip) < ihl+16:
                continue
            udp = ip[ihl:]
            if struct.unpack("!H", udp[2:4])[0] != 2152:
                continue
            gtp = udp[8:]
            if len(gtp) < 8 or (gtp[0]&0xf0) != 0x30 or gtp[1] != 255 or nonce not in gtp:
                continue
            teid = struct.unpack("!I", gtp[4:8])[0]
            require(teid > 0, "zero user-plane TEID")
            source, destination = socket.inet_ntoa(ip[12:16]), socket.inet_ntoa(ip[16:20])
            direction = "uplink" if (source,destination) == (RAN_IP,CORE_IP) else "downlink" if (source,destination) == (CORE_IP,RAN_IP) else None
            if direction:
                require(struct.unpack("!H",gtp[2:4])[0] == len(gtp)-8, "invalid GTP-U payload length")
                offset = 8
                if gtp[0]&7:
                    require(len(gtp) >= 12, "truncated GTP-U optional header")
                    offset, extension = 12, gtp[11]
                    require(gtp[0]&4 or not extension, "GTP-U extension flag absent")
                    for _ in range(32):
                        if not extension:
                            break
                        require(offset < len(gtp), "truncated GTP-U extension")
                        size = gtp[offset]*4
                        require(size >= 4 and offset+size <= len(gtp), "invalid GTP-U extension length")
                        extension, offset = gtp[offset+size-1], offset+size
                    require(not extension, "GTP-U extension chain too long")
                inner = gtp[offset:]
                if (len(inner)>=20 and inner[0]>>4==4 and inner[9]!=17) or (len(inner)>=40 and inner[0]>>4==6 and inner[6]!=17):
                    # A closed DN UDP port may return ICMP quoting the nonce;
                    # it is not a UDP echo or evidence of a working DN peer.
                    continue
                require(len(inner)>=28 and inner[0]>>4==family,f"DN payload is not expected IPv{family} UDP")
                if family==4:
                    length,ihl=struct.unpack('!H',inner[2:4])[0],(inner[0]&15)*4
                    require(length==len(inner) and ihl>=20 and inner[9]==17,'invalid inner IPv4/UDP packet')
                    peers=(socket.inet_ntoa(inner[12:16]),socket.inet_ntoa(inner[16:20]))
                else:
                    require(len(inner)>=48 and len(inner)==40+struct.unpack('!H',inner[4:6])[0] and inner[6]==17,'invalid inner IPv6/UDP packet')
                    ihl=40
                    peers=(socket.inet_ntop(socket.AF_INET6,inner[8:24]),socket.inet_ntop(socket.AF_INET6,inner[24:40]))
                    require(inner[46:48]!=b'\0\0','IPv6 UDP checksum absent')
                    require(checksum(inner[8:40]+struct.pack('!I3xB',len(inner)-40,17)+inner[40:])==0,'invalid IPv6 UDP checksum')
                require(peers==((ue_ip,dn_ip) if direction=='uplink' else (dn_ip,ue_ip)),'DN payload has wrong allocated UE/DN addresses')
                require(len(inner) >= ihl+8 and struct.unpack("!H",inner[ihl+4:ihl+6])[0] == len(inner)-ihl, "invalid inner UDP length")
                payload = inner[ihl+8:]
                require(payload.startswith(nonce) and len(payload) == len(nonce)+4, "unique payload missing from inner UDP")
                sequences[direction].add(struct.unpack("!I",payload[-4:])[0])
                port = struct.unpack("!H",inner[ihl+(2 if direction == "uplink" else 0):ihl+(4 if direction == "uplink" else 2)])[0]
                require(port == 9000, "DN payload has wrong UDP peer port")
                packets[direction].append(teid)
    expected = set(expected_sequences)
    require(bool(expected), 'at least one unique sequence is required')
    require(len(packets['uplink']) >= len(expected) and sequences['uplink'] == expected, 'N3 capture lacks distinct uplink payloads')
    if bidirectional:
        require(len(packets['downlink']) >= len(expected) and sequences['downlink'] == expected, 'N3 capture lacks distinct DN echo responses')
    else:
        require(not packets['downlink'], 'missing-DN negative control received a UDP echo')
    return {direction: {"packets": len(teids), "teids": sorted(set(teids))} for direction, teids in packets.items()}


def families(profile):
    return profile.get('user_plane_families',[{'version':4,'ue':UE_IP,'dn':DN_IP}])


def family_nonce(nonce, version):
    # Neither family's identifier is a substring of the other's in dual-stack.
    return f'IPv{version}/'.encode()+nonce


def traffic_route(state, family):
    version=family['version']
    command=['ip','-6' if version==6 else '-4','-json','route','get',family['dn'],'from',family['ue']]
    def selected():
        route=subprocess.check_output(command,text=True)
        require(json.loads(route)[0].get('dev')=='val0000000120',f"IPv{version} UE traffic does not select persistent endpoint: {route}")
        return route
    # IPv6 setup needs an actual UPF/SMF RS/RA exchange after NAS acceptance.
    route=until(selected,15,'actual IPv6 prefix/route') if version==6 else selected()
    suffix='-ipv6' if version==6 else ''
    (state/f'ue-route{suffix}.json').write_text(route)
    if version==6:
        endpoint=json.loads(subprocess.check_output(['ip','-6','-json','addr','show','dev','val0000000120'],text=True))
        require(any(addr.get('local')==family['ue'] and addr.get('prefixlen')==128 for addr in endpoint[0]['addr_info']),'learned IPv6 address is not owned /128')
        (state/'ue-endpoint-ipv6.json').write_text(json.dumps(endpoint,indent=2)+'\n')
        (state/'ue-rule-ipv6.json').write_bytes(subprocess.check_output(['ip','-6','-json','rule','show']))


def echo_probe(family, nonce, sequences, expect_reply=True):
    version=family['version']
    with socket.socket(socket.AF_INET6 if version==6 else socket.AF_INET,socket.SOCK_DGRAM) as sock:
        if version==6:
            sock.setsockopt(socket.IPPROTO_IPV6,socket.IPV6_V6ONLY,1)
        sock.bind((family['ue'],0))
        sock.settimeout(5 if expect_reply else 3)
        for sequence in sequences:
            payload=nonce+struct.pack('!I',sequence)
            sock.sendto(payload,(family['dn'],9000))
            if expect_reply:
                data,peer=sock.recvfrom(4096)
                require(data==payload and peer[:2]==(family['dn'],9000),f'IPv{version} DN response payload/peer mismatch')
            else:
                try:
                    sock.recvfrom(4096)
                except (TimeoutError,ConnectionRefusedError):
                    pass
                else:
                    raise AssertionError(f'IPv{version} missing-DN received a UDP response')

def probe(binary, state):
    state = Path(state).resolve()
    profile = json.loads((state / "profile.json").read_text())
    sessions = profile.get('sessions',0 if profile["core"] == "free5gc" else 1)
    backend = profile.get('tunnel_backend','userspace')
    result, process, capture = {"core": profile["core"], "sessions":sessions, "tunnel_backend":backend if sessions else None, "upf_implementation":profile.get('upf_implementation'), "success": False}, None, None
    traffic=families(profile)
    if profile.get('pdu_session_type'):
        result['pdu_session_type']=profile['pdu_session_type']
    nonce = b"PACKETRUSHER-EXTERNAL-" + os.urandom(32).hex().encode()
    log = (state / "packetrusher.log").open("wb")
    capture_log = (state / "capture.log").open("wb")
    try:
        if backend == 'ebpf' and not profile.get('native'):
            subprocess.run(['ethtool','-K','eth0','tx','off','rx','off','tso','off','gso','off','gro','off'],check=True)
        if sessions or profile.get("native"):
            until(lambda: (state/"core-ready").exists(), 65, "real NF registration and accepted SMF/UPF PFCP association")
        else:
            until(lambda: free_registered(profile), 65, "real free5GC NRF registrations")
        await_core_count(profile,state,"initial",0,60,initial=True)
        command = [str(binary), "--config", str(state/"config/packetrusher.yaml"), "--tunnel-backend", backend, "--report-json", str(state/"report.json"), "multi-ue", "-n", "1", "--numPduSessions", str(sessions), "--control-socket", str(state/"control.sock")]
        if sessions:
            command.extend(["--tunnel","--tunnel-vrf=false"])
            capture = subprocess.Popen(["tcpdump", "-n", "-U", "--immediate-mode", "-i", "eth0", "-w", str(state/"n3.pcap"), "udp", "port", "2152"], stdout=capture_log, stderr=subprocess.STDOUT)
            until(lambda: capture.poll() is None and (state/"n3.pcap").exists() and (state/"n3.pcap").stat().st_size >= 24, 5, "N3 capture startup")
        else:
            command.extend(["--timeBeforeDeregistration", "5000"])
        process = subprocess.Popen(command, stdout=log, stderr=subprocess.STDOUT)
        until(lambda: (state/"control.sock").is_socket() and process.poll() is None, 20, "control socket")
        def control(action, label=None):
            response = subprocess.run([str(binary), "control", "--socket", str(state/"control.sock"), "--ue", "1", "--action", action, "--timeout", "45s"], check=True, capture_output=True, text=True, timeout=50)
            value = json.loads(response.stdout)
            (state/f"{label or action}.json").write_text(json.dumps(value,indent=2)+"\n")
            return value["ues"][0]
        ready = control("wait")
        require(ready["ready"] and ready["connected"] and ready["state"] == "registered" and ready["active_pdu_sessions"] == ([1] if sessions else []), f"incomplete real NAS readiness: {ready}")
        await_core_count(profile,state,"registered",1,5)
        if sessions:
            family_results={}
            for family in traffic:
                traffic_route(state,family)
                version=family['version']
                identifier=family_nonce(nonce,version)
                echo_probe(family,identifier,range(3))
                options={'family':version,'ue_ip':family['ue'],'dn_ip':family['dn']}
                until(lambda:gtpu_proof(state/'n3.pcap',identifier,**options),3,f'IPv{version} three observed uplink/downlink sequences')
                family_results[str(version)]={'ue':family['ue'],'dn':family['dn'],'probe_nonce':identifier.decode(),'user_plane':gtpu_proof(state/'n3.pcap',identifier,**options)}
                if version==6:
                    family_results['6']['router_advertisement']=router_advertisement_proof(state/'n3.pcap',family['ue'])
            result['user_plane']=family_results[str(traffic[0]['version'])]['user_plane']
            # Keep the real core and PDU active while removing only our DN
            # application. This proves that echo acceptance depends on that
            # downstream peer, not control-plane readiness or a local shortcut.
            (state/'dn-disable').write_text('negative control: close only owned DN UDP peer\n')
            until(lambda: (state/'dn-disabled').exists(),3,'owned DN peer closure')
            negative_nonce = b'PACKETRUSHER-MISSING-DN-'+os.urandom(32).hex().encode()
            for family in traffic:
                version=family['version']
                identifier=family_nonce(negative_nonce,version)
                echo_probe(family,identifier,(3,),False)
                options={'family':version,'ue_ip':family['ue'],'dn_ip':family['dn']}
                until(lambda:gtpu_proof(state/'n3.pcap',identifier,(3,),False,**options),3,f'IPv{version} missing-DN uplink with no UDP echo')
                family_results[str(version)]['missing_dn_nonce']=identifier.decode()
                family_results[str(version)]['missing_dn']=gtpu_proof(state/'n3.pcap',identifier,(3,),False,**options)
            still_ready = control('wait','wait-dn-disabled')
            require(still_ready['ready'] and still_ready['connected'] and still_ready['state']=='registered' and still_ready['active_pdu_sessions'] == [1], 'DN closure changed PDU readiness')
            await_core_count(profile,state,'dn-disabled',1,5)
            capture.send_signal(signal.SIGINT)
            require(capture.wait(timeout=5) == 0, "N3 capture failed")
            capture = None
            for family in traffic:
                version=family['version']
                options={'family':version,'ue_ip':family['ue'],'dn_ip':family['dn']}
                family_results[str(version)]['user_plane']=gtpu_proof(state/'n3.pcap',family_nonce(nonce,version),**options)
                family_results[str(version)]['missing_dn']=gtpu_proof(state/'n3.pcap',family_nonce(negative_nonce,version),(3,),False,**options)
            result['user_plane']=family_results[str(traffic[0]['version'])]['user_plane']
            result['missing_dn']=family_results[str(traffic[0]['version'])]['missing_dn']
            if profile.get('pdu_session_type'):
                result['families']=family_results
            retired = control('deregister')
            require(retired['state'] == 'parked' and not retired['ready'] and not retired['connected'] and retired['active_pdu_sessions'] == [], 'manual deregistration did not retire the active PDU and connection')
            if profile.get('pdu_session_type'):
                remaining=json.loads(subprocess.check_output(['ip','-6','-json','addr','show'],text=True))
                require(not any(addr.get('local')==family['ue'] for link in remaining for addr in link['addr_info'] for family in traffic if family['version']==6),'retired PDU retained its global IPv6 address')
                (state/'retired-ipv6-endpoints.json').write_text(json.dumps(remaining,indent=2)+'\n')
                rules=json.loads(subprocess.check_output(['ip','-6','-json','rule','show'],text=True))
                require(not any(rule.get('src','').split('/')[0]==family['ue'] for rule in rules for family in traffic if family['version']==6),'retired PDU retained IPv6 source policy')
                (state/'retired-ipv6-rules.json').write_text(json.dumps(rules,indent=2)+'\n')
        # Free5GC automatic termination must take the positive readiness path,
        # well before the legacy30s unresolved-readiness cleanup fallback.
        await_core_count(profile,state,"deregistered",0,12)
        result["core_registered_counts"] = [0,1,0]
        process.send_signal(signal.SIGINT)
        require(process.wait(timeout=10) == 0, "PacketRusher shutdown failed")
        validate_report(json.loads((state/"report.json").read_text()), sessions)
        require("readiness deadline" not in (state/"packetrusher.log").read_text(), "automatic deregistration used readiness timeout fallback")
        result["success"] = True
    except Exception as error:
        result["error"] = str(error)
        inspect(state,"ran-failure")
        raise
    finally:
        if process and process.poll() is None:
            process.send_signal(signal.SIGINT)
            try: process.wait(timeout=10)
            except subprocess.TimeoutExpired:
                process.kill(); process.wait()
        if capture and capture.poll() is None:
            capture.send_signal(signal.SIGINT)
            try: capture.wait(timeout=5)
            except subprocess.TimeoutExpired:
                capture.kill(); capture.wait()
        log.close(); capture_log.close()
        (state/"result.json").write_text(json.dumps(result,indent=2)+"\n")

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--packetrusher", default="/usr/bin/packetrusher")
    parser.add_argument("--state", required=True)
    args = parser.parse_args()
    probe(Path(args.packetrusher).resolve(), args.state)
